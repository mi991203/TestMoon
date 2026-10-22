import vm from 'node:vm';
import readline from 'node:readline';
import { AsyncLocalStorage } from 'node:async_hooks';
import { STATUS_CODES } from 'node:http';
import { expect, Assertion } from 'chai';

// This process runs only inside a per-run, network-disabled container. vm is an API
// surface, not the security boundary; the supervisor owns network and credentials.
const output = (message) => process.stdout.write(`${JSON.stringify(message)}\n`);
const input = readline.createInterface({ input: process.stdin });
const replies = new Map();
let launched = false;
input.on('line', (line) => {
  try {
    const message = JSON.parse(line);
    if (!launched) { launched = true; execute(message); }
    else if (message.type === 'http_result') {
      const pending = replies.get(message.id);
      if (pending) { replies.delete(message.id); pending(message.response); }
    }
  } catch { output({ type: 'done', error: '无效脚本通信数据' }); process.exit(1); }
});

async function execute(command) {
  const state = command.state;
  const assertions = [], logs = [], pending = new Set();
  const names = new AsyncLocalStorage();
  const original = Assertion.prototype.assert;
  Assertion.prototype.assert = function (...args) {
    try { const result = original.apply(this, args); assertions.push({ name: names.getStore() || '脚本断言', passed: true }); return result; }
    catch (error) { assertions.push({ name: names.getStore() || '脚本断言', passed: false, message: String(error.message).slice(0, 2048) }); throw error; }
  };
  const unsupported = (object, label) => new Proxy(object, { get(target, key, receiver) {
    if (typeof key === 'symbol' || key === 'then' || key in target) return Reflect.get(target, key, receiver);
    throw new Error(`首版不支持 ${label}.${String(key)}`);
  }});
  const resolve = (name) => {
    for (const scope of [state.locals, state.case_variables, state.environment, state.collection, state.globals]) {
      if (Object.hasOwn(scope, name)) return scope[name];
    }
    throw new Error(`变量未定义: ${name}`);
  };
  const replaceIn = (value) => String(value).replace(/\{\{\s*([^{}]+?)\s*\}\}/g, (_, key) => resolve(key));
  const scope = (values, label, lookup = false) => unsupported({
    get: (key) => { if (lookup) { try { return resolve(String(key)); } catch { return undefined; } } return values[key]; },
    set: (key, value) => { values[String(key)] = typeof value === 'string' ? value : JSON.stringify(value); },
    has: (key) => lookup ? [state.locals, state.case_variables, state.environment, state.collection, state.globals].some(x => Object.hasOwn(x, key)) : Object.hasOwn(values, key),
    unset: (key) => { delete values[key]; },
    clear: () => { for (const key of Object.keys(values)) delete values[key]; },
    toObject: () => lookup ? Object.assign({}, state.globals, state.collection, state.environment, state.case_variables, state.locals) : { ...values },
    replaceIn,
  }, label);
  function headers(values) {
    return unsupported({
      get: (key) => Object.entries(values).find(([name]) => name.toLowerCase() === String(key).toLowerCase())?.[1],
      has: (key) => Object.keys(values).some(name => name.toLowerCase() === String(key).toLowerCase()),
      upsert: ({ key, value }) => { for (const k of Object.keys(values)) if (k.toLowerCase() === String(key).toLowerCase()) delete values[k]; values[key] = String(value); },
      add: ({ key, value }) => { values[key] = String(value); },
      remove: (key) => { for (const k of Object.keys(values)) if (k.toLowerCase() === String(key).toLowerCase()) delete values[k]; },
      toObject: () => ({ ...values }),
      each: (fn) => { for (const [key, value] of Object.entries(values)) fn({ key, value }); },
    }, 'headers');
  }
  function response(value) {
    const hs = Object.fromEntries(Object.entries(value.headers || {}).map(([k, v]) => [k, Array.isArray(v) ? v.join(', ') : v]));
    const obj = { code: value.status, status: STATUS_CODES[value.status] || '', responseTime: value.duration_ms,
      json: () => JSON.parse(value.body), text: () => value.body, headers: headers(hs),
      to: { have: { status: (code) => expect(value.status).to.equal(code), header: (key, expected) => {
        const found = Object.entries(hs).find(([k]) => k.toLowerCase() === key.toLowerCase()); expect(found, `header ${key}`).to.exist;
        if (expected !== undefined) expect(found?.[1]).to.equal(expected);
      }, jsonBody: (expected) => { const body = JSON.parse(value.body); if (expected !== undefined) expect(body).to.deep.equal(expected); else expect(body).to.not.equal(undefined); } }, be: {} },
    };
    Object.defineProperty(obj.to.be, 'json', { get: () => { expect(() => JSON.parse(value.body)).not.to.throw(); return true; } });
    return unsupported(obj, 'pm.response');
  }
  const request = structuredClone(command.request);
  request.headers ||= {};
  const requestView = { ...request, headers: headers(request.headers), body: { mode: 'raw', raw: request.body || '', update(value) { this.raw = typeof value === 'string' ? value : value.raw; } } };
  let sequence = 0;
  const sendRequest = (definition, callback) => {
    const id = String(++sequence);
    const req = typeof definition === 'string' ? { url: definition } : definition;
    const headerMap = Array.isArray(req.header) ? Object.fromEntries(req.header.filter(h => !h.disabled).map(h => [h.key, h.value])) : (req.headers || req.header || {});
    let body = typeof req.body === 'string' ? req.body : req.body?.raw || '';
    let bodyType = req.body_type || 'text';
    if (req.body?.mode === 'urlencoded') { body = new URLSearchParams((req.body.urlencoded || []).filter(x => !x.disabled).map(x => [x.key, x.value])).toString(); bodyType = 'form'; }
    const normalized = { method: req.method || 'GET', url: replaceIn(typeof req.url === 'string' ? req.url : req.url?.raw || ''),
      headers: Object.fromEntries(Object.entries(headerMap).map(([k, v]) => [k, replaceIn(v)])), query: req.query || {}, body: bodyType === 'form' ? body : replaceIn(body), body_type: bodyType,
      auth: req.auth || { type: 'none' }, timeout_ms: command.timeout_ms };
    const promise = new Promise((resolvePromise, reject) => {
      replies.set(id, value => {
        if (value.error) { const error = new Error(value.error); if (callback) { try { callback(error, null); resolvePromise(); } catch (e) { reject(e); } } else reject(error); }
        else { const reply = response(value); try { if (callback) callback(null, reply); resolvePromise(reply); } catch (e) { reject(e); } }
      });
      output({ type: 'http', id, request: normalized });
    });
    pending.add(promise); promise.catch(() => {});
    return callback ? undefined : promise;
  };
  const tests = [];
  const pm = unsupported({
    variables: scope(state.locals, 'pm.variables', true), environment: scope(state.environment, 'pm.environment'),
    collectionVariables: scope(state.collection, 'pm.collectionVariables'), globals: scope(state.globals, 'pm.globals'),
    request: unsupported(requestView, 'pm.request'), response: response(command.response || { status: 0, body: '', headers: {} }),
    info: Object.freeze({ requestName: command.name, eventName: command.phase, iteration: 0, iterationCount: 1 }),
    expect, sendRequest,
    test: (name, fn) => {
      const task = names.run(String(name), async () => {
        const before = assertions.length;
        try { if (fn.length) await new Promise((resolveTest, reject) => fn(error => error ? reject(error) : resolveTest())); else await fn(); }
        catch (error) { if (assertions.length === before || assertions.at(-1)?.passed) assertions.push({ name: String(name), passed: false, message: String(error.message).slice(0, 2048) }); }
      });
      tests.push(task); return pm;
    },
  }, 'pm');
  const consoleAPI = Object.fromEntries(['log', 'info', 'warn', 'error', 'debug'].map(level => [level, (...args) => {
    if (logs.length >= 200) return;
    const text = args.map(x => { try { return typeof x === 'string' ? x : JSON.stringify(x); } catch { return '[unserializable]'; } }).join(' ').slice(0, 2048);
    if (logs.reduce((n, s) => n + s.length, 0) + text.length <= 32768) logs.push(`[${level}] ${text}`);
  }]));
  let error;
  try {
    const context = vm.createContext({ pm, console: Object.freeze(consoleAPI), URLSearchParams, TextEncoder, TextDecoder });
    const script = new vm.Script(`(async () => {\n${command.script}\n})()`, { filename: 'user-script.js' });
    const deadline = new Promise((_, reject) => setTimeout(() => reject(new Error('JS脚本超时')), command.timeout_ms));
    const run = async () => {
      await script.runInContext(context, { timeout: command.timeout_ms });
      for (let rounds = 0; rounds < 100; rounds++) {
        const count = pending.size + tests.length;
        await Promise.all([...pending, ...tests]);
        if (count === pending.size + tests.length) return;
      }
      throw new Error('脚本异步操作过多');
    };
    await Promise.race([run(), deadline]);
  } catch (e) { error = String(e.message).slice(0, 2048); }
  output({ type: 'done', state, request: { ...request, method: String(requestView.method), url: String(requestView.url), body: String(requestView.body.raw), headers: request.headers }, assertions, logs, error });
  process.exit(error ? 1 : 0);
}

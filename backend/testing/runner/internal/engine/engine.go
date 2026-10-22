package engine

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	p "github.com/mi991203/TestMoon/libs/go/platform"
	m "github.com/mi991203/TestMoon/libs/go/testing/domain"
	"io"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type State struct {
	Locals        map[string]string `json:"locals"`
	Environment   map[string]string `json:"environment"`
	Collection    map[string]string `json:"collection"`
	Globals       map[string]string `json:"globals"`
	CaseVariables map[string]string `json:"case_variables"`
}
type Transport func(context.Context, string, m.HTTPRequest) m.Response
type Engine struct {
	Transport  Transport
	ScriptPath string
	Event      func(string, m.StepResult)
	Secrets    []string
}

func copyMap(in map[string]string) map[string]string {
	o := map[string]string{}
	for k, v := range in {
		o[k] = v
	}
	return o
}
func (s *State) expand(value string) (string, error) {
	var err error
	re := regexp.MustCompile(`\{\{\s*([^{}]+?)\s*\}\}`)
	for depth := 0; depth < 10 && re.MatchString(value); depth++ {
		value = re.ReplaceAllStringFunc(value, func(full string) string {
			key := strings.TrimSpace(re.FindStringSubmatch(full)[1])
			for _, scope := range []map[string]string{s.Locals, s.CaseVariables, s.Environment, s.Collection, s.Globals} {
				if v, ok := scope[key]; ok {
					return v
				}
			}
			err = fmt.Errorf("变量未定义: %s", key)
			return full
		})
		if err != nil {
			return value, err
		}
	}
	if re.MatchString(value) {
		return value, errors.New("变量存在循环引用")
	}
	return value, nil
}
func (s *State) request(r m.HTTPRequest) (m.HTTPRequest, error) {
	var e error
	r.URL, e = s.expand(r.URL)
	if e != nil {
		return r, e
	}
	if r.BodyType == "form" {
		form, err := url.ParseQuery(r.Body)
		if err != nil {
			return r, errors.New("表单编码无效")
		}
		for key, values := range form {
			for i, value := range values {
				values[i], e = s.expand(value)
				if e != nil {
					return r, e
				}
			}
			form[key] = values
		}
		r.Body = form.Encode()
	} else {
		r.Body, e = s.expand(r.Body)
	}
	if e != nil {
		return r, e
	}
	r.Headers = copyMap(r.Headers)
	r.Query = copyMap(r.Query)
	for _, values := range []map[string]string{r.Headers, r.Query} {
		for k, v := range values {
			values[k], e = s.expand(v)
			if e != nil {
				return r, e
			}
		}
	}
	for _, field := range []*string{&r.Auth.Username, &r.Auth.Password, &r.Auth.Token, &r.Auth.Key, &r.Auth.Value} {
		*field, e = s.expand(*field)
		if e != nil {
			return r, e
		}
	}
	return r, nil
}
func Pointer(value any, path string) (any, bool) {
	if path == "" {
		return value, true
	}
	if !strings.HasPrefix(path, "/") {
		return nil, false
	}
	for _, part := range strings.Split(path[1:], "/") {
		key := strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
		switch current := value.(type) {
		case map[string]any:
			var ok bool
			value, ok = current[key]
			if !ok {
				return nil, false
			}
		case []any:
			index, e := strconv.Atoi(key)
			if e != nil || index < 0 || index >= len(current) {
				return nil, false
			}
			value = current[index]
		default:
			return nil, false
		}
	}
	return value, true
}
func extract(response m.Response, source, path string) (any, bool) {
	switch source {
	case "status":
		return response.Status, true
	case "duration":
		return response.DurationMS, true
	case "text":
		return response.Body, true
	case "header":
		for k, v := range response.Headers {
			if strings.EqualFold(k, path) {
				return strings.Join(v, ", "), true
			}
		}
		return nil, false
	case "json":
		var v any
		if json.Unmarshal([]byte(response.Body), &v) != nil {
			return nil, false
		}
		return Pointer(v, path)
	}
	return nil, false
}
func assert(got any, exists bool, op string, want any) bool {
	if op == "exists" {
		return exists
	}
	if !exists {
		return false
	}
	a, _ := json.Marshal(got)
	b, _ := json.Marshal(want)
	switch op {
	case "eq":
		return string(a) == string(b)
	case "neq":
		return string(a) != string(b)
	case "contains":
		return strings.Contains(fmt.Sprint(got), fmt.Sprint(want))
	case "regex":
		if len(fmt.Sprint(want)) > 512 {
			return false
		}
		re, e := regexp.Compile(fmt.Sprint(want))
		return e == nil && re.MatchString(fmt.Sprint(got))
	case "lt", "lte", "gt", "gte":
		x, e := strconv.ParseFloat(fmt.Sprint(got), 64)
		y, f := strconv.ParseFloat(fmt.Sprint(want), 64)
		if e != nil || f != nil {
			return false
		}
		switch op {
		case "lt":
			return x < y
		case "lte":
			return x <= y
		case "gt":
			return x > y
		case "gte":
			return x >= y
		}
	}
	return false
}
func Aggregate(states []string) string {
	result := "passed"
	rank := map[string]int{"passed": 0, "unverified": 1, "skipped": 1, "failed": 2, "error": 3, "timed_out": 4, "canceled": 5}
	for _, s := range states {
		if rank[s] > rank[result] {
			result = s
		}
	}
	return result
}
func (e *Engine) Run(ctx context.Context, snapshot m.Snapshot) m.Result {
	start := time.Now()
	result := m.Result{Status: "passed"}
	for _, v := range snapshot.Environment.Variables {
		if v.Secret && v.Value != "" {
			e.Secrets = append(e.Secrets, v.Value)
		}
	}
	for _, item := range snapshot.Items {
		state := State{Locals: map[string]string{}, Environment: map[string]string{}, Collection: copyMap(item.CollectionVariables), Globals: map[string]string{}, CaseVariables: map[string]string{}}
		for k, v := range snapshot.Environment.Variables {
			state.Environment[k] = v.Value
		}
		ir := m.ItemResult{ID: item.ID, Name: item.Name}
		stopped := false
		statuses := []string{}
		for _, step := range item.Steps {
			var sr m.StepResult
			if stopped || ctx.Err() != nil {
				sr = m.StepResult{ID: step.ID, Name: step.Name, Status: "skipped"}
			} else {
				state.CaseVariables = copyMap(step.Case.Variables)
				if len(state.Collection) == 0 {
					state.Collection = copyMap(step.Case.CollectionVariables)
				}
				sr = e.step(ctx, item.ID, step, &state)
			}
			safe := p.Sanitize(sr, e.Secrets)
			b, err := json.Marshal(safe)
			var stored m.StepResult
			if err == nil {
				err = json.Unmarshal(b, &stored)
			}
			if err != nil {
				// Return a reportable failure without falling back to unsanitized data.
				stored = m.StepResult{ID: step.ID, Name: step.Name, Status: "error", Error: "执行结果脱敏转换失败"}
			}
			statuses = append(statuses, stored.Status)
			if stored.Status != "passed" && stored.Status != "unverified" && !step.ContinueOnFailure {
				stopped = true
			}
			trimStep(&stored, 16384)
			ir.Steps = append(ir.Steps, stored)
			if e.Event != nil {
				e.Event(item.ID, stored)
			}
		}
		ir.Status = Aggregate(statuses)
		result.Items = append(result.Items, ir)
	}
	states := []string{}
	for _, v := range result.Items {
		states = append(states, v.Status)
	}
	result.Status = Aggregate(states)
	if ctx.Err() != nil {
		result.Status = "timed_out"
		result.Error = "执行超时或已取消"
	}
	result.DurationMS = time.Since(start).Milliseconds()
	Compact(&result)
	return result
}
func (e *Engine) step(ctx context.Context, itemID string, step m.CompiledStep, state *State) m.StepResult {
	start := time.Now()
	c := step.Case
	result := m.StepResult{ID: step.ID, Name: step.Name, Status: "unverified", Request: c.Request}
	recordError := func(err error) {
		if err != nil {
			result.Status = "error"
			result.Error = err.Error()
		}
	}
	for _, op := range c.Before {
		if err := e.script(ctx, itemID, step.Name, "prerequest", op, state, &result); err != nil {
			recordError(err)
			result.DurationMS = time.Since(start).Milliseconds()
			return result
		}
		for _, a := range result.Assertions {
			if !a.Passed {
				result.Status = "failed"
				result.DurationMS = time.Since(start).Milliseconds()
				return result
			}
		}
	}
	req, err := state.request(result.Request)
	if err != nil {
		recordError(err)
		return result
	}
	result.Request = req
	result.Response = e.Transport(ctx, itemID, req)
	if result.Response.Error != "" {
		recordError(errors.New(result.Response.Error))
	} else {
		for _, x := range c.Extractions {
			value, ok := extract(result.Response, x.Source, x.Path)
			if !ok {
				recordError(fmt.Errorf("无法提取变量 %s", x.Variable))
				continue
			}
			v := fmt.Sprint(value)
			if _, ok := value.(string); !ok {
				b, _ := json.Marshal(value)
				v = string(b)
			}
			state.Locals[x.Variable] = v
			if x.Secret || p.Sensitive(x.Variable) {
				e.Secrets = append(e.Secrets, v)
			}
		}
		for _, a := range c.Assertions {
			value, exists := extract(result.Response, a.Source, a.Path)
			passed := assert(value, exists, a.Operator, a.Expected)
			message := ""
			if !passed {
				message = "断言未满足"
			}
			result.Assertions = append(result.Assertions, m.AssertionResult{Name: a.Source + ":" + a.Path + " " + a.Operator, Passed: passed, Message: message})
		}
	}
	for _, op := range c.After {
		recordError(e.script(ctx, itemID, step.Name, "test", op, state, &result))
	}
	if result.Status != "error" {
		if len(result.Assertions) > 0 {
			result.Status = "passed"
		}
		for _, a := range result.Assertions {
			if !a.Passed {
				result.Status = "failed"
			}
		}
	}
	for name, value := range state.Environment {
		if p.Sensitive(name) {
			e.Secrets = append(e.Secrets, value)
		}
	}
	for name, value := range state.Locals {
		if p.Sensitive(name) {
			e.Secrets = append(e.Secrets, value)
		}
	}
	for _, v := range []string{req.Auth.Password, req.Auth.Token, req.Auth.Value} {
		if v != "" {
			e.Secrets = append(e.Secrets, v)
		}
	}
	result.DurationMS = time.Since(start).Milliseconds()
	return result
}
func (e *Engine) script(ctx context.Context, itemID, name, phase string, op m.Operation, state *State, result *m.StepResult) error {
	if op.Type != "javascript" || op.Version != 1 {
		return errors.New("不支持的前后置操作")
	}
	timeout := 10 * time.Second
	scriptCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	path := e.ScriptPath
	if path == "" {
		path = "/app/worker/worker.mjs"
	}
	cmd := exec.CommandContext(scriptCtx, "node", path)
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "NODE_ENV=production"}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = io.Discard
	if err = cmd.Start(); err != nil {
		return errors.New("Node脚本进程启动失败")
	}
	encoder := json.NewEncoder(stdin)
	if err = encoder.Encode(map[string]any{"script": op.Script, "state": state, "request": result.Request, "response": result.Response, "phase": phase, "name": name, "timeout_ms": int(timeout.Milliseconds())}); err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return err
	}
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 65536), 4<<20)
	done := false
	var scriptError string
	for scanner.Scan() {
		var frame struct {
			Type       string              `json:"type"`
			ID         string              `json:"id"`
			Request    m.HTTPRequest       `json:"request"`
			State      State               `json:"state"`
			Assertions []m.AssertionResult `json:"assertions"`
			Logs       []string            `json:"logs"`
			Error      string              `json:"error"`
		}
		if json.Unmarshal(scanner.Bytes(), &frame) != nil {
			scriptError = "无效脚本输出"
			break
		}
		switch frame.Type {
		case "http":
			req, er := state.request(frame.Request)
			response := m.Response{}
			if er != nil {
				response.Error = er.Error()
			} else {
				response = e.Transport(scriptCtx, itemID, req)
			}
			if er := encoder.Encode(map[string]any{"type": "http_result", "id": frame.ID, "response": response}); er != nil {
				scriptError = "脚本通信失败"
			}
		case "done":
			*state = frame.State
			if phase == "prerequest" {
				result.Request = frame.Request
			}
			result.Assertions = append(result.Assertions, frame.Assertions...)
			result.Logs = append(result.Logs, frame.Logs...)
			scriptError = frame.Error
			done = true
		}
		if done || scriptError != "" {
			break
		}
	}
	_ = stdin.Close()
	if !done {
		_ = cmd.Process.Kill()
	}
	waitErr := cmd.Wait()
	if scriptCtx.Err() != nil {
		return errors.New("JS脚本超时")
	}
	if scriptError != "" {
		return errors.New(scriptError)
	}
	if !done || waitErr != nil {
		return errors.New("JS脚本执行失败")
	}
	return nil
}
func trimStep(step *m.StepResult, limit int) {
	if len(step.Response.Body) > limit {
		step.Response.Body = step.Response.Body[:limit]
		step.Response.Truncated = true
	}
	if len(step.Request.Body) > limit {
		step.Request.Body = step.Request.Body[:limit] + " [truncated]"
	}
	if len(step.Error) > 2048 {
		step.Error = step.Error[:2048]
	}
	for i := range step.Assertions {
		if len(step.Assertions[i].Message) > 2048 {
			step.Assertions[i].Message = step.Assertions[i].Message[:2048]
		}
	}
}
func Compact(result *m.Result) {
	for _, limit := range []int{16384, 4096, 1024, 0} {
		for i := range result.Items {
			for j := range result.Items[i].Steps {
				trimStep(&result.Items[i].Steps[j], limit)
				if limit < 4096 {
					result.Items[i].Steps[j].Logs = nil
				}
			}
		}
		b, _ := json.Marshal(result)
		if len(b) < 7<<20 {
			return
		}
	}
}

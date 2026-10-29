---
name: apimoon
description: "根据需求文档和项目代码生成 HTTP 接口测试，使用 APIMoon CLI 将接口、用例、场景和任务同步到 TestMoon，通过本机私有 Runner 执行并分析报告。当用户要求生成接口测试、运行 TestMoon 测试或分析平台报告时使用。"
---

# APIMoon

使用本机 `apimoon` 命令操作 TestMoon。需求分析和测试设计由 Agent 完成，测试执行与报告由 TestMoon 提供。按用户要求完成生成、同步、执行或报告分析；只要求生成计划时，交付计划即可。

## 检查调用环境

```bash
command -v apimoon
apimoon version
apimoon --help
```

使用具体命令前可通过 `apimoon plan sync --help`、`apimoon task run --help` 等逐级查看其参数和示例；帮助输出纯文本，无需登录或配置。参数按命令定义，不要将其他命令的选项传入。

仅生成本地计划时可离线读取 `plan schema`。需要访问平台时再检查鉴权与作用域，提交执行前检查本机 Runner 在线：

```bash
apimoon auth status
apimoon context show
apimoon runner list
```

- CLI 默认读取 `~/.api_moon/config.yaml`；可用 `--config FILE` 选择其他配置。查看配置使用 `config show`，它会隐藏登录凭证。
- 使用用户选定的组织、项目和环境。本机服务选择属于当前项目的在线私有 Runner；`localhost` 指 Runner 所在机器。
- 若 CLI、登录或本机 Runner 未就绪，读取 [references/setup.md](references/setup.md)。当前 `runner list` 查询平台状态，本机 Runner 使用独立程序启动。
- 普通使用者从 GitHub Releases 下载预编译 APIMoon、Runner，无需本机 Go。Worker 通过发布包的 `build-worker.sh` 下载对应源码并在本机 Docker 中构建；本机组件尚未就绪时按接入文档准备。
- 密码仅通过 `auth login --password-stdin` 提供；不要把密码、会话 Cookie、CSRF 或 Runner 凭证写入计划、日志或回复。

## 从需求和代码生成计划

先确认被测接口的路由、方法、请求类型、返回值、业务约束和部署地址。沿代码查找依赖数据，例如创建班级依赖已存在的学校；区分代码确认的行为与需求期望，报告二者的偏差。

读取当前 CLI 的计划协议：

```bash
apimoon plan schema
```

stdout 是 JSON 信封，Schema 位于 `data`。结构和示例见 [references/test-plans.md](references/test-plans.md)。

- 为同一测试意图使用稳定 `plan_id` 和条目 `key`，写入可复用的本地 JSON 文件。
- 通过 `ref:<key>` 引用计划内资源；使用平台 ID 引用已有资源。更新提供准确的 ID、revision 和完整 document。
- 根据实际约束覆盖正常、参数边界和错误场景；每个执行用例至少有一条有效断言。
- 有数据依赖的请求放入同一个场景，先创建前置数据、提取 ID，再执行测试、回读和清理。独立任务项不共享变量。
- 只清理本次测试创建的数据，并用回读验证清理结果。当前执行器失败后可能跳过后续步骤；使用 `continue_on_failure` 时考虑依赖是否仍成立，不能据此宣称清理必定执行。
- 静态断言的 `expected` 不做变量替换。比较响应值与动态提取变量时，使用后置 JavaScript，例如 `pm.expect(pm.response.json().school_id).to.equal(pm.variables.get('school_id'))`。

## 校验和同步

```bash
apimoon plan validate --file test-plan.json
apimoon plan sync --file test-plan.json
```

校验只读取平台资源，不请求被测系统或执行脚本。校验通过后，按用户授权同步。同步成功后从 `data.state.entries` 读取资源 ID、revision，从 `data.urls` 读取平台链接。

保留配置文件同级的 `state/` 目录，并保持计划标识稳定。同步按条目逐个保存，已经成功的条目不会随后续失败而回滚；删掉计划条目也不会自动删除平台资源。

出现版本冲突或写入结果不明时，读取 [references/recovery.md](references/recovery.md)，不要通过新建 plan_id 或删状态文件来绕过问题。

## 执行并读取报告

先确认任务保存的环境和 Runner。已有任务执行不接受环境或 Runner 临时覆盖，配置默认值的变化也不会修改已保存任务。

```bash
apimoon task get --id <任务ID>
apimoon task run --id <任务ID> --idempotency-key <本次执行唯一键> --wait --wait-timeout 5m
apimoon run get --id <执行ID>
```

提交前记录本次执行的幂等键；提交结果不明时用原键重试。用户发起一次新的执行时使用新键。

业务命令 stdout 为单份 JSON，进度在 stderr；帮助和 Shell 补全输出纯文本。即使命令非零退出，也要读取 JSON 中可能存在的 `data.run.id`、`data.idempotency_key`、同步进度和 `error`。

- 退出码 5：等待超时，使用已有执行 ID 继续 `run wait` 或 `run get`；超时或 Ctrl+C 不会自动取消平台执行。
- 退出码 6：执行已结束但没有通过，读取报告分析，不要直接当作 CLI 故障重试。
- `run get` 查询成功返回 0，即使报告是失败状态。以 `data.run.state` 和步骤结果判断业务结论。
- 没有断言的 `unverified` 不能计为通过。

分析报告的 `run`、`result.items[].steps[]`、`summary` 和 `failures`。结合请求、响应、断言和脚本日志，区分被测接口问题、用例设计问题及 Runner/环境问题。执行成功也要核对清理步骤。

回复包含计划文件、平台任务/报告链接、运行结果、关键断言和失败原因。准确说明哪些是静态校验、哪些已经真实执行；尚未验证的运行行为标为待确认。

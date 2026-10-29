# APIMoon

供 AI Agent 调用 TestMoon 的 Go CLI。Agent 读取需求和项目代码、生成测试计划；APIMoon 负责校验、同步资源、提交执行和读取结构化报告。测试在平台 Runner 执行，CLI 不调用模型。

## 构建和配置

在仓库根目录执行（Go 版本同根 `go.mod`）：

```bash
make build-apimoon
export PATH="$PWD/.local/bin:$PATH"
apimoon config init --base-url http://localhost:18080
# 从标准输入提供密码，避免写入命令参数或历史。
apimoon auth login --username <用户名> --password-stdin
apimoon org list
apimoon config set context.org_id <组织ID>
apimoon project list
apimoon config set context.project_id <项目ID>
apimoon runner list
apimoon config set context.runner_id <RunnerID>
# 按需设置默认环境：
apimoon config set context.environment_id <环境ID>
apimoon context show
```

`auth login` 读取 stdin 至 EOF；交互输入后按 Ctrl+D，也可由 Agent 通过子进程 stdin 传入。使用现有用户登录会话（Cookie + CSRF）；密码不落盘。需要预先在平台创建组织、项目和 Runner，并为登录用户授予相应项目权限。

默认配置为 `~/.api_moon/config.yaml`，可通过 `--config FILE` 指定；[配置示例](examples/config.yaml)。配置与同步状态原子写入，文件权限为 `0600`，新建目录为 `0700`。`config show` 隐去会话凭证；会话到期后重新登录，改变 `base_url` 会清除鉴权信息。切换组织清除项目、环境、Runner；切换项目清除环境、Runner。

`--org`、`--project` 可临时选择作用域，`--environment`、`--runner` 用于创建资源的默认值。已有任务执行时使用任务保存的环境和 Runner；修改配置不会改写已保存任务，需显式更新其完整 document。

## Agent 调用流程

1. 用 `plan schema` 读取输入协议，结合需求和代码生成计划；[School 完整计划](examples/school.plan.json)包含接口、接口用例、嵌套场景、环境和任务。
2. `plan validate` 预校验协议、编排树、引用、已有资源版本及当前用户的写入/执行权限。
3. `plan sync` 按依赖顺序保存资源，读取输出的任务 ID。
4. `task run --wait` 执行，或不加 `--wait` 后使用 `run wait/get` 获取报告。
5. Agent 根据报告中的请求、响应、断言、日志及失败摘要分析结果。

```bash
apimoon plan schema
apimoon plan validate --file tools/apimoon/examples/school.plan.json
apimoon plan sync --file tools/apimoon/examples/school.plan.json
apimoon task run --id <任务ID> --idempotency-key <本次执行的唯一键> --wait
apimoon run get --id <执行ID>
apimoon run list
apimoon run cancel --id <执行ID>
```

运行示例前启动 [School fixture](../../fixtures/school_management/README.md)，选择能访问它的本机私有 Runner。环境使用 `http://127.0.0.1:8096`，远程执行时需要修改地址。示例按顺序创建学校、创建班级、查询、删除学校并确认班级返回 404。平台当前失败后可能跳过后续步骤，因此该示例不保证失败路径执行清理；需要在测试环境检查残留数据。

预校验不会请求被测系统或执行 JavaScript，不能保证动态变量、脚本和断言的运行结果。任务保存时平台还会构建快照并进行自己的权限及依赖校验。

## 计划协议与重复同步

计划使用 `version: 1`、稳定 `plan_id` 和 `entries`。每个条目包含稳定 `key`、`kind`、`name`、`document`，更新可指定 `id` 与正整数 `revision`。字段采用平台的 `snake_case`；[JSON Schema](examples/testplan.schema.json)由同一套 protobuf 描述符生成，最新协议以 `plan schema` 输出为准。

```json
{
  "version": 1,
  "plan_id": "health-check-v1",
  "entries": [
    {
      "key": "health",
      "kind": "api",
      "name": "Health",
      "document": {
        "protocol": "http",
        "request": {"method": "GET", "url": "https://example.com/healthz"}
      }
    },
    {
      "key": "health-case",
      "kind": "case",
      "name": "Health returns 200",
      "document": {
        "api_id": "ref:health",
        "assertions": [{"source": "status", "operator": "eq", "expected": 200}]
      }
    }
  ]
}
```

`ref:<key>` 只在资源 ID 引用字段替换，不改写请求体、脚本或其他普通文本。依赖可以先后任意排列；不允许不存在的引用、循环或类型不符。JSON Pointer 使用 `/id`，不是 `$.id`。请求体是字符串，headers/query 是字符串映射；前后置操作使用 `type: javascript`、`version: 1`。

同步状态位于配置文件同级 `state/`，按平台地址、登录用户、组织、项目及 plan_id 隔离；记录 key 到平台 ID、revision、内容摘要及写入状态。相同内容跳过保存，变更按已知 revision 更新；检测到平台人工编辑会报冲突。不会按名称查找或自动删除已从计划移除的资源。保留 plan_id、key 和状态目录才能重复同步；本地状态不提供跨机器恢复或跨服务事务。

```bash
apimoon plan status --file plan.json
# 仅在写入响应丢失等结果不明时，人工核对平台 ID 后使用：
apimoon plan bind --file plan.json --key <条目key> --id <平台ID>
```

写入前先记为 pending；结果不明时停止，禁止盲目重试创建。`bind` 会读取目标资源并核对内容，匹配后才继续同步。使用写入时的计划和默认配置；内容不匹配会拒绝绑定。若进程异常退出遗留锁，确认没有同步进程后按错误中的路径移除锁文件。多个资源逐个保存，失败前已成功的资源保留。

## 单资源命令与输出

`api`、`case`、`scenario`、`task`、`environment`、`folder` 均支持 `list/get/create/update/delete`。使用 `--file FILE` 或 `--file -` 提供平台请求 JSON；`--id`、`--revision` 等参数覆盖输入文件同名字段。组织和项目由当前配置注入，输入中不允许指定其他作用域。更新必须提供 ID、revision、名称和完整 document；不提供局部合并语义。

```bash
apimoon api list --search school
apimoon case get --id <用例ID>
apimoon api create --file api.json
apimoon case update --id <用例ID> --revision <版本> --file case.json
apimoon task delete --id <任务ID>
apimoon auth status
apimoon auth logout
apimoon --help --output text
```

默认 stdout 为单份 JSON：`{"ok":true,"data":...}` 或 `{"ok":false,"error":{"code":...,"message":...}}`；部分失败仍会携带 `data`（同步进度或执行 ID）。进度写 stderr，`--output text` 使用易读缩进。报告查询成功的退出码为 0，即使报告状态为 failed；等待执行结束时只有 passed 返回 0，unverified 也返回 6。

| 退出码 | 含义 |
| --- | --- |
| 0 | 命令成功 / 执行通过 |
| 1 | 本地、网络或其他平台错误 |
| 2 | 参数、配置或校验错误 |
| 3 | 登录或权限错误 |
| 4 | 版本冲突、并发锁或写入结果不明 |
| 5 | 等待超时 |
| 6 | 执行结束但未通过 |

HTTP 请求默认超时 30 秒；等待默认 30 分钟、轮询 2 秒，可设置 `--wait-timeout 5m --poll-interval 2s`。等待超时或 Ctrl+C 不会自动取消平台执行。提交任务会返回 idempotency_key；提交响应丢失时使用原键重试，以避免重复执行。

## 包结构与验证

```text
cmd/main.go           进程入口
cli/                  命令路由、参数、鉴权与输出适配
config/               YAML 配置、作用域与配置校验
platform/             类型协议、HTTP 会话客户端和资源操作
validate/             资源文档与编排校验
plan/                 计划协议、引用、Schema、同步状态和恢复
execution/            运行轮询与报告摘要
clierror/             结构化错误
fileio/               私有文件原子写入
jsonutil/             严格 JSON 解码与对象操作
tests/                模拟平台的跨包功能测试
```

功能包直接放在 APIMoon 目录下，可被大仓其他模块导入复用。包依赖从命令层指向计划/执行，再指向平台客户端；配置、错误、JSON 和文件操作为基础能力。同步状态和摘要的包内测试留在所属包中。

```bash
go test -race ./tools/apimoon/...
make build-apimoon
```

当前范围不含模型接入、稳定性标注及自动筛选执行、平台集成令牌、跨机器同步和 Runner 执行语义修改。

# APIMoon

供 AI Agent 调用 TestMoon 的 Go CLI。Agent 读取需求和项目代码、生成测试计划；APIMoon 负责校验、同步资源、提交执行和读取结构化报告。测试在平台 Runner 执行，CLI 不调用模型。

## 从 GitHub Releases 安装

APIMoon 和 Runner 均由 GitHub Actions 编译并在 [GitHub Releases](https://github.com/mi991203/TestMoon/releases) 独立提供下载；Worker 从 Docker Hub 拉取 `flyingcode/testmoon-worker:0.1.0`。用户只需 Docker，无需下载源码或安装 Git、Go、Node。

支持 macOS、Linux、Windows 的 amd64、arm64。每个系统/架构下载同版本的两个文件和 `SHA256SUMS`，不使用压缩包：

- `apimoon_vX.Y.Z_<系统>_<架构>`
- `testmoon-runner_vX.Y.Z_<系统>_<架构>`

Windows 文件增加 `.exe` 后缀。Mac 苹果芯片选 `darwin_arm64`，Intel 选 `darwin_amd64`；Linux x86-64 选 `linux_amd64`；Windows 常见电脑选 `windows_amd64`，ARM64 电脑选 `windows_arm64`。

校验哈希后，macOS/Linux 用 `install -m 755` 安装到 `~/.local/bin/apimoon`、`~/.local/bin/testmoon-runner`；Windows 复制并重命名为 `apimoon.exe`、`testmoon-runner.exe`，将所在目录加入 PATH。Skill 已内嵌在 APIMoon 中。完整命令见 [安装及 Runner 启动说明](distribution/README.md)。

安装、升级保留 `~/.api_moon/` 的登录、作用域、同步状态和 Runner 状态；升级前先停止旧 Runner。示例版本号需替换为实际已发布版本。

## 通过 APIMoon 安装 Agent Skill

[APIMoon Skill](skills/apimoon/SKILL.md) 及引用文档直接内嵌在 CLI 中，安装后即使只保留 `apimoon` 可执行文件，也可离线安装 Skill，无需登录平台、源码目录、Node 或额外安装脚本。

```bash
# Codex，用户级安装，跨项目使用。
apimoon skill install --agent codex --global

# Trae 国际版 / 国内版，用户级安装。
apimoon skill install --agent trae --global
apimoon skill install --agent trae-cn --global

# 默认安装到当前业务项目；也可明确指定项目目录。
apimoon skill install --agent codex
apimoon skill install --agent trae-cn --project-dir /业务项目
```

| Agent | 当前项目默认目录 | 用户级目录（--global） |
| --- | --- | --- |
| `codex` | `.agents/skills/apimoon` | `~/.agents/skills/apimoon` |
| `trae` | `.trae/skills/apimoon` | `~/.trae/skills/apimoon` |
| `trae-cn` | `.trae/skills/apimoon` | `~/.trae-cn/skills/apimoon` |

命令形态参考 [Vercel Skills CLI](https://github.com/vercel-labs/skills/blob/main/docs/commands.md) 的 agent 与 project/global 选择方式；目录依据 [Codex 官方文档](https://learn.chatgpt.com/docs/build-skills)、[TRAE 国内版官方文档](https://docs.trae.cn/ide_skills) 及 [Skills Agent 兼容表](https://github.com/vercel-labs/skills/blob/main/docs/agents.md)。Trae 国际版和国内版的用户级目录不同，需要选择正确的 agent。

检查及更新使用：

```bash
apimoon skill status --agent codex --global
apimoon skill install --agent codex --global --force
```

相同内容重复安装返回 `unchanged`；状态查询返回 `not_installed/current/different`。已有内容不同时，安装返回退出码 4；确认更新后使用 `--force`，原目录备份到 skills 根目录之外，备份路径在 JSON 的 `data.backup` 返回，避免备份也被 Agent 当作 Skill 加载。符号链接目标需要先自行处理，不会被强制覆盖。

`--dir /自定义/skills` 可显式指定 skills 根目录，命令会追加 `apimoon`；它与 `--global`、`--project-dir` 不能同时使用。安装结果返回 agent、scope、path、status 和 bundled_version。

在 Codex 中可显式请求 `使用 $apimoon，根据需求和当前项目代码生成接口测试，使用 TestMoon 执行并分析报告。` Trae 可在技能面板确认已安装，再在对话中请求使用 APIMoon Skill；两者也可根据描述按需加载。安装 Skill 不会启动 Runner，Agent 执行环境仍需能找到 `apimoon` 并执行本机命令。

## 连接平台与本机 Runner

CLI 安装完成后配置连接：

```bash
apimoon config init --base-url http://localhost:18080
# 从标准输入提供密码，避免写入命令参数或历史。
apimoon auth login --username <用户名> --password-stdin
apimoon org list
apimoon config set context.org_id <组织ID>
apimoon project list
apimoon config set context.project_id <项目ID>
apimoon runner list
apimoon config set context.runner_id <RunnerID>
# 按需选择已有环境：
apimoon environment list
apimoon config set context.environment_id <环境ID>
apimoon context show
```

`auth login` 读取 stdin 至 EOF；Mac/Windows 隐藏密码输入的完整命令见 [配置 APIMoon](distribution/README.md#配置-apimoon)，Agent 也可通过子进程 stdin 传入。使用现有用户登录会话（Cookie + CSRF）；密码不落盘。需要预先在平台创建组织、项目和 Runner，并为登录用户授予相应项目权限。

本机 Docker 启动后，拉取 Worker：

```bash
docker pull flyingcode/testmoon-worker:0.1.0
```

Windows Docker Desktop 使用 Linux 容器模式；PowerShell 命令见 [使用说明](distribution/README.md)。APIMoon/Runner 的 Release 版本与 Worker 镜像版本分别管理，不能根据 CLI 版本推导镜像标签。

`testmoon-runner` 使用该镜像创建执行容器；启动前显式设置 `TESTMOON_WORKER_IMAGE=flyingcode/testmoon-worker:0.1.0`。当前程序默认镜像仍是 `testmoon/worker:0.1.0`，不能省略该配置。首次启动使用项目私有节点的一次性注册凭证，详细步骤见 [安装及 Runner 启动说明](distribution/README.md)。当前 APIMoon 的 Runner 命令仅支持 `runner list`，本机 Runner 的自动注册和进程管理尚未接入 CLI。

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
apimoon --help
apimoon plan sync --help
apimoon skill install --help
```

命令使用 Cobra 注册，每一级命令都支持 `--help` / `-h`，也可使用 `apimoon help plan sync`。帮助按当前命令展示说明、参数、默认值及示例，不读取平台配置或发起请求。`--config`、`--output`、`--org`、`--project` 为公共参数；其余选项属于对应命令，无关参数会报错。资源 ID 和 revision 仍可从 `--file` 提供，无需重复传递对应选项。

业务命令的 stdout 默认是单份 JSON：`{"ok":true,"data":...}` 或 `{"ok":false,"error":{"code":...,"message":...}}`；部分失败仍会携带 `data`（同步进度或执行 ID）。进度写 stderr，`--output text` 使用易读缩进。帮助和 Shell 补全输出纯文本。报告查询成功的退出码为 0，即使报告状态为 failed；等待执行结束时只有 passed 返回 0，unverified 也返回 6。

可生成 Bash、Zsh、Fish 和 PowerShell 的补全脚本，例如：

```bash
apimoon completion bash > apimoon.bash
apimoon completion zsh > _apimoon
```

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
cli/                  Cobra 命令树、按命令注册参数、鉴权与输出适配
config/               YAML 配置、作用域与配置校验
platform/             类型协议、HTTP 会话客户端和资源操作
validate/             资源文档与编排校验
plan/                 计划协议、引用、Schema、同步状态和恢复
execution/            运行轮询与报告摘要
clierror/             结构化错误
fileio/               私有文件原子写入
jsonutil/             严格 JSON 解码与对象操作
skills/               内嵌 Skill、Codex/Trae 安装与状态检查
tests/                模拟平台的跨包功能测试
```

功能包直接放在 APIMoon 目录下，可被大仓其他模块导入复用。包依赖从命令层指向计划/执行，再指向平台客户端；配置、错误、JSON 和文件操作为基础能力。同步状态和摘要的包内测试留在所属包中。

```bash
go test -race ./tools/apimoon/...
make build-apimoon
```

当前范围不含模型接入、稳定性标注及自动筛选执行、平台集成令牌、跨机器同步和 Runner 执行语义修改。

## 维护者构建和发布

本节面向维护者，需要 Git、Make 和根 `go.mod` 指定的 Go 工具链。开发时仍可执行：

```bash
make build-apimoon
make install-apimoon APIMOON_BIN_DIR=/自定义/bin
make install-apimoon-skill APIMOON_SKILLS_DIR=/自定义/skills
```

`make build-apimoon` 构建当前系统的程序，位于 `.local/bin/apimoon/apimoon`（Windows 为 `apimoon.exe`）。旧版位于 `.local/bin/apimoon` 的单个程序会自动迁入该目录；`make install-apimoon` 的用户级安装目录不变。

一次构建 macOS、Linux、Windows 的 amd64 和 arm64 程序：

```bash
make build-apimoon-all
```

产物位于 `.local/bin/apimoon/`：

```text
apimoon_darwin_amd64
apimoon_darwin_arm64
apimoon_linux_amd64
apimoon_linux_arm64
apimoon_windows_amd64.exe
apimoon_windows_arm64.exe
```

交叉编译设置 `CGO_ENABLED=0`，无需安装目标平台的 C 编译器。可指定构建目标或输出目录，例如 `make build-apimoon-all APIMOON_BUILD_TARGETS="linux/amd64 windows/amd64" APIMOON_BUILD_DIR=.local/custom-apimoon`。

本地验证发布产物（示例产物在 `.local/apimoon-release/v0.1.3`）：

```bash
make release-apimoon APIMOON_VERSION=v0.1.3
```

会交叉编译六个平台的 APIMoon 和 Runner，共生成 12 个独立可执行文件及覆盖全部文件的 `SHA256SUMS`。macOS/Linux 无文件后缀，Windows 为 `.exe`；所有平台均不打包、不压缩。两个程序分别命名为 `apimoon_vX.Y.Z_<系统>_<架构>`、`testmoon-runner_vX.Y.Z_<系统>_<架构>`。`APIMOON_RELEASE_TARGETS="windows/amd64 windows/arm64"` 可限制本地验证目标；正式工作流构建全部六个平台。

`.github/workflows/release-apimoon.yml` 在推送 `apimoon-vX.Y.Z` 标签时，先执行 CLI/Runner 测试，再构建并上传 GitHub Release 附件；也可通过 Actions 手动选择一个已存在的标签发布。测试版标签使用 `apimoon-vX.Y.Z-rc.1`，工作流会标为 prerelease。

先合入完整发布代码，再创建并推送标签：

```bash
git tag apimoon-v0.1.0
git push origin apimoon-v0.1.0
```

版本、提交和构建时间通过链接参数写入两个程序。Worker 由维护者使用对应源码单独构建并发布到 Docker Hub，用户只拉取已发布镜像；不要移动已发布标签或覆盖已有镜像版本。当前 GitHub Release 工作流不负责推送 Worker 镜像。

首次在 Docker Hub 创建公开仓库 `flyingcode/testmoon-worker` 后，在仓库根目录发布：

```bash
docker login --username flyingcode
# 首次创建构建器；已存在时改用 docker buildx use testmoon-builder。
docker buildx create --name testmoon-builder --driver docker-container --use
docker buildx inspect --bootstrap
docker buildx build --platform linux/amd64,linux/arm64 -f deploy/runner/worker.Dockerfile -t flyingcode/testmoon-worker:0.1.0 --push .
docker buildx imagetools inspect flyingcode/testmoon-worker:0.1.0
```

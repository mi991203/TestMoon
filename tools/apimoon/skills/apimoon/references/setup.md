# 安装、连接与本机 Runner

仅在 CLI 未安装、配置缺失或 Runner 未就绪时读取本文件。安装 Skill 不会自动安装可执行程序或授予命令执行权限。

## 安装预编译 APIMoon 和 Runner

普通使用者无需 Go 或 Node。从 [GitHub Releases](https://github.com/mi991203/TestMoon/releases) 选择与平台兼容的 `apimoon-vX.Y.Z`，下载本机系统/架构对应的 `apimoon_vX.Y.Z_<系统>_<架构>.tar.gz` 和 `SHA256SUMS`。支持 macOS/Linux 的 amd64、arm64。

用 `shasum -a 256 <安装包>` 或 `sha256sum <安装包>` 与校验文件中的对应行比较，解压后在安装包目录执行：

```bash
bash install.sh
export PATH="$HOME/.local/bin:$PATH"
apimoon version
testmoon-runner --version
```

默认安装两个程序到 `~/.local/bin`。自定义目录用 `bash install.sh /自定义/bin`，将其加入终端和 Agent 进程的 PATH。安装和升级保留 `~/.api_moon/` 的配置、同步记录与 Runner 状态；升级 Runner 前等待当前任务结束并停止旧进程。

APIMoon 已内嵌 Skill，安装包包含 `build-worker.sh`、`VERSION`、`COMMIT` 和 `REPOSITORY`。保留解压目录，用于构建 Worker。没有已发布附件时，说明发布尚未完成，不要把维护者示例版本视为可下载的正式版本。

维护者可从 TestMoon 源码运行 `make release-apimoon APIMOON_VERSION=vX.Y.Z` 打包；这一步需要 Go，仅用于维护者构建，不是使用者的安装步骤。

## 连接 TestMoon

```bash
apimoon config init --base-url https://testmoon.example.com
apimoon auth login --username <用户名> --password-stdin
apimoon org list
apimoon config set context.org_id <组织ID>
apimoon project list
apimoon config set context.project_id <项目ID>
apimoon runner list
apimoon config set context.runner_id <本机私有RunnerID>
apimoon environment list
apimoon config set context.environment_id <环境ID>
```

配置已存在时先用 `config show` 查看脱敏内容，保留用户已有的服务地址和作用域。更换组织会清除项目、环境及 Runner；更换项目会清除环境及 Runner。环境可由计划创建，也可引用已有环境。

`auth login` 从 stdin 读取至 EOF；交互输入后按 Ctrl+D，自动化调用通过子进程 stdin 提供用户给定的密码。登录会话过期后重新登录。

## 本地构建 Worker

前置条件：本机 Docker 已启动。安装包解压目录中执行：

```bash
bash build-worker.sh
```

脚本用 curl 下载本发布包 `COMMIT` 指定的源码，通过 Dockerfile 在容器内完成 Go/Node 构建，生成 `testmoon/worker:<发布版本，去掉v前缀>`。宿主机需要 Docker、curl、tar、Bash，无需 Go、Node、Make；构建会获取基础镜像和依赖。

已有对应源码或私有仓库时，可以传入干净的 Git checkout：

```bash
bash build-worker.sh /本地/TestMoon
```

此模式需要 Git，脚本会验证 HEAD 与安装包 `COMMIT` 一致且无本地修改。需要修改镜像标签时，在构建和启动 Runner 时使用相同的 `TESTMOON_WORKER_IMAGE`。

## 启动本机私有 Runner

用户在 TestMoon 所选项目的「Runner 管理」创建私有节点并获得一次性注册凭证后，在安装包解压目录的独立 Bash 终端中启动已安装的 Runner：

```bash
release_version="$(cat VERSION)"
export TESTMOON_URL='https://testmoon.example.com'
export TESTMOON_WORKER_IMAGE="testmoon/worker:${release_version#v}"
export TESTMOON_RUNNER_STATE="$HOME/.api_moon/runner/state.json"
export DOCKER_HOST="${DOCKER_HOST:-$(docker context inspect --format '{{.Endpoints.docker.Host}}')}"
docker info >/dev/null
docker image inspect "$TESTMOON_WORKER_IMAGE" >/dev/null

# 只在首次注册时输入；已有状态文件时省略这三行。
read -r -s -p 'Runner enrollment token: ' TESTMOON_ENROLLMENT_TOKEN
echo
export TESTMOON_ENROLLMENT_TOKEN
testmoon-runner
```

注册状态创建后，后续启动保持相同平台地址、状态文件和 Worker 镜像，无需再次输入注册凭证。运行凭证独立于 CLI 登录会话；不要展示状态文件内容。

本机开发平台可用 `http://localhost:18080` 并显式设置 `TESTMOON_ALLOW_HTTP=1`。被测服务的 `base_url` 必须从本机 Runner 可达。

在另一终端用 `apimoon runner list` 确认节点在线，保存其 ID 到 `context.runner_id`，再创建或显式更新任务的 Runner。保持 Runner 终端运行，Ctrl+C 停止它。

当前没有 `apimoon runner setup/start/stop` 或 `apimoon doctor` 命令；不要调用这些尚不存在的命令。

## 使用 CLI 安装 Codex / Trae Skill

安装 APIMoon 二进制后，可从任意目录执行；无需安装脚本、网络或 TestMoon 登录：

```bash
apimoon skill install --agent codex --global
apimoon skill install --agent trae --global
apimoon skill install --agent trae-cn --global
```

按所用 Agent 选择一条命令。用户级目录分别是 `~/.agents/skills/apimoon`、`~/.trae/skills/apimoon`、`~/.trae-cn/skills/apimoon`。Trae 国际版与国内版的用户目录不同。

不加 `--global` 时默认安装到当前业务项目；Codex 为 `.agents/skills/apimoon`，两种 Trae 均为 `.trae/skills/apimoon`。指定其他项目：

```bash
apimoon skill install --agent codex --project-dir /被测项目
```

`--dir /自定义/skills` 表示显式 skills 根目录，会追加 `apimoon`。`--dir`、`--project-dir`、`--global` 只能选择一种。

```bash
apimoon skill status --agent codex --global
apimoon skill install --agent codex --global --force
```

状态为 `not_installed/current/different`。相同内容重复安装返回 `unchanged`；已有内容不同则退出码 4，默认不覆盖。`--force` 更新时会备份原目录，输出 `data.backup`，备份放在 skills 根目录外，避免重复发现。符号链接不自动覆盖。

复制的内容包含 `SKILL.md`、`agents/` 和 `references/`。Codex 可通过 `$apimoon` 显式选择，Trae 在技能面板确认后可在对话中要求使用 APIMoon；Agent 也可根据描述按需加载。若界面尚未显示，按对应 Agent 的说明刷新或重启。

维护者可参考 [Codex 文档](https://learn.chatgpt.com/docs/build-skills)、[Trae 国内版文档](https://docs.trae.cn/ide_skills) 和 [Skills Agent 兼容表](https://github.com/vercel-labs/skills/blob/main/docs/agents.md) 更新路径映射。源码中的 `make install-apimoon-skill` 仅是调用同一 CLI 安装能力的开发便利入口。

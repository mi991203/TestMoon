# APIMoon 与本机 Runner

本发布包包含当前操作系统/架构的 `apimoon` 和 `testmoon-runner`，以及程序安装脚本和 Worker 构建脚本。Skill 已内嵌在 APIMoon 可执行文件中。使用者无需安装 Go 或 Node。构建及运行 Worker 需要已启动的 Docker。

## 安装预编译程序

从 GitHub Releases 下载本机对应的 `.tar.gz` 和 `SHA256SUMS`，用 `shasum -a 256 <安装包>` 或 `sha256sum <安装包>` 与校验文件中的对应行比较，再解压。

在解压目录执行：

```bash
./apimoon version
./testmoon-runner --version
bash install.sh
export PATH="$HOME/.local/bin:$PATH"
apimoon skill install --agent codex --global
```

默认安装两个程序到 `~/.local/bin`。自定义程序目录使用 `bash install.sh /自定义/bin`。需要在新终端或 Agent 中调用时，将程序目录加入对应 PATH。

## 安装 Codex / Trae Skill

```bash
apimoon skill install --agent codex --global
apimoon skill install --agent trae --global
apimoon skill install --agent trae-cn --global
```

按所用 Agent 选择一条命令：Codex 的用户级目录为 `~/.agents/skills/apimoon`，Trae 国际版为 `~/.trae/skills/apimoon`，国内版为 `~/.trae-cn/skills/apimoon`。

不加 `--global` 时安装到当前业务项目，Codex 使用 `.agents/skills/apimoon`，Trae 使用 `.trae/skills/apimoon`。可用 `--project-dir /业务项目` 选择项目，或用 `--dir /自定义/skills` 指定 skills 根目录。

使用 `apimoon skill status --agent codex --global` 检查状态。相同内容重复安装会成功；已有内容不同时默认报冲突，使用 `--force` 会先备份原 Skill，并在输出中返回备份路径。安装 Skill 只需 APIMoon 二进制，不依赖发布包目录或网络。

安装和升级保留 `~/.api_moon/` 中的登录、作用域、同步记录和 Runner 状态。升级 Runner 前等待当前任务结束，停止旧进程后安装。

## 本地构建 Worker

在解压目录执行：

```bash
bash build-worker.sh
```

脚本下载本发布包 `COMMIT` 指定的 TestMoon 源码，在 Docker 内完成 Go、Node 构建，生成 `testmoon/worker:<本发布版本，去掉v前缀>`。宿主机只需要 Docker、curl、tar、Bash；不需要 Go、Node、Make，也不需要维护者上传 Worker 镜像。

私有仓库或已经准备源码时，可提供干净且与 `COMMIT` 完全一致的 Git checkout：

```bash
bash build-worker.sh /本地/TestMoon
```

此模式需要 Git。源码版本不匹配或有本地修改时会拒绝构建，避免无意混用不同版本。构建会获取基础镜像和包依赖。

## 连接和执行

先配置 APIMoon，选择组织和项目，并在该项目「Runner 管理」创建私有节点、获得一次性注册凭证：

```bash
apimoon config init --base-url https://testmoon.example.com
apimoon auth login --username <用户名> --password-stdin
apimoon org list
apimoon config set context.org_id <组织ID>
apimoon project list
apimoon config set context.project_id <项目ID>
```

在解压目录的独立 Bash 终端中启动预编译 Runner：

```bash
release_version="$(cat VERSION)"
export TESTMOON_URL='https://testmoon.example.com'
export TESTMOON_WORKER_IMAGE="testmoon/worker:${release_version#v}"
export TESTMOON_RUNNER_STATE="$HOME/.api_moon/runner/state.json"
export DOCKER_HOST="${DOCKER_HOST:-$(docker context inspect --format '{{.Endpoints.docker.Host}}')}"
docker image inspect "$TESTMOON_WORKER_IMAGE" >/dev/null

# 只在首次注册时输入；已有状态文件时省略这三行。
read -r -s -p 'Runner enrollment token: ' TESTMOON_ENROLLMENT_TOKEN
echo
export TESTMOON_ENROLLMENT_TOKEN
testmoon-runner
```

本机开发平台可使用 `http://localhost:18080` 并显式设置 `TESTMOON_ALLOW_HTTP=1`。保持 Runner 终端运行；Ctrl+C 停止它，后续使用相同状态文件启动无需重新注册。

在另一终端确认在线并选择节点：

```bash
apimoon runner list
apimoon config set context.runner_id <本机私有RunnerID>
apimoon plan schema
```

让 Agent 使用已安装的 APIMoon Skill，根据被测项目需求和代码完成计划、同步、执行和报告分析。当前 APIMoon 的 `runner list` 查询平台节点，尚未提供本机 Runner 的注册、启动和停止命令。

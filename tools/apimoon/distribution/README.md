# 安装 APIMoon 与 Runner

APIMoon、Runner 均从 [GitHub Releases](https://github.com/mi991203/TestMoon/releases) 下载预编译程序，不需要用户构建。只有 Worker 镜像在本地 Docker 中构建；用户无需安装 Go、Node。

支持 macOS（darwin）、Linux、Windows 的 amd64、arm64。每个架构分别提供两个文件，全部直接下载，不提供压缩包：

| 系统 | APIMoon | Runner |
| --- | --- | --- |
| macOS / Linux | `apimoon_vX.Y.Z_<系统>_<架构>` | `testmoon-runner_vX.Y.Z_<系统>_<架构>` |
| Windows | `apimoon_vX.Y.Z_windows_<架构>.exe` | `testmoon-runner_vX.Y.Z_windows_<架构>.exe` |

同时下载 `SHA256SUMS`。下文以新格式版本 v0.1.3 为例，替换为实际已发布版本；程序、源码标签及 Worker 镜像版本保持一致。运行前用文件哈希与校验文件对应行比较。

## macOS / Linux 安装

Mac 苹果芯片选 darwin_arm64，Intel 选 darwin_amd64。Linux 将文件名中的 darwin 换为 linux。以下在下载目录执行：

```bash
shasum -a 256 apimoon_v0.1.3_darwin_arm64 testmoon-runner_v0.1.3_darwin_arm64
# 校验通过后安装；Linux 也可用 sha256sum 计算哈希。
mkdir -p "$HOME/.local/bin"
install -m 755 apimoon_v0.1.3_darwin_arm64 "$HOME/.local/bin/apimoon"
install -m 755 testmoon-runner_v0.1.3_darwin_arm64 "$HOME/.local/bin/testmoon-runner"
export PATH="$HOME/.local/bin:$PATH"
apimoon version
testmoon-runner --version
```

将 PATH 设置写入 shell 配置以便新终端及 Agent 使用：Mac 默认使用 `~/.zshrc`，Linux 按所用 shell 配置。升级前停止旧 Runner，保留 `~/.api_moon/`。

## Windows 安装

常见电脑选 amd64，ARM64 电脑选 arm64。在下载目录的 PowerShell 执行：

```powershell
Get-FileHash .\apimoon_v0.1.3_windows_amd64.exe -Algorithm SHA256
Get-FileHash .\testmoon-runner_v0.1.3_windows_amd64.exe -Algorithm SHA256
# 校验通过后安装。
$bin = "$HOME\.local\bin"
New-Item -ItemType Directory -Force $bin | Out-Null
Copy-Item .\apimoon_v0.1.3_windows_amd64.exe "$bin\apimoon.exe"
Copy-Item .\testmoon-runner_v0.1.3_windows_amd64.exe "$bin\testmoon-runner.exe"
[Environment]::SetEnvironmentVariable('Path', "$bin;" + [Environment]::GetEnvironmentVariable('Path', 'User'), 'User')
$env:Path = "$bin;$env:Path"
apimoon version
testmoon-runner --version
```

## 只构建 Worker

安装 Git，启动 Docker（Windows 使用 Docker Desktop 的 Linux 容器模式）。从同一发布标签获取源码，只执行 Worker 的 Docker 构建，不执行 Go、Make 或 Runner 构建。

macOS / Linux：

```bash
git clone --branch apimoon-v0.1.3 --depth 1 https://github.com/mi991203/TestMoon.git TestMoon-v0.1.3
docker build -f TestMoon-v0.1.3/deploy/runner/worker.Dockerfile -t testmoon/worker:0.1.3 TestMoon-v0.1.3
```

Windows PowerShell：

```powershell
git clone --branch apimoon-v0.1.3 --depth 1 https://github.com/mi991203/TestMoon.git TestMoon-v0.1.3
docker build -f .\TestMoon-v0.1.3\deploy\runner\worker.Dockerfile -t testmoon/worker:0.1.3 .\TestMoon-v0.1.3
```

## 启动已下载的 Runner

在 TestMoon 选择空间、项目，进入「Runner 管理」创建私有节点，复制一次性注册凭证。将下面网关地址替换为实际平台地址。

Mac 默认 zsh：

```bash
export TESTMOON_URL='https://testmoon.example.com'
export TESTMOON_WORKER_IMAGE='testmoon/worker:0.1.3'
export TESTMOON_RUNNER_STATE="$HOME/.api_moon/runner/state.json"
export DOCKER_HOST="$(docker context inspect --format '{{.Endpoints.docker.Host}}')"
# 仅首次注册需要这三行；Linux Bash 使用 read -r -s -p 'Runner token: ' TESTMOON_ENROLLMENT_TOKEN。
read -r -s 'TESTMOON_ENROLLMENT_TOKEN?Runner 注册凭证：'
echo
export TESTMOON_ENROLLMENT_TOKEN
testmoon-runner
```

Windows PowerShell：

```powershell
$env:TESTMOON_URL = 'https://testmoon.example.com'
$env:TESTMOON_WORKER_IMAGE = 'testmoon/worker:0.1.3'
$env:TESTMOON_RUNNER_STATE = "$HOME/.api_moon/runner/state.json"
$env:DOCKER_HOST = docker context inspect --format '{{.Endpoints.docker.Host}}'
# 仅首次注册需要这两行。
$token = Read-Host 'Runner 注册凭证' -AsSecureString
$env:TESTMOON_ENROLLMENT_TOKEN = [System.Net.NetworkCredential]::new('', $token).Password
testmoon-runner
```

平台显示节点「在线」即接入成功。保持终端运行，Ctrl+C 停止。后续启动保留同一平台地址、镜像和状态文件，无需再次输入凭证。仅本机 HTTP 平台需额外设置 `TESTMOON_ALLOW_HTTP=1`。

## CLI 与 Skill

安装 Skill 无需登录、网络或源码：

```bash
apimoon skill install --agent codex --global
apimoon skill install --agent trae --global
apimoon skill install --agent trae-cn --global
```

选择与 Agent 对应的一条命令。不加 `--global` 默认安装到当前项目；指定其他项目可用 `--project-dir DIR`。检查状态用 `skill status`；已有内容不同，`skill install --force` 会先备份后更新。

配置 CLI、选择作用域和在线 Runner：

```bash
apimoon config init --base-url https://testmoon.example.com
apimoon auth login --username <用户名> --password-stdin
apimoon org list
apimoon config set context.org_id <组织ID>
apimoon project list
apimoon config set context.project_id <项目ID>
apimoon runner list
apimoon config set context.runner_id <RunnerID>
```

密码通过 stdin 提供。当前 `runner list` 只查询平台节点，本机 Runner 使用已下载的独立程序启动。

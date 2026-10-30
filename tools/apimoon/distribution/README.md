# 安装 APIMoon 与 Runner

APIMoon、Runner 均从 [GitHub Releases](https://github.com/mi991203/TestMoon/releases) 下载预编译程序；Worker 从 Docker Hub 拉取 `flyingcode/testmoon-worker:0.1.0`。用户只需安装并启动 Docker，无需下载源码或安装 Git、Go、Node；Windows 使用 Linux 容器模式。

支持 macOS（darwin）、Linux、Windows 的 amd64、arm64。每个架构分别提供两个文件，全部直接下载，不提供压缩包：

| 系统 | APIMoon | Runner |
| --- | --- | --- |
| macOS / Linux | `apimoon_vX.Y.Z_<系统>_<架构>` | `testmoon-runner_vX.Y.Z_<系统>_<架构>` |
| Windows | `apimoon_vX.Y.Z_windows_<架构>.exe` | `testmoon-runner_vX.Y.Z_windows_<架构>.exe` |

同时下载 `SHA256SUMS`。下文程序以 v0.1.3 为例，替换为实际已发布版本；APIMoon 与 Runner 下载同一 Release 的程序。Worker 镜像版本独立管理，当前为 `0.1.0`。运行前用文件哈希与校验文件对应行比较。

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

Mac 默认 zsh 将 PATH 写入配置，新终端及 Agent 才能找到程序：

```bash
printf '\nexport PATH="$HOME/.local/bin:$PATH"\n' >> "$HOME/.zshrc"
source "$HOME/.zshrc"
```

Linux 按所用 shell 配置 PATH。升级前停止旧 Runner，保留 `~/.api_moon/`。

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

## 拉取 Worker

Mac、Linux、Windows PowerShell 均执行：

```bash
docker pull flyingcode/testmoon-worker:0.1.0
```

Docker 自动选择匹配的 CPU 架构。启动 Runner 时必须显式指定该镜像地址；当前程序默认地址仍为 `testmoon/worker:0.1.0`。

## 启动已下载的 Runner

在 TestMoon 选择空间、项目，进入「Runner 管理」创建私有节点，复制一次性注册凭证。将下面网关地址替换为实际平台地址。

Mac 默认 zsh：

```bash
export TESTMOON_URL='https://testmoon.example.com'
export TESTMOON_WORKER_IMAGE='flyingcode/testmoon-worker:0.1.0'
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
$env:TESTMOON_WORKER_IMAGE = 'flyingcode/testmoon-worker:0.1.0'
$env:TESTMOON_RUNNER_STATE = "$HOME/.api_moon/runner/state.json"
$env:DOCKER_HOST = docker context inspect --format '{{.Endpoints.docker.Host}}'
# 仅首次注册需要这两行。
$token = Read-Host 'Runner 注册凭证' -AsSecureString
$env:TESTMOON_ENROLLMENT_TOKEN = [System.Net.NetworkCredential]::new('', $token).Password
testmoon-runner
```

平台显示节点「在线」即接入成功。保持终端运行，Ctrl+C 停止；在另一个终端配置 APIMoon。后续启动重新设置上述环境变量并保留同一状态文件，省略首次注册的凭证输入步骤。

仅本机 HTTP 平台需额外设置：Mac/Linux 执行 `export TESTMOON_ALLOW_HTTP=1`；Windows 执行 `$env:TESTMOON_ALLOW_HTTP = '1'`。

## 配置 APIMoon

### 初始化并登录

将地址替换为与 Runner 相同的 TestMoon 网关根地址，不带 `/api` 或页面路径。默认配置位于 `~/.api_moon/config.yaml`（Windows 为 `$HOME\.api_moon\config.yaml`），无需手工填写鉴权信息。

```bash
apimoon config init --base-url https://testmoon.example.com
```

已有配置先用 `apimoon config show` 查看；需要更换地址时执行 `apimoon config set base_url https://testmoon.example.com`，然后重新登录。

Mac 默认 zsh 隐藏输入密码，将用户名替换为平台账号：

```bash
read -r -s 'APIMOON_PASSWORD?TestMoon 登录密码：'
echo
printf '%s' "$APIMOON_PASSWORD" | apimoon auth login --username '你的用户名' --password-stdin
unset APIMOON_PASSWORD
```

Windows PowerShell：

```powershell
$password = Read-Host 'TestMoon 登录密码' -AsSecureString
$OutputEncoding = [System.Text.UTF8Encoding]::new($false)
[System.Net.NetworkCredential]::new('', $password).Password | apimoon auth login --username '你的用户名' --password-stdin
Remove-Variable password
```

Linux Bash 将读取密码的一行替换为 `read -r -s -p 'TestMoon password: ' APIMOON_PASSWORD`，其余同 Mac。密码通过 stdin 传入，CLI 自动保存登录会话；会话到期后重新登录。平台账号密码与 Runner 一次性注册凭证用途不同。

### 选择空间、项目、环境和 Runner

Mac、Windows 使用相同命令。按顺序执行，将示例 ID 替换为列表输出中所选对象的实际 ID；「空间」在 CLI 中使用 `org`：

```bash
apimoon org list
apimoon config set context.org_id '空间ID'
apimoon project list
apimoon config set context.project_id '项目ID'
apimoon environment list
apimoon config set context.environment_id '环境ID'
apimoon runner list
apimoon config set context.runner_id '本机在线RunnerID'
apimoon auth status
apimoon context show
```

空间、项目需预先在平台创建，账号需具备项目权限。环境尚未创建时，可在平台创建或由测试计划创建，再绑定实际环境 ID。选择前面启动的项目私有 Runner，并确认在线。切换空间会清除项目、环境和 Runner；切换项目会清除环境和 Runner，需重新选择。

当前 `runner list` 只查询平台节点，本机 Runner 由独立程序启动。环境和 Runner 配置作为新建资源的默认值，不会修改已有任务保存的配置。

## 安装 Agent Skill 并使用

安装 Skill 无需登录、网络或源码：

```bash
apimoon skill install --agent codex --global
apimoon skill install --agent trae --global
apimoon skill install --agent trae-cn --global
```

选择与 Agent 对应的一条命令。不加 `--global` 默认安装到当前项目；指定其他项目可用 `--project-dir DIR`。检查状态用 `skill status`；已有内容不同，`skill install --force` 会先备份后更新。

完成配置后，可让 Agent「根据需求文档和当前项目代码生成接口测试，通过 APIMoon 同步、执行并分析报告」。使用已有计划文件时：

```bash
apimoon plan validate --file test-plan.json
apimoon plan sync --file test-plan.json
apimoon task run --id '同步返回的任务ID' --idempotency-key '本次执行唯一键' --wait
apimoon run get --id '执行返回的ID'
```

每次新的执行使用新的幂等键；报告依据请求、响应和断言分析。APIMoon 的帮助可逐级查看，例如 `apimoon task run --help`。

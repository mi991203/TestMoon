# TestMoon

面向开发测试的大仓应用集。首个应用为低代码 HTTP 测试平台，统一门户、网关及账号组织能力供后续应用复用。

已实现 API、接口用例、顺序场景、手动任务、执行报告、HTTP Mock、接口文档分享、环境、项目及公共/项目私有 Runner。首版执行 HTTP；协议和前后置操作保留版本化扩展入口。

## 本地启动

推荐使用 Docker Compose 运行整个平台，本机只运行私有 Runner。需要 Docker、Compose v2+、Python 3 和 Go 1.27（toolchain 自动选择 1.27.3）；Node、pnpm 和前端均在镜像内构建。

从原本机进程部署迁移时，先执行 `./scripts/dev.sh stop`，并停止原先手动运行的 Runner。保留 `deploy/local/.env`、`.local/runner.json` 和已有数据卷；平台地址继续使用 `http://localhost:18080`，无需重新注册节点。

```bash
./scripts/local-prepare.sh  # 首次：准备私有配置、Worker镜像、本机Runner
docker compose --env-file deploy/local/.env -f deploy/local/compose.yaml up -d --build --wait
./scripts/runner.sh start
```

第一次使用 Runner，需要先在界面创建私有节点，将一次性凭证通过 `TESTMOON_ENROLLMENT_TOKEN` 环境变量传给启动脚本；已有 `.local/runner.json` 时不需要凭证。

```bash
docker compose --env-file deploy/local/.env -f deploy/local/compose.yaml ps
docker compose --env-file deploy/local/.env -f deploy/local/compose.yaml logs --tail 100 gateway execution
./scripts/runner.sh status
./scripts/runner.sh stop
docker compose --env-file deploy/local/.env -f deploy/local/compose.yaml stop
```

日常启动可省略 `--build`；修改平台代码后重新执行带 `--build` 的启动命令。修改 Runner/Worker 后重跑准备命令并重启 Runner。`stop` 保留容器，`down` 删除容器及网络但保留数据卷。停机前等待任务结束，先停 Runner 再停平台；平台停止期间，Runner 持续心跳失败会退出，平台恢复后需再次启动 Runner。

Compose 内的服务通过服务名访问，网关发布到本机 `18080`；本机 Runner 经此入口连接平台，并使用本机网络访问被测 API。Worker 仍按执行创建且没有直接网络。VPN 需提供本机可用的 DNS/直连路由；Runner 当前不使用系统 HTTP 代理。

### 本机源码开发

原有开发方式仍保留，与 Compose 应用部署二选一，避免端口冲突。需要额外安装 Node 24.22.0 LTS、pnpm 11.28.6。协议文件已生成并提交；只有修改 proto 才需要 protoc 36.2 和生成插件。

```bash
make deps
make infra-up          # MySQL、Redis、Kafka 健康检查及四个业务数据库
make build-all         # 分别使用 Go 和 pnpm 构建，Makefile 统一入口
make worker-image      # 执行时必须有本地 Worker 镜像
make dev
```

打开 <http://localhost:18080>。初始账号密码在自动生成的 `deploy/local/.env` 中；文件权限为 0600 且已被 Git 忽略。启动脚本不打印密码。MySQL/Redis/Kafka 数据通过 Docker volume 保留。

测试应用左侧使用统一“API测试”入口，接口树可打开接口，并在Header切换API调试和关联用例；主要业务详情使用独立页面路由。

首次由平台管理员创建组织并指定组织管理员；组织管理员进入测试应用创建项目。平台管理员身份不会自动取得组织业务权限。

```bash
make stop              # 停止本仓库脚本启动的业务服务
make infra-down        # 停止此Compose项目（包括已启动的应用），保留数据卷
./scripts/infra.sh up mysql redis  # 可只启动需要的中间件
./scripts/dev.sh status
```

日志在 `.local/logs/`，本地二进制在 `.local/bin/`。修改业务代码后执行 `make build-all`、`make stop`、`make dev`。前端可分别用 `pnpm --filter @testmoon/portal dev` 与 `pnpm --filter @testmoon/testing dev` 进行热更新；Cookie/Origin 的完整联调以网关生产构建为准。

## Runner 启动

进入项目「Runner 管理」生成私有节点的一次性注册凭证。公共节点由门户中的平台管理员创建。未部署在线节点时拒绝新执行。

```bash
make runner-image
# 远程机器应拉取或加载相同版本的 runner 和 worker 两个镜像。
export TESTMOON_URL=https://testmoon.example.com
export TESTMOON_ENROLLMENT_TOKEN='<界面生成的一次性凭证>'
docker compose -f deploy/runner/compose.yaml up -d
unset TESTMOON_ENROLLMENT_TOKEN
```

注册成功后凭证保存在 Runner 独立状态卷；后续重启不需要注册凭证。一个私有 Runner 仅服务一个项目。公共 Runner 只允许访问公网地址；私有 Runner 可以访问它所在部署网络。容器中的 localhost 指 Runner 管理进程所在容器，而非 Docker 宿主机。

本地可在宿主机运行 Go Runner，便于调用本机服务：

```bash
export TESTMOON_URL=http://localhost:18080
export TESTMOON_ALLOW_HTTP=1 # 仅本地联调
export TESTMOON_ENROLLMENT_TOKEN='<私有节点凭证>'
export TESTMOON_RUNNER_STATE="$PWD/.local/runner.json"
export DOCKER_HOST="$(docker context inspect --format '{{.Endpoints.docker.Host}}')"
.local/bin/runner
```

Runner 管理进程需要 Docker API 权限，应部署在专用执行机器。每次执行创建独立、无直接网络、只读、非 root、限资源容器；执行容器不挂 Docker socket、宿主文件或平台凭证。`vm` 用于 JS 接口环境，隔离边界是每次执行的 Docker 容器。

## Mock 与文档

Mock 独立域名为 `http://mock.localhost:18080/m/<endpoint-id>/<path>`；若操作系统不解析 `mock.localhost`，本地 hosts 将其映射到 `127.0.0.1`。生产必须使用与 Web 不同的 Mock 域名。测试接口可使用 `{{base_url}}`，由项目环境配置。

文档分享在 `/share/<随机token>`，可设置密码、有效期并撤销；内容为所选接口的当前定义，不包含项目环境与用例。

## 构建与文档

```bash
make tools             # 安装固定版本 protoc Go/gRPC/gateway/OpenAPI 插件
make generate          # 原生 protoc，统一生成 Go、网关、OpenAPI、TS客户端
make typecheck
make build-go
make build-web
```

- [需求目录](docs/requirements/README.md)：统一存放需求及验收条件。
- [目录与技术方案](docs/architecture/项目结构与技术选型.md)：依赖、服务边界、数据与运行机制。
- [接口与扩展约定](docs/architecture/接口与扩展约定.md)：HTTP/gRPC、变量、脚本和导入兼容边界。
- [数据模型与迁移](docs/architecture/数据模型与迁移.md)：分服务建表定义、历史回填与升级边界。
- [部署配置](docs/architecture/部署配置.md)：端口、配置优先级、限额和生产部署条件。
- [实现与验证记录](docs/requirements/delivery.md)：已验证路径及未验证范围。

平台支持 Compose 镜像构建，也保留原生工具与 Makefile 开发入口，未引入 Bazel 或 CI。

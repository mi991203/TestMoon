# TestMoon

低代码 HTTP 测试平台，支持 API 调试、用例与场景编排、任务执行、报告、Mock 和接口文档分享；提供统一门户、组织权限及公共/项目私有 Runner。

使用 AI Agent 测试已有项目时，按 [APIMoon 安装说明](tools/apimoon/README.md) 从 GitHub Releases 安装预编译 CLI 和本机 Runner，并在 Docker 中本地构建 Worker；使用者无需安装 Go。以下章节用于启动或开发 TestMoon 服务端。

## 本地启动

需要已启动的 Docker、Compose v2+、Python 3 和 Go 1.27（toolchain 1.27.3）。平台在容器内构建运行，Runner 在本机运行。以下命令均在仓库根目录执行。

1. 准备配置、Runner 和 Worker 镜像，启动平台：

   ```bash
   ./scripts/local-prepare.sh
   docker compose --env-file deploy/local/.env -f deploy/local/compose.yaml up -d --build --wait
   ```

2. 打开 <http://localhost:18080>，使用自动生成的 `deploy/local/.env` 中的 `TESTMOON_BOOTSTRAP_USERNAME`、`TESTMOON_BOOTSTRAP_PASSWORD` 登录。平台管理员创建组织并指定组织管理员，由组织管理员进入测试应用创建项目；平台管理员不自动拥有组织业务权限。

3. 在项目「Runner 管理」创建私有节点，生成一次性注册凭证并启动 Runner：

   ```bash
   TESTMOON_ENROLLMENT_TOKEN='<一次性注册凭证>' ./scripts/runner.sh start
   ```

注册状态保存在 `.local/runner.json`，后续无需凭证。有在线 Runner 后即可执行测试。

### 日常操作

启动已有服务：

```bash
docker compose --env-file deploy/local/.env -f deploy/local/compose.yaml up -d --wait
./scripts/runner.sh start
```

等待执行任务结束后停机：

```bash
./scripts/runner.sh stop
docker compose --env-file deploy/local/.env -f deploy/local/compose.yaml stop
```

更新平台代码：执行首次启动中带 `--build` 的命令，再执行 `./scripts/runner.sh start`。修改 Runner/Worker：先停止 Runner、执行 `./scripts/local-prepare.sh`，再启动。两者均修改时先准备 Runner/Worker，再更新平台，最后启动 Runner。

查看状态和日志：

```bash
docker compose --env-file deploy/local/.env -f deploy/local/compose.yaml ps
docker compose --env-file deploy/local/.env -f deploy/local/compose.yaml logs --tail 100 gateway execution
./scripts/runner.sh status
```

本机日志在 `.local/logs/`。`stop` 保留容器和数据；`down` 删除容器及网络，保留数据卷。保管好 `deploy/local/.env` 和 `.local/runner.json`，勿提交到 Git。

## 源码开发

额外安装 Node 24.22.0 LTS、pnpm 11.28.6、Make 和 curl。与 Compose 平台部署二选一，切换前停止另一种方式的应用服务。

```bash
make deps
make infra-up          # 启动 MySQL、Redis、Kafka，创建业务数据库
make build-all         # 构建后端、Runner 和前端
make worker-image
make dev
```

首次登录及 Runner 的注册、更新步骤同上。修改业务代码后执行 `make stop && make build-all && make dev`。

```bash
./scripts/dev.sh status
./scripts/runner.sh stop
make stop              # 停止源码启动的业务服务
make infra-down        # 删除本地 Compose 容器和网络，保留数据卷
```

前端热更新：`pnpm --filter @testmoon/portal dev`、`pnpm --filter @testmoon/testing dev`。完整联调使用网关中的前端构建产物。

```bash
pnpm format            # 格式化前端及共享库，排除生成的 API 客户端
pnpm format:check
make typecheck
```

仅修改 proto 时需安装 protoc 36.2，再执行：

```bash
make tools             # 安装固定版本生成插件
make generate          # 生成 Go、网关、OpenAPI 和 TS 客户端
make typecheck
make build-all
```

## 远程 Runner

执行机器需要 Docker 和平台 HTTPS 地址。公共节点在管理员门户创建，私有节点在项目「Runner 管理」创建并生成凭证。

在仓库中构建镜像，将两个镜像及 `deploy/runner/compose.yaml` 分发到执行机器：

```bash
make runner-image worker-image
```

在执行机器启动（按实际位置调整 Compose 文件路径）：

```bash
export TESTMOON_URL=https://testmoon.example.com
export TESTMOON_ENROLLMENT_TOKEN='<一次性注册凭证>'
docker compose -f deploy/runner/compose.yaml up -d
unset TESTMOON_ENROLLMENT_TOKEN
```

注册状态保存在 `runner-state` 卷，后续无需凭证。Runner 需要 Docker API 权限，应使用专用执行机器；每次执行使用独立 Worker 容器。私有节点仅服务所属项目，可访问部署网络；公共节点仅能访问公网。

容器中的 `localhost` 指 Runner 容器。本机 Runner 使用宿主机网络，不使用系统 HTTP 代理；VPN 需提供可用的 DNS 和直连路由。

## Mock 与文档分享

- Mock：`http://mock.localhost:18080/m/<endpoint-id>/<path>`。无法解析时，在 hosts 中添加 `127.0.0.1 mock.localhost`；生产环境使用与 Web 不同的 Mock 域名。
- 环境变量：接口地址可用 `{{base_url}}`，在项目环境中配置其值。
- 文档分享：`/share/<随机token>`，支持密码、有效期和撤销，展示所选接口的当前定义。

## 参考文档

- [APIMoon CLI](tools/apimoon/README.md)（Agent 生成计划、同步、执行及报告读取）
- [学生管理测试服务](fixtures/school_management/README.md)（独立启动，提供 HTTP/gRPC CRUD）
- [需求与验收](docs/requirements/README.md)
- [项目结构与技术选型](docs/architecture/项目结构与技术选型.md)
- [API 文档](API文档.md)
- [接口与扩展约定](docs/architecture/接口与扩展约定.md)
- [数据模型与迁移](docs/architecture/数据模型与迁移.md)
- [部署配置](docs/architecture/部署配置.md)
- [实现与验证记录](docs/requirements/delivery.md)

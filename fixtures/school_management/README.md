# School Management fixture

供 testing 调用的简单学生管理服务，使用真实 MySQL 存储，提供学校、班级、学生 CRUD。独立运行，不接入 Docker Compose、`make dev` 或平台网关，也不需要平台登录、服务密钥、Redis 或 Kafka。

## 数据库准备

复用已有 MySQL，默认独立 schema 为 `testmoon_school_management`。使用管理员执行本目录的 `init.sql`，创建 schema 并授权现有的 `testmoon` 用户。例如已有本地 MySQL 容器时，在仓库根目录执行：

```bash
docker exec -i testmoon-mysql-1 sh -c 'MYSQL_PWD="$MYSQL_ROOT_PASSWORD" exec mysql -uroot' \
  < fixtures/school_management/init.sql
```

SQL 只创建新 schema 和授权，不修改平台业务库。如果使用其他账号或 schema，请相应调整初始化 SQL 和 DSN。服务启动前 schema 必须存在；服务启动时自动创建/补齐表，不执行清库或预置数据。

## 启动

在仓库根目录设置 DSN 后启动：

```bash
export TESTMOON_SCHOOL_MANAGEMENT_DSN='testmoon:<MySQL密码>@tcp(127.0.0.1:13306)/testmoon_school_management?charset=utf8mb4&parseTime=True&loc=UTC&timeout=5s&readTimeout=30s&writeTimeout=30s'
go run ./fixtures/school_management/cmd/server
```

复用 `deploy/local/.env` 的现有本地配置时，也可执行：

```bash
source deploy/local/.env
export TESTMOON_SCHOOL_MANAGEMENT_DSN="testmoon:${MYSQL_PASSWORD}@tcp(127.0.0.1:${MYSQL_PORT:-13306})/testmoon_school_management?charset=utf8mb4&parseTime=True&loc=UTC&timeout=5s&readTimeout=30s&writeTimeout=30s"
go run ./fixtures/school_management/cmd/server
```

不要提交含真实密码的 DSN。Ctrl+C 停止服务，数据保留在 MySQL 中。

| 环境变量 | 默认值 |
| --- | --- |
| `TESTMOON_SCHOOL_MANAGEMENT_DSN` | 必填，已有 MySQL 的 DSN |
| `TESTMOON_SCHOOL_MANAGEMENT_HTTP` | `127.0.0.1:8096` |
| `TESTMOON_SCHOOL_MANAGEMENT_GRPC` | `127.0.0.1:9096` |

HTTP 与 gRPC 均无需鉴权，默认监听本机。`GET /healthz` 检查数据库连通性。需要其他机器访问时，将对应监听地址配置为可访问地址，并在测试环境内使用。

可单独构建二进制：

```bash
go build -o .local/bin/school_management ./fixtures/school_management/cmd/server
.local/bin/school_management
```

## HTTP 接口

基础路径：`http://127.0.0.1:8096/api/school-management/v1`。

| 操作 | 方法及路径 |
| --- | --- |
| 创建 | `POST /schools`、`POST /classes`、`POST /students` |
| 详情 | `GET /schools/{id}`、`GET /classes/{id}`、`GET /students/{id}` |
| 列表 | `GET /schools`、`GET /classes`、`GET /students` |
| 更新 | `PUT /schools/{id}`、`PUT /classes/{id}`、`PUT /students/{id}` |
| 删除 | `DELETE /schools/{id}`、`DELETE /classes/{id}`、`DELETE /students/{id}` |

学校可编辑字段为 `name`、`address`；班级为 `name`、`school_id`；学生为 `name`、`student_no`、`class_id`。`name`、`student_no` 和关联 ID 必填；`address` 可为空。不要求名称或学号唯一。

创建和更新返回资源本身，包含 UUIDv7 `id`、RFC3339 `created_at` / `updated_at`。PUT 替换全部可编辑字段，路径中的 ID 为更新目标。成功状态码统一为 200；删除返回 `{}`。列表返回 `{"items":[...]}`，无匹配时为 `{"items":[]}`，无分页。班级列表支持 `?school_id=...`；学生列表支持 `?school_id=...&class_id=...`，同时填写时按交集筛选。

```bash
# 1. 创建学校，记录响应 id 为 SCHOOL_ID。
curl -sS http://127.0.0.1:8096/api/school-management/v1/schools \
  -H 'Content-Type: application/json' \
  -d '{"name":"第一中学","address":"测试路1号"}'

# 2. 创建班级，记录响应 id 为 CLASS_ID。
curl -sS http://127.0.0.1:8096/api/school-management/v1/classes \
  -H 'Content-Type: application/json' \
  -d '{"name":"一年一班","school_id":"<SCHOOL_ID>"}'

# 3. 创建学生。
curl -sS http://127.0.0.1:8096/api/school-management/v1/students \
  -H 'Content-Type: application/json' \
  -d '{"name":"张三","student_no":"S001","class_id":"<CLASS_ID>"}'

# 4. 查询学校的学生。
curl -sS 'http://127.0.0.1:8096/api/school-management/v1/students?school_id=<SCHOOL_ID>'
```

删除采用物理删除：删学校会级联删除所属班级及学生，删班级会级联删除所属学生。学生的学校归属由班级推导，不重复存储。非法字段、无效关联返回 400；详情/更新/删除目标不存在返回 404。错误使用标准 gRPC 状态映射及 `code`、`message`、`details` 字段。

## gRPC 与协议生成

gRPC 默认地址为 `127.0.0.1:9096`，服务名为 `testmoon.fixtures.school_management.v1.SchoolManagement`。启用了 reflection，可通过 grpcurl 直接调用：

```bash
grpcurl -plaintext 127.0.0.1:9096 list
grpcurl -plaintext -d '{"name":"第二中学","address":"测试地址"}' \
  127.0.0.1:9096 testmoon.fixtures.school_management.v1.SchoolManagement/CreateSchool
grpcurl -plaintext -d '{}' \
  127.0.0.1:9096 testmoon.fixtures.school_management.v1.SchoolManagement/ListSchools
```

协议与 OpenAPI 在 `protocols/fixtures/school_management/v1/`。安装项目固定版本 protoc 插件后，独立生成或随根生成流程执行：

```bash
make tools
bash fixtures/school_management/scripts/generate-proto.sh
# 或 make generate
```

fixture 生成 Go、gRPC、grpc-gateway、OpenAPI，不加入平台的 TS 客户端生成。

## testing 调用

为 testing 项目环境配置 `base_url=http://127.0.0.1:8096`，接口 URL 使用 `{{base_url}}/api/school-management/v1/schools` 等路径。使用本机私有 Runner 执行；公共 Runner 不允许访问本机/私网地址。

可编排“创建学校 → 提取 `/id` → 创建班级 → 提取 `/id` → 创建学生 → 更新/查询 → 删除学校 → 确认班级和学生均返回 404”。当前 testing 执行器支持 HTTP；gRPC 用生成的 Go 客户端或 grpcurl 单独调用。

Runner 在远程机器或容器运行时，`127.0.0.1` 指 Runner 所在环境；此时需要调整服务监听地址和 `base_url` 为 Runner 可达的服务地址。

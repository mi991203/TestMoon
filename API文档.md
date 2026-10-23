# TestMoon API 文档

本文档对应仓库当前实现。HTTP API 由网关 `http://localhost:18080` 提供；生产环境请替换为实际域名。接口定义来源于 `protocols/**/v1/*.proto`，字段使用 JSON 的 lowerCamelCase 命名。

## 通用约定

- 除特别注明外，管理及业务 API 均使用 `POST`、`Content-Type: application/json`，请求字段放在 JSON body。
- 登录成功后服务端设置会话 Cookie：本地为 `testmoon_session`，启用安全 Cookie 时为 `__Host-testmoon_session`。浏览器后续请求自动携带 Cookie。
- 除登录、分享文档读取及 Runner 通信外，API 请求需要有效会话 Cookie；除会话查询外，还需要 `X-CSRF-Token` 请求头。登录或会话响应中的 `csrf` 值用于该请求头。登出、改密等已登录请求也需要 CSRF。
- 网关在 API 请求上限制 body 最大为 12 MiB；Mock 被测请求 body 最大为 1 MiB。
- Protobuf `int64` 在 JSON 中通常以字符串编码；时间戳采用 RFC 3339 字符串。proto3 的未赋值字段可能以空字符串、`false`、`0` 或空对象/数组返回。
- 成功响应的 JSON 字段对应各接口的 `*Response` 消息；失败时使用 HTTP 状态码，网关校验失败可能返回纯文本，gRPC-Gateway 业务错误为 JSON 状态对象（`code`、`message`、`details`）。调用方应同时检查 HTTP 状态和业务响应。

## 账号与组织

前缀：`/api/platform/v1`。登录、登出、会话、密码、用户及组织管理接口如下。

| 方法与路径 | 请求 body 字段 | 说明 |
|---|---|---|
| `POST /api/platform/v1/login` | `username`, `password` | 登录；响应含 `csrf`、`user`，并设置会话 Cookie。 |
| `POST /api/platform/v1/logout` | `{}` | 注销当前会话并清理 Cookie。 |
| `POST /api/platform/v1/session` | `token` | 校验会话并返回 `principal`、`organizations`、`csrf`；浏览器通常由网关/前端使用 Cookie 建立会话状态。 |
| `POST /api/platform/v1/change-password` | `oldPassword`, `newPassword` | 修改当前用户密码。 |
| `POST /api/platform/v1/users` | `{}` | 平台管理员查询用户。 |
| `POST /api/platform/v1/create-user` | `username`, `name`, `password`, `admin` | 创建用户。 |
| `POST /api/platform/v1/reset-password` | `id`, `password` | 重置用户密码。 |
| `POST /api/platform/v1/set-user-status` | `id`, `disabled` | 启用或停用用户。 |
| `POST /api/platform/v1/organizations` | `{}` | 查询当前用户可见的组织；平台管理员可查询组织列表。 |
| `POST /api/platform/v1/create-organization` | `name`, `adminUserId` | 创建组织并指定组织管理员。 |
| `POST /api/platform/v1/update-organization` | `id`, `name` | 更新组织名称。 |
| `POST /api/platform/v1/members` | `orgId` | 查询组织成员。 |
| `POST /api/platform/v1/set-member` | `orgId`, `username`, `role` | 设置组织成员及角色。 |
| `POST /api/platform/v1/remove-member` | `orgId`, `username` | 移除组织成员。 |

登录请求示例：

```http
POST /api/platform/v1/login
Content-Type: application/json

{"username":"admin","password":"your-password"}
```

## 测试资源、项目与分享

前缀：`/api/testing/v1`。项目、API、用例、场景、环境、文件夹、成员、导入和分享管理接口如下。

| 方法与路径 | 请求 body 字段 | 说明 |
|---|---|---|
| `POST /api/testing/v1/projects` | `orgId` | 查询组织下的项目。 |
| `POST /api/testing/v1/create-project` | `orgId`, `name`, `description` | 创建项目。 |
| `POST /api/testing/v1/update-project` | `orgId`, `id`, `name`, `description` | 更新项目。 |
| `POST /api/testing/v1/delete-project` | `orgId`, `projectId` | 删除项目。 |
| `POST /api/testing/v1/authorize` | `orgId`, `projectId`, `permission` | 查询当前用户对项目的授权状态。 |
| `POST /api/testing/v1/members` | `orgId`, `projectId` | 查询项目成员。 |
| `POST /api/testing/v1/set-member` | `orgId`, `projectId`, `userId`, `role` | 设置项目成员角色。 |
| `POST /api/testing/v1/remove-member` | `orgId`, `projectId`, `userId` | 移除项目成员。 |
| `POST /api/testing/v1/apis/list` | `orgId`, `projectId`, `search`, `folderId` | 查询 API；`folderId` 可省略，支持按名称搜索。 |
| `POST /api/testing/v1/apis/get` | `orgId`, `projectId`, `id` | 获取 API 详情。 |
| `POST /api/testing/v1/apis/create` | `orgId`, `projectId`, `name`, `folderId`, `document` | 创建 API 定义。 |
| `POST /api/testing/v1/apis/update` | `orgId`, `projectId`, `id`, `revision`, `name`, `folderId`, `document` | 更新 API；使用 `revision` 做版本冲突检查。 |
| `POST /api/testing/v1/apis/delete` | `orgId`, `projectId`, `id` | 删除 API。 |
| `POST /api/testing/v1/cases/list` | `orgId`, `projectId`, `search`, `folderId`, `apiId` | 查询用例，可按 API 过滤。 |
| `POST /api/testing/v1/cases/get` | `orgId`, `projectId`, `id` | 获取用例详情。 |
| `POST /api/testing/v1/cases/create` | `orgId`, `projectId`, `name`, `folderId`, `document` | 创建接口用例。 |
| `POST /api/testing/v1/cases/update` | `orgId`, `projectId`, `id`, `revision`, `name`, `folderId`, `document` | 更新用例；使用 `revision` 做版本冲突检查。 |
| `POST /api/testing/v1/cases/delete` | `orgId`, `projectId`, `id` | 删除用例。 |
| `POST /api/testing/v1/scenarios/list` | `orgId`, `projectId`, `search`, `folderId` | 查询顺序场景。 |
| `POST /api/testing/v1/scenarios/get` | `orgId`, `projectId`, `id` | 获取场景详情。 |
| `POST /api/testing/v1/scenarios/create` | `orgId`, `projectId`, `name`, `folderId`, `document` | 创建场景。 |
| `POST /api/testing/v1/scenarios/update` | `orgId`, `projectId`, `id`, `revision`, `name`, `folderId`, `document` | 更新场景；使用 `revision` 做版本冲突检查。 |
| `POST /api/testing/v1/scenarios/delete` | `orgId`, `projectId`, `id` | 删除场景。 |
| `POST /api/testing/v1/environments/list` | `orgId`, `projectId`, `search`, `folderId` | 查询环境。 |
| `POST /api/testing/v1/environments/get` | `orgId`, `projectId`, `id` | 获取环境详情。 |
| `POST /api/testing/v1/environments/create` | `orgId`, `projectId`, `name`, `folderId`, `document` | 创建环境。 |
| `POST /api/testing/v1/environments/update` | `orgId`, `projectId`, `id`, `revision`, `name`, `folderId`, `document` | 更新环境；使用 `revision` 做版本冲突检查。 |
| `POST /api/testing/v1/environments/delete` | `orgId`, `projectId`, `id` | 删除环境。 |
| `POST /api/testing/v1/folders/list` | `orgId`, `projectId`, `search`, `folderId` | 查询资源文件夹。 |
| `POST /api/testing/v1/folders/get` | `orgId`, `projectId`, `id` | 获取文件夹详情。 |
| `POST /api/testing/v1/folders/create` | `orgId`, `projectId`, `name`, `folderId`, `document` | 创建文件夹。 |
| `POST /api/testing/v1/folders/update` | `orgId`, `projectId`, `id`, `revision`, `name`, `folderId`, `document` | 更新文件夹。 |
| `POST /api/testing/v1/folders/delete` | `orgId`, `projectId`, `id` | 删除文件夹。 |
| `POST /api/testing/v1/import-preview` | `orgId`, `projectId`, `format`, `content` | 预览导入内容，返回批次 `id`、条目和诊断信息。 |
| `POST /api/testing/v1/import-commit` | `orgId`, `projectId`, `id`, `selected`, `updates`, `targetFolderId` | 提交预览批次；可指定导入条目、更新目标及目标文件夹。 |
| `POST /api/testing/v1/snapshot` | `orgId`, `projectId`, `items`, `environmentId`, `runnerId`, `timeoutSeconds`, `debug` | 创建执行快照。 |
| `POST /api/testing/v1/shares` | `orgId`, `projectId` | 查询文档分享。 |
| `POST /api/testing/v1/create-share` | `orgId`, `projectId`, `name`, `apiIds`, `password`, `expiresAt` | 创建 API 文档分享；密码和过期时间可选。 |
| `POST /api/testing/v1/delete-share` | `orgId`, `projectId`, `id` | 撤销分享。 |

### 资源定义字段

创建和更新资源时，`document` 使用对应的结构化定义：

- `APIDefinition`：`protocol`、`request`、`description`、`schema`、`environmentId`、`runnerId`。`request` 为 `HTTPRequest`：`method`、`url`、`headers`、`query`、`bodyType`、`body`、`auth`、`timeoutMs`。
- `TestCaseDefinition`：`apiId`、`name`、`request`、`override`、`variables`、`collectionVariables`、`before`、`after`、`assertions`、`extractions`、`environmentId`、`runnerId`。
- `ScenarioDefinition`：`name`、`steps`、`collectionVariables`。步骤可引用 `caseId` 或携带内联用例，并可定义请求覆盖、变量和失败后继续策略。
- `EnvironmentDefinition`：`variables` 映射；每个变量包含 `value` 和 `secret`。
- `FolderDefinition`：`kind`。

## 执行与 Runner

执行管理前缀：`/api/execution/v1`。Runner 通信前缀：`/api/runner/v1`。Runner 接口不使用浏览器会话 Cookie，由网关开放给 Runner 客户端并由 Runner 凭证鉴权。

| 方法与路径 | 请求 body 字段 | 说明 |
|---|---|---|
| `POST /api/execution/v1/tasks` | `orgId`, `projectId` | 查询手动执行任务。 |
| `POST /api/execution/v1/create-task` | `orgId`, `projectId`, `name`, `document` | 创建任务；`document` 为 `TaskDefinition`。 |
| `POST /api/execution/v1/update-task` | `orgId`, `projectId`, `id`, `revision`, `name`, `document` | 更新任务。 |
| `POST /api/execution/v1/delete-task` | `orgId`, `projectId`, `id` | 删除任务。 |
| `POST /api/execution/v1/debug-http` | `orgId`, `projectId`, `debug`, `environmentId`, `runnerId`, `timeoutSeconds`, `idempotencyKey` | 调试 HTTP 请求，可直接提交当前草稿。 |
| `POST /api/execution/v1/run-case` | `orgId`, `projectId`, `caseId`, `environmentId`, `runnerId`, `timeoutSeconds`, `idempotencyKey`, `draft` | 执行用例；`draft` 可携带未保存用例定义。 |
| `POST /api/execution/v1/run-scenario` | `orgId`, `projectId`, `scenarioId`, `environmentId`, `runnerId`, `timeoutSeconds`, `idempotencyKey` | 执行场景。 |
| `POST /api/execution/v1/run-task` | `orgId`, `projectId`, `taskId`, `idempotencyKey` | 执行任务。 |
| `POST /api/execution/v1/runs` | `orgId`, `projectId` | 查询执行记录。 |
| `POST /api/execution/v1/run` | `orgId`, `projectId`, `id` | 查询执行详情及结果。 |
| `POST /api/execution/v1/cancel` | `orgId`, `projectId`, `id` | 取消执行。 |
| `POST /api/execution/v1/project-runners/list` | `orgId`, `projectId` | 查询项目可用 Runner。 |
| `POST /api/execution/v1/private-runners/create` | `orgId`, `projectId`, `name`, `concurrency` | 创建项目私有 Runner；响应中的 `enrollmentToken` 为一次性注册凭证。 |
| `POST /api/execution/v1/private-runners/revoke` | `orgId`, `projectId`, `id` | 撤销私有 Runner。 |
| `POST /api/execution/v1/public-runners/list` | `{}` | 平台管理员查询公共 Runner。 |
| `POST /api/execution/v1/public-runners/create` | `name`, `concurrency` | 平台管理员创建公共 Runner。 |
| `POST /api/execution/v1/public-runners/revoke` | `id` | 平台管理员撤销公共 Runner。 |
| `POST /api/runner/v1/register` | `token`, `version`, `capabilities` | Runner 使用一次性凭证注册，响应返回 `runnerId` 和长期 `credential`。 |
| `POST /api/runner/v1/heartbeat` | `leases` | Runner 心跳并续租。 |
| `POST /api/runner/v1/poll` | `{}` | Runner 拉取待执行任务。 |
| `POST /api/runner/v1/progress` | `id`, `lease`, `result` | Runner 上报执行进度。 |
| `POST /api/runner/v1/complete` | `id`, `lease`, `result` | Runner 上报执行完成结果。 |

执行状态、请求/响应、断言和日志字段定义见 `protocols/testing/contracts/v1/testing.proto` 中的 `ExecutionResult`、`ItemResult`、`StepResult`、`HTTPRequest` 和 `HTTPResponse`。

## Mock

管理接口前缀：`/api/mock/v1`，需要登录和 CSRF。Mock 配置由 Endpoint 和 Rule 组成。

| 方法与路径 | 请求 body 字段 | 说明 |
|---|---|---|
| `POST /api/mock/v1/endpoints` | `orgId`, `projectId` | 查询 Mock Endpoint。 |
| `POST /api/mock/v1/create-endpoint` | `orgId`, `projectId`, `name` | 创建 Endpoint；响应返回 `path`。 |
| `POST /api/mock/v1/update-endpoint` | `orgId`, `projectId`, `id`, `name` | 更新 Endpoint。 |
| `POST /api/mock/v1/delete-endpoint` | `orgId`, `projectId`, `id` | 删除 Endpoint。 |
| `POST /api/mock/v1/rules` | `orgId`, `projectId` | 查询 Mock 规则。 |
| `POST /api/mock/v1/create-rule` | `orgId`, `projectId`, `name`, `endpointId`, `method`, `path`, `priority`, `document` | 创建规则。 |
| `POST /api/mock/v1/update-rule` | `orgId`, `projectId`, `id`, `name`, `endpointId`, `method`, `path`, `priority`, `document` | 更新规则。 |
| `POST /api/mock/v1/delete-rule` | `orgId`, `projectId`, `id` | 删除规则。 |

被测服务使用 Mock 公共入口：`http(s)://<Mock域名>/m/<endpoint-id>/<path>`。该入口支持转发对应 HTTP 方法；按请求方法、路径、Header、Query 和 JSON Body 匹配规则。匹配到规则后返回其 `document` 中的状态码、响应头、响应体，并按 `delayMs` 延迟；未匹配返回 `404`。Mock 请求不接受浏览器会话作为鉴权凭证。

`MockResponse` 字段：`conditions`、`status`、`headers`、`body`、`delayMs`。每个条件包含 `source`、`path`、`operator`、`expected`。

## 文档分享

创建分享后，客户端可通过以下公开接口读取：

```http
POST /api/share/v1/document
Content-Type: application/json

{"token":"分享链接中的随机 token","password":"分享密码（未设置密码时可省略）"}
```

响应返回分享名称 `name` 和 API 列表 `items`（包含 `id`、`name`、`document`）。分享读取无需登录；设置了密码的分享必须提交正确密码。分享只包含接口定义，不包含项目环境和用例。

## 网关辅助端点

| 方法与路径 | 说明 |
|---|---|
| `GET /healthz` | 网关健康检查，成功返回 HTTP 200。 |
| `GET /config.json` | 返回前端运行配置，目前包含 `mock_origin`。 |

## 协议定义

完整字段类型、响应结构和 OpenAPI 定义以以下仓库文件为准：

- `protocols/platform/identity/v1/service.proto`、`service.swagger.json`
- `protocols/testing/core/v1/service.proto`、`service.swagger.json`
- `protocols/testing/execution/v1/service.proto`、`service.swagger.json`
- `protocols/testing/mock/v1/service.proto`、`service.swagger.json`
- `protocols/testing/contracts/v1/testing.proto`（资源定义、执行结果与 Mock 响应）

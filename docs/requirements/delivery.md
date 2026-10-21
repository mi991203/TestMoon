# 实现与验证记录

验证日期：2026-10-21。本记录描述当前代码及本地真实联调，不作为生产容量或安全审计结论。按要求未新增单元测试，未执行go test。

## 实现范围

| 需求 | 当前实现 |
| --- | --- |
| 共享平台 | 密码登录/退出/改密、账号创建/重置/禁用、组织与成员、平台/组织/项目角色、同域会话与CSRF |
| 测试资源 | HTTP API、接口用例、按类型目录和名称搜索、顺序场景、引用/内联步骤、断言/提取、JS前后置、环境与密钥 |
| 调度执行 | 保存手动任务、独立快照、Kafka Outbox与去重、公共/项目私有Runner、隔离容器、并发限额、取消/超时/失联处理 |
| 结果 | 汇总与步骤、断言、请求响应、脚本日志、敏感字段与已知密钥脱敏、截断、默认30天终态清理 |
| 协作 | 项目与成员、固定响应HTTP Mock、独立Mock域名、接口实时文档、带密码/到期/撤销的文档分享 |
| 导入 | Postman2.1、OpenAPI2/3.0/3.1 JSON/YAML、预览/诊断、显式新增/更新及revision冲突控制 |
| 工程 | 根Go module、pnpm workspace/catalog/lockfile、proto及生成桩、独立服务二进制、两个Web应用、Makefile、中间件Compose、Runner/Worker镜像 |

## 构建证据

本机macOS arm64，使用Go1.27.3 toolchain、Node24.22.0、pnpm11.28.6、protoc36.2与Docker。依赖和镜像版本见架构文档、go.sum与pnpm-lock.yaml。

| 检查 | 结果 |
| --- | --- |
| pnpm install / 锁文件校验 | 通过 |
| pnpm typecheck | 门户、测试应用、JS Worker通过 |
| make build-go / go build ./... | 所有服务、协议及Runner编译通过 |
| pnpm build | 门户、测试应用、JS Worker生产构建通过 |
| make generate | 原生protoc、四个固定版本插件及仓内TS生成器通过 |
| bash -n scripts/*.sh | 通过 |
| make worker-image / make runner-image | 两个镜像构建通过；Worker镜像Chai加载通过 |
| scripts/infra.sh up | MySQL、Redis、Kafka健康；四个业务库创建与授权通过 |
| scripts/dev.sh start | 五服务healthz通过 |

首次构建中修复了数据库初始化stdin被Docker命令消耗、Monaco package exports路径以及共享包类型依赖。前端仍有大chunk体积提示，构建成功；进一步按页面切分和编辑器按语言裁剪属于后续性能工作。

## 真实业务联调

以下均通过HTTP网关、真实MySQL/Redis/Kafka验证；执行使用宿主Go Runner管理每次运行的真实Docker容器，未使用模拟调度器或伪造通过报告。

- 账号登录、HttpOnly Cookie会话、改密/管理员重置/禁用导致旧会话失效、重新启用与退出。
- 组织和项目创建、只读可读但不可写/执行/开通Runner；环境敏感值遮蔽，跨组织资源拒绝访问，平台管理员无隐式业务权，错误CSRF拒绝。
- 复核只读API Key通用value字段、URL查询密钥及嵌套JSON请求体脱敏；已有只读账号读取Mock敏感响应字段时遮蔽。
- HTTP JSON/文本/URL编码表单、特殊字符变量正确编码、Basic/Bearer/Query APIKey；同场景Cookie复用。
- 配置断言和JSON Pointer、变量提取、前后置JS、Chai、pm.sendRequest回调/Promise、脚本辅助表单请求及日志脱敏。
- 有实际断言的成功/失败、无断言标记unverified、未知pm接口报错、死循环JS按期限终止。
- 顺序场景变量传递、环境运行修改不回写、保存手动任务的运行及任务定义删除不影响历史。
- 提交后修改API，已开始运行仍使用原快照；资源旧revision更新返回冲突。
- Kafka重复发布同一Outbox事件，不重复执行，原终态/完成时间不改变；重复幂等提交返回原run。
- 公共Runner拒绝环回/内网目标；私有Runner可调用部署网络。离线节点拒绝启动、取消保留终态、总超时、租约到期变error且无自动重试/回退、撤销节点凭证拒绝后续请求。
- Mock独立Host、Query条件、固定状态/响应；不满足条件返回404。
- 分享密码校验、最新API定义、到期与撤销后拒绝访问；Postman目录/API/用例/场景预览及事务提交、预览重复使用冲突；OpenAPI三个版本JSON和YAML预览。
- 将本轮联调的终态记录设置为保留期之前，后台实际清理后返回404；非终态记录即使带旧完成时间仍保留。只修改了本轮生成的验收数据。

浏览器实际检查了登录页、门户进入测试应用、项目/API列表、请求体Monaco加载和报告汇总页面；检查时未见浏览器控制台error。未逐项自动化覆盖全部前端交互、浏览器种类或屏幕尺寸。

本地联调产生的示例组织、账号、项目、接口、节点和历史记录保留在本地数据卷；辅助联调脚本、日志、结果与凭证在被忽略的.local目录，不进入源码。初始管理员凭证仍在私有deploy/local/.env。

## 当前边界

- 首版HTTP，未实现gRPC/WS执行、文件上传、OAuth、定时/重试任务、完整Postman沙箱、外部OpenAPI引用或压测。兼容子集见接口与扩展约定。
- 项目删除仅支持清空资源且无任务、有效私有Runner、Mock和历史记录的项目，不提供跨微服务级联删除。报告/资源列表有当前数量上限，尚无大规模游标分页。
- 单机Docker隔离与关键网络拒绝路径已验证；未做恶意脚本逃逸审计、DNS全部组合、压力/故障注入矩阵或生产容量验证。
- 已构建Runner部署镜像及Compose示例；实际功能联调使用宿主Runner，未验收远端HTTPS、证书部署、容器版Runner跨网部署、Kafka生产认证/高可用。
- 本地Compose为单机开发设施；生产HTTPS、备份、监控告警、按库凭证、密钥托管和升级迁移需要部署环境单独实施。

## API测试导航与详情页面调整（2026-10-21）

- 左侧合并为“API测试”，树同时展示API目录与接口；接口树点击进入API调试页，Header切换“API调试 / 接口用例”。
- 用例按document.api_id聚合，后端支持resources的api_id筛选；不传参数继续保留原查询。关联仍在原加密文档中，无表结构变更、迁移、清表或历史数据重写。
- 历史独立用例和失效关联保留“未关联接口用例”入口，可重新关联。接口、用例、场景、环境、任务、Mock规则、接口文档和执行报告详情改为路由页面；操作参数、目录维护、成员及分享等仍保留对话框。
- URL保留组织/项目与资源ID，支持深层路由刷新、浏览器前进后退；修复直达带项目参数页面在会话加载前请求权限导致的CSRF失败。未授权项目不会自动回落到其他项目展示资源。
- 本轮真实HTTP验证：历史用例ID仍可读；按接口查询排除其他接口用例；独立用例能显示并重新关联，变量保留且revision正常递增；无效接口/资源类型校验、既有viewer只读及跨组织拒绝仍有效。
- 浏览器验证：接口树跳转、Header切换、历史用例详情刷新和返回、新建关联用例保存、场景详情、任务页面创建保存、Mock规则详情及原Header恢复。详情访问未使用资源弹窗。
- 验证命令：pnpm typecheck、pnpm --filter @testmoon/testing build、go build ./...；均通过。未新增单元测试，未重新验收执行引擎或生产部署。前端大chunk提示仍存在。

重新构建期间已打开的旧页面可能引用被替换的懒加载产物，更新后需刷新浏览器加载新版本。此次联调只新增少量本地示例用例/任务，未删除原历史数据。

## 代码分层、类型化协议与持久化调整（2026-10-21）

当前实现以本节、最新proto和[数据模型与迁移](../architecture/数据模型与迁移.md)为准。前一节的通用resources查询和无数据库迁移描述仅对应当时的导航调整。

- identity/core/execution/mock按api、cmd、conf、dao、model、logic、serializer及必要的internal/thirdparty拆分。移除按操作名分派的Handle和common Request/Reply；84个RPC各有明确Request/Response。资源类型、创建/更新、成员设置/移除、不同执行来源分别定义。
- 固定低代码字段使用contracts PB；Struct仅用于开放Schema，Value用于任意JSON期望值。导入条目使用oneof，未知固定字段返回400。前端、服务和Runner已统一迁移直接JSON结构；旧data信封不再兼容。
- 资源目录与五类定义分表；用例api_id显式索引。任务配置、Mock响应改为明确字段和对应加密片段。资源元数据与定义在事务中读写，保留版本冲突和软删除。
- 历史字段逐项核对：17个资源、2个任务、4条Mock规则（含软删除）与原加密业务定义一致；原UUID、revision、删除标记和创建/更新时间校验值保持一致。补充保留导入用例/场景内部name的迁移，未清表。重复启动迁移通过。
- 前端拆为app/features/components/hooks/services/models，通用UI和会话放web库，TS契约/命名客户端放api库；移除全局Doc/any类型。请求、脚本、断言、提取、场景、环境等编辑器独立封装。修复PB默认空字段使请求覆盖开关误显示启用的问题。
- 构建：make generate可重复生成且文件摘要不变；go build ./...、make build-go、go vet ./...、pnpm typecheck及两个Web生产构建通过；Runner/Worker镜像构建通过。Monaco相关大chunk提示仍存在。
- 真实HTTP：旧资源读取、按接口查询用例、独立CRUD、乐观锁、被引用删除保护、缺少ID/跨类型删除拒绝、租户隔离、已有viewer脱敏及禁止写入/执行、快照、任务、Mock、分享、OpenAPI及Postman类型化预览/提交通过。新Postman导入的内部名称持久化和目录/定义软删除一致性通过。
- 真实执行：已有私有Runner使用新机器协议心跳、领取和回传；run-case、debug-http、run-scenario、run-task均运行真实隔离容器并通过HTTP/JS/断言，重复幂等提交返回原run，退出后会话失效。使用已注册Runner状态，未重新创建用户或扩大成员权限。
- 浏览器：最终构建的接口树跳转、Header用例聚合、接口保存、独立用例详情及未启用覆盖选项正常；未逐项重新验收全部页面交互和浏览器尺寸。

按要求未新增单元测试，未运行go test。辅助联调代码与日志仍在忽略的.local目录。生产容量、远端部署和恶意脚本安全审计仍属前述未验证范围。

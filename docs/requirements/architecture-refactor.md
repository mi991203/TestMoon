# 代码分层与类型化契约调整

范围：在已有业务和微服务边界下改善包结构、PB、接口、持久化及前端可维护性，不增加新业务能力，不新增单元测试。

- 后端参考CodeMatrix的api、cmd、conf、dao、internal、logic、serializer、thirdparty分层；数据库实体单独放model。按实际职责创建包，共享依赖仍在根go.mod。
- 每种资源及每种操作使用独立Request/Response和命名RPC；固定字段明确建模，Struct/Value仅承载动态Schema或任意JSON期望值。
- 创建与更新、成员设置与移除、公共与私有Runner操作分别定义；执行接口按HTTP调试、接口用例、场景和任务区分。
- 数据库定义明确固定字段、资源归属与索引，保持soft_delete.DeletedAt；历史资源、任务和Mock定义通过兼容迁移保留。
- 前端按app、features、components、hooks、services、models组织，封装共享编辑器；API类型和命名客户端从PB生成，页面不使用通用data信封或操作名分派。
- 保持API树、API调试/接口用例Header切换、页面详情及org/project路由上下文。

验收：协议可重复生成；全部Go编译、前端类型/生产构建通过；历史迁移幂等且数据保留；主要CRUD、导入、权限、独立执行API、Runner及浏览器页面可联通。验证记录见delivery.md。

# TestMoon 首版需求基线

本目录集中管理全部业务需求；技术方案在 [architecture](../architecture/项目结构与技术选型.md)，功能状态与联调证据在 [delivery.md](delivery.md)。需求编号在后续变更中保持稳定。

- [共享平台](shared-platform.md)
- [测试业务](testing-platform.md)
- [执行与 Runner](runner-and-execution.md)
- [导入与脚本兼容](import-and-script-compatibility.md)
- [验收](acceptance.md)
- [代码分层与类型化契约调整](architecture-refactor.md)

首版为 SaaS Web 应用，组织即租户，组织下管理项目。需求在 2026-10-21 会话中逐项确认。实现不新增单元测试，验证采用编译、构建和业务联调。

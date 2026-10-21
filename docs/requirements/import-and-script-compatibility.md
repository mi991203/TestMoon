# 导入与脚本兼容

- IMP-001：手工、Postman Collection v2.1 JSON、OpenAPI 2.0/3.0/3.1 JSON/YAML；不支持格式明确报错。
- IMP-002：先预览并诊断，确认后提交；更新须选择既有资源及其版本，不自动覆盖。Postman 请求生成接口、用例和顺序场景，文件夹层级保留。
- IMP-003：集合、文件夹、请求脚本按来源次序保存；集合变量保留。远程/文件外部引用不自动下载。
- JS-001：支持变量、request/response、pm.test、pm.expect、pm.sendRequest、日志；支持 Promise 和回调式额外 HTTP。
- JS-002：不提供文件、数据库、任意 require/import、Vault、Visualizer、执行跳转或迭代数据集。未知接口运行时明确失败；静态检查不承诺识别所有动态调用。
- JS-003：变量优先级为运行局部 > 当前用例 > 环境 > 集合 > 运行全局。pm.environment.set 和 pm.globals.set 均不回写保存配置。

兼容矩阵细节在技术设计中列出；不承诺完全兼容 Postman。

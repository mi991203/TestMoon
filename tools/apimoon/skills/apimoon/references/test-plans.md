# TestPlan 与测试设计

以 `apimoon plan schema` 输出的 `data` 为准；不要复制 CLI 的 `ok/data` 信封作为计划输入。

## 最小执行计划

示例使用被测服务的 `/healthz`；实际生成时替换环境地址、路由和断言。Runner 默认值来自当前 CLI 配置。

```json
{
  "version": 1,
  "plan_id": "health-check-v1",
  "entries": [
    {
      "key": "env.local",
      "kind": "environment",
      "name": "本机测试环境",
      "document": {"variables": {"base_url": {"value": "http://127.0.0.1:8096"}}}
    },
    {
      "key": "api.health",
      "kind": "api",
      "name": "健康检查接口",
      "document": {
        "protocol": "http",
        "request": {"method": "GET", "url": "{{base_url}}/healthz"}
      }
    },
    {
      "key": "case.health",
      "kind": "case",
      "name": "健康检查返回200",
      "document": {
        "api_id": "ref:api.health",
        "assertions": [{"source": "status", "operator": "eq", "expected": 200}]
      }
    },
    {
      "key": "task.health",
      "kind": "task",
      "name": "健康检查任务",
      "document": {
        "environment_id": "ref:env.local",
        "items": [{"kind": "case", "id": "ref:case.health", "node_id": "health-instance"}]
      }
    }
  ]
}
```

## 字段和引用

- `kind` 支持 `api`、`case`、`scenario`、`task`、`environment`、`folder`，使用 snake_case 字段。
- HTTP `body` 为字符串；JSON 请求使用 `body_type: "json"` 和序列化后的 JSON 字符串。headers/query 为字符串映射。
- `ref:<key>` 只在资源引用字段替换：folder_id、api_id、environment_id、场景 case_id 和任务选择的 id。Runner ID 直接取平台节点，不是计划条目。
- 请求体、脚本和普通变量里的 `ref:...` 是普通文本；请求变量使用 `{{variable}}`。
- JSON 响应路径为 JSON Pointer，例如 `/id`、`/items/0/name`，不是 `$.id`。
- 环境变量采用 `{"value":"...","secret":false}`；用例变量是字符串映射。
- 更新条目提供 `id`、正整数 `revision`、`name` 和完整 `document`；单资源更新也不做局部合并。

## 场景中的依赖与清理

创建班级时，可按代码设计“创建学校 → 提取 school_id → 创建班级 → 提取 class_id → 查询班级 → 删除学校 → 确认班级返回404”。预期状态码依据被测服务实际协议；示例学校服务创建成功为 200。

```json
{
  "key": "scenario.class",
  "kind": "scenario",
  "name": "创建班级场景",
  "document": {
    "steps": [
      {"id": "setup", "case_id": "ref:case.school-create"},
      {"id": "create", "case_id": "ref:case.class-create"},
      {"id": "verify", "case_id": "ref:case.class-get"},
      {"id": "cleanup", "case_id": "ref:case.school-delete"}
    ]
  }
}
```

这是场景条目片段，需将引用的接口用例补全后才是可校验的计划。用例提取配置示例：

```json
{"extractions": [{"source": "json", "path": "/id", "variable": "school_id"}]}
```

同一场景的步骤可使用已提取变量。目录节点 `kind: "folder"` 只分组，`children` 包含子节点；可执行步骤为叶子。任务中重复引用同一资源时使用不同 `node_id`。

前后置操作格式为 `{"type":"javascript","version":1,"script":"..."}`。例如在 `after` 中进行动态比较：

```javascript
pm.test('班级属于本次创建的学校', () => {
  pm.expect(pm.response.json().school_id).to.equal(pm.variables.get('school_id'));
});
```

若一个负向测试意外创建成功，应记录该资源并清理，避免只根据预期 400 而忽略实际写入。清理失败应单独列入结果。

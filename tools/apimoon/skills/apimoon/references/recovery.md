# 同步与执行恢复

默认输出为 `{"ok":true,"data":...}`；错误包含 `error.code/message`，也可能同时包含部分 `data`。

| 退出码 | 处理 |
| --- | --- |
| 1 | 检查本地、网络或平台错误；写入结果是否确定要结合 error.code 和状态日志判断 |
| 2 | 修正参数、配置或计划协议后再校验 |
| 3 | 检查登录、当前作用域和项目权限 |
| 4 | 检查版本冲突、同步锁或结果不明，禁止盲目重试创建 |
| 5 | 等待超时，继续查询原执行 ID |
| 6 | 执行结束未通过，读取报告分析 |

## 同步状态

配置同级 `state/` 中的绑定按平台、登录用户、组织、项目和 plan_id 隔离。保留计划文件、稳定的 key、配置及状态目录；换机器或换登录用户不保证沿用原绑定。

同一计划重复同步会跳过相同内容，按绑定 ID 和 revision 更新变化的内容。创建条目的默认环境、Runner 来自配置，改变默认值可能改变下次同步的文档。

### 版本冲突

先读取当前计划状态和对应平台资源：

```bash
apimoon plan status --file test-plan.json
apimoon case get --id <用例ID>
```

与计划和已记录 revision 比较，指出人工修改与本次需求之间的差异。明确选择保留或更新后再调整计划；不要修改状态中的 revision 或自动使用最新 revision 强行覆盖。

### 写入结果不明

网络失败、响应丢失或写后读取失败可能留下 `pending`。已成功条目保留；不要重建 plan_id、删除状态或按名称猜测绑定。

先通过平台列表、详情或 UI 核实结果，确认目标 ID，随后使用写入时的计划和默认配置：

```bash
apimoon plan bind --file test-plan.json --key <条目key> --id <已核实的平台ID>
apimoon plan sync --file test-plan.json
```

`bind` 会验证平台内容与待确认写入的摘要一致，更新只能绑定原平台 ID。没有可核实的目标或校验不一致时，保留证据并报告阻塞原因。

锁文件可能表示另一个同步进程正在运行；异常退出后，先确认没有同步进程，再处理错误中列出的锁路径。保留状态 JSON。

## 任务提交与等待

任务提交使用事先记录的幂等键。提交响应丢失时，用相同任务 ID 和原键重试；不要生成新键造成另一次执行。

```bash
apimoon task run --id <任务ID> --idempotency-key <原键>
apimoon run wait --id <已有执行ID> --wait-timeout 5m
apimoon run get --id <已有执行ID>
```

等待超时或中断没有取消平台执行。只有用户要求取消时调用 `run cancel`；取消也不能撤销已经发出的业务请求。

任务因所选 Runner 离线被拒绝时，检查本机 Runner 进程和 `runner list` 的在线状态；不要切换到其他项目节点或公共 Runner 来绕过本机网络要求。

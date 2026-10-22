import type { ContractsAssertionResult } from "@testmoon/api";
import { active } from "../../features/reports/state";
import { useRun } from "../../hooks/queries";
import {
  Button,
  Card,
  Collapse,
  Descriptions,
  Empty,
  Space,
  Spin,
  Table,
  Tabs,
  Typography,
} from "antd";
import {
  ErrorState,
  JSONView,
  Scope,
  Status,
} from "@testmoon/web";
export function RunResult({ scope, id, onReport }: {
  scope: Scope;
  id: string;
  onReport?: () => void;
}) {
  const query = useRun(scope, id);
  const data = query.data;
  return (
    <>
      <Card
        title="执行结果"
        extra={onReport && <Button onClick={onReport}>查看完整报告</Button>}
        className="run-result"
      >
        {query.isPending && <Spin tip="正在获取执行结果" />}
        <ErrorState error={query.error} />
        {data && (
          <>
            <Descriptions
              bordered
              size="small"
              column={2}
              items={[
                { key: "name", label: "名称", children: data.run.name },
                {
                  key: "status",
                  label: "状态",
                  children: <Status value={data.run.state} />,
                },
                { key: "id", label: "执行ID", children: data.run.id },
                { key: "duration", label: "执行耗时", children: data.result ? `${data.result.duration_ms} ms` : "—" },
                {
                  key: "time",
                  label: "提交时间",
                  children: new Date(data.run.created_at).toLocaleString(),
                },
                {
                  key: "error",
                  label: "说明",
                  children: data.run.error || data.result?.error || "—",
                },
              ]}
            />
            {!data.result ? (
              <Empty
                style={{ margin: 60 }}
                description={active(data.run.state) ? "等待 Runner 返回执行结果" : "执行已结束，暂无响应结果"}
              />
            ) : (
              (data.result.items || []).map((item) => (
                <Card
                  style={{ marginTop: 20 }}
                  key={item.id}
                  title={
                    <Space>
                      {item.name}
                      <Status value={item.status} />
                    </Space>
                  }
                >
                  <Collapse
                    defaultActiveKey={item.steps?.length === 1 ? [item.steps[0].id] : []}
                    items={(item.steps || []).map((step) => ({
                      key: step.id,
                      label: (
                        <Space>
                          {step.name}
                          <Status value={step.status} />
                          <Typography.Text type="secondary">
                            {step.duration_ms} ms
                          </Typography.Text>
                        </Space>
                      ),
                      children: (
                        <>
                          <Typography.Text type="danger">
                            {step.error}
                          </Typography.Text>
                          <Tabs
                            defaultActiveKey="response"
                            items={[
                              {
                                key: "assert",
                                label: `断言 (${step.assertions?.length || 0})`,
                                children: (
                                  <Table<ContractsAssertionResult>
                                    pagination={false}
                                    rowKey={(_, i) => String(i)}
                                    dataSource={step.assertions || []}
                                    columns={[
                                      { title: "断言", dataIndex: "name" },
                                      {
                                        title: "结果",
                                        dataIndex: "passed",
                                        render: (v) => (v ? "通过" : "失败"),
                                      },
                                      { title: "说明", dataIndex: "message" },
                                    ]}
                                  />
                                ),
                              },
                              {
                                key: "request",
                                label: "请求",
                                children: <JSONView value={step.request} />,
                              },
                              {
                                key: "response",
                                label: `响应 ${step.response?.status || ""}${step.response?.truncated ? "（已截断）" : ""}`,
                                children: (
                                  <>
                                    <JSONView value={step.response?.body} />
                                    <Collapse
                                      ghost
                                      items={[{
                                        key: "headers",
                                        label: "响应头",
                                        children: <JSONView value={step.response?.headers} />,
                                      }]}
                                    />
                                  </>
                                ),
                              },
                              {
                                key: "logs",
                                label: "脚本日志",
                                children: (
                                  <JSONView
                                    value={(step.logs || []).join("\n")}
                                  />
                                ),
                              },
                            ]}
                          />
                        </>
                      ),
                    }))}
                  />
                </Card>
              ))
            )}
          </>
        )}
      </Card>
    </>
  );
}

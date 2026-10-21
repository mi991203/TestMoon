import type { ContractsAssertionResult } from "@testmoon/api";
import { active } from "./state";
import { execution } from "@testmoon/api";

import { useQuery } from "@tanstack/react-query";
import {
  Button,
  Card,
  Collapse,
  Descriptions,
  Empty,
  Space,
  Table,
  Tabs,
  Typography,
} from "antd";
import {
  ErrorState,
  JSONView,
  PageHeading,
  Scope,
  Status,
} from "@testmoon/web";
export function ReportDetail({
  scope,
  id,
  onBack,
}: {
  scope: Scope;
  id: string;
  onBack: () => void;
}) {
  const query = useQuery({
    queryKey: ["run", scope.org_id, scope.project_id, id],
    enabled: !!id,
    queryFn: () => execution.getRun({ ...scope, id: id }),
    refetchInterval: (q) =>
      active(q.state.data?.run?.state || "") ? 2000 : false,
  });
  const data = query.data;
  return (
    <>
      <PageHeading
        title="执行详情"
        extra={<Button onClick={onBack}>返回报告列表</Button>}
      />
      <Card>
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
                description="等待 Runner 返回执行结果"
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
                                    <JSONView value={step.response?.headers} />
                                    <JSONView value={step.response?.body} />
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

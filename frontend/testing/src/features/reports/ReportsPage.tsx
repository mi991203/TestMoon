import type { ExecutionRun } from "@testmoon/api";
import { execution } from "@testmoon/api";

import { useQuery } from "@tanstack/react-query";
import { Button, Card, Space, Statistic, Table } from "antd";
import {
  ErrorState,
  PageHeading,
  Scope,
  Status,
  useAction,
} from "@testmoon/web";
import { active } from "./state";
export function Reports({
  scope,
  role,
  onOpen,
}: {
  scope: Scope;
  role: string;
  onOpen: (id: string) => void;
}) {
  const query = useQuery({
    queryKey: ["runs", scope.org_id, scope.project_id],
    queryFn: () => execution.listRuns({ ...scope }),
    refetchInterval: 5000,
  });
  const { run } = useAction();
  const items: ExecutionRun[] = query.data?.items || [];
  return (
    <>
      <PageHeading
        title="执行报告"
        description="每次执行独立快照，结果默认保留30天。"
      />
      <div className="stat-grid">
        {[
          ["最近执行", items.length],
          ["执行中", items.filter((v) => active(v.state)).length],
          ["通过", items.filter((v) => v.state === "passed").length],
          ["未校验", items.filter((v) => v.state === "unverified").length],
        ].map(([title, value]) => (
          <Card key={String(title)}>
            <Statistic title={title} value={value} />
          </Card>
        ))}
      </div>
      <Card>
        <ErrorState error={query.error} />
        <Table<ExecutionRun>
          rowKey="id"
          loading={query.isLoading}
          dataSource={items}
          pagination={{ pageSize: 15 }}
          columns={[
            {
              title: "执行名称",
              dataIndex: "name",
              render: (name, row) => (
                <a onClick={() => onOpen(row.id)}>{name}</a>
              ),
            },
            {
              title: "状态",
              dataIndex: "state",
              render: (value) => <Status value={value} />,
            },
            {
              title: "Runner",
              dataIndex: "target_runner_id",
              render: (v) => (v === "public" ? "公共Runner" : v),
            },
            {
              title: "提交时间",
              dataIndex: "created_at",
              render: (v) => new Date(v).toLocaleString(),
            },
            {
              title: "操作",
              render: (_, row) => (
                <Space>
                  <Button size="small" onClick={() => onOpen(row.id)}>
                    详情
                  </Button>
                  {active(row.state) && role !== "viewer" && (
                    <Button
                      size="small"
                      danger
                      onClick={() =>
                        run(
                          () => execution.cancelRun({ ...scope, id: row.id }),
                          "已请求取消",
                        )
                      }
                    >
                      取消
                    </Button>
                  )}
                </Space>
              ),
            },
          ]}
        />
      </Card>
    </>
  );
}

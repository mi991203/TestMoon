import type { ExecutionTask } from "@testmoon/api";
import { saveTask } from "../../services/workspace";
import { execution } from "@testmoon/api";
import { useEffect } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  Button,
  Card,
  Form,
  Input,
  InputNumber,
  Popconfirm,
  Select,
  Space,
  Spin,
  Table,
} from "antd";
import { ErrorState, PageHeading, Scope, useAction } from "@testmoon/web";
import { useAssets, useRunners } from "../../hooks/queries";
import { useParams } from "react-router-dom";
import { useScopedNavigate } from "../../hooks/navigation";
import { TaskItemsFields, taskNodes, type TaskNode } from "./TaskItemsFields";
import { leafNodes } from "../../components/orchestration/tree";

export function Tasks({
  scope,
  role,
  onRun,
}: {
  scope: Scope;
  role: string;
  onRun: (id: string) => void;
}) {
  const query = useQuery({
    queryKey: ["tasks", scope.org_id, scope.project_id],
    queryFn: () => execution.listTasks({ ...scope }),
  });
  const cases = useAssets(scope, "case");
  const scenarios = useAssets(scope, "scenario");
  const envs = useAssets(scope, "environment");
  const runners = useRunners(scope);
  const { id } = useParams();
  const go = useScopedNavigate();
  const row: Partial<ExecutionTask> | null =
    id === "new"
      ? {}
      : (query.data?.items || []).find((v) => v.id === id) || null;
  const setRow = (v: Partial<ExecutionTask> | null) =>
    go(v ? `/tasks/${v.id || "new"}` : "/tasks");
  const [form] = Form.useForm();
  const { busy, run } = useAction();
  const editable = role === "admin" || role === "editor";
  useEffect(() => {
    if (!id || query.isPending || !row) return;
    form.resetFields();
    form.setFieldsValue(
      row?.id
        ? {
            ...row.document,
            name: row.name,
            items: taskNodes(row.document!.items || []),
          }
        : { items: [], runner_id: "public", timeout_seconds: 900 },
    );
  }, [id, row?.id, row?.revision, query.isPending, form]);
  if (id)
    return (
      <>
        <PageHeading
          title={id === "new" ? "新建任务" : row?.name || "任务详情"}
          extra={
            <Space>
              <Button onClick={() => setRow(null)}>返回任务列表</Button>
              {editable && row && (
                <Button
                  type="primary"
                  loading={busy}
                  onClick={() => form.submit()}
                >
                  保存
                </Button>
              )}
            </Space>
          }
        />
        <ErrorState error={query.error} />
        {query.isPending ? (
          <Spin />
        ) : !row ? (
          <ErrorState error={new Error("任务不存在或已删除")} />
        ) : (
          <Card>
            <Form
              form={form}
              layout="vertical"
              disabled={!editable}
              onFinish={(v) =>
                run(async () => {
                  const { items, name, ...rest } = v;
                  await saveTask(
                    scope,
                    {
                      name,
                      revision: row?.revision || 0,
                      document: {
                        ...rest,
                        items,
                      },
                    },
                    row?.id,
                  );
                  setRow(null);
                })
              }
            >
              <Form.Item
                name="name"
                label="名称"
                rules={[{ required: true, max: 128 }]}
              >
                <Input />
              </Form.Item>
              <Form.Item
                name="items"
                label="测试项编排"
                rules={[
                  {
                    validator: (_, nodes: TaskNode[]) =>
                      leafNodes(nodes || []).length
                        ? Promise.resolve()
                        : Promise.reject(new Error("请至少添加一个测试项")),
                  },
                ]}
              >
                <TaskItemsFields
                  cases={cases.data?.items || []}
                  scenarios={scenarios.data?.items || []}
                  readOnly={!editable}
                />
              </Form.Item>
              <Form.Item name="environment_id" label="环境">
                <Select
                  allowClear
                  options={(envs.data?.items || []).map((v) => ({
                    value: v.id,
                    label: v.name,
                  }))}
                />
              </Form.Item>
              <Form.Item
                name="runner_id"
                label="Runner"
                rules={[{ required: true }]}
              >
                <Select
                  options={(runners.data?.items || []).map((v) => ({
                    value: v.id,
                    label: `${v.name} · ${v.online ? "在线" : "离线"}`,
                  }))}
                />
              </Form.Item>
              <Form.Item name="timeout_seconds" label="执行总超时（秒）">
                <InputNumber min={1} max={3600} />
              </Form.Item>
            </Form>
          </Card>
        )}
      </>
    );
  return (
    <>
      <PageHeading
        title="测试任务"
        description="通过目录树编排测试项，配置环境与 Runner 后执行。"
        extra={
          editable && (
            <Button type="primary" onClick={() => setRow({})}>
              新建任务
            </Button>
          )
        }
      />
      <ErrorState error={query.error} />
      <Card>
        <Table<ExecutionTask>
          rowKey="id"
          dataSource={query.data?.items || []}
          columns={[
            { title: "名称", dataIndex: "name" },
            {
              title: "测试项",
              render: (_, v) => leafNodes(v.document.items || []).length,
            },
            { title: "版本", dataIndex: "revision" },
            {
              title: "操作",
              render: (_, v) => (
                <Space>
                  <Button onClick={() => setRow(v)}>
                    {editable ? "编辑" : "查看"}
                  </Button>
                  {editable && (
                    <>
                      <Button
                        type="primary"
                        ghost
                        onClick={() =>
                          run(async () => {
                            const out = await execution.runTask({
                              ...scope,
                              task_id: v.id,
                              ...{ idempotency_key: crypto.randomUUID() },
                            });
                            onRun(out.run.id);
                          }, "执行已提交")
                        }
                      >
                        运行
                      </Button>
                      <Popconfirm
                        title="删除任务定义？历史执行保留。"
                        onConfirm={() =>
                          run(
                            () => execution.deleteTask({ ...scope, id: v.id }),
                            "已删除",
                          )
                        }
                      >
                        <Button danger>删除</Button>
                      </Popconfirm>
                    </>
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

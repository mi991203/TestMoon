import type { MockDraft } from "../../models/editors";
import type { JsonValue } from "@testmoon/api";

import type { MockRule, MockEndpoint } from "@testmoon/api";
import { saveRule } from "../../services/workspace";
import { mock } from "@testmoon/api";
import { useEffect, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  Button,
  Card,
  Form,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Select,
  Space,
  Spin,
  Table,
  Tabs,
  Typography,
} from "antd";
import {
  Code,
  ErrorState,
  PageHeading,
  Pairs,
  Scope,
  useAction,
} from "@testmoon/web";

import { useParams } from "react-router-dom";
import { useScopedNavigate } from "../../hooks/navigation";

export function Mocks({ scope, role }: { scope: Scope; role: string }) {
  const endpoints = useQuery({
    queryKey: ["mock-endpoints", scope.org_id, scope.project_id],
    queryFn: () => mock.listEndpoints({ ...scope }),
  });
  const rules = useQuery({
    queryKey: ["mock-rules", scope.org_id, scope.project_id],
    queryFn: () => mock.listRules({ ...scope }),
  });
  const config = useQuery({
    queryKey: ["runtime-config"],
    queryFn: () => fetch("/config.json").then((r) => r.json()),
  });
  const [endpointOpen, setEndpointOpen] = useState(false);
  const { id } = useParams();
  const go = useScopedNavigate();
  const row: Partial<MockRule> | null =
    id === "new"
      ? {}
      : (rules.data?.items || []).find((v) => v.id === id) || null;
  const setRow = (v: Partial<MockRule> | null) =>
    go(v ? `/mock/${v.id || "new"}` : "/mock");
  const [doc, setDoc] = useState<MockDraft>({});
  const [loadedId, setLoadedId] = useState("");
  const [form] = Form.useForm();
  const { busy, run } = useAction();
  const editable = role === "admin" || role === "editor";
  useEffect(() => {
    if (!id || rules.isPending || !row) return;
    setDoc(
      row?.id
        ? structuredClone(row.document!)
        : {
            status: 200,
            headers: { "Content-Type": "application/json" },
            body: "{}",
            delay_ms: 0,
            conditions: [],
          },
    );
    setLoadedId(id);
  }, [id, row?.id, rules.isPending]);
  useEffect(() => {
    if (!row || loadedId !== id) return;
    form.resetFields();
    form.setFieldsValue(
      row.id ? row : { method: "GET", path: "/", priority: 0 },
    );
  }, [loadedId, id, row?.id, form]);
  if (id)
    return (
      <>
        <PageHeading
          title={id === "new" ? "新建Mock规则" : row?.name || "Mock规则"}
          extra={
            <Space>
              <Button onClick={() => setRow(null)}>返回Mock列表</Button>
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
        <ErrorState error={rules.error} />
        {rules.isPending ? (
          <Spin />
        ) : !row ? (
          <ErrorState error={new Error("Mock规则不存在或已删除")} />
        ) : loadedId !== id ? (
          <Spin />
        ) : (
          <Card>
            <Form
              form={form}
              layout="vertical"
              disabled={!editable}
              onFinish={(v) =>
                run(async () => {
                  await saveRule(scope, { ...v, document: doc }, row?.id);
                  setRow(null);
                })
              }
            >
              <Form.Item name="name" label="名称" rules={[{ required: true }]}>
                <Input />
              </Form.Item>
              <Form.Item
                name="endpoint_id"
                label="端点"
                rules={[{ required: true }]}
              >
                <Select
                  options={(endpoints.data?.items || []).map((v) => ({
                    value: v.id,
                    label: v.name,
                  }))}
                />
              </Form.Item>
              <Space>
                <Form.Item name="method" label="方法">
                  <Select
                    style={{ width: 130 }}
                    options={[
                      "GET",
                      "POST",
                      "PUT",
                      "PATCH",
                      "DELETE",
                      "HEAD",
                      "OPTIONS",
                    ].map((value) => ({ value }))}
                  />
                </Form.Item>
                <Form.Item
                  name="path"
                  label="相对路径"
                  rules={[{ required: true }]}
                >
                  <Input placeholder="/users" />
                </Form.Item>
                <Form.Item name="priority" label="优先级">
                  <InputNumber />
                </Form.Item>
              </Space>
            </Form>
            <Tabs
              items={[
                {
                  key: "response",
                  label: "固定响应",
                  children: (
                    <Space direction="vertical" style={{ width: "100%" }}>
                      <Space>
                        状态码
                        <InputNumber
                          disabled={!editable}
                          min={200}
                          max={599}
                          value={doc.status}
                          onChange={(status) =>
                            setDoc({ ...doc, status: status ?? 200 })
                          }
                        />
                        延迟(ms)
                        <InputNumber
                          disabled={!editable}
                          min={0}
                          max={30000}
                          value={doc.delay_ms}
                          onChange={(delay_ms) =>
                            setDoc({ ...doc, delay_ms: delay_ms ?? 0 })
                          }
                        />
                      </Space>
                      <Pairs
                        key={`header-${row?.id || "new"}`}
                        value={doc.headers}
                        readOnly={!editable}
                        onChange={(headers) => setDoc({ ...doc, headers })}
                      />
                      <Code
                        value={doc.body || ""}
                        onChange={(body) => setDoc({ ...doc, body })}
                        readOnly={!editable}
                      />
                    </Space>
                  ),
                },
                {
                  key: "conditions",
                  label: "请求匹配条件",
                  children: (
                    <Space direction="vertical" style={{ width: "100%" }}>
                      {(doc.conditions || []).map((c, i: number) => (
                        <Space.Compact key={i} style={{ width: "100%" }}>
                          <Select
                            disabled={!editable}
                            style={{ width: 130 }}
                            value={c.source}
                            options={[
                              { value: "query", label: "Query" },
                              { value: "header", label: "Header" },
                              { value: "body", label: "JSON请求体" },
                            ]}
                            onChange={(source) =>
                              setDoc({
                                ...doc,
                                conditions: (doc.conditions || []).map(
                                  (v, j: number) =>
                                    i === j ? { ...v, source } : v,
                                ),
                              })
                            }
                          />
                          <Input
                            disabled={!editable}
                            placeholder="字段名 / JSON Pointer"
                            value={c.path}
                            onChange={(e) =>
                              setDoc({
                                ...doc,
                                conditions: (doc.conditions || []).map(
                                  (v, j: number) =>
                                    i === j
                                      ? { ...v, path: e.target.value }
                                      : v,
                                ),
                              })
                            }
                          />
                          <Select
                            disabled={!editable}
                            style={{ width: 110 }}
                            value={c.operator}
                            options={["eq", "contains", "exists", "regex"].map(
                              (value) => ({ value }),
                            )}
                            onChange={(operator) =>
                              setDoc({
                                ...doc,
                                conditions: (doc.conditions || []).map(
                                  (v, j: number) =>
                                    i === j ? { ...v, operator } : v,
                                ),
                              })
                            }
                          />
                          <Input
                            disabled={!editable}
                            placeholder="预期：JSON或文本"
                            value={
                              typeof c.expected === "string"
                                ? c.expected
                                : JSON.stringify(c.expected)
                            }
                            onChange={(e) => {
                              let expected: JsonValue = e.target.value;
                              try {
                                expected = JSON.parse(expected);
                              } catch {}
                              setDoc({
                                ...doc,
                                conditions: (doc.conditions || []).map(
                                  (v, j: number) =>
                                    i === j ? { ...v, expected } : v,
                                ),
                              });
                            }}
                          />
                          <Button
                            disabled={!editable}
                            onClick={() =>
                              setDoc({
                                ...doc,
                                conditions: (doc.conditions || []).filter(
                                  (_, j: number) => i !== j,
                                ),
                              })
                            }
                          >
                            移除
                          </Button>
                        </Space.Compact>
                      ))}
                      {editable && (
                        <Button
                          type="dashed"
                          onClick={() =>
                            setDoc({
                              ...doc,
                              conditions: [
                                ...(doc.conditions || []),
                                {
                                  source: "query",
                                  path: "",
                                  operator: "eq",
                                  expected: "",
                                },
                              ],
                            })
                          }
                        >
                          添加条件
                        </Button>
                      )}
                    </Space>
                  ),
                },
              ]}
            />
          </Card>
        )}
      </>
    );
  return (
    <>
      <PageHeading
        title="HTTP Mock"
        description="匿名调用、独立域名。规则按优先级从高到低匹配，条件全部满足才返回。"
        extra={
          editable && (
            <Space>
              <Button onClick={() => setEndpointOpen(true)}>新建端点</Button>
              <Button type="primary" onClick={() => setRow({})}>
                新建规则
              </Button>
            </Space>
          )
        }
      />
      <Tabs
        items={[
          {
            key: "rules",
            label: "响应规则",
            children: (
              <Card>
                <ErrorState error={rules.error} />
                <Table<MockRule>
                  rowKey="id"
                  dataSource={rules.data?.items || []}
                  columns={[
                    { title: "名称", dataIndex: "name" },
                    {
                      title: "请求",
                      render: (_, v) => `${v.method} ${v.path}`,
                    },
                    { title: "优先级", dataIndex: "priority" },
                    {
                      title: "操作",
                      render: (_, v) => (
                        <Space>
                          <Button onClick={() => setRow(v)}>
                            {editable ? "编辑" : "查看"}
                          </Button>
                          {editable && (
                            <Popconfirm
                              title="删除规则？"
                              onConfirm={() =>
                                run(
                                  () => mock.deleteRule({ ...scope, id: v.id }),
                                  "已删除",
                                )
                              }
                            >
                              <Button danger>删除</Button>
                            </Popconfirm>
                          )}
                        </Space>
                      ),
                    },
                  ]}
                />
              </Card>
            ),
          },
          {
            key: "endpoints",
            label: "访问端点",
            children: (
              <Card>
                <ErrorState error={endpoints.error} />
                <Table<MockEndpoint>
                  rowKey="id"
                  dataSource={endpoints.data?.items || []}
                  columns={[
                    { title: "名称", dataIndex: "name" },
                    {
                      title: "基础地址",
                      render: (_, v) => (
                        <Typography.Text
                          copyable
                        >{`${config.data?.mock_origin || ""}/m/${v.id}`}</Typography.Text>
                      ),
                    },
                    {
                      title: "操作",
                      render: (_, v) =>
                        editable ? (
                          <Popconfirm
                            title="删除端点及其全部规则？"
                            onConfirm={() =>
                              run(
                                () =>
                                  mock.deleteEndpoint({ ...scope, id: v.id }),
                                "已删除",
                              )
                            }
                          >
                            <Button danger>删除</Button>
                          </Popconfirm>
                        ) : null,
                    },
                  ]}
                />
              </Card>
            ),
          },
        ]}
      />
      <Modal
        title="Mock端点"
        open={endpointOpen}
        onCancel={() => setEndpointOpen(false)}
        footer={null}
        destroyOnHidden
      >
        <Form
          layout="vertical"
          onFinish={(v) =>
            run(async () => {
              await mock.createEndpoint({ ...scope, ...v });
              setEndpointOpen(false);
            })
          }
        >
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Button type="primary" htmlType="submit" loading={busy}>
            创建匿名端点
          </Button>
        </Form>
      </Modal>
    </>
  );
}

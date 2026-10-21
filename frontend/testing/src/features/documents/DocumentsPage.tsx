import type { Asset } from "../../models/resources";
import type { CoreShare } from "@testmoon/api";

import { core } from "@testmoon/api";
import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  Button,
  Card,
  Form,
  Input,
  Modal,
  Popconfirm,
  Select,
  Table,
  Tabs,
  Tag,
  Typography,
} from "antd";
import { ErrorState, PageHeading, Scope, useAction } from "@testmoon/web";
import { useAssets } from "../../hooks/queries";

export function Documents({
  scope,
  role,
  onOpen,
}: {
  scope: Scope;
  role: string;
  onOpen: (id: string) => void;
}) {
  const apis = useAssets(scope, "api");
  const shares = useQuery({
    queryKey: ["shares", scope.org_id, scope.project_id],
    queryFn: () => core.listShares({ ...scope }),
  });
  const [open, setOpen] = useState(false);
  const [link, setLink] = useState("");
  const { busy, run } = useAction();
  const editable = role === "admin" || role === "editor";
  return (
    <>
      <PageHeading
        title="接口文档"
        description="文档随接口定义更新，分享仅包含选定接口。"
        extra={
          editable && (
            <Button type="primary" onClick={() => setOpen(true)}>
              创建分享
            </Button>
          )
        }
      />
      <ErrorState error={apis.error} />
      <Tabs
        items={[
          {
            key: "documents",
            label: "项目接口文档",
            children: (
              <Card>
                <Table<Asset<"api">>
                  rowKey="id"
                  dataSource={apis.data?.items || []}
                  columns={[
                    {
                      title: "接口",
                      dataIndex: "name",
                      render: (name, row) => (
                        <a onClick={() => onOpen(row.id)}>{name}</a>
                      ),
                    },
                    {
                      title: "方法",
                      render: (_, v) => <Tag>{v.document.request.method}</Tag>,
                    },
                    { title: "URL", render: (_, v) => v.document.request.url },
                  ]}
                />
              </Card>
            ),
          },
          {
            key: "shares",
            label: "分享管理",
            children: (
              <Card>
                <ErrorState error={shares.error} />
                <Table<CoreShare>
                  rowKey="id"
                  dataSource={shares.data?.items || []}
                  columns={[
                    { title: "名称", dataIndex: "name" },
                    {
                      title: "保护",
                      render: (_, v) =>
                        v.password_required ? "需要密码" : "无密码",
                    },
                    {
                      title: "有效期",
                      dataIndex: "expires_at",
                      render: (v) =>
                        v ? new Date(v).toLocaleString() : "长期",
                    },
                    {
                      title: "链接",
                      render: (_, v) =>
                        v.url ? (
                          <Typography.Text copyable>{v.url}</Typography.Text>
                        ) : (
                          "—"
                        ),
                    },
                    {
                      title: "操作",
                      render: (_, v) =>
                        editable ? (
                          <Popconfirm
                            title="撤销后链接立即不可访问，确认？"
                            onConfirm={() =>
                              run(
                                () => core.deleteShare({ ...scope, id: v.id }),
                                "已撤销",
                              )
                            }
                          >
                            <Button danger>撤销</Button>
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
        title="创建文档分享"
        open={open}
        footer={null}
        onCancel={() => setOpen(false)}
        destroyOnHidden
      >
        <Form
          layout="vertical"
          onFinish={(v) =>
            run(async () => {
              const out = await core.createShare({
                ...scope,
                ...{
                  ...v,
                  expires_at: v.expires_at
                    ? new Date(v.expires_at).toISOString()
                    : null,
                },
              });
              setLink(out.url);
              setOpen(false);
            }, "分享已创建")
          }
        >
          <Form.Item name="name" label="分享名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item
            name="api_ids"
            label="接口范围"
            rules={[{ required: true }]}
          >
            <Select
              mode="multiple"
              options={(apis.data?.items || []).map((v) => ({
                value: v.id,
                label: v.name,
              }))}
            />
          </Form.Item>
          <Form.Item name="password" label="访问密码（可选）">
            <Input.Password />
          </Form.Item>
          <Form.Item name="expires_at" label="到期时间（可选）">
            <Input type="datetime-local" />
          </Form.Item>
          <Button htmlType="submit" loading={busy} type="primary">
            创建链接
          </Button>
        </Form>
      </Modal>
      <Modal
        title="分享链接"
        open={!!link}
        onCancel={() => setLink("")}
        onOk={() => setLink("")}
      >
        <Typography.Paragraph copyable>{link}</Typography.Paragraph>
      </Modal>
    </>
  );
}

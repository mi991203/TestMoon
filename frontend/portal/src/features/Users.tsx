import type { IdentityUser } from "@testmoon/api";
import { identity } from "@testmoon/api";
import { useState } from "react";

import { useQuery } from "@tanstack/react-query";
import {
  Button,
  Checkbox,
  Form,
  Input,
  Modal,
  Popconfirm,
  Space,
  Table,
  Tag,
} from "antd";
import { ErrorState, PageHeading, useAction } from "@testmoon/web";

export function Users() {
  const query = useQuery({
    queryKey: ["users"],
    queryFn: () => identity.listUsers({}),
  });
  const [mode, setMode] = useState("");
  const [row, setRow] = useState<Partial<IdentityUser>>({});
  const [form] = Form.useForm();
  const { busy, run } = useAction();
  return (
    <>
      <PageHeading
        title="账号管理"
        description="账号由平台统一开通，组织成员关系单独管理。"
        extra={
          <Button
            type="primary"
            onClick={() => {
              setRow({});
              form.resetFields();
              setMode("create");
            }}
          >
            创建账号
          </Button>
        }
      />
      <ErrorState error={query.error} />
      <Table<IdentityUser>
        rowKey="id"
        loading={query.isLoading}
        dataSource={query.data?.items || []}
        columns={[
          { title: "账号", dataIndex: "username" },
          { title: "姓名", dataIndex: "name" },
          {
            title: "平台角色",
            render: (_, v) =>
              v.admin ? <Tag color="blue">平台管理员</Tag> : "普通账号",
          },
          {
            title: "状态",
            render: (_, v) =>
              v.disabled ? (
                <Tag color="red">已禁用</Tag>
              ) : (
                <Tag color="green">正常</Tag>
              ),
          },
          {
            title: "操作",
            render: (_, v) => (
              <Space>
                <Button
                  onClick={() => {
                    setRow(v);
                    form.resetFields();
                    setMode("reset");
                  }}
                >
                  重置密码
                </Button>
                <Popconfirm
                  title={v.disabled ? "启用此账号？" : "禁用此账号并撤销登录？"}
                  onConfirm={() =>
                    run(() =>
                      identity.setUserStatus({
                        ...{ disabled: !v.disabled },
                        id: v.id,
                      }),
                    )
                  }
                >
                  <Button danger={!v.disabled}>
                    {v.disabled ? "启用" : "禁用"}
                  </Button>
                </Popconfirm>
              </Space>
            ),
          },
        ]}
      />
      <Modal
        title={mode === "create" ? "创建账号" : "重置密码"}
        open={!!mode}
        onCancel={() => setMode("")}
        onOk={() => form.submit()}
        confirmLoading={busy}
        destroyOnHidden
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={(v) =>
            run(async () => {
              await (mode === "create"
                ? identity.createUser({ ...v, id: row.id })
                : identity.resetPassword({ ...v, id: row.id }));
              setMode("");
            })
          }
        >
          {mode === "create" && (
            <>
              <Form.Item
                name="username"
                label="账号"
                rules={[
                  { required: true, pattern: /^[a-z0-9][a-z0-9_.-]{2,63}$/ },
                ]}
              >
                <Input />
              </Form.Item>
              <Form.Item name="name" label="显示名称">
                <Input />
              </Form.Item>
              <Form.Item name="admin" valuePropName="checked">
                <Checkbox>平台管理员（不自动获取组织业务权限）</Checkbox>
              </Form.Item>
            </>
          )}
          <Form.Item
            name="password"
            label="密码"
            rules={[{ required: true, min: 10, max: 128 }]}
          >
            <Input.Password autoComplete="new-password" />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}

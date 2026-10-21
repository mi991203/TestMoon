import type {
  IdentityOrganization,
  IdentityOrganizationMember,
  IdentitySessionResponse,
} from "@testmoon/api";
import { identity } from "@testmoon/api";
import { useState } from "react";

import { useQuery } from "@tanstack/react-query";
import { Button, Card, Form, Input, Modal, Select, Table } from "antd";
import { ErrorState, PageHeading, useAction } from "@testmoon/web";

export function Organizations({
  session,
  org,
}: {
  session: IdentitySessionResponse;
  org: string;
}) {
  const query = useQuery({
    queryKey: ["organizations"],
    queryFn: () => identity.listOrganizations({}),
  });
  const users = useQuery({
    queryKey: ["users"],
    enabled: !!session.principal.admin,
    queryFn: () => identity.listUsers({}),
  });
  const members = useQuery({
    queryKey: ["org-members", org],
    enabled: !!org,
    queryFn: () => identity.listOrganizationMembers({ ...{ org_id: org } }),
  });
  const [mode, setMode] = useState("");
  const [row, setRow] = useState<Partial<IdentityOrganization>>({});
  const [form] = Form.useForm();
  const { busy, run } = useAction();
  const orgAdmin = session.principal.organizations?.[org] === "admin";
  return (
    <>
      <PageHeading
        title="组织与成员"
        description="组织即独立租户；一个账号可以加入多个组织。"
        extra={
          session.principal.admin && (
            <Button
              type="primary"
              onClick={() => {
                setMode("organization");
                setRow({});
                form.resetFields();
              }}
            >
              创建组织
            </Button>
          )
        }
      />
      <ErrorState error={query.error} />
      <Table<IdentityOrganization>
        rowKey="id"
        dataSource={query.data?.items || []}
        columns={[
          { title: "组织", dataIndex: "name" },
          { title: "组织ID", dataIndex: "id" },
          {
            title: "操作",
            render: (_, v) =>
              session.principal.admin ? (
                <Button
                  onClick={() => {
                    setMode("organization");
                    setRow(v);
                    form.setFieldsValue(v);
                  }}
                >
                  编辑名称
                </Button>
              ) : null,
          },
        ]}
      />
      {org && (
        <Card
          title="当前组织成员"
          extra={
            orgAdmin && (
              <Button
                onClick={() => {
                  setMode("member");
                  setRow({});
                  form.resetFields();
                }}
              >
                添加 / 调整成员
              </Button>
            )
          }
        >
          <ErrorState error={members.error} />
          <Table<IdentityOrganizationMember>
            rowKey="id"
            dataSource={members.data?.items || []}
            columns={[
              { title: "账号", dataIndex: "username" },
              { title: "姓名", dataIndex: "name" },
              {
                title: "角色",
                dataIndex: "role",
                render: (v) => (v === "admin" ? "组织管理员" : "组织成员"),
              },
              {
                title: "账号状态",
                render: (_, v) => (v.disabled ? "禁用" : "正常"),
              },
            ]}
          />
        </Card>
      )}
      <Modal
        open={!!mode}
        title={mode === "organization" ? "组织管理" : "组织成员"}
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
              if (mode === "organization")
                await (row.id
                  ? identity.updateOrganization({ ...v, id: row.id })
                  : identity.createOrganization({ ...v, id: row.id }));
              else
                await (v.role === "remove"
                  ? identity.removeOrganizationMember({
                      ...{ org_id: org },
                      username: v.username,
                    })
                  : identity.setOrganizationMember({
                      ...{ org_id: org },
                      ...v,
                    }));
              setMode("");
            })
          }
        >
          {mode === "organization" ? (
            <>
              <Form.Item
                name="name"
                label="组织名称"
                rules={[{ required: true }]}
              >
                <Input />
              </Form.Item>
              {!row.id && (
                <Form.Item
                  name="admin_user_id"
                  label="组织管理员"
                  rules={[{ required: true }]}
                >
                  <Select
                    showSearch
                    optionFilterProp="label"
                    options={(users.data?.items || [])
                      .filter((u) => !u.disabled)
                      .map((u) => ({ value: u.id, label: u.username }))}
                  />
                </Form.Item>
              )}
            </>
          ) : (
            <>
              <Form.Item
                name="username"
                label="已开通账号"
                rules={[{ required: true }]}
              >
                <Input placeholder="输入准确账号" />
              </Form.Item>
              <Form.Item name="role" label="角色" rules={[{ required: true }]}>
                <Select
                  options={[
                    { value: "admin", label: "组织管理员" },
                    { value: "member", label: "组织成员" },
                    { value: "remove", label: "移出组织" },
                  ]}
                />
              </Form.Item>
            </>
          )}
        </Form>
      </Modal>
    </>
  );
}

import type { CoreProjectMember } from "@testmoon/api";
import { setProjectMember } from "../../services/workspace";
import { core, identity } from "@testmoon/api";
import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Button, Card, Form, Modal, Select, Table } from "antd";
import { ErrorState, PageHeading, Scope, useAction } from "@testmoon/web";

export function Members({ scope, role }: { scope: Scope; role: string }) {
  const query = useQuery({
    queryKey: ["project-members", scope.org_id, scope.project_id],
    queryFn: () => core.listProjectMembers({ ...scope }),
  });
  const orgUsers = useQuery({
    queryKey: ["org-members", scope.org_id],
    queryFn: () =>
      identity.listOrganizationMembers({ ...{ org_id: scope.org_id } }),
  });
  const [open, setOpen] = useState(false);
  const { busy, run } = useAction();
  return (
    <>
      <PageHeading
        title="项目成员"
        description="组织管理员自动管理本组织全部项目；项目角色单独授权。"
        extra={
          role === "admin" && (
            <Button type="primary" onClick={() => setOpen(true)}>
              添加 / 调整成员
            </Button>
          )
        }
      />
      <ErrorState error={query.error} />
      <Card>
        <Table<CoreProjectMember>
          rowKey="id"
          dataSource={query.data?.items || []}
          columns={[
            {
              title: "账号",
              render: (_, v) =>
                orgUsers.data?.items?.find((u) => u.user_id === v.user_id)
                  ?.username || v.user_id,
            },
            {
              title: "角色",
              dataIndex: "role",
              render: (v) =>
                (
                  ({
                    admin: "项目管理员",
                    editor: "编辑者",
                    viewer: "只读成员",
                  }) as Record<string, string>
                )[v] || v,
            },
          ]}
        />
      </Card>
      <Modal
        title="项目成员"
        open={open}
        footer={null}
        onCancel={() => setOpen(false)}
        destroyOnHidden
      >
        <Form
          layout="vertical"
          onFinish={(v) =>
            run(async () => {
              await setProjectMember(scope, v);
              setOpen(false);
            })
          }
        >
          <Form.Item
            name="user_id"
            label="组织成员"
            rules={[{ required: true }]}
          >
            <Select
              showSearch
              optionFilterProp="label"
              options={(orgUsers.data?.items || [])
                .filter((u) => !u.disabled)
                .map((u) => ({ value: u.user_id, label: u.username }))}
            />
          </Form.Item>
          <Form.Item name="role" label="项目角色" rules={[{ required: true }]}>
            <Select
              options={[
                { value: "admin", label: "项目管理员" },
                { value: "editor", label: "编辑者" },
                { value: "viewer", label: "只读成员" },
                { value: "remove", label: "移出项目" },
              ]}
            />
          </Form.Item>
          <Button loading={busy} htmlType="submit" type="primary">
            保存
          </Button>
        </Form>
      </Modal>
    </>
  );
}

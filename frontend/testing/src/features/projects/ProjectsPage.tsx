import type { CoreProject } from "@testmoon/api";
import { saveProject } from "../../services/workspace";
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
  Space,
  Table,
} from "antd";
import { PageHeading, Scope, useAction } from "@testmoon/web";

export function Projects({
  scope,
  role,
  orgAdmin,
  onSelect,
}: {
  scope: Scope;
  role: string;
  orgAdmin: boolean;
  onSelect: (id: string) => void;
}) {
  const query = useQuery({
    queryKey: ["projects", scope.org_id],
    enabled: !!scope.org_id,
    queryFn: () => core.listProjects({ ...{ org_id: scope.org_id } }),
  });
  const [row, setRow] = useState<Partial<CoreProject> | null>(null);
  const [form] = Form.useForm();
  const { busy, run } = useAction();
  return (
    <>
      <PageHeading
        title="项目管理"
        description="项目隔离接口、环境、用例、任务与私有 Runner。"
        extra={
          orgAdmin && (
            <Button
              type="primary"
              onClick={() => {
                setRow({});
                form.resetFields();
              }}
            >
              创建项目
            </Button>
          )
        }
      />
      <Card>
        <Table<CoreProject>
          rowKey="id"
          dataSource={query.data?.items || []}
          columns={[
            {
              title: "名称",
              dataIndex: "name",
              render: (name, v) => <a onClick={() => onSelect(v.id)}>{name}</a>,
            },
            { title: "说明", dataIndex: "description" },
            {
              title: "操作",
              render: (_, v) =>
                orgAdmin || (v.id === scope.project_id && role === "admin") ? (
                  <Space>
                    <Button
                      onClick={() => {
                        setRow(v);
                        form.setFieldsValue(v);
                      }}
                    >
                      编辑
                    </Button>
                    <Popconfirm
                      title="仅允许删除无资源、任务、有效 Runner 和历史执行的空项目。"
                      onConfirm={() =>
                        run(
                          () =>
                            core.deleteProject({
                              ...{
                                org_id: scope.org_id,
                                project_id: v.id,
                              },
                            }),
                          "已删除",
                        )
                      }
                    >
                      <Button danger>删除</Button>
                    </Popconfirm>
                  </Space>
                ) : null,
            },
          ]}
        />
      </Card>
      <Modal
        title={row?.id ? "编辑项目" : "创建项目"}
        open={row !== null}
        onCancel={() => setRow(null)}
        onOk={() => form.submit()}
        confirmLoading={busy}
        destroyOnHidden
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={(v) =>
            run(async () => {
              const result = await saveProject(
                { org_id: scope.org_id },
                v,
                row?.id,
              );
              setRow(null);
              onSelect(result.project.id);
            })
          }
        >
          <Form.Item
            label="项目名称"
            name="name"
            rules={[{ required: true, max: 128 }]}
          >
            <Input />
          </Form.Item>
          <Form.Item label="说明" name="description">
            <Input.TextArea maxLength={4096} />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}

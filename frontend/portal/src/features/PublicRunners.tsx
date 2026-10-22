import type { ExecutionRunner } from "@testmoon/api";
import { execution } from "@testmoon/api";
import { useState } from "react";

import { useQuery } from "@tanstack/react-query";
import {
  Alert,
  Button,
  Form,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Table,
  Typography,
} from "antd";
import { ErrorState, PageHeading, useAction } from "@testmoon/web";

export function PublicRunners() {
  const query = useQuery({
    queryKey: ["public-runners"],
    queryFn: () => execution.listPublicRunners({}),
    refetchInterval: 10000,
  });
  const [open, setOpen] = useState(false);
  const [token, setToken] = useState("");
  const { busy, run } = useAction();
  return (
    <>
      <PageHeading
        title="公共 Runner"
        description="由平台运维部署，执行时限制为公网访问。"
        extra={
          <Button type="primary" onClick={() => setOpen(true)}>
            开通公共节点
          </Button>
        }
      />
      <ErrorState error={query.error} />
      <Table<ExecutionRunner>
        rowKey="id"
        dataSource={query.data?.items || []}
        columns={[
          { title: "名称", dataIndex: "name" },
          { title: "在线", render: (_, v) => (v.online ? "在线" : "离线") },
          { title: "并发", dataIndex: "concurrency" },
          {
            title: "操作",
            render: (_, v) => (
              <Popconfirm
                title="撤销此公共节点并停止其执行？"
                onConfirm={() =>
                  run(() => execution.revokePublicRunner({ ...{}, id: v.id }))
                }
              >
                <Button danger>撤销</Button>
              </Popconfirm>
            ),
          },
        ]}
      />
      <Modal
        title="开通公共节点"
        open={open}
        onCancel={() => setOpen(false)}
        footer={null}
        destroyOnHidden
      >
        <Form
          layout="vertical"
          initialValues={{ concurrency: 1 }}
          onFinish={(v) =>
            run(async () => {
              const result = await execution.createPublicRunner({
                ...{},
                ...v,
              });
              setToken(result.enrollment_token);
              setOpen(false);
            }, "节点已开通")
          }
        >
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="concurrency" label="并发上限">
            <InputNumber min={1} max={8} />
          </Form.Item>
          <Button type="primary" htmlType="submit" loading={busy}>
            生成一次性注册凭证
          </Button>
        </Form>
      </Modal>
      <Modal
        title="保存注册凭证"
        open={!!token}
        onCancel={() => setToken("")}
        onOk={() => setToken("")}
      >
        <Alert type="warning" message="凭证只显示一次，完成注册后失效。" />
        <Typography.Paragraph copyable style={{ marginTop: 20 }}>
          {token}
        </Typography.Paragraph>
        <Typography.Text>
          按 README 中 Runner 部署步骤启动该节点。
        </Typography.Text>
      </Modal>
    </>
  );
}

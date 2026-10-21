import type { ExecutionRunner } from "@testmoon/api";

import { execution } from "@testmoon/api";
import { useState } from "react";

import {
  Alert,
  Button,
  Card,
  Form,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Table,
  Tag,
  Typography,
} from "antd";
import { ErrorState, PageHeading, Scope, useAction } from "@testmoon/web";
import { useRunners } from "../../hooks/queries";

export function Runners({ scope, role }: { scope: Scope; role: string }) {
  const query = useRunners(scope);
  const [open, setOpen] = useState(false);
  const [token, setToken] = useState("");
  const { busy, run } = useAction();
  return (
    <>
      <PageHeading
        title="Runner 管理"
        description="项目私有节点独享；公共节点只可用于公网接口。"
        extra={
          role === "admin" && (
            <Button type="primary" onClick={() => setOpen(true)}>
              注册私有 Runner
            </Button>
          )
        }
      />
      <ErrorState error={query.error} />
      <Card>
        <Table<ExecutionRunner>
          rowKey="id"
          dataSource={query.data?.items || []}
          columns={[
            { title: "名称", dataIndex: "name" },
            {
              title: "类型",
              render: (_, v) => (v.public ? "公共池" : "项目私有"),
            },
            {
              title: "状态",
              render: (_, v) => (
                <Tag color={v.online ? "green" : "default"}>
                  {v.revoked ? "已撤销" : v.online ? "在线" : "离线"}
                </Tag>
              ),
            },
            {
              title: "最近心跳",
              dataIndex: "last_seen",
              render: (v) => (v ? new Date(v).toLocaleString() : "—"),
            },
            {
              title: "操作",
              render: (_, v) =>
                !v.public && role === "admin" && !v.revoked ? (
                  <Popconfirm
                    title="撤销后节点凭证立即失效，执行将停止。"
                    onConfirm={() =>
                      run(
                        () =>
                          execution.revokePrivateRunner({ ...scope, id: v.id }),
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
      <Modal
        title="开通私有 Runner"
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
              const result = await execution.createPrivateRunner({
                ...scope,
                ...v,
              });
              setToken(result.enrollment_token);
              setOpen(false);
            }, "Runner已开通")
          }
        >
          <Form.Item label="节点名称" name="name" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item label="并发上限" name="concurrency">
            <InputNumber min={1} max={8} />
          </Form.Item>
          <Button loading={busy} type="primary" htmlType="submit">
            生成一次性注册凭证
          </Button>
        </Form>
      </Modal>
      <Modal
        title="部署私有 Runner"
        open={!!token}
        onCancel={() => setToken("")}
        onOk={() => setToken("")}
        width={750}
      >
        <Alert
          type="warning"
          showIcon
          message="注册凭证仅显示一次；执行容器不获得此凭证。"
        />
        <Typography.Paragraph copyable style={{ marginTop: 20 }}>
          {token}
        </Typography.Paragraph>
        <Typography.Paragraph>
          平台地址：{location.origin}。按 README 的 Runner
          部署说明设置地址与上述注册凭证，挂载专用状态目录和 Docker API。仅本地
          HTTP 联调设置 TESTMOON_ALLOW_HTTP=1。
        </Typography.Paragraph>
      </Modal>
    </>
  );
}

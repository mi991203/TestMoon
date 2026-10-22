import { identity } from "@testmoon/api";

import { Button, Card, Form, Input } from "antd";
import { login, PageHeading, useAction } from "@testmoon/web";

export function Security() {
  const { busy, run } = useAction();
  return (
    <>
      <PageHeading title="账号安全" description="修改密码后需重新登录。" />
      <Card style={{ maxWidth: 520 }}>
        <Form
          layout="vertical"
          onFinish={(v) =>
            run(async () => {
              await identity.changePassword({ ...v });
              location.href = "/login";
            }, "密码已修改")
          }
        >
          <Form.Item
            name="old_password"
            label="原密码"
            rules={[{ required: true }]}
          >
            <Input.Password autoComplete="current-password" />
          </Form.Item>
          <Form.Item
            name="new_password"
            label="新密码"
            rules={[{ required: true, min: 10, max: 128 }]}
          >
            <Input.Password autoComplete="new-password" />
          </Form.Item>
          <Button type="primary" loading={busy} htmlType="submit">
            修改密码
          </Button>
        </Form>
      </Card>
    </>
  );
}

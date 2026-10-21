import { Button, Form, Input, Typography } from "antd";
import { login, useAction } from "@testmoon/web";

export function Login() {
  const { busy, run } = useAction();
  return (
    <div className="login-page">
      <div className="login-story">
        <span>◒ TESTMOON</span>
        <h1>
          连接接口。
          <br />
          验证每个可能。
        </h1>
        <p>
          在统一的组织空间里管理接口、构建测试场景，并将执行交给合适的 Runner。
        </p>
        <span className="subtle">HTTP TESTING · SCENARIOS · MOCK</span>
      </div>
      <div className="login-form">
        <div className="login-card">
          <Typography.Title level={2}>登录工作台</Typography.Title>
          <Typography.Paragraph type="secondary">
            使用管理员为你开通的账号
          </Typography.Paragraph>
          <Form
            layout="vertical"
            onFinish={(v) =>
              run(async () => {
                await login(v.username, v.password);
                location.href = "/";
              }, "登录成功")
            }
          >
            <Form.Item
              label="账号"
              name="username"
              rules={[{ required: true }]}
            >
              <Input autoComplete="username" size="large" />
            </Form.Item>
            <Form.Item
              label="密码"
              name="password"
              rules={[{ required: true }]}
            >
              <Input.Password autoComplete="current-password" size="large" />
            </Form.Item>
            <Button
              type="primary"
              block
              htmlType="submit"
              size="large"
              loading={busy}
            >
              登录
            </Button>
          </Form>
        </div>
      </div>
    </div>
  );
}

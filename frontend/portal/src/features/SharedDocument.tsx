import { core } from "@testmoon/api";
import { useState } from "react";

import { useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import {
  Button,
  Card,
  Descriptions,
  Form,
  Input,
  Spin,
  Tag,
  Typography,
} from "antd";
import { ErrorState, JSONView } from "@testmoon/web";

export function SharedDocument() {
  const { token } = useParams();
  const [password, setPassword] = useState("");
  const [attempt, setAttempt] = useState(0);
  const query = useQuery({
    queryKey: ["shared-document", token, attempt],
    queryFn: () => core.readSharedDocument({ ...{ token, password } }),
  });
  return (
    <div style={{ maxWidth: 1100, margin: "50px auto", padding: 24 }}>
      <Typography.Text type="secondary">
        TestMoon / 共享接口文档
      </Typography.Text>
      <Typography.Title>{query.data?.name || "接口文档"}</Typography.Title>
      {query.error && (
        <>
          <ErrorState error={query.error} />
          <Form
            layout="inline"
            style={{ margin: "20px 0" }}
            onFinish={() => setAttempt((x) => x + 1)}
          >
            <Input.Password
              aria-label="分享密码"
              placeholder="输入分享密码"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
            <Button type="primary" htmlType="submit">
              访问文档
            </Button>
          </Form>
        </>
      )}
      {query.isLoading && <Spin />}
      {query.data?.items?.map((item) => (
        <Card key={item.id} title={item.name} style={{ marginTop: 20 }}>
          <Tag color="blue">{item.document.request.method}</Tag>
          <Typography.Text code>{item.document.request.url}</Typography.Text>
          <Typography.Paragraph>
            {item.document.description}
          </Typography.Paragraph>
          <Descriptions
            items={[
              {
                key: "body",
                label: "请求体类型",
                children: item.document.request.body_type || "无",
              },
            ]}
          />
          <JSONView value={item.document} />
        </Card>
      ))}
    </div>
  );
}

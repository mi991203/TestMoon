import type { RequestDraft } from "../../models/editors";

import { Input, InputNumber, Select, Space, Tabs, Typography } from "antd";
import { Code, Pairs } from "@testmoon/web";
import { blankRequest } from "./defaults";
export function RequestFields({
  value,
  onChange,
  readOnly = false,
}: {
  value: RequestDraft;
  onChange: (v: RequestDraft) => void;
  readOnly?: boolean;
}) {
  const r: RequestDraft = { ...blankRequest(), ...value };
  const patch = (v: RequestDraft) => onChange({ ...r, ...v });
  return (
    <>
      <div className="request-line">
        <Select
          disabled={readOnly}
          value={r.method}
          onChange={(method) => patch({ method })}
          options={[
            "GET",
            "POST",
            "PUT",
            "PATCH",
            "DELETE",
            "HEAD",
            "OPTIONS",
          ].map((value) => ({ value }))}
        />
        <Input
          disabled={readOnly}
          aria-label="请求URL"
          placeholder="https://api.example.com/path 或 {{base_url}}/path"
          value={r.url}
          onChange={(e) => patch({ url: e.target.value })}
        />
      </div>
      <Tabs
        items={[
          {
            key: "query",
            label: "Query",
            children: (
              <Pairs
                key="query"
                value={r.query}
                readOnly={readOnly}
                onChange={(query) => patch({ query })}
              />
            ),
          },
          {
            key: "headers",
            label: "Headers",
            children: (
              <Pairs
                key="headers"
                value={r.headers}
                readOnly={readOnly}
                onChange={(headers) => patch({ headers })}
              />
            ),
          },
          {
            key: "body",
            label: "Body",
            children: (
              <>
                <Select
                  style={{ marginBottom: 12, width: 230 }}
                  disabled={readOnly}
                  value={r.body_type || "text"}
                  options={[
                    { value: "json", label: "JSON" },
                    { value: "text", label: "原始文本" },
                    { value: "form", label: "URL编码表单" },
                  ]}
                  onChange={(body_type) => patch({ body_type })}
                />
                <Code
                  language={r.body_type === "json" ? "json" : "plaintext"}
                  value={r.body || ""}
                  onChange={(body) => patch({ body })}
                  readOnly={readOnly}
                />
                {r.body_type === "form" && (
                  <Typography.Text type="secondary">
                    格式：username=alice&amp;password={"{{password}}"}
                    。值在变量替换后编码。
                  </Typography.Text>
                )}
              </>
            ),
          },
          {
            key: "auth",
            label: "认证",
            children: (
              <Space direction="vertical" style={{ width: "100%" }}>
                <Select
                  style={{ width: 220 }}
                  disabled={readOnly}
                  value={r.auth?.type || "none"}
                  options={[
                    { value: "none", label: "无认证" },
                    { value: "basic", label: "Basic" },
                    { value: "bearer", label: "Bearer Token" },
                    { value: "api_key", label: "API Key" },
                  ]}
                  onChange={(type) => patch({ auth: { ...r.auth, type } })}
                />
                {r.auth?.type === "basic" && (
                  <>
                    <Input
                      disabled={readOnly}
                      placeholder="用户名"
                      value={r.auth.username}
                      onChange={(e) =>
                        patch({ auth: { ...r.auth, username: e.target.value } })
                      }
                    />
                    <Input.Password
                      disabled={readOnly}
                      placeholder="密码或变量"
                      value={r.auth.password}
                      onChange={(e) =>
                        patch({ auth: { ...r.auth, password: e.target.value } })
                      }
                    />
                  </>
                )}
                {r.auth?.type === "bearer" && (
                  <Input.Password
                    disabled={readOnly}
                    placeholder="Token 或 {{token}}"
                    value={r.auth.token}
                    onChange={(e) =>
                      patch({ auth: { ...r.auth, token: e.target.value } })
                    }
                  />
                )}
                {r.auth?.type === "api_key" && (
                  <>
                    <Select
                      disabled={readOnly}
                      value={r.auth.in || "header"}
                      onChange={(place) =>
                        patch({ auth: { ...r.auth, in: place } })
                      }
                      options={[
                        { value: "header", label: "Header" },
                        { value: "query", label: "Query" },
                      ]}
                    />
                    <Input
                      disabled={readOnly}
                      placeholder="Key名称"
                      value={r.auth.key}
                      onChange={(e) =>
                        patch({ auth: { ...r.auth, key: e.target.value } })
                      }
                    />
                    <Input.Password
                      disabled={readOnly}
                      placeholder="Key值"
                      value={r.auth.value}
                      onChange={(e) =>
                        patch({ auth: { ...r.auth, value: e.target.value } })
                      }
                    />
                  </>
                )}
              </Space>
            ),
          },
          {
            key: "settings",
            label: "设置",
            children: (
              <Space>
                请求超时(ms)
                <InputNumber
                  disabled={readOnly}
                  min={1}
                  max={120000}
                  value={r.timeout_ms || 30000}
                  onChange={(timeout_ms) =>
                    patch({ timeout_ms: timeout_ms ?? 30000 })
                  }
                />
              </Space>
            ),
          },
        ]}
      />
    </>
  );
}

import type { EnvironmentDraft } from "../../models/editors";
import { useState } from "react";
import { Button, Checkbox, Input, Space, Typography } from "antd";

export function EnvironmentFields({
  value,
  onChange,
  readOnly = false,
}: {
  value: EnvironmentDraft;
  onChange: (v: EnvironmentDraft) => void;
  readOnly?: boolean;
}) {
  const [rows, setRows] = useState(() =>
    Object.entries(value.variables || {}).map(([key, v]) => ({
      id: crypto.randomUUID(),
      key,
      value: v.value || "",
      secret: !!v.secret,
    })),
  );
  function update(next: typeof rows) {
    setRows(next);
    onChange({
      variables: Object.fromEntries(
        next
          .filter((v) => v.key)
          .map((v) => [v.key, { value: v.value, secret: v.secret }]),
      ),
    });
  }
  return (
    <Space direction="vertical" style={{ width: "100%" }}>
      {rows.map((row, i) => (
        <Space.Compact style={{ width: "100%" }} key={row.id}>
          <Input
            disabled={readOnly}
            placeholder="变量名称"
            value={row.key}
            onChange={(e) =>
              update(
                rows.map((v, j) =>
                  i === j ? { ...v, key: e.target.value } : v,
                ),
              )
            }
          />
          {row.secret ? (
            <Input.Password
              disabled={readOnly}
              placeholder="密钥值"
              value={row.value}
              onChange={(e) =>
                update(
                  rows.map((v, j) =>
                    i === j ? { ...v, value: e.target.value } : v,
                  ),
                )
              }
            />
          ) : (
            <Input
              disabled={readOnly}
              placeholder="变量值"
              value={row.value}
              onChange={(e) =>
                update(
                  rows.map((v, j) =>
                    i === j ? { ...v, value: e.target.value } : v,
                  ),
                )
              }
            />
          )}
          <Checkbox
            disabled={readOnly}
            checked={row.secret}
            onChange={(e) =>
              update(
                rows.map((v, j) =>
                  i === j ? { ...v, secret: e.target.checked } : v,
                ),
              )
            }
          >
            密钥
          </Checkbox>
          <Button
            disabled={readOnly}
            danger
            onClick={() => update(rows.filter((_, j) => i !== j))}
          >
            移除
          </Button>
        </Space.Compact>
      ))}
      {!readOnly && (
        <Button
          type="dashed"
          onClick={() =>
            update([
              ...rows,
              { id: crypto.randomUUID(), key: "", value: "", secret: false },
            ])
          }
        >
          添加变量
        </Button>
      )}
      <Typography.Text type="secondary">
        每次运行复制环境。脚本和提取操作不会修改此处保存的值。
      </Typography.Text>
    </Space>
  );
}

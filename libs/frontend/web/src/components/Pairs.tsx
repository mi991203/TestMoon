import { useState } from "react";
import { Button, Input, Space } from "antd";

export function Pairs({
  value = {},
  onChange,
  readOnly = false,
  secret = false,
}: {
  value?: Record<string, string>;
  onChange?: (v: Record<string, string>) => void;
  readOnly?: boolean;
  secret?: boolean;
}) {
  const [rows, setRows] = useState(() =>
    Object.entries(value).map(([key, value]) => ({
      id: crypto.randomUUID(),
      key,
      value,
    })),
  );
  function update(next: typeof rows) {
    setRows(next);
    onChange?.(
      Object.fromEntries(
        next.filter((x) => x.key).map((x) => [x.key, x.value]),
      ),
    );
  }
  return (
    <Space direction="vertical" style={{ width: "100%" }}>
      {rows.map((row, i) => (
        <Space.Compact key={row.id} style={{ width: "100%" }}>
          <Input
            aria-label="字段名称"
            placeholder="名称"
            value={row.key}
            disabled={readOnly}
            onChange={(e) =>
              update(
                rows.map((x, j) =>
                  j === i ? { ...x, key: e.target.value } : x,
                ),
              )
            }
          />
          {secret ? (
            <Input.Password
              placeholder="值"
              value={row.value}
              disabled={readOnly}
              onChange={(e) =>
                update(
                  rows.map((x, j) =>
                    j === i ? { ...x, value: e.target.value } : x,
                  ),
                )
              }
            />
          ) : (
            <Input
              placeholder="值，支持 {{变量}}"
              value={row.value}
              disabled={readOnly}
              onChange={(e) =>
                update(
                  rows.map((x, j) =>
                    j === i ? { ...x, value: e.target.value } : x,
                  ),
                )
              }
            />
          )}
          <Button
            disabled={readOnly}
            onClick={() => update(rows.filter((_, j) => j !== i))}
          >
            移除
          </Button>
        </Space.Compact>
      ))}
      {!readOnly && (
        <Button
          type="dashed"
          onClick={() =>
            update([...rows, { id: crypto.randomUUID(), key: "", value: "" }])
          }
        >
          添加字段
        </Button>
      )}
    </Space>
  );
}

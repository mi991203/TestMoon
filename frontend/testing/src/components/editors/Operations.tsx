import type { ContractsOperation } from "@testmoon/api";

import { Button, Space, Typography } from "antd";
import { Code } from "@testmoon/web";

export function Operations({
  value = [],
  onChange,
  readOnly = false,
}: {
  value?: ContractsOperation[];
  onChange: (v: ContractsOperation[]) => void;
  readOnly?: boolean;
}) {
  return (
    <Space direction="vertical" style={{ width: "100%" }}>
      {value.map((op, i) => (
        <div key={i}>
          <Space style={{ marginBottom: 8 }}>
            <Typography.Text>
              JavaScript v1 {op.source ? `· ${op.source}` : ""}
            </Typography.Text>
            {!readOnly && (
              <>
                <Button
                  size="small"
                  disabled={!i}
                  onClick={() => {
                    const next = [...value];
                    [next[i - 1], next[i]] = [next[i], next[i - 1]];
                    onChange(next);
                  }}
                >
                  上移
                </Button>
                <Button
                  size="small"
                  danger
                  onClick={() => onChange(value.filter((_, j) => j !== i))}
                >
                  删除
                </Button>
              </>
            )}
          </Space>
          <Code
            language="javascript"
            value={op.script || ""}
            onChange={(script) =>
              onChange(value.map((v, j) => (j === i ? { ...v, script } : v)))
            }
            height={180}
            readOnly={readOnly}
          />
        </div>
      ))}
      {!readOnly && (
        <Button
          type="dashed"
          onClick={() =>
            onChange([
              ...value,
              {
                type: "javascript",
                version: 1,
                source: "",
                script:
                  '// pm.environment.get("base_url")\n// pm.test("状态码", () => pm.expect(pm.response.code).to.equal(200));',
              },
            ])
          }
        >
          添加 JS 操作
        </Button>
      )}
    </Space>
  );
}

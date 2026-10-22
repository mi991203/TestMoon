import type { ContractsExtraction } from "@testmoon/api";

import { Button, Checkbox, Input, Select, Space } from "antd";

import { sources } from "./defaults";
export function Extractions({
  value = [],
  onChange,
  readOnly = false,
}: {
  value?: ContractsExtraction[];
  onChange: (value: ContractsExtraction[]) => void;
  readOnly?: boolean;
}) {
  return (
    <Space direction="vertical" style={{ width: "100%" }}>
      {value.map((v, i) => (
        <Space.Compact style={{ width: "100%" }} key={i}>
          <Select
            disabled={readOnly}
            style={{ width: 180 }}
            options={sources}
            value={v.source}
            onChange={(source: string) =>
              onChange(value.map((x, j) => (i === j ? { ...x, source } : x)))
            }
          />
          <Input
            disabled={readOnly}
            value={v.path}
            placeholder="/data/token"
            onChange={(e) =>
              onChange(
                value.map((x, j) =>
                  i === j ? { ...x, path: e.target.value } : x,
                ),
              )
            }
          />
          <Input
            disabled={readOnly}
            value={v.variable}
            placeholder="变量名"
            onChange={(e) =>
              onChange(
                value.map((x, j) =>
                  i === j ? { ...x, variable: e.target.value } : x,
                ),
              )
            }
          />
          <Checkbox
            disabled={readOnly}
            checked={v.secret}
            onChange={(e) =>
              onChange(
                value.map((x, j) =>
                  i === j ? { ...x, secret: e.target.checked } : x,
                ),
              )
            }
          >
            密钥
          </Checkbox>
          <Button
            disabled={readOnly}
            onClick={() => onChange(value.filter((_, j) => j !== i))}
          >
            移除
          </Button>
        </Space.Compact>
      ))}
      {!readOnly && (
        <Button
          type="dashed"
          onClick={() =>
            onChange([
              ...value,
              { source: "json", path: "", variable: "", secret: false },
            ])
          }
        >
          添加提取规则
        </Button>
      )}
    </Space>
  );
}

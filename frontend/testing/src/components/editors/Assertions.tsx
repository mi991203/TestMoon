import type { ContractsAssertion } from "@testmoon/api";
import type { JsonValue } from "@testmoon/api";

import { Button, Input, Select, Space, Typography } from "antd";

import { sources } from "./defaults";
export function Assertions({
  value = [],
  onChange,
  readOnly = false,
}: {
  value?: ContractsAssertion[];
  onChange: (v: ContractsAssertion[]) => void;
  readOnly?: boolean;
}) {
  return (
    <Space direction="vertical" style={{ width: "100%" }}>
      {value.map((a, i) => (
        <Space.Compact key={i} style={{ width: "100%" }}>
          <Select
            disabled={readOnly}
            style={{ width: 190 }}
            value={a.source}
            options={sources}
            onChange={(source) =>
              onChange(value.map((x, j) => (j === i ? { ...x, source } : x)))
            }
          />
          <Input
            disabled={
              readOnly || a.source === "status" || a.source === "duration"
            }
            placeholder={a.source === "json" ? "/data/token" : "字段 / 路径"}
            value={a.path}
            onChange={(e) =>
              onChange(
                value.map((x, j) =>
                  j === i ? { ...x, path: e.target.value } : x,
                ),
              )
            }
          />
          <Select
            disabled={readOnly}
            style={{ width: 120 }}
            value={a.operator}
            options={[
              "eq",
              "neq",
              "contains",
              "exists",
              "regex",
              "lt",
              "lte",
              "gt",
              "gte",
            ].map((value) => ({ value }))}
            onChange={(operator) =>
              onChange(value.map((x, j) => (j === i ? { ...x, operator } : x)))
            }
          />
          <Input
            disabled={readOnly || a.operator === "exists"}
            placeholder="预期：数字、文本或JSON"
            value={
              typeof a.expected === "string"
                ? a.expected
                : JSON.stringify(a.expected)
            }
            onChange={(e) => {
              let expected: JsonValue = e.target.value;
              try {
                expected = JSON.parse(expected);
              } catch {}
              onChange(value.map((x, j) => (j === i ? { ...x, expected } : x)));
            }}
          />
          <Button
            disabled={readOnly}
            danger
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
              { source: "status", path: "", operator: "eq", expected: 200 },
            ])
          }
        >
          添加断言
        </Button>
      )}
      <Typography.Text type="secondary">
        JSON 字段路径采用 JSON Pointer，例如
        /data/items/0/id；无断言的用例标记为未校验。
      </Typography.Text>
    </Space>
  );
}

import type { CaseDraft } from "../../models/editors";
import type { Asset } from "../../models/resources";
import { useState } from "react";
import { Checkbox, Collapse, Select } from "antd";
import { Pairs } from "@testmoon/web";
import { blankRequest, requestPatch, hasRequestPatch } from "./defaults";
import { RequestFields } from "./RequestFields";
import { Operations } from "./Operations";
import { Assertions } from "./Assertions";
import { Extractions } from "./Extractions";
export function CaseFields({
  value,
  onChange,
  apis = [],
  readOnly = false,
  inline = false,
  hideRequest = false,
}: {
  value: CaseDraft;
  onChange: (v: CaseDraft) => void;
  apis?: Asset<"api">[];
  readOnly?: boolean;
  inline?: boolean;
  hideRequest?: boolean;
}) {
  const [override, setOverride] = useState(hasRequestPatch(value.override));
  const base =
    apis.find((v) => v.id === value.api_id)?.document?.request ||
    blankRequest();
  const patch = (data: CaseDraft) => onChange({ ...value, ...data });
  return (
    <>
      {!inline && (
        <>
          <div className="form-section">关联接口</div>
          <Select
            disabled={readOnly}
            style={{ width: "100%" }}
            allowClear
            showSearch
            optionFilterProp="label"
            value={value.api_id || undefined}
            options={apis.map((v) => ({ value: v.id, label: v.name }))}
            onChange={(api_id) => patch({ api_id: api_id || "" })}
          />
        </>
      )}
      {!hideRequest &&
        (value.api_id ? (
          <>
            <div style={{ margin: "16px 0" }}>
              <Checkbox
                disabled={readOnly}
                checked={override}
                onChange={(e) => {
                  setOverride(e.target.checked);
                  if (!e.target.checked) patch({ override: {} });
                }}
              >
                覆盖本用例的请求输入
              </Checkbox>
            </div>
            {override && (
              <RequestFields
                value={{ ...base, ...value.override }}
                readOnly={readOnly}
                onChange={(request) =>
                  patch({ override: requestPatch(request) })
                }
              />
            )}
          </>
        ) : (
          <RequestFields
            value={value.request || blankRequest()}
            onChange={(request) => patch({ request })}
            readOnly={readOnly}
          />
        ))}
      <Collapse
        style={{ marginTop: 20 }}
        defaultActiveKey={["assertions"]}
        items={[
          {
            key: "assertions",
            label: "响应断言",
            children: (
              <Assertions
                value={value.assertions}
                onChange={(assertions) => patch({ assertions })}
                readOnly={readOnly}
              />
            ),
          },
          {
            key: "extractions",
            label: "变量提取",
            children: (
              <Extractions
                value={value.extractions}
                onChange={(extractions) => patch({ extractions })}
                readOnly={readOnly}
              />
            ),
          },
          {
            key: "variables",
            label: "用例变量",
            children: (
              <Pairs
                value={value.variables}
                onChange={(variables) => patch({ variables })}
                readOnly={readOnly}
              />
            ),
          },
          {
            key: "before",
            label: "前置操作",
            children: (
              <Operations
                value={value.before}
                onChange={(before) => patch({ before })}
                readOnly={readOnly}
              />
            ),
          },
          {
            key: "after",
            label: "后置操作",
            children: (
              <Operations
                value={value.after}
                onChange={(after) => patch({ after })}
                readOnly={readOnly}
              />
            ),
          },
        ]}
      />
    </>
  );
}

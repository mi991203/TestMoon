import type { ScenarioDraft } from "../../models/editors";
import type { Asset } from "../../models/resources";
import type { StepDraft } from "../../models/editors";
import { RequestFields } from "./RequestFields";
import { useState } from "react";
import {
  Button,
  Checkbox,
  Input,
  Modal,
  Select,
  Space,
  Typography,
} from "antd";
import { Pairs } from "@testmoon/web";
import { blankRequest, requestPatch, hasRequestPatch } from "./defaults";
import { CaseFields } from "./CaseFields";
import { TagIndex } from "./TagIndex";
export function ScenarioFields({
  value,
  onChange,
  cases,
  apis,
  readOnly = false,
}: {
  value: ScenarioDraft;
  onChange: (v: ScenarioDraft) => void;
  cases: Asset<"case">[];
  apis: Asset<"api">[];
  readOnly?: boolean;
}) {
  const steps: StepDraft[] = value.steps || [];
  const [index, setIndex] = useState<number | null>(null);
  const [step, setStep] = useState<StepDraft>({});
  const [over, setOver] = useState(false);
  function edit(i: number) {
    const v = structuredClone(steps[i]);
    setStep(v);
    setOver(hasRequestPatch(v.override));
    setIndex(i);
  }
  return (
    <>
      <Typography.Paragraph type="secondary">
        步骤按顺序执行。默认失败停止当前场景，可对单步配置失败后继续。
      </Typography.Paragraph>
      {steps.map((st, i) => (
        <div className="step-card" key={st.id || i}>
          <Space>
            <TagIndex index={i} />
            <span>
              {st.name ||
                cases.find((v) => v.id === st.case_id)?.name ||
                "独立请求"}
            </span>
            {st.continue_on_failure && (
              <Typography.Text type="secondary">失败后继续</Typography.Text>
            )}
          </Space>
          <Space>
            <Button onClick={() => edit(i)}>
              {readOnly ? "查看" : "编辑"}
            </Button>
            {!readOnly && (
              <>
                <Button
                  disabled={!i}
                  onClick={() => {
                    const next = [...steps];
                    [next[i - 1], next[i]] = [next[i], next[i - 1]];
                    onChange({ ...value, steps: next });
                  }}
                >
                  上移
                </Button>
                <Button
                  danger
                  onClick={() =>
                    onChange({
                      ...value,
                      steps: steps.filter((_, j) => j !== i),
                    })
                  }
                >
                  删除
                </Button>
              </>
            )}
          </Space>
        </div>
      ))}
      {!readOnly && (
        <Button
          type="dashed"
          block
          onClick={() => {
            setStep({
              id: crypto.randomUUID(),
              name: "",
              case_id: "",
              inline: { request: blankRequest() },
            });
            setOver(false);
            setIndex(steps.length);
          }}
        >
          添加步骤
        </Button>
      )}
      <Modal
        width={1000}
        title="场景步骤"
        open={index !== null}
        onCancel={() => setIndex(null)}
        onOk={() => {
          if (index === null) return;
          const next = [...steps];
          next[index] = step;
          onChange({ ...value, steps: next });
          setIndex(null);
        }}
        okButtonProps={{ disabled: readOnly }}
        destroyOnHidden
      >
        <Space direction="vertical" style={{ width: "100%" }}>
          <Input
            disabled={readOnly}
            placeholder="步骤名称"
            value={step.name}
            onChange={(e) => setStep({ ...step, name: e.target.value })}
          />
          <Select
            disabled={readOnly}
            style={{ width: 250 }}
            value={step.case_id ? "reference" : "inline"}
            options={[
              { value: "reference", label: "引用已有接口用例" },
              { value: "inline", label: "场景内独立请求" },
            ]}
            onChange={(type) =>
              setStep(
                type === "inline"
                  ? {
                      ...step,
                      case_id: "",
                      inline: { request: blankRequest() },
                    }
                  : { ...step, case_id: cases[0]?.id || "", inline: undefined },
              )
            }
          />
          {step.case_id ? (
            <>
              <Select
                disabled={readOnly}
                style={{ width: "100%" }}
                value={step.case_id}
                options={cases.map((v) => ({ value: v.id, label: v.name }))}
                onChange={(case_id) => setStep({ ...step, case_id })}
              />
              <Checkbox
                disabled={readOnly}
                checked={over}
                onChange={(e) => {
                  setOver(e.target.checked);
                  if (!e.target.checked) setStep({ ...step, override: {} });
                }}
              >
                覆盖步骤请求（填写需要覆盖的完整输入）
              </Checkbox>
              {over && (
                <RequestFields
                  value={step.override || blankRequest()}
                  readOnly={readOnly}
                  onChange={(override) =>
                    setStep({ ...step, override: requestPatch(override) })
                  }
                />
              )}
            </>
          ) : (
            <CaseFields
              key={index}
              value={step.inline || { request: blankRequest() }}
              inline
              apis={apis}
              readOnly={readOnly}
              onChange={(inline) => setStep({ ...step, inline })}
            />
          )}
          <div className="form-section">步骤变量</div>
          <Pairs
            key={`vars-${index}`}
            value={step.variables}
            readOnly={readOnly}
            onChange={(variables) => setStep({ ...step, variables })}
          />
          <Checkbox
            disabled={readOnly}
            checked={step.continue_on_failure}
            onChange={(e) =>
              setStep({ ...step, continue_on_failure: e.target.checked })
            }
          >
            失败后继续后续步骤
          </Checkbox>
        </Space>
      </Modal>
    </>
  );
}

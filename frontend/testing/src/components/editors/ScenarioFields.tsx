import type { ScenarioDraft } from "../../models/editors";
import type { Asset } from "../../models/resources";
import type { StepDraft } from "../../models/editors";
import { RequestFields } from "./RequestFields";
import { useMemo, useState } from "react";
import {
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
import { OrchestrationTree } from "../orchestration/OrchestrationTree";
import { insertNode, updateNode } from "../orchestration/tree";
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
  const steps = useMemo(() => {
    const normalize = (nodes: StepDraft[]): StepDraft[] => nodes.map((node) => ({
      ...node, id: node.id || crypto.randomUUID(),
      ...(node.kind === "folder" ? { children: normalize(node.children || []) } : {}),
    }));
    return normalize(value.steps || []);
  }, [value.steps]);
  const [editing, setEditing] = useState<{ key?: string; parent: string }>();
  const [stepType, setStepType] = useState("inline");
  const [step, setStep] = useState<StepDraft>({});
  const [over, setOver] = useState(false);
  function edit(node: StepDraft) {
    const v = structuredClone(node);
    setStep(v);
    setStepType(v.case_id ? "reference" : "inline");
    setOver(hasRequestPatch(v.override));
    setEditing({ key: v.id, parent: "" });
  }
  return (
    <>
      <Typography.Paragraph type="secondary">
        默认失败停止当前场景，可对单步配置失败后继续。
      </Typography.Paragraph>
      <OrchestrationTree
        value={steps}
        onChange={(steps) => onChange({ ...value, steps })}
        readOnly={readOnly}
        title={(st) => (st.name || cases.find((v) => v.id === st.case_id)?.name || "独立请求") + (st.continue_on_failure ? " · 失败后继续" : "")}
        createFolder={(name) => ({ id: crypto.randomUUID(), kind: "folder", name, children: [] })}
        onEdit={edit}
        addLabel="添加步骤"
        onAdd={(parent) => {
          setStep({ id: crypto.randomUUID(), kind: "step", name: "", inline: { request: blankRequest() } });
          setStepType("inline");
          setOver(false);
          setEditing({ parent });
        }}
      />
      <Modal
        width={1000}
        title="场景步骤"
        open={!!editing}
        onCancel={() => setEditing(undefined)}
        onOk={() => {
          if (!editing) return;
          const next = editing.key
            ? updateNode(steps, editing.key, () => step)
            : insertNode(steps, editing.parent, step);
          onChange({ ...value, steps: next });
          setEditing(undefined);
        }}
        okButtonProps={{ disabled: readOnly || (stepType === "reference" && !step.case_id) }}
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
            value={stepType}
            options={[
              { value: "reference", label: "引用已有接口用例" },
              { value: "inline", label: "场景内独立请求" },
            ]}
            onChange={(type) => {
              setStepType(type);
              setStep(
                type === "inline"
                  ? {
                      ...step,
                      case_id: "",
                      inline: { request: blankRequest() },
                    }
                  : { ...step, case_id: cases[0]?.id || "", inline: undefined },
              );
            }}
          />
          {stepType === "reference" ? (
            <>
              <Select
                disabled={readOnly}
                style={{ width: "100%" }}
                showSearch
                optionFilterProp="label"
                placeholder="选择接口用例"
                value={step.case_id || undefined}
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
              key={step.id}
              value={step.inline || { request: blankRequest() }}
              inline
              apis={apis}
              readOnly={readOnly}
              onChange={(inline) => setStep({ ...step, inline })}
            />
          )}
          <div className="form-section">步骤变量</div>
          <Pairs
            key={`vars-${step.id}`}
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

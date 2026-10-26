import type { Asset } from "../../models/resources";
import type {
  ResourceDraft,
  EditableDefinition,
  CaseDraft,
  ScenarioDraft,
  EnvironmentDraft,
} from "../../models/editors";
import { resources } from "../../services/resources";

import { useEffect, useRef, useState } from "react";
import { execution } from "@testmoon/api";

import { Button, Card, Form, Input, Select, Space, Typography } from "antd";
import { ErrorState, PageHeading, Scope, useAction } from "@testmoon/web";
import {
  CaseFields,
  EnvironmentFields,
  RequestFields,
  ScenarioFields,
} from "../editors";

import { useScopedNavigate } from "../../hooks/navigation";
import { titles, folderOptions } from "./defaults";
import { RunDialog } from "./RunDialog";
import { ExecutionFields, useExecutionOptions } from "./ExecutionFields";
import { RunResult } from "./RunResult";
import { useRun } from "../../hooks/queries";
import { active } from "../../features/reports/state";
export function ResourceEditor({
  row,
  kind,
  scope,
  role,
  apis,
  cases,
  folders,
  onSaved,
  back,
  debug,
  onDebugChange,
}: {
  row: ResourceDraft;
  kind: string;
  scope: Scope;
  role: string;
  apis: Asset<"api">[];
  cases: Asset<"case">[];
  folders: Asset<"folder">[];
  onSaved: (row: Asset) => void;
  back: () => void;
  debug?: CaseDraft;
  onDebugChange?: (v: CaseDraft) => void;
}) {
  const [doc, setDoc] = useState<EditableDefinition>(() =>
    structuredClone(row.document),
  );
  const [form] = Form.useForm();
  const { busy, run } = useAction();
  const submit = useAction();
  const [executing, setExecuting] = useState(false);
  const [runId, setRunId] = useState("");
  const [executionError, setExecutionError] = useState<unknown>();
  const resultRef = useRef<HTMLDivElement>(null);
  const result = useRun(scope, runId);
  const options = useExecutionOptions(scope, doc);
  const running =
    !!runId && (result.isPending || active(result.data?.run.state || ""));
  const go = useScopedNavigate();
  const readOnly = role !== "admin" && role !== "editor";
  const inlineExecution = kind === "api" || kind === "case";
  useEffect(() => {
    if (runId)
      resultRef.current?.scrollIntoView({ behavior: "smooth", block: "start" });
  }, [runId]);
  async function execute() {
    if (busy || submit.busy || running) return;
    setExecutionError(undefined);
    await submit.run(async () => {
      try {
        const current = await options.runners.refetch();
        if (current.error) throw current.error;
        const runner_id = doc.runner_id || "public";
        if (!current.data?.items.some((v) => v.id === runner_id && v.online))
          throw new Error("所选 Runner 离线、已撤销或不可用，请重新选择。");
        const config = {
          ...scope,
          environment_id: doc.environment_id || "",
          runner_id,
          timeout_seconds: 900,
          idempotency_key: crypto.randomUUID(),
        };
        const name =
          form.getFieldValue("name") ||
          row.name ||
          (kind === "api" ? "接口调试" : "用例草稿");
        const out =
          kind === "api"
            ? await execution.debugHTTP({
                ...config,
                debug: { ...debug, name, request: doc.request },
              })
            : await execution.runCase({
                ...config,
                case_id: row.id,
                draft: { ...(doc as CaseDraft), name },
              });
        setRunId(out.run.id);
      } catch (error) {
        setExecutionError(error);
        throw error;
      }
    }, "执行已提交");
  }
  const heading = (
    <PageHeading
      title={row.id ? row.name : `新建${titles[kind]}`}
      description={
        readOnly
          ? "只读查看"
          : inlineExecution
            ? "执行当前表单草稿；保存后可复用执行配置"
            : "保存定义后可继续配置和执行"
      }
      extra={
        <Space>
          <Button onClick={back}>返回列表</Button>
          {!readOnly && (
            <>
              <Button
                onClick={() => form.submit()}
                loading={busy}
                disabled={submit.busy || running}
                type="primary"
              >
                保存
              </Button>
              {(inlineExecution || (row.id && kind === "scenario")) && (
                <Button
                  loading={submit.busy || running}
                  disabled={
                    busy ||
                    (inlineExecution &&
                      (options.loading ||
                        !!options.unavailable ||
                        !!options.envs.error ||
                        !!options.runners.error))
                  }
                  onClick={inlineExecution ? execute : () => setExecuting(true)}
                >
                  {kind === "api"
                    ? "发送请求"
                    : kind === "case"
                      ? "执行当前草稿"
                      : "执行已保存版本"}
                </Button>
              )}
            </>
          )}
        </Space>
      }
    />
  );
  return (
    <>
      {heading}
      <Card>
        <Form
          layout="vertical"
          form={form}
          initialValues={{ name: row.name, folder_id: row.folder_id || "" }}
          onFinish={(value) => {
            if (submit.busy || running) return;
            return run(async () => {
              const result = await resources.save(
                scope,
                kind,
                { ...value, kind, revision: row.revision || 0, document: doc },
                row.id,
              );
              onSaved(result.resource);
            });
          }}
        >
          <Form.Item
            name="name"
            label="名称"
            rules={[{ required: true, max: 128 }]}
          >
            <Input disabled={readOnly} />
          </Form.Item>
          {kind !== "environment" && (
            <Form.Item name="folder_id" label="目录">
              <Select
                disabled={readOnly}
                options={folderOptions(folders, kind)}
              />
            </Form.Item>
          )}
        </Form>
        {inlineExecution && (
          <ExecutionFields
            value={{
              environment_id: doc.environment_id,
              runner_id: doc.runner_id,
            }}
            onChange={(value) => setDoc({ ...doc, ...value })}
            readOnly={readOnly || submit.busy || running}
            options={options}
          />
        )}
        {kind === "api" ? (
          <>
            <RequestFields
              value={doc.request || {}}
              readOnly={readOnly}
              onChange={(request) =>
                setDoc({ ...doc, protocol: "http", request })
              }
            />
            <div className="form-section">接口说明</div>
            <Input.TextArea
              value={doc.description}
              disabled={readOnly}
              onChange={(e) => setDoc({ ...doc, description: e.target.value })}
            />
            {debug && onDebugChange && (
              <>
                <div className="form-section">本次调试配置</div>
                <Typography.Paragraph type="secondary">
                  调试断言、变量和脚本仅用于本次发送；长期配置请保存为接口用例。
                </Typography.Paragraph>
                <CaseFields
                  value={debug}
                  onChange={onDebugChange}
                  inline
                  hideRequest
                  readOnly={readOnly}
                />
              </>
            )}
          </>
        ) : kind === "case" ? (
          <CaseFields
            value={doc as CaseDraft}
            onChange={setDoc}
            apis={apis}
            readOnly={readOnly}
          />
        ) : kind === "scenario" ? (
          <ScenarioFields
            value={doc as ScenarioDraft}
            onChange={setDoc}
            apis={apis}
            cases={cases}
            readOnly={readOnly}
          />
        ) : (
          <EnvironmentFields
            value={doc as EnvironmentDraft}
            onChange={setDoc}
            readOnly={readOnly}
          />
        )}
      </Card>
      <ErrorState error={executionError} />
      {runId && (
        <div ref={resultRef} className="inline-run-result">
          <RunResult
            scope={scope}
            id={runId}
            onReport={() => go(`/reports/${runId}`)}
          />
        </div>
      )}
      {executing && (
        <RunDialog
          scope={scope}
          selection={kind !== "api" ? { kind, id: row.id } : undefined}
          debug={
            kind === "api" ? { ...debug, request: doc.request } : undefined
          }
          onClose={() => setExecuting(false)}
          onRun={(id) => go(`/reports/${id}`)}
        />
      )}
    </>
  );
}

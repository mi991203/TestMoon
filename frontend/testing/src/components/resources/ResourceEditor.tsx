import type { Asset } from "../../models/resources";
import type {
  ResourceDraft,
  EditableDefinition,
  CaseDraft,
  ScenarioDraft,
  EnvironmentDraft,
} from "../../models/editors";
import { resources } from "../../services/resources";

import { useState } from "react";

import { Button, Card, Form, Input, Select, Space, Typography } from "antd";
import { PageHeading, Scope, useAction } from "@testmoon/web";
import {
  CaseFields,
  EnvironmentFields,
  RequestFields,
  ScenarioFields,
} from "../editors";

import { useScopedNavigate } from "../../hooks/navigation";
import { titles, folderOptions } from "./defaults";
import { RunDialog } from "./RunDialog";
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
  const [executing, setExecuting] = useState(false);
  const go = useScopedNavigate();
  const readOnly = role !== "admin" && role !== "editor";
  const heading = (
    <PageHeading
      title={row.id ? row.name : `新建${titles[kind]}`}
      description={readOnly ? "只读查看" : "保存定义后可继续配置和执行"}
      extra={
        <Space>
          <Button onClick={back}>返回列表</Button>
          {!readOnly && (
            <>
              <Button
                onClick={() => form.submit()}
                loading={busy}
                type="primary"
              >
                保存
              </Button>
              {(kind === "api" ||
                (row.id && (kind === "case" || kind === "scenario"))) && (
                <Button onClick={() => setExecuting(true)}>
                  {kind === "api" ? "发送请求" : "执行已保存版本"}
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
          onFinish={(value) =>
            run(async () => {
              const result = await resources.save(
                scope,
                kind,
                { ...value, kind, revision: row.revision || 0, document: doc },
                row.id,
              );
              onSaved(result.resource);
            })
          }
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

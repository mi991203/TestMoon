type ImportEntry = Awaited<ReturnType<typeof previewImport>>["entries"][number];

import type { Int64 } from "@testmoon/api";

import { previewImport } from "../../services/workspace";
import { core } from "@testmoon/api";
import React, { useState } from "react";

import {
  Alert,
  Button,
  Modal,
  Select,
  Space,
  Table,
  Typography,
  Upload,
} from "antd";
import { Code, Scope, useAction } from "@testmoon/web";
import { useAssets } from "../../hooks/queries";
import { folderOptions } from "../../components/resources/defaults";

export function Imports({
  scope,
  role,
  initialFolderId,
  onClose,
  onImported,
}: {
  scope: Scope;
  role: string;
  initialFolderId: string;
  onClose: () => void;
  onImported: (folder: string) => void;
}) {
  const [targetFolder, setTargetFolder] = useState(initialFolderId);
  const [format, setFormat] = useState("postman");
  const [content, setContent] = useState("");
  const [preview, setPreview] = useState<Awaited<
    ReturnType<typeof previewImport>
  > | null>(null);
  const [selected, setSelected] = useState<React.Key[]>([]);
  const [updates, setUpdates] = useState<
    Record<string, { id: string; revision: Int64 }>
  >({});
  const apis = useAssets(scope, "api"),
    cases = useAssets(scope, "case"),
    scenarios = useAssets(scope, "scenario"),
    folders = useAssets(scope, "folder");
  const existing = [
    ...(apis.data?.items || []),
    ...(cases.data?.items || []),
    ...(scenarios.data?.items || []),
    ...(folders.data?.items || []),
  ];
  const { busy, run } = useAction();
  const editable = role === "admin" || role === "editor";
  return (
    <Modal
      title="接口与用例导入"
      open
      width={1100}
      footer={null}
      onCancel={onClose}
      closable={!busy}
      keyboard={!busy}
      maskClosable={!busy}
    >
      <Space wrap style={{ marginBottom: 16 }}>
        <span>导入到</span>
        <Select
          aria-label="导入目标目录"
          className="import-target"
          value={targetFolder}
          disabled={!editable || busy}
          showSearch
          optionFilterProp="label"
          options={folderOptions(folders.data?.items || [], "api")}
          onChange={setTargetFolder}
        />
        <Select
          value={format}
          disabled={!editable || busy}
          style={{ width: 250 }}
          options={[
            { value: "postman", label: "Postman Collection v2.1" },
            { value: "openapi", label: "OpenAPI 2.0 / 3.0 / 3.1" },
          ]}
          onChange={(value) => {
            setFormat(value);
            setPreview(null);
          }}
        />
        <Upload
          accept=".json,.yaml,.yml"
          disabled={!editable || busy}
          showUploadList={false}
          beforeUpload={async (file) => {
            setContent(await file.text());
            setPreview(null);
            return false;
          }}
        >
          <Button disabled={!editable || busy}>选择文件</Button>
        </Upload>
        <Button
          type="primary"
          disabled={!editable || !content || busy}
          loading={busy}
          onClick={() =>
            run(async () => {
              const out = await previewImport({
                ...scope,
                ...{
                  format,
                  content,
                },
              });
              setPreview(out);
              setSelected(out.entries.map((v) => v.temp_id));
              setUpdates({});
            }, "预览已生成")
          }
        >
          生成预览
        </Button>
      </Space>
      <Code
        language={format === "openapi" ? "yaml" : "json"}
        value={content}
        onChange={(value) => {
          setContent(value);
          setPreview(null);
        }}
        height={280}
        readOnly={!editable || busy}
      />
      {preview && (
        <>
          <div style={{ margin: "20px 0" }}>
            {preview.diagnostics?.length > 0 ? (
              <Alert
                type="warning"
                showIcon
                message="兼容诊断"
                description={
                  <ul>
                    {preview.diagnostics.map((v: string, i: number) => (
                      <li key={i}>{v}</li>
                    ))}
                  </ul>
                }
              />
            ) : (
              <Alert
                type="success"
                message="解析成功，请确认资源选择和更新目标。"
              />
            )}
          </div>
          <Table<ImportEntry>
            rowKey="temp_id"
            dataSource={preview.entries}
            rowSelection={{
              selectedRowKeys: selected,
              onChange: setSelected,
              getCheckboxProps: () => ({ disabled: busy }),
            }}
            pagination={{ pageSize: 15 }}
            scroll={{ x: 650 }}
            columns={[
              { title: "类型", dataIndex: "kind" },
              { title: "名称", dataIndex: "name" },
              {
                title: "写入方式",
                render: (_, v) => (
                  <Select
                    style={{ width: 280 }}
                    disabled={busy}
                    allowClear
                    placeholder="新增资源"
                    value={updates[v.temp_id]?.id}
                    options={existing
                      .filter(
                        (x) =>
                          x.kind === v.kind &&
                          (v.kind !== "folder" ||
                            ("kind" in x.document &&
                              "kind" in (v.document || {}) &&
                              x.document.kind ===
                                (v.document && "kind" in v.document
                                  ? v.document.kind
                                  : undefined))),
                      )
                      .map((x) => ({ value: x.id, label: x.name }))}
                    onChange={(id) => {
                      const next = { ...updates };
                      if (!id) delete next[v.temp_id];
                      else
                        next[v.temp_id] = {
                          id,
                          revision:
                            existing.find((x) => x.id === id)?.revision || 0,
                        };
                      setUpdates(next);
                    }}
                  />
                ),
              },
            ]}
          />
          <Button
            type="primary"
            disabled={!selected.length || !editable}
            loading={busy}
            onClick={() =>
              run(async () => {
                await core.commitImport({
                  ...scope,
                  ...{ selected: selected.map(String), updates },
                  id: preview.batch_id,
                  target_folder_id: targetFolder,
                });
                setPreview(null);
                setContent("");
                onImported(targetFolder);
              }, "导入完成")
            }
          >
            确认提交 {selected.length} 项
          </Button>
          <Typography.Paragraph type="secondary" style={{ marginTop: 12 }}>
            新增接口保留源目录层级并导入所选目录；用例、场景保留各自目录和接口关联。更新已有资源时保留原位置。
            引用的接口、用例和父目录必须同时选中；更新目标若已变化，整个提交会拒绝。
          </Typography.Paragraph>
        </>
      )}
    </Modal>
  );
}

import type { Asset } from "../../models/resources";
import type { FolderDraft } from "../../models/editors";
import { resources } from "../../services/resources";

import { useState } from "react";

import { Input, Modal, Select, Space } from "antd";
import { Scope, useAction } from "@testmoon/web";

import { folderOptions } from "./defaults";
export function FolderDialog({
  row,
  kind,
  scope,
  folders,
  onClose,
}: {
  row: FolderDraft;
  kind: string;
  scope: Scope;
  folders: Asset<"folder">[];
  onClose: () => void;
}) {
  const [value, setValue] = useState<FolderDraft>(() => ({ ...row }));
  const { busy, run } = useAction();
  const descendants = new Set<string>(row.id ? [row.id] : []);
  let previous = 0;
  while (previous !== descendants.size) {
    previous = descendants.size;
    for (const f of folders)
      if (descendants.has(f.folder_id)) descendants.add(f.id);
  }
  return (
    <Modal
      title="资源目录"
      open
      onCancel={onClose}
      confirmLoading={busy}
      onOk={() =>
        run(async () => {
          await resources.save(
            scope,
            "folder",
            {
              kind: "folder",
              name: value.name,
              folder_id: value.folder_id || "",
              revision: row.revision || 0,
              document: { kind },
            },
            row.id,
          );
          onClose();
        })
      }
    >
      <Space direction="vertical" style={{ width: "100%" }}>
        <Input
          aria-label="目录名称"
          placeholder="目录名称"
          value={value.name}
          onChange={(e) => setValue({ ...value, name: e.target.value })}
        />
        <Select
          aria-label="父目录"
          style={{ width: "100%" }}
          value={value.folder_id || ""}
          options={folderOptions(folders, kind).filter(
            (v) => !descendants.has(v.value),
          )}
          onChange={(folder_id) => setValue({ ...value, folder_id })}
        />
      </Space>
    </Modal>
  );
}

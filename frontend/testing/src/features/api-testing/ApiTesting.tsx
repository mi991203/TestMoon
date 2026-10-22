import type { Asset } from "../../models/resources";
import type { FolderDraft } from "../../models/editors";

import type { DataNode } from "antd/es/tree";
import { resources } from "../../services/resources";
import React, { useEffect, useState } from "react";
import { matchPath, Route, Routes, useLocation } from "react-router-dom";
import {
  Button,
  Empty,
  Input,
  Popconfirm,
  Space,
  Tag,
  Tree,
  Typography,
} from "antd";
import { ErrorState, useAction } from "@testmoon/web";
import { useAssets, useResource } from "../../hooks/queries";
import { useScopedNavigate } from "../../hooks/navigation";
import { FolderDialog } from "../../components/resources";

import { Context, editable } from "./context";
import { ApiList } from "./ApiList";
import { ApiDebug } from "./ApiDebug";
import { CaseList } from "./CaseList";
import { CaseDetail } from "./CaseDetail";
export function ApiTesting({ scope, role }: Context) {
  const location = useLocation();
  const go = useScopedNavigate();
  const apis = useAssets(scope, "api"),
    folders = useAssets(scope, "folder");
  const caseMatch = matchPath("/api-testing/cases/:id", location.pathname);
  const selectedCase = useResource(scope, "case", caseMatch?.params.id);
  const apiMatch = matchPath("/api-testing/apis/:id/*", location.pathname);
  const parentId =
    apiMatch?.params.id !== "new"
      ? apiMatch?.params.id
      : selectedCase.data?.document.api_id;
  const actualParent = parentId || selectedCase.data?.document.api_id;
  const parent = useResource(scope, "api", actualParent);
  const [search, setSearch] = useState("");
  const [folderEdit, setFolderEdit] = useState<FolderDraft | null>(null);
  const [expanded, setExpanded] = useState<React.Key[]>([]);
  const action = useAction();
  const allFolders: Asset<"folder">[] = (folders.data?.items || []).filter(
    (v) => v.document.kind === "api",
  );
  const allApis: Asset<"api">[] = [...(apis.data?.items || [])];
  if (parent.data && !allApis.some((v) => v.id === parent.data.id))
    allApis.push(parent.data);
  const filtered = allApis.filter((v) =>
    v.name.toLowerCase().includes(search.toLowerCase()),
  );
  const selectedFolder = matchPath(
    "/api-testing/folders/:id/*",
    location.pathname,
  )?.params.id;
  useEffect(() => {
    const ids: string[] = [];
    let current = parent.data?.folder_id;
    while (current && !ids.includes(current)) {
      ids.push(current);
      current = (folders.data?.items || []).find(
        (v) => v.id === current,
      )?.folder_id;
    }
    setExpanded((old) => [
      ...new Set([...old, ...ids.map((id) => `folder:${id}`)]),
    ]);
  }, [parent.data?.folder_id, folders.data]);
  function branches(id: string, seen: string[] = []): DataNode[] {
    if (seen.length > 32) return [];
    const children = allFolders
      .filter((v) => v.folder_id === id && !seen.includes(v.id))
      .map((v) => ({
        key: `folder:${v.id}`,
        title: v.name,
        children: branches(v.id, [...seen, v.id]),
      }))
      .filter(
        (v) =>
          !search ||
          v.children.length ||
          v.title.toLowerCase().includes(search.toLowerCase()),
      );
    return [
      ...children,
      ...filtered
        .filter((v) => v.folder_id === id)
        .map((v) => ({
          key: `api:${v.id}`,
          title: (
            <span className="api-tree-label">
              <Tag color="blue">{v.document.request?.method}</Tag>
              <span>{v.name}</span>
            </span>
          ),
          isLeaf: true,
        })),
    ];
  }
  const orphanApis = filtered.filter(
    (v) => v.folder_id && !allFolders.some((f) => f.id === v.folder_id),
  );
  const nodes = branches("");
  for (const v of orphanApis)
    nodes.push({ key: `api:${v.id}`, title: v.name, isLeaf: true });
  const active = selectedFolder
    ? `folder:${selectedFolder}`
    : parent.data
      ? `api:${parent.data.id}`
      : location.pathname.includes("unlinked") || caseMatch
        ? "unlinked"
        : "all";
  return (
    <div className="api-testing-layout">
      <aside className="folder-panel api-tree-panel">
        <div className="toolbar">
          <Typography.Text strong>API测试</Typography.Text>
          {editable(role) && (
            <Button
              size="small"
              onClick={() =>
                setFolderEdit({ name: "", folder_id: selectedFolder || "" })
              }
            >
              新建目录
            </Button>
          )}
        </div>
        <Input.Search
          aria-label="搜索接口树"
          placeholder="搜索接口或目录"
          value={search}
          allowClear
          onChange={(e) => setSearch(e.target.value)}
        />
        <ErrorState error={apis.error || folders.error} />
        <Tree
          blockNode
          showLine
          selectedKeys={[active]}
          expandedKeys={
            search ? allFolders.map((v) => `folder:${v.id}`) : expanded
          }
          onExpand={setExpanded}
          treeData={[
            { key: "all", title: "全部接口" },
            ...nodes,
            { key: "unlinked", title: "未关联接口用例", isLeaf: true },
          ]}
          onSelect={(keys) => {
            const key = String(keys[0] || "");
            if (key === "all") go("/api-testing");
            else if (key === "unlinked") go("/api-testing/unlinked");
            else if (key.startsWith("api:"))
              go(`/api-testing/apis/${key.slice(4)}/debug`);
            else if (key.startsWith("folder:"))
              go(`/api-testing/folders/${key.slice(7)}`);
          }}
          onDoubleClick={(_, node) => {
            const folder = allFolders.find(
              (v) => `folder:${v.id}` === node.key,
            );
            if (folder && editable(role)) setFolderEdit(folder);
          }}
        />
        {selectedFolder && editable(role) && (
          <Space wrap className="folder-actions">
            <Button
              size="small"
              onClick={() => {
                const f = allFolders.find((v) => v.id === selectedFolder);
                if (f) setFolderEdit(f);
              }}
            >
              编辑目录
            </Button>
            <Popconfirm
              title="仅可删除空目录，确认删除？"
              onConfirm={() =>
                action.run(async () => {
                  await resources.remove(scope, "folder", selectedFolder);
                  go("/api-testing");
                }, "目录已删除")
              }
            >
              <Button danger size="small">
                删除目录
              </Button>
            </Popconfirm>
          </Space>
        )}
      </aside>
      <section className="api-testing-content">
        <Routes>
          <Route index element={<ApiList scope={scope} role={role} />} />
          <Route
            path="folders/:folderId"
            element={<ApiList scope={scope} role={role} />}
          />
          <Route
            path="apis/new"
            element={<ApiDebug scope={scope} role={role} />}
          />
          <Route
            path="folders/:folderId/new"
            element={<ApiDebug scope={scope} role={role} />}
          />
          <Route
            path="apis/:apiId/debug"
            element={<ApiDebug scope={scope} role={role} />}
          />
          <Route
            path="apis/:apiId/cases"
            element={<CaseList scope={scope} role={role} />}
          />
          <Route
            path="apis/:apiId/cases/new"
            element={<CaseDetail scope={scope} role={role} />}
          />
          <Route
            path="cases/:caseId"
            element={<CaseDetail scope={scope} role={role} />}
          />
          <Route
            path="unlinked"
            element={<CaseList scope={scope} role={role} unlinked />}
          />
          <Route
            path="*"
            element={
              <Empty description="页面不存在">
                <Button onClick={() => go("/api-testing")}>返回API测试</Button>
              </Empty>
            }
          />
        </Routes>
      </section>
      {folderEdit && (
        <FolderDialog
          key={folderEdit.id || "new"}
          row={folderEdit}
          kind="api"
          scope={scope}
          folders={folders.data?.items || []}
          onClose={() => setFolderEdit(null)}
        />
      )}
    </div>
  );
}

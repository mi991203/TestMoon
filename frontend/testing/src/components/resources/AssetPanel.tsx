import type { ResourceDraft, FolderDraft } from "../../models/editors";
import type { ContractsSelection } from "@testmoon/api";
import type { DataNode } from "antd/es/tree";
import type { Asset } from "../../models/resources";
import { resources } from "../../services/resources";

import { useState } from "react";
import { useLocation, useParams } from "react-router-dom";
import {
  Button,
  Card,
  Input,
  Popconfirm,
  Space,
  Spin,
  Table,
  Tree,
  Typography,
} from "antd";
import { ErrorState, PageHeading, Scope, useAction } from "@testmoon/web";

import { useAssets, useResource } from "../../hooks/queries";
import { useScopedNavigate } from "../../hooks/navigation";
import { titles, emptyDocument } from "./defaults";
import { ResourceEditor } from "./ResourceEditor";
import { RunDialog } from "./RunDialog";
import { FolderDialog } from "./FolderDialog";
export function AssetPanel({
  kind,
  scope,
  role,
}: {
  kind: string;
  scope: Scope;
  role: string;
}) {
  const { id } = useParams();
  const location = useLocation();
  const go = useScopedNavigate();
  const query = useAssets(scope, kind),
    apis = useAssets(scope, "api"),
    cases = useAssets(scope, "case"),
    folders = useAssets(scope, "folder"),
    detail = useResource(scope, kind, id);
  const [folder, setFolder] = useState("*");
  const [search, setSearch] = useState("");
  const [folderEdit, setFolderEdit] = useState<FolderDraft | null>(null);
  const [selection, setSelection] =
    useState<Partial<ContractsSelection> | null>(null);
  const action = useAction();
  const readOnly = role !== "admin" && role !== "editor";
  const allFolders: Asset<"folder">[] = (folders.data?.items || []).filter(
    (v) => v.document.kind === kind,
  );
  const root = `/${kind}`;
  const back = () => go(root);
  if (id) {
    if (id !== "new" && detail.isPending) return <Spin />;
    if (id !== "new" && detail.error)
      return (
        <>
          <Button onClick={back}>返回列表</Button>
          <ErrorState error={detail.error} />
        </>
      );
    const row: ResourceDraft =
      id === "new"
        ? {
            name: "",
            folder_id: location.state?.folder || "",
            document: emptyDocument(kind),
          }
        : detail.data!;
    return (
      <ResourceEditor
        key={`${id}/${row.revision || 0}`}
        row={row}
        kind={kind}
        scope={scope}
        role={role}
        apis={apis.data?.items || []}
        cases={cases.data?.items || []}
        folders={folders.data?.items || []}
        back={back}
        onSaved={(v) => go(`${root}/${v.id}`, id === "new")}
      />
    );
  }
  function tree(parent: string, seen: string[] = []): DataNode[] {
    return allFolders
      .filter((v) => v.folder_id === parent && !seen.includes(v.id))
      .map((v) => ({
        key: v.id,
        title: v.name,
        children: tree(v.id, [...seen, v.id]),
      }));
  }
  return (
    <>
      <PageHeading
        title={kind === "environment" ? "环境管理" : titles[kind]}
        extra={
          !readOnly && (
            <Button
              type="primary"
              onClick={() =>
                go(`${root}/new`, false, {
                  folder: folder === "*" ? "" : folder,
                })
              }
            >
              新建{titles[kind]}
            </Button>
          )
        }
      />
      <ErrorState error={query.error} />
      <div className={kind === "environment" ? "" : "asset-grid"}>
        {kind !== "environment" && (
          <aside className="folder-panel">
            <Space>
              <Typography.Text strong>资源目录</Typography.Text>
              {!readOnly && (
                <Button
                  size="small"
                  onClick={() =>
                    setFolderEdit({
                      name: "",
                      folder_id: folder === "*" ? "" : folder,
                    })
                  }
                >
                  新建
                </Button>
              )}
            </Space>
            <Tree
              selectedKeys={[folder]}
              treeData={[
                { key: "*", title: "全部资源" },
                { key: "", title: "根目录", children: tree("") },
              ]}
              onSelect={(keys) => setFolder(String(keys[0] ?? "*"))}
              onDoubleClick={(_, node) => {
                const row = allFolders.find((v) => v.id === node.key);
                if (row && !readOnly) setFolderEdit(row);
              }}
            />
            {folder && folder !== "*" && !readOnly && (
              <Popconfirm
                title="确认删除空目录？"
                onConfirm={() =>
                  action.run(async () => {
                    await resources.remove(scope, "folder", folder);
                    setFolder("*");
                  }, "目录已删除")
                }
              >
                <Button danger size="small">
                  删除当前目录
                </Button>
              </Popconfirm>
            )}
          </aside>
        )}
        <Card>
          <div className="toolbar">
            <Input.Search
              placeholder="按名称搜索"
              allowClear
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
            <Button onClick={() => query.refetch()}>刷新</Button>
          </div>
          <Table<Asset>
            rowKey="id"
            loading={query.isLoading}
            dataSource={(query.data?.items || []).filter(
              (v) =>
                (folder === "*" || v.folder_id === folder) &&
                v.name.toLowerCase().includes(search.toLowerCase()),
            )}
            columns={[
              {
                title: "名称",
                dataIndex: "name",
                render: (name, v) => (
                  <a onClick={() => go(`${root}/${v.id}`)}>{name}</a>
                ),
              },
              { title: "版本", dataIndex: "revision" },
              {
                title: "操作",
                render: (_, v) => (
                  <Space>
                    <Button onClick={() => go(`${root}/${v.id}`)}>
                      {readOnly ? "查看" : "编辑"}
                    </Button>
                    {!readOnly && (
                      <>
                        {kind === "scenario" && (
                          <Button
                            onClick={() => setSelection({ kind, id: v.id })}
                          >
                            执行
                          </Button>
                        )}
                        <Popconfirm
                          title="确认软删除资源？"
                          onConfirm={() =>
                            action.run(
                              () => resources.remove(scope, kind, v.id),
                              "已删除",
                            )
                          }
                        >
                          <Button danger>删除</Button>
                        </Popconfirm>
                      </>
                    )}
                  </Space>
                ),
              },
            ]}
          />
        </Card>
      </div>
      {folderEdit && (
        <FolderDialog
          row={folderEdit}
          kind={kind}
          scope={scope}
          folders={folders.data?.items || []}
          onClose={() => setFolderEdit(null)}
        />
      )}{" "}
      {selection && (
        <RunDialog
          scope={scope}
          selection={selection}
          onClose={() => setSelection(null)}
          onRun={(id) => go(`/reports/${id}`)}
        />
      )}
    </>
  );
}

import type { Asset } from "../../models/resources";
import { resources } from "../../services/resources";
import { useState } from "react";
import { useParams } from "react-router-dom";
import {
  Button,
  Card,
  Input,
  Popconfirm,
  Space,
  Table,
  Tag,
  Typography,
} from "antd";
import { ErrorState, PageHeading, useAction } from "@testmoon/web";
import { useAssets } from "../../hooks/queries";
import { useScopedNavigate } from "../../hooks/navigation";

import { Context, editable } from "./context";

export function ApiList({ scope, role }: Context) {
  const { folderId } = useParams();
  const go = useScopedNavigate();
  const query = useAssets(scope, "api"),
    folders = useAssets(scope, "folder");
  const [search, setSearch] = useState("");
  const { run } = useAction();
  const descendants = new Set(folderId ? [folderId] : []);
  const pending = [...descendants];
  for (let i = 0; i < pending.length; i++) {
    for (const folder of folders.data?.items || []) {
      if (folder.document.kind === "api" && folder.folder_id === pending[i] && !descendants.has(folder.id)) {
        descendants.add(folder.id);
        pending.push(folder.id);
      }
    }
  }
  const items: Asset<"api">[] = (query.data?.items || []).filter(
    (v) =>
      (!folderId || descendants.has(v.folder_id)) &&
      v.name.toLowerCase().includes(search.toLowerCase()),
  );
  return (
    <>
      <PageHeading
        title={
          folderId
            ? folders.data?.items?.find((v) => v.id === folderId)?.name ||
              "接口目录"
            : "API测试"
        }
        description="选择接口进行调试，或查看与它关联的接口用例。"
        extra={
          editable(role) && (
            <Button
              type="primary"
              onClick={() =>
                go(
                  folderId
                    ? `/api-testing/folders/${folderId}/new`
                    : "/api-testing/apis/new",
                )
              }
            >
              新建接口
            </Button>
          )
        }
      />
      <ErrorState error={query.error} />
      <Card>
        <div className="toolbar">
          <Input.Search
            placeholder="按接口名称搜索"
            value={search}
            allowClear
            onChange={(e) => setSearch(e.target.value)}
          />
          <Button onClick={() => query.refetch()}>刷新</Button>
        </div>
        <Table<Asset<"api">>
          rowKey="id"
          loading={query.isLoading}
          dataSource={items}
          pagination={{ pageSize: 12 }}
          columns={[
            {
              title: "接口",
              dataIndex: "name",
              render: (name, v) => (
                <a onClick={() => go(`/api-testing/apis/${v.id}/debug`)}>
                  {name}
                </a>
              ),
            },
            {
              title: "请求",
              render: (_, v) => (
                <Space>
                  <Tag>{v.document.request?.method}</Tag>
                  <Typography.Text ellipsis style={{ maxWidth: 250 }}>
                    {v.document.request?.url}
                  </Typography.Text>
                </Space>
              ),
            },
            {
              title: "操作",
              render: (_, v) => (
                <Space wrap>
                  <Button onClick={() => go(`/api-testing/apis/${v.id}/debug`)}>
                    API调试
                  </Button>
                  <Button onClick={() => go(`/api-testing/apis/${v.id}/cases`)}>
                    接口用例
                  </Button>
                  {editable(role) && (
                    <Popconfirm
                      title="确认删除接口？已被用例引用时需先解除关联。"
                      onConfirm={() =>
                        run(
                          () => resources.remove(scope, "api", v.id),
                          "已删除",
                        )
                      }
                    >
                      <Button danger>删除</Button>
                    </Popconfirm>
                  )}
                </Space>
              ),
            },
          ]}
        />
      </Card>
    </>
  );
}

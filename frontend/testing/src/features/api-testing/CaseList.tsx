import { execution } from "@testmoon/api";

import type { Asset } from "../../models/resources";
import { resources } from "../../services/resources";
import { useEffect, useRef, useState } from "react";
import { useParams } from "react-router-dom";
import { Button, Card, Input, Popconfirm, Space, Spin, Table } from "antd";
import { ErrorState, PageHeading, useAction } from "@testmoon/web";
import { useAssets, useResource, useRun } from "../../hooks/queries";
import { useScopedNavigate } from "../../hooks/navigation";
import { RunResult } from "../../components/resources/RunResult";
import { active } from "../reports/state";

import { Context, editable } from "./context";
import { ApiHeader } from "./ApiHeader";
export function CaseList({
  scope,
  role,
  unlinked = false,
}: Context & { unlinked?: boolean }) {
  const { apiId } = useParams();
  const parent = useResource(scope, "api", apiId);
  const query = useAssets(scope, "case", unlinked ? "" : apiId);
  const [search, setSearch] = useState("");
  const [runId, setRunId] = useState("");
  const [executionError, setExecutionError] = useState<unknown>();
  const result = useRun(scope, runId);
  const resultRef = useRef<HTMLDivElement>(null);
  const running = !!runId && (result.isPending || active(result.data?.run.state || ""));
  const go = useScopedNavigate();
  const { busy, run } = useAction();
  useEffect(() => {
    setRunId("");
    setExecutionError(undefined);
  }, [apiId, unlinked]);
  useEffect(() => {
    if (runId) resultRef.current?.scrollIntoView({ behavior: "smooth", block: "start" });
  }, [runId]);
  if (apiId && parent.isPending) return <Spin />;
  if (parent.error) return <ErrorState error={parent.error} />;
  return (
    <>
      <ApiHeader id={apiId} name={parent.data?.name} tab="cases" />
      <PageHeading
        title={
          unlinked
            ? "未关联接口用例"
            : `${parent.data?.name || "接口"} · 接口用例`
        }
        description={
          unlinked
            ? "保留历史独立用例及接口已不存在的用例，可打开详情重新关联接口。"
            : "当前列表只展示所选接口关联的用例。"
        }
        extra={
          apiId &&
          editable(role) && (
            <Button
              type="primary"
              onClick={() => go(`/api-testing/apis/${apiId}/cases/new`)}
            >
              新建接口用例
            </Button>
          )
        }
      />
      <ErrorState error={query.error} />
      <Card>
        <div className="toolbar">
          <Input.Search
            placeholder="按用例名称搜索"
            value={search}
            allowClear
            onChange={(e) => setSearch(e.target.value)}
          />
          <Button onClick={() => query.refetch()}>刷新</Button>
        </div>
        <Table<Asset<"case">>
          rowKey="id"
          loading={query.isLoading}
          dataSource={(query.data?.items || []).filter((v) =>
            v.name.toLowerCase().includes(search.toLowerCase()),
          )}
          columns={[
            {
              title: "用例",
              dataIndex: "name",
              render: (name, v) => (
                <a onClick={() => go(`/api-testing/cases/${v.id}`)}>{name}</a>
              ),
            },
            { title: "版本", dataIndex: "revision" },
            {
              title: "操作",
              render: (_, v) => (
                <Space>
                  <Button onClick={() => go(`/api-testing/cases/${v.id}`)}>
                    {editable(role) ? "编辑" : "查看"}
                  </Button>
                  {editable(role) && (
                    <>
                      <Button
                        disabled={busy || running}
                        onClick={() => {
                          setExecutionError(undefined);
                          run(async () => {
                            try {
                              const out = await execution.runCase({
                                ...scope,
                                case_id: v.id,
                                environment_id: v.document.environment_id || "",
                                runner_id: v.document.runner_id || "public",
                                idempotency_key: crypto.randomUUID(),
                              });
                              setRunId(out.run.id);
                            } catch (error) {
                              setExecutionError(error);
                              throw error;
                            }
                          }, "执行已提交");
                        }}
                      >
                        执行
                      </Button>
                      <Popconfirm
                        title="确认删除接口用例？"
                        onConfirm={() =>
                          run(
                            () => resources.remove(scope, "case", v.id),
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
      <ErrorState error={executionError} />
      {runId && <div ref={resultRef} className="inline-run-result">
        <RunResult scope={scope} id={runId} onReport={() => go(`/reports/${runId}`)} />
      </div>}
    </>
  );
}

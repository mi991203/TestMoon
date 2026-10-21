import { Button, Card, Space } from "antd";
import { ErrorState, JSONView, PageHeading, Scope } from "@testmoon/web";

import { useParams } from "react-router-dom";
import { useScopedNavigate } from "../../hooks/navigation";
import { useResource } from "../../hooks/queries";

export function DocumentRoute({ scope }: { scope: Scope }) {
  const { id } = useParams();
  const query = useResource(scope, "api", id);
  const go = useScopedNavigate();
  return (
    <>
      <PageHeading
        title={query.data?.name || "接口文档"}
        extra={
          <Space>
            <Button onClick={() => go("/documents")}>返回文档列表</Button>
            <Button
              disabled={!query.data}
              onClick={() => go(`/api-testing/apis/${id}/debug`)}
            >
              API调试
            </Button>
          </Space>
        }
      />
      <ErrorState error={query.error} />
      <Card loading={query.isPending}>
        <JSONView value={query.data?.document} />
      </Card>
    </>
  );
}

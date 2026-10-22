import type { Asset } from "../../models/resources";
import type { ResourceDraft } from "../../models/editors";

import { useParams } from "react-router-dom";
import { Alert, Spin } from "antd";
import { ErrorState } from "@testmoon/web";
import { useAssets, useResource } from "../../hooks/queries";
import { useScopedNavigate } from "../../hooks/navigation";
import { emptyDocument, ResourceEditor } from "../../components/resources";

import { Context } from "./context";
import { ApiHeader } from "./ApiHeader";
export function CaseDetail({ scope, role }: Context) {
  const { caseId, apiId } = useParams();
  const query = useResource(scope, "case", caseId),
    apis = useAssets(scope, "api"),
    folders = useAssets(scope, "folder");
  const go = useScopedNavigate();
  const owner = apiId || query.data?.document.api_id;
  const parent = useResource(scope, "api", owner);
  if ((caseId && query.isPending) || (owner && parent.isPending))
    return <Spin />;
  if (caseId && query.error) return <ErrorState error={query.error} />;
  if (apiId && parent.error) return <ErrorState error={parent.error} />;
  const row: ResourceDraft = caseId
    ? query.data!
    : { name: "", document: emptyDocument("case", apiId) };
  const all: Asset<"api">[] = [...(apis.data?.items || [])];
  if (parent.data && !all.some((v) => v.id === parent.data.id))
    all.push(parent.data);
  const existing = !!parent.data;
  return (
    <>
      <ApiHeader
        id={existing ? owner : undefined}
        name={parent.data?.name}
        tab="cases"
      />
      {caseId && owner && parent.error && (
        <Alert
          type="warning"
          showIcon
          message="原关联接口已不存在，请重新关联接口或补充独立请求后保存。"
        />
      )}
      <ResourceEditor
        key={`${caseId || `new-${apiId}`}/${row.revision || 0}`}
        row={row}
        kind="case"
        scope={scope}
        role={role}
        apis={all}
        cases={[]}
        folders={folders.data?.items || []}
        back={() =>
          go(
            existing
              ? `/api-testing/apis/${owner}/cases`
              : "/api-testing/unlinked",
          )
        }
        onSaved={(v) => go(`/api-testing/cases/${v.id}`, !caseId)}
      />
    </>
  );
}

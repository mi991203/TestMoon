import type { ResourceDraft, CaseDraft } from "../../models/editors";

import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { Spin } from "antd";
import { ErrorState } from "@testmoon/web";
import { useAssets, useResource } from "../../hooks/queries";
import { useScopedNavigate } from "../../hooks/navigation";
import { emptyDocument, ResourceEditor } from "../../components/resources";

import { Context } from "./context";
import { ApiHeader } from "./ApiHeader";
export function ApiDebug({ scope, role }: Context) {
  const { apiId, folderId } = useParams();
  const query = useResource(scope, "api", apiId),
    apis = useAssets(scope, "api"),
    folders = useAssets(scope, "folder");
  const go = useScopedNavigate();
  const [debug, setDebug] = useState<CaseDraft>({ assertions: [] });
  useEffect(() => setDebug({ assertions: [] }), [apiId]);
  if (apiId && query.isPending) return <Spin />;
  if (apiId && query.error) return <ErrorState error={query.error} />;
  const row: ResourceDraft = apiId
    ? query.data!
    : { name: "", folder_id: folderId || "", document: emptyDocument("api") };
  return (
    <>
      {apiId && <ApiHeader id={apiId} name={row.name} tab="debug" />}
      <ResourceEditor
        key={`${apiId || "new"}/${row.revision || 0}`}
        row={row}
        kind="api"
        scope={scope}
        role={role}
        apis={apis.data?.items || []}
        cases={[]}
        folders={folders.data?.items || []}
        debug={debug}
        onDebugChange={setDebug}
        back={() => go("/api-testing")}
        onSaved={(v) => go(`/api-testing/apis/${v.id}/debug`, !apiId)}
      />
    </>
  );
}

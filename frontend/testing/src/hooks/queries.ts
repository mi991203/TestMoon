import { useQuery } from "@tanstack/react-query";
import { execution, type Scope } from "@testmoon/api";
import type { Asset, AssetKind } from "../models/resources";
import { resources } from "../services/resources";
export function useAssets<K extends string>(
  scope: Scope,
  kind: K,
  apiId?: string,
) {
  return useQuery({
    queryKey: ["assets", scope.org_id, scope.project_id, kind, apiId],
    enabled: !!scope.org_id && !!scope.project_id,
    queryFn: async () =>
      (await resources.list(scope, kind, apiId)) as {
        items: Asset<K extends AssetKind ? K : AssetKind>[];
      },
  });
}
export function useResource<K extends string>(
  scope: Scope,
  kind: K,
  id?: string,
) {
  return useQuery({
    queryKey: ["resource", scope.org_id, scope.project_id, kind, id],
    enabled: !!scope.org_id && !!scope.project_id && !!id && id !== "new",
    queryFn: async () =>
      (await resources.get(scope, kind, id!)) as Asset<
        K extends AssetKind ? K : AssetKind
      >,
  });
}
export function useRunners(scope: Scope) {
  return useQuery({
    queryKey: ["runners", scope.org_id, scope.project_id],
    enabled: !!scope.project_id,
    queryFn: async () => {
      const result = await execution.listProjectRunners(scope);
      return { ...result, items: result.items.filter((v) => !v.revoked) };
    },
    refetchInterval: 10000,
  });
}

export function useRun(scope: Scope, id?: string) {
  return useQuery({
    queryKey: ["run", scope.org_id, scope.project_id, id],
    enabled: !!id,
    queryFn: () => execution.getRun({ ...scope, id: id! }),
    refetchInterval: (q) =>
      ["created", "queued", "running"].includes(q.state.data?.run.state || "")
        ? 2000
        : false,
  });
}

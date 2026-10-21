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
    queryFn: () => execution.listProjectRunners(scope),
    refetchInterval: 10000,
  });
}

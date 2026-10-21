import {
  core,
  type Scope,
  type Input,
  type CoreCreateAPIRequest,
  type CoreCreateTestCaseRequest,
  type CoreCreateScenarioRequest,
  type CoreCreateEnvironmentRequest,
  type CoreCreateFolderRequest,
} from "@testmoon/api";

import type { AssetKind, Asset, AssetDraft } from "../models/resources";
function resource<K extends AssetKind>(
  kind: K,
  value: Omit<Asset<K>, "kind">,
): Asset<K> {
  return { ...value, kind };
}
/** Adapts concrete API contracts to the editor's resource view model. */
export const resources = {
  async list(
    scope: Scope,
    kind: string,
    apiID?: string,
  ): Promise<{ items: Asset[] }> {
    switch (kind) {
      case "api":
        return {
          items: (await core.listAPIs(scope)).items.map((v) =>
            resource("api", v),
          ),
        };
      case "case":
        return {
          items: (
            await core.listTestCases({
              ...scope,
              ...(apiID !== undefined ? { api_id: apiID } : {}),
            })
          ).items.map((v) => resource("case", v)),
        };
      case "scenario":
        return {
          items: (await core.listScenarios(scope)).items.map((v) =>
            resource("scenario", v),
          ),
        };
      case "environment":
        return {
          items: (await core.listEnvironments(scope)).items.map((v) =>
            resource("environment", v),
          ),
        };
      case "folder":
        return {
          items: (await core.listFolders(scope)).items.map((v) =>
            resource("folder", v),
          ),
        };
      default:
        throw new Error("无效资源类型");
    }
  },
  async get(scope: Scope, kind: string, id: string): Promise<Asset> {
    const request = { ...scope, id };
    switch (kind) {
      case "api":
        return resource("api", (await core.getAPI(request)).resource);
      case "case":
        return resource("case", (await core.getTestCase(request)).resource);
      case "scenario":
        return resource("scenario", (await core.getScenario(request)).resource);
      case "environment":
        return resource(
          "environment",
          (await core.getEnvironment(request)).resource,
        );
      case "folder":
        return resource("folder", (await core.getFolder(request)).resource);
      default:
        throw new Error("无效资源类型");
    }
  },
  async save(
    scope: Scope,
    kind: string,
    draft: AssetDraft,
    id?: string,
  ): Promise<{ resource: Asset }> {
    const request = {
      ...scope,
      name: draft.name,
      folder_id: draft.folder_id,
      document: draft.document,
      ...(id ? { id, revision: draft.revision } : {}),
    };
    switch (kind) {
      case "api":
        return {
          resource: resource(
            "api",
            (
              await (id ? core.updateAPI : core.createAPI)(
                request as Input<CoreCreateAPIRequest>,
              )
            ).resource,
          ),
        };
      case "case":
        return {
          resource: resource(
            "case",
            (
              await (id ? core.updateTestCase : core.createTestCase)(
                request as Input<CoreCreateTestCaseRequest>,
              )
            ).resource,
          ),
        };
      case "scenario":
        return {
          resource: resource(
            "scenario",
            (
              await (id ? core.updateScenario : core.createScenario)(
                request as Input<CoreCreateScenarioRequest>,
              )
            ).resource,
          ),
        };
      case "environment":
        return {
          resource: resource(
            "environment",
            (
              await (id ? core.updateEnvironment : core.createEnvironment)(
                request as Input<CoreCreateEnvironmentRequest>,
              )
            ).resource,
          ),
        };
      case "folder":
        return {
          resource: resource(
            "folder",
            (
              await (id ? core.updateFolder : core.createFolder)(
                request as Input<CoreCreateFolderRequest>,
              )
            ).resource,
          ),
        };
      default:
        throw new Error("无效资源类型");
    }
  },
  remove(scope: Scope, kind: string, id: string) {
    const request = { ...scope, id };
    switch (kind) {
      case "api":
        return core.deleteAPI(request);
      case "case":
        return core.deleteTestCase(request);
      case "scenario":
        return core.deleteScenario(request);
      case "environment":
        return core.deleteEnvironment(request);
      case "folder":
        return core.deleteFolder(request);
      default:
        throw new Error("无效资源类型");
    }
  },
};

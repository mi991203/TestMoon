import type { Asset } from "../../models/resources";
import type { EditableDefinition } from "../../models/editors";

import { blankRequest } from "../editors";

export const titles: Record<string, string> = {
  api: "接口",
  case: "接口用例",
  scenario: "场景用例",
  environment: "环境",
};
export const emptyDocument = (kind: string, apiId = ""): EditableDefinition =>
  kind === "api"
    ? { protocol: "http", request: blankRequest(), description: "" }
    : kind === "scenario"
      ? { steps: [] }
      : kind === "environment"
        ? { variables: {} }
        : { api_id: apiId, request: blankRequest() };
export function folderOptions(folders: Asset<"folder">[], kind: string) {
  const all = folders.filter((v) => v.document.kind === kind);
  const name = (id: string, seen: string[] = []): string => {
    const f = all.find((v) => v.id === id);
    return !f || seen.includes(id)
      ? ""
      : (f.folder_id ? name(f.folder_id, [...seen, id]) + "/" : "") + f.name;
  };
  return [
    { value: "", label: "根目录" },
    ...all.map((v) => ({ value: v.id, label: name(v.id) })),
  ];
}

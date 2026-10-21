import type {
  Input,
  Int64,
  ContractsAPIDefinition,
  ContractsTestCaseDefinition,
  ContractsScenarioDefinition,
  ContractsEnvironmentDefinition,
  ContractsFolderDefinition,
} from "@testmoon/api";
export interface Definitions {
  api: ContractsAPIDefinition;
  case: ContractsTestCaseDefinition;
  scenario: ContractsScenarioDefinition;
  environment: ContractsEnvironmentDefinition;
  folder: ContractsFolderDefinition;
}
export type AssetKind = keyof Definitions;
export type Asset<K extends AssetKind = AssetKind> = {
  id: string;
  name: string;
  kind: K;
  folder_id: string;
  revision: Int64;
  updated_at: string;
  document: Definitions[K];
};
export type AssetDraft = {
  name?: string;
  folder_id?: string;
  revision?: Int64;
  document?: Input<Definitions[AssetKind]>;
  kind?: string;
};

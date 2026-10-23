import type {
  ContractsAPIDefinition,
  ContractsHTTPRequest,
  ContractsAuth,
  ContractsRequestPatch,
  ContractsTestCaseDefinition,
  ContractsScenarioDefinition,
  ContractsScenarioStep,
  ContractsEnvironmentDefinition,
  ContractsFolderDefinition,
  ContractsVariable,
  Int64,
  ContractsMockResponse,
  ContractsMockCondition,
} from "@testmoon/api";
import type { Asset } from "../models/resources";
export type RequestDraft = Partial<Omit<ContractsHTTPRequest, "auth">> & {
  auth?: Partial<ContractsAuth>;
};
export type PatchDraft = Partial<Omit<ContractsRequestPatch, "auth">> & {
  auth?: Partial<ContractsAuth>;
};
export type APIDraft = Partial<Omit<ContractsAPIDefinition, "request">> & {
  request?: RequestDraft;
};
export type CaseDraft = Partial<
  Omit<ContractsTestCaseDefinition, "request" | "override">
> & { request?: RequestDraft; override?: PatchDraft };
export type StepDraft = Partial<
  Omit<ContractsScenarioStep, "inline" | "override" | "children">
> & { inline?: CaseDraft; override?: PatchDraft; children?: StepDraft[] };
export type ScenarioDraft = Partial<
  Omit<ContractsScenarioDefinition, "steps">
> & { steps?: StepDraft[] };
export type EnvironmentDraft = Partial<ContractsEnvironmentDefinition>;
export type MockDraft = Partial<ContractsMockResponse>;
/** Shared editor shell fields; feature editors keep their concrete definitions. */
export type EditableDefinition = Partial<
  Omit<ContractsTestCaseDefinition, "request" | "variables" | "override"> &
    Omit<ContractsAPIDefinition, "request"> &
    Omit<ContractsScenarioDefinition, "steps"> &
    ContractsFolderDefinition
> & {
  request?: RequestDraft;
  override?: PatchDraft;
  steps?: StepDraft[];
  variables?: Record<string, string> | Record<string, ContractsVariable>;
};
export interface ResourceDraft {
  id?: string;
  name: string;
  folder_id?: string;
  revision?: Int64;
  document: EditableDefinition;
}
export type FolderDraft = Partial<Asset<"folder">>;
export type ConditionDraft = Partial<ContractsMockCondition>;

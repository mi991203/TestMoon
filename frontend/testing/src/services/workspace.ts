import {
  core,
  execution,
  mock,
  type Scope,
  type Input,
  type CoreCreateProjectRequest,
  type ExecutionUpdateTaskRequest,
  type MockCreateRuleRequest,
  type CoreSetProjectMemberRequest,
} from "@testmoon/api";
export const saveProject = (
  scope: Scope,
  values: Input<CoreCreateProjectRequest>,
  id?: string,
) =>
  (id ? core.updateProject : core.createProject)({
    ...scope,
    name: values.name,
    description: values.description,
    ...(id ? { id } : {}),
  });
export const saveTask = (
  scope: Scope,
  values: Input<ExecutionUpdateTaskRequest>,
  id?: string,
) =>
  (id ? execution.updateTask : execution.createTask)({
    ...scope,
    name: values.name,
    document: values.document,
    ...(id ? { id, revision: values.revision } : {}),
  });
export const saveRule = (
  scope: Scope,
  values: Input<MockCreateRuleRequest>,
  id?: string,
) =>
  (id ? mock.updateRule : mock.createRule)({
    ...scope,
    ...values,
    ...(id ? { id } : {}),
  });
export function setProjectMember(
  scope: Scope,
  values: Input<CoreSetProjectMemberRequest>,
) {
  return values.role === "remove"
    ? core.removeProjectMember({ ...scope, user_id: values.user_id })
    : core.setProjectMember({ ...scope, ...values });
}
export async function previewImport(
  request: Input<Parameters<typeof core.previewImport>[0]>,
) {
  const value = await core.previewImport(request);
  return {
    ...value,
    entries: value.entries.map((e) => ({
      ...e,
      document:
        e.api ??
        e.test_case ??
        e.scenario ??
        e.environment ??
        e.folder_definition,
    })),
  };
}

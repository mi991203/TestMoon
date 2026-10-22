import { Alert, Form, Select } from "antd";
import { ErrorState, type Scope } from "@testmoon/web";
import { useAssets, useRunners } from "../../hooks/queries";

export type ExecutionConfig = { environment_id?: string; runner_id?: string };

export function useExecutionOptions(scope: Scope, config: ExecutionConfig) {
  const envs = useAssets(scope, "environment");
  const runners = useRunners(scope);
  const runner = runners.data?.items.find((v) => v.id === (config.runner_id || "public"));
  const unavailable = runners.data && !runner
    ? "所选 Runner 已撤销或不可用，请重新选择。"
    : runner && !runner.online
      ? "所选 Runner 离线，请等待上线或重新选择。"
      : config.environment_id && envs.data && !envs.data.items.some((v) => v.id === config.environment_id)
        ? "所选环境已删除或不可用，请重新选择。"
        : "";
  return { envs, runners, unavailable, loading: envs.isPending || runners.isPending };
}

export function ExecutionFields({ value, onChange, readOnly, options }: {
  value: ExecutionConfig;
  onChange: (value: ExecutionConfig) => void;
  readOnly: boolean;
  options: ReturnType<typeof useExecutionOptions>;
}) {
  return (
    <>
      <div className="execution-fields">
        <Form.Item label="执行环境">
          <Select
            aria-label="执行环境"
            allowClear
            showSearch
            optionFilterProp="label"
            placeholder="不使用环境"
            disabled={readOnly}
            loading={options.envs.isPending}
            value={value.environment_id || undefined}
            labelRender={({ label }) => label || "环境已不可用"}
            options={(options.envs.data?.items || []).map((v) => ({ value: v.id, label: v.name }))}
            onChange={(environment_id) => onChange({ ...value, environment_id: environment_id || "" })}
          />
        </Form.Item>
        <Form.Item label="Runner">
          <Select
            aria-label="执行 Runner"
            showSearch
            optionFilterProp="label"
            disabled={readOnly}
            loading={options.runners.isPending}
            value={value.runner_id || "public"}
            labelRender={({ label }) => label || "Runner 已不可用"}
            options={(options.runners.data?.items || []).map((v) => ({
              value: v.id,
              label: `${v.name} · ${v.online ? "在线" : "离线"}`,
              disabled: !v.online,
            }))}
            onChange={(runner_id) => onChange({ ...value, runner_id })}
          />
        </Form.Item>
      </div>
      <ErrorState error={options.envs.error || options.runners.error} />
      {options.unavailable && <Alert type="warning" showIcon message={options.unavailable} />}
    </>
  );
}

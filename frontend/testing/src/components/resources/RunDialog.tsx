import type { ContractsSelection } from "@testmoon/api";
import type { CaseDraft } from "../../models/editors";

import { execution } from "@testmoon/api";

import { Alert, Form, Modal, Select } from "antd";
import { Scope, useAction } from "@testmoon/web";

import { useAssets, useRunners } from "../../hooks/queries";

export function RunDialog({
  scope,
  selection,
  debug,
  onClose,
  onRun,
}: {
  scope: Scope;
  selection?: Partial<ContractsSelection>;
  debug?: CaseDraft;
  onClose: () => void;
  onRun: (id: string) => void;
}) {
  const envs = useAssets(scope, "environment");
  const runners = useRunners(scope);
  const [form] = Form.useForm();
  const { busy, run } = useAction();
  return (
    <Modal
      title={debug ? "接口调试" : "执行测试"}
      open
      onCancel={onClose}
      onOk={() => form.submit()}
      confirmLoading={busy}
      width={debug ? 1000 : 540}
      destroyOnHidden
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={{ runner_id: "public", timeout_seconds: 900 }}
        onFinish={(v) =>
          run(async () => {
            const result = await (debug
              ? execution.debugHTTP({
                  ...scope,
                  ...v,
                  debug,
                  idempotency_key: crypto.randomUUID(),
                })
              : selection?.kind === "case"
                ? execution.runCase({
                    ...scope,
                    ...v,
                    case_id: selection.id,
                    idempotency_key: crypto.randomUUID(),
                  })
                : execution.runScenario({
                    ...scope,
                    ...v,
                    scenario_id: selection?.id,
                    idempotency_key: crypto.randomUUID(),
                  }));
            onRun(result.run.id);
            onClose();
          }, "执行已提交")
        }
      >
        <Form.Item label="环境" name="environment_id">
          <Select
            allowClear
            placeholder="不使用环境"
            options={(envs.data?.items || []).map((v) => ({
              value: v.id,
              label: v.name,
            }))}
          />
        </Form.Item>
        <Form.Item label="Runner" name="runner_id" rules={[{ required: true }]}>
          <Select
            options={(runners.data?.items || []).map((v) => ({
              value: v.id,
              label: `${v.name} · ${v.online ? "在线" : "离线"}`,
              disabled: !v.online,
            }))}
          />
        </Form.Item>
        <Alert
          type="info"
          showIcon
          message="离线 Runner 不接受新执行；运行中不会自动切换 Runner。"
        />
      </Form>
    </Modal>
  );
}

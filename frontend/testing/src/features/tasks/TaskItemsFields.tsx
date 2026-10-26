import { useState } from "react";
import { Form, Modal, Select } from "antd";
import type { ContractsSelection, Input } from "@testmoon/api";
import type { Asset } from "../../models/resources";
import { OrchestrationTree } from "../../components/orchestration/OrchestrationTree";
import { insertNode, updateNode } from "../../components/orchestration/tree";

export type TaskNode = Input<ContractsSelection>;

export function taskNodes(nodes: TaskNode[]): TaskNode[] {
  return nodes.map((node) => ({
    ...node,
    node_id: node.node_id || crypto.randomUUID(),
    ...(node.kind === "folder"
      ? { children: taskNodes(node.children || []) }
      : {}),
  }));
}

export function TaskItemsFields({
  value = [],
  onChange,
  cases,
  scenarios,
  readOnly,
}: {
  value?: TaskNode[];
  onChange?: (value: TaskNode[]) => void;
  cases: Asset<"case">[];
  scenarios: Asset<"scenario">[];
  readOnly: boolean;
}) {
  const [editing, setEditing] = useState<{ key?: string; parent: string }>();
  const [item, setItem] = useState<TaskNode>({ kind: "case" });
  const assets = item.kind === "scenario" ? scenarios : cases;
  const change = (nodes: TaskNode[]) => onChange?.(nodes);
  return (
    <>
      <OrchestrationTree
        value={value}
        onChange={change}
        readOnly={readOnly}
        title={(node) =>
          node.kind === "folder"
            ? node.name || "目录"
            : `${node.kind === "scenario" ? "场景" : "用例"} · ${(node.kind === "scenario" ? scenarios : cases).find((asset) => asset.id === node.id)?.name || "资源不可用"}`
        }
        createFolder={(name) => ({
          node_id: crypto.randomUUID(),
          kind: "folder",
          name,
          children: [],
        })}
        addLabel="添加测试项"
        onAdd={(parent) => {
          setItem({ kind: "case", node_id: crypto.randomUUID() });
          setEditing({ parent });
        }}
        onEdit={(node) => {
          setItem(structuredClone(node));
          setEditing({ key: node.node_id, parent: "" });
        }}
      />
      <Modal
        title="任务测试项"
        open={!!editing}
        onCancel={() => setEditing(undefined)}
        okButtonProps={{
          disabled:
            readOnly ||
            !item.id ||
            !assets.some((asset) => asset.id === item.id),
        }}
        onOk={() => {
          if (!editing || !item.id) return;
          change(
            editing.key
              ? updateNode(value, editing.key, () => item)
              : insertNode(value, editing.parent, item),
          );
          setEditing(undefined);
        }}
        destroyOnHidden
      >
        <Form layout="vertical" component={false}>
          <Form.Item label="测试项类型">
            <Select
              aria-label="测试项类型"
              disabled={readOnly}
              value={item.kind}
              options={[
                { value: "case", label: "接口用例" },
                { value: "scenario", label: "场景用例" },
              ]}
              onChange={(kind) => setItem({ node_id: item.node_id, kind })}
            />
          </Form.Item>
          <Form.Item label="测试项">
            <Select
              aria-label="选择测试项"
              disabled={readOnly}
              showSearch
              optionFilterProp="label"
              placeholder="选择一个测试项"
              value={item.id}
              options={assets.map((asset) => ({
                value: asset.id,
                label: asset.name,
              }))}
              onChange={(id) => setItem({ ...item, id })}
            />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}

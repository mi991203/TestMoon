import { useState } from "react";
import {
  Button,
  Empty,
  Input,
  Modal,
  Popconfirm,
  Space,
  Tag,
  Tree,
  Typography,
} from "antd";
import type { DataNode } from "antd/es/tree";
import { FileOutlined, FolderOutlined } from "@ant-design/icons";
import {
  findNode,
  folderKeys,
  insertNode,
  leafNodes,
  moveNode,
  nodeKey,
  parentKey,
  removeNode,
  updateNode,
  type ArrangementNode,
} from "./tree";
import "./orchestration.css";

export function OrchestrationTree<T extends ArrangementNode<T>>({
  value,
  onChange,
  title,
  readOnly = false,
  createFolder,
  onAdd,
  onEdit,
  addLabel,
}: {
  value: T[];
  onChange: (nodes: T[]) => void;
  title: (node: T) => string;
  readOnly?: boolean;
  createFolder: (name: string) => T;
  onAdd: (parent: string) => void;
  onEdit: (node: T) => void;
  addLabel: string;
}) {
  const [selected, setSelected] = useState("");
  const [expanded, setExpanded] = useState<React.Key[]>(() =>
    folderKeys(value),
  );
  const [folder, setFolder] = useState<{
    key?: string;
    parent: string;
    name: string;
  }>();
  const current = findNode(value, selected);
  const parent =
    current?.kind === "folder" ? selected : parentKey(value, selected);
  const siblings = parentKey(value, selected)
    ? findNode(value, parentKey(value, selected))?.children || []
    : value;
  const index = siblings.findIndex((v) => nodeKey(v) === selected);
  let sequence = 0;
  function data(nodes: T[]): DataNode[] {
    return nodes.map((node) => ({
      key: nodeKey(node),
      isLeaf: node.kind !== "folder",
      icon: node.kind === "folder" ? <FolderOutlined /> : <FileOutlined />,
      title: (
        <Space size={8}>
          {node.kind !== "folder" && <Tag>{++sequence}</Tag>}
          <span>{title(node)}</span>
          {node.kind === "folder" && (
            <Typography.Text type="secondary">
              {leafNodes(node.children || []).length} 项
            </Typography.Text>
          )}
        </Space>
      ),
      children: node.kind === "folder" ? data(node.children || []) : undefined,
    }));
  }
  function expand(key: string) {
    if (key) setExpanded((keys) => [...new Set([...keys, key])]);
  }
  return (
    <div className="orchestration">
      <Typography.Paragraph type="secondary">
        按树中从上到下的叶子顺序执行，目录仅用于分组。
        {!readOnly && "可拖动节点排序或移入目录，测试项不能包含子节点。"}
      </Typography.Paragraph>
      {!readOnly && (
        <div className="orchestration-toolbar">
          <Space wrap>
            <Button
              onClick={() => {
                expand(parent);
                onAdd(parent);
              }}
            >
              {addLabel}
            </Button>
            <Button onClick={() => setFolder({ parent, name: "" })}>
              新建目录
            </Button>
            <Button onClick={() => setSelected("")}>根目录</Button>
          </Space>
          <Typography.Text type="secondary">
            添加位置：{parent ? title(findNode(value, parent)!) : "根目录"}
          </Typography.Text>
        </div>
      )}
      <div className="orchestration-body">
        {value.length ? (
          <Tree
            blockNode
            showIcon
            showLine={{ showLeafIcon: false }}
            treeData={data(value)}
            selectedKeys={current ? [selected] : []}
            expandedKeys={expanded}
            onExpand={(keys) => setExpanded(keys)}
            onSelect={(keys) => {
              const key = String(keys[0] || "");
              setSelected(key);
              if (findNode(value, key)?.kind === "folder") expand(key);
            }}
            onDoubleClick={(_, treeNode) => {
              const node = findNode(value, String(treeNode.key));
              if (node && node.kind !== "folder") onEdit(node);
            }}
            draggable={!readOnly}
            allowDrop={({ dragNode, dropNode, dropPosition }) => {
              const source = findNode(value, String(dragNode.key));
              return (
                !!source &&
                dragNode.key !== dropNode.key &&
                !findNode(source.children || [], String(dropNode.key)) &&
                (dropPosition !== 0 ||
                  findNode(value, String(dropNode.key))?.kind === "folder")
              );
            }}
            onDragEnter={({ expandedKeys }) => setExpanded(expandedKeys)}
            onDrop={(info) => {
              const target = String(info.node.key);
              const offset =
                info.dropPosition - Number(info.node.pos.split("-").at(-1));
              const position = !info.dropToGap
                ? "inside"
                : offset < 0
                  ? "before"
                  : "after";
              onChange(
                moveNode(value, String(info.dragNode.key), target, position),
              );
              expand(target);
            }}
          />
        ) : (
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description="暂无测试项，添加测试项或创建目录开始编排"
          />
        )}
      </div>
      {current && (
        <div className="orchestration-toolbar">
          <Typography.Text ellipsis style={{ maxWidth: 360 }}>
            {title(current)}
          </Typography.Text>
          <Space wrap>
            {current.kind !== "folder" && (
              <Button disabled={false} onClick={() => onEdit(current)}>
                {readOnly ? "查看" : "编辑"}
              </Button>
            )}
            {!readOnly && (
              <>
                {current.kind === "folder" && (
                  <Button
                    onClick={() =>
                      setFolder({
                        key: selected,
                        parent: "",
                        name: current.name || "",
                      })
                    }
                  >
                    重命名
                  </Button>
                )}
                <Button
                  disabled={index <= 0}
                  onClick={() =>
                    onChange(
                      moveNode(
                        value,
                        selected,
                        nodeKey(siblings[index - 1]),
                        "before",
                      ),
                    )
                  }
                >
                  上移
                </Button>
                <Button
                  disabled={index < 0 || index >= siblings.length - 1}
                  onClick={() =>
                    onChange(
                      moveNode(
                        value,
                        selected,
                        nodeKey(siblings[index + 1]),
                        "after",
                      ),
                    )
                  }
                >
                  下移
                </Button>
                {parentKey(value, selected) && (
                  <Button
                    onClick={() => {
                      onChange(
                        insertNode(removeNode(value, selected), "", current),
                      );
                    }}
                  >
                    移到根目录
                  </Button>
                )}
                <Popconfirm
                  title={
                    current.kind === "folder"
                      ? "删除目录及其中的全部测试项？"
                      : "移除此测试项？"
                  }
                  onConfirm={() => {
                    onChange(removeNode(value, selected));
                    setSelected("");
                  }}
                >
                  <Button danger>删除</Button>
                </Popconfirm>
              </>
            )}
          </Space>
        </div>
      )}
      <Typography.Text type="secondary">
        共 {leafNodes(value).length} 个测试项
      </Typography.Text>
      <Modal
        title={folder?.key ? "重命名目录" : "新建目录"}
        open={!!folder}
        onCancel={() => setFolder(undefined)}
        okButtonProps={{ disabled: !folder?.name.trim() }}
        onOk={() => {
          if (!folder?.name.trim()) return;
          if (folder.key)
            onChange(
              updateNode(value, folder.key, (node) => ({
                ...node,
                name: folder.name.trim(),
              })),
            );
          else {
            const node = createFolder(folder.name.trim());
            onChange(insertNode(value, folder.parent, node));
            expand(folder.parent);
            setSelected(nodeKey(node));
            expand(nodeKey(node));
          }
          setFolder(undefined);
        }}
      >
        <Input
          autoFocus
          aria-label="目录名称"
          placeholder="目录名称"
          maxLength={128}
          value={folder?.name}
          onChange={(e) =>
            setFolder((v) => v && { ...v, name: e.target.value })
          }
        />
      </Modal>
    </div>
  );
}

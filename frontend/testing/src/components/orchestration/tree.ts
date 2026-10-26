export interface ArrangementNode<T> {
  id?: string;
  node_id?: string;
  kind?: string;
  name?: string;
  children?: T[];
}

export const nodeKey = <T extends ArrangementNode<T>>(node: T) =>
  node.node_id || node.id || "";

export function folderKeys<T extends ArrangementNode<T>>(nodes: T[]): string[] {
  return nodes.flatMap((node) =>
    node.kind === "folder"
      ? [nodeKey(node), ...folderKeys(node.children || [])]
      : [],
  );
}

export function findNode<T extends ArrangementNode<T>>(
  nodes: T[],
  key: string,
): T | undefined {
  for (const node of nodes) {
    if (nodeKey(node) === key) return node;
    const found = findNode(node.children || [], key);
    if (found) return found;
  }
}

export function parentKey<T extends ArrangementNode<T>>(
  nodes: T[],
  key: string,
  parent = "",
): string {
  for (const node of nodes) {
    if (nodeKey(node) === key) return parent;
    if (findNode(node.children || [], key))
      return parentKey(node.children || [], key, nodeKey(node));
  }
  return "";
}

export function updateNode<T extends ArrangementNode<T>>(
  nodes: T[],
  key: string,
  update: (node: T) => T,
): T[] {
  return nodes.map((node) =>
    nodeKey(node) === key
      ? update(node)
      : {
          ...node,
          ...(node.children
            ? { children: updateNode(node.children, key, update) }
            : {}),
        },
  );
}

export function removeNode<T extends ArrangementNode<T>>(
  nodes: T[],
  key: string,
): T[] {
  return nodes
    .filter((node) => nodeKey(node) !== key)
    .map((node) => ({
      ...node,
      ...(node.children ? { children: removeNode(node.children, key) } : {}),
    }));
}

export function insertNode<T extends ArrangementNode<T>>(
  nodes: T[],
  parent: string,
  node: T,
  index?: number,
): T[] {
  if (parent)
    return updateNode(nodes, parent, (folder) => ({
      ...folder,
      children: insertNode(folder.children || [], "", node, index),
    }));
  const next = [...nodes];
  next.splice(index ?? next.length, 0, node);
  return next;
}

export function moveNode<T extends ArrangementNode<T>>(
  nodes: T[],
  source: string,
  target: string,
  position: "inside" | "before" | "after",
): T[] {
  const node = findNode(nodes, source);
  const destination = findNode(nodes, target);
  if (
    !node ||
    !destination ||
    source === target ||
    findNode(node.children || [], target)
  )
    return nodes;
  if (position === "inside" && destination.kind !== "folder") return nodes;
  const rest = removeNode(nodes, source);
  if (position === "inside") return insertNode(rest, target, node);
  const parent = parentKey(rest, target);
  const siblings = parent ? findNode(rest, parent)!.children || [] : rest;
  const index = siblings.findIndex((v) => nodeKey(v) === target);
  return insertNode(rest, parent, node, index + (position === "after" ? 1 : 0));
}

export function leafNodes<T extends ArrangementNode<T>>(nodes: T[]): T[] {
  return nodes.flatMap((node) =>
    node.kind === "folder" ? leafNodes(node.children || []) : [node],
  );
}

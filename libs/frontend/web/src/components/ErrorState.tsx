import { Alert, message } from "antd";

export function ErrorState({ error }: { error: unknown }) {
  return error ? (
    <Alert
      type="error"
      showIcon
      message={error instanceof Error ? error.message : "加载失败"}
    />
  ) : null;
}

import { Tag } from "antd";

export function Status({ value }: { value: string }) {
  const names: Record<string, string> = {
    created: "待投递",
    queued: "排队中",
    running: "执行中",
    passed: "通过",
    failed: "失败",
    unverified: "未校验",
    error: "执行异常",
    canceled: "已取消",
    timed_out: "超时",
    skipped: "跳过",
  };
  const colors: Record<string, string> = {
    passed: "success",
    failed: "error",
    error: "error",
    running: "processing",
    unverified: "warning",
    timed_out: "warning",
  };
  return <Tag color={colors[value] || "default"}>{names[value] || value}</Tag>;
}

import { useState } from "react";
import { message } from "antd";

import "./style.css";

import { client } from "./query";
export function useAction() {
  const [busy, setBusy] = useState(false);
  async function run<T>(action: () => Promise<T>, success = "已保存") {
    setBusy(true);
    try {
      const value = await action();
      await client.invalidateQueries();
      if (success) message.success(success);
      return value;
    } catch (error) {
      message.error(error instanceof Error ? error.message : "操作失败");
      return undefined;
    } finally {
      setBusy(false);
    }
  }
  return { busy, run };
}

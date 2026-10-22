import React from "react";
import { ConfigProvider } from "antd";
import { QueryClientProvider } from "@tanstack/react-query";
import "./style.css";
import zhCN from "antd/locale/zh_CN";
import { client } from "./query";
export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <ConfigProvider locale={zhCN}>
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    </ConfigProvider>
  );
}

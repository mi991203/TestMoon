import { identity } from "@testmoon/api";
import React from "react";
import { Button, Layout, Select, Space, Tag, Typography } from "antd";

import { client } from "../query";
import type { IdentitySessionResponse } from "@testmoon/api";
export function Shell({
  session,
  org,
  onOrg,
  menu,
  selected,
  onSelect,
  children,
}: {
  session: IdentitySessionResponse;
  org: string;
  onOrg: (id: string) => void;
  menu: { key: string; label: React.ReactNode }[];
  selected: string;
  onSelect: (key: string) => void;
  children: React.ReactNode;
}) {
  return (
    <Layout className="shell">
      <Layout.Sider width={224} className="sidebar">
        <a className="brand" href="/">
          ◒{" "}
          <span>
            TestMoon<small>开发与测试工作台</small>
          </span>
        </a>
        <div className="side-nav">
          {menu.map((item) => (
            <button
              key={item.key}
              className={selected === item.key ? "selected" : ""}
              onClick={() => onSelect(item.key)}
            >
              {item.label}
            </button>
          ))}
        </div>
        <div className="side-foot">LOW CODE · CONNECT · VERIFY</div>
      </Layout.Sider>
      <Layout>
        <Layout.Header className="topbar">
          <Space>
            <Typography.Text type="secondary">组织空间</Typography.Text>
            <Select
              aria-label="当前组织"
              style={{ width: 220 }}
              value={org || undefined}
              placeholder="选择组织"
              options={(session.organizations || []).map((o) => ({
                value: o.id,
                label: o.name,
              }))}
              onChange={onOrg}
            />
          </Space>
          <Space>
            <Tag>{session.principal?.username}</Tag>
            <a href="/">统一门户</a>
            <Button
              onClick={async () => {
                await identity.logout({});
                client.clear();
                location.href = "/login";
              }}
            >
              退出
            </Button>
          </Space>
        </Layout.Header>
        <Layout.Content className="content">{children}</Layout.Content>
      </Layout>
    </Layout>
  );
}

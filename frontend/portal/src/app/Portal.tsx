import { Login } from "../features/Login";

import { useState } from "react";

import { Alert, Button, Card, Space, Spin, Tag, Typography } from "antd";
import {
  client,
  chooseOrg,
  ErrorState,
  PageHeading,
  setOrgPreference,
  Shell,
  useSession,
} from "@testmoon/web";

import { Organizations } from "../features/Organizations";
import { Users } from "../features/Users";
import { Security } from "../features/Security";
import { PublicRunners } from "../features/PublicRunners";
export function Portal() {
  const session = useSession();
  const [org, setOrg] = useState("");
  const [page, setPage] = useState("overview");
  if (session.isPending) return <Spin fullscreen />;
  if (!session.data) {
    if ((session.error as any)?.status !== 401)
      return (
        <div style={{ padding: 40 }}>
          <ErrorState error={session.error} />
          <Button onClick={() => session.refetch()}>重试</Button>
        </div>
      );
    return <Login />;
  }
  const selected = org || chooseOrg(session.data);
  const isAdmin = !!session.data.principal.admin;
  const menus = [
    { key: "overview", label: "工作台" },
    { key: "organizations", label: "组织与成员" },
    ...(isAdmin
      ? [
          { key: "users", label: "账号管理" },
          { key: "public-runners", label: "公共 Runner" },
        ]
      : []),
    { key: "security", label: "账号安全" },
  ];
  return (
    <Shell
      session={session.data}
      org={selected}
      onOrg={(id) => {
        setOrgPreference(id);
        setOrg(id);
        client.clear();
      }}
      menu={menus}
      selected={page}
      onSelect={setPage}
    >
      {page === "overview" ? (
        <>
          <PageHeading
            title="欢迎回到 TestMoon"
            description="统一账号与组织，连接你的开发测试应用。"
          />
          <Card title="测试平台" style={{ maxWidth: 700 }}>
            <Typography.Paragraph>
              管理 HTTP 接口、编排顺序场景、运行测试任务，并使用 Mock
              与共享接口文档。
            </Typography.Paragraph>
            <Space>
              <Tag>HTTP</Tag>
              <Tag>JavaScript</Tag>
              <Tag>公共 / 私有 Runner</Tag>
            </Space>
            <div style={{ marginTop: 24 }}>
              <Button
                type="primary"
                disabled={!selected}
                href={`/apps/testing/?org=${encodeURIComponent(selected)}`}
              >
                进入测试平台 →
              </Button>
            </div>
            {!selected && (
              <Alert
                style={{ marginTop: 20 }}
                type="info"
                message="当前账号尚未加入组织，请先开通组织成员关系。"
              />
            )}
          </Card>
        </>
      ) : page === "users" ? (
        <Users />
      ) : page === "organizations" ? (
        <Organizations session={session.data} org={selected} />
      ) : page === "security" ? (
        <Security />
      ) : (
        <PublicRunners />
      )}
    </Shell>
  );
}

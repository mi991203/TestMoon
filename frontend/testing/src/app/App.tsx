import { lazy, Suspense } from "react";
import { core } from "@testmoon/api";
import { useEffect } from "react";
import { useQuery } from "@tanstack/react-query";
import { Button, Card, Select, Space, Spin, Tag } from "antd";
import {
  chooseOrg,
  EmptyProject,
  ErrorState,
  Scope,
  setOrgPreference,
  Shell,
  useSession,
} from "@testmoon/web";
import {
  Navigate,
  Route,
  Routes,
  useLocation,
  useNavigate,
} from "react-router-dom";
import { useScopedNavigate } from "../hooks/navigation";

const AssetPanel = lazy(() =>
  import("../components/resources").then((module) => ({
    default: module.AssetPanel,
  })),
);
const ApiTesting = lazy(() =>
  import("../features/api-testing").then((module) => ({
    default: module.ApiTesting,
  })),
);

const Documents = lazy(() =>
  import("../features/documents/DocumentsPage").then((module) => ({
    default: module.Documents,
  })),
);
const Members = lazy(() =>
  import("../features/members/MembersPage").then((module) => ({
    default: module.Members,
  })),
);
const Mocks = lazy(() =>
  import("../features/mocks/MocksPage").then((module) => ({
    default: module.Mocks,
  })),
);
const Runners = lazy(() =>
  import("../features/runners/RunnersPage").then((module) => ({
    default: module.Runners,
  })),
);
const Tasks = lazy(() =>
  import("../features/tasks/TasksPage").then((module) => ({
    default: module.Tasks,
  })),
);
const Reports = lazy(() =>
  import("../features/reports/ReportsPage").then((m) => ({
    default: m.Reports,
  })),
);

const Projects = lazy(() =>
  import("../features/projects/ProjectsPage").then((module) => ({
    default: module.Projects,
  })),
);
const ReportRoute = lazy(() =>
  import("./ReportRoute").then((module) => ({ default: module.ReportRoute })),
);
const DocumentRoute = lazy(() =>
  import("../features/documents/DocumentDetail").then((module) => ({
    default: module.DocumentRoute,
  })),
);
export function App() {
  const session = useSession();
  const location = useLocation();
  const navigate = useNavigate();
  const go = useScopedNavigate();
  const params = new URLSearchParams(location.search);
  const selected =
    params.get("org") || (session.data ? chooseOrg(session.data) : "");
  const project = params.get("project") || "";
  const scope: Scope = { org_id: selected, project_id: project };
  const projects = useQuery({
    queryKey: ["projects", selected],
    enabled: !!selected && !!session.data,
    queryFn: () => core.listProjects({ ...{ org_id: selected } }),
  });
  const access = useQuery({
    queryKey: ["access", selected, project],
    enabled: !!selected && !!project && !!session.data,
    queryFn: () => core.authorize({ ...scope, ...{ permission: "read" } }),
  });
  useEffect(() => {
    if (!selected) return;
    const next = new URLSearchParams(location.search);
    let changed = false;
    if (!next.get("org")) {
      next.set("org", selected);
      changed = true;
    }
    if (!project && projects.data?.items?.length) {
      next.set("project", projects.data.items[0].id);
      changed = true;
    }
    if (changed)
      navigate(
        { pathname: location.pathname, search: next.toString() },
        { replace: true },
      );
  }, [
    selected,
    project,
    projects.data,
    location.pathname,
    location.search,
    navigate,
  ]);
  function selectScope(org: string, id: string, path: string) {
    const next = new URLSearchParams({ org });
    if (id) next.set("project", id);
    navigate({ pathname: path, search: next.toString() });
  }
  if (session.isPending) return <Spin fullscreen />;
  if (!session.data) {
    if ((session.error as any)?.status === 401) {
      window.location.replace("/login");
      return null;
    }
    return (
      <Card>
        <ErrorState error={session.error} />
        <Button onClick={() => session.refetch()}>重试</Button>
      </Card>
    );
  }
  const menus = [
    { key: "api-testing", label: "API测试" },
    { key: "scenario", label: "场景用例" },
    { key: "tasks", label: "测试任务" },
    { key: "reports", label: "执行报告" },
    { key: "mock", label: "HTTP Mock" },
    { key: "documents", label: "接口文档" },
    { key: "environment", label: "环境管理" },
    { key: "runners", label: "Runner 管理" },
    { key: "members", label: "项目成员" },
    { key: "projects", label: "项目管理" },
  ];
  const page = location.pathname.split("/")[1] || "api-testing";
  const role = access.data?.role || "";
  const onRun = (id: string) => go(`/reports/${id}`);
  return (
    <Shell
      session={session.data}
      org={selected}
      onOrg={(id) => {
        setOrgPreference(id);
        selectScope(id, "", "/api-testing");
      }}
      menu={menus}
      selected={page}
      onSelect={(key) => go(`/${key}`)}
      scopeControls={
        <Space>
          <span>当前项目</span>
          <Select
            aria-label="当前项目"
            style={{ width: 260 }}
            value={project || undefined}
            placeholder="选择项目"
            options={(projects.data?.items || []).map((v) => ({
              value: v.id,
              label: v.name,
            }))}
            onChange={(id) =>
              selectScope(
                selected,
                id,
                `/${menus.some((v) => v.key === page) ? page : "api-testing"}`,
              )
            }
          />
          {role && (
            <Tag>
              {
                (
                  {
                    admin: "管理员",
                    editor: "编辑者",
                    viewer: "只读成员",
                  } as Record<string, string>
                )[role]
              }
            </Tag>
          )}
        </Space>
      }
    >
      <ErrorState error={projects.error} />
      {page === "projects" ? (
        <Projects
          scope={scope}
          role={role}
          orgAdmin={
            session.data.principal.organizations?.[selected] === "admin"
          }
          onSelect={(id) => selectScope(selected, id, "/api-testing")}
        />
      ) : !project ? (
        <EmptyProject />
      ) : access.isPending ? (
        <Spin />
      ) : access.error ? (
        <ErrorState error={access.error} />
      ) : (
        <Suspense fallback={<Spin />}>
          <Routes key={`${selected}/${project}`}>
            <Route
              index
              element={
                <Navigate
                  to={{ pathname: "/api-testing", search: location.search }}
                  replace
                />
              }
            />
            <Route
              path="api-testing/*"
              element={<ApiTesting scope={scope} role={role} />}
            />
            <Route
              path="api"
              element={
                <Navigate
                  to={{ pathname: "/api-testing", search: location.search }}
                  replace
                />
              }
            />
            <Route
              path="case"
              element={
                <Navigate
                  to={{
                    pathname: "/api-testing/unlinked",
                    search: location.search,
                  }}
                  replace
                />
              }
            />
            <Route
              path="scenario"
              element={<AssetPanel kind="scenario" scope={scope} role={role} />}
            />
            <Route
              path="scenario/:id"
              element={<AssetPanel kind="scenario" scope={scope} role={role} />}
            />
            <Route
              path="environment"
              element={
                <AssetPanel kind="environment" scope={scope} role={role} />
              }
            />
            <Route
              path="environment/:id"
              element={
                <AssetPanel kind="environment" scope={scope} role={role} />
              }
            />
            <Route
              path="tasks"
              element={<Tasks scope={scope} role={role} onRun={onRun} />}
            />
            <Route
              path="reports"
              element={<Reports scope={scope} role={role} onOpen={onRun} />}
            />
            <Route path="reports/:id" element={<ReportRoute scope={scope} />} />
            <Route
              path="runners"
              element={<Runners scope={scope} role={role} />}
            />
            <Route
              path="members"
              element={<Members scope={scope} role={role} />}
            />
            <Route
              path="imports"
              element={
                <Navigate
                  to={{ pathname: "/api-testing", search: location.search }}
                  replace
                  state={{ openImport: true }}
                />
              }
            />
            <Route
              path="documents"
              element={
                <Documents
                  scope={scope}
                  role={role}
                  onOpen={(id) => go(`/documents/${id}`)}
                />
              }
            />
            <Route
              path="documents/:id"
              element={<DocumentRoute scope={scope} />}
            />
            <Route path="mock" element={<Mocks scope={scope} role={role} />} />
            <Route
              path="tasks/:id"
              element={<Tasks scope={scope} role={role} onRun={onRun} />}
            />
            <Route
              path="mock/:id"
              element={<Mocks scope={scope} role={role} />}
            />
            <Route
              path="*"
              element={
                <Card>
                  <ErrorState error={new Error("页面不存在")} />
                  <Button onClick={() => go("/api-testing")}>
                    返回API测试
                  </Button>
                </Card>
              }
            />
          </Routes>
        </Suspense>
      )}
    </Shell>
  );
}

import { Breadcrumb, Tabs } from "antd";

import { useScopedNavigate } from "../../hooks/navigation";

export function ApiHeader({
  id,
  name,
  tab,
}: {
  id?: string;
  name?: string;
  tab: string;
}) {
  const go = useScopedNavigate();
  return (
    <>
      <Breadcrumb
        className="detail-breadcrumb"
        items={[
          { title: <a onClick={() => go("/api-testing")}>API测试</a> },
          { title: name || "未关联接口用例" },
        ]}
      />
      <Tabs
        activeKey={tab}
        onChange={(key) => {
          if (key === "debug" && id) go(`/api-testing/apis/${id}/debug`);
          else
            go(id ? `/api-testing/apis/${id}/cases` : "/api-testing/unlinked");
        }}
        items={[
          { key: "debug", label: "API调试", disabled: !id },
          { key: "cases", label: "接口用例" },
        ]}
      />
    </>
  );
}

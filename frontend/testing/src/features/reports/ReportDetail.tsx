import { Button } from "antd";
import { PageHeading, type Scope } from "@testmoon/web";
import { RunResult } from "../../components/resources/RunResult";

export function ReportDetail({ scope, id, onBack }: {
  scope: Scope;
  id: string;
  onBack: () => void;
}) {
  return <>
    <PageHeading title="执行详情" extra={<Button onClick={onBack}>返回报告列表</Button>} />
    <RunResult scope={scope} id={id} />
  </>;
}

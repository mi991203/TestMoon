import { Scope } from "@testmoon/web";

import { useParams } from "react-router-dom";
import { useScopedNavigate } from "../hooks/navigation";

import { ReportDetail } from "../features/reports";

export function ReportRoute({ scope }: { scope: Scope }) {
  const { id } = useParams();
  const go = useScopedNavigate();
  return (
    <ReportDetail scope={scope} id={id || ""} onBack={() => go("/reports")} />
  );
}

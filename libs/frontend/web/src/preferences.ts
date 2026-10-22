import type { IdentitySessionResponse } from "@testmoon/api";
export function chooseOrg(session: IdentitySessionResponse) {
  const search = new URLSearchParams(location.search);
  const preferred =
    search.get("org") || localStorage.getItem("testmoon.org") || "";
  return (session.organizations || []).some((o) => o.id === preferred)
    ? preferred
    : session.organizations?.[0]?.id || "";
}
export function setOrgPreference(id: string) {
  localStorage.setItem("testmoon.org", id);
}

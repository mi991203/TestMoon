import { Scope } from "@testmoon/web";

export type Context = { scope: Scope; role: string };
export const editable = (role: string) => role === "admin" || role === "editor";

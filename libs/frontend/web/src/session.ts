import { identity, setCSRF } from "@testmoon/api";

import { useQuery } from "@tanstack/react-query";
import "./style.css";

import { client } from "./query";
export function useSession(enabled = true) {
  return useQuery({
    queryKey: ["session"],
    enabled,
    queryFn: async () => {
      const value = await identity.session({});
      setCSRF(value.csrf);
      return value;
    },
  });
}
export async function login(username: string, password: string) {
  const value = await identity.login({ ...{ username, password } });
  setCSRF(value.csrf);
  await client.invalidateQueries({ queryKey: ["session"] });
  return value;
}

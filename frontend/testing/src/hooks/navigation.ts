import { useLocation, useNavigate } from "react-router-dom";

export function useScopedNavigate() {
  const navigate = useNavigate();
  const location = useLocation();
  return (path: string, replace = false, state?: unknown) =>
    navigate({ pathname: path, search: location.search }, { replace, state });
}

let csrf = "";
export function setCSRF(token: string) {
  csrf = token;
}
export class APIError extends Error {
  constructor(
    public readonly status: number,
    message: string,
  ) {
    super(message);
  }
}
/** Same-origin gateway transport. Business operations live in generated clients. */
export async function request<T>(path: string, body: unknown): Promise<T> {
  const response = await fetch(path, {
    method: "POST",
    credentials: "same-origin",
    headers: { "Content-Type": "application/json", "X-CSRF-Token": csrf },
    body: JSON.stringify(body),
  });
  const raw = await response.text();
  let value: unknown;
  try {
    value = JSON.parse(raw);
  } catch {
    if (!response.ok)
      throw new APIError(response.status, raw.slice(0, 2048) || "请求失败");
    throw new APIError(response.status, "服务返回了无效的 JSON");
  }
  if (!response.ok) {
    const message =
      typeof value === "object" && value !== null && "message" in value
        ? String(value.message)
        : "请求失败";
    throw new APIError(response.status, message.slice(0, 2048));
  }
  return value as T;
}

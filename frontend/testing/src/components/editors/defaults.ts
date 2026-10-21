export const blankRequest = () => ({
  method: "GET",
  url: "",
  headers: {},
  query: {},
  body_type: "json",
  body: "",
  auth: { type: "none" },
  timeout_ms: 30000,
});
export const sources = [
  { value: "status", label: "状态码" },
  { value: "json", label: "JSON字段（JSON Pointer）" },
  { value: "header", label: "响应头" },
  { value: "text", label: "响应文本" },
  { value: "duration", label: "耗时(ms)" },
];

/** Request overrides support a defined subset of HTTP request fields. */
export function requestPatch(
  request: import("../../models/editors").RequestDraft,
) {
  const { timeout_ms: _, ...patch } = request;
  return patch;
}

/** Protobuf JSON emits empty maps/null messages even for an unset override. */
export function hasRequestPatch(
  patch?: import("../../models/editors").PatchDraft | null,
): boolean {
  return (
    !!patch &&
    Object.values(patch).some(
      (value) =>
        value !== undefined &&
        value !== null &&
        (typeof value !== "object" || Object.keys(value).length > 0),
    )
  );
}

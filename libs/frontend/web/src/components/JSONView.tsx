export function JSONView({ value }: { value: unknown }) {
  return (
    <pre className="json-view">
      {typeof value === "string" ? value : JSON.stringify(value, null, 2)}
    </pre>
  );
}

export function TagIndex({ index }: { index: number }) {
  return (
    <span
      style={{
        background: "#e7effe",
        padding: "4px 9px",
        borderRadius: 6,
        color: "#315996",
      }}
    >
      {String(index + 1).padStart(2, "0")}
    </span>
  );
}

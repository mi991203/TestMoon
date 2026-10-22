import React from "react";
import { Spin } from "antd";

const LazyEditor = React.lazy(() => import("../editor"));
export function Code({
  value,
  onChange,
  language = "json",
  height = 250,
  readOnly = false,
}: {
  value: string;
  onChange?: (v: string) => void;
  language?: string;
  height?: number;
  readOnly?: boolean;
}) {
  return (
    <div className="code">
      <React.Suspense fallback={<Spin />}>
        <LazyEditor
          value={value}
          onChange={onChange}
          language={language}
          height={height}
          readOnly={readOnly}
        />
      </React.Suspense>
    </div>
  );
}

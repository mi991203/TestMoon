import Editor, { loader } from "@monaco-editor/react";
import * as monaco from "monaco-editor";
import EditorWorker from "monaco-editor/editor/editor.worker.js?worker";
import JSONWorker from "monaco-editor/language/json/json.worker.js?worker";
import TSWorker from "monaco-editor/language/typescript/ts.worker.js?worker";
(self as any).MonacoEnvironment = {
  getWorker(_id: string, label: string) {
    if (label === "json") return new JSONWorker();
    if (label === "typescript" || label === "javascript") return new TSWorker();
    return new EditorWorker();
  },
};
loader.config({ monaco });
export default function CodeEditor({
  value,
  onChange,
  language,
  height,
  readOnly,
}: {
  value: string;
  onChange?: (value: string) => void;
  language?: string;
  height?: number;
  readOnly?: boolean;
}) {
  return (
    <Editor
      height={height}
      language={language}
      value={value}
      onChange={(v) => onChange?.(v || "")}
      options={{
        minimap: { enabled: false },
        readOnly,
        scrollBeyondLastLine: false,
        fontSize: 13,
        tabSize: 2,
        automaticLayout: true,
        wordWrap: "on",
      }}
    />
  );
}

"use client";

import { Card } from "@/components/ui/card";

export type ToolCallView = {
  id: string;
  name: string;
  args: unknown;
  status: "running" | "success" | "error";
  display?: unknown;
  error?: string;
};

export function AiToolCallCard({ call }: { call: ToolCallView }) {
  const statusColor =
    call.status === "error"
      ? "text-red-600"
      : call.status === "running"
        ? "text-muted-foreground"
        : "text-green-600";

  return (
    <Card className="my-2 border-dashed p-3 text-sm">
      <div className="flex items-center gap-2 font-mono">
        <span aria-hidden>🔧</span>
        <span className="font-semibold">{call.name}</span>
        <span className={`ml-auto text-xs ${statusColor}`}>{call.status}</span>
      </div>
      <pre className="mt-1 overflow-x-auto text-xs text-muted-foreground">
        {JSON.stringify(call.args, null, 2)}
      </pre>
      {call.error && (
        <div className="mt-2 text-xs text-red-600">error: {call.error}</div>
      )}
      {call.display !== undefined && call.status === "success" && (
        <pre className="mt-2 max-h-48 overflow-auto rounded bg-muted p-2 text-xs">
          {JSON.stringify(call.display, null, 2)}
        </pre>
      )}
    </Card>
  );
}

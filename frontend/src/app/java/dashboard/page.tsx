import { Suspense } from "react";
import { DashboardClient } from "./dashboard-client";

export default function DashboardPage() {
  return (
    <Suspense fallback={<div className="p-6 text-sm text-muted-foreground">Loading…</div>}>
      <DashboardClient />
    </Suspense>
  );
}

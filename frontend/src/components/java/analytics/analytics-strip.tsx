"use client";

import Link from "next/link";
import { useQuery } from "@apollo/client/react";
import { PROJECT_HEALTH, type ProjectHealthData } from "./project-health.graphql";
import { StatCard } from "./stat-card";

interface Props { projectId: string; }

export function AnalyticsStrip({ projectId }: Props) {
  const { data, loading, error } = useQuery<ProjectHealthData>(PROJECT_HEALTH, {
    variables: { projectId },
  });

  if (loading || error || !data) return null;

  const { stats, velocity, activity } = data.projectHealth;
  const completedThisWeek = velocity.weeklyThroughput[0]?.completed ?? 0;

  return (
    <div
      data-testid="analytics-strip"
      className="mb-6 flex flex-col gap-4 rounded-lg border bg-muted/30 p-4 md:flex-row md:items-center md:justify-between"
    >
      <div className="grid flex-1 grid-cols-2 gap-3 md:grid-cols-4">
        <StatCard label="Overdue" value={stats.overdueCount} />
        <StatCard label="Completed This Week" value={completedThisWeek} />
        <StatCard label="Avg Lead Time" value={velocity.avgLeadTimeHours} unit="h" />
        <StatCard label="Active Contributors" value={activity.activeContributors} />
      </div>
      <Link
        href={`/java/dashboard?projectId=${projectId}`}
        className="text-sm font-medium underline hover:text-foreground"
      >
        View full dashboard →
      </Link>
    </div>
  );
}

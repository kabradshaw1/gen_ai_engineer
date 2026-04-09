"use client";

import { useEffect } from "react";
import { useQuery } from "@apollo/client/react";
import { gql } from "@apollo/client";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { PROJECT_HEALTH, type ProjectHealthData } from "@/components/java/analytics/project-health.graphql";
import { ProjectSelector } from "@/components/java/analytics/project-selector";
import { StatCard } from "@/components/java/analytics/stat-card";
import { TaskBreakdownCharts } from "@/components/java/analytics/task-breakdown-charts";
import { VelocitySection } from "@/components/java/analytics/velocity-section";
import { MemberWorkloadTable } from "@/components/java/analytics/member-workload-table";
import { ActivitySection } from "@/components/java/analytics/activity-section";

const MY_PROJECTS = gql`
  query MyProjects { myProjects { id name } }
`;

interface Project { id: string; name: string; }
interface MyProjectsData { myProjects: Project[]; }

export function DashboardClient() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const projectId = searchParams.get("projectId");

  const { data: projectsData, loading: projectsLoading, error: projectsError } =
    useQuery<MyProjectsData>(MY_PROJECTS);

  const projects = projectsData?.myProjects ?? [];

  useEffect(() => {
    if (!projectId && projects.length > 0) {
      router.replace(`/java/dashboard?projectId=${projects[0].id}`);
    }
  }, [projectId, projects, router]);

  const { data, loading, error } = useQuery<ProjectHealthData>(PROJECT_HEALTH, {
    variables: { projectId: projectId ?? "" },
    skip: !projectId,
  });

  if (projectsLoading) {
    return <div className="p-6 text-sm text-muted-foreground">Loading projects…</div>;
  }
  if (projectsError) {
    return <ErrorCard message="Failed to load projects." />;
  }
  if (projects.length === 0) {
    return (
      <div className="mx-auto max-w-2xl p-6">
        <h1 className="text-2xl font-semibold">Project Analytics</h1>
        <p className="mt-4 text-muted-foreground">
          You don&rsquo;t have any projects yet.{" "}
          <Link className="underline" href="/java/tasks">Create one to get started →</Link>
        </p>
      </div>
    );
  }

  const currentId = projectId ?? projects[0].id;

  return (
    <div className="mx-auto max-w-6xl p-6">
      <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <h1 className="text-2xl font-semibold">Project Analytics</h1>
        <ProjectSelector
          projects={projects}
          currentId={currentId}
          onChange={(id) => router.replace(`/java/dashboard?projectId=${id}`)}
        />
      </div>

      {loading && <DashboardSkeleton />}
      {error && <ErrorCard message="Failed to load analytics." />}
      {data && (
        <div className="mt-6 space-y-6">
          <div className="grid gap-3 sm:grid-cols-2 md:grid-cols-4">
            <StatCard label="Overdue" value={data.projectHealth.stats.overdueCount} />
            <StatCard label="Avg Completion" value={data.projectHealth.stats.avgCompletionTimeHours} unit="h" />
            <StatCard label="Active Contributors" value={data.projectHealth.activity.activeContributors} />
            <StatCard label="Total Events" value={data.projectHealth.activity.totalEvents} />
          </div>
          <TaskBreakdownCharts stats={data.projectHealth.stats} />
          <VelocitySection velocity={data.projectHealth.velocity} />
          <MemberWorkloadTable members={data.projectHealth.stats.memberWorkload} />
          <ActivitySection activity={data.projectHealth.activity} />
        </div>
      )}
    </div>
  );
}

function DashboardSkeleton() {
  return (
    <div className="mt-6 grid gap-3 sm:grid-cols-2 md:grid-cols-4">
      {Array.from({ length: 4 }).map((_, i) => (
        <div key={i} className="h-24 animate-pulse rounded-lg border bg-muted/40" />
      ))}
    </div>
  );
}

function ErrorCard({ message }: { message: string }) {
  return (
    <div className="mt-6 rounded-lg border border-destructive/40 bg-destructive/10 p-4 text-sm text-destructive">
      {message}
    </div>
  );
}

"use client";

import { Bar, BarChart, CartesianGrid, Line, LineChart, XAxis, YAxis } from "recharts";
import { ChartContainer, ChartTooltip, ChartTooltipContent } from "@/components/ui/chart";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { ActivityStats } from "./project-health.graphql";

interface Props { activity: ActivityStats; }

export function ActivitySection({ activity }: Props) {
  const typeEmpty = activity.eventCountByType.length === 0;
  const weeklyData = [...activity.weeklyActivity].reverse();
  const weeklyEmpty = weeklyData.length === 0;

  const typeConfig = { count: { label: "Events", color: "hsl(var(--chart-3))" } };
  const weeklyConfig = {
    events: { label: "Events", color: "hsl(var(--chart-1))" },
    comments: { label: "Comments", color: "hsl(var(--chart-2))" },
  };

  return (
    <div className="grid gap-4 md:grid-cols-2">
      <Card>
        <CardHeader><CardTitle>Events by Type</CardTitle></CardHeader>
        <CardContent>
          {typeEmpty ? (
            <p className="text-sm text-muted-foreground">No data yet</p>
          ) : (
            <ChartContainer config={typeConfig} className="h-56 w-full">
              <BarChart data={activity.eventCountByType} layout="vertical">
                <CartesianGrid horizontal={false} />
                <XAxis type="number" allowDecimals={false} />
                <YAxis dataKey="eventType" type="category" width={140} />
                <ChartTooltip content={<ChartTooltipContent />} />
                <Bar dataKey="count" fill="var(--color-count)" radius={4} />
              </BarChart>
            </ChartContainer>
          )}
        </CardContent>
      </Card>
      <Card>
        <CardHeader><CardTitle>Weekly Activity</CardTitle></CardHeader>
        <CardContent>
          {weeklyEmpty ? (
            <p className="text-sm text-muted-foreground">No data yet</p>
          ) : (
            <ChartContainer config={weeklyConfig} className="h-56 w-full">
              <LineChart data={weeklyData}>
                <CartesianGrid vertical={false} />
                <XAxis dataKey="week" tickLine={false} axisLine={false} />
                <YAxis allowDecimals={false} />
                <ChartTooltip content={<ChartTooltipContent />} />
                <Line type="monotone" dataKey="events" stroke="var(--color-events)" strokeWidth={2} dot={false} />
                <Line type="monotone" dataKey="comments" stroke="var(--color-comments)" strokeWidth={2} dot={false} />
              </LineChart>
            </ChartContainer>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

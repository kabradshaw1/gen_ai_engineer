import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { MemberWorkload } from "./project-health.graphql";

interface Props { members: MemberWorkload[]; }

export function MemberWorkloadTable({ members }: Props) {
  return (
    <Card>
      <CardHeader><CardTitle>Member Workload</CardTitle></CardHeader>
      <CardContent>
        {members.length === 0 ? (
          <p className="text-sm text-muted-foreground">No members assigned</p>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b text-left text-muted-foreground">
                <th className="py-2 font-medium">Name</th>
                <th className="py-2 font-medium text-right">Assigned</th>
                <th className="py-2 font-medium text-right">Completed</th>
              </tr>
            </thead>
            <tbody>
              {members.map((m) => (
                <tr key={m.userId} className="border-b last:border-0">
                  <td className="py-2">{m.name}</td>
                  <td className="py-2 text-right">{m.assignedCount}</td>
                  <td className="py-2 text-right">{m.completedCount}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </CardContent>
    </Card>
  );
}

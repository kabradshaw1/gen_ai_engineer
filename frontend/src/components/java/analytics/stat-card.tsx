import { Card, CardContent } from "@/components/ui/card";

interface Props {
  label: string;
  value: number | null;
  unit?: string;
}

export function StatCard({ label, value, unit }: Props) {
  const display = value === null || value === undefined ? "—" : formatNumber(value);
  return (
    <Card>
      <CardContent className="py-4">
        <div className="text-xs text-muted-foreground uppercase tracking-wide">{label}</div>
        <div className="mt-2 flex items-baseline gap-1">
          <span className="text-2xl font-semibold">{display}</span>
          {unit && value !== null && <span className="text-sm text-muted-foreground">{unit}</span>}
        </div>
      </CardContent>
    </Card>
  );
}

function formatNumber(n: number): string {
  if (Number.isInteger(n)) return n.toString();
  return n.toFixed(1);
}

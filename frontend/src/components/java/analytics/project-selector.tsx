"use client";

interface Project { id: string; name: string; }

interface Props {
  projects: Project[];
  currentId: string;
  onChange: (id: string) => void;
}

export function ProjectSelector({ projects, currentId, onChange }: Props) {
  return (
    <select
      className="rounded-md border bg-background px-3 py-2 text-sm"
      value={currentId}
      onChange={(e) => onChange(e.target.value)}
      aria-label="Select project"
    >
      {projects.map((p) => (
        <option key={p.id} value={p.id}>
          {p.name}
        </option>
      ))}
    </select>
  );
}

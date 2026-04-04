"use client";

import { useQuery, useMutation } from "@apollo/client/react";
import { gql } from "@apollo/client";
import Link from "next/link";
import { Trash2, FolderOpen } from "lucide-react";
import { useState } from "react";
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { CreateProjectDialog } from "./CreateProjectDialog";

const MY_PROJECTS = gql`
  query MyProjects {
    myProjects {
      id
      name
      description
      ownerName
      createdAt
    }
  }
`;

const DELETE_PROJECT = gql`
  mutation DeleteProject($id: ID!) {
    deleteProject(id: $id)
  }
`;

interface Project {
  id: string;
  name: string;
  description: string | null;
  ownerName: string;
  createdAt: string;
}

interface MyProjectsData {
  myProjects: Project[];
}

export function ProjectList() {
  const { data, loading, refetch } = useQuery<MyProjectsData>(MY_PROJECTS);
  const [deleteProject] = useMutation(DELETE_PROJECT);
  const [showCreate, setShowCreate] = useState(false);

  if (loading) {
    return <p className="text-muted-foreground">Loading projects...</p>;
  }

  const projects = data?.myProjects ?? [];

  return (
    <div>
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-semibold">My Projects</h2>
        <Button onClick={() => setShowCreate(true)}>New Project</Button>
      </div>

      {projects.length === 0 ? (
        <p className="mt-6 text-muted-foreground">
          No projects yet. Create one to get started.
        </p>
      ) : (
        <div className="mt-6 grid gap-4">
          {projects.map((project) => (
              <Link
                key={project.id}
                href={`/java/tasks/${project.id}`}
                className="block"
              >
                <Card className="hover:ring-foreground/20 transition-all group">
                  <CardHeader className="flex flex-row items-start justify-between">
                    <div className="flex items-start gap-3">
                      <FolderOpen className="mt-0.5 size-5 text-muted-foreground" />
                      <div>
                        <CardTitle>{project.name}</CardTitle>
                        {project.description && (
                          <CardDescription className="mt-1">
                            {project.description}
                          </CardDescription>
                        )}
                      </div>
                    </div>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="opacity-0 group-hover:opacity-100 transition-opacity"
                      onClick={async (e) => {
                        e.preventDefault();
                        e.stopPropagation();
                        await deleteProject({ variables: { id: project.id } });
                        refetch();
                      }}
                    >
                      <Trash2 className="size-4 text-destructive" />
                    </Button>
                  </CardHeader>
                </Card>
              </Link>
            )
          )}
        </div>
      )}

      {showCreate && (
        <CreateProjectDialog
          onClose={() => setShowCreate(false)}
          onCreated={() => {
            setShowCreate(false);
            refetch();
          }}
        />
      )}
    </div>
  );
}

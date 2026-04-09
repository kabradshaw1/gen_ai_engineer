"use client";

import { NotificationBell } from "@/components/java/NotificationBell";
import { JavaUserDropdown } from "@/components/java/JavaUserDropdown";
import { useAuth } from "@/components/java/AuthProvider";

export function JavaSubHeader() {
  const { isLoggedIn } = useAuth();

  return (
    <div className="border-b border-foreground/10 bg-background">
      <div className="mx-auto grid h-12 max-w-5xl grid-cols-[1fr_auto] items-center gap-4 px-6">
        <h1 className="text-lg font-semibold">Tasks</h1>
        <div className="flex items-center justify-end gap-4">
          {isLoggedIn && <NotificationBell />}
          <JavaUserDropdown />
        </div>
      </div>
    </div>
  );
}

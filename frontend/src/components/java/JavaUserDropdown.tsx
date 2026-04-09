"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useAuth } from "@/components/java/AuthProvider";

export function JavaUserDropdown() {
  const router = useRouter();
  const { user, isLoggedIn, logout } = useAuth();

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        className="flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground transition-colors outline-none"
        aria-label="Account menu"
      >
        {isLoggedIn && user ? (
          <>
            {user.avatarUrl && (
              <img
                src={user.avatarUrl}
                alt=""
                className="size-7 rounded-full"
              />
            )}
            <span>{user.name}</span>
          </>
        ) : (
          "Sign in"
        )}
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        {isLoggedIn && user ? (
          <>
            <DropdownMenuGroup>
              <DropdownMenuLabel className="font-normal">
                <div className="flex flex-col">
                  <span className="text-sm font-medium">{user.name}</span>
                  <span className="text-xs text-muted-foreground">
                    {user.email}
                  </span>
                </div>
              </DropdownMenuLabel>
            </DropdownMenuGroup>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={() => {
                logout();
                router.push("/java/tasks");
              }}
            >
              Sign out
            </DropdownMenuItem>
          </>
        ) : (
          <DropdownMenuItem render={<Link href="/java/tasks" />}>
            Sign in
          </DropdownMenuItem>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

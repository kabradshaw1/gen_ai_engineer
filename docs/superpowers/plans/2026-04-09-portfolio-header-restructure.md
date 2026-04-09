# Portfolio Header Restructure Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Split the global header into a slim brand/nav/resume header, a Java-specific subheader for `/java/tasks/*`, and a site-wide footer for profile links; fix mermaid label rendering; and add architecture + sequence diagrams to `/java` and `/go`.

**Architecture:** Global `SiteHeader` keeps only brand + AI/Java/Go nav + resume. A new `JavaSubHeader` (mounted via a new `app/java/tasks/layout.tsx`) hosts `NotificationBell` and `JavaUserDropdown`, mirroring how `GoSubHeader` serves `/go/*`. A new `SiteFooter` rendered from the root layout holds GitHub/LinkedIn. `MermaidDiagram` passes the SVG profile to DOMPurify so flowchart text stops getting stripped. New mermaid diagrams live inline in `/java/page.tsx` and `/go/page.tsx`.

**Tech Stack:** Next.js (App Router), TypeScript, Tailwind, Apollo Client, mermaid, DOMPurify, Playwright.

**Spec:** `docs/superpowers/specs/2026-04-09-portfolio-header-restructure-design.md`

**Branch:** `frontend-header-restructure` (already created and has the spec commit).

> ⚠️ **Read first:** `frontend/AGENTS.md` warns that this Next.js has breaking changes from training data. Before writing any new `layout.tsx` file, skim the layout docs in `frontend/node_modules/next/dist/docs/` to confirm current conventions.

---

## File Structure

**New files:**
- `frontend/src/components/SiteFooter.tsx` — site-wide footer
- `frontend/src/components/java/JavaSubHeader.tsx` — Java subheader (mirrors `GoSubHeader`)
- `frontend/src/components/java/JavaUserDropdown.tsx` — avatar/name/sign-out dropdown
- `frontend/src/app/java/tasks/layout.tsx` — mounts `JavaSubHeader` under `/java/tasks/*`

**Modified files:**
- `frontend/src/components/SiteHeader.tsx` — strip to brand + nav + resume
- `frontend/src/components/MermaidDiagram.tsx` — pass SVG profile to DOMPurify
- `frontend/src/app/layout.tsx` — render `SiteFooter`
- `frontend/src/app/java/page.tsx` — add diagrams section
- `frontend/src/app/go/page.tsx` — add diagrams section

---

## Task 1: Fix MermaidDiagram label rendering

**Files:**
- Modify: `frontend/src/components/MermaidDiagram.tsx`

- [ ] **Step 1: Read the current implementation**

Read `frontend/src/components/MermaidDiagram.tsx`. Locate the `mermaid.render(...).then(({ svg }) => { ... })` callback. It currently passes `svg` to `DOMPurify.sanitize` with only one argument, then injects the result into `ref.current`.

- [ ] **Step 2: Pass the SVG profile to DOMPurify**

Edit `frontend/src/components/MermaidDiagram.tsx`. Update the sanitize call to enable the SVG profile. The patched callback body should read:

```ts
mermaid.render(id, chart).then(({ svg }) => {
  if (ref.current) {
    const clean = DOMPurify.sanitize(svg, {
      USE_PROFILES: { svg: true, svgFilters: true },
    });
    ref.current.innerHTML = clean;
  }
});
```

Keep every other line identical (the `mermaid.initialize` config, id generation, effect dependencies, JSX, etc.).

- [ ] **Step 3: Verify type-check and lint**

Run from `frontend/`:
```
npm run lint
npx tsc --noEmit
```
Expected: both pass.

- [ ] **Step 4: Manual smoke (dev server)**

Start the frontend dev server (`cd frontend && npm run dev`) and load `/ai`. Confirm the mermaid diagrams on the page now show their node labels and arrow labels. Before the fix they were blank; after the fix they must be legible.

- [ ] **Step 5: Commit**

```
git add frontend/src/components/MermaidDiagram.tsx
git commit -m "fix(frontend): preserve mermaid text via DOMPurify SVG profile"
```

---

## Task 2: Create SiteFooter component

**Files:**
- Create: `frontend/src/components/SiteFooter.tsx`
- Modify: `frontend/src/app/layout.tsx`

- [ ] **Step 1: Write the component**

Create `frontend/src/components/SiteFooter.tsx` with this exact content:

```tsx
import { Github, Linkedin } from "lucide-react";

export function SiteFooter() {
  return (
    <footer className="mt-auto border-t border-foreground/10 bg-background">
      <div className="mx-auto flex max-w-5xl flex-col items-center gap-2 px-6 py-6 text-sm text-muted-foreground sm:flex-row sm:justify-between">
        <p>© {new Date().getFullYear()} Kyle Bradshaw</p>
        <nav className="flex items-center gap-5">
          <a
            href="https://github.com/kabradshaw1"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-1.5 hover:text-foreground transition-colors"
          >
            <Github className="size-4" />
            GitHub
          </a>
          <a
            href="https://www.linkedin.com/in/kyle-bradshaw-15950988/"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-1.5 hover:text-foreground transition-colors"
          >
            <Linkedin className="size-4" />
            LinkedIn
          </a>
        </nav>
      </div>
    </footer>
  );
}
```

- [ ] **Step 2: Mount the footer in the root layout**

Edit `frontend/src/app/layout.tsx`. Import `SiteFooter` next to the existing `SiteHeader` import, and render `<SiteFooter />` as a sibling below `{children}` inside `<AuthProvider>`:

```tsx
import { SiteHeader } from "@/components/SiteHeader";
import { SiteFooter } from "@/components/SiteFooter";
// ...
<AuthProvider>
  <SiteHeader />
  {children}
  <SiteFooter />
</AuthProvider>
```

The `<body>` already has `min-h-full flex flex-col`, so `mt-auto` on the footer will pin it to the bottom on short pages.

- [ ] **Step 3: Verify**

From `frontend/`:
```
npm run lint
npx tsc --noEmit
```
Then dev-smoke `/`, `/ai`, `/java`, `/go` — footer visible on all, GitHub and LinkedIn links open correct URLs in new tabs.

- [ ] **Step 4: Commit**

```
git add frontend/src/components/SiteFooter.tsx frontend/src/app/layout.tsx
git commit -m "feat(frontend): add SiteFooter with GitHub and LinkedIn links"
```

---

## Task 3: Create JavaUserDropdown component

**Files:**
- Create: `frontend/src/components/java/JavaUserDropdown.tsx`

- [ ] **Step 1: Write the component**

Create `frontend/src/components/java/JavaUserDropdown.tsx`. It must use the existing `useAuth` hook from `@/components/java/AuthProvider` (which exposes `user: { name, email, avatarUrl } | null`, `isLoggedIn`, and `logout`) and the existing shadcn `DropdownMenu` primitives, mirroring `GoUserDropdown`.

```tsx
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
```

- [ ] **Step 2: Verify type-check**

```
cd frontend && npx tsc --noEmit
```
Expected: passes. If the `DropdownMenuItem render={<Link ... />}` prop shape has changed in this Next/shadcn version, copy the exact pattern from `GoUserDropdown.tsx` — it is known to work.

- [ ] **Step 3: Commit**

```
git add frontend/src/components/java/JavaUserDropdown.tsx
git commit -m "feat(frontend): add JavaUserDropdown mirroring GoUserDropdown"
```

---

## Task 4: Create JavaSubHeader component

**Files:**
- Create: `frontend/src/components/java/JavaSubHeader.tsx`

- [ ] **Step 1: Write the component**

Create `frontend/src/components/java/JavaSubHeader.tsx`. Follow the same container/grid styling as `GoSubHeader` so the two subheaders feel consistent.

```tsx
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
```

Rationale: `NotificationBell` polls a GraphQL query that requires auth — only render it when logged in to avoid unauthenticated requests. `JavaUserDropdown` handles both the logged-in and logged-out affordance itself.

- [ ] **Step 2: Verify**

```
cd frontend && npx tsc --noEmit && npm run lint
```

- [ ] **Step 3: Commit**

```
git add frontend/src/components/java/JavaSubHeader.tsx
git commit -m "feat(frontend): add JavaSubHeader with notifications and user menu"
```

---

## Task 5: Mount JavaSubHeader under /java/tasks

**Files:**
- Create: `frontend/src/app/java/tasks/layout.tsx`

- [ ] **Step 1: Confirm current Next.js layout conventions**

From `frontend/`, check that nested layouts are still the standard pattern in the bundled Next.js version:
```
ls node_modules/next/dist/docs/ 2>/dev/null | head
```
Read whichever file covers layouts. Confirm that a `layout.tsx` in a route segment automatically wraps that segment's pages. If the convention differs from the standard App Router pattern used by `frontend/src/app/go/layout.tsx`, adapt accordingly.

- [ ] **Step 2: Create the layout file**

Create `frontend/src/app/java/tasks/layout.tsx`:

```tsx
import { JavaSubHeader } from "@/components/java/JavaSubHeader";

export default function JavaTasksLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <>
      <JavaSubHeader />
      {children}
    </>
  );
}
```

Note: `AuthProvider` is already mounted by the root layout, so the subheader can freely use `useAuth` and `NotificationBell`.

- [ ] **Step 3: Verify**

From `frontend/`:
```
npx tsc --noEmit
npm run build
```
Then dev-smoke:
- `/java` → subheader **not** present (overview page stays clean).
- `/java/tasks` logged out → subheader present, "Sign in" label visible, no NotificationBell.
- `/java/tasks` logged in → subheader present, NotificationBell + user dropdown visible.

- [ ] **Step 4: Commit**

```
git add frontend/src/app/java/tasks/layout.tsx
git commit -m "feat(frontend): mount JavaSubHeader under /java/tasks"
```

---

## Task 6: Slim down SiteHeader

**Files:**
- Modify: `frontend/src/components/SiteHeader.tsx`

- [ ] **Step 1: Rewrite the component**

Replace the entire contents of `frontend/src/components/SiteHeader.tsx` with:

```tsx
"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { FileText } from "lucide-react";

export function SiteHeader() {
  const pathname = usePathname();

  const isActive = (prefix: string) =>
    pathname === prefix || pathname.startsWith(prefix + "/");

  const navLinkClass = (prefix: string) =>
    isActive(prefix)
      ? "text-sm text-foreground border-b-2 border-foreground pb-px transition-colors"
      : "text-sm text-muted-foreground hover:text-foreground transition-colors";

  return (
    <header className="border-b border-foreground/10 bg-background">
      <div className="mx-auto flex h-14 max-w-5xl items-center justify-between px-6">
        <div className="flex items-center gap-6">
          <Link href="/" className="text-lg font-semibold">
            Kyle Bradshaw
          </Link>
          <nav className="flex items-center gap-4">
            <Link href="/ai" className={navLinkClass("/ai")}>
              AI
            </Link>
            <Link href="/java" className={navLinkClass("/java")}>
              Java
            </Link>
            <Link href="/go" className={navLinkClass("/go")}>
              Go
            </Link>
          </nav>
        </div>
        <a
          href="/resume.pdf"
          aria-label="Resume"
          className="text-muted-foreground hover:text-foreground transition-colors"
        >
          <FileText className="size-5" />
        </a>
      </div>
    </header>
  );
}
```

This removes:
- the `useAuth` / `NotificationBell` imports and the logged-in block
- the Portfolio, GitHub, LinkedIn links (GitHub/LinkedIn are now in `SiteFooter`; the Portfolio link pointed to a GitHub repo and is redundant with the footer's GitHub link)

- [ ] **Step 2: Verify type-check, lint, and build**

```
cd frontend && npx tsc --noEmit && npm run lint && npm run build
```

- [ ] **Step 3: Dev smoke**

- `/`, `/ai`, `/java`, `/go` — header shows only brand + nav + resume icon; no GitHub/LinkedIn/NotificationBell/avatar.
- `/java/tasks` logged in — global header is still clean; the NotificationBell + user dropdown appear in the Java subheader instead.
- Resize to mobile width — brand, nav, and resume still fit on one row (if they do not, add `flex-wrap` to the outer flex container).

- [ ] **Step 4: Commit**

```
git add frontend/src/components/SiteHeader.tsx
git commit -m "refactor(frontend): slim SiteHeader to brand, nav, and resume"
```

---

## Task 7: Update Playwright E2E for header changes

**Files:**
- Modify: any file under `frontend/e2e/` that asserts on removed header elements (GitHub/LinkedIn/Portfolio links, `NotificationBell`, sign-out button in the global header, avatar in the global header)

- [ ] **Step 1: Find affected specs**

From repo root, run:
```
grep -rn -E "GitHub|LinkedIn|Portfolio|Sign out|NotificationBell|avatarUrl|bell" frontend/e2e || true
```
List every match. For each, decide whether the assertion:
- (a) should move to `/java/tasks` scope (notifications, sign-out, avatar), or
- (b) should move to the footer scope (GitHub, LinkedIn), or
- (c) should be deleted (Portfolio link, which no longer exists).

- [ ] **Step 2: Update the specs**

Apply the categorization from Step 1. When asserting against the footer, scope the locator with a `footer` selector (e.g., `page.locator('footer').getByRole('link', { name: 'GitHub' })`). When asserting against the Java subheader, navigate to `/java/tasks` first and scope the locator to that subheader region.

- [ ] **Step 3: Run mocked E2E**

```
make preflight-e2e
```
Expected: pass. If a test fails because of a genuinely changed flow (e.g., the login button is now inside the subheader), update the selector — do not weaken the assertion.

- [ ] **Step 4: Commit**

```
git add frontend/e2e
git commit -m "test(frontend): update E2E selectors for header restructure"
```

---

## Task 8: Add diagrams section to /java page

**Files:**
- Modify: `frontend/src/app/java/page.tsx`

- [ ] **Step 1: Add a diagrams section above the "Open Task Manager" CTA**

Edit `frontend/src/app/java/page.tsx`. Import `MermaidDiagram` at the top:

```tsx
import { MermaidDiagram } from "@/components/MermaidDiagram";
```

Then insert a new `<section>` between the "Task Management System" section and the "Open Task Manager" CTA section:

```tsx
<section className="mt-12">
  <h2 className="text-2xl font-semibold">Architecture</h2>
  <p className="mt-4 text-muted-foreground leading-relaxed">
    Four Spring Boot services fronted by a GraphQL gateway. Writes land in
    Postgres; events fan out through RabbitMQ to the activity and
    notification services, which persist into MongoDB.
  </p>
  <div className="mt-6">
    <MermaidDiagram
      chart={`flowchart LR
  FE[Next.js Frontend]
  GW[gateway-service<br/>GraphQL]
  TS[task-service<br/>Spring Boot]
  AS[activity-service<br/>Spring Boot]
  NS[notification-service<br/>Spring Boot]
  PG[(PostgreSQL)]
  MG[(MongoDB)]
  RD[(Redis cache)]
  MQ{{RabbitMQ}}
  FE -->|GraphQL| GW
  GW -->|REST| TS
  GW -->|REST| AS
  GW -->|REST| NS
  TS --> PG
  TS -->|publish task.events| MQ
  MQ -->|consume| AS
  MQ -->|consume| NS
  AS --> MG
  AS --> RD
  NS --> MG`}
    />
  </div>

  <h3 className="mt-10 text-xl font-semibold">Request flow: Create a task</h3>
  <p className="mt-4 text-muted-foreground leading-relaxed">
    One click traces through the gateway, into task-service, onto RabbitMQ,
    and fans out to activity and notification consumers in parallel.
  </p>
  <div className="mt-6">
    <MermaidDiagram
      chart={`sequenceDiagram
  participant U as User
  participant FE as Frontend
  participant GW as gateway
  participant TS as task-service
  participant PG as Postgres
  participant MQ as RabbitMQ
  participant AS as activity-service
  participant NS as notification-service
  participant MG as MongoDB
  U->>FE: Click "Create task"
  FE->>GW: mutation createTask
  GW->>TS: POST /tasks
  TS->>PG: INSERT task
  PG-->>TS: ok
  TS->>MQ: publish task.created
  TS-->>GW: 201 task
  GW-->>FE: task payload
  par fan-out
    MQ->>AS: task.created
    AS->>MG: insert activity
  and
    MQ->>NS: task.created
    NS->>MG: insert notification
  end
  FE->>GW: poll myNotifications (30s)
  GW->>NS: GET /notifications
  NS-->>FE: unread badge updates`}
    />
  </div>
</section>
```

Note: mermaid chokes on literal parentheses in `participant` aliases, so the gateway participant is `gateway` (not `gateway (GraphQL)`).

- [ ] **Step 2: Verify**

```
cd frontend && npx tsc --noEmit && npm run lint && npm run build
```

- [ ] **Step 3: Dev smoke**

Load `/java`. Both diagrams must render with visible text (the Task 1 fix enables this). The page layout should still look clean at `max-w-3xl` — the `overflow-x-auto` on `MermaidDiagram` handles wide diagrams.

- [ ] **Step 4: Commit**

```
git add frontend/src/app/java/page.tsx
git commit -m "feat(frontend): add architecture and request-flow diagrams to /java"
```

---

## Task 9: Add diagrams section to /go page

**Files:**
- Modify: `frontend/src/app/go/page.tsx`

- [ ] **Step 1: Verify the checkout flow against the actual service**

Before writing the sequence diagram, read `go/ecommerce-service/` (handler for `POST /orders` and any worker pool code) to confirm these claims from the spec:
- JWT is validated inside `ecommerce-service`
- Cart is read from Redis with key `cart:<userId>`
- `INSERT order (status=pending)` happens before publish
- A RabbitMQ message is published on the `order.created` topic
- A worker pool consumes and updates the order to `confirmed`, then deletes the Redis cart key

If any of these are wrong, adjust the mermaid in Step 2 to match reality. Do not ship a diagram that misrepresents the code — the point of the portfolio is accuracy.

- [ ] **Step 2: Add the diagrams section**

Edit `frontend/src/app/go/page.tsx`. Import `MermaidDiagram`:

```tsx
import { MermaidDiagram } from "@/components/MermaidDiagram";
```

Insert a new `<section>` between the "Ecommerce Platform" section (ending with the tech stack `<ul>`) and the "View Project" CTA:

```tsx
<section className="mt-12">
  <h2 className="text-2xl font-semibold">Architecture</h2>
  <p className="mt-4 text-muted-foreground leading-relaxed">
    Two Go services — auth and ecommerce — sharing Postgres. The ecommerce
    service caches reads in Redis and offloads order finalization to a
    RabbitMQ-driven goroutine worker pool.
  </p>
  <div className="mt-6">
    <MermaidDiagram
      chart={`flowchart LR
  FE[Next.js Frontend]
  AUTH[auth-service<br/>Go + JWT]
  EC[ecommerce-service<br/>Go]
  PG[(PostgreSQL)]
  RD[(Redis cache)]
  MQ{{RabbitMQ}}
  WP[Worker pool<br/>goroutines]
  FE -->|REST /go-auth| AUTH
  FE -->|REST /go-api| EC
  AUTH --> PG
  EC --> PG
  EC --> RD
  EC -->|publish order.events| MQ
  MQ --> WP
  WP --> PG`}
    />
  </div>

  <h3 className="mt-10 text-xl font-semibold">Request flow: Checkout order</h3>
  <p className="mt-4 text-muted-foreground leading-relaxed">
    The HTTP handler returns immediately after persisting a pending order and
    publishing to RabbitMQ. A worker pool finalizes the order asynchronously.
  </p>
  <div className="mt-6">
    <MermaidDiagram
      chart={`sequenceDiagram
  participant U as User
  participant FE as Frontend
  participant AUTH as auth-service
  participant EC as ecommerce-service
  participant RD as Redis
  participant PG as Postgres
  participant MQ as RabbitMQ
  participant WP as Worker pool
  U->>FE: Click "Checkout"
  FE->>AUTH: POST /login
  AUTH->>PG: verify creds
  AUTH-->>FE: JWT
  FE->>EC: POST /orders (Bearer JWT)
  EC->>EC: validate JWT
  EC->>RD: GET cart:userId
  RD-->>EC: cart items
  EC->>PG: INSERT order (status=pending)
  EC->>MQ: publish order.created
  EC-->>FE: 202 orderId
  MQ->>WP: order.created
  WP->>PG: charge + update status=confirmed
  WP->>RD: DEL cart:userId
  FE->>EC: GET /orders/{id}
  EC-->>FE: status=confirmed`}
    />
  </div>
</section>
```

If Step 1 revealed a different flow, adjust the arrows accordingly before running the build.

- [ ] **Step 3: Verify**

```
cd frontend && npx tsc --noEmit && npm run lint && npm run build
```

- [ ] **Step 4: Dev smoke**

Load `/go` — both diagrams render with visible text. Verify `/go/ecommerce` still works (subheader, categories, cart).

- [ ] **Step 5: Commit**

```
git add frontend/src/app/go/page.tsx
git commit -m "feat(frontend): add architecture and checkout-flow diagrams to /go"
```

---

## Task 10: Full preflight and manual smoke

- [ ] **Step 1: Run full frontend preflight**

```
make preflight-frontend
```
Expected: pass.

- [ ] **Step 2: Run E2E preflight**

```
make preflight-e2e
```
Expected: pass. Any remaining failures must be diagnosed and fixed — do not merge past red tests.

- [ ] **Step 3: Manual smoke checklist**

Start `cd frontend && npm run dev` and walk through:

- [ ] `/` — slim header (brand + AI/Java/Go + resume), footer visible, no Java-specific UI.
- [ ] `/ai` — mermaid diagrams render **with visible node and edge labels** (Task 1 regression gate).
- [ ] `/java` — slim header, no subheader, two new diagrams render with text, footer visible.
- [ ] `/java/tasks` logged out — Java subheader visible with "Sign in" affordance, no NotificationBell, login form still rendered by `TasksPageContent`.
- [ ] `/java/tasks` logged in — Java subheader shows NotificationBell (with unread badge if any) + avatar/name dropdown with Sign out; notifications dropdown opens, "Mark all read" works.
- [ ] `/go` — slim header, two new diagrams render with text.
- [ ] `/go/ecommerce` — `GoSubHeader` unchanged and still functional (categories + cart + user dropdown).
- [ ] Mobile width (≤ 400px) — global header does not overflow; subheaders and footer reflow acceptably.

- [ ] **Step 4: Done**

Leave the branch committed locally. Per CLAUDE.md, Kyle handles all pushes and merges — do not `git push`.

---

## Self-Review Notes

- Spec coverage: Goals 1 (Task 6), 2 (Tasks 3–5), 3 (Task 2), 4 (Task 1), 5 (Tasks 8–9). ✓
- Task 1 fixes the mermaid bug before Tasks 8 and 9 add new mermaid diagrams, so the new diagrams never ship broken.
- Task 6 (header slim-down) happens after Tasks 2–5 (footer + subheader) so at no point is Java auth UI homeless — users can always reach sign-in during the transition.
- Task 7 (E2E updates) depends on Task 6 being merged conceptually; if E2E runs before Task 6, it will still pass (old selectors still present) — tests genuinely break only after the header slim-down. Order as written is fine.
- No placeholders, every code step has complete code, every command is exact.

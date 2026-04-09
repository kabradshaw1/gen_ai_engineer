# Portfolio Header Restructure + Section Diagrams

**Date:** 2026-04-09
**Status:** Approved (design)
**Owner:** Kyle Bradshaw

## Motivation

The current `SiteHeader` conflates four concerns: brand/navigation, external profile links (GitHub/LinkedIn), the Java-only `NotificationBell`, and the Java-only login/user menu. This makes the header visually busy and semantically confused — Java-specific UI appears on AI and Go pages where it has no meaning.

Separately, the existing `/ai` mermaid diagrams render with missing label text because `DOMPurify.sanitize` strips `<foreignObject>` elements (which modern mermaid uses for flowchart labels). And `/java` and `/go` have no architecture or request-flow diagrams, which is a gap for a portfolio aimed at microservices roles.

## Goals

1. A clean global header that is the same on every page and contains only brand, section nav, and resume.
2. Java-specific UI (login, user dropdown, notifications) lives in a Java subheader visible only on `/java/tasks/*`, mirroring how `GoSubHeader` works for `/go/*`.
3. GitHub and LinkedIn profile links move into a new site-wide footer.
4. Fix the mermaid text rendering bug so labels are visible on `/ai` and on the new diagrams.
5. Add architecture and request-flow mermaid diagrams to `/java` and `/go`, featuring RabbitMQ fan-out and the Go worker pool — the flows that demonstrate microservice understanding.

## Non-Goals

- No changes to backend services.
- No changes to the `/java` or `/go` page copy beyond adding diagram sections.
- No auth/login UX changes. Login form still lives in `TasksPageContent`.
- No redesign of `GoSubHeader`.
- No footer content beyond GitHub, LinkedIn, copyright, and optional tech tagline.

## Architecture

### Layout tree after the change

```
app/layout.tsx
├── SiteHeader          (global, always visible)
├── {children}
│   ├── app/java/tasks/layout.tsx
│   │   └── JavaSubHeader   (only under /java/tasks/*)
│   └── app/go/layout.tsx
│       └── GoSubHeader     (unchanged, only under /go/*)
└── SiteFooter          (global, always visible)
```

### Global header (`SiteHeader`)

- **Left cluster:** `Kyle Bradshaw` wordmark (routes to `/`) immediately followed by nav links `AI · Java · Go`. Treated as one visual unit — the brand and nav belong together.
- **Right cluster:** Resume PDF link only.
- **Removed:** GitHub, LinkedIn, `NotificationBell`, and the Java login/avatar/sign-out block.
- Height and background styling unchanged. Active link state preserved.

### Java subheader (`JavaSubHeader`)

- New component at `frontend/src/components/java/JavaSubHeader.tsx`.
- Visual treatment mirrors `GoSubHeader`: same container width, same border/padding, same typography.
- **Mount point:** a new `frontend/src/app/java/tasks/layout.tsx`. The subheader appears only on `/java/tasks` and descendants. The `/java` overview page stays clean.
- **Contents:**
  - Left: section title `"Tasks"`.
  - Right: `NotificationBell` (existing component, moved), then `JavaUserDropdown` when logged in, or a `"Sign in"` link when logged out.
- When logged out, the login form itself continues to render inside `TasksPageContent` — the subheader only shows a "Sign in" affordance that anchors/scrolls to it (or is a no-op link to `/java/tasks`).

### Java user dropdown (`JavaUserDropdown`)

- New component at `frontend/src/components/java/JavaUserDropdown.tsx`.
- Extracted from the current `SiteHeader` logged-in block: avatar, display name, sign-out button.
- Visually mirrors `GoUserDropdown` so the two sections feel consistent.

### Site footer (`SiteFooter`)

- New component at `frontend/src/components/SiteFooter.tsx`.
- Rendered in `app/layout.tsx` after `{children}`.
- Contents:
  - GitHub link (icon + label)
  - LinkedIn link (icon + label)
  - Copyright line: `© 2026 Kyle Bradshaw`
  - Optional small tagline: `Built with Next.js, FastAPI, Spring Boot, and Go`
- Centered, muted foreground color, thin top border, modest vertical padding. Non-sticky.

### Mermaid fix

- In `frontend/src/components/MermaidDiagram.tsx`, update the `DOMPurify.sanitize` call to pass the SVG profile: `{ USE_PROFILES: { svg: true, svgFilters: true } }`. The existing code injects the sanitized SVG into the ref via the React DOM API it already uses; only the sanitize options change.
- Root cause: DOMPurify's default HTML profile strips `<foreignObject>`, which modern mermaid uses to render all flowchart label text. The SVG profile preserves it. This fix applies to every page that uses `MermaidDiagram`, including `/ai`.

### Section diagrams

Both `/java/page.tsx` and `/go/page.tsx` get a new section rendered below their existing tech-stack content, containing two `MermaidDiagram` instances each. Mermaid source is inline in the page file (matches the existing `/ai` pattern).

**Java — architecture flowchart:**
```
flowchart LR
  FE[Next.js Frontend]
  GW[gateway-service\nGraphQL]
  TS[task-service\nSpring Boot]
  AS[activity-service\nSpring Boot]
  NS[notification-service\nSpring Boot]
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
  NS --> MG
```

**Java — sequence diagram, "Create a task":**
```
sequenceDiagram
  participant U as User
  participant FE as Frontend
  participant GW as gateway (GraphQL)
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
  NS-->>FE: unread badge updates
```

**Go — architecture flowchart:**
```
flowchart LR
  FE[Next.js Frontend]
  AUTH[auth-service\nGo + JWT]
  EC[ecommerce-service\nGo]
  PG[(PostgreSQL)]
  RD[(Redis cache)]
  MQ{{RabbitMQ}}
  WP[Worker pool\ngoroutines]
  FE -->|REST /go-auth| AUTH
  FE -->|REST /go-api| EC
  AUTH --> PG
  EC --> PG
  EC --> RD
  EC -->|publish order.events| MQ
  MQ --> WP
  WP --> PG
```

**Go — sequence diagram, "Checkout order":**
```
sequenceDiagram
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
  EC-->>FE: status=confirmed
```

Before merging, Kyle should sanity-check the Go checkout flow against the actual `ecommerce-service` implementation — if the real flow differs (e.g., no JWT validation inside the service, or no Redis cart cache), the sequence should be adjusted to match reality.

## Files Touched

- `frontend/src/components/SiteHeader.tsx` — strip down to brand+nav+resume
- `frontend/src/components/SiteFooter.tsx` — **new**
- `frontend/src/components/java/JavaSubHeader.tsx` — **new**
- `frontend/src/components/java/JavaUserDropdown.tsx` — **new**
- `frontend/src/app/layout.tsx` — mount `SiteFooter`
- `frontend/src/app/java/tasks/layout.tsx` — **new**, mounts `JavaSubHeader`
- `frontend/src/app/java/page.tsx` — add diagrams section
- `frontend/src/app/go/page.tsx` — add diagrams section
- `frontend/src/components/MermaidDiagram.tsx` — DOMPurify SVG profile fix
- `frontend/e2e/**` — update any selectors broken by removed header elements

## Testing

- `make preflight-frontend` (tsc, lint, Next.js build)
- `make preflight-e2e` (Playwright mocked E2E) — expect selector fixes for the header move
- Manual smoke in dev:
  - `/` — header, footer present
  - `/ai` — diagrams render with **visible text** (verifies DOMPurify fix)
  - `/java` — header has no Java login/notification UI; architecture + sequence diagrams render
  - `/java/tasks` logged out — `JavaSubHeader` visible with "Sign in" affordance; login form still works
  - `/java/tasks` logged in — `JavaSubHeader` shows `NotificationBell` + `JavaUserDropdown`; notifications poll and mark-read work
  - `/go` — architecture + sequence diagrams render
  - `/go/ecommerce` — `GoSubHeader` unchanged and still functional
- Responsive check at mobile width for the slimmer header.

## Open Risks

- **E2E selector drift.** Playwright tests that find login controls or GitHub/LinkedIn in the header will break. Expected and in-scope to fix.
- **Go sequence accuracy.** The proposed checkout flow is based on the architecture summary, not a line-by-line read of `ecommerce-service`. Verify against actual code during implementation and adjust the mermaid source if reality differs.
- **New Next.js conventions.** `frontend/AGENTS.md` warns that this Next.js has breaking changes from training data. When creating `app/java/tasks/layout.tsx`, consult `node_modules/next/dist/docs/` before writing the file to confirm current layout conventions.

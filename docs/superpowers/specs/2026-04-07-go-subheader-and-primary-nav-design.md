# Primary Nav Restructure + Go Sub-Header Design

**Date:** 2026-04-07
**Status:** Approved
**Stack:** Next.js 15, TypeScript, shadcn/ui, Tailwind

## Summary

Two related UI changes:

1. Replace the single hardcoded "Go" link in the top header with three section
   links — `AI`, `Java`, `Go` — each with a prefix-matched active-route
   indicator (color + underline).
2. Add a secondary header that renders on every `/go/*` route, containing a
   cart icon (with live badge) and a user dropdown menu. Remove the inline
   login/register form currently inlined on `/go/ecommerce`; auth UX moves to
   the dropdown (Sign in / Register links) and the dedicated pages built in
   the previous Google OAuth spec.

## Goals

- Users can see at a glance which portfolio section they are in.
- `/go/*` routes have a consistent secondary navigation with cart and user
  state.
- `/go/ecommerce` page becomes cleaner by shedding auth UI it shouldn't own.

## Non-goals

- Sub-headers for `/ai` or `/java` (these sections have no equivalent needs
  yet).
- Profile page or user settings.
- Mobile hamburger nav (tablet+ only in this spec).
- Product detail page changes, orders page changes.
- Cart mutation via context (`addItem`, `removeItem`) — pages that mutate
  still POST directly and call `refresh()` on the context.
- Automated tests for the new components (manual smoke test only, consistent
  with the Go OAuth spec).
- Changes to Java's `SiteHeader` auth/notification block.

## Architecture

### Primary header (`frontend/src/components/SiteHeader.tsx`)

Replace the hardcoded `<Link href="/go">Go</Link>` with three section links
driven by `usePathname()` from `next/navigation`:

```tsx
const pathname = usePathname();
const isActive = (prefix: string) =>
  pathname === prefix || pathname.startsWith(prefix + "/");

const navLinkClass = (prefix: string) =>
  isActive(prefix)
    ? "text-sm text-foreground border-b-2 border-foreground pb-px transition-colors"
    : "text-sm text-muted-foreground hover:text-foreground transition-colors";
```

Applied to:

```tsx
<Link href="/ai" className={navLinkClass("/ai")}>AI</Link>
<Link href="/java" className={navLinkClass("/java")}>Java</Link>
<Link href="/go" className={navLinkClass("/go")}>Go</Link>
```

The three section links sit where the single `Go` link was. External links
(Portfolio, GitHub, LinkedIn, resume icon) and the Java auth block are
unchanged.

**Matching:** prefix match. `/go`, `/go/ecommerce`, `/go/login`, and
`/go/ecommerce/cart` all light up the "Go" link. `/` (homepage) matches none
of the three — intentional.

**Styling:** active link uses `text-foreground` plus `border-b-2
border-foreground pb-px`. The `pb-px` nudges the underline one pixel below
descenders. Inactive links keep the current muted → foreground hover
vocabulary.

### Go sub-header (new)

Mounted inside `frontend/src/app/go/layout.tsx`, renders on every `/go/*`
page just below the primary header. Height `h-12` (primary is `h-14`) so it
reads as secondary. Right-aligned content — no left-side nav items because
`/go` has a single content area and the section is already indicated by the
primary nav's active "Go" link.

Three new components (one concern each):

- **`GoSubHeader.tsx`** — layout shell. Renders `<GoCartIcon />` and
  `<GoUserDropdown />`.
- **`GoCartIcon.tsx`** — cart icon with badge; hidden when logged out.
- **`GoUserDropdown.tsx`** — shadcn `DropdownMenu` with conditional items.

### Go cart context (new)

**`frontend/src/components/go/GoCartProvider.tsx`** — wraps `/go/*` inside
the layout. Fetches `/cart` from the ecommerce service once on login,
exposes a minimal API:

```ts
interface GoCartContextType {
  items: CartItem[];
  count: number;           // sum of quantities
  refresh: () => Promise<void>;
}
```

Deliberately minimal:
- No `addItem` / `removeItem` mutators. Pages that mutate the cart POST
  directly to the API and then call `refresh()` on the context.
- Fetch errors are swallowed silently — the badge stays stale rather than
  showing an error. Real failures surface on the cart page.
- Not persisted to localStorage. Server is the source of truth; one fetch on
  login is cheap.
- On logout the context clears `items` to `[]`.

The `CartItem` shape will be verified against the actual `/cart` response
during implementation. The badge `count` depends only on `quantity`, so even
if field names shift, the badge continues to work.

## Components

### `frontend/src/app/go/layout.tsx` (modify)

```tsx
import { GoAuthProvider } from "@/components/go/GoAuthProvider";
import { GoCartProvider } from "@/components/go/GoCartProvider";
import { GoSubHeader } from "@/components/go/GoSubHeader";

export default function GoLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <GoAuthProvider>
      <GoCartProvider>
        <GoSubHeader />
        {children}
      </GoCartProvider>
    </GoAuthProvider>
  );
}
```

### `GoCartProvider.tsx` (create)

Standard React context. Hooks:
- `useGoAuth()` → when `isLoggedIn` flips true, trigger `refresh()`. When
  flips false, clear items.
- `getGoAccessToken()` from `@/lib/go-auth` → Bearer token for the fetch.

Fetch target: `${GO_ECOMMERCE_URL}/cart`. Response assumed to be
`{ items: CartItem[] }` — verified during implementation.

### `GoSubHeader.tsx` (create)

```tsx
"use client";

import { GoCartIcon } from "./GoCartIcon";
import { GoUserDropdown } from "./GoUserDropdown";

export function GoSubHeader() {
  return (
    <div className="border-b border-foreground/10 bg-background">
      <div className="mx-auto flex h-12 max-w-5xl items-center justify-end gap-4 px-6">
        <GoCartIcon />
        <GoUserDropdown />
      </div>
    </div>
  );
}
```

### `GoCartIcon.tsx` (create)

- Reads `isLoggedIn` from `useGoAuth()`; returns `null` when logged out
- Reads `count` from `useGoCart()`
- Renders a `lucide-react` `ShoppingCart` icon inside a `Link` to
  `/go/ecommerce/cart`
- Badge: absolute-positioned dot at `-right-2 -top-2`, `size-4`, rounded
  full, `bg-foreground text-background`, rendered only when `count > 0`.
  Caps at `99+`.

### `GoUserDropdown.tsx` (create)

Uses shadcn `DropdownMenu`. Trigger content:
- **Logged in + avatar:** `<img>` of `user.avatarUrl`, `size-7 rounded-full`
- **Logged in + no avatar:** initials in a circular muted background
- **Logged out:** text "Welcome" in muted-foreground

Menu content:
- **Logged in:** `DropdownMenuLabel` with name + email, separator,
  `DropdownMenuItem` → `/go/ecommerce/orders`, `DropdownMenuItem` → sign out
  (calls `logout()` then `router.push("/go/ecommerce")`)
- **Logged out:** `DropdownMenuItem` → `/go/login`, `DropdownMenuItem` →
  `/go/register`

### `frontend/src/app/go/ecommerce/page.tsx` (modify)

Strip auth UI. Remove:
- `authMode`, `email`, `password`, `name`, `authError` state
- `handleAuth` callback
- The entire header-bar block (login form / user chip / sign-out button)

Keep:
- Page title ("Store")
- Category filter
- Product grid
- Empty state
- `useEffect` that fetches products and categories

The page becomes substantially shorter.

### `frontend/src/components/SiteHeader.tsx` (modify)

- Add `usePathname` import
- Add the `isActive` / `navLinkClass` helpers
- Replace the single `Go` link with three section links using
  `navLinkClass`

No other changes to `SiteHeader`.

## Data flow

```
User navigates to /go/*
  │
  ├──> GoLayout mounts
  │      │
  │      ├──> GoAuthProvider provides { user, isLoggedIn, ... }
  │      │
  │      └──> GoCartProvider mounts
  │             │
  │             ├──> useEffect watches isLoggedIn
  │             │      └──> if true: GET /cart → setItems
  │             │
  │             └──> GoSubHeader renders
  │                    │
  │                    ├──> GoCartIcon reads { count } from useGoCart()
  │                    └──> GoUserDropdown reads { user, isLoggedIn } from useGoAuth()
  │
  └──> /go/ecommerce/page.tsx (child) fetches products

User adds item to cart (cart page)
  │
  ├──> POST /cart with product id
  └──> call refresh() → GET /cart → setItems → count updates → badge re-renders
```

## Prerequisites

Before implementation, verify `@/components/ui/dropdown-menu` exists:

```bash
ls frontend/src/components/ui/dropdown-menu.tsx
```

If missing, install it:

```bash
cd frontend && npx shadcn@latest add dropdown-menu
```

The implementation plan will call this out as step 1.

## Testing

Manual smoke test only (no automated tests in this spec):

1. **Primary nav active indicator**
   - Visit `/`. None of AI/Java/Go highlighted.
   - Visit `/ai`. "AI" highlighted.
   - Visit `/ai/rag`. "AI" still highlighted.
   - Visit `/java/tasks`. "Java" highlighted.
   - Visit `/go/ecommerce`. "Go" highlighted.
   - Visit `/go/login`. "Go" still highlighted.

2. **Go sub-header — logged out**
   - Visit `/go/ecommerce`. Cart icon hidden. Dropdown shows "Welcome".
   - Open dropdown. Items: Sign in, Register. Clicking each routes correctly.

3. **Go sub-header — logged in (email path)**
   - Sign in via `/go/login` with email/password.
   - Cart icon visible, badge hidden (count = 0).
   - Dropdown shows avatar initials (no Google avatar).
   - Open dropdown. Label shows name + email. Items: Orders, Sign out.
   - Click Sign out. Dropdown reverts to "Welcome".

4. **Go sub-header — logged in (Google path)**
   - Sign in via `/go/login` with Google.
   - Dropdown trigger shows Google profile picture.
   - Same items as above.

5. **Cart badge**
   - Log in, verify badge absent.
   - Add a product to cart on `/go/ecommerce/cart` (existing flow).
   - Verify badge shows count after the cart page's own refresh triggers the
     provider (if the cart page doesn't trigger a refresh, that's an
     implementation note — wire it up).

6. **Ecommerce page cleanup**
   - `/go/ecommerce` no longer shows an inline login form or user chip.
   - Product grid still renders; category filter still works.

## Rollout

Single feature branch, single PR. No migrations, no backend changes. Pair
with the uncommitted `.gitignore` change from the OAuth work.

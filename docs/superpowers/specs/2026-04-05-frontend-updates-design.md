# Frontend Updates & Password Auth Design

**Date:** 2026-04-05
**Status:** Approved

## Overview

A collection of frontend updates to the portfolio site: filling in placeholder bios, fixing Mermaid diagram rendering, updating tech stack text, adding a global header, implementing password authentication for the Java task manager, and renaming the repo from `gen_ai_engineer` to `portfolio`.

---

## 1. Bio Content

Replace placeholder text on three pages with short professional bios.

### `/` (Home — `frontend/src/app/page.tsx`)

> Software engineer focused on building production systems with modern tooling. Since August 2022, I've been working full-time on personal projects and consulting, with a focus on Go, TypeScript, and cloud-native infrastructure. This portfolio showcases two areas of specialization — AI/ML engineering and full-stack Java development.

### `/ai` (AI Engineer — `frontend/src/app/ai/page.tsx`)

> Building intelligent systems with retrieval-augmented generation and agentic architectures. This section demonstrates RAG pipelines, vector search, LLM orchestration, and tool-using agents — built with FastAPI, Qdrant, and Ollama, deployed on Kubernetes.

### `/java` (Java Developer — `frontend/src/app/java/page.tsx`)

> Full-stack microservices architecture with Spring Boot, GraphQL, and event-driven communication. This section demonstrates a task management platform built with four Java services, PostgreSQL, MongoDB, Redis, and RabbitMQ — deployed on Kubernetes.

---

## 2. Mermaid Diagram Fix

**File:** `frontend/src/components/MermaidDiagram.tsx`

**Problem:** Two `MermaidDiagram` components on `/ai` both use `Date.now()` to generate render IDs. When both render in the same tick, they get the same ID, causing the second diagram's text to not render.

**Fix:** Replace `Date.now()` with a stable unique ID. Use React's `useId()` hook (sanitized for mermaid's ID requirements) or a module-level counter that increments per instance. Each call to `mermaid.render()` must receive a globally unique ID string.

---

## 3. Tech Stack Updates

Update the tech stack bullet points to reflect that production runs on Minikube.

### `/ai` — Document Q&A Assistant

Change:
- "Docker Compose orchestration"

To:
- "Minikube Kubernetes deployment (production), Docker Compose (local dev)"

### `/ai` — Debug Assistant

Add bullet:
- "Minikube Kubernetes deployment (production)"

### `/java` — Task Management System

Change:
- "Docker Compose + Minikube Kubernetes manifests"

To:
- "Minikube Kubernetes deployment (production), Docker Compose (local dev)"

---

## 4. Global Header

**Current state:** `SiteHeader` component lives at `frontend/src/components/java/SiteHeader.tsx` and is only rendered in `frontend/src/app/java/layout.tsx`. The `/` and `/ai` pages use inline `← Home` links for navigation.

**Changes:**

1. Move `SiteHeader` from `frontend/src/components/java/` to `frontend/src/components/SiteHeader.tsx` (it's now global).
2. Add it to the root layout (`frontend/src/app/layout.tsx`).
3. Remove it from the Java layout (`frontend/src/app/java/layout.tsx`).
4. Update the "Kyle Bradshaw" link from `/java` to `/`.
5. Add a "Portfolio" link pointing to `https://github.com/kabradshaw1/portfolio`.
6. Remove the inline `← Home` / `← Back` links from `/ai` and `/java` pages (the header now handles navigation).
7. The authenticated user section (avatar, name, sign out, notifications) only renders when the user is logged in — this continues to work as-is since `useAuth()` returns `isLoggedIn: false` when not on Java task pages.

**Note:** The `SiteHeader` uses `useAuth()` from `AuthProvider`. The `AuthProvider` currently wraps only the Java layout. Since the header is now global, `AuthProvider` needs to wrap the root layout (or the header needs to gracefully handle the absence of auth context). The simplest approach: move `AuthProvider` to root layout — it's a no-op on non-Java pages since there are no auth tokens in localStorage for those routes.

---

## 5. Password Authentication

### 5.1 Backend (task-service)

#### User Entity Changes

Add `passwordHash` field to `User` entity (`java/task-service/src/main/java/dev/kylebradshaw/task/entity/User.java`):

```java
@Column(name = "password_hash")
private String passwordHash; // nullable — Google-only users won't have one
```

#### New Entity: PasswordResetToken

```java
@Entity
@Table(name = "password_reset_tokens")
public class PasswordResetToken {
    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    private UUID id;

    @Column(nullable = false, unique = true)
    private String token;

    @ManyToOne
    @JoinColumn(name = "user_id", nullable = false)
    private User user;

    @Column(nullable = false)
    private Instant expiresAt; // 1 hour from creation
}
```

#### New Endpoints (AuthController)

All endpoints on the existing `AuthController` (`/api/auth/*`):

| Endpoint | Method | Request Body | Response | Description |
|---|---|---|---|---|
| `/api/auth/register` | POST | `{ email, password, name }` | `{ accessToken, refreshToken, user }` | Create user with BCrypt-hashed password, return JWT tokens |
| `/api/auth/login` | POST | `{ email, password }` | `{ accessToken, refreshToken, user }` | Validate credentials, return JWT tokens |
| `/api/auth/forgot-password` | POST | `{ email }` | `204 No Content` | Generate reset token, send email via Resend. Always returns 204 (don't reveal if email exists) |
| `/api/auth/reset-password` | POST | `{ token, password }` | `204 No Content` | Validate token, update password hash, delete token |

#### Dependencies

Add to `java/task-service/build.gradle`:
- `spring-boot-starter-mail` is **not** needed — using Resend HTTP API directly
- Resend Java SDK: `com.resend:resend-java:<latest>`
- BCrypt is already included in `spring-boot-starter-security` (`BCryptPasswordEncoder`)

#### Email Service

New `EmailService` class using the Resend SDK:
- Sends password reset emails with a link to `https://kylebradshaw.dev/java/tasks/reset-password?token=<token>`
- From address: uses Resend's shared domain (e.g., `onboarding@resend.dev` or similar)
- Environment variable: `RESEND_API_KEY`

#### Validation Rules

- Password: minimum 8 characters
- Email: must be valid format, unique in database
- Name: required, non-empty
- Reset token: expires after 1 hour, single-use (deleted after use)

#### Security Config

Update `SecurityConfig.java` to permit unauthenticated access to:
- `/api/auth/register`
- `/api/auth/login`
- `/api/auth/forgot-password`
- `/api/auth/reset-password`

(These join the existing `/api/auth/google` and `/api/auth/refresh` permits.)

### 5.2 Frontend

#### Login Page Updates (`frontend/src/components/java/TasksPageContent.tsx`)

The existing login view shows a Google sign-in button. Update to show:

1. **Email/password form** — email input, password input, "Sign in" button
2. **"Forgot password?"** link below the form
3. **"Create account"** link below the form
4. **Divider** — "or"
5. **Google sign-in button** (existing)

#### New Components

**`RegisterForm.tsx`** — Shown when user clicks "Create account":
- Fields: name, email, password, confirm password
- On success: stores tokens in localStorage, redirects to tasks page
- "Already have an account?" link back to login

**`ForgotPasswordForm.tsx`** — Shown when user clicks "Forgot password?":
- Field: email
- On submit: calls `/api/auth/forgot-password`, shows "Check your email" message
- "Back to sign in" link

#### New Route

**`/java/tasks/reset-password` page** (`frontend/src/app/java/tasks/reset-password/page.tsx`):
- Reads `token` from URL query params
- Fields: new password, confirm password
- On submit: calls `/api/auth/reset-password`
- On success: redirects to `/java/tasks` login with a "Password reset successful" message

#### Auth Flow

The existing `auth.ts` utilities and `AuthProvider` handle token storage and Apollo Client headers. The new password auth endpoints return the same JWT token format as Google OAuth, so no changes needed to the token handling — just add the new API calls for register/login/forgot/reset.

### 5.3 Testing

**Backend unit tests:**
- Registration: success, duplicate email, invalid password (too short)
- Login: success, wrong password, non-existent email
- Forgot password: sends email, doesn't reveal if email exists
- Reset password: success, expired token, invalid token

**Frontend:** Existing E2E tests are mocked and won't be affected. New mocked Playwright tests for the login/register/forgot-password forms are optional but recommended.

---

## 6. Google OAuth Fix

**Not a code change.** The "Access blocked: Authorization Error" comes from Google Cloud Console configuration.

To fix:
1. Go to Google Cloud Console → APIs & Services → OAuth consent screen
2. Ensure the app is set to "External" and published (not "Testing" mode with limited users)
3. Go to Credentials → OAuth 2.0 Client ID
4. Add `https://kylebradshaw.dev/java/tasks` to Authorized redirect URIs
5. Add `https://kylebradshaw.dev` to Authorized JavaScript origins

---

## 7. Repo Rename

The GitHub repo has already been renamed from `gen_ai_engineer` to `portfolio`. The following active files need their references updated:

### GHCR Image References (critical — builds and deploys will break)

All `ghcr.io/kabradshaw1/gen_ai_engineer/...` → `ghcr.io/kabradshaw1/portfolio/...`

| File | Current Image |
|---|---|
| `docker-compose.yml` (3 refs) | `ghcr.io/kabradshaw1/gen_ai_engineer/ingestion`, `/chat`, `/debug` |
| `k8s/ai-services/deployments/ingestion.yml` | `ghcr.io/kabradshaw1/gen_ai_engineer/ingestion` |
| `k8s/ai-services/deployments/chat.yml` | `ghcr.io/kabradshaw1/gen_ai_engineer/chat` |
| `k8s/ai-services/deployments/debug.yml` | `ghcr.io/kabradshaw1/gen_ai_engineer/debug` |
| `java/k8s/deployments/task-service.yml` | `ghcr.io/kabradshaw1/gen_ai_engineer/java-task-service` |
| `java/k8s/deployments/gateway-service.yml` | `ghcr.io/kabradshaw1/gen_ai_engineer/java-gateway-service` |
| `java/k8s/deployments/activity-service.yml` | `ghcr.io/kabradshaw1/gen_ai_engineer/java-activity-service` |
| `java/k8s/deployments/notification-service.yml` | `ghcr.io/kabradshaw1/gen_ai_engineer/java-notification-service` |

### Other Active Files

| File | What to change |
|---|---|
| `Makefile` | SSH path: `C:\Users\PC\repos\gen_ai_engineer` → `C:\Users\PC\repos\portfolio` |
| `k8s/setup-windows.ps1` | Default repo path suggestion |
| `CLAUDE.md` | Multiple references to paths and project structure |

### Local Directory Rename

Rename `/Users/kylebradshaw/repos/gen_ai_engineer` → `/Users/kylebradshaw/repos/portfolio`.

After renaming:
- Update the Windows PC clone path: `C:\Users\PC\repos\gen_ai_engineer` → `C:\Users\PC\repos\portfolio`
- Claude Code memory files at `~/.claude/projects/-Users-kylebradshaw-repos-gen-ai-engineer/` will need the project path updated (Claude Code creates a new project directory automatically when opened from the new path)

### GitHub Actions CI

The CI workflows (`.github/workflows/`) don't hardcode the repo name — they use `${{ github.repository }}` for GHCR paths. **No CI workflow changes needed** as long as GitHub's repo rename redirect is in place.

### Docs/Plans (not updating)

~20+ files in `docs/superpowers/specs/` and `docs/superpowers/plans/` reference the old name. These are historical records and will not be updated.

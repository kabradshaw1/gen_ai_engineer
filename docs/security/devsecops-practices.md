# DevSecOps & Security Practices

A factual inventory of the security and DevSecOps practices currently implemented in this portfolio project, written for resume and interview use. Each practice cites the exact file(s) that implement it so claims can be verified.

Gaps (things NOT implemented) are called out at the bottom — being honest about them is a feature, not a bug, because it shows scoping judgment.

---

## Shift-left security in CI

Six dedicated security jobs in `.github/workflows/ci.yml` gate every push and block deploy on failure.

| Practice | Tool | Evidence |
|---|---|---|
| Python SAST | Bandit `-ll` (high-confidence only) | `.github/workflows/ci.yml:461-489` (`security-bandit`) |
| Python dependency CVE scan | `pip-audit` per service | `.github/workflows/ci.yml:491-521` (`security-pip-audit`) |
| JS dependency CVE scan | `npm audit --audit-level=high` | `.github/workflows/ci.yml:522-542` (`security-npm-audit`) |
| Full-history secret scan | `gitleaks-action@v2` with custom allowlist | `.github/workflows/ci.yml:544-555`, `.gitleaks.toml` |
| Dockerfile lint | Hadolint across 9 Dockerfiles | `.github/workflows/ci.yml:557-578` |
| Homegrown CORS guardrail | grep for `allow_origins=["*"]` | `.github/workflows/ci.yml:580-594` |
| Java dependency tree (manual review) | OWASP Dependency Check (Gradle) | `.github/workflows/java-ci.yml:88-105` |
| Automated dependency PRs | Dependabot (pip/npm/Actions, weekly) | `.github/dependabot.yml` |

**Deploy gating:** the `deploy` job declares `needs:` on gitleaks + hadolint + all build jobs (`.github/workflows/ci.yml:630-643`), so a single security failure blocks production.

**Resume phrasing:** "Implemented a multi-language DevSecOps pipeline in GitHub Actions gating every deploy on SAST (Bandit), SCA (pip-audit, npm audit, OWASP Dependency Check), secret scanning (gitleaks), and container linting (Hadolint), with automated dependency updates via Dependabot."

---

## Infrastructure-as-Code security (policy-as-code)

The strongest recent story — built in response to real production incidents.

- **`kubeconform`** — static K8s OpenAPI schema validation across `k8s/`, `java/k8s/`, `go/k8s/` (`.github/workflows/ci.yml:352-404`).
- **`kind` cluster + server-side dry-run** — real API-server admission checks, not just schema validation. Catches CRD references, admission webhooks, and cross-resource validation.
- **Custom policy-as-code script** — `scripts/k8s-policy-check.sh` enforces two rules derived from real production incidents:
  - **R1:** any `Deployment` running `postgres`/`mongo`/`redis` must have a `readinessProbe`. Derived from a 2026-04-08 incident where `kubectl rollout status` returned before postgres bound its port, causing Go migration Jobs to race startup.
  - **R2:** any `ConfigMap` `DATABASE_URL` on `postgres://` must include `sslmode=disable`. Derived from the same incident, where the `lib/pq` driver's default `sslmode=require` failed against a non-SSL postgres.
  - Test harness at `scripts/test-k8s-policy-check.sh` covers 5 fixtures (both rules, positive and negative cases, plus a non-applicable case).

**Resume phrasing:** "Authored a policy-as-code lint layer for Kubernetes manifests (kubeconform + kind server-side dry-run + custom bash/yq policy rules derived from production incidents), catching an entire class of deployment failures at PR time instead of during rollout."

---

## Supply chain

- **Multi-stage builds** on every service (Python, Go, Java) — builder stage compiles, final image is slim/alpine with only artifacts. Evidence: `go/auth-service/Dockerfile:1-11`, `services/chat/Dockerfile`.
- **Non-root containers** — explicit `USER appuser` (uid 1001 on Go) across all 9 services. Evidence: `services/chat/Dockerfile:13-14`, `go/auth-service/Dockerfile:17,23`.
- **Private registry (GHCR)** with `imagePullSecrets` on every Deployment and GITHUB_TOKEN-based push auth.
- **Pinned tool versions** in CI (`kubeconform v0.6.7`, `golang-migrate v4.17.0`, `yq v4.44.3`, `ruff v0.11.6`, `Hadolint` via pinned action).

**Do NOT claim:** image signing, SLSA provenance, cosign, or SHA-pinned image digests — the project uses `:latest` tags.

---

## Secrets hygiene

- **Gitignored secrets:** `.env` and `**/k8s/secrets/*.yml` are gitignored (`.gitignore:18-23`); only `*.yml.template` files are committed.
- **Kubernetes-native secret injection:** K8s `Secret` resources are injected via `secretKeyRef` at the Deployment level — never baked into images or ConfigMaps. Evidence: `java/k8s/deployments/task-service.yml:27-46`, `go/k8s/deployments/auth-service.yml:28-42`.
- **GitHub Actions secrets** (`TAILSCALE_AUTHKEY`, `SSH_PRIVATE_KEY`, `SMOKE_GO_PASSWORD`, `GITHUB_TOKEN`) used for CI/deploy auth and smoke testing.
- **Full-history gitleaks scan** — runs with `fetch-depth: 0` so it scans the entire repository history, not just the latest commit.

---

## Application-layer AuthN/AuthZ

The Go and Java stacks both implement the same pattern independently, showing the pattern is the contribution rather than the framework.

- **JWT + bcrypt** — 15-minute access tokens + 7-day refresh tokens, HMAC-SHA256 signed, bcrypt `DefaultCost` password hashes. Evidence: `go/auth-service/internal/service/auth.go`, `java/task-service/src/main/java/.../SecurityConfig.java:71-73`.
- **Google OAuth 2.0** as a federated auth path with server-side code exchange; both auth methods return the same JWT envelope so the frontend is agnostic. Evidence: `go/auth-service/internal/handler/auth.go:74-91`, `docs/adr/password-authentication.md`.
- **Gateway-validated JWTs with `X-User-Id` forwarding** (Java side) — only the gateway and task-service validate tokens; downstream activity/notification services trust the forwarded header, reducing attack surface. Evidence: `java/gateway-service/src/main/java/.../SecurityConfig.java:33-51`, architecture diagram in `docs/adr/java-task-management/03_authentication_and_security.md`.
- **Strict CORS** — environment-driven allowlists, no wildcards. Runtime enforced in `go/auth-service/internal/middleware/cors.go:1-33` and `java/gateway-service/src/main/java/.../SecurityConfig.java:54-64`. Also enforced by the CI CORS guardrail so a wildcard can never be committed.
- **Security headers** — Spring Security sets `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, and HSTS `max-age=31536000; includeSubDomains`. Evidence: `java/gateway-service/src/main/java/.../SecurityConfig.java:38-43`.
- **Stateless sessions** (`SessionCreationPolicy.STATELESS`) — no server-side session state, all auth via JWT.

**Resume phrasing:** "Designed and implemented JWT + bcrypt auth with short-lived access/refresh token separation and Google OAuth 2.0 federation across independent Go and Java microservice stacks. Implemented a gateway-validated-JWT pattern with trusted `X-User-Id` header forwarding to reduce auth surface on downstream services."

---

## Transport security

- **Cloudflare Tunnel** — backend is never directly internet-exposed. TLS terminates at the Cloudflare edge; the tunnel forwards to the home-lab Windows PC. Evidence: `CLAUDE.md:36-40`.
- **NGINX Ingress** with path-based routing inside Minikube.
- **Frontend on Vercel with HTTPS**; Cloudflare-managed certs for the API subdomain.

---

## Developer-side guardrails

- **Pre-commit hooks** — ruff, checkstyle, `tsc`, ESLint, golangci-lint, plus a `pre-push` stage that runs `next build`. Evidence: `.pre-commit-config.yaml`.
- **`make preflight*` targets** mirror CI locally so issues are caught before pushing. Evidence: `Makefile`.
- **Feature-branch workflow** enforced by convention: `feature → staging → main`, with distinct CI triggers per branch and a deploy-only-on-main-push guard. Staging runs mocked Playwright E2E tests; main runs deploy + production smoke tests.

---

## Post-deploy and pre-deploy verification

- **Production smoke tests** via Playwright against `https://api.kylebradshaw.dev` after every deploy (`.github/workflows/ci.yml:716-749`). Tests: frontend loads, health checks, document upload, chat query, collection cleanup.
- **Compose-smoke CI job** — stands up the full Python stack via `docker-compose.ci.yml` with a mocked Ollama stub (`services/mock-ollama/`) and runs a RAG happy-path Playwright test (`frontend/e2e/smoke-ci.spec.ts`). Catches contract drift before staging merge.
- **Go migration pipeline test** — runs the real `golang-migrate` auth + ecommerce migrations plus seed SQL against a postgres service container on every push. Reproduces the exact failure modes from the 2026-04-08 incident (sslmode, Job ordering, cross-service schema refs) in CI.
- **Readiness probes** on every stateful service (postgres, mongo, redis) — made un-skippable by the policy-as-code rule R1.

---

## Observability (foundation only — NOT security monitoring)

- Prometheus scraping every service + host + GPU (`k8s/monitoring/configmaps/prometheus-config.yml`).
- Grafana dashboard at `https://grafana.kylebradshaw.dev` (public read-only).
- Health endpoints on every service used by readiness probes and production smoke tests.

**Do NOT claim:** alerting on auth failures, anomaly detection, SIEM integration, log aggregation, or audit logging. Current framing should be "observability foundation in place" rather than "security monitoring."

---

## Explicit gaps (future work)

Listing these demonstrates scoping judgment — a portfolio project doesn't need enterprise controls, but being explicit about where the line is drawn is itself a DevSecOps signal.

- **No Kubernetes `NetworkPolicy`** — pods can talk to anything. Fine for a portfolio; production would enforce pod-to-pod restrictions.
- **No `PodSecurityContext` / `readOnlyRootFilesystem` / `runAsNonRoot: true`** at the k8s level. Containers run as non-root via Dockerfile `USER` directive, but Kubernetes isn't enforcing it — a future PSS (`restricted` profile) would.
- **No image signing** (cosign / SLSA provenance / Sigstore). All images are `:latest` tag, not SHA-pinned.
- **No IaC scanning** (checkov, tfsec, kube-linter, kubesec). A targeted homegrown policy script fills part of this gap but doesn't cover the breadth of a dedicated scanner.
- **No runtime security** (Falco, Tetragon, eBPF-based detection).
- **No WAF rules beyond Cloudflare defaults.**
- **Java OWASP Dependency Check** only generates a report; it does not fail the build.
- **No centralized log aggregation** (no Loki, ELK, or equivalent).
- **No audit logging** of auth events, privileged actions, or config changes.

---

## One-line resume summary

> Designed and operated a full DevSecOps pipeline for a polyglot (Python/Go/Java/TypeScript) microservice portfolio: shift-left SAST/SCA/secret-scanning gates, Dockerfile and Kubernetes policy-as-code linting with custom rules derived from production incidents, multi-stage non-root container images, strict CORS + HSTS enforcement, JWT + bcrypt + Google OAuth 2.0 with gateway-validated tokens, and automated pre- and post-deploy smoke verification — all gating a GitHub Actions → GHCR → Minikube → Cloudflare Tunnel delivery pipeline.

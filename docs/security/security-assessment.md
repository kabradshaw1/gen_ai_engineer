# Security Assessment

An evaluation of the security and DevSecOps controls currently implemented in this portfolio project. Each finding cites the exact file(s) that implement the control so every claim can be independently verified.

**Scope:** the entire repository — Python AI services, Go and Java microservices, Next.js frontend, Kubernetes manifests, CI/CD workflows, and supporting scripts.

**Methodology:** source-code review of `.github/workflows/`, Dockerfiles, Kubernetes manifests, application auth code, and operational scripts. No dynamic testing was performed.

**Date:** 2026-04-08

## Summary of findings

| Area | Status | Notes |
|---|---|---|
| Shift-left security in CI | **Strong** | Six gating jobs covering SAST, SCA, secret scanning, Dockerfile lint, and a custom CORS guardrail |
| Infrastructure-as-Code validation | **Strong** | `kubeconform`, `kind` server-side dry-run, and a custom policy-as-code script |
| Supply chain | **Adequate for portfolio** | Multi-stage non-root builds and a private registry; no image signing |
| Secrets management | **Strong** | Gitignored files, full-history secret scanning, and k8s-native `secretKeyRef` injection |
| Application AuthN/AuthZ | **Strong** | JWT + bcrypt + Google OAuth 2.0, gateway-validated tokens with trusted header forwarding |
| Transport security | **Adequate** | TLS terminated at Cloudflare; no direct internet exposure of the backend |
| Developer guardrails | **Strong** | Pre-commit hooks, preflight Makefile targets, and a structured branch workflow |
| Post-deploy verification | **Strong** | Production Playwright smoke tests plus a new pre-deploy compose-smoke job |
| Kubernetes runtime posture | **Weak** | No `NetworkPolicy`, `PodSecurityContext`, or `readOnlyRootFilesystem` — see accepted risks |
| Observability | **Foundation only** | Metrics and dashboards exist; no security monitoring or alerting |

The overall posture is strong for a portfolio-scale project. The most notable gap is Kubernetes-level runtime security (NetworkPolicy, PSS), which is documented as an accepted risk below.

---

## 1. Shift-left security in CI

**Status:** Strong.

Six dedicated security jobs in `.github/workflows/ci.yml` gate every push and block deployment on failure.

| Control | Tool | Evidence |
|---|---|---|
| Python SAST | Bandit `-ll` (high-confidence only) | `.github/workflows/ci.yml:461-489` (`security-bandit`) |
| Python dependency CVE scan | `pip-audit` per service | `.github/workflows/ci.yml:491-521` (`security-pip-audit`) |
| JavaScript dependency CVE scan | `npm audit --audit-level=high` | `.github/workflows/ci.yml:522-542` (`security-npm-audit`) |
| Full-history secret scan | `gitleaks-action@v2` with custom allowlist | `.github/workflows/ci.yml:544-555`, `.gitleaks.toml` |
| Dockerfile lint | Hadolint across 9 Dockerfiles | `.github/workflows/ci.yml:557-578` |
| CORS guardrail | `grep` for `allow_origins=["*"]` | `.github/workflows/ci.yml:580-594` |
| Java dependency tree (manual review) | OWASP Dependency Check (Gradle) | `.github/workflows/java-ci.yml:88-105` |
| Automated dependency PRs | Dependabot (pip / npm / Actions, weekly) | `.github/dependabot.yml` |

**Effectiveness:** the `deploy` job declares `needs:` on gitleaks, Hadolint, and all build jobs (`.github/workflows/ci.yml:630-643`), so a single security failure blocks production deployment. Dependabot feeds into the same pipeline, so proposed dependency updates are subject to the same gates.

**Limitations:** the Java OWASP Dependency Check currently generates a report but does not fail the build; it is effectively an informational control.

---

## 2. Infrastructure-as-Code validation

**Status:** Strong.

This layer was built in response to a real production incident on 2026-04-08, where Kubernetes manifests without a `readinessProbe` on postgres and without `sslmode=disable` on the Go services' `DATABASE_URL` caused cascading migration failures.

- **`kubeconform`** — static Kubernetes OpenAPI schema validation across `k8s/`, `java/k8s/`, and `go/k8s/`. Evidence: `.github/workflows/ci.yml:352-404`.
- **`kind` cluster with server-side `--dry-run=server`** — real API-server admission validation, not only schema validation. Catches CRD references, admission webhooks, and cross-resource validation.
- **Custom policy-as-code script** — `scripts/k8s-policy-check.sh` enforces two rules derived from the 2026-04-08 incident:
  - **R1:** any `Deployment` whose container image is `postgres`, `mongo`, or `redis` must define a `readinessProbe`. Addresses a class of bug where `kubectl rollout status` returns before the database binds its port, allowing downstream Jobs to race startup.
  - **R2:** any `ConfigMap` data key ending in `DATABASE_URL` and starting with `postgres://` must include `sslmode=disable`. Addresses the `lib/pq` driver's default of `sslmode=require`, which fails against a non-SSL postgres.
- **Test harness** at `scripts/test-k8s-policy-check.sh` covers five fixtures, including both positive and negative cases for each rule plus a non-applicable case to prevent overreach.

**Effectiveness:** the policy script has already surfaced real regressions — missing probes on `mongodb` and `redis` Deployments — during the implementation of this assessment, forcing them to be remediated before merge.

---

## 3. Supply chain

**Status:** Adequate for a portfolio project.

- **Multi-stage Docker builds** on every service. The builder stage compiles code; the final image contains only artifacts on a slim or Alpine base, reducing attack surface. Evidence: `go/auth-service/Dockerfile:1-11`, `services/chat/Dockerfile`.
- **Non-root containers** — explicit `USER appuser` (uid 1001 on Go) across all 9 services. Evidence: `services/chat/Dockerfile:13-14`, `go/auth-service/Dockerfile:17,23`.
- **Private registry (GHCR)** with `imagePullSecrets` on every Deployment and `GITHUB_TOKEN`-based push authentication.
- **Pinned tool versions** in CI for reproducibility: `kubeconform v0.6.7`, `golang-migrate v4.17.0`, `yq v4.44.3`, `ruff v0.11.6`, and Hadolint via a pinned action SHA.

**Accepted risks:**
- Images are tagged `:latest` rather than pinned to commit SHAs or digests, which means the tag is mutable. This is acceptable for portfolio use but would not be acceptable in production.
- No image signing (Cosign / SLSA provenance) or attestation verification. The registry is private and push is gated by the CI workflow, which is a partial mitigation.

---

## 4. Secrets management

**Status:** Strong.

- **Gitignored secrets:** both `.env` and `**/k8s/secrets/*.yml` are gitignored (`.gitignore:18-23`). Only `*.yml.template` files are committed, providing a public contract for required secret fields without leaking values.
- **Kubernetes-native secret injection:** all production secrets are stored as `Secret` resources and injected at the Deployment level via `valueFrom.secretKeyRef`. They are never baked into images or ConfigMaps. Evidence: `java/k8s/deployments/task-service.yml:27-46`, `go/k8s/deployments/auth-service.yml:28-42`.
- **GitHub Actions secrets** (`TAILSCALE_AUTHKEY`, `SSH_PRIVATE_KEY`, `SMOKE_GO_PASSWORD`, `GITHUB_TOKEN`) are the only place CI-level credentials are referenced.
- **Full-history gitleaks scan** runs with `fetch-depth: 0`, so the entire repository history is scanned, not only the latest commit.

**Observations:**
- Secret rotation is manual. The `TAILSCALE_AUTHKEY` is documented as requiring rotation every 90 days (`CLAUDE.md:181`), but no automation enforces this.
- There is no envelope encryption (SOPS / sealed-secrets / external-secrets-operator). Kubernetes Secrets are therefore only as well protected as the cluster's etcd backing store, which in a Minikube deployment is not encrypted at rest.

---

## 5. Application AuthN/AuthZ

**Status:** Strong. The Go and Java stacks independently implement the same patterns, which demonstrates the patterns rather than the framework choices.

- **JWT + bcrypt.** 15-minute access tokens and 7-day refresh tokens, HMAC-SHA256 signed, with bcrypt `DefaultCost` password hashes. Evidence: `go/auth-service/internal/service/auth.go`, `java/task-service/src/main/java/.../SecurityConfig.java:71-73`.
- **Google OAuth 2.0.** Federated authentication with server-side code exchange. Both password and OAuth flows return the same JWT envelope so the frontend is agnostic. Evidence: `go/auth-service/internal/handler/auth.go:74-91`, `docs/adr/password-authentication.md`.
- **Gateway-validated JWTs with `X-User-Id` header forwarding** on the Java side. Only the gateway and `task-service` validate tokens; `activity-service` and `notification-service` trust the forwarded header. This reduces the auth attack surface but introduces a trust boundary that must be enforced at the network layer (see accepted risks). Evidence: `java/gateway-service/src/main/java/.../SecurityConfig.java:33-51`, `docs/adr/java-task-management/03_authentication_and_security.md`.
- **Strict CORS.** Environment-driven allowlists with no wildcards. Runtime-enforced in `go/auth-service/internal/middleware/cors.go:1-33` and `java/gateway-service/src/main/java/.../SecurityConfig.java:54-64`. Also enforced at CI time via the CORS guardrail (§1), so a wildcard cannot be committed.
- **Security headers.** Spring Security sets `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, and HSTS `max-age=31536000; includeSubDomains`. Evidence: `java/gateway-service/src/main/java/.../SecurityConfig.java:38-43`.
- **Stateless sessions** (`SessionCreationPolicy.STATELESS`). No server-side session state; all authorization is via JWT.

**Accepted risk:** the gateway/downstream trust model relies on `activity-service` and `notification-service` only being reachable from inside the cluster. With no `NetworkPolicy` in place (§9), this boundary is not enforced at the network layer — a compromised pod in another namespace could in principle forge `X-User-Id`. Acceptable for portfolio, not acceptable for production.

---

## 6. Transport security

**Status:** Adequate.

- **Cloudflare Tunnel** — the backend is never directly exposed to the public internet. TLS terminates at the Cloudflare edge; the tunnel forwards traffic to the home-lab Windows PC running Minikube. Evidence: `CLAUDE.md:36-40`.
- **NGINX Ingress** with path-based routing inside Minikube.
- **Frontend on Vercel** with managed HTTPS; Cloudflare-managed certificates for the API subdomain.

**Observations:** intra-cluster traffic between pods is unencrypted. For a home-lab deployment this is acceptable; production workloads with sensitive data would require a service mesh (e.g. Linkerd, Istio) for mTLS.

---

## 7. Developer guardrails

**Status:** Strong.

- **Pre-commit hooks** for ruff, Checkstyle, `tsc`, ESLint, and golangci-lint, plus a `pre-push` stage that runs `next build`. Evidence: `.pre-commit-config.yaml`.
- **`make preflight*` targets** mirror CI locally so issues are caught before pushing. Evidence: `Makefile`.
- **Structured branch workflow** — `feature → staging → main`, with distinct CI triggers per branch and a deploy-only-on-`main`-push guard. Staging runs mocked Playwright E2E tests; `main` runs deploy and production smoke tests.

**Observations:** the branch workflow is enforced by convention rather than branch protection rules. A sufficiently motivated contributor (or an errant script) could push directly to `main`. For a single-maintainer portfolio this is acceptable.

---

## 8. Pre-deploy and post-deploy verification

**Status:** Strong.

- **Production smoke tests** — Playwright suite against `https://api.kylebradshaw.dev` after every deploy. Tests cover frontend load, health checks, document upload, chat query, and cleanup. Evidence: `.github/workflows/ci.yml:716-749`.
- **Compose smoke CI job** — stands up the full Python stack via `docker-compose.ci.yml` with a mocked Ollama stub (`services/mock-ollama/`) and runs a RAG happy-path Playwright test (`frontend/e2e/smoke-ci.spec.ts`). Catches contract drift before a staging merge.
- **Go migration pipeline test** — runs the real `golang-migrate` auth and ecommerce migrations plus seed SQL against a postgres service container on every push. Reproduces the 2026-04-08 incident failure modes (sslmode, Job ordering, cross-service schema references) in CI.
- **Readiness probes** on every stateful service (postgres, mongo, redis) — made non-skippable by the policy-as-code rule R1 in §2.

---

## 9. Kubernetes runtime posture

**Status:** Weak. These are the most significant gaps in the current assessment.

- **No `NetworkPolicy`.** Pods across all namespaces can reach each other without restriction. The gateway/downstream auth trust model (§5) depends on implicit network isolation that does not exist.
- **No `PodSecurityContext` at the Kubernetes level.** Containers run as non-root via the Dockerfile `USER` directive, but Kubernetes itself is not enforcing `runAsNonRoot: true`, `runAsUser`, `fsGroup`, or `seccompProfile`.
- **No `readOnlyRootFilesystem: true`.** Containers can write anywhere on their root filesystem at runtime.
- **No Pod Security Standards (PSS) profile** — neither `baseline` nor `restricted` is enforced at the namespace level.

**Accepted risk:** all four gaps are acknowledged as future work. A single commit adding PSS `restricted` profiles plus targeted `NetworkPolicy` resources would close the majority of this surface. See "Recommended next steps" below.

---

## 10. Observability

**Status:** Foundation only. Not a security monitoring control in its current form.

- **Prometheus** scrapes every service, the host, and the GPU (`k8s/monitoring/configmaps/prometheus-config.yml`).
- **Grafana dashboard** at `https://grafana.kylebradshaw.dev` (public read-only).
- **Health endpoints** on every service, used by readiness probes and production smoke tests.

**Gaps:** there is no alerting on authentication failures, no anomaly detection, no SIEM integration, no centralized log aggregation (Loki / ELK), and no audit logging of privileged actions. For security purposes the current stack is observational rather than detective.

---

## Recommended next steps

Ordered by impact-to-effort ratio:

1. **Add a PSS `restricted` label** to the `ai-services`, `java-tasks`, and `go-ecommerce` namespaces and fix any resulting violations. High impact, low effort.
2. **Add minimal `NetworkPolicy` resources** — default-deny ingress in each namespace, plus explicit allow rules for gateway→downstream and ingress-controller→gateway. Closes the §5 trust-model gap.
3. **Pin production container images by digest** (`@sha256:…`) instead of `:latest`. Eliminates the mutable-tag supply chain risk.
4. **Promote Java OWASP Dependency Check from reporting to gating.** One-line CI change.
5. **Add Prometheus alert rules** for 4xx/5xx rate spikes and authentication failure bursts. Converts observability into detection.
6. **Run a dedicated IaC scanner** (e.g. `kube-linter`, `checkov`) alongside the existing custom policy script. Catches the broader class of rules the homegrown script was intentionally narrow about.
7. **Migrate to envelope-encrypted secrets** (SOPS with age, or external-secrets-operator with a cloud KMS) to remove dependence on Kubernetes Secrets' at-rest model.

---

## Evidence index

Every file path cited in this assessment, for quick navigation:

- `.github/workflows/ci.yml` — primary CI/CD workflow (Python, frontend, security, k8s validation, deploy, smoke)
- `.github/workflows/java-ci.yml` — Java CI workflow
- `.github/workflows/go-ci.yml` — Go CI workflow
- `.github/dependabot.yml` — automated dependency updates
- `.gitleaks.toml` — secret scanner allowlist
- `.gitignore` — secret and environment file exclusions
- `.pre-commit-config.yaml` — pre-commit hooks
- `Makefile` — preflight targets
- `scripts/k8s-policy-check.sh` — custom policy-as-code
- `scripts/test-k8s-policy-check.sh` — policy script test harness
- `services/mock-ollama/` — CI-only Ollama stub for compose smoke tests
- `docker-compose.ci.yml` — compose overlay for CI smoke runs
- `frontend/e2e/smoke.spec.ts` — post-deploy production smoke tests
- `frontend/e2e/smoke-ci.spec.ts` — pre-deploy compose smoke tests
- `go/auth-service/internal/service/auth.go` — JWT and bcrypt implementation
- `go/auth-service/internal/middleware/cors.go` — Go CORS enforcement
- `java/gateway-service/src/main/java/.../SecurityConfig.java` — Spring Security configuration
- `java/task-service/src/main/java/.../SecurityConfig.java` — Spring Security configuration
- `java/k8s/deployments/postgres.yml`, `mongodb.yml`, `redis.yml` — stateful service manifests with readiness probes
- `go/k8s/configmaps/*.yml` — Go service ConfigMaps with `sslmode=disable`
- `docs/adr/password-authentication.md` — auth architecture ADR
- `docs/adr/java-task-management/03_authentication_and_security.md` — Java auth ADR

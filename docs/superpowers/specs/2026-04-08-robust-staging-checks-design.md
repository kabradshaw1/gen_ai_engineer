# Robust Staging Checks — Design

**Status:** Approved, ready for implementation plan
**Date:** 2026-04-08
**Branch:** `robust-staging-checks`

## Problem

Recent main-branch deploys have failed on issues that staging's current checks couldn't catch, because staging only runs mocked Playwright E2E tests — nothing that exercises real infrastructure. Examples from this week:

- `pq: SSL is not enabled on the server` — Go migration Jobs using `lib/pq`'s default `sslmode=require` against a non-SSL postgres. Never surfaced until migration Jobs ran on the real PC.
- `relation "users" does not exist` — Go migration Jobs ran in parallel with no ordering, so ecommerce seed tried to insert into a table that auth-service's migrations hadn't created yet.
- Missing postgres `readinessProbe` — caused `kubectl rollout status` to return before postgres actually bound port 5432, making the Go migration Jobs race postgres startup.
- PVC data wipe on first deploy of the new postgres volume mount — not preventable but would have been visible earlier with better pre-deploy checks.

The shared root cause: **staging has no job that runs the real deploy artifacts against real dependencies.** Every infra regression waits until main.

## Goal

Catch infra/integration failures on staging (or earlier) so the main-branch deploy action stays green. Keep the solution hermetic — no changes to the Windows PC, no parallel Minikube namespaces, no new Cloudflare tunnels.

## Non-goals

- A real parallel staging environment on Minikube (rejected: higher fidelity but much more complexity for a portfolio project).
- Running Java integration tests in CI (stays on-demand via `preflight-java-integration`).
- Running real Ollama in CI (too slow, too flaky — mock instead).
- Kustomize refactor of the existing flat k8s manifest layout.

## Approach

Add three new path-filtered jobs to `.github/workflows/ci.yml`, each gated on the same `needs.changes` outputs the existing pipeline already uses. All three run entirely inside GitHub Actions runners.

### Job 1 — `go-migration-test`

**Trigger:** `needs.changes.outputs.go == 'true' || needs.changes.outputs.k8s == 'true'`

**Steps:**
1. Launch `postgres:17-alpine` as a GitHub Actions service container (same tag as prod — mirrors the no-SSL configuration).
2. Create the `ecommercedb` database in an init step (mirrors what the `postgres-initdb` ConfigMap does on first boot of a fresh PVC).
3. Install `golang-migrate` `v4.17.0` (same version baked into the service Dockerfiles).
4. Run `migrate -path go/auth-service/migrations -database "$DATABASE_URL" up` and assert exit 0.
5. Run `migrate -path go/ecommerce-service/migrations -database "$DATABASE_URL" up` and assert exit 0.
6. Run `psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f go/ecommerce-service/seed.sql` and assert exit 0.
7. `DATABASE_URL` is set with `?sslmode=disable` explicitly to match prod exactly.

**Failure modes caught:** sslmode regressions, migration ordering bugs (auth must run before ecommerce), cross-service schema refs (ecommerce seed references auth's `users` table), SQL syntax errors, seed idempotency regressions, drift between seed column names and migration schemas.

**Expected runtime:** ~60s.

### Job 2 — `k8s-manifest-validation`

**Trigger:** `needs.changes.outputs.k8s == 'true'`

Three stages, fail-fast ordered so cheap checks run first.

**Stage A — `kubeconform` (static schema validation)**
- Install `kubeconform` via release binary.
- Run `kubeconform -strict -summary -schema-location default` against `k8s/`, `java/k8s/`, `go/k8s/`.
- Catches: field typos, invalid enum values, missing required fields, wrong `apiVersion`/`kind` combinations.
- Runtime: <5s.

**Stage B — `kind` cluster server-side dry-run**
- Stand up a `kind` cluster via `helm/kind-action@v1` (cached image, ~20s startup).
- Create namespaces: `ai-services`, `java-tasks`, `go-ecommerce`, `monitoring`.
- `kubectl apply --dry-run=server -f` each manifest directory recursively.
- Catches: admission webhook errors, cross-resource reference errors, immutability violations, invalid probe configs, volumeMount combinations the real API server rejects.
- Runtime: ~45s.

**Stage C — homegrown policy check**
- `scripts/k8s-policy-check.sh` using bash + `yq`, enforcing portfolio-specific rules:
  - Every `Deployment` whose container image references `postgres`, `mongo`, or `redis` **must** define a `readinessProbe` (would have caught this week's postgres regression).
  - Every `ConfigMap` data key ending in `DATABASE_URL` containing `postgres://` **must** include `sslmode=disable` (belt-and-suspenders for the sslmode regression).
- Exit 1 with a clear error on any violation.
- Runtime: <5s.
- Intentionally extensible — add rules reactively as new failure modes are discovered.

**Total expected runtime:** ~60s on k8s changes only.

### Job 3 — `compose-smoke`

**Trigger:** `needs.changes.outputs.python == 'true' || needs.changes.outputs.frontend == 'true'`

Runs the Python AI stack end-to-end via `docker compose` with a mocked Ollama, and exercises the RAG happy-path with a headless Playwright smoke test.

**New assets:**
- `services/mock-ollama/` — tiny FastAPI app, ~30 LOC, with two endpoints:
  - `POST /api/embeddings` — returns a fixed 768-dimension vector (matches `nomic-embed-text` dims so Qdrant collection schema stays valid).
  - `POST /api/chat` — returns a canned streaming response matching Ollama's chat streaming format, so the chat service's streaming parser is exercised correctly.
  - Dockerfile alongside.
- `docker-compose.ci.yml` — overlay on top of the main `docker-compose.yml`:
  - Adds the `mock-ollama` service.
  - Overrides `OLLAMA_BASE_URL` on `ingestion`, `chat`, `debug` services to `http://mock-ollama:11434`.
  - Keeps everything else (Qdrant, nginx, service ports) identical to the real compose file.
- `frontend/e2e/smoke-ci.spec.ts` — Playwright test that runs the RAG happy path: upload a small PDF → ask a question → assert a streamed response arrives through nginx at `localhost:8000`. Can be a sibling of the existing `smoke.spec.ts` with a separate Playwright config targeting `localhost:8000` instead of `api.kylebradshaw.dev`.

**Job steps:**
1. `docker compose -f docker-compose.yml -f docker-compose.ci.yml up -d --build`.
2. Poll `/health` on ingestion/chat/debug through nginx until all return 200 (or timeout).
3. Run `npx playwright test smoke-ci.spec.ts`.
4. On failure: dump `docker compose logs` for all services for debugging.
5. `docker compose down -v` in an `always()` step.

**Failure modes caught:** Dockerfile regressions, docker-compose wiring errors, service startup failures, env var propagation bugs, nginx routing regressions, CORS regressions, Qdrant schema/dimensionality drift, chunking regressions, streaming response parsing bugs, frontend↔backend contract drift on chat/ingestion endpoints.

**Expected runtime:** ~2 minutes.

## New files

- `services/mock-ollama/main.py` — FastAPI stub.
- `services/mock-ollama/Dockerfile` — minimal Python image.
- `services/mock-ollama/requirements.txt` — just `fastapi` + `uvicorn`.
- `docker-compose.ci.yml` — CI overlay.
- `scripts/k8s-policy-check.sh` — probe + sslmode policy script.
- `frontend/e2e/smoke-ci.spec.ts` — RAG happy-path smoke against localhost.
- `frontend/playwright.smoke-ci.config.ts` — Playwright config targeting `localhost:8000`.

## Modified files

- `.github/workflows/ci.yml` — three new jobs, wired into the existing `changes` outputs, running in parallel with existing lint/test jobs. None of them gate the `deploy` job (deploy still only runs on main).

## Out of scope for this spec

- Making any of the new jobs block merges (we can add branch protection rules later once they're proven stable).
- Adding similar real-integration checks for Java services — that stays on-demand via `preflight-java-integration`.
- Publishing the mock-ollama image to GHCR — it lives only in the repo and is built fresh in CI.

## Risks

- **docker-compose drift:** the compose file must stay in sync with what k8s actually runs. This is already true today for local dev, so the drift risk isn't new — but Job 3 makes the consequences of drift more visible. Mitigation: a CLAUDE.md note reminding both Kyle and Claude that changes to Python service configuration must be reflected in both compose and k8s manifests.
- **mock-ollama realism:** the stub may diverge from real Ollama's response format over time. Mitigation: pin the exact response schemas in a short comment in `services/mock-ollama/main.py`, and re-check after any Ollama version bump.
- **CI runtime creep:** adding ~3 minutes of jobs to every relevant push. Acceptable per user direction ("plenty of CI minutes, want depth").

## Open questions

None — all design questions resolved during brainstorming.

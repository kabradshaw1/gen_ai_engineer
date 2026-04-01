# DevSecOps Security Hardening — Design Spec

## Goal

Harden the production deployment and CI pipeline against common security vulnerabilities while demonstrating DevSecOps best practices for a portfolio project. Both genuinely improve security and showcase tooling knowledge.

## Scope

- CORS configuration hardening (both backend services)
- Dockerfile hardening (both backend services)
- Automated security scanning in CI (5 new jobs + 1 custom guardrail)
- Local secrets detection config
- No changes to runtime dependencies, lesson notebooks, or frontend code

## 1. CORS Hardening

### Problem

Both `services/chat/app/main.py` and `services/ingestion/app/main.py` use `allow_origins=["*"]`, allowing any origin to make cross-origin requests to the API. This is shipped to production.

### Solution

Make allowed origins configurable via environment variable.

**Config changes** (`services/chat/app/config.py` and `services/ingestion/app/config.py`):
- Add `allowed_origins: str = "https://kylebradshaw.dev"` to the `Settings` class
- Value is a comma-separated string, parsed into a list at usage

**Middleware changes** (`services/chat/app/main.py` and `services/ingestion/app/main.py`):
- Replace `allow_origins=["*"]` with `allow_origins=settings.allowed_origins.split(",")`
- `allow_methods=["*"]` and `allow_headers=["*"]` remain unchanged (low-risk, required for API function)

**Environment config:**
- `.env.example` updated with `ALLOWED_ORIGINS=https://kylebradshaw.dev`
- Local dev override: `ALLOWED_ORIGINS=https://kylebradshaw.dev,http://localhost:3000`

## 2. Dockerfile Hardening

### Problem

Both Dockerfiles (`services/chat/Dockerfile`, `services/ingestion/Dockerfile`) run as root with an unpinned base image tag.

### Solution

For both Dockerfiles:
- Pin Python base image to specific minor version (`python:3.11-slim` is acceptable — full digest pinning is overkill for this project)
- Add `PYTHONDONTWRITEBYTECODE=1` and `PYTHONUNBUFFERED=1` environment variables
- Create a non-root user (`appuser`) and switch to it before `CMD`

## 3. CI Security Jobs

Five new jobs added to `.github/workflows/ci.yml`. All run in parallel with existing jobs. All must pass before `deploy` runs.

### 3.1 security-bandit — Python SAST

- Installs and runs Bandit against `services/`
- Flags: `-r services/ -f json -ll` (medium+ severity)
- Catches: hardcoded passwords, unsafe deserialization, use of `eval()`, weak crypto, etc.

### 3.2 security-pip-audit — Python Dependency CVE Scanning

- Matrix job: runs for both `ingestion` and `chat` services
- Installs service dependencies, then runs `pip-audit`
- Checks installed packages against the OSV vulnerability database
- Fails on any known CVE

### 3.3 security-npm-audit — Node Dependency CVE Scanning

- Runs `npm ci` in `frontend/`
- Runs `npm audit --audit-level=high`
- Only fails on high+ severity (avoids noise from low/moderate transitive dependency findings)

### 3.4 security-gitleaks — Secrets Detection

- Uses `gitleaks/gitleaks-action@v2` GitHub Action
- Scans codebase and git history for API keys, tokens, passwords, private keys
- Uses `.gitleaks.toml` config from repo root
- Fails if any secrets are detected

### 3.5 security-hadolint — Dockerfile Linting

- Uses `hadolint/hadolint-action@v3` GitHub Action
- Runs on `services/chat/Dockerfile` and `services/ingestion/Dockerfile`
- Enforces: non-root user, pinned base images, `--no-cache-dir` on pip, etc.

### 3.6 security-cors-check — Custom CORS Guardrail

- Simple grep-based check: fails if `allow_origins=["*"]` appears in any Python file under `services/`
- Project-specific safety net — cheap and effective

### Deploy Gate

The `deploy` job's `needs` list is updated to include all security jobs, ensuring nothing deploys without passing security checks.

## 4. Gitleaks Local Config

- `.gitleaks.toml` at repo root with sensible defaults
- Used by both the CI gitleaks job and available for local pre-commit use
- Defense-in-depth: catch secrets before they reach CI

## Files Changed

| File | Change |
|------|--------|
| `services/chat/app/config.py` | Add `allowed_origins` setting |
| `services/ingestion/app/config.py` | Add `allowed_origins` setting |
| `services/chat/app/main.py` | Use `settings.allowed_origins` for CORS |
| `services/ingestion/app/main.py` | Use `settings.allowed_origins` for CORS |
| `services/chat/Dockerfile` | Non-root user, env vars |
| `services/ingestion/Dockerfile` | Non-root user, env vars |
| `.github/workflows/ci.yml` | 6 new security jobs, updated deploy gate |
| `.env.example` | Add `ALLOWED_ORIGINS` |
| `.gitleaks.toml` | New file — gitleaks configuration |

## Files NOT Changed

- Lesson notebooks — these are educational, `allow_origins=["*"]` in examples is fine
- Frontend code — no changes needed
- Runtime Python dependencies — all security tools are CI-only

# DevSecOps Security Hardening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Harden CORS, Dockerfiles, and CI pipeline with automated security scanning.

**Architecture:** Fix CORS wildcard by making origins configurable via env vars. Harden Dockerfiles with non-root user. Add 6 security scanning jobs to CI that gate deployment.

**Tech Stack:** FastAPI/Pydantic (CORS config), Docker (container hardening), GitHub Actions (CI jobs), Bandit, pip-audit, npm audit, gitleaks, Hadolint.

---

### Task 1: CORS Hardening — Ingestion Service

**Files:**
- Modify: `services/ingestion/app/config.py:4-12`
- Modify: `services/ingestion/app/main.py:17-22`
- Modify: `services/ingestion/tests/test_main.py`

- [ ] **Step 1: Add `allowed_origins` to ingestion config**

In `services/ingestion/app/config.py`, add the `allowed_origins` field to the `Settings` class:

```python
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    ollama_base_url: str = "http://host.docker.internal:11434"
    embedding_model: str = "nomic-embed-text"
    qdrant_host: str = "qdrant"
    qdrant_port: int = 6333
    collection_name: str = "documents"
    chunk_size: int = 1000
    chunk_overlap: int = 200
    max_file_size_mb: int = 50
    allowed_origins: str = "https://kylebradshaw.dev"


settings = Settings()
```

- [ ] **Step 2: Update ingestion CORS middleware to use config**

In `services/ingestion/app/main.py`, replace the hardcoded wildcard CORS:

```python
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.allowed_origins.split(","),
    allow_methods=["*"],
    allow_headers=["*"],
)
```

- [ ] **Step 3: Add CORS test to ingestion**

Add to the end of `services/ingestion/tests/test_main.py`:

```python
def test_cors_rejects_unknown_origin():
    response = client.options(
        "/health",
        headers={
            "Origin": "https://evil.example.com",
            "Access-Control-Request-Method": "GET",
        },
    )
    assert response.headers.get("access-control-allow-origin") != "*"
    assert "evil.example.com" not in response.headers.get(
        "access-control-allow-origin", ""
    )
```

- [ ] **Step 4: Run ingestion tests**

Run: `cd services/ingestion && pip install -r requirements.txt && pytest tests/ -v`
Expected: All tests pass, including the new CORS test.

- [ ] **Step 5: Commit**

```bash
git add services/ingestion/app/config.py services/ingestion/app/main.py services/ingestion/tests/test_main.py
git commit -m "fix: harden CORS on ingestion service — configurable allowed origins"
```

---

### Task 2: CORS Hardening — Chat Service

**Files:**
- Modify: `services/chat/app/config.py:4-10`
- Modify: `services/chat/app/main.py:15-20`
- Modify: `services/chat/tests/test_main.py`

- [ ] **Step 1: Add `allowed_origins` to chat config**

In `services/chat/app/config.py`, add the `allowed_origins` field:

```python
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    ollama_base_url: str = "http://host.docker.internal:11434"
    chat_model: str = "mistral"
    embedding_model: str = "nomic-embed-text"
    qdrant_host: str = "qdrant"
    qdrant_port: int = 6333
    collection_name: str = "documents"
    allowed_origins: str = "https://kylebradshaw.dev"


settings = Settings()
```

- [ ] **Step 2: Update chat CORS middleware to use config**

In `services/chat/app/main.py`, replace the hardcoded wildcard CORS:

```python
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.allowed_origins.split(","),
    allow_methods=["*"],
    allow_headers=["*"],
)
```

- [ ] **Step 3: Add CORS test to chat**

Add to the end of `services/chat/tests/test_main.py`:

```python
def test_cors_rejects_unknown_origin():
    response = client.options(
        "/health",
        headers={
            "Origin": "https://evil.example.com",
            "Access-Control-Request-Method": "GET",
        },
    )
    assert response.headers.get("access-control-allow-origin") != "*"
    assert "evil.example.com" not in response.headers.get(
        "access-control-allow-origin", ""
    )
```

- [ ] **Step 4: Run chat tests**

Run: `cd services/chat && pip install -r requirements.txt && pytest tests/ -v`
Expected: All tests pass, including the new CORS test.

- [ ] **Step 5: Commit**

```bash
git add services/chat/app/config.py services/chat/app/main.py services/chat/tests/test_main.py
git commit -m "fix: harden CORS on chat service — configurable allowed origins"
```

---

### Task 3: Update .env.example

**Files:**
- Modify: `.env.example`

- [ ] **Step 1: Add ALLOWED_ORIGINS to .env.example**

Add to the end of `.env.example`:

```
# CORS
ALLOWED_ORIGINS=https://kylebradshaw.dev
```

- [ ] **Step 2: Commit**

```bash
git add .env.example
git commit -m "docs: add ALLOWED_ORIGINS to .env.example"
```

---

### Task 4: Dockerfile Hardening — Ingestion Service

**Files:**
- Modify: `services/ingestion/Dockerfile`

- [ ] **Step 1: Rewrite ingestion Dockerfile with hardening**

Replace `services/ingestion/Dockerfile` with:

```dockerfile
FROM python:3.11-slim

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY app/ ./app/

RUN useradd --create-home appuser
USER appuser

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
```

- [ ] **Step 2: Verify Docker build**

Run: `docker build -t ingestion-test services/ingestion/`
Expected: Build succeeds with no errors.

- [ ] **Step 3: Commit**

```bash
git add services/ingestion/Dockerfile
git commit -m "fix: harden ingestion Dockerfile — non-root user, Python env vars"
```

---

### Task 5: Dockerfile Hardening — Chat Service

**Files:**
- Modify: `services/chat/Dockerfile`

- [ ] **Step 1: Rewrite chat Dockerfile with hardening**

Replace `services/chat/Dockerfile` with:

```dockerfile
FROM python:3.11-slim

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY app/ ./app/

RUN useradd --create-home appuser
USER appuser

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
```

- [ ] **Step 2: Verify Docker build**

Run: `docker build -t chat-test services/chat/`
Expected: Build succeeds with no errors.

- [ ] **Step 3: Commit**

```bash
git add services/chat/Dockerfile
git commit -m "fix: harden chat Dockerfile — non-root user, Python env vars"
```

---

### Task 6: Gitleaks Config

**Files:**
- Create: `.gitleaks.toml`

- [ ] **Step 1: Create gitleaks config**

Create `.gitleaks.toml` at the repo root:

```toml
title = "gitleaks config"

[allowlist]
paths = [
    '''\.env\.example''',
    '''lessons/.*\.ipynb''',
    '''frontend/node_modules/.*''',
]
```

This allows `.env.example` (which has placeholder values, not real secrets), lesson notebooks (educational content), and node_modules (third-party code).

- [ ] **Step 2: Commit**

```bash
git add .gitleaks.toml
git commit -m "chore: add gitleaks config for secrets scanning"
```

---

### Task 7: CI Security Jobs — Bandit (Python SAST)

**Files:**
- Modify: `.github/workflows/ci.yml`

- [ ] **Step 1: Add security-bandit job**

Add after the `docker-build` job in `.github/workflows/ci.yml`:

```yaml
  security-bandit:
    name: Security - Bandit (Python SAST)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install Bandit
        run: pip install bandit

      - name: Run Bandit
        run: bandit -r services/ -f json -ll -o bandit-report.json || true

      - name: Check results
        run: |
          if [ -f bandit-report.json ]; then
            python -c "
          import json, sys
          with open('bandit-report.json') as f:
              data = json.load(f)
          results = data.get('results', [])
          if results:
              for r in results:
                  print(f\"{r['severity']}: {r['filename']}:{r['line_number']} - {r['issue_text']}\")
              sys.exit(1)
          print('No issues found')
          "
          fi
```

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: add Bandit Python SAST scanning job"
```

---

### Task 8: CI Security Jobs — pip-audit (Dependency CVEs)

**Files:**
- Modify: `.github/workflows/ci.yml`

- [ ] **Step 1: Add security-pip-audit job**

Add after the `security-bandit` job in `.github/workflows/ci.yml`:

```yaml
  security-pip-audit:
    name: Security - pip-audit (${{ matrix.service }})
    runs-on: ubuntu-latest
    strategy:
      matrix:
        service: [ingestion, chat]
    steps:
      - uses: actions/checkout@v4

      - name: Set up Python
        uses: actions/setup-python@v5
        with:
          python-version: "3.11"

      - name: Install dependencies
        run: pip install -r services/${{ matrix.service }}/requirements.txt

      - name: Install pip-audit
        run: pip install pip-audit

      - name: Run pip-audit
        run: pip-audit
```

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: add pip-audit dependency vulnerability scanning"
```

---

### Task 9: CI Security Jobs — npm audit (Node CVEs)

**Files:**
- Modify: `.github/workflows/ci.yml`

- [ ] **Step 1: Add security-npm-audit job**

Add after the `security-pip-audit` job in `.github/workflows/ci.yml`:

```yaml
  security-npm-audit:
    name: Security - npm audit
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: frontend
    steps:
      - uses: actions/checkout@v4

      - name: Set up Node
        uses: actions/setup-node@v4
        with:
          node-version: "20"

      - name: Install dependencies
        run: npm ci

      - name: Run npm audit
        run: npm audit --audit-level=high
```

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: add npm audit dependency vulnerability scanning"
```

---

### Task 10: CI Security Jobs — gitleaks (Secrets Detection)

**Files:**
- Modify: `.github/workflows/ci.yml`

- [ ] **Step 1: Add security-gitleaks job**

Add after the `security-npm-audit` job in `.github/workflows/ci.yml`:

```yaml
  security-gitleaks:
    name: Security - Gitleaks (Secrets)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Run Gitleaks
        uses: gitleaks/gitleaks-action@v2
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: add gitleaks secrets detection scanning"
```

---

### Task 11: CI Security Jobs — Hadolint (Dockerfile Linting)

**Files:**
- Modify: `.github/workflows/ci.yml`

- [ ] **Step 1: Add security-hadolint job**

Add after the `security-gitleaks` job in `.github/workflows/ci.yml`:

```yaml
  security-hadolint:
    name: Security - Hadolint (${{ matrix.dockerfile }})
    runs-on: ubuntu-latest
    strategy:
      matrix:
        dockerfile:
          - services/ingestion/Dockerfile
          - services/chat/Dockerfile
    steps:
      - uses: actions/checkout@v4

      - name: Run Hadolint
        uses: hadolint/hadolint-action@v3.1.0
        with:
          dockerfile: ${{ matrix.dockerfile }}
```

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: add Hadolint Dockerfile linting"
```

---

### Task 12: CI Security Jobs — CORS Guardrail

**Files:**
- Modify: `.github/workflows/ci.yml`

- [ ] **Step 1: Add security-cors-check job**

Add after the `security-hadolint` job in `.github/workflows/ci.yml`:

```yaml
  security-cors-check:
    name: Security - CORS Guardrail
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Check for wildcard CORS
        run: |
          if grep -r 'allow_origins=\["\*"\]' services/; then
            echo "ERROR: Wildcard CORS found in services/. Use ALLOWED_ORIGINS env var instead."
            exit 1
          fi
          echo "CORS check passed — no wildcard origins found."
```

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: add CORS wildcard guardrail check"
```

---

### Task 13: Update Deploy Gate

**Files:**
- Modify: `.github/workflows/ci.yml:109`

- [ ] **Step 1: Update deploy job needs**

In `.github/workflows/ci.yml`, update the `deploy` job's `needs` list from:

```yaml
    needs: [backend-lint, backend-tests, frontend-checks, docker-build]
```

to:

```yaml
    needs:
      - backend-lint
      - backend-tests
      - frontend-checks
      - docker-build
      - security-bandit
      - security-pip-audit
      - security-npm-audit
      - security-gitleaks
      - security-hadolint
      - security-cors-check
```

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: gate deploy on all security scanning jobs"
```

---

### Task 14: Final Verification

- [ ] **Step 1: Run all backend tests**

Run: `pytest services/ingestion/tests/ services/chat/tests/ -v`
Expected: All tests pass (including new CORS tests).

- [ ] **Step 2: Validate CI workflow syntax**

Run: `python -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yml'))" && echo "Valid YAML"`
Expected: "Valid YAML"

- [ ] **Step 3: Verify no wildcard CORS in services**

Run: `grep -r 'allow_origins=\["\*"\]' services/ && echo "FAIL" || echo "PASS"`
Expected: "PASS"

- [ ] **Step 4: Verify Dockerfiles have non-root user**

Run: `grep -l "USER appuser" services/*/Dockerfile`
Expected: Both Dockerfiles listed.

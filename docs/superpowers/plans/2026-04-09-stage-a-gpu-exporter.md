# Stage A — GPU Exporter Restore Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restore GPU metrics in Grafana by installing `nvidia_gpu_exporter` v1.4.1 on the Windows host as an auto-starting Windows service, and verify Prometheus is scraping it.

**Architecture:** `nvidia_gpu_exporter` (utkuozdemir) is a standalone Windows binary that wraps `nvidia-smi` and exposes Prometheus metrics on `:9835`. It does NOT require GeForce Experience. We install it to `C:\tools\nvidia_gpu_exporter`, wrap it with NSSM so it runs as a Windows service (same pattern as `windows_exporter` and `cloudflared`), and verify end-to-end by curling the existing Prometheus scrape job which already targets `host.minikube.internal:9835`.

**Tech Stack:** NSSM (Non-Sucking Service Manager), PowerShell, `nvidia_gpu_exporter` v1.4.1, existing k8s Prometheus in `monitoring` namespace.

**Parent spec:** `docs/superpowers/specs/2026-04-09-grafana-overhaul-design.md`

**Target host:** `PC@100.79.113.84` (via Tailscale SSH from Kyle's Mac).

---

## Pre-flight context (read before starting)

- `nvidia-smi` on the PC already works: `nvidia-smi --query-gpu=name,memory.used,memory.total --format=csv` returns a row for the RTX 3090. Verified 2026-04-09.
- `:9835` is currently unbound. `Get-NetTCPConnection -LocalPort 9835` returns nothing.
- No `C:\tools` directory exists yet; we create it in this plan.
- `windows_exporter` and `cloudflared` already run as Windows services (`Running` state) — we follow the same pattern.
- The existing k8s Prometheus config (`k8s/monitoring/configmaps/prometheus-config.yml`) already has a `nvidia-gpu` scrape job pointing at `host.minikube.internal:9835`. We do NOT modify Prometheus config in this stage — we just make the target return 200.
- All PowerShell commands are run over SSH as `ssh PC@100.79.113.84 'powershell -Command "<cmd>"'`. Elevation (admin rights) is required for service creation — verify the SSH user is an admin on the box (it is; `cloudflared` was installed the same way).

---

## Task 1: Download NSSM on the PC

**Files:**
- No repo changes.

- [ ] **Step 1: Create the tools directory**

Run (from Mac):
```bash
ssh PC@100.79.113.84 'powershell -Command "New-Item -ItemType Directory -Force -Path C:\tools | Out-Null; Test-Path C:\tools"'
```
Expected: `True`

- [ ] **Step 2: Download NSSM 2.24 zip**

Run:
```bash
ssh PC@100.79.113.84 'powershell -Command "Invoke-WebRequest -Uri https://nssm.cc/release/nssm-2.24.zip -OutFile C:\tools\nssm-2.24.zip; (Get-Item C:\tools\nssm-2.24.zip).Length"'
```
Expected: a byte count > 300000 (the file is ~348 KB).

- [ ] **Step 3: Extract NSSM**

Run:
```bash
ssh PC@100.79.113.84 'powershell -Command "Expand-Archive -Force -Path C:\tools\nssm-2.24.zip -DestinationPath C:\tools; Copy-Item C:\tools\nssm-2.24\win64\nssm.exe C:\tools\nssm.exe -Force; Test-Path C:\tools\nssm.exe"'
```
Expected: `True`

- [ ] **Step 4: Verify NSSM runs**

Run:
```bash
ssh PC@100.79.113.84 'powershell -Command "C:\tools\nssm.exe version"'
```
Expected: output starting with `NSSM: The non-sucking service manager` and `2.24`.

---

## Task 2: Download and stage nvidia_gpu_exporter v1.4.1

**Files:**
- No repo changes.

- [ ] **Step 1: Create install dir**

Run:
```bash
ssh PC@100.79.113.84 'powershell -Command "New-Item -ItemType Directory -Force -Path C:\tools\nvidia_gpu_exporter | Out-Null; Test-Path C:\tools\nvidia_gpu_exporter"'
```
Expected: `True`

- [ ] **Step 2: Download the Windows x86_64 release zip**

Run:
```bash
ssh PC@100.79.113.84 'powershell -Command "Invoke-WebRequest -Uri https://github.com/utkuozdemir/nvidia_gpu_exporter/releases/download/v1.4.1/nvidia_gpu_exporter_1.4.1_windows_x86_64.zip -OutFile C:\tools\nvidia_gpu_exporter\exporter.zip; (Get-Item C:\tools\nvidia_gpu_exporter\exporter.zip).Length"'
```
Expected: a byte count > 1000000.

- [ ] **Step 3: Extract the binary**

Run:
```bash
ssh PC@100.79.113.84 'powershell -Command "Expand-Archive -Force -Path C:\tools\nvidia_gpu_exporter\exporter.zip -DestinationPath C:\tools\nvidia_gpu_exporter; Get-ChildItem C:\tools\nvidia_gpu_exporter\*.exe"'
```
Expected: a line showing `nvidia_gpu_exporter.exe`.

- [ ] **Step 4: Smoke-test the binary in the foreground**

Run:
```bash
ssh PC@100.79.113.84 'powershell -Command "C:\tools\nvidia_gpu_exporter\nvidia_gpu_exporter.exe --version"'
```
Expected: output containing `1.4.1`.

---

## Task 3: Create the Windows service via NSSM

**Files:**
- No repo changes.

- [ ] **Step 1: Install the service**

Run:
```bash
ssh PC@100.79.113.84 'powershell -Command "C:\tools\nssm.exe install nvidia_gpu_exporter C:\tools\nvidia_gpu_exporter\nvidia_gpu_exporter.exe"'
```
Expected: `Service \"nvidia_gpu_exporter\" installed successfully!`

- [ ] **Step 2: Configure service arguments**

The exporter listens on `:9835` by default but we pin it explicitly so future changes are visible in the service config.

Run:
```bash
ssh PC@100.79.113.84 'powershell -Command "C:\tools\nssm.exe set nvidia_gpu_exporter AppParameters \"--web.listen-address=:9835\""'
```
Expected: `Set parameter \"AppParameters\" successfully.`

- [ ] **Step 3: Configure stdout/stderr log files**

Run:
```bash
ssh PC@100.79.113.84 'powershell -Command "New-Item -ItemType Directory -Force -Path C:\tools\nvidia_gpu_exporter\logs | Out-Null; C:\tools\nssm.exe set nvidia_gpu_exporter AppStdout C:\tools\nvidia_gpu_exporter\logs\stdout.log; C:\tools\nssm.exe set nvidia_gpu_exporter AppStderr C:\tools\nvidia_gpu_exporter\logs\stderr.log"'
```
Expected: two `Set parameter ... successfully.` lines.

- [ ] **Step 4: Set the service to auto-start on boot**

Run:
```bash
ssh PC@100.79.113.84 'powershell -Command "C:\tools\nssm.exe set nvidia_gpu_exporter Start SERVICE_AUTO_START"'
```
Expected: `Set parameter \"Start\" successfully.`

- [ ] **Step 5: Start the service**

Run:
```bash
ssh PC@100.79.113.84 'powershell -Command "Start-Service nvidia_gpu_exporter; (Get-Service nvidia_gpu_exporter).Status"'
```
Expected: `Running`

- [ ] **Step 6: Verify the port is listening**

Run:
```bash
ssh PC@100.79.113.84 'powershell -Command "Get-NetTCPConnection -LocalPort 9835 -State Listen | Select-Object LocalAddress,LocalPort,State"'
```
Expected: at least one row with `LocalPort: 9835`, `State: Listen`.

---

## Task 4: Verify metrics end-to-end

**Files:**
- No repo changes.

- [ ] **Step 1: Curl `/metrics` on the PC itself**

Run:
```bash
ssh PC@100.79.113.84 'powershell -Command "(Invoke-WebRequest -UseBasicParsing http://localhost:9835/metrics).Content" | grep -E "nvidia_smi_memory_used_bytes|nvidia_smi_utilization_gpu_ratio|nvidia_smi_temperature_gpu"'
```
Expected: at least three lines, one each containing `nvidia_smi_memory_used_bytes`, `nvidia_smi_utilization_gpu_ratio`, and `nvidia_smi_temperature_gpu`, each with a non-`NaN` numeric value.

- [ ] **Step 2: Confirm VRAM reading matches `nvidia-smi`**

Run:
```bash
ssh PC@100.79.113.84 'nvidia-smi --query-gpu=memory.used --format=csv,noheader,nounits'
```
Record the value (MiB). Then:
```bash
ssh PC@100.79.113.84 'powershell -Command "(Invoke-WebRequest -UseBasicParsing http://localhost:9835/metrics).Content" | grep "^nvidia_smi_memory_used_bytes"'
```
Expected: the exported value in bytes should be roughly `<MiB value> * 1024 * 1024` (within a few % — GPU is live, small drift is normal).

- [ ] **Step 3: Curl from inside the Minikube cluster**

The existing Prometheus scrape job targets `host.minikube.internal:9835`. Verify that hostname resolves and reaches the exporter from inside the `monitoring` namespace.

Run (from Mac):
```bash
ssh PC@100.79.113.84 'kubectl -n monitoring run curl-gpu-check --rm -it --restart=Never --image=curlimages/curl:8.9.1 -- curl -sf http://host.minikube.internal:9835/metrics | head -20'
```
Expected: the first 20 lines of the Prometheus exposition format, including several `nvidia_smi_*` metric `# HELP` / `# TYPE` lines.

- [ ] **Step 4: Verify Prometheus target is UP**

Run (from Mac — requires the existing `kubectl port-forward` or minikube service access):
```bash
ssh PC@100.79.113.84 'kubectl -n monitoring port-forward svc/prometheus 9090:9090 &' && sleep 3 && \
  curl -s 'http://localhost:9090/api/v1/targets?state=active' | \
  python3 -c 'import sys,json; d=json.load(sys.stdin); t=[x for x in d["data"]["activeTargets"] if x["labels"]["job"]=="nvidia-gpu"]; print(t[0]["health"], t[0]["lastError"]) if t else print("NO_TARGET")'
```
Expected: `up ` (no error). If the port-forward step is awkward, an equivalent acceptance check is: open Grafana, navigate to Explore, and run PromQL `up{job="nvidia-gpu"}` — expected result `1`.

Kill the port-forward afterward:
```bash
ssh PC@100.79.113.84 'pkill -f "port-forward svc/prometheus"'
```

- [ ] **Step 5: Verify a VRAM query returns data**

With port-forward still active (or re-run it), query:
```bash
curl -s 'http://localhost:9090/api/v1/query?query=nvidia_smi_memory_used_bytes' | \
  python3 -c 'import sys,json; d=json.load(sys.stdin); r=d["data"]["result"]; print("OK", r[0]["value"]) if r else print("EMPTY")'
```
Expected: `OK ['<timestamp>', '<bytes>']` with bytes > 0.

---

## Task 5: Document the install in monitoring/README.md

**Files:**
- Modify: `monitoring/README.md`

- [ ] **Step 1: Replace the stale Troubleshooting "GPU metrics missing" section and add an install section**

Read the current file first (`monitoring/README.md`), then apply the edit below.

Replace the block:

```markdown
**GPU metrics missing:**
- Ensure Docker has NVIDIA GPU support: `docker run --rm --gpus all nvidia/cuda:12.0-base nvidia-smi`
- Check exporter logs: `docker compose logs nvidia-gpu-exporter`
```

with:

```markdown
**GPU metrics missing:**

The `nvidia_gpu_exporter` runs as a Windows service on the PC (not in Docker or Minikube) because it needs direct `nvidia-smi` access on the host. To diagnose:

1. SSH to the PC and check the service: `Get-Service nvidia_gpu_exporter` — expected `Running`.
2. Check the port: `Get-NetTCPConnection -LocalPort 9835 -State Listen` — expected one row.
3. Curl metrics on the host: `(Invoke-WebRequest -UseBasicParsing http://localhost:9835/metrics).Content | Select-String nvidia_smi_memory_used_bytes`.
4. Tail the log: `Get-Content C:\tools\nvidia_gpu_exporter\logs\stderr.log -Tail 50`.
5. Verify Prometheus can reach it: PromQL `up{job="nvidia-gpu"}` in Grafana Explore — expected `1`.

### Reinstalling nvidia_gpu_exporter on the Windows host

The exporter is installed at `C:\tools\nvidia_gpu_exporter\nvidia_gpu_exporter.exe` and wrapped as a Windows service via NSSM (`C:\tools\nssm.exe`). It runs as `nvidia_gpu_exporter`, listens on `:9835`, and auto-starts on boot. Logs are in `C:\tools\nvidia_gpu_exporter\logs\`. The exporter wraps `nvidia-smi` directly and does **not** require GeForce Experience or the NVIDIA Display Container LS service.

To reinstall after a wipe, follow the Stage A plan in `docs/superpowers/plans/2026-04-09-stage-a-gpu-exporter.md`.
```

- [ ] **Step 2: Verify the edit**

Run:
```bash
grep -n "nvidia_gpu_exporter runs as a Windows service" monitoring/README.md
```
Expected: one matching line.

- [ ] **Step 3: Run Python/Go/Java preflight checks**

This task only edits a Markdown file, but the pre-commit hook runs regardless. There are no Python, frontend, Java, or Go changes — preflight for those should be no-ops or skip.

Run:
```bash
make preflight-frontend 2>&1 | tail -5
```
Expected: no failures (or "no files to check").

If any preflight fails due to unrelated lint drift, fix inline before committing.

- [ ] **Step 4: Commit**

```bash
git add monitoring/README.md
git commit -m "$(cat <<'EOF'
docs(monitoring): document nvidia_gpu_exporter Windows service install

Stage A of the Grafana observability overhaul. The exporter is now
running as a Windows service on the PC, auto-starting on boot, and
Prometheus is scraping it again.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>
EOF
)"
```

Expected: commit succeeds.

---

## Acceptance criteria (run at the end)

All four must be true:

1. `ssh PC@... 'powershell -Command "(Get-Service nvidia_gpu_exporter).Status"'` prints `Running`.
2. `(Get-Service nvidia_gpu_exporter).StartType` (via `Get-Service nvidia_gpu_exporter | Select-Object StartType`) prints `Automatic`.
3. PromQL `up{job="nvidia-gpu"}` returns `1` in Grafana.
4. PromQL `nvidia_smi_memory_used_bytes` returns a non-zero scalar matching current `nvidia-smi` output.

If all four pass, Stage A is done and we can move on to writing the Stage B plan.

---

## Rollback

If the service misbehaves:

```bash
ssh PC@100.79.113.84 'powershell -Command "Stop-Service nvidia_gpu_exporter -Force; C:\tools\nssm.exe remove nvidia_gpu_exporter confirm"'
```

This leaves the binary in `C:\tools\nvidia_gpu_exporter\` for later reuse without re-downloading.

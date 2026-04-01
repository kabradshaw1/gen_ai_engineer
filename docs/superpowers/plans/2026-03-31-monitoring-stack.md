# Monitoring Stack Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Prometheus + Grafana monitoring to the Windows Docker host, tracking CPU, RAM, GPU, and container health, with a public read-only dashboard at `grafana.kylebradshaw.dev`.

**Architecture:** Prometheus scrapes three exporters (windows_exporter for host metrics, nvidia_gpu_exporter for GPU, cAdvisor for containers) and feeds Grafana. All new services run in the existing Docker Compose stack except windows_exporter which runs as a native Windows service. Grafana is exposed publicly via the existing Cloudflare Tunnel.

**Tech Stack:** Prometheus, Grafana, cAdvisor, nvidia_gpu_exporter, windows_exporter, Docker Compose

---

### Task 1: Create Prometheus Configuration

**Files:**
- Create: `monitoring/prometheus.yml`

- [ ] **Step 1: Create the monitoring directory and prometheus.yml**

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: "prometheus"
    static_configs:
      - targets: ["localhost:9090"]

  - job_name: "windows"
    static_configs:
      - targets: ["host.docker.internal:9182"]

  - job_name: "nvidia-gpu"
    static_configs:
      - targets: ["nvidia-gpu-exporter:9835"]

  - job_name: "cadvisor"
    static_configs:
      - targets: ["cadvisor:8080"]
```

- [ ] **Step 2: Commit**

```bash
git add monitoring/prometheus.yml
git commit -m "feat(monitoring): add Prometheus scrape config"
```

---

### Task 2: Create Grafana Provisioning Files

**Files:**
- Create: `monitoring/grafana/provisioning/datasources/prometheus.yml`
- Create: `monitoring/grafana/provisioning/dashboards/dashboard.yml`

- [ ] **Step 1: Create Grafana datasource provisioning**

Create `monitoring/grafana/provisioning/datasources/prometheus.yml`:

```yaml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: false
```

- [ ] **Step 2: Create Grafana dashboard provisioning config**

Create `monitoring/grafana/provisioning/dashboards/dashboard.yml`:

```yaml
apiVersion: 1

providers:
  - name: "default"
    orgId: 1
    folder: ""
    type: file
    disableDeletion: false
    editable: true
    options:
      path: /var/lib/grafana/dashboards
      foldersFromFilesStructure: false
```

- [ ] **Step 3: Commit**

```bash
git add monitoring/grafana/provisioning/
git commit -m "feat(monitoring): add Grafana datasource and dashboard provisioning"
```

---

### Task 3: Create Grafana Dashboard JSON

**Files:**
- Create: `monitoring/grafana/dashboards/system-overview.json`

This is the main dashboard with three rows: System, GPU, Containers.

- [ ] **Step 1: Create the dashboard JSON**

Create `monitoring/grafana/dashboards/system-overview.json` with the following content. This is a complete Grafana dashboard JSON with 10 panels across 3 rows.

```json
{
  "annotations": { "list": [] },
  "editable": true,
  "fiscalYearStartMonth": 0,
  "graphTooltip": 1,
  "id": null,
  "links": [],
  "panels": [
    {
      "collapsed": false,
      "gridPos": { "h": 1, "w": 24, "x": 0, "y": 0 },
      "id": 100,
      "title": "System",
      "type": "row"
    },
    {
      "title": "CPU Usage %",
      "type": "timeseries",
      "gridPos": { "h": 8, "w": 8, "x": 0, "y": 1 },
      "id": 1,
      "datasource": { "type": "prometheus", "uid": "" },
      "fieldConfig": {
        "defaults": {
          "unit": "percent",
          "min": 0,
          "max": 100,
          "thresholds": {
            "mode": "absolute",
            "steps": [
              { "color": "green", "value": null },
              { "color": "yellow", "value": 60 },
              { "color": "red", "value": 85 }
            ]
          },
          "color": { "mode": "palette-classic" }
        },
        "overrides": []
      },
      "options": {
        "legend": { "displayMode": "list", "placement": "bottom" },
        "tooltip": { "mode": "single" }
      },
      "targets": [
        {
          "expr": "(1 - avg(rate(windows_cpu_time_total{mode=\"idle\"}[1m]))) * 100",
          "legendFormat": "CPU Usage",
          "refId": "A"
        }
      ]
    },
    {
      "title": "RAM Usage",
      "type": "gauge",
      "gridPos": { "h": 8, "w": 8, "x": 8, "y": 1 },
      "id": 2,
      "datasource": { "type": "prometheus", "uid": "" },
      "fieldConfig": {
        "defaults": {
          "unit": "decbytes",
          "min": 0,
          "thresholds": {
            "mode": "absolute",
            "steps": [
              { "color": "green", "value": null },
              { "color": "yellow", "value": 24000000000 },
              { "color": "red", "value": 28000000000 }
            ]
          },
          "color": { "mode": "thresholds" }
        },
        "overrides": []
      },
      "options": {
        "reduceOptions": { "calcs": ["lastNotNull"] },
        "showThresholdLabels": false,
        "showThresholdMarkers": true
      },
      "targets": [
        {
          "expr": "windows_memory_physical_total_bytes - windows_memory_physical_free_bytes",
          "legendFormat": "Used RAM",
          "refId": "A"
        }
      ]
    },
    {
      "title": "Disk Usage",
      "type": "gauge",
      "gridPos": { "h": 8, "w": 8, "x": 16, "y": 1 },
      "id": 3,
      "datasource": { "type": "prometheus", "uid": "" },
      "fieldConfig": {
        "defaults": {
          "unit": "percentunit",
          "min": 0,
          "max": 1,
          "thresholds": {
            "mode": "absolute",
            "steps": [
              { "color": "green", "value": null },
              { "color": "yellow", "value": 0.7 },
              { "color": "red", "value": 0.9 }
            ]
          },
          "color": { "mode": "thresholds" }
        },
        "overrides": []
      },
      "options": {
        "reduceOptions": { "calcs": ["lastNotNull"] },
        "showThresholdLabels": false,
        "showThresholdMarkers": true
      },
      "targets": [
        {
          "expr": "1 - (windows_logical_disk_free_bytes{volume=\"C:\"} / windows_logical_disk_size_bytes{volume=\"C:\"})",
          "legendFormat": "C: Drive",
          "refId": "A"
        }
      ]
    },
    {
      "collapsed": false,
      "gridPos": { "h": 1, "w": 24, "x": 0, "y": 9 },
      "id": 101,
      "title": "GPU — RTX 3090",
      "type": "row"
    },
    {
      "title": "GPU Utilization %",
      "type": "timeseries",
      "gridPos": { "h": 8, "w": 8, "x": 0, "y": 10 },
      "id": 4,
      "datasource": { "type": "prometheus", "uid": "" },
      "fieldConfig": {
        "defaults": {
          "unit": "percent",
          "min": 0,
          "max": 100,
          "thresholds": {
            "mode": "absolute",
            "steps": [
              { "color": "green", "value": null },
              { "color": "yellow", "value": 70 },
              { "color": "red", "value": 90 }
            ]
          },
          "color": { "mode": "palette-classic" }
        },
        "overrides": []
      },
      "options": {
        "legend": { "displayMode": "list", "placement": "bottom" },
        "tooltip": { "mode": "single" }
      },
      "targets": [
        {
          "expr": "nvidia_smi_utilization_gpu_ratio * 100",
          "legendFormat": "GPU Utilization",
          "refId": "A"
        }
      ]
    },
    {
      "title": "VRAM Usage",
      "type": "gauge",
      "gridPos": { "h": 8, "w": 8, "x": 8, "y": 10 },
      "id": 5,
      "datasource": { "type": "prometheus", "uid": "" },
      "fieldConfig": {
        "defaults": {
          "unit": "decbytes",
          "min": 0,
          "thresholds": {
            "mode": "absolute",
            "steps": [
              { "color": "green", "value": null },
              { "color": "yellow", "value": 16000000000 },
              { "color": "red", "value": 22000000000 }
            ]
          },
          "color": { "mode": "thresholds" }
        },
        "overrides": []
      },
      "options": {
        "reduceOptions": { "calcs": ["lastNotNull"] },
        "showThresholdLabels": false,
        "showThresholdMarkers": true
      },
      "targets": [
        {
          "expr": "nvidia_smi_memory_used_bytes",
          "legendFormat": "VRAM Used",
          "refId": "A"
        }
      ]
    },
    {
      "title": "GPU Temperature",
      "type": "stat",
      "gridPos": { "h": 8, "w": 8, "x": 16, "y": 10 },
      "id": 6,
      "datasource": { "type": "prometheus", "uid": "" },
      "fieldConfig": {
        "defaults": {
          "unit": "celsius",
          "thresholds": {
            "mode": "absolute",
            "steps": [
              { "color": "green", "value": null },
              { "color": "yellow", "value": 70 },
              { "color": "red", "value": 85 }
            ]
          },
          "color": { "mode": "thresholds" }
        },
        "overrides": []
      },
      "options": {
        "reduceOptions": { "calcs": ["lastNotNull"] },
        "colorMode": "value",
        "graphMode": "area",
        "textMode": "auto"
      },
      "targets": [
        {
          "expr": "nvidia_smi_temperature_gpu",
          "legendFormat": "GPU Temp",
          "refId": "A"
        }
      ]
    },
    {
      "collapsed": false,
      "gridPos": { "h": 1, "w": 24, "x": 0, "y": 18 },
      "id": 102,
      "title": "Containers",
      "type": "row"
    },
    {
      "title": "Container CPU Usage %",
      "type": "timeseries",
      "gridPos": { "h": 8, "w": 12, "x": 0, "y": 19 },
      "id": 7,
      "datasource": { "type": "prometheus", "uid": "" },
      "fieldConfig": {
        "defaults": {
          "unit": "percent",
          "min": 0,
          "color": { "mode": "palette-classic" }
        },
        "overrides": []
      },
      "options": {
        "legend": { "displayMode": "list", "placement": "bottom" },
        "tooltip": { "mode": "multi" }
      },
      "targets": [
        {
          "expr": "rate(container_cpu_usage_seconds_total{name=~\".+\"}[1m]) * 100",
          "legendFormat": "{{name}}",
          "refId": "A"
        }
      ]
    },
    {
      "title": "Container Memory Usage",
      "type": "timeseries",
      "gridPos": { "h": 8, "w": 12, "x": 12, "y": 19 },
      "id": 8,
      "datasource": { "type": "prometheus", "uid": "" },
      "fieldConfig": {
        "defaults": {
          "unit": "decbytes",
          "min": 0,
          "color": { "mode": "palette-classic" }
        },
        "overrides": []
      },
      "options": {
        "legend": { "displayMode": "list", "placement": "bottom" },
        "tooltip": { "mode": "multi" }
      },
      "targets": [
        {
          "expr": "container_memory_usage_bytes{name=~\".+\"}",
          "legendFormat": "{{name}}",
          "refId": "A"
        }
      ]
    },
    {
      "title": "Container Network I/O",
      "type": "timeseries",
      "gridPos": { "h": 8, "w": 12, "x": 0, "y": 27 },
      "id": 9,
      "datasource": { "type": "prometheus", "uid": "" },
      "fieldConfig": {
        "defaults": {
          "unit": "Bps",
          "color": { "mode": "palette-classic" }
        },
        "overrides": []
      },
      "options": {
        "legend": { "displayMode": "list", "placement": "bottom" },
        "tooltip": { "mode": "multi" }
      },
      "targets": [
        {
          "expr": "rate(container_network_receive_bytes_total{name=~\".+\"}[1m])",
          "legendFormat": "{{name}} RX",
          "refId": "A"
        },
        {
          "expr": "rate(container_network_transmit_bytes_total{name=~\".+\"}[1m])",
          "legendFormat": "{{name}} TX",
          "refId": "B"
        }
      ]
    },
    {
      "title": "Container Uptime",
      "type": "table",
      "gridPos": { "h": 8, "w": 12, "x": 12, "y": 27 },
      "id": 10,
      "datasource": { "type": "prometheus", "uid": "" },
      "fieldConfig": {
        "defaults": {
          "unit": "s",
          "color": { "mode": "thresholds" },
          "thresholds": {
            "mode": "absolute",
            "steps": [
              { "color": "red", "value": null },
              { "color": "yellow", "value": 300 },
              { "color": "green", "value": 3600 }
            ]
          }
        },
        "overrides": [
          {
            "matcher": { "id": "byName", "options": "name" },
            "properties": [{ "id": "custom.width", "value": 200 }]
          }
        ]
      },
      "options": {
        "showHeader": true,
        "sortBy": [{ "displayName": "name", "desc": false }]
      },
      "transformations": [
        { "id": "labelsToFields", "options": {} },
        {
          "id": "organize",
          "options": {
            "excludeByName": { "Time": true, "Value": false, "__name__": true, "id": true, "image": true, "instance": true, "job": true },
            "renameByName": { "Value": "Uptime" }
          }
        }
      ],
      "targets": [
        {
          "expr": "time() - container_start_time_seconds{name=~\".+\"}",
          "legendFormat": "{{name}}",
          "refId": "A",
          "instant": true
        }
      ]
    }
  ],
  "schemaVersion": 39,
  "tags": ["monitoring", "portfolio"],
  "templating": { "list": [] },
  "time": { "from": "now-1h", "to": "now" },
  "timepicker": {},
  "timezone": "browser",
  "title": "System Overview",
  "uid": "system-overview",
  "version": 1
}
```

- [ ] **Step 2: Commit**

```bash
git add monitoring/grafana/dashboards/system-overview.json
git commit -m "feat(monitoring): add Grafana system overview dashboard"
```

---

### Task 4: Update Docker Compose

**Files:**
- Modify: `docker-compose.yml`

- [ ] **Step 1: Add monitoring services to docker-compose.yml**

Add these services after the existing `chat` service block:

```yaml
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - prometheus_data:/prometheus
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml:ro
    extra_hosts:
      - "host.docker.internal:host-gateway"
    restart: unless-stopped

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    volumes:
      - grafana_data:/var/lib/grafana
      - ./monitoring/grafana/provisioning:/etc/grafana/provisioning:ro
      - ./monitoring/grafana/dashboards:/var/lib/grafana/dashboards:ro
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_ADMIN_PASSWORD:-admin}
      - GF_AUTH_ANONYMOUS_ENABLED=true
      - GF_AUTH_ANONYMOUS_ORG_ROLE=Viewer
      - GF_SERVER_ROOT_URL=https://grafana.kylebradshaw.dev
    depends_on:
      - prometheus
    restart: unless-stopped

  cadvisor:
    image: gcr.io/cadvisor/cadvisor:latest
    ports:
      - "8080:8080"
    volumes:
      - //var/run/docker.sock:/var/run/docker.sock:ro
    privileged: true
    restart: unless-stopped

  nvidia-gpu-exporter:
    image: utkuozdemir/nvidia_gpu_exporter:1.2.1
    ports:
      - "9835:9835"
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: 1
              capabilities: [gpu]
    restart: unless-stopped
```

- [ ] **Step 2: Add named volumes**

Update the `volumes:` section at the bottom of `docker-compose.yml`:

```yaml
volumes:
  qdrant_data:
  prometheus_data:
  grafana_data:
```

- [ ] **Step 3: Commit**

```bash
git add docker-compose.yml
git commit -m "feat(monitoring): add Prometheus, Grafana, cAdvisor, nvidia-gpu-exporter to Docker Compose"
```

---

### Task 5: Update Environment Config

**Files:**
- Modify: `.env.example`

- [ ] **Step 1: Add Grafana password to .env.example**

Append to the end of `.env.example`:

```
# Grafana
GRAFANA_ADMIN_PASSWORD=changeme
```

- [ ] **Step 2: Commit**

```bash
git add .env.example
git commit -m "feat(monitoring): add GRAFANA_ADMIN_PASSWORD to .env.example"
```

---

### Task 6: Install windows_exporter on Windows Host

This task is manual — it must be done on the Windows machine via SSH. These are instructions, not automatable steps.

- [ ] **Step 1: SSH into the Windows machine**

```bash
ssh PC@100.79.113.84
```

- [ ] **Step 2: Download and install windows_exporter**

In PowerShell on the Windows machine:

```powershell
# Download the MSI installer
Invoke-WebRequest -Uri "https://github.com/prometheus-community/windows_exporter/releases/latest/download/windows_exporter-0.30.4-amd64.msi" -OutFile "$env:TEMP\windows_exporter.msi"

# Install with default collectors (cpu, memory, logical_disk, net, os, cs)
msiexec /i "$env:TEMP\windows_exporter.msi" ENABLED_COLLECTORS="cpu,memory,logical_disk,net,os,cs" /qn
```

Note: Check https://github.com/prometheus-community/windows_exporter/releases for the latest version and update the URL accordingly.

- [ ] **Step 3: Verify windows_exporter is running**

```powershell
# Check the service is running
Get-Service windows_exporter

# Test the metrics endpoint
Invoke-WebRequest -Uri "http://localhost:9182/metrics" -UseBasicParsing | Select-Object -First 20
```

Expected: Service status `Running`, and metrics output containing `windows_cpu_time_total` and `windows_memory_physical_total_bytes`.

---

### Task 7: Add Cloudflare Tunnel Route for Grafana

This task is manual — it must be done on the Windows machine where cloudflared is configured.

- [ ] **Step 1: SSH into the Windows machine**

```bash
ssh PC@100.79.113.84
```

- [ ] **Step 2: Edit the cloudflared config**

Open the cloudflared config file (typically at `C:\Users\PC\.cloudflared\config.yml` or wherever the service config lives) and add a new ingress rule for Grafana:

```yaml
- hostname: grafana.kylebradshaw.dev
  service: http://localhost:3000
```

Add this **before** the catch-all `- service: http_status:404` rule at the bottom.

- [ ] **Step 3: Add DNS record**

```powershell
cloudflared tunnel route dns <tunnel-name> grafana.kylebradshaw.dev
```

Use the same tunnel name as the existing `api-chat` and `api-ingestion` routes.

- [ ] **Step 4: Restart cloudflared service**

```powershell
Restart-Service cloudflared
```

- [ ] **Step 5: Verify**

Wait 30 seconds, then open `https://grafana.kylebradshaw.dev` in a browser. The Grafana login page should appear (anonymous access will show the dashboard directly).

---

### Task 8: Deploy and Verify

This task runs on the Windows machine.

- [ ] **Step 1: Pull latest code on Windows machine**

```bash
ssh PC@100.79.113.84
cd /path/to/gen_ai_engineer
git pull
```

- [ ] **Step 2: Add GRAFANA_ADMIN_PASSWORD to .env**

```bash
echo "GRAFANA_ADMIN_PASSWORD=<your-secure-password>" >> .env
```

- [ ] **Step 3: Start the stack**

```bash
docker compose up -d
```

Expected: All services start, including prometheus, grafana, cadvisor, nvidia-gpu-exporter.

- [ ] **Step 4: Verify Prometheus targets**

Open `http://localhost:9090/targets` in a browser (or via SSH tunnel).

Expected: All 4 targets show as UP:
- `prometheus` (localhost:9090)
- `windows` (host.docker.internal:9182)
- `nvidia-gpu` (nvidia-gpu-exporter:9835)
- `cadvisor` (cadvisor:8080)

- [ ] **Step 5: Verify Grafana dashboard locally**

Open `http://localhost:3000` in a browser.

Expected: System Overview dashboard loads without login, showing live data for CPU, RAM, Disk, GPU utilization, VRAM, GPU temp, and container metrics.

- [ ] **Step 6: Verify public access**

Open `https://grafana.kylebradshaw.dev` in a browser.

Expected: Same dashboard, accessible without login, read-only.

- [ ] **Step 7: Commit any final adjustments**

If any dashboard tweaks or config fixes were needed during verification:

```bash
git add -A
git commit -m "fix(monitoring): adjustments from deployment verification"
```

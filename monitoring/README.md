# Monitoring Stack

Prometheus + Grafana monitoring for the Document Q&A Assistant infrastructure. Tracks system resources (CPU, RAM, disk), GPU usage (RTX 3090), and service health.

**Live dashboard:** [grafana.kylebradshaw.dev](https://grafana.kylebradshaw.dev)

## Architecture

```
windows_exporter (host:9182) ──┐
nvidia_gpu_exporter (:9835) ───┤
cAdvisor (:8080) ──────────────┼──> Prometheus (:9090) ──> Grafana (:3000) ──> Cloudflare Tunnel
qdrant /healthz ───────────────┤
ingestion /health ─────────────┤
chat /health ──────────────────┤
grafana /api/health ───────────┘
```

Prometheus scrapes all targets every 15 seconds and stores the metrics. Grafana queries Prometheus and renders the dashboard. Public access is read-only via Cloudflare Tunnel.

## Services

| Service | Port | Purpose |
|---------|------|---------|
| Prometheus | 9090 | Scrapes and stores metrics |
| Grafana | 3000 | Dashboard visualization (public read-only) |
| cAdvisor | 8080 | Docker container metrics |
| nvidia_gpu_exporter | 9835 | GPU utilization, VRAM, temperature |
| windows_exporter | 9182 | Host CPU, RAM, disk (runs as Windows service, not Docker) |

## Dashboard Panels

### System Row
- **CPU Usage %** — timeseries graph over time
- **RAM Used** — current bytes used
- **RAM %** — percentage of total RAM
- **Disk C: %** — percentage of C: drive used

### GPU Row (RTX 3090)
- **GPU Utilization %** — timeseries graph over time
- **VRAM Used** — current VRAM bytes used
- **VRAM %** — percentage of total VRAM
- **GPU Temp** — current temperature in Celsius

### Services Row
- **Services Running** — count of healthy services (green = all up, yellow = some down)
- **Total Services** — total number of monitored services

## Setup

### Prerequisites

- Docker Desktop with NVIDIA GPU support
- `windows_exporter` installed as a Windows service ([releases](https://github.com/prometheus-community/windows_exporter/releases))

### Environment

Add to your `.env` file:

```
GRAFANA_ADMIN_PASSWORD=<your-secure-password>
```

### Start

```bash
docker compose up -d
```

This starts Prometheus, Grafana, cAdvisor, and nvidia_gpu_exporter alongside the application services.

### Verify

1. **Prometheus targets:** http://localhost:9090/targets — all 8 targets should show as UP
2. **Grafana dashboard:** http://localhost:3000 — loads without login, shows live data
3. **Public access:** https://grafana.kylebradshaw.dev — same dashboard via Cloudflare Tunnel

### Grafana Access

| Role | How |
|------|-----|
| **Viewer (public)** | Visit the URL — no login required, read-only |
| **Admin** | Login at `/login` with user `admin` and your `GRAFANA_ADMIN_PASSWORD` |

## File Structure

```
monitoring/
├── README.md
├── prometheus.yml                          # Scrape targets and intervals
└── grafana/
    ├── dashboards/
    │   └── system-overview.json            # Dashboard definition (auto-provisioned)
    └── provisioning/
        ├── datasources/
        │   └── prometheus.yml              # Prometheus as default datasource
        └── dashboards/
            └── dashboard.yml               # Points Grafana to dashboard JSON files
```

## Customizing the Dashboard

The dashboard is provisioned from `monitoring/grafana/dashboards/system-overview.json`. To modify:

1. Edit the JSON file
2. Restart Grafana: `docker compose restart grafana`

Alternatively, edit in the Grafana UI (login as admin), then export the JSON and save it back to the file.

## Troubleshooting

**A Prometheus target shows as DOWN:**
- Check if the service is running: `docker compose ps`
- For windows_exporter: verify the Windows service is running with `Get-Service windows_exporter`

**GPU metrics missing:**
- Ensure Docker has NVIDIA GPU support: `docker run --rm --gpus all nvidia/cuda:12.0-base nvidia-smi`
- Check exporter logs: `docker compose logs nvidia-gpu-exporter`

**Dashboard shows "No data":**
- Wait 30 seconds for the first scrape to complete
- Check Prometheus is scraping: http://localhost:9090/targets

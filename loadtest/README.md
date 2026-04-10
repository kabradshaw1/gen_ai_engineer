# Load Testing Suite

k6-based stress tests for the Go ecommerce, auth, and AI agent services.

## Prerequisites

```bash
brew install k6
```

Services must be running (either locally via Docker Compose or on Minikube via SSH tunnel).

## Quick Start

```bash
# Ensure SSH tunnel is active
ssh -f -N -L 8000:localhost:8000 PC@100.79.113.84

# Run a single phase
k6 run loadtest/scripts/phase1-ecommerce.js

# Run a single scenario within a phase
k6 run --env SCENARIO=browse loadtest/scripts/phase1-ecommerce.js
k6 run --env SCENARIO=stockContention loadtest/scripts/phase1-ecommerce.js

# Override base URL (e.g., for local Docker Compose)
k6 run --env BASE_URL=http://localhost:8092 loadtest/scripts/phase1-ecommerce.js
```

## Pushing Metrics to Prometheus

To see k6 metrics in Grafana alongside service metrics:

```bash
# Open SSH tunnel for Prometheus (separate from the nginx tunnel)
ssh -f -N -L 9090:localhost:9090 PC@100.79.113.84

# Run with Prometheus remote-write output
k6 run \
  -o experimental-prometheus-rw \
  --env K6_PROMETHEUS_RW_SERVER_URL=http://localhost:9090/api/v1/write \
  loadtest/scripts/phase1-ecommerce.js
```

Then open the "k6 Load Test Results" dashboard in Grafana.

## Phases

| Phase | Script | What it tests |
|-------|--------|---------------|
| 1 | `phase1-ecommerce.js` | Product browsing, cart ops, checkout, stock contention |
| 2 | `phase2-auth.js` | Registration burst, login sustained load, token refresh |
| 3 | `phase3-ai-agent.js` | Simple AI queries, multi-step flows, rate limiter |

## Scenarios

Each script supports `--env SCENARIO=<name>` to run a single scenario:

**Phase 1:** `browse`, `cart`, `checkout`, `stockContention`
**Phase 2:** `registrationBurst`, `loginSustained`, `tokenRefresh`
**Phase 3:** `simpleQuery`, `multiStep`, `rateLimiter`

## Thresholds

| Endpoint | p95 Target | Error Rate |
|----------|-----------|------------|
| Product browse | < 500ms | < 1% |
| Cart operations | < 500ms | < 1% |
| Checkout | < 1s | < 1% |
| Auth login | < 2s | < 1% |
| AI agent turn | < 15s | N/A |

## Grafana Dashboard

Import `dashboards/k6-load-test.json` into Grafana, or copy it to the Grafana provisioning directory on the Windows PC.

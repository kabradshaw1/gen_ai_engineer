# Go Ecommerce Platform — Design Spec

## Overview

A Go microservices ecommerce platform demonstrating backend engineering skills for a Go backend developer role. Two Go services: an **auth service** (user management, JWT tokens) and an **ecommerce service** (products, cart, orders, async processing). Together they demonstrate: microservice decomposition, RESTful API design with Gin, JWT authentication, PostgreSQL data modeling, Redis caching, RabbitMQ async order processing, Prometheus/Grafana observability, and comprehensive testing (unit, integration, benchmark).

Frontend pages at `/go` (bio + links) and `/go/ecommerce` (storefront) in the existing Next.js app.

**Future expansion note:** Current scope covers auth, product catalog, cart, checkout, and async order processing. Potential additions: seller accounts, inventory management dashboard, search, product reviews, payment integration.

## Tech Stack

- **Go** with **Gin** framework
- **pgx** — PostgreSQL driver
- **golang-jwt/jwt/v5** — JWT authentication
- **amqp091-go** — RabbitMQ client
- **go-redis/redis/v9** — Redis caching
- **prometheus/client_golang** — Metrics
- **golangci-lint** — Linting
- **testcontainers-go** or test DB — Integration tests

## Service Architecture

Two Go microservices with shared infrastructure:

- **auth-service** — Owns user registration, login, token refresh. Owns the `users` table in `ecommercedb`.
- **ecommerce-service** — Owns products, cart, orders, async processing. Validates JWTs locally using shared secret (no inter-service call needed for auth).

Both services share the same JWT secret so tokens issued by auth-service are validated statelessly by ecommerce-service.

## Project Structure

```
go/
├── auth-service/
│   ├── cmd/server/main.go        # Entry point
│   ├── internal/
│   │   ├── handler/              # Gin HTTP handlers
│   │   │   ├── auth.go
│   │   │   └── health.go
│   │   ├── service/              # Business logic
│   │   │   └── auth.go
│   │   ├── repository/           # Database access (pgx)
│   │   │   └── user.go
│   │   ├── model/                # Domain types
│   │   │   └── user.go
│   │   └── middleware/           # Logging, metrics, CORS
│   │       ├── logging.go
│   │       ├── metrics.go
│   │       └── cors.go
│   ├── migrations/
│   │   └── 001_create_users.sql
│   ├── Dockerfile
│   └── go.mod
│
├── ecommerce-service/
│   ├── cmd/server/main.go        # Entry point
│   ├── internal/
│   │   ├── handler/              # Gin HTTP handlers
│   │   │   ├── product.go
│   │   │   ├── cart.go
│   │   │   ├── order.go
│   │   │   └── health.go
│   │   ├── service/              # Business logic
│   │   │   ├── product.go
│   │   │   ├── cart.go
│   │   │   └── order.go
│   │   ├── repository/           # Database access (pgx)
│   │   │   ├── product.go
│   │   │   ├── cart.go
│   │   │   └── order.go
│   │   ├── model/                # Domain types
│   │   │   ├── product.go
│   │   │   ├── cart.go
│   │   │   └── order.go
│   │   ├── middleware/           # Auth, logging, metrics, CORS
│   │   │   ├── auth.go           # JWT validation (shared secret)
│   │   │   ├── logging.go
│   │   │   ├── metrics.go
│   │   │   └── cors.go
│   │   └── worker/               # RabbitMQ consumers
│   │       └── order_processor.go
│   ├── migrations/
│   │   ├── 001_create_products.sql
│   │   ├── 002_create_cart_items.sql
│   │   └── 003_create_orders.sql
│   ├── Dockerfile
│   └── go.mod
│
└── k8s/                          # Shared K8s manifests
    ├── namespace.yml
    ├── configmaps/
    ├── deployments/
    ├── services/
    └── ingress.yml
```

## Data Model

### Users

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK, generated |
| email | VARCHAR(255) | Unique, not null |
| password_hash | VARCHAR(255) | bcrypt |
| name | VARCHAR(255) | Not null |
| created_at | TIMESTAMPTZ | Default now() |

### Products

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK, generated |
| name | VARCHAR(255) | Not null |
| description | TEXT | |
| price | INTEGER | Cents, not null |
| category | VARCHAR(100) | Indexed |
| image_url | VARCHAR(500) | |
| stock | INTEGER | Default 0 |
| created_at | TIMESTAMPTZ | Default now() |
| updated_at | TIMESTAMPTZ | Default now() |

Indexes: `idx_products_category` on `category`, `idx_products_price` on `price`.

### Cart Items

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK, generated |
| user_id | UUID | FK → users, not null |
| product_id | UUID | FK → products, not null |
| quantity | INTEGER | Not null, > 0 |
| created_at | TIMESTAMPTZ | Default now() |

Unique constraint on `(user_id, product_id)`.

### Orders

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK, generated |
| user_id | UUID | FK → users, not null |
| status | VARCHAR(20) | pending/processing/completed/failed |
| total | INTEGER | Cents |
| created_at | TIMESTAMPTZ | Default now() |
| updated_at | TIMESTAMPTZ | Default now() |

### Order Items

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK, generated |
| order_id | UUID | FK → orders, not null |
| product_id | UUID | FK → products, not null |
| quantity | INTEGER | Not null |
| price_at_purchase | INTEGER | Cents, snapshot at order time |

## REST API

### Auth Service (port 8091)

All public endpoints:

| Method | Path | Description |
|--------|------|-------------|
| POST | /auth/register | Register, returns tokens |
| POST | /auth/login | Login, returns tokens |
| POST | /auth/refresh | Refresh access token |
| GET | /health | Postgres connectivity check |
| GET | /metrics | Prometheus scrape endpoint |

Request/response format matches existing Java auth so the frontend auth utilities work as-is.

**Token config:** Access token 15min TTL, refresh token 7 day TTL. Same JWT secret as Java services (shared K8s secret).

### Ecommerce Service (port 8092)

#### Products (public)

| Method | Path | Description |
|--------|------|-------------|
| GET | /products | List with `?category=`, `?sort=price_asc`, `?page=`, `?limit=` |
| GET | /products/:id | Product detail |
| GET | /categories | List distinct categories |

Product list and categories cached in Redis with 5min TTL. Cache invalidated when stock changes.

#### Cart (authenticated)

| Method | Path | Description |
|--------|------|-------------|
| GET | /cart | User's cart with product details joined |
| POST | /cart | Add item `{productId, quantity}` |
| PUT | /cart/:itemId | Update quantity `{quantity}` |
| DELETE | /cart/:itemId | Remove item |

#### Orders (authenticated)

| Method | Path | Description |
|--------|------|-------------|
| POST | /orders | Checkout: cart → order, publish to RabbitMQ |
| GET | /orders | User's order history |
| GET | /orders/:id | Order detail with items |

#### Health & Metrics

| Method | Path | Description |
|--------|------|-------------|
| GET | /health | Postgres, Redis, RabbitMQ connectivity |
| GET | /metrics | Prometheus scrape endpoint |

### Error Responses (both services)

All errors return consistent JSON:

```json
{"error": "description of what went wrong"}
```

Standard HTTP status codes: 400 (bad request), 401 (no/invalid token), 403 (forbidden), 404 (not found), 409 (conflict), 500 (internal).

## Async Order Processing

### Flow

1. `POST /orders` handler:
   - Validates cart is not empty
   - Creates order with status `pending` + order items (snapshot prices)
   - Clears cart
   - Publishes `order.created` message to RabbitMQ
   - Returns order ID immediately

2. RabbitMQ worker (goroutine in ecommerce-service process):
   - Consumes from `ecommerce.orders` queue
   - Validates stock for all items
   - Decrements inventory in a transaction
   - Updates order status: `processing` → `completed`
   - On insufficient stock: status → `failed`

### RabbitMQ Setup

- Exchange: `ecommerce` (topic)
- Queue: `ecommerce.orders`
- Routing key: `order.created`

### Worker Architecture

Worker pool pattern with configurable concurrency (default 3 goroutines). Graceful shutdown via context cancellation on SIGTERM.

## Redis Caching

| Key Pattern | TTL | Invalidation |
|-------------|-----|-------------|
| `ecom:products:list:<params>` | 5min | On stock change via order worker |
| `ecom:categories` | 30min | Rarely changes |
| `ecom:product:<id>` | 5min | On stock change |

## Authentication

- **auth-service** owns user registration, login, token issuance
- **ecommerce-service** validates JWTs locally using the shared secret — no inter-service call needed
- Same JWT structure as Java services: `{sub: userId, email, iat, exp}`
- Same shared secret from K8s secret (`go-secrets` in `go-ecommerce` namespace)
- bcrypt password hashing (compatible with Java's BCryptPasswordEncoder)
- Gin middleware in ecommerce-service extracts user ID from JWT, sets in Gin context

## Frontend

### `/go` Page

Bio page about Go experience. Links to the ecommerce storefront. Matches the style of existing `/ai` and `/java` landing pages.

### `/go/ecommerce` Pages

- **Login/Register** — Same pattern as `/java/tasks` auth flow
- **Product grid** — Category filtering, price sorting, pagination
- **Product detail** — Description, price, add to cart
- **Cart** — Item list, quantity controls, total, checkout button
- **Checkout confirmation** — Order placed, order ID, status
- **Order history** — List of past orders with status

Uses existing `SiteHeader` for consistent navigation. New `GO_ECOMMERCE_URL` env var for API base URL (similar to `NEXT_PUBLIC_GATEWAY_URL`).

Auth utilities (`auth.ts`, `refreshAccessToken`) reused — same JWT format, same localStorage keys pattern (with `go_` prefix to avoid collisions with Java tokens).

## Observability

### Prometheus Metrics

Gin middleware exposes:
- `http_requests_total{method, path, status}` — request counter
- `http_request_duration_seconds{method, path}` — latency histogram
- `orders_total{status}` — business metric counter
- `rabbitmq_messages_processed_total{result}` — worker counter

Scraped by existing Prometheus instance. New Grafana dashboard for the ecommerce service.

### Structured Logging

JSON logs via `slog` (Go stdlib):
- Request ID, user ID, method, path, status, latency
- Correlation IDs on RabbitMQ messages
- Error context on failures

## Testing

### Unit Tests

- Service layer with mocked repository interfaces
- Handler tests with `httptest` recorder + Gin test context
- JWT middleware tests with valid/invalid/expired tokens
- Worker logic with mocked repository + message

### Integration Tests

- Repository layer against real Postgres (testcontainers or dedicated test DB)
- Full API tests: register → login → browse → cart → checkout flow

### Benchmark Tests

- `go test -bench` on hot paths: product listing, cart operations, JWT validation
- Demonstrates Go benchmark proficiency from job description

## CI/CD

### `.github/workflows/go-ci.yml`

Matrix strategy across both services (`auth-service`, `ecommerce-service`):

- **Lint:** `golangci-lint run`
- **Test:** `go test ./... -v -race -coverprofile=coverage.out`
- **Build:** `go build ./cmd/server`
- **Docker:** Build and push to GHCR on main (per service)

### K8s Deployment

Manifests in `go/k8s/` following existing pattern:
- `go-ecommerce` namespace
- Deployments, Services, ConfigMaps for both auth-service and ecommerce-service
- Shares existing Postgres, Redis, RabbitMQ (in `java-tasks` namespace via ExternalName or cross-namespace service references)
- Ingress routes: `/go-auth/*` → auth-service, `/go-api/*` → ecommerce-service
- Included in CI deploy step's `find` + pipe pattern

### Preflight

`make preflight-go` — lint + test locally. Added to `make preflight` umbrella command.

## Infrastructure

### Shared Resources

| Resource | How Shared |
|----------|-----------|
| Postgres | Same instance, new `ecommercedb` database |
| Redis | Same instance, `ecom:` key prefix |
| RabbitMQ | Same instance, `ecommerce` exchange |
| Prometheus | Existing instance, new scrape target |
| Grafana | Existing instance, new dashboard |

### Docker Compose (local dev)

Separate `go/docker-compose.yml`:

```yaml
auth-service:
  build: ./auth-service
  ports:
    - "8091:8091"
  environment:
    DATABASE_URL: postgres://taskuser:taskpass@postgres:5432/ecommercedb
    JWT_SECRET: ${JWT_SECRET}
    ALLOWED_ORIGINS: http://localhost:3000

ecommerce-service:
  build: ./ecommerce-service
  ports:
    - "8092:8092"
  environment:
    DATABASE_URL: postgres://taskuser:taskpass@postgres:5432/ecommercedb
    REDIS_URL: redis://redis:6379
    RABBITMQ_URL: amqp://guest:guest@rabbitmq:5672
    JWT_SECRET: ${JWT_SECRET}
    ALLOWED_ORIGINS: http://localhost:3000
```

Both services depend on the shared Postgres, Redis, and RabbitMQ from the main `docker-compose.yml`.

### Seed Data

Migration includes seed data: ~20 products across 4-5 categories with realistic names, descriptions, and prices. Enough to make the demo look real.

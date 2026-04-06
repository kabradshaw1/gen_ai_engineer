# Go Ecommerce Platform

Two Go microservices powering an ecommerce storefront: **auth-service** (user management, JWT tokens) and **ecommerce-service** (products, cart, orders, async processing).

## Architecture

```
                    ┌─────────────┐
                    │   Frontend  │
                    │  (Next.js)  │
                    └──────┬──────┘
                           │
              ┌────────────┴────────────┐
              │                         │
       ┌──────▼──────┐          ┌───────▼───────┐
       │ auth-service │          │  ecommerce-   │
       │  (port 8091) │          │   service     │
       │              │          │  (port 8092)  │
       │  /auth/*     │          │  /products    │
       │              │          │  /cart        │
       │  PostgreSQL  │          │  /orders      │
       │  (users)     │          │               │
       └──────────────┘          │  PostgreSQL   │
                                 │  Redis cache  │
                                 │  RabbitMQ     │
                                 └───────┬───────┘
                                         │
                                  ┌──────▼──────┐
                                  │   Worker    │
                                  │   Pool (3)  │
                                  │  (in-proc)  │
                                  └─────────────┘
```

## Quick Start (Docker Compose)

```bash
cd go
docker compose up --build
```

This starts:
- PostgreSQL (port 5433)
- Redis (port 6380)
- RabbitMQ (port 5673, management UI at 15673)
- auth-service (port 8091)
- ecommerce-service (port 8092)

### Run migrations

Migrations are SQL files applied manually (or via a startup script):

```bash
# Connect to Postgres
psql postgres://taskuser:taskpass@localhost:5433/ecommercedb

# Apply migrations in order
\i auth-service/migrations/001_create_users.sql
\i ecommerce-service/migrations/001_create_products.sql
\i ecommerce-service/migrations/002_create_cart_items.sql
\i ecommerce-service/migrations/003_create_orders.sql
```

## API Reference

### Auth Service (port 8091)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/auth/register` | No | Register a new user |
| POST | `/auth/login` | No | Login, returns JWT tokens |
| POST | `/auth/refresh` | No | Refresh access token |
| GET | `/health` | No | Postgres health check |
| GET | `/metrics` | No | Prometheus metrics |

#### Register

```bash
curl -X POST http://localhost:8091/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123","name":"Test User"}'
```

Response:
```json
{
  "accessToken": "eyJhbG...",
  "refreshToken": "eyJhbG...",
  "userId": "uuid",
  "email": "user@example.com",
  "name": "Test User"
}
```

#### Login

```bash
curl -X POST http://localhost:8091/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

#### Refresh

```bash
curl -X POST http://localhost:8091/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refreshToken":"eyJhbG..."}'
```

### Ecommerce Service (port 8092)

#### Products (public)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/products` | List products (paginated, filterable) |
| GET | `/products/:id` | Product detail |
| GET | `/categories` | List distinct categories |

```bash
# List all products
curl http://localhost:8092/products

# Filter by category
curl http://localhost:8092/products?category=Electronics

# Sort by price
curl http://localhost:8092/products?sort=price_asc

# Paginate
curl http://localhost:8092/products?page=2&limit=10

# Get categories
curl http://localhost:8092/categories
```

#### Cart (authenticated)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/cart` | Get cart with product details and total |
| POST | `/cart` | Add item to cart |
| PUT | `/cart/:itemId` | Update item quantity |
| DELETE | `/cart/:itemId` | Remove item from cart |

```bash
TOKEN="eyJhbG..."

# Add to cart
curl -X POST http://localhost:8092/cart \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"productId":"uuid","quantity":2}'

# View cart
curl http://localhost:8092/cart \
  -H "Authorization: Bearer $TOKEN"

# Update quantity
curl -X PUT http://localhost:8092/cart/ITEM_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"quantity":3}'

# Remove item
curl -X DELETE http://localhost:8092/cart/ITEM_ID \
  -H "Authorization: Bearer $TOKEN"
```

#### Orders (authenticated)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/orders` | Checkout (cart to order) |
| GET | `/orders` | List user's orders |
| GET | `/orders/:id` | Order detail with items |

```bash
# Checkout
curl -X POST http://localhost:8092/orders \
  -H "Authorization: Bearer $TOKEN"

# List orders
curl http://localhost:8092/orders \
  -H "Authorization: Bearer $TOKEN"

# Order detail
curl http://localhost:8092/orders/ORDER_ID \
  -H "Authorization: Bearer $TOKEN"
```

#### Health & Metrics

```bash
curl http://localhost:8092/health
curl http://localhost:8092/metrics
```

## Environment Variables

### auth-service

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgres://taskuser:taskpass@localhost:5432/ecommercedb` | PostgreSQL connection string |
| `JWT_SECRET` | `dev-secret-key-at-least-32-characters-long` | JWT signing secret |
| `ALLOWED_ORIGINS` | `http://localhost:3000` | CORS allowed origins (comma-separated) |
| `PORT` | `8091` | Server port |

### ecommerce-service

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgres://taskuser:taskpass@localhost:5432/ecommercedb` | PostgreSQL connection string |
| `REDIS_URL` | `redis://localhost:6379` | Redis connection string |
| `RABBITMQ_URL` | `amqp://guest:guest@localhost:5672` | RabbitMQ connection string |
| `JWT_SECRET` | `dev-secret-key-at-least-32-characters-long` | JWT signing secret (must match auth-service) |
| `ALLOWED_ORIGINS` | `http://localhost:3000` | CORS allowed origins (comma-separated) |
| `PORT` | `8092` | Server port |

## Development

### Prerequisites

- Go 1.26+
- Docker (for infrastructure)
- golangci-lint (`go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`)

### Run without Docker

```bash
# Start infrastructure only
docker compose up postgres redis rabbitmq -d

# Run auth-service
cd auth-service && go run ./cmd/server/

# Run ecommerce-service (separate terminal)
cd ecommerce-service && go run ./cmd/server/
```

### Run tests

```bash
# All tests with race detection
cd auth-service && go test ./... -v -race
cd ecommerce-service && go test ./... -v -race

# Benchmarks
cd auth-service && go test ./internal/service/ -bench=. -benchmem
cd ecommerce-service && go test ./internal/service/ -bench=. -benchmem
```

### Lint

```bash
cd auth-service && golangci-lint run ./...
cd ecommerce-service && golangci-lint run ./...
```

### Preflight (from repo root)

```bash
make preflight-go
```

## Project Structure

```
go/
├── auth-service/
│   ├── cmd/server/main.go          # Entry point
│   ├── internal/
│   │   ├── handler/                # HTTP handlers (auth, health)
│   │   ├── service/                # Business logic (JWT, bcrypt)
│   │   ├── repository/             # PostgreSQL queries
│   │   ├── model/                  # Domain types and DTOs
│   │   └── middleware/             # Logging, metrics, CORS
│   ├── migrations/                 # SQL migrations
│   └── Dockerfile
├── ecommerce-service/
│   ├── cmd/server/main.go          # Entry point
│   ├── internal/
│   │   ├── handler/                # HTTP handlers (product, cart, order, health)
│   │   ├── service/                # Business logic with Redis caching
│   │   ├── repository/             # PostgreSQL queries
│   │   ├── model/                  # Domain types and DTOs
│   │   ├── middleware/             # JWT auth, logging, metrics, CORS
│   │   └── worker/                 # RabbitMQ order processor (worker pool)
│   ├── migrations/                 # SQL migrations + seed data
│   └── Dockerfile
├── k8s/                            # Kubernetes manifests
│   ├── namespace.yml
│   ├── configmaps/
│   ├── deployments/
│   ├── services/
│   └── ingress.yml
└── docker-compose.yml              # Local dev stack
```

## Token Flow

```
1. User registers/logs in via auth-service
2. auth-service returns access token (15min) + refresh token (7 days)
3. Frontend stores tokens in localStorage (go_access_token, go_refresh_token)
4. Requests to ecommerce-service include: Authorization: Bearer <access_token>
5. ecommerce-service validates JWT locally (shared secret, no auth-service call)
6. On 403, frontend calls auth-service /auth/refresh and retries
```

## Order Processing Flow

```
1. POST /orders → validates cart, creates order (status: pending), clears cart
2. Publishes order.created message to RabbitMQ (ecommerce exchange)
3. Worker goroutine picks up message from ecommerce.orders queue
4. Worker validates stock, decrements inventory in a transaction
5. Order status updated: pending → processing → completed (or failed)
6. Product cache invalidated in Redis
```

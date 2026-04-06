# Go Ecommerce Platform Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build two Go microservices (auth + ecommerce) with a Next.js frontend to demonstrate Go backend engineering skills for a job application.

**Architecture:** Two Gin-based Go services sharing Postgres (`ecommercedb`), Redis, and RabbitMQ. Auth-service owns users/tokens. Ecommerce-service owns products/cart/orders and validates JWTs statelessly via shared secret. RabbitMQ worker pool processes orders asynchronously.

**Tech Stack:** Go 1.22+, Gin, pgx, golang-jwt, amqp091-go, go-redis, prometheus/client_golang, Next.js, TypeScript, Docker, K8s

---

## File Map

### Auth Service (`go/auth-service/`)

| File | Responsibility |
|------|---------------|
| `cmd/server/main.go` | Entry point, wires dependencies, starts Gin server |
| `internal/model/user.go` | User struct, AuthResponse, request DTOs |
| `internal/repository/user.go` | Postgres CRUD for users table |
| `internal/service/auth.go` | Business logic: register, login, refresh, JWT signing |
| `internal/handler/auth.go` | Gin handlers for /auth/* routes |
| `internal/handler/health.go` | Health check handler |
| `internal/middleware/logging.go` | Request logging middleware (slog) |
| `internal/middleware/metrics.go` | Prometheus metrics middleware |
| `internal/middleware/cors.go` | CORS middleware |
| `migrations/001_create_users.sql` | Users table DDL |
| `go.mod` | Module definition |
| `Dockerfile` | Multi-stage Docker build |

### Ecommerce Service (`go/ecommerce-service/`)

| File | Responsibility |
|------|---------------|
| `cmd/server/main.go` | Entry point, wires dependencies, starts Gin + workers |
| `internal/model/product.go` | Product struct, list params |
| `internal/model/cart.go` | CartItem struct, add/update DTOs |
| `internal/model/order.go` | Order, OrderItem structs, order status enum |
| `internal/repository/product.go` | Postgres CRUD for products |
| `internal/repository/cart.go` | Postgres CRUD for cart_items |
| `internal/repository/order.go` | Postgres CRUD for orders + order_items |
| `internal/service/product.go` | Product listing, detail, category logic + Redis cache |
| `internal/service/cart.go` | Cart add/update/remove/get logic |
| `internal/service/order.go` | Checkout flow, order queries |
| `internal/handler/product.go` | Gin handlers for /products, /categories |
| `internal/handler/cart.go` | Gin handlers for /cart |
| `internal/handler/order.go` | Gin handlers for /orders |
| `internal/handler/health.go` | Health check handler |
| `internal/middleware/auth.go` | JWT validation middleware (shared secret) |
| `internal/middleware/logging.go` | Request logging middleware (slog) |
| `internal/middleware/metrics.go` | Prometheus metrics middleware |
| `internal/middleware/cors.go` | CORS middleware |
| `internal/worker/order_processor.go` | RabbitMQ consumer, worker pool, stock processing |
| `migrations/001_create_products.sql` | Products table DDL + seed data |
| `migrations/002_create_cart_items.sql` | Cart items table DDL |
| `migrations/003_create_orders.sql` | Orders + order_items tables DDL |
| `go.mod` | Module definition |
| `Dockerfile` | Multi-stage Docker build |

### Frontend (`frontend/src/`)

| File | Responsibility |
|------|---------------|
| `app/go/page.tsx` | Go landing page (bio + link to ecommerce) |
| `app/go/layout.tsx` | Go section layout |
| `app/go/ecommerce/page.tsx` | Ecommerce storefront (product grid) |
| `app/go/ecommerce/[productId]/page.tsx` | Product detail page |
| `app/go/ecommerce/cart/page.tsx` | Cart page |
| `app/go/ecommerce/orders/page.tsx` | Order history |
| `components/go/GoAuthProvider.tsx` | Auth context for Go services |
| `components/go/ProductCard.tsx` | Product card component |
| `components/go/CartItem.tsx` | Cart line item component |
| `lib/go-auth.ts` | Go auth token utilities (go_ prefix keys) |
| `lib/go-api.ts` | Fetch wrapper for Go ecommerce API |

### Infrastructure

| File | Responsibility |
|------|---------------|
| `go/docker-compose.yml` | Local dev compose for both Go services |
| `go/k8s/namespace.yml` | go-ecommerce namespace |
| `go/k8s/configmaps/auth-service-config.yml` | Auth service env vars |
| `go/k8s/configmaps/ecommerce-service-config.yml` | Ecommerce service env vars |
| `go/k8s/deployments/auth-service.yml` | Auth service deployment |
| `go/k8s/deployments/ecommerce-service.yml` | Ecommerce service deployment |
| `go/k8s/services/auth-service.yml` | Auth service ClusterIP |
| `go/k8s/services/ecommerce-service.yml` | Ecommerce service ClusterIP |
| `go/k8s/ingress.yml` | Ingress routes for /go-auth/*, /go-api/* |
| `.github/workflows/go-ci.yml` | CI pipeline: lint, test, build, Docker push |
| `Makefile` | Add preflight-go target |

---

## Part 1: Auth Service

### Task 1: Scaffold auth-service module and models

**Files:**
- Create: `go/auth-service/go.mod`
- Create: `go/auth-service/internal/model/user.go`

- [ ] **Step 1: Initialize Go module**

```bash
mkdir -p go/auth-service && cd go/auth-service
go mod init github.com/kabradshaw1/portfolio/go/auth-service
```

- [ ] **Step 2: Create user model and DTOs**

Create `go/auth-service/internal/model/user.go`:

```go
package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"createdAt"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type AuthResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	UserID       string `json:"userId"`
	Email        string `json:"email"`
	Name         string `json:"name"`
}
```

- [ ] **Step 3: Add dependencies**

```bash
cd go/auth-service
go get github.com/gin-gonic/gin
go get github.com/google/uuid
go get github.com/jackc/pgx/v5
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto
go get github.com/prometheus/client_golang
```

- [ ] **Step 4: Commit**

```bash
git add go/auth-service/
git commit -m "feat(go): scaffold auth-service module with user model"
```

### Task 2: Database migration and user repository

**Files:**
- Create: `go/auth-service/migrations/001_create_users.sql`
- Create: `go/auth-service/internal/repository/user.go`
- Create: `go/auth-service/internal/repository/user_test.go`

- [ ] **Step 1: Create users migration**

Create `go/auth-service/migrations/001_create_users.sql`:

```sql
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_email ON users (email);
```

- [ ] **Step 2: Write the user repository tests**

Create `go/auth-service/internal/repository/user_test.go`:

```go
package repository_test

import (
	"context"
	"testing"

	"github.com/kabradshaw1/portfolio/go/auth-service/internal/model"
	"github.com/kabradshaw1/portfolio/go/auth-service/internal/repository"
)

// MockDB implements a simple in-memory store for unit tests.
// Integration tests against real Postgres are in a separate _integration_test.go file.

type mockPool struct {
	users map[string]*model.User
}

func TestCreateUser(t *testing.T) {
	repo := repository.NewUserRepository(nil) // nil pool — we test interface behavior
	if repo == nil {
		t.Fatal("expected non-nil repository")
	}
}
```

Note: Full repository tests will use a real Postgres in integration tests. Unit tests verify the interface is wired correctly.

- [ ] **Step 3: Write the user repository**

Create `go/auth-service/internal/repository/user.go`:

```go
package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kabradshaw1/portfolio/go/auth-service/internal/model"
)

var ErrUserNotFound = errors.New("user not found")
var ErrEmailExists = errors.New("email already registered")

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, email, passwordHash, name string) (*model.User, error) {
	user := &model.User{}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, name) VALUES ($1, $2, $3)
		 RETURNING id, email, password_hash, name, created_at`,
		email, passwordHash, name,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.CreatedAt)
	if err != nil {
		if isDuplicateKeyError(err) {
			return nil, ErrEmailExists
		}
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, name, created_at FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	user := &model.User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, name, created_at FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func isDuplicateKeyError(err error) bool {
	return err != nil && contains(err.Error(), "duplicate key")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: Run tests**

```bash
cd go/auth-service && go test ./internal/repository/ -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add go/auth-service/migrations/ go/auth-service/internal/repository/
git commit -m "feat(go): add users migration and user repository"
```

### Task 3: Auth service business logic (JWT + bcrypt)

**Files:**
- Create: `go/auth-service/internal/service/auth.go`
- Create: `go/auth-service/internal/service/auth_test.go`

- [ ] **Step 1: Write auth service tests**

Create `go/auth-service/internal/service/auth_test.go`:

```go
package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kabradshaw1/portfolio/go/auth-service/internal/model"
	"github.com/kabradshaw1/portfolio/go/auth-service/internal/service"
)

type mockUserRepo struct {
	users map[string]*model.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*model.User)}
}

func (m *mockUserRepo) Create(ctx context.Context, email, passwordHash, name string) (*model.User, error) {
	if _, exists := m.users[email]; exists {
		return nil, fmt.Errorf("email already registered")
	}
	user := &model.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
		Name:         name,
	}
	m.users[email] = user
	return user, nil
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	user, exists := m.users[email]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*model.User, error) {
	for _, u := range m.users {
		if u.ID.String() == id {
			return u, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func TestRegister(t *testing.T) {
	repo := newMockUserRepo()
	svc := service.NewAuthService(repo, "test-secret-at-least-32-characters-long!!", 900000, 604800000)

	resp, err := svc.Register(context.Background(), "test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", resp.Email)
	}
	if resp.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if resp.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
}

func TestLogin(t *testing.T) {
	repo := newMockUserRepo()
	svc := service.NewAuthService(repo, "test-secret-at-least-32-characters-long!!", 900000, 604800000)

	_, err := svc.Register(context.Background(), "test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	resp, err := svc.Login(context.Background(), "test@example.com", "password123")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	repo := newMockUserRepo()
	svc := service.NewAuthService(repo, "test-secret-at-least-32-characters-long!!", 900000, 604800000)

	_, _ = svc.Register(context.Background(), "test@example.com", "password123", "Test User")

	_, err := svc.Login(context.Background(), "test@example.com", "wrongpassword")
	if err == nil {
		t.Error("expected error for wrong password")
	}
}

func TestRefresh(t *testing.T) {
	repo := newMockUserRepo()
	svc := service.NewAuthService(repo, "test-secret-at-least-32-characters-long!!", 900000, 604800000)

	resp, _ := svc.Register(context.Background(), "test@example.com", "password123", "Test User")

	newResp, err := svc.Refresh(context.Background(), resp.RefreshToken)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if newResp.AccessToken == "" {
		t.Error("expected non-empty access token after refresh")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd go/auth-service && go test ./internal/service/ -v
```

Expected: FAIL — `service.NewAuthService` not defined

- [ ] **Step 3: Write the auth service**

Create `go/auth-service/internal/service/auth.go`:

```go
package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kabradshaw1/portfolio/go/auth-service/internal/model"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

type UserRepo interface {
	Create(ctx context.Context, email, passwordHash, name string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
}

type AuthService struct {
	repo            UserRepo
	jwtSecret       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewAuthService(repo UserRepo, jwtSecret string, accessTTLMs, refreshTTLMs int64) *AuthService {
	return &AuthService{
		repo:            repo,
		jwtSecret:       []byte(jwtSecret),
		accessTokenTTL:  time.Duration(accessTTLMs) * time.Millisecond,
		refreshTokenTTL: time.Duration(refreshTTLMs) * time.Millisecond,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, name string) (*model.AuthResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.Create(ctx, email, string(hash), name)
	if err != nil {
		return nil, err
	}

	return s.generateTokens(user)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*model.AuthResponse, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.generateTokens(user)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*model.AuthResponse, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	user, err := s.repo.FindByID(ctx, sub)
	if err != nil {
		return nil, ErrInvalidToken
	}

	return s.generateTokens(user)
}

func (s *AuthService) generateTokens(user *model.User) (*model.AuthResponse, error) {
	now := time.Now()

	accessClaims := jwt.MapClaims{
		"sub":   user.ID.String(),
		"email": user.Email,
		"iat":   now.Unix(),
		"exp":   now.Add(s.accessTokenTTL).Unix(),
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	refreshClaims := jwt.MapClaims{
		"sub": user.ID.String(),
		"iat": now.Unix(),
		"exp": now.Add(s.refreshTokenTTL).Unix(),
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserID:       user.ID.String(),
		Email:        user.Email,
		Name:         user.Name,
	}, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd go/auth-service && go test ./internal/service/ -v
```

Expected: PASS — all 4 tests

- [ ] **Step 5: Commit**

```bash
git add go/auth-service/internal/service/
git commit -m "feat(go): add auth service with register, login, refresh"
```

### Task 4: Auth middleware (shared — used by ecommerce-service too)

**Files:**
- Create: `go/auth-service/internal/middleware/logging.go`
- Create: `go/auth-service/internal/middleware/metrics.go`
- Create: `go/auth-service/internal/middleware/cors.go`

- [ ] **Step 1: Create logging middleware**

Create `go/auth-service/internal/middleware/logging.go`:

```go
package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set("requestId", requestID)
		c.Header("X-Request-ID", requestID)

		start := time.Now()
		c.Next()
		latency := time.Since(start)

		slog.Info("request",
			"requestId", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency", latency.String(),
			"ip", c.ClientIP(),
		)
	}
}
```

- [ ] **Step 2: Create metrics middleware**

Create `go/auth-service/internal/middleware/metrics.go`:

```go
package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})
)

func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()

		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}
		status := strconv.Itoa(c.Writer.Status())

		httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}
```

- [ ] **Step 3: Create CORS middleware**

Create `go/auth-service/internal/middleware/cors.go`:

```go
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func CORS(allowedOrigins string) gin.HandlerFunc {
	origins := strings.Split(allowedOrigins, ",")
	originSet := make(map[string]bool, len(origins))
	for _, o := range origins {
		originSet[strings.TrimSpace(o)] = true
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if originSet[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With")
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
```

- [ ] **Step 4: Commit**

```bash
git add go/auth-service/internal/middleware/
git commit -m "feat(go): add logging, metrics, and CORS middleware"
```

### Task 5: Auth handlers and server entry point

**Files:**
- Create: `go/auth-service/internal/handler/auth.go`
- Create: `go/auth-service/internal/handler/health.go`
- Create: `go/auth-service/internal/handler/auth_test.go`
- Create: `go/auth-service/cmd/server/main.go`

- [ ] **Step 1: Write auth handler tests**

Create `go/auth-service/internal/handler/auth_test.go`:

```go
package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kabradshaw1/portfolio/go/auth-service/internal/handler"
	"github.com/kabradshaw1/portfolio/go/auth-service/internal/model"
	"github.com/kabradshaw1/portfolio/go/auth-service/internal/service"
)

func setupTestRouter(authSvc *service.AuthService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := handler.NewAuthHandler(authSvc)
	auth := r.Group("/auth")
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.Refresh)
	return r
}

func TestRegisterHandler(t *testing.T) {
	// Uses the same mock repo from service tests
	repo := newMockUserRepo()
	svc := service.NewAuthService(repo, "test-secret-at-least-32-characters-long!!", 900000, 604800000)
	router := setupTestRouter(svc)

	body, _ := json.Marshal(model.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp model.AuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
}

func TestRegisterHandler_InvalidEmail(t *testing.T) {
	repo := newMockUserRepo()
	svc := service.NewAuthService(repo, "test-secret-at-least-32-characters-long!!", 900000, 604800000)
	router := setupTestRouter(svc)

	body, _ := json.Marshal(map[string]string{
		"email":    "not-an-email",
		"password": "password123",
		"name":     "Test",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd go/auth-service && go test ./internal/handler/ -v
```

Expected: FAIL — `handler.NewAuthHandler` not defined

- [ ] **Step 3: Write the auth handler**

Create `go/auth-service/internal/handler/auth.go`:

```go
package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kabradshaw1/portfolio/go/auth-service/internal/model"
)

type AuthServiceInterface interface {
	Register(ctx context.Context, email, password, name string) (*model.AuthResponse, error)
	Login(ctx context.Context, email, password string) (*model.AuthResponse, error)
	Refresh(ctx context.Context, refreshToken string) (*model.AuthResponse, error)
}

type AuthHandler struct {
	svc AuthServiceInterface
}

func NewAuthHandler(svc AuthServiceInterface) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.Register(c.Request.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req model.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
		return
	}

	c.JSON(http.StatusOK, resp)
}
```

- [ ] **Step 4: Write health handler**

Create `go/auth-service/internal/handler/health.go`:

```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
	pool *pgxpool.Pool
}

func NewHealthHandler(pool *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{pool: pool}
}

func (h *HealthHandler) Health(c *gin.Context) {
	if err := h.pool.Ping(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
```

- [ ] **Step 5: Write server entry point**

Create `go/auth-service/cmd/server/main.go`:

```go
package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/kabradshaw1/portfolio/go/auth-service/internal/handler"
	"github.com/kabradshaw1/portfolio/go/auth-service/internal/middleware"
	"github.com/kabradshaw1/portfolio/go/auth-service/internal/repository"
	"github.com/kabradshaw1/portfolio/go/auth-service/internal/service"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	dbURL := envOrDefault("DATABASE_URL", "postgres://taskuser:taskpass@localhost:5432/ecommercedb")
	jwtSecret := envOrDefault("JWT_SECRET", "dev-secret-key-at-least-32-characters-long")
	allowedOrigins := envOrDefault("ALLOWED_ORIGINS", "http://localhost:3000")
	port := envOrDefault("PORT", "8091")

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	slog.Info("connected to database")

	userRepo := repository.NewUserRepository(pool)
	authSvc := service.NewAuthService(userRepo, jwtSecret, 900000, 604800000)
	authHandler := handler.NewAuthHandler(authSvc)
	healthHandler := handler.NewHealthHandler(pool)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logging())
	r.Use(middleware.Metrics())
	r.Use(middleware.CORS(allowedOrigins))

	auth := r.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.Refresh)

	r.GET("/health", healthHandler.Health)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		slog.Info("auth-service starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down auth-service")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
	slog.Info("auth-service stopped")
}

func envOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
```

- [ ] **Step 6: Verify build compiles**

```bash
cd go/auth-service && go build ./cmd/server/
```

Expected: builds successfully

- [ ] **Step 7: Run all auth-service tests**

```bash
cd go/auth-service && go test ./... -v
```

Expected: PASS

- [ ] **Step 8: Commit**

```bash
git add go/auth-service/internal/handler/ go/auth-service/cmd/
git commit -m "feat(go): add auth handlers and server entry point"
```

### Task 6: Auth service Dockerfile

**Files:**
- Create: `go/auth-service/Dockerfile`

- [ ] **Step 1: Create multi-stage Dockerfile**

Create `go/auth-service/Dockerfile`:

```dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth-service ./cmd/server

FROM alpine:3.19

RUN adduser -D -u 1001 appuser
COPY --from=builder /auth-service /auth-service
USER appuser

EXPOSE 8091
ENTRYPOINT ["/auth-service"]
```

- [ ] **Step 2: Commit**

```bash
git add go/auth-service/Dockerfile
git commit -m "feat(go): add auth-service Dockerfile"
```

---

## Part 2: Ecommerce Service

### Task 7: Scaffold ecommerce-service module and models

**Files:**
- Create: `go/ecommerce-service/go.mod`
- Create: `go/ecommerce-service/internal/model/product.go`
- Create: `go/ecommerce-service/internal/model/cart.go`
- Create: `go/ecommerce-service/internal/model/order.go`

- [ ] **Step 1: Initialize Go module**

```bash
mkdir -p go/ecommerce-service && cd go/ecommerce-service
go mod init github.com/kabradshaw1/portfolio/go/ecommerce-service
```

- [ ] **Step 2: Create product model**

Create `go/ecommerce-service/internal/model/product.go`:

```go
package model

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       int       `json:"price"` // cents
	Category    string    `json:"category"`
	ImageURL    string    `json:"imageUrl"`
	Stock       int       `json:"stock"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type ProductListParams struct {
	Category string
	Sort     string // "price_asc", "price_desc", "name_asc"
	Page     int
	Limit    int
}

type ProductListResponse struct {
	Products []Product `json:"products"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	Limit    int       `json:"limit"`
}
```

- [ ] **Step 3: Create cart model**

Create `go/ecommerce-service/internal/model/cart.go`:

```go
package model

import (
	"time"

	"github.com/google/uuid"
)

type CartItem struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	ProductID uuid.UUID `json:"productId"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"createdAt"`
	// Joined fields from product
	ProductName  string `json:"productName,omitempty"`
	ProductPrice int    `json:"productPrice,omitempty"`
	ProductImage string `json:"productImage,omitempty"`
}

type AddToCartRequest struct {
	ProductID string `json:"productId" binding:"required,uuid"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

type UpdateCartRequest struct {
	Quantity int `json:"quantity" binding:"required,min=1"`
}

type CartResponse struct {
	Items []CartItem `json:"items"`
	Total int        `json:"total"` // total price in cents
}
```

- [ ] **Step 4: Create order model**

Create `go/ecommerce-service/internal/model/order.go`:

```go
package model

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusCompleted  OrderStatus = "completed"
	OrderStatusFailed     OrderStatus = "failed"
)

type Order struct {
	ID        uuid.UUID   `json:"id"`
	UserID    uuid.UUID   `json:"userId"`
	Status    OrderStatus `json:"status"`
	Total     int         `json:"total"` // cents
	CreatedAt time.Time   `json:"createdAt"`
	UpdatedAt time.Time   `json:"updatedAt"`
	Items     []OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
	ID              uuid.UUID `json:"id"`
	OrderID         uuid.UUID `json:"orderId"`
	ProductID       uuid.UUID `json:"productId"`
	Quantity        int       `json:"quantity"`
	PriceAtPurchase int       `json:"priceAtPurchase"` // cents
	ProductName     string    `json:"productName,omitempty"`
}

type OrderMessage struct {
	OrderID string `json:"orderId"`
}
```

- [ ] **Step 5: Add dependencies**

```bash
cd go/ecommerce-service
go get github.com/gin-gonic/gin
go get github.com/google/uuid
go get github.com/jackc/pgx/v5
go get github.com/golang-jwt/jwt/v5
go get github.com/redis/go-redis/v9
go get github.com/rabbitmq/amqp091-go
go get github.com/prometheus/client_golang
```

- [ ] **Step 6: Commit**

```bash
git add go/ecommerce-service/
git commit -m "feat(go): scaffold ecommerce-service module with models"
```

### Task 8: Ecommerce database migrations with seed data

**Files:**
- Create: `go/ecommerce-service/migrations/001_create_products.sql`
- Create: `go/ecommerce-service/migrations/002_create_cart_items.sql`
- Create: `go/ecommerce-service/migrations/003_create_orders.sql`

- [ ] **Step 1: Create products migration with seed data**

Create `go/ecommerce-service/migrations/001_create_products.sql`:

```sql
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price INTEGER NOT NULL,
    category VARCHAR(100) NOT NULL,
    image_url VARCHAR(500),
    stock INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_products_category ON products (category);
CREATE INDEX idx_products_price ON products (price);

-- Seed data: ~20 products across 5 categories
INSERT INTO products (name, description, price, category, stock) VALUES
-- Electronics
('Wireless Bluetooth Headphones', 'Noise-canceling over-ear headphones with 30hr battery', 7999, 'Electronics', 50),
('USB-C Fast Charger', 'GaN 65W charger with 3 ports', 3499, 'Electronics', 100),
('Mechanical Keyboard', 'RGB backlit with Cherry MX switches', 12999, 'Electronics', 30),
('Portable SSD 1TB', 'NVMe external drive, USB-C, 1050MB/s', 8999, 'Electronics', 40),
-- Clothing
('Classic Cotton T-Shirt', 'Heavyweight premium cotton, unisex fit', 2499, 'Clothing', 200),
('Slim Fit Chinos', 'Stretch cotton blend, multiple colors available', 4999, 'Clothing', 80),
('Lightweight Rain Jacket', 'Packable waterproof shell with sealed seams', 6999, 'Clothing', 60),
('Merino Wool Beanie', 'Breathable and temperature regulating', 1999, 'Clothing', 150),
-- Home
('Pour-Over Coffee Maker', 'Borosilicate glass with stainless steel filter', 3999, 'Home', 70),
('Cast Iron Skillet 12"', 'Pre-seasoned, oven safe to 500F', 4499, 'Home', 45),
('LED Desk Lamp', 'Adjustable brightness and color temperature', 5999, 'Home', 55),
('Ceramic Planter Set', 'Set of 3, drainage holes included', 2999, 'Home', 90),
-- Books
('The Go Programming Language', 'Donovan & Kernighan — comprehensive Go guide', 3499, 'Books', 120),
('Designing Data-Intensive Applications', 'Martin Kleppmann — distributed systems bible', 3999, 'Books', 100),
('Clean Architecture', 'Robert C. Martin — software design principles', 2999, 'Books', 80),
('System Design Interview', 'Alex Xu — practical system design guide', 3499, 'Books', 95),
-- Sports
('Yoga Mat 6mm', 'Non-slip TPE material with carrying strap', 2999, 'Sports', 110),
('Adjustable Dumbbells', 'Quick-change weight from 5-52.5 lbs per hand', 29999, 'Sports', 20),
('Resistance Band Set', '5 bands with handles, door anchor, and bag', 1999, 'Sports', 130),
('Water Bottle 32oz', 'Insulated stainless steel, keeps cold 24hrs', 2499, 'Sports', 200);
```

- [ ] **Step 2: Create cart items migration**

Create `go/ecommerce-service/migrations/002_create_cart_items.sql`:

```sql
CREATE TABLE IF NOT EXISTS cart_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    product_id UUID NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, product_id)
);

CREATE INDEX idx_cart_items_user ON cart_items (user_id);
```

- [ ] **Step 3: Create orders migration**

Create `go/ecommerce-service/migrations/003_create_orders.sql`:

```sql
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    total INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_orders_user ON orders (user_id);
CREATE INDEX idx_orders_status ON orders (status);

CREATE TABLE IF NOT EXISTS order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL,
    price_at_purchase INTEGER NOT NULL
);

CREATE INDEX idx_order_items_order ON order_items (order_id);
```

- [ ] **Step 4: Commit**

```bash
git add go/ecommerce-service/migrations/
git commit -m "feat(go): add ecommerce database migrations with seed data"
```

### Task 9: Product repository and service with Redis caching

**Files:**
- Create: `go/ecommerce-service/internal/repository/product.go`
- Create: `go/ecommerce-service/internal/service/product.go`
- Create: `go/ecommerce-service/internal/service/product_test.go`

- [ ] **Step 1: Write product service tests**

Create `go/ecommerce-service/internal/service/product_test.go`:

```go
package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/model"
	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/service"
)

type mockProductRepo struct {
	products []model.Product
}

func newMockProductRepo() *mockProductRepo {
	return &mockProductRepo{
		products: []model.Product{
			{ID: uuid.New(), Name: "Test Product", Price: 1999, Category: "Electronics", Stock: 10, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: uuid.New(), Name: "Another Product", Price: 2999, Category: "Books", Stock: 5, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
	}
}

func (m *mockProductRepo) List(ctx context.Context, params model.ProductListParams) ([]model.Product, int, error) {
	var filtered []model.Product
	for _, p := range m.products {
		if params.Category != "" && p.Category != params.Category {
			continue
		}
		filtered = append(filtered, p)
	}
	return filtered, len(filtered), nil
}

func (m *mockProductRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	for _, p := range m.products {
		if p.ID == id {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("product not found")
}

func (m *mockProductRepo) Categories(ctx context.Context) ([]string, error) {
	seen := map[string]bool{}
	var cats []string
	for _, p := range m.products {
		if !seen[p.Category] {
			cats = append(cats, p.Category)
			seen[p.Category] = true
		}
	}
	return cats, nil
}

func (m *mockProductRepo) DecrementStock(ctx context.Context, productID uuid.UUID, qty int) error {
	for i, p := range m.products {
		if p.ID == productID {
			if p.Stock < qty {
				return fmt.Errorf("insufficient stock")
			}
			m.products[i].Stock -= qty
			return nil
		}
	}
	return fmt.Errorf("product not found")
}

func TestListProducts(t *testing.T) {
	repo := newMockProductRepo()
	svc := service.NewProductService(repo, nil) // nil Redis for unit tests

	resp, err := svc.List(context.Background(), model.ProductListParams{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Products) != 2 {
		t.Errorf("expected 2 products, got %d", len(resp.Products))
	}
}

func TestListProductsByCategory(t *testing.T) {
	repo := newMockProductRepo()
	svc := service.NewProductService(repo, nil)

	resp, err := svc.List(context.Background(), model.ProductListParams{Category: "Electronics", Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Products) != 1 {
		t.Errorf("expected 1 product, got %d", len(resp.Products))
	}
}

func TestGetCategories(t *testing.T) {
	repo := newMockProductRepo()
	svc := service.NewProductService(repo, nil)

	cats, err := svc.Categories(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cats) != 2 {
		t.Errorf("expected 2 categories, got %d", len(cats))
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd go/ecommerce-service && go test ./internal/service/ -v
```

Expected: FAIL — `service.NewProductService` not defined

- [ ] **Step 3: Write product repository**

Create `go/ecommerce-service/internal/repository/product.go`:

```go
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/model"
)

var ErrProductNotFound = errors.New("product not found")
var ErrInsufficientStock = errors.New("insufficient stock")

type ProductRepository struct {
	pool *pgxpool.Pool
}

func NewProductRepository(pool *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{pool: pool}
}

func (r *ProductRepository) List(ctx context.Context, params model.ProductListParams) ([]model.Product, int, error) {
	query := `SELECT id, name, description, price, category, image_url, stock, created_at, updated_at FROM products WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM products WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if params.Category != "" {
		query += fmt.Sprintf(" AND category = $%d", argIdx)
		countQuery += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, params.Category)
		argIdx++
	}

	switch params.Sort {
	case "price_asc":
		query += " ORDER BY price ASC"
	case "price_desc":
		query += " ORDER BY price DESC"
	case "name_asc":
		query += " ORDER BY name ASC"
	default:
		query += " ORDER BY created_at DESC"
	}

	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	page := params.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Category, &p.ImageURL, &p.Stock, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, err
		}
		products = append(products, p)
	}

	return products, total, nil
}

func (r *ProductRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	var p model.Product
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, description, price, category, image_url, stock, created_at, updated_at FROM products WHERE id = $1`, id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Category, &p.ImageURL, &p.Stock, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepository) Categories(ctx context.Context) ([]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT DISTINCT category FROM products ORDER BY category`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var cat string
		if err := rows.Scan(&cat); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}
	return categories, nil
}

func (r *ProductRepository) DecrementStock(ctx context.Context, productID uuid.UUID, qty int) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE products SET stock = stock - $1, updated_at = now() WHERE id = $2 AND stock >= $1`,
		qty, productID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrInsufficientStock
	}
	return nil
}
```

- [ ] **Step 4: Write product service with Redis caching**

Create `go/ecommerce-service/internal/service/product.go`:

```go
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/model"
	"github.com/redis/go-redis/v9"
)

type ProductRepo interface {
	List(ctx context.Context, params model.ProductListParams) ([]model.Product, int, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Product, error)
	Categories(ctx context.Context) ([]string, error)
	DecrementStock(ctx context.Context, productID uuid.UUID, qty int) error
}

type ProductService struct {
	repo  ProductRepo
	redis *redis.Client
}

func NewProductService(repo ProductRepo, redisClient *redis.Client) *ProductService {
	return &ProductService{repo: repo, redis: redisClient}
}

func (s *ProductService) List(ctx context.Context, params model.ProductListParams) (*model.ProductListResponse, error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}

	cacheKey := fmt.Sprintf("ecom:products:list:%s:%s:%d:%d", params.Category, params.Sort, params.Page, params.Limit)

	if s.redis != nil {
		cached, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var resp model.ProductListResponse
			if json.Unmarshal([]byte(cached), &resp) == nil {
				return &resp, nil
			}
		}
	}

	products, total, err := s.repo.List(ctx, params)
	if err != nil {
		return nil, err
	}

	resp := &model.ProductListResponse{
		Products: products,
		Total:    total,
		Page:     params.Page,
		Limit:    params.Limit,
	}

	if s.redis != nil {
		if data, err := json.Marshal(resp); err == nil {
			s.redis.Set(ctx, cacheKey, data, 5*time.Minute)
		}
	}

	return resp, nil
}

func (s *ProductService) GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	cacheKey := fmt.Sprintf("ecom:product:%s", id)

	if s.redis != nil {
		cached, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var p model.Product
			if json.Unmarshal([]byte(cached), &p) == nil {
				return &p, nil
			}
		}
	}

	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if s.redis != nil {
		if data, err := json.Marshal(p); err == nil {
			s.redis.Set(ctx, cacheKey, data, 5*time.Minute)
		}
	}

	return p, nil
}

func (s *ProductService) Categories(ctx context.Context) ([]string, error) {
	cacheKey := "ecom:categories"

	if s.redis != nil {
		cached, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var cats []string
			if json.Unmarshal([]byte(cached), &cats) == nil {
				return cats, nil
			}
		}
	}

	cats, err := s.repo.Categories(ctx)
	if err != nil {
		return nil, err
	}

	if s.redis != nil {
		if data, err := json.Marshal(cats); err == nil {
			s.redis.Set(ctx, cacheKey, data, 30*time.Minute)
		}
	}

	return cats, nil
}

func (s *ProductService) InvalidateCache(ctx context.Context) {
	if s.redis == nil {
		return
	}
	iter := s.redis.Scan(ctx, 0, "ecom:products:*", 100).Iterator()
	for iter.Next(ctx) {
		s.redis.Del(ctx, iter.Val())
	}
	s.redis.Del(ctx, "ecom:categories")
}
```

- [ ] **Step 5: Run tests**

```bash
cd go/ecommerce-service && go test ./internal/service/ -v
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add go/ecommerce-service/internal/repository/product.go go/ecommerce-service/internal/service/
git commit -m "feat(go): add product repository and service with Redis caching"
```

### Task 10: Cart repository and service

**Files:**
- Create: `go/ecommerce-service/internal/repository/cart.go`
- Create: `go/ecommerce-service/internal/service/cart.go`
- Create: `go/ecommerce-service/internal/service/cart_test.go`

- [ ] **Step 1: Write cart service tests**

Create `go/ecommerce-service/internal/service/cart_test.go`:

```go
package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/model"
	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/service"
)

type mockCartRepo struct {
	items []model.CartItem
}

func newMockCartRepo() *mockCartRepo {
	return &mockCartRepo{items: []model.CartItem{}}
}

func (m *mockCartRepo) GetByUser(ctx context.Context, userID uuid.UUID) ([]model.CartItem, error) {
	var result []model.CartItem
	for _, item := range m.items {
		if item.UserID == userID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (m *mockCartRepo) AddItem(ctx context.Context, userID, productID uuid.UUID, quantity int) (*model.CartItem, error) {
	for i, item := range m.items {
		if item.UserID == userID && item.ProductID == productID {
			m.items[i].Quantity += quantity
			return &m.items[i], nil
		}
	}
	item := model.CartItem{
		ID:        uuid.New(),
		UserID:    userID,
		ProductID: productID,
		Quantity:  quantity,
	}
	m.items = append(m.items, item)
	return &item, nil
}

func (m *mockCartRepo) UpdateQuantity(ctx context.Context, itemID uuid.UUID, userID uuid.UUID, quantity int) error {
	for i, item := range m.items {
		if item.ID == itemID && item.UserID == userID {
			m.items[i].Quantity = quantity
			return nil
		}
	}
	return fmt.Errorf("cart item not found")
}

func (m *mockCartRepo) RemoveItem(ctx context.Context, itemID uuid.UUID, userID uuid.UUID) error {
	for i, item := range m.items {
		if item.ID == itemID && item.UserID == userID {
			m.items = append(m.items[:i], m.items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("cart item not found")
}

func (m *mockCartRepo) ClearCart(ctx context.Context, userID uuid.UUID) error {
	var remaining []model.CartItem
	for _, item := range m.items {
		if item.UserID != userID {
			remaining = append(remaining, item)
		}
	}
	m.items = remaining
	return nil
}

func TestAddToCart(t *testing.T) {
	repo := newMockCartRepo()
	svc := service.NewCartService(repo)

	userID := uuid.New()
	productID := uuid.New()

	item, err := svc.AddItem(context.Background(), userID, productID, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Quantity != 2 {
		t.Errorf("expected quantity 2, got %d", item.Quantity)
	}
}

func TestGetCart(t *testing.T) {
	repo := newMockCartRepo()
	svc := service.NewCartService(repo)

	userID := uuid.New()
	svc.AddItem(context.Background(), userID, uuid.New(), 1)
	svc.AddItem(context.Background(), userID, uuid.New(), 3)

	items, err := svc.GetCart(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestRemoveFromCart(t *testing.T) {
	repo := newMockCartRepo()
	svc := service.NewCartService(repo)

	userID := uuid.New()
	item, _ := svc.AddItem(context.Background(), userID, uuid.New(), 1)

	err := svc.RemoveItem(context.Background(), item.ID, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items, _ := svc.GetCart(context.Background(), userID)
	if len(items) != 0 {
		t.Errorf("expected empty cart, got %d items", len(items))
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd go/ecommerce-service && go test ./internal/service/ -v -run TestAddToCart
```

Expected: FAIL — `service.NewCartService` not defined

- [ ] **Step 3: Write cart repository**

Create `go/ecommerce-service/internal/repository/cart.go`:

```go
package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/model"
)

var ErrCartItemNotFound = errors.New("cart item not found")

type CartRepository struct {
	pool *pgxpool.Pool
}

func NewCartRepository(pool *pgxpool.Pool) *CartRepository {
	return &CartRepository{pool: pool}
}

func (r *CartRepository) GetByUser(ctx context.Context, userID uuid.UUID) ([]model.CartItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT ci.id, ci.user_id, ci.product_id, ci.quantity, ci.created_at,
		        p.name, p.price, p.image_url
		 FROM cart_items ci
		 JOIN products p ON ci.product_id = p.id
		 WHERE ci.user_id = $1
		 ORDER BY ci.created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.CartItem
	for rows.Next() {
		var item model.CartItem
		if err := rows.Scan(&item.ID, &item.UserID, &item.ProductID, &item.Quantity, &item.CreatedAt,
			&item.ProductName, &item.ProductPrice, &item.ProductImage); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *CartRepository) AddItem(ctx context.Context, userID, productID uuid.UUID, quantity int) (*model.CartItem, error) {
	var item model.CartItem
	err := r.pool.QueryRow(ctx,
		`INSERT INTO cart_items (user_id, product_id, quantity)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, product_id)
		 DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity
		 RETURNING id, user_id, product_id, quantity, created_at`,
		userID, productID, quantity,
	).Scan(&item.ID, &item.UserID, &item.ProductID, &item.Quantity, &item.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *CartRepository) UpdateQuantity(ctx context.Context, itemID, userID uuid.UUID, quantity int) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE cart_items SET quantity = $1 WHERE id = $2 AND user_id = $3`,
		quantity, itemID, userID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrCartItemNotFound
	}
	return nil
}

func (r *CartRepository) RemoveItem(ctx context.Context, itemID, userID uuid.UUID) error {
	result, err := r.pool.Exec(ctx,
		`DELETE FROM cart_items WHERE id = $1 AND user_id = $2`,
		itemID, userID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrCartItemNotFound
	}
	return nil
}

func (r *CartRepository) ClearCart(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM cart_items WHERE user_id = $1`, userID)
	return err
}
```

- [ ] **Step 4: Write cart service**

Create `go/ecommerce-service/internal/service/cart.go`:

```go
package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/model"
)

type CartRepo interface {
	GetByUser(ctx context.Context, userID uuid.UUID) ([]model.CartItem, error)
	AddItem(ctx context.Context, userID, productID uuid.UUID, quantity int) (*model.CartItem, error)
	UpdateQuantity(ctx context.Context, itemID, userID uuid.UUID, quantity int) error
	RemoveItem(ctx context.Context, itemID, userID uuid.UUID) error
	ClearCart(ctx context.Context, userID uuid.UUID) error
}

type CartService struct {
	repo CartRepo
}

func NewCartService(repo CartRepo) *CartService {
	return &CartService{repo: repo}
}

func (s *CartService) GetCart(ctx context.Context, userID uuid.UUID) ([]model.CartItem, error) {
	return s.repo.GetByUser(ctx, userID)
}

func (s *CartService) AddItem(ctx context.Context, userID, productID uuid.UUID, quantity int) (*model.CartItem, error) {
	return s.repo.AddItem(ctx, userID, productID, quantity)
}

func (s *CartService) UpdateQuantity(ctx context.Context, itemID, userID uuid.UUID, quantity int) error {
	return s.repo.UpdateQuantity(ctx, itemID, userID, quantity)
}

func (s *CartService) RemoveItem(ctx context.Context, itemID, userID uuid.UUID) error {
	return s.repo.RemoveItem(ctx, itemID, userID)
}

func (s *CartService) ClearCart(ctx context.Context, userID uuid.UUID) error {
	return s.repo.ClearCart(ctx, userID)
}
```

- [ ] **Step 5: Run tests**

```bash
cd go/ecommerce-service && go test ./internal/service/ -v -run TestAddToCart\|TestGetCart\|TestRemoveFromCart
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add go/ecommerce-service/internal/repository/cart.go go/ecommerce-service/internal/service/cart.go go/ecommerce-service/internal/service/cart_test.go
git commit -m "feat(go): add cart repository and service"
```

### Task 11: Order repository, service, and RabbitMQ publisher

**Files:**
- Create: `go/ecommerce-service/internal/repository/order.go`
- Create: `go/ecommerce-service/internal/service/order.go`
- Create: `go/ecommerce-service/internal/service/order_test.go`

- [ ] **Step 1: Write order service tests**

Create `go/ecommerce-service/internal/service/order_test.go`:

```go
package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/model"
	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/service"
)

type mockOrderRepo struct {
	orders []model.Order
}

func newMockOrderRepo() *mockOrderRepo {
	return &mockOrderRepo{orders: []model.Order{}}
}

func (m *mockOrderRepo) Create(ctx context.Context, userID uuid.UUID, total int, items []model.OrderItem) (*model.Order, error) {
	order := model.Order{
		ID:     uuid.New(),
		UserID: userID,
		Status: model.OrderStatusPending,
		Total:  total,
		Items:  items,
	}
	m.orders = append(m.orders, order)
	return &order, nil
}

func (m *mockOrderRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	for _, o := range m.orders {
		if o.ID == id {
			return &o, nil
		}
	}
	return nil, fmt.Errorf("order not found")
}

func (m *mockOrderRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]model.Order, error) {
	var result []model.Order
	for _, o := range m.orders {
		if o.UserID == userID {
			result = append(result, o)
		}
	}
	return result, nil
}

func (m *mockOrderRepo) UpdateStatus(ctx context.Context, orderID uuid.UUID, status model.OrderStatus) error {
	for i, o := range m.orders {
		if o.ID == orderID {
			m.orders[i].Status = status
			return nil
		}
	}
	return fmt.Errorf("order not found")
}

type mockPublisher struct {
	published []string
}

func (m *mockPublisher) PublishOrderCreated(orderID string) error {
	m.published = append(m.published, orderID)
	return nil
}

func TestCheckout(t *testing.T) {
	cartRepo := newMockCartRepo()
	orderRepo := newMockOrderRepo()
	publisher := &mockPublisher{}

	userID := uuid.New()
	productID := uuid.New()

	// Add items to cart with product details
	cartRepo.items = []model.CartItem{
		{ID: uuid.New(), UserID: userID, ProductID: productID, Quantity: 2, ProductPrice: 1999, ProductName: "Test"},
	}

	svc := service.NewOrderService(orderRepo, cartRepo, publisher)

	order, err := svc.Checkout(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.Status != model.OrderStatusPending {
		t.Errorf("expected status pending, got %s", order.Status)
	}
	if order.Total != 3998 { // 1999 * 2
		t.Errorf("expected total 3998, got %d", order.Total)
	}
	if len(publisher.published) != 1 {
		t.Errorf("expected 1 published message, got %d", len(publisher.published))
	}
}

func TestCheckoutEmptyCart(t *testing.T) {
	cartRepo := newMockCartRepo()
	orderRepo := newMockOrderRepo()
	publisher := &mockPublisher{}

	svc := service.NewOrderService(orderRepo, cartRepo, publisher)

	_, err := svc.Checkout(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected error for empty cart")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd go/ecommerce-service && go test ./internal/service/ -v -run TestCheckout
```

Expected: FAIL — `service.NewOrderService` not defined

- [ ] **Step 3: Write order repository**

Create `go/ecommerce-service/internal/repository/order.go`:

```go
package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/model"
)

var ErrOrderNotFound = errors.New("order not found")

type OrderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

func (r *OrderRepository) Create(ctx context.Context, userID uuid.UUID, total int, items []model.OrderItem) (*model.Order, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var order model.Order
	err = tx.QueryRow(ctx,
		`INSERT INTO orders (user_id, status, total) VALUES ($1, $2, $3)
		 RETURNING id, user_id, status, total, created_at, updated_at`,
		userID, model.OrderStatusPending, total,
	).Scan(&order.ID, &order.UserID, &order.Status, &order.Total, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		_, err := tx.Exec(ctx,
			`INSERT INTO order_items (order_id, product_id, quantity, price_at_purchase) VALUES ($1, $2, $3, $4)`,
			order.ID, item.ProductID, item.Quantity, item.PriceAtPurchase,
		)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	order.Items = items
	return &order, nil
}

func (r *OrderRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	var order model.Order
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, status, total, created_at, updated_at FROM orders WHERE id = $1`, id,
	).Scan(&order.ID, &order.UserID, &order.Status, &order.Total, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price_at_purchase, p.name
		 FROM order_items oi JOIN products p ON oi.product_id = p.id
		 WHERE oi.order_id = $1`, id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item model.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.PriceAtPurchase, &item.ProductName); err != nil {
			return nil, err
		}
		order.Items = append(order.Items, item)
	}

	return &order, nil
}

func (r *OrderRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]model.Order, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, status, total, created_at, updated_at FROM orders WHERE user_id = $1 ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.Total, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID uuid.UUID, status model.OrderStatus) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE orders SET status = $1, updated_at = now() WHERE id = $2`,
		status, orderID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrOrderNotFound
	}
	return nil
}
```

- [ ] **Step 4: Write order service**

Create `go/ecommerce-service/internal/service/order.go`:

```go
package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/model"
)

var ErrEmptyCart = errors.New("cart is empty")

type OrderRepo interface {
	Create(ctx context.Context, userID uuid.UUID, total int, items []model.OrderItem) (*model.Order, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Order, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]model.Order, error)
	UpdateStatus(ctx context.Context, orderID uuid.UUID, status model.OrderStatus) error
}

type OrderPublisher interface {
	PublishOrderCreated(orderID string) error
}

type OrderService struct {
	orderRepo OrderRepo
	cartRepo  CartRepo
	publisher OrderPublisher
}

func NewOrderService(orderRepo OrderRepo, cartRepo CartRepo, publisher OrderPublisher) *OrderService {
	return &OrderService{
		orderRepo: orderRepo,
		cartRepo:  cartRepo,
		publisher: publisher,
	}
}

func (s *OrderService) Checkout(ctx context.Context, userID uuid.UUID) (*model.Order, error) {
	cartItems, err := s.cartRepo.GetByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(cartItems) == 0 {
		return nil, ErrEmptyCart
	}

	var total int
	var orderItems []model.OrderItem
	for _, ci := range cartItems {
		lineTotal := ci.ProductPrice * ci.Quantity
		total += lineTotal
		orderItems = append(orderItems, model.OrderItem{
			ProductID:       ci.ProductID,
			Quantity:        ci.Quantity,
			PriceAtPurchase: ci.ProductPrice,
		})
	}

	order, err := s.orderRepo.Create(ctx, userID, total, orderItems)
	if err != nil {
		return nil, err
	}

	if err := s.cartRepo.ClearCart(ctx, userID); err != nil {
		return nil, err
	}

	if err := s.publisher.PublishOrderCreated(order.ID.String()); err != nil {
		// Log but don't fail — order is created, worker will retry
		return order, nil
	}

	return order, nil
}

func (s *OrderService) GetOrder(ctx context.Context, orderID uuid.UUID) (*model.Order, error) {
	return s.orderRepo.FindByID(ctx, orderID)
}

func (s *OrderService) ListOrders(ctx context.Context, userID uuid.UUID) ([]model.Order, error) {
	return s.orderRepo.ListByUser(ctx, userID)
}
```

- [ ] **Step 5: Run tests**

```bash
cd go/ecommerce-service && go test ./internal/service/ -v -run TestCheckout
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add go/ecommerce-service/internal/repository/order.go go/ecommerce-service/internal/service/order.go go/ecommerce-service/internal/service/order_test.go
git commit -m "feat(go): add order repository, service, and checkout flow"
```

### Task 12: RabbitMQ order processor worker

**Files:**
- Create: `go/ecommerce-service/internal/worker/order_processor.go`
- Create: `go/ecommerce-service/internal/worker/order_processor_test.go`

- [ ] **Step 1: Write worker test**

Create `go/ecommerce-service/internal/worker/order_processor_test.go`:

```go
package worker_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/model"
	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/worker"
)

type mockOrderRepoForWorker struct {
	orders map[string]*model.Order
}

func (m *mockOrderRepoForWorker) FindByID(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	o, ok := m.orders[id.String()]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return o, nil
}

func (m *mockOrderRepoForWorker) UpdateStatus(ctx context.Context, id uuid.UUID, status model.OrderStatus) error {
	o, ok := m.orders[id.String()]
	if !ok {
		return fmt.Errorf("not found")
	}
	o.Status = status
	return nil
}

type mockProductRepoForWorker struct {
	stock map[string]int
}

func (m *mockProductRepoForWorker) DecrementStock(ctx context.Context, productID uuid.UUID, qty int) error {
	s, ok := m.stock[productID.String()]
	if !ok {
		return fmt.Errorf("product not found")
	}
	if s < qty {
		return fmt.Errorf("insufficient stock")
	}
	m.stock[productID.String()] = s - qty
	return nil
}

func TestProcessOrder_Success(t *testing.T) {
	productID := uuid.New()
	orderID := uuid.New()

	orderRepo := &mockOrderRepoForWorker{
		orders: map[string]*model.Order{
			orderID.String(): {
				ID:     orderID,
				Status: model.OrderStatusPending,
				Items: []model.OrderItem{
					{ProductID: productID, Quantity: 2},
				},
			},
		},
	}

	productRepo := &mockProductRepoForWorker{
		stock: map[string]int{productID.String(): 10},
	}

	processor := worker.NewOrderProcessor(orderRepo, productRepo, nil)
	err := processor.ProcessOrder(context.Background(), orderID.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	order := orderRepo.orders[orderID.String()]
	if order.Status != model.OrderStatusCompleted {
		t.Errorf("expected completed, got %s", order.Status)
	}
	if productRepo.stock[productID.String()] != 8 {
		t.Errorf("expected stock 8, got %d", productRepo.stock[productID.String()])
	}
}

func TestProcessOrder_InsufficientStock(t *testing.T) {
	productID := uuid.New()
	orderID := uuid.New()

	orderRepo := &mockOrderRepoForWorker{
		orders: map[string]*model.Order{
			orderID.String(): {
				ID:     orderID,
				Status: model.OrderStatusPending,
				Items: []model.OrderItem{
					{ProductID: productID, Quantity: 100},
				},
			},
		},
	}

	productRepo := &mockProductRepoForWorker{
		stock: map[string]int{productID.String(): 5},
	}

	processor := worker.NewOrderProcessor(orderRepo, productRepo, nil)
	err := processor.ProcessOrder(context.Background(), orderID.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	order := orderRepo.orders[orderID.String()]
	if order.Status != model.OrderStatusFailed {
		t.Errorf("expected failed, got %s", order.Status)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd go/ecommerce-service && go test ./internal/worker/ -v
```

Expected: FAIL — `worker.NewOrderProcessor` not defined

- [ ] **Step 3: Write the order processor**

Create `go/ecommerce-service/internal/worker/order_processor.go`:

```go
package worker

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/model"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	ordersProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "orders_total",
		Help: "Total orders processed by status",
	}, []string{"status"})

	messagesProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rabbitmq_messages_processed_total",
		Help: "Total RabbitMQ messages processed",
	}, []string{"result"})
)

type OrderRepoForWorker interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.Order, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.OrderStatus) error
}

type ProductRepoForWorker interface {
	DecrementStock(ctx context.Context, productID uuid.UUID, qty int) error
}

type CacheInvalidator interface {
	InvalidateCache(ctx context.Context)
}

type OrderProcessor struct {
	orderRepo  OrderRepoForWorker
	productRepo ProductRepoForWorker
	cache       CacheInvalidator
}

func NewOrderProcessor(orderRepo OrderRepoForWorker, productRepo ProductRepoForWorker, cache CacheInvalidator) *OrderProcessor {
	return &OrderProcessor{
		orderRepo:  orderRepo,
		productRepo: productRepo,
		cache:       cache,
	}
}

func (p *OrderProcessor) ProcessOrder(ctx context.Context, orderIDStr string) error {
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		return err
	}

	order, err := p.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return err
	}

	if err := p.orderRepo.UpdateStatus(ctx, orderID, model.OrderStatusProcessing); err != nil {
		return err
	}

	for _, item := range order.Items {
		if err := p.productRepo.DecrementStock(ctx, item.ProductID, item.Quantity); err != nil {
			slog.Error("insufficient stock, failing order", "orderId", orderIDStr, "productId", item.ProductID, "error", err)
			p.orderRepo.UpdateStatus(ctx, orderID, model.OrderStatusFailed)
			ordersProcessed.WithLabelValues("failed").Inc()
			return nil
		}
	}

	if err := p.orderRepo.UpdateStatus(ctx, orderID, model.OrderStatusCompleted); err != nil {
		return err
	}

	if p.cache != nil {
		p.cache.InvalidateCache(ctx)
	}

	ordersProcessed.WithLabelValues("completed").Inc()
	return nil
}

func (p *OrderProcessor) StartConsumer(ctx context.Context, ch *amqp.Channel, concurrency int) error {
	err := ch.ExchangeDeclare("ecommerce", "topic", true, false, false, false, nil)
	if err != nil {
		return err
	}

	q, err := ch.QueueDeclare("ecommerce.orders", true, false, false, false, nil)
	if err != nil {
		return err
	}

	if err := ch.QueueBind(q.Name, "order.created", "ecommerce", false, nil); err != nil {
		return err
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for i := 0; i < concurrency; i++ {
		go func(workerID int) {
			slog.Info("order worker started", "workerId", workerID)
			for {
				select {
				case <-ctx.Done():
					slog.Info("order worker stopping", "workerId", workerID)
					return
				case msg, ok := <-msgs:
					if !ok {
						return
					}
					var orderMsg model.OrderMessage
					if err := json.Unmarshal(msg.Body, &orderMsg); err != nil {
						slog.Error("failed to unmarshal message", "error", err)
						msg.Nack(false, false)
						messagesProcessed.WithLabelValues("error").Inc()
						continue
					}

					slog.Info("processing order", "orderId", orderMsg.OrderID, "workerId", workerID)
					if err := p.ProcessOrder(ctx, orderMsg.OrderID); err != nil {
						slog.Error("failed to process order", "orderId", orderMsg.OrderID, "error", err)
						msg.Nack(false, true) // requeue
						messagesProcessed.WithLabelValues("error").Inc()
					} else {
						msg.Ack(false)
						messagesProcessed.WithLabelValues("success").Inc()
					}
				}
			}
		}(i)
	}

	return nil
}
```

- [ ] **Step 4: Run tests**

```bash
cd go/ecommerce-service && go test ./internal/worker/ -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add go/ecommerce-service/internal/worker/
git commit -m "feat(go): add RabbitMQ order processor with worker pool"
```

### Task 13: Ecommerce JWT middleware and handlers

**Files:**
- Create: `go/ecommerce-service/internal/middleware/auth.go`
- Create: `go/ecommerce-service/internal/middleware/logging.go`
- Create: `go/ecommerce-service/internal/middleware/metrics.go`
- Create: `go/ecommerce-service/internal/middleware/cors.go`
- Create: `go/ecommerce-service/internal/handler/product.go`
- Create: `go/ecommerce-service/internal/handler/cart.go`
- Create: `go/ecommerce-service/internal/handler/order.go`
- Create: `go/ecommerce-service/internal/handler/health.go`

- [ ] **Step 1: Create JWT auth middleware**

Create `go/ecommerce-service/internal/middleware/auth.go`:

```go
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims := jwt.MapClaims{}
		_, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid or expired token"})
			return
		}

		sub, ok := claims["sub"].(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid token claims"})
			return
		}

		c.Set("userId", sub)
		c.Next()
	}
}
```

- [ ] **Step 2: Copy logging, metrics, CORS middleware from auth-service**

Create `go/ecommerce-service/internal/middleware/logging.go`, `metrics.go`, `cors.go` — identical to auth-service (Tasks 4). These are separate Go modules so they cannot share code directly.

`go/ecommerce-service/internal/middleware/logging.go`:

```go
package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set("requestId", requestID)
		c.Header("X-Request-ID", requestID)

		start := time.Now()
		c.Next()
		latency := time.Since(start)

		slog.Info("request",
			"requestId", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency", latency.String(),
			"ip", c.ClientIP(),
			"userId", c.GetString("userId"),
		)
	}
}
```

`go/ecommerce-service/internal/middleware/metrics.go`:

```go
package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})
)

func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()

		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}
		status := strconv.Itoa(c.Writer.Status())

		httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}
```

`go/ecommerce-service/internal/middleware/cors.go`:

```go
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func CORS(allowedOrigins string) gin.HandlerFunc {
	origins := strings.Split(allowedOrigins, ",")
	originSet := make(map[string]bool, len(origins))
	for _, o := range origins {
		originSet[strings.TrimSpace(o)] = true
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if originSet[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With")
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
```

- [ ] **Step 3: Create product handler**

Create `go/ecommerce-service/internal/handler/product.go`:

```go
package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/model"
)

type ProductServiceInterface interface {
	List(ctx context.Context, params model.ProductListParams) (*model.ProductListResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error)
	Categories(ctx context.Context) ([]string, error)
}

type ProductHandler struct {
	svc ProductServiceInterface
}

func NewProductHandler(svc ProductServiceInterface) *ProductHandler {
	return &ProductHandler{svc: svc}
}

func (h *ProductHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	params := model.ProductListParams{
		Category: c.Query("category"),
		Sort:     c.Query("sort"),
		Page:     page,
		Limit:    limit,
	}

	resp, err := h.svc.List(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list products"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ProductHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	product, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) Categories(c *gin.Context) {
	cats, err := h.svc.Categories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list categories"})
		return
	}

	c.JSON(http.StatusOK, cats)
}
```

- [ ] **Step 4: Create cart handler**

Create `go/ecommerce-service/internal/handler/cart.go`:

```go
package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/model"
)

type CartServiceInterface interface {
	GetCart(ctx context.Context, userID uuid.UUID) ([]model.CartItem, error)
	AddItem(ctx context.Context, userID, productID uuid.UUID, quantity int) (*model.CartItem, error)
	UpdateQuantity(ctx context.Context, itemID, userID uuid.UUID, quantity int) error
	RemoveItem(ctx context.Context, itemID, userID uuid.UUID) error
}

type CartHandler struct {
	svc CartServiceInterface
}

func NewCartHandler(svc CartServiceInterface) *CartHandler {
	return &CartHandler{svc: svc}
}

func (h *CartHandler) GetCart(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("userId"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	items, err := h.svc.GetCart(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get cart"})
		return
	}

	var total int
	for _, item := range items {
		total += item.ProductPrice * item.Quantity
	}

	c.JSON(http.StatusOK, model.CartResponse{Items: items, Total: total})
}

func (h *CartHandler) AddItem(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("userId"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	var req model.AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	item, err := h.svc.AddItem(c.Request.Context(), userID, productID, req.Quantity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add item"})
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *CartHandler) UpdateQuantity(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("userId"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	itemID, err := uuid.Parse(c.Param("itemId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item ID"})
		return
	}

	var req model.UpdateCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.UpdateQuantity(c.Request.Context(), itemID, userID, req.Quantity); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cart item not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *CartHandler) RemoveItem(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("userId"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	itemID, err := uuid.Parse(c.Param("itemId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item ID"})
		return
	}

	if err := h.svc.RemoveItem(c.Request.Context(), itemID, userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cart item not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "removed"})
}
```

- [ ] **Step 5: Create order handler**

Create `go/ecommerce-service/internal/handler/order.go`:

```go
package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/model"
)

type OrderServiceInterface interface {
	Checkout(ctx context.Context, userID uuid.UUID) (*model.Order, error)
	GetOrder(ctx context.Context, orderID uuid.UUID) (*model.Order, error)
	ListOrders(ctx context.Context, userID uuid.UUID) ([]model.Order, error)
}

type OrderHandler struct {
	svc OrderServiceInterface
}

func NewOrderHandler(svc OrderServiceInterface) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func (h *OrderHandler) Checkout(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("userId"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	order, err := h.svc.Checkout(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) List(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("userId"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	orders, err := h.svc.ListOrders(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list orders"})
		return
	}

	c.JSON(http.StatusOK, orders)
}

func (h *OrderHandler) GetByID(c *gin.Context) {
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	order, err := h.svc.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}
```

- [ ] **Step 6: Create health handler**

Create `go/ecommerce-service/internal/handler/health.go`:

```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	pool  *pgxpool.Pool
	redis *redis.Client
}

func NewHealthHandler(pool *pgxpool.Pool, redisClient *redis.Client) *HealthHandler {
	return &HealthHandler{pool: pool, redis: redisClient}
}

func (h *HealthHandler) Health(c *gin.Context) {
	ctx := c.Request.Context()
	checks := gin.H{}

	if err := h.pool.Ping(ctx); err != nil {
		checks["postgres"] = "unhealthy"
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "checks": checks})
		return
	}
	checks["postgres"] = "healthy"

	if h.redis != nil {
		if err := h.redis.Ping(ctx).Err(); err != nil {
			checks["redis"] = "unhealthy"
		} else {
			checks["redis"] = "healthy"
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "healthy", "checks": checks})
}
```

- [ ] **Step 7: Commit**

```bash
git add go/ecommerce-service/internal/middleware/ go/ecommerce-service/internal/handler/
git commit -m "feat(go): add ecommerce handlers with JWT auth middleware"
```

### Task 14: Ecommerce server entry point and RabbitMQ publisher

**Files:**
- Create: `go/ecommerce-service/cmd/server/main.go`

- [ ] **Step 1: Write server entry point**

Create `go/ecommerce-service/cmd/server/main.go`:

```go
package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"

	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/handler"
	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/middleware"
	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/model"
	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/repository"
	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/service"
	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/worker"
)

type rabbitPublisher struct {
	ch *amqp.Channel
}

func (p *rabbitPublisher) PublishOrderCreated(orderID string) error {
	body, _ := json.Marshal(model.OrderMessage{OrderID: orderID})
	return p.ch.Publish("ecommerce", "order.created", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	dbURL := envOrDefault("DATABASE_URL", "postgres://taskuser:taskpass@localhost:5432/ecommercedb")
	redisURL := envOrDefault("REDIS_URL", "redis://localhost:6379")
	rabbitURL := envOrDefault("RABBITMQ_URL", "amqp://guest:guest@localhost:5672")
	jwtSecret := envOrDefault("JWT_SECRET", "dev-secret-key-at-least-32-characters-long")
	allowedOrigins := envOrDefault("ALLOWED_ORIGINS", "http://localhost:3000")
	port := envOrDefault("PORT", "8092")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Postgres
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()
	slog.Info("connected to database")

	// Redis
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("failed to parse Redis URL: %v", err)
	}
	redisClient := redis.NewClient(opts)
	defer redisClient.Close()
	slog.Info("connected to Redis")

	// RabbitMQ
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to open RabbitMQ channel: %v", err)
	}
	defer ch.Close()
	slog.Info("connected to RabbitMQ")

	// Repositories
	productRepo := repository.NewProductRepository(pool)
	cartRepo := repository.NewCartRepository(pool)
	orderRepo := repository.NewOrderRepository(pool)

	// Services
	productSvc := service.NewProductService(productRepo, redisClient)
	cartSvc := service.NewCartService(cartRepo)
	publisher := &rabbitPublisher{ch: ch}
	orderSvc := service.NewOrderService(orderRepo, cartRepo, publisher)

	// Handlers
	productHandler := handler.NewProductHandler(productSvc)
	cartHandler := handler.NewCartHandler(cartSvc)
	orderHandler := handler.NewOrderHandler(orderSvc)
	healthHandler := handler.NewHealthHandler(pool, redisClient)

	// Worker
	processor := worker.NewOrderProcessor(orderRepo, productRepo, productSvc)
	if err := processor.StartConsumer(ctx, ch, 3); err != nil {
		log.Fatalf("failed to start order worker: %v", err)
	}
	slog.Info("order workers started", "concurrency", 3)

	// Router
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logging())
	r.Use(middleware.Metrics())
	r.Use(middleware.CORS(allowedOrigins))

	// Public routes
	r.GET("/products", productHandler.List)
	r.GET("/products/:id", productHandler.GetByID)
	r.GET("/categories", productHandler.Categories)
	r.GET("/health", healthHandler.Health)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Authenticated routes
	authed := r.Group("/")
	authed.Use(middleware.Auth(jwtSecret))
	authed.GET("/cart", cartHandler.GetCart)
	authed.POST("/cart", cartHandler.AddItem)
	authed.PUT("/cart/:itemId", cartHandler.UpdateQuantity)
	authed.DELETE("/cart/:itemId", cartHandler.RemoveItem)
	authed.POST("/orders", orderHandler.Checkout)
	authed.GET("/orders", orderHandler.List)
	authed.GET("/orders/:id", orderHandler.GetByID)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		slog.Info("ecommerce-service starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down ecommerce-service")

	cancel() // stop workers

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
	slog.Info("ecommerce-service stopped")
}

func envOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
```

- [ ] **Step 2: Verify build compiles**

```bash
cd go/ecommerce-service && go build ./cmd/server/
```

Expected: builds successfully

- [ ] **Step 3: Commit**

```bash
git add go/ecommerce-service/cmd/
git commit -m "feat(go): add ecommerce server entry point with RabbitMQ publisher"
```

### Task 15: Ecommerce Dockerfile

**Files:**
- Create: `go/ecommerce-service/Dockerfile`

- [ ] **Step 1: Create multi-stage Dockerfile**

Create `go/ecommerce-service/Dockerfile`:

```dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /ecommerce-service ./cmd/server

FROM alpine:3.19

RUN adduser -D -u 1001 appuser
COPY --from=builder /ecommerce-service /ecommerce-service
USER appuser

EXPOSE 8092
ENTRYPOINT ["/ecommerce-service"]
```

- [ ] **Step 2: Commit**

```bash
git add go/ecommerce-service/Dockerfile
git commit -m "feat(go): add ecommerce-service Dockerfile"
```

---

## Part 3: Docker Compose and Infrastructure

### Task 16: Docker Compose for local dev

**Files:**
- Create: `go/docker-compose.yml`

- [ ] **Step 1: Create docker-compose.yml**

Create `go/docker-compose.yml`:

```yaml
services:
  postgres:
    image: postgres:17-alpine
    ports:
      - "5433:5432"
    environment:
      POSTGRES_USER: taskuser
      POSTGRES_PASSWORD: taskpass
      POSTGRES_DB: ecommercedb
    volumes:
      - go-pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U taskuser -d ecommercedb"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6380:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5

  rabbitmq:
    image: rabbitmq:3-management-alpine
    ports:
      - "5673:5672"
      - "15673:15672"
    healthcheck:
      test: ["CMD", "rabbitmq-diagnostics", "check_running"]
      interval: 10s
      timeout: 10s
      retries: 5

  auth-service:
    build: ./auth-service
    ports:
      - "8091:8091"
    environment:
      DATABASE_URL: postgres://taskuser:taskpass@postgres:5432/ecommercedb
      JWT_SECRET: dev-secret-key-at-least-32-characters-long
      ALLOWED_ORIGINS: http://localhost:3000
      PORT: "8091"
    depends_on:
      postgres:
        condition: service_healthy

  ecommerce-service:
    build: ./ecommerce-service
    ports:
      - "8092:8092"
    environment:
      DATABASE_URL: postgres://taskuser:taskpass@postgres:5432/ecommercedb
      REDIS_URL: redis://redis:6379
      RABBITMQ_URL: amqp://guest:guest@rabbitmq:5672
      JWT_SECRET: dev-secret-key-at-least-32-characters-long
      ALLOWED_ORIGINS: http://localhost:3000
      PORT: "8092"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
      rabbitmq:
        condition: service_healthy

volumes:
  go-pgdata:
```

- [ ] **Step 2: Commit**

```bash
git add go/docker-compose.yml
git commit -m "feat(go): add Docker Compose for local development"
```

### Task 17: Makefile and preflight targets

**Files:**
- Modify: `Makefile`

- [ ] **Step 1: Add Go preflight targets**

Add to the existing `Makefile`:

```makefile
# Go
preflight-go:
	@echo "==> Running Go linting..."
	cd go/auth-service && golangci-lint run ./...
	cd go/ecommerce-service && golangci-lint run ./...
	@echo "==> Running Go tests..."
	cd go/auth-service && go test ./... -v -race
	cd go/ecommerce-service && go test ./... -v -race
```

Update the `preflight` target to include `preflight-go`.

- [ ] **Step 2: Commit**

```bash
git add Makefile
git commit -m "feat(go): add preflight-go to Makefile"
```

---

## Part 4: Frontend

### Task 18: Go auth utilities and API client

**Files:**
- Create: `frontend/src/lib/go-auth.ts`
- Create: `frontend/src/lib/go-api.ts`

- [ ] **Step 1: Create Go auth utilities**

Create `frontend/src/lib/go-auth.ts`:

```typescript
const ACCESS_TOKEN_KEY = "go_access_token";
const REFRESH_TOKEN_KEY = "go_refresh_token";

export function getGoAccessToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(ACCESS_TOKEN_KEY);
}

export function getGoRefreshToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(REFRESH_TOKEN_KEY);
}

export function setGoTokens(accessToken: string, refreshToken: string): void {
  localStorage.setItem(ACCESS_TOKEN_KEY, accessToken);
  localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken);
}

export function clearGoTokens(): void {
  localStorage.removeItem(ACCESS_TOKEN_KEY);
  localStorage.removeItem(REFRESH_TOKEN_KEY);
}

export function isGoLoggedIn(): boolean {
  return getGoAccessToken() !== null;
}

export const GO_AUTH_URL =
  process.env.NEXT_PUBLIC_GO_AUTH_URL || "http://localhost:8091";
export const GO_ECOMMERCE_URL =
  process.env.NEXT_PUBLIC_GO_ECOMMERCE_URL || "http://localhost:8092";

let refreshPromise: Promise<string | null> | null = null;

export async function refreshGoAccessToken(): Promise<string | null> {
  if (refreshPromise) return refreshPromise;

  refreshPromise = (async () => {
    const refreshToken = getGoRefreshToken();
    if (!refreshToken) return null;

    try {
      const res = await fetch(`${GO_AUTH_URL}/auth/refresh`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ refreshToken }),
      });
      if (!res.ok) {
        clearGoTokens();
        return null;
      }
      const data = await res.json();
      setGoTokens(data.accessToken, data.refreshToken);
      return data.accessToken as string;
    } catch {
      clearGoTokens();
      return null;
    } finally {
      refreshPromise = null;
    }
  })();

  return refreshPromise;
}
```

- [ ] **Step 2: Create API client with token refresh**

Create `frontend/src/lib/go-api.ts`:

```typescript
import {
  getGoAccessToken,
  refreshGoAccessToken,
  GO_ECOMMERCE_URL,
} from "./go-auth";

export async function goApiFetch(
  path: string,
  options: RequestInit = {}
): Promise<Response> {
  let token = getGoAccessToken();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  let res = await fetch(`${GO_ECOMMERCE_URL}${path}`, {
    ...options,
    headers,
  });

  if (res.status === 403 && token) {
    token = await refreshGoAccessToken();
    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
      res = await fetch(`${GO_ECOMMERCE_URL}${path}`, {
        ...options,
        headers,
      });
    }
  }

  return res;
}
```

- [ ] **Step 3: Run TypeScript check**

```bash
cd frontend && npx tsc --noEmit
```

Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add frontend/src/lib/go-auth.ts frontend/src/lib/go-api.ts
git commit -m "feat(frontend): add Go auth utilities and API client"
```

### Task 19: Go landing page and ecommerce layout

**Files:**
- Create: `frontend/src/app/go/page.tsx`
- Create: `frontend/src/app/go/layout.tsx`
- Create: `frontend/src/components/go/GoAuthProvider.tsx`

- [ ] **Step 1: Create Go auth provider**

Create `frontend/src/components/go/GoAuthProvider.tsx`:

```tsx
"use client";

import { createContext, useCallback, useContext, useState } from "react";
import {
  clearGoTokens,
  isGoLoggedIn as checkIsGoLoggedIn,
  setGoTokens,
  GO_AUTH_URL,
} from "@/lib/go-auth";

interface GoAuthUser {
  userId: string;
  email: string;
  name: string;
}

interface GoAuthContextType {
  user: GoAuthUser | null;
  isLoggedIn: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, name: string) => Promise<void>;
  logout: () => void;
}

const GoAuthContext = createContext<GoAuthContextType>({
  user: null,
  isLoggedIn: false,
  login: async () => {},
  register: async () => {},
  logout: () => {},
});

export function useGoAuth() {
  return useContext(GoAuthContext);
}

function handleAuthResponse(data: {
  accessToken: string;
  refreshToken: string;
  userId: string;
  email: string;
  name: string;
}): GoAuthUser {
  setGoTokens(data.accessToken, data.refreshToken);
  const user: GoAuthUser = {
    userId: data.userId,
    email: data.email,
    name: data.name,
  };
  localStorage.setItem("go_user", JSON.stringify(user));
  return user;
}

export function GoAuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<GoAuthUser | null>(() => {
    if (typeof window === "undefined" || !checkIsGoLoggedIn()) return null;
    const stored = localStorage.getItem("go_user");
    return stored ? JSON.parse(stored) : null;
  });
  const [isAuthenticated, setIsAuthenticated] = useState(
    () => typeof window !== "undefined" && checkIsGoLoggedIn()
  );

  const login = useCallback(async (email: string, password: string) => {
    const res = await fetch(`${GO_AUTH_URL}/auth/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password }),
    });
    if (!res.ok) {
      const errorText = await res.text();
      throw new Error(errorText || "Invalid email or password");
    }
    const data = await res.json();
    const authUser = handleAuthResponse(data);
    setUser(authUser);
    setIsAuthenticated(true);
  }, []);

  const register = useCallback(
    async (email: string, password: string, name: string) => {
      const res = await fetch(`${GO_AUTH_URL}/auth/register`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password, name }),
      });
      if (!res.ok) {
        const errorText = await res.text();
        throw new Error(errorText || "Registration failed");
      }
      const data = await res.json();
      const authUser = handleAuthResponse(data);
      setUser(authUser);
      setIsAuthenticated(true);
    },
    []
  );

  const logout = useCallback(() => {
    clearGoTokens();
    localStorage.removeItem("go_user");
    setUser(null);
    setIsAuthenticated(false);
  }, []);

  return (
    <GoAuthContext.Provider
      value={{ user, isLoggedIn: isAuthenticated, login, register, logout }}
    >
      {children}
    </GoAuthContext.Provider>
  );
}
```

- [ ] **Step 2: Create Go layout**

Create `frontend/src/app/go/layout.tsx`:

```tsx
import { GoAuthProvider } from "@/components/go/GoAuthProvider";

export default function GoLayout({ children }: { children: React.ReactNode }) {
  return <GoAuthProvider>{children}</GoAuthProvider>;
}
```

- [ ] **Step 3: Create Go landing page**

Create `frontend/src/app/go/page.tsx`:

```tsx
import Link from "next/link";

export default function GoPage() {
  return (
    <div className="mx-auto max-w-3xl px-6 py-12">
      <h1 className="text-3xl font-bold">Go Backend Developer</h1>
      <p className="mt-4 text-muted-foreground leading-relaxed">
        A microservices ecommerce platform built with Go, demonstrating REST API
        design, JWT authentication, PostgreSQL, Redis caching, RabbitMQ async
        processing, and Prometheus observability.
      </p>

      <div className="mt-6 space-y-2 text-sm text-muted-foreground">
        <p>
          <strong>Services:</strong> Auth Service (Gin, pgx, JWT) + Ecommerce
          Service (Gin, pgx, Redis, RabbitMQ)
        </p>
        <p>
          <strong>Infrastructure:</strong> PostgreSQL, Redis, RabbitMQ, Docker,
          Kubernetes, Prometheus/Grafana
        </p>
        <p>
          <strong>Testing:</strong> Unit tests, integration tests, benchmark
          tests with go test
        </p>
      </div>

      <div className="mt-8">
        <Link
          href="/go/ecommerce"
          className="inline-block rounded-md bg-primary px-6 py-3 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
        >
          Open Store &rarr;
        </Link>
      </div>
    </div>
  );
}
```

- [ ] **Step 4: Run TypeScript check and lint**

```bash
cd frontend && npx tsc --noEmit && npm run lint
```

Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add frontend/src/app/go/ frontend/src/components/go/
git commit -m "feat(frontend): add Go landing page and auth provider"
```

### Task 20: Ecommerce storefront pages

**Files:**
- Create: `frontend/src/app/go/ecommerce/page.tsx`
- Create: `frontend/src/app/go/ecommerce/[productId]/page.tsx`
- Create: `frontend/src/app/go/ecommerce/cart/page.tsx`
- Create: `frontend/src/app/go/ecommerce/orders/page.tsx`
- Create: `frontend/src/components/go/ProductCard.tsx`

- [ ] **Step 1: Create ProductCard component**

Create `frontend/src/components/go/ProductCard.tsx`:

```tsx
import Link from "next/link";

interface ProductCardProps {
  id: string;
  name: string;
  price: number;
  category: string;
  imageUrl: string;
}

export function ProductCard({
  id,
  name,
  price,
  category,
  imageUrl,
}: ProductCardProps) {
  return (
    <Link href={`/go/ecommerce/${id}`} className="block group">
      <div className="rounded-lg border border-foreground/10 p-4 hover:ring-1 hover:ring-foreground/20 transition-all">
        <div className="aspect-square rounded-md bg-muted flex items-center justify-center text-muted-foreground text-xs">
          {imageUrl || "No image"}
        </div>
        <div className="mt-3">
          <p className="text-xs text-muted-foreground">{category}</p>
          <h3 className="font-medium">{name}</h3>
          <p className="mt-1 font-semibold">${(price / 100).toFixed(2)}</p>
        </div>
      </div>
    </Link>
  );
}
```

- [ ] **Step 2: Create storefront page (product grid)**

Create `frontend/src/app/go/ecommerce/page.tsx`:

```tsx
"use client";

import { useEffect, useState } from "react";
import { useGoAuth } from "@/components/go/GoAuthProvider";
import { ProductCard } from "@/components/go/ProductCard";
import { GO_ECOMMERCE_URL } from "@/lib/go-auth";

interface Product {
  id: string;
  name: string;
  price: number;
  category: string;
  imageUrl: string;
}

export default function EcommercePage() {
  const { isLoggedIn, login, register, logout } = useGoAuth();
  const [products, setProducts] = useState<Product[]>([]);
  const [categories, setCategories] = useState<string[]>([]);
  const [selectedCategory, setSelectedCategory] = useState("");
  const [loading, setLoading] = useState(true);

  // Auth form state
  const [authView, setAuthView] = useState<"login" | "register">("login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [name, setName] = useState("");
  const [authError, setAuthError] = useState("");

  useEffect(() => {
    fetch(`${GO_ECOMMERCE_URL}/categories`)
      .then((r) => r.json())
      .then(setCategories)
      .catch(() => {});
  }, []);

  useEffect(() => {
    setLoading(true);
    const params = new URLSearchParams();
    if (selectedCategory) params.set("category", selectedCategory);
    fetch(`${GO_ECOMMERCE_URL}/products?${params}`)
      .then((r) => r.json())
      .then((data) => {
        setProducts(data.products || []);
        setLoading(false);
      })
      .catch(() => setLoading(false));
  }, [selectedCategory]);

  const handleAuth = async (e: React.FormEvent) => {
    e.preventDefault();
    setAuthError("");
    try {
      if (authView === "login") {
        await login(email, password);
      } else {
        await register(email, password, name);
      }
    } catch (err) {
      setAuthError(err instanceof Error ? err.message : "Auth failed");
    }
  };

  return (
    <div className="mx-auto max-w-5xl px-6 py-12">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Store</h1>
        {isLoggedIn ? (
          <div className="flex gap-3">
            <a
              href="/go/ecommerce/cart"
              className="text-sm text-muted-foreground hover:text-foreground"
            >
              Cart
            </a>
            <a
              href="/go/ecommerce/orders"
              className="text-sm text-muted-foreground hover:text-foreground"
            >
              Orders
            </a>
            <button
              onClick={logout}
              className="text-sm text-muted-foreground hover:text-foreground"
            >
              Sign out
            </button>
          </div>
        ) : (
          <form onSubmit={handleAuth} className="flex items-center gap-2">
            {authView === "register" && (
              <input
                type="text"
                placeholder="Name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="rounded border px-2 py-1 text-sm"
              />
            )}
            <input
              type="email"
              placeholder="Email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              className="rounded border px-2 py-1 text-sm"
            />
            <input
              type="password"
              placeholder="Password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              className="rounded border px-2 py-1 text-sm"
            />
            <button
              type="submit"
              className="rounded bg-primary px-3 py-1 text-sm text-primary-foreground"
            >
              {authView === "login" ? "Sign in" : "Register"}
            </button>
            <button
              type="button"
              onClick={() =>
                setAuthView(authView === "login" ? "register" : "login")
              }
              className="text-xs text-muted-foreground hover:text-foreground"
            >
              {authView === "login" ? "Register" : "Sign in"}
            </button>
            {authError && (
              <span className="text-xs text-red-500">{authError}</span>
            )}
          </form>
        )}
      </div>

      <div className="mt-6 flex gap-2">
        <button
          onClick={() => setSelectedCategory("")}
          className={`rounded-full px-3 py-1 text-sm ${!selectedCategory ? "bg-primary text-primary-foreground" : "bg-muted text-muted-foreground"}`}
        >
          All
        </button>
        {categories.map((cat) => (
          <button
            key={cat}
            onClick={() => setSelectedCategory(cat)}
            className={`rounded-full px-3 py-1 text-sm ${selectedCategory === cat ? "bg-primary text-primary-foreground" : "bg-muted text-muted-foreground"}`}
          >
            {cat}
          </button>
        ))}
      </div>

      {loading ? (
        <p className="mt-8 text-muted-foreground">Loading...</p>
      ) : (
        <div className="mt-6 grid grid-cols-2 gap-4 md:grid-cols-4">
          {products.map((p) => (
            <ProductCard key={p.id} {...p} />
          ))}
        </div>
      )}
    </div>
  );
}
```

- [ ] **Step 3: Create product detail page**

Create `frontend/src/app/go/ecommerce/[productId]/page.tsx`:

```tsx
"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { useGoAuth } from "@/components/go/GoAuthProvider";
import { goApiFetch } from "@/lib/go-api";
import { GO_ECOMMERCE_URL } from "@/lib/go-auth";

interface Product {
  id: string;
  name: string;
  description: string;
  price: number;
  category: string;
  stock: number;
}

export default function ProductDetailPage() {
  const params = useParams();
  const productId = params.productId as string;
  const { isLoggedIn } = useGoAuth();
  const [product, setProduct] = useState<Product | null>(null);
  const [loading, setLoading] = useState(true);
  const [addedToCart, setAddedToCart] = useState(false);

  useEffect(() => {
    fetch(`${GO_ECOMMERCE_URL}/products/${productId}`)
      .then((r) => r.json())
      .then((data) => {
        setProduct(data);
        setLoading(false);
      })
      .catch(() => setLoading(false));
  }, [productId]);

  const addToCart = async () => {
    const res = await goApiFetch("/cart", {
      method: "POST",
      body: JSON.stringify({ productId, quantity: 1 }),
    });
    if (res.ok) {
      setAddedToCart(true);
      setTimeout(() => setAddedToCart(false), 2000);
    }
  };

  if (loading) {
    return (
      <div className="mx-auto max-w-3xl px-6 py-12">
        <p className="text-muted-foreground">Loading...</p>
      </div>
    );
  }

  if (!product) {
    return (
      <div className="mx-auto max-w-3xl px-6 py-12">
        <p>Product not found.</p>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-3xl px-6 py-12">
      <Link
        href="/go/ecommerce"
        className="text-sm text-muted-foreground hover:text-foreground"
      >
        &larr; Back to Store
      </Link>

      <div className="mt-6">
        <p className="text-sm text-muted-foreground">{product.category}</p>
        <h1 className="text-2xl font-bold">{product.name}</h1>
        <p className="mt-2 text-muted-foreground">{product.description}</p>
        <p className="mt-4 text-2xl font-bold">
          ${(product.price / 100).toFixed(2)}
        </p>
        <p className="mt-1 text-sm text-muted-foreground">
          {product.stock > 0 ? `${product.stock} in stock` : "Out of stock"}
        </p>

        {isLoggedIn && product.stock > 0 && (
          <button
            onClick={addToCart}
            className="mt-6 rounded-md bg-primary px-6 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
          >
            {addedToCart ? "Added!" : "Add to Cart"}
          </button>
        )}
      </div>
    </div>
  );
}
```

- [ ] **Step 4: Create cart page**

Create `frontend/src/app/go/ecommerce/cart/page.tsx`:

```tsx
"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { goApiFetch } from "@/lib/go-api";

interface CartItem {
  id: string;
  productId: string;
  quantity: number;
  productName: string;
  productPrice: number;
}

export default function CartPage() {
  const router = useRouter();
  const [items, setItems] = useState<CartItem[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);

  const loadCart = async () => {
    const res = await goApiFetch("/cart");
    if (res.ok) {
      const data = await res.json();
      setItems(data.items || []);
      setTotal(data.total || 0);
    }
    setLoading(false);
  };

  useEffect(() => {
    loadCart();
  }, []);

  const removeItem = async (itemId: string) => {
    await goApiFetch(`/cart/${itemId}`, { method: "DELETE" });
    loadCart();
  };

  const checkout = async () => {
    const res = await goApiFetch("/orders", { method: "POST" });
    if (res.ok) {
      router.push("/go/ecommerce/orders");
    }
  };

  if (loading) {
    return (
      <div className="mx-auto max-w-3xl px-6 py-12">
        <p className="text-muted-foreground">Loading...</p>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-3xl px-6 py-12">
      <Link
        href="/go/ecommerce"
        className="text-sm text-muted-foreground hover:text-foreground"
      >
        &larr; Back to Store
      </Link>
      <h1 className="mt-4 text-2xl font-bold">Cart</h1>

      {items.length === 0 ? (
        <p className="mt-6 text-muted-foreground">Your cart is empty.</p>
      ) : (
        <>
          <div className="mt-6 space-y-4">
            {items.map((item) => (
              <div
                key={item.id}
                className="flex items-center justify-between rounded-lg border p-4"
              >
                <div>
                  <p className="font-medium">{item.productName}</p>
                  <p className="text-sm text-muted-foreground">
                    Qty: {item.quantity} &times; $
                    {(item.productPrice / 100).toFixed(2)}
                  </p>
                </div>
                <div className="flex items-center gap-4">
                  <p className="font-semibold">
                    ${((item.productPrice * item.quantity) / 100).toFixed(2)}
                  </p>
                  <button
                    onClick={() => removeItem(item.id)}
                    className="text-sm text-red-500 hover:text-red-400"
                  >
                    Remove
                  </button>
                </div>
              </div>
            ))}
          </div>

          <div className="mt-6 flex items-center justify-between border-t pt-4">
            <p className="text-lg font-bold">
              Total: ${(total / 100).toFixed(2)}
            </p>
            <button
              onClick={checkout}
              className="rounded-md bg-primary px-6 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
            >
              Checkout
            </button>
          </div>
        </>
      )}
    </div>
  );
}
```

- [ ] **Step 5: Create orders page**

Create `frontend/src/app/go/ecommerce/orders/page.tsx`:

```tsx
"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { goApiFetch } from "@/lib/go-api";

interface Order {
  id: string;
  status: string;
  total: number;
  createdAt: string;
}

export default function OrdersPage() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    goApiFetch("/orders")
      .then((r) => r.json())
      .then((data) => {
        setOrders(data || []);
        setLoading(false);
      })
      .catch(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div className="mx-auto max-w-3xl px-6 py-12">
        <p className="text-muted-foreground">Loading...</p>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-3xl px-6 py-12">
      <Link
        href="/go/ecommerce"
        className="text-sm text-muted-foreground hover:text-foreground"
      >
        &larr; Back to Store
      </Link>
      <h1 className="mt-4 text-2xl font-bold">Order History</h1>

      {orders.length === 0 ? (
        <p className="mt-6 text-muted-foreground">No orders yet.</p>
      ) : (
        <div className="mt-6 space-y-4">
          {orders.map((order) => (
            <div
              key={order.id}
              className="flex items-center justify-between rounded-lg border p-4"
            >
              <div>
                <p className="font-medium">
                  Order {order.id.slice(0, 8)}...
                </p>
                <p className="text-sm text-muted-foreground">
                  {new Date(order.createdAt).toLocaleDateString()}
                </p>
              </div>
              <div className="text-right">
                <p className="font-semibold">
                  ${(order.total / 100).toFixed(2)}
                </p>
                <p
                  className={`text-sm ${
                    order.status === "completed"
                      ? "text-green-500"
                      : order.status === "failed"
                        ? "text-red-500"
                        : "text-yellow-500"
                  }`}
                >
                  {order.status}
                </p>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
```

- [ ] **Step 6: Run TypeScript check and lint**

```bash
cd frontend && npx tsc --noEmit && npm run lint
```

Expected: no errors

- [ ] **Step 7: Commit**

```bash
git add frontend/src/app/go/ecommerce/ frontend/src/components/go/ProductCard.tsx
git commit -m "feat(frontend): add ecommerce storefront pages"
```

### Task 21: Update SiteHeader navigation

**Files:**
- Modify: `frontend/src/components/SiteHeader.tsx`

- [ ] **Step 1: Add Go link to navigation**

Add a "Go" navigation link alongside the existing "AI" and "Java" links in the SiteHeader component. Follow the existing pattern for how the other section links are structured.

- [ ] **Step 2: Run lint and type check**

```bash
cd frontend && npx tsc --noEmit && npm run lint
```

Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/SiteHeader.tsx
git commit -m "feat(frontend): add Go section to site navigation"
```

---

## Part 5: CI/CD and K8s

### Task 22: Go CI workflow

**Files:**
- Create: `.github/workflows/go-ci.yml`

- [ ] **Step 1: Create Go CI workflow**

Create `.github/workflows/go-ci.yml`:

```yaml
name: Go CI

on:
  push:
    branches: ["**"]
    paths:
      - "go/**"
      - ".github/workflows/go-ci.yml"
  pull_request:
    branches: [main]
    paths:
      - "go/**"

jobs:
  lint:
    name: Lint (${{ matrix.service }})
    runs-on: ubuntu-latest
    strategy:
      matrix:
        service: [auth-service, ecommerce-service]
    defaults:
      run:
        working-directory: go/${{ matrix.service }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
      - name: Install golangci-lint
        run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
      - name: Run linter
        run: golangci-lint run ./...

  test:
    name: Test (${{ matrix.service }})
    runs-on: ubuntu-latest
    strategy:
      matrix:
        service: [auth-service, ecommerce-service]
    defaults:
      run:
        working-directory: go/${{ matrix.service }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
      - name: Run tests
        run: go test ./... -v -race -coverprofile=coverage.out
      - name: Upload coverage
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: coverage-go-${{ matrix.service }}
          path: go/${{ matrix.service }}/coverage.out

  build:
    name: Docker Build (${{ matrix.service }})
    runs-on: ubuntu-latest
    needs: [lint, test]
    permissions:
      packages: write
    strategy:
      matrix:
        service: [auth-service, ecommerce-service]
    steps:
      - uses: actions/checkout@v4
      - uses: docker/setup-buildx-action@v3
      - name: Log in to GHCR
        if: github.ref == 'refs/heads/main' && github.event_name == 'push'
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - name: Build and push
        uses: docker/build-push-action@v6
        with:
          context: go/${{ matrix.service }}
          push: ${{ github.ref == 'refs/heads/main' && github.event_name == 'push' }}
          tags: ghcr.io/${{ github.repository }}/go-${{ matrix.service }}:latest
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/go-ci.yml
git commit -m "feat(ci): add Go CI workflow with lint, test, Docker build"
```

### Task 23: K8s manifests

**Files:**
- Create: `go/k8s/namespace.yml`
- Create: `go/k8s/configmaps/auth-service-config.yml`
- Create: `go/k8s/configmaps/ecommerce-service-config.yml`
- Create: `go/k8s/deployments/auth-service.yml`
- Create: `go/k8s/deployments/ecommerce-service.yml`
- Create: `go/k8s/services/auth-service.yml`
- Create: `go/k8s/services/ecommerce-service.yml`
- Create: `go/k8s/ingress.yml`

- [ ] **Step 1: Create namespace**

Create `go/k8s/namespace.yml`:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: go-ecommerce
```

- [ ] **Step 2: Create configmaps**

Create `go/k8s/configmaps/auth-service-config.yml`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: auth-service-config
  namespace: go-ecommerce
data:
  DATABASE_URL: postgres://taskuser:taskpass@postgres.java-tasks.svc.cluster.local:5432/ecommercedb
  ALLOWED_ORIGINS: http://localhost:3000,https://kylebradshaw.dev
  PORT: "8091"
```

Create `go/k8s/configmaps/ecommerce-service-config.yml`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: ecommerce-service-config
  namespace: go-ecommerce
data:
  DATABASE_URL: postgres://taskuser:taskpass@postgres.java-tasks.svc.cluster.local:5432/ecommercedb
  REDIS_URL: redis://redis.java-tasks.svc.cluster.local:6379
  RABBITMQ_URL: amqp://guest:guest@rabbitmq.java-tasks.svc.cluster.local:5672
  ALLOWED_ORIGINS: http://localhost:3000,https://kylebradshaw.dev
  PORT: "8092"
```

- [ ] **Step 3: Create deployments**

Create `go/k8s/deployments/auth-service.yml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: go-auth-service
  namespace: go-ecommerce
spec:
  replicas: 1
  selector:
    matchLabels:
      app: go-auth-service
  template:
    metadata:
      labels:
        app: go-auth-service
    spec:
      imagePullSecrets:
        - name: ghcr-secret
      containers:
        - name: go-auth-service
          image: ghcr.io/kabradshaw1/portfolio/go-auth-service:latest
          imagePullPolicy: Always
          ports:
            - containerPort: 8091
          envFrom:
            - configMapRef:
                name: auth-service-config
          env:
            - name: JWT_SECRET
              valueFrom:
                secretKeyRef:
                  name: go-secrets
                  key: jwt-secret
          resources:
            requests:
              memory: "64Mi"
              cpu: "100m"
            limits:
              memory: "128Mi"
              cpu: "250m"
          readinessProbe:
            httpGet:
              path: /health
              port: 8091
            initialDelaySeconds: 5
            periodSeconds: 10
```

Create `go/k8s/deployments/ecommerce-service.yml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: go-ecommerce-service
  namespace: go-ecommerce
spec:
  replicas: 1
  selector:
    matchLabels:
      app: go-ecommerce-service
  template:
    metadata:
      labels:
        app: go-ecommerce-service
    spec:
      imagePullSecrets:
        - name: ghcr-secret
      containers:
        - name: go-ecommerce-service
          image: ghcr.io/kabradshaw1/portfolio/go-ecommerce-service:latest
          imagePullPolicy: Always
          ports:
            - containerPort: 8092
          envFrom:
            - configMapRef:
                name: ecommerce-service-config
          env:
            - name: JWT_SECRET
              valueFrom:
                secretKeyRef:
                  name: go-secrets
                  key: jwt-secret
          resources:
            requests:
              memory: "64Mi"
              cpu: "100m"
            limits:
              memory: "256Mi"
              cpu: "500m"
          readinessProbe:
            httpGet:
              path: /health
              port: 8092
            initialDelaySeconds: 5
            periodSeconds: 10
```

- [ ] **Step 4: Create services**

Create `go/k8s/services/auth-service.yml`:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: go-auth-service
  namespace: go-ecommerce
spec:
  selector:
    app: go-auth-service
  ports:
    - port: 8091
      targetPort: 8091
```

Create `go/k8s/services/ecommerce-service.yml`:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: go-ecommerce-service
  namespace: go-ecommerce
spec:
  selector:
    app: go-ecommerce-service
  ports:
    - port: 8092
      targetPort: 8092
```

- [ ] **Step 5: Create ingress**

Create `go/k8s/ingress.yml`:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: go-ecommerce-ingress
  namespace: go-ecommerce
  annotations:
    nginx.ingress.kubernetes.io/use-regex: "true"
spec:
  ingressClassName: nginx
  rules:
    - http:
        paths:
          - path: /go-auth/(.*)
            pathType: ImplementationSpecific
            backend:
              service:
                name: go-auth-service
                port:
                  number: 8091
          - path: /go-api/(.*)
            pathType: ImplementationSpecific
            backend:
              service:
                name: go-ecommerce-service
                port:
                  number: 8092
```

- [ ] **Step 6: Commit**

```bash
git add go/k8s/
git commit -m "feat(k8s): add Go ecommerce K8s manifests"
```

### Task 24: Update CI deploy step and deploy.sh

**Files:**
- Modify: `.github/workflows/ci.yml`
- Modify: `k8s/deploy.sh`

- [ ] **Step 1: Add Go manifests to CI deploy step**

Add a new line after the `java/k8s` manifest apply in the deploy step:

```bash
for f in $(find go/k8s -name '*.yml' -not -path '*/secrets/*'); do echo '---'; cat "$f"; done | $SSH "kubectl apply -f -"
```

Also add rollout restart and status for the go-ecommerce namespace.

- [ ] **Step 2: Update deploy.sh for Go services**

Add Go namespace creation, manifest apply, and wait steps to `k8s/deploy.sh` following the existing pattern.

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/ci.yml k8s/deploy.sh
git commit -m "feat(ci): add Go ecommerce services to deploy pipeline"
```

---

## Part 6: Benchmark Tests

### Task 25: Benchmark tests for hot paths

**Files:**
- Create: `go/ecommerce-service/internal/service/product_bench_test.go`
- Create: `go/auth-service/internal/service/auth_bench_test.go`

- [ ] **Step 1: Write product listing benchmark**

Create `go/ecommerce-service/internal/service/product_bench_test.go`:

```go
package service_test

import (
	"context"
	"testing"

	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/model"
	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/service"
)

func BenchmarkProductList(b *testing.B) {
	repo := newMockProductRepo()
	svc := service.NewProductService(repo, nil)
	ctx := context.Background()
	params := model.ProductListParams{Page: 1, Limit: 20}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.List(ctx, params)
	}
}

func BenchmarkProductListByCategory(b *testing.B) {
	repo := newMockProductRepo()
	svc := service.NewProductService(repo, nil)
	ctx := context.Background()
	params := model.ProductListParams{Category: "Electronics", Page: 1, Limit: 20}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.List(ctx, params)
	}
}
```

- [ ] **Step 2: Write JWT validation benchmark**

Create `go/auth-service/internal/service/auth_bench_test.go`:

```go
package service_test

import (
	"context"
	"testing"

	"github.com/kabradshaw1/portfolio/go/auth-service/internal/service"
)

func BenchmarkRegister(b *testing.B) {
	repo := newMockUserRepo()
	svc := service.NewAuthService(repo, "test-secret-at-least-32-characters-long!!", 900000, 604800000)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.users = make(map[string]*model.User) // reset
		svc.Register(ctx, "bench@test.com", "password123", "Bench User")
	}
}

func BenchmarkLogin(b *testing.B) {
	repo := newMockUserRepo()
	svc := service.NewAuthService(repo, "test-secret-at-least-32-characters-long!!", 900000, 604800000)
	ctx := context.Background()
	svc.Register(ctx, "bench@test.com", "password123", "Bench User")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.Login(ctx, "bench@test.com", "password123")
	}
}
```

- [ ] **Step 3: Run benchmarks**

```bash
cd go/auth-service && go test ./internal/service/ -bench=. -benchmem
cd go/ecommerce-service && go test ./internal/service/ -bench=. -benchmem
```

- [ ] **Step 4: Commit**

```bash
git add go/auth-service/internal/service/auth_bench_test.go go/ecommerce-service/internal/service/product_bench_test.go
git commit -m "feat(go): add benchmark tests for auth and product services"
```

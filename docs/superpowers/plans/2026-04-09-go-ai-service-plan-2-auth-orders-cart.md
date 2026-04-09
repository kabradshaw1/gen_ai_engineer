# Plan 2 — `go/ai-service` Auth + Order/Cart Tools + Returns Endpoint

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend `go/ai-service` with JWT authentication and six user-scoped tools (`list_orders`, `get_order`, `summarize_orders`, `view_cart`, `add_to_cart`, `initiate_return`), and add the one new `ecommerce-service` endpoint the returns tool needs. After Plan 2 the agent can operate on a logged-in user's real orders, cart, and return flow.

**Architecture:** JWT is validated at the ai-service HTTP layer using the same HS256 `JWT_SECRET` `ecommerce-service` and `auth-service` share. The userId from the `sub` claim is threaded through `agent.Turn.UserID` to every tool. User-scoped tools forward the original Bearer token to `ecommerce-service`, so the backend remains the sole authorization authority. A new returns table, repo, service, and `POST /orders/:id/returns` endpoint land in `ecommerce-service`. A pre-existing authz gap on `GET /orders/:id` (no ownership check) is fixed as part of this plan because the agent amplifies its blast radius.

**Tech Stack:** Go 1.26, Gin, `github.com/golang-jwt/jwt/v5` (already in go.sum via ecommerce-service), golang-migrate for the new returns table, same `httptest`-based testing pattern as Plan 1.

**Scope boundaries:**
- No Redis caching, no evals, no metrics, no guardrails beyond the existing input-length check (→ Plan 3).
- No frontend (→ Plan 4).
- No K8s/CI changes (→ Plan 5).
- `summarize_orders` is kept with its sub-LLM call as agreed in the spec, but the sub-call reuses the parent turn's `context.Context` and therefore its wall-clock deadline. No separate budget.
- No `place_order` / checkout tool. The agent still cannot move money.
- JWT validation uses a shared HS256 secret. No JWKS, no RS256, no key rotation. Matches `auth-service`/`ecommerce-service` today.

**Reference:** spec `2026-04-09-go-ai-service-agent-design.md` sections 2.3 (auth), 4.2 (tool catalog), 4.3 (new endpoints), 4.4 (design rules).

**Module path:** `github.com/kabradshaw1/portfolio/go/ai-service`

---

## File Map

New files:

```
go/ai-service/
├── internal/
│   ├── auth/
│   │   ├── jwt.go            # ParseBearer(jwt, secret) -> userID, error
│   │   └── jwt_test.go
│   ├── http/
│   │   └── chat.go           # (modified) reads Authorization header, calls auth.ParseBearer, threads JWT into Turn
│   ├── tools/
│   │   ├── orders.go         # list_orders, get_order, summarize_orders
│   │   ├── orders_test.go
│   │   ├── cart.go           # view_cart, add_to_cart
│   │   ├── cart_test.go
│   │   ├── returns.go        # initiate_return
│   │   ├── returns_test.go
│   │   └── clients/
│   │       └── ecommerce.go  # (modified) new authenticated methods
│   │       └── ecommerce_test.go  # (modified) add auth-header assertions

go/ecommerce-service/
├── migrations/
│   ├── 004_create_returns.up.sql
│   └── 004_create_returns.down.sql
├── internal/
│   ├── model/
│   │   └── returns.go        # Return, ReturnRequest, ReturnResponse
│   ├── repository/
│   │   ├── return.go
│   │   └── return_test.go    # (optional — skip if no existing repo integration tests)
│   ├── service/
│   │   └── return.go
│   ├── handler/
│   │   ├── return.go
│   │   └── return_test.go
│   │   └── order.go          # (modified) GET /orders/:id ownership check
│   │   └── order_test.go     # (new or extended) ownership test
│   └── repository/
│       └── order.go          # (modified) ensure GetByID returns user_id so handler can check
```

Modified files:
- `go/ai-service/internal/http/chat.go`
- `go/ai-service/internal/tools/clients/ecommerce.go`
- `go/ai-service/internal/tools/clients/ecommerce_test.go`
- `go/ai-service/cmd/server/main.go` — wire JWT secret, register new tools
- `go/ai-service/go.mod` / `go.sum` — add `github.com/golang-jwt/jwt/v5`
- `go/ecommerce-service/cmd/server/main.go` — register returns routes

---

## Shared conventions for this plan

All new ai-service tools take dependencies via small interfaces defined in their own file, mirroring the `ecommerceAPI` pattern from Plan 1's `catalog.go`. This keeps tests fake-friendly without exposing the full `*clients.EcommerceClient` surface to every tool.

All user-scoped ecommerce client methods take the JWT as their last string argument — e.g. `ListOrders(ctx, jwt string)` — rather than storing a token on the client struct. The client is stateless and safe to share across goroutines.

Tools receive the JWT via a new argument on `Tool.Call`? **No — we don't change the `Tool` interface.** Instead, each user-scoped tool is constructed with a small "token source" that reads the current request's JWT from context. The HTTP handler stashes the JWT on the `context.Context` under a private key; tools extract it in their `Call` method. This keeps the `Tool` interface stable and keeps future MCP-adapter work trivial.

The context helper lives in `internal/http/context.go`:

```go
package http

import "context"

type ctxKey int

const jwtKey ctxKey = iota

func ContextWithJWT(ctx context.Context, jwt string) context.Context {
	return context.WithValue(ctx, jwtKey, jwt)
}

func JWTFromContext(ctx context.Context) string {
	v, _ := ctx.Value(jwtKey).(string)
	return v
}
```

User-scoped tools read the JWT with `http.JWTFromContext(ctx)`. Tests pass a context pre-populated with `http.ContextWithJWT(ctx, "test-token")`.

This is the one shared piece. Everything else is task-local.

---

## Task 1: JWT parsing in ai-service + context helper

**Files:**
- Create: `go/ai-service/internal/auth/jwt.go`
- Create: `go/ai-service/internal/auth/jwt_test.go`
- Create: `go/ai-service/internal/http/context.go`

**Dependency:** add `github.com/golang-jwt/jwt/v5` to ai-service `go.mod`.

- [ ] **Step 1: Add the dependency**

```bash
cd go/ai-service && go get github.com/golang-jwt/jwt/v5@v5.3.1
```

Expected: updates `go.mod` and `go.sum`, no errors.

- [ ] **Step 2: Write `internal/http/context.go`**

```go
package http

import "context"

type ctxKey int

const jwtKey ctxKey = iota

// ContextWithJWT returns a new context that carries the user's bearer token.
// User-scoped tools extract it with JWTFromContext.
func ContextWithJWT(ctx context.Context, jwt string) context.Context {
	return context.WithValue(ctx, jwtKey, jwt)
}

// JWTFromContext returns the bearer token attached by ContextWithJWT, or "".
func JWTFromContext(ctx context.Context) string {
	v, _ := ctx.Value(jwtKey).(string)
	return v
}
```

- [ ] **Step 3: Write the failing test for `auth.ParseBearer`**

Contents of `go/ai-service/internal/auth/jwt_test.go`:

```go
package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func mintToken(t *testing.T, secret, sub string, exp time.Duration) string {
	t.Helper()
	claims := jwt.MapClaims{
		"sub": sub,
		"exp": time.Now().Add(exp).Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := tok.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return s
}

func TestParseBearer_Valid(t *testing.T) {
	secret := "dev-secret-key-at-least-32-characters-long"
	tok := mintToken(t, secret, "11111111-1111-1111-1111-111111111111", time.Hour)

	userID, err := ParseBearer("Bearer "+tok, secret)
	if err != nil {
		t.Fatalf("ParseBearer: %v", err)
	}
	if userID != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("userID = %q", userID)
	}
}

func TestParseBearer_Missing(t *testing.T) {
	if _, err := ParseBearer("", "secret"); err == nil {
		t.Error("expected error for empty header")
	}
}

func TestParseBearer_WrongScheme(t *testing.T) {
	if _, err := ParseBearer("Basic abc", "secret"); err == nil {
		t.Error("expected error for non-Bearer scheme")
	}
}

func TestParseBearer_BadSignature(t *testing.T) {
	tok := mintToken(t, "right-secret-key-at-least-32-characters-long", "u1", time.Hour)
	if _, err := ParseBearer("Bearer "+tok, "wrong-secret-key-at-least-32-characters-long"); err == nil {
		t.Error("expected signature error")
	}
}

func TestParseBearer_Expired(t *testing.T) {
	secret := "dev-secret-key-at-least-32-characters-long"
	tok := mintToken(t, secret, "u1", -time.Minute)
	if _, err := ParseBearer("Bearer "+tok, secret); err == nil {
		t.Error("expected expiration error")
	}
}

func TestParseBearer_NoSubClaim(t *testing.T) {
	secret := "dev-secret-key-at-least-32-characters-long"
	claims := jwt.MapClaims{"exp": time.Now().Add(time.Hour).Unix()}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := tok.SignedString([]byte(secret))
	if _, err := ParseBearer("Bearer "+s, secret); err == nil {
		t.Error("expected missing-sub error")
	}
}
```

- [ ] **Step 4: Run, expect compile failure**

```bash
cd go/ai-service && go test ./internal/auth/...
```

Expected: `undefined: ParseBearer`.

- [ ] **Step 5: Implement `internal/auth/jwt.go`**

```go
package auth

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// ParseBearer validates an HTTP "Authorization: Bearer <token>" header using
// the provided HS256 secret and returns the "sub" claim (userID).
// It matches the exact same scheme used by ecommerce-service's auth middleware
// so tokens minted by auth-service are accepted as-is.
func ParseBearer(header, secret string) (string, error) {
	if header == "" {
		return "", errors.New("missing authorization header")
	}
	if !strings.HasPrefix(header, "Bearer ") {
		return "", errors.New("authorization header must start with Bearer")
	}
	tokenStr := strings.TrimPrefix(header, "Bearer ")

	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return "", fmt.Errorf("parse token: %w", err)
	}
	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", errors.New("token missing sub claim")
	}
	return sub, nil
}
```

- [ ] **Step 6: Run tests, expect pass**

```bash
cd go/ai-service && go test ./internal/auth/... -v
```

Expected: 6/6 PASS.

- [ ] **Step 7: Commit**

```bash
git add go/ai-service/
git commit -m "feat(ai-service): add JWT bearer parsing and JWT context helper"
```

---

## Task 2: Thread JWT through `POST /chat` into `agent.Turn.UserID`

**Files:**
- Modify: `go/ai-service/internal/http/chat.go`
- Modify: `go/ai-service/internal/http/chat_test.go`

Auth is **optional** on `/chat`: catalog-only conversations work without a token. When a token is present, it must validate. When a token is present and valid, the `userID` is set on the turn and the raw token is stashed on the context so tools can forward it.

- [ ] **Step 1: Extend `RegisterChatRoutes` to accept a secret and parse tokens**

New signature:
```go
func RegisterChatRoutes(r *gin.Engine, runner Runner, jwtSecret string)
```

Inside the handler, after parsing the request body:

```go
var userID string
authHeader := c.GetHeader("Authorization")
if authHeader != "" {
    uid, err := auth.ParseBearer(authHeader, jwtSecret)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }
    userID = uid
}

// Build a ctx that carries the raw bearer so forwarding tools can use it.
ctx := c.Request.Context()
if authHeader != "" {
    ctx = ContextWithJWT(ctx, strings.TrimPrefix(authHeader, "Bearer "))
}

turn := agent.Turn{UserID: userID, Messages: req.Messages}
if err := runner.Run(ctx, turn, emit); err != nil {
    emit(agent.Event{Error: &agent.ErrorEvent{Reason: err.Error()}})
}
```

Add `"strings"` and the `auth` import. The full updated file: keep everything from Plan 1 Task 8, just replace the handler closure with the above and change the function signature. `jwtSecret` may be `""` during tests that don't care about auth — when empty, the code path that calls `auth.ParseBearer` is still exercised only if a header was sent.

Critical: if `jwtSecret == ""` AND a bearer header is sent, `auth.ParseBearer` will fail because the signature can't validate. That's fine — it's a test misconfiguration, not something to silently paper over. In production `main.go` always supplies the secret.

- [ ] **Step 2: Update existing chat tests to the new signature**

In `chat_test.go`, all `RegisterChatRoutes(r, runner)` call sites become `RegisterChatRoutes(r, runner, "")`. The existing tests don't send Authorization headers, so nothing else changes.

Add two new tests:

```go
func TestChatHandler_AcceptsValidJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "dev-secret-key-at-least-32-characters-long"
	// Sign a token using the same shape as auth-service.
	tok := mintChatTestToken(t, secret, "user-abc", time.Hour)

	var capturedTurn agent.Turn
	var capturedJWT string
	runner := &fakeRunner{events: []agent.Event{{Final: &agent.FinalEvent{Text: "ok"}}}}
	runner.onRun = func(ctx context.Context, turn agent.Turn) {
		capturedTurn = turn
		capturedJWT = JWTFromContext(ctx)
	}

	r := gin.New()
	RegisterChatRoutes(r, runner, secret)
	req := httptest.NewRequest(http.MethodPost, "/chat",
		strings.NewReader(`{"messages":[{"role":"user","content":"hi"}]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tok)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", w.Code, w.Body.String())
	}
	if capturedTurn.UserID != "user-abc" {
		t.Errorf("turn.UserID = %q", capturedTurn.UserID)
	}
	if capturedJWT != tok {
		t.Errorf("ctx jwt = %q", capturedJWT)
	}
}

func TestChatHandler_RejectsInvalidJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterChatRoutes(r, &fakeRunner{}, "dev-secret-key-at-least-32-characters-long")

	req := httptest.NewRequest(http.MethodPost, "/chat",
		strings.NewReader(`{"messages":[{"role":"user","content":"hi"}]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer not-a-real-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// mintChatTestToken mirrors the helper in auth/jwt_test.go.
func mintChatTestToken(t *testing.T, secret, sub string, d time.Duration) string {
	t.Helper()
	claims := jwt.MapClaims{"sub": sub, "exp": time.Now().Add(d).Unix()}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := tok.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return s
}
```

And extend `fakeRunner` in the same file:

```go
type fakeRunner struct {
	events []agent.Event
	err    error
	onRun  func(ctx context.Context, turn agent.Turn)
}

func (f *fakeRunner) Run(ctx context.Context, turn agent.Turn, emit func(agent.Event)) error {
	if f.onRun != nil {
		f.onRun(ctx, turn)
	}
	for _, e := range f.events {
		emit(e)
	}
	return f.err
}
```

Imports to add at the top of `chat_test.go`:
```go
import (
    // ... existing imports ...
    "time"
    "github.com/golang-jwt/jwt/v5"
)
```

- [ ] **Step 3: Run tests**

```bash
cd go/ai-service && go test ./internal/http/... -v
```

Expected: all http tests PASS, including the two new ones.

- [ ] **Step 4: Commit**

```bash
git add go/ai-service/internal/http/
git commit -m "feat(ai-service): validate JWT on /chat and thread userID into Turn"
```

---

## Task 3: Fix `GET /orders/:id` ownership check in `ecommerce-service`

**Pre-existing security gap:** `handler/order.go` `GetByID` calls `h.svc.GetOrder(ctx, orderID)` and returns the result with no check that the authenticated user owns the order. Any logged-in user can read any order by UUID. The agent amplifies this — making this a one-step fix is in scope for Plan 2.

**Files:**
- Modify: `go/ecommerce-service/internal/handler/order.go`
- Create or extend: `go/ecommerce-service/internal/handler/order_test.go`

- [ ] **Step 1: Write the failing test**

If `order_test.go` doesn't exist, create it (mirror the pattern from Plan 1's `product_test.go`):

```go
package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/model"
)

type fakeOrderService struct {
	order *model.Order
}

func (f *fakeOrderService) Checkout(ctx context.Context, userID uuid.UUID) (*model.Order, error) {
	return nil, nil
}
func (f *fakeOrderService) GetOrder(ctx context.Context, orderID uuid.UUID) (*model.Order, error) {
	return f.order, nil
}
func (f *fakeOrderService) ListOrders(ctx context.Context, userID uuid.UUID) ([]model.Order, error) {
	return nil, nil
}

func TestOrderHandler_GetByID_ForbidsOtherUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	owner := uuid.New()
	other := uuid.New()
	orderID := uuid.New()
	svc := &fakeOrderService{order: &model.Order{ID: orderID, UserID: owner}}
	h := NewOrderHandler(svc)

	r := gin.New()
	r.GET("/orders/:id", func(c *gin.Context) {
		c.Set("userId", other.String())
		h.GetByID(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/orders/"+orderID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for non-owner, got %d", w.Code)
	}
}

func TestOrderHandler_GetByID_AllowsOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)
	owner := uuid.New()
	orderID := uuid.New()
	svc := &fakeOrderService{order: &model.Order{ID: orderID, UserID: owner}}
	h := NewOrderHandler(svc)

	r := gin.New()
	r.GET("/orders/:id", func(c *gin.Context) {
		c.Set("userId", owner.String())
		h.GetByID(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/orders/"+orderID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for owner, got %d", w.Code)
	}
}
```

Verify the `Order` model has a `UserID uuid.UUID` field before writing the test. If the field name is different (e.g. `UserId`), adjust. Read `go/ecommerce-service/internal/model/order.go` first.

- [ ] **Step 2: Run, expect the forbid test to fail with 200**

```bash
cd go/ecommerce-service && go test ./internal/handler/... -run TestOrderHandler_GetByID_ForbidsOtherUsers -v
```

Expected: FAIL — handler currently returns 200 regardless of owner.

- [ ] **Step 3: Fix the handler**

Replace `GetByID` in `go/ecommerce-service/internal/handler/order.go`:

```go
func (h *OrderHandler) GetByID(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	order, err := h.svc.GetOrder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	// Ownership check: 404 (not 403) to avoid leaking existence of other users' orders.
	if order.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}
```

- [ ] **Step 4: Run both ownership tests**

```bash
cd go/ecommerce-service && go test ./internal/handler/... -run "TestOrderHandler_GetByID" -v
```

Expected: both PASS.

- [ ] **Step 5: Run the full ecommerce suite**

```bash
cd go/ecommerce-service && go test ./... -count=1
```

Expected: all PASS, no regressions.

- [ ] **Step 6: Commit**

```bash
git add go/ecommerce-service/internal/handler/
git commit -m "fix(ecommerce): enforce order ownership on GET /orders/:id"
```

---

## Task 4: Extend ecommerce client with JWT-forwarding methods

**Files:**
- Modify: `go/ai-service/internal/tools/clients/ecommerce.go`
- Modify: `go/ai-service/internal/tools/clients/ecommerce_test.go`

New methods on `EcommerceClient`, each taking `jwt string` as the last argument and forwarding it as `Authorization: Bearer <jwt>`:

- `ListOrders(ctx, jwt) ([]Order, error)` — hits `GET /orders`, decodes `{"orders":[...]}` envelope
- `GetOrder(ctx, jwt, orderID) (Order, error)` — hits `GET /orders/{id}`
- `GetCart(ctx, jwt) (Cart, error)` — hits `GET /cart`, decodes `CartResponse`
- `AddToCart(ctx, jwt, productID, qty) (CartItem, error)` — hits `POST /cart`, JSON body
- `InitiateReturn(ctx, jwt, orderID, itemIDs, reason) (Return, error)` — hits `POST /orders/{id}/returns` (endpoint added in Task 7)

**Types the client needs to expose.** Only the fields the tools actually use — do not mirror every ecommerce-service model field.

- [ ] **Step 1: Write the failing test**

Add to `go/ai-service/internal/tools/clients/ecommerce_test.go`:

```go
func TestEcommerceClient_ListOrders_ForwardsJWT(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/orders" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("auth header = %q", got)
		}
		_, _ = w.Write([]byte(`{"orders":[
			{"id":"00000000-0000-0000-0000-000000000001","status":"paid","total":12999,"createdAt":"2026-04-01T00:00:00Z"},
			{"id":"00000000-0000-0000-0000-000000000002","status":"pending","total":8900,"createdAt":"2026-04-02T00:00:00Z"}
		]}`))
	}))
	defer server.Close()

	c := NewEcommerceClient(server.URL)
	orders, err := c.ListOrders(context.Background(), "test-token")
	if err != nil {
		t.Fatalf("ListOrders: %v", err)
	}
	if len(orders) != 2 || orders[0].Status != "paid" || orders[0].Total != 12999 {
		t.Errorf("orders = %+v", orders)
	}
}

func TestEcommerceClient_GetOrder_404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	_, err := NewEcommerceClient(server.URL).GetOrder(context.Background(), "t", "id-x")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestEcommerceClient_GetCart(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cart" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer t" {
			t.Fatal("missing auth")
		}
		_, _ = w.Write([]byte(`{
			"items":[
				{"id":"i1","productId":"p1","productName":"Jacket","productPrice":12999,"quantity":1}
			],
			"total":12999
		}`))
	}))
	defer server.Close()

	cart, err := NewEcommerceClient(server.URL).GetCart(context.Background(), "t")
	if err != nil {
		t.Fatalf("GetCart: %v", err)
	}
	if cart.Total != 12999 || len(cart.Items) != 1 || cart.Items[0].ProductName != "Jacket" {
		t.Errorf("cart = %+v", cart)
	}
}

func TestEcommerceClient_AddToCart_BodyShape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/cart" {
			t.Fatalf("method/path = %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer t" {
			t.Fatal("missing auth")
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["productId"] != "p1" || body["quantity"].(float64) != 2 {
			t.Errorf("body = %+v", body)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"i1","productId":"p1","productName":"Jacket","productPrice":12999,"quantity":2}`))
	}))
	defer server.Close()

	item, err := NewEcommerceClient(server.URL).AddToCart(context.Background(), "t", "p1", 2)
	if err != nil {
		t.Fatalf("AddToCart: %v", err)
	}
	if item.Quantity != 2 || item.ProductName != "Jacket" {
		t.Errorf("item = %+v", item)
	}
}

func TestEcommerceClient_InitiateReturn(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		if r.URL.Path != "/orders/order-1/returns" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer t" {
			t.Fatal("missing auth")
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["reason"] != "doesn't fit" {
			t.Errorf("body = %+v", body)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"r1","orderId":"order-1","status":"requested","reason":"doesn't fit"}`))
	}))
	defer server.Close()

	ret, err := NewEcommerceClient(server.URL).InitiateReturn(context.Background(), "t", "order-1", []string{"i1"}, "doesn't fit")
	if err != nil {
		t.Fatalf("InitiateReturn: %v", err)
	}
	if ret.Status != "requested" || ret.ID != "r1" {
		t.Errorf("ret = %+v", ret)
	}
}
```

Add `"encoding/json"` to the test imports if not already there.

- [ ] **Step 2: Run, expect compile failure**

```bash
cd go/ai-service && go test ./internal/tools/clients/...
```

Expected: undefined types/methods.

- [ ] **Step 3: Extend `ecommerce.go`**

Add to `go/ai-service/internal/tools/clients/ecommerce.go` (keep all existing types and methods):

```go
// ---------- user-scoped types ----------

type Order struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Total     int       `json:"total"` // cents
	CreatedAt time.Time `json:"createdAt"`
}

type ordersResponse struct {
	Orders []Order `json:"orders"`
}

type CartItem struct {
	ID           string `json:"id"`
	ProductID    string `json:"productId"`
	ProductName  string `json:"productName"`
	ProductPrice int    `json:"productPrice"` // cents
	Quantity     int    `json:"quantity"`
}

type Cart struct {
	Items []CartItem `json:"items"`
	Total int        `json:"total"` // cents
}

type Return struct {
	ID      string `json:"id"`
	OrderID string `json:"orderId"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
}

// ---------- authenticated helper ----------

// authedRequest builds a request with the standard headers and a bearer token.
func (c *EcommerceClient) authedRequest(ctx context.Context, method, path string, body io.Reader, jwt string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if jwt != "" {
		req.Header.Set("Authorization", "Bearer "+jwt)
	}
	return req, nil
}

func (c *EcommerceClient) doJSON(req *http.Request, out any) error {
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		payload, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s %s: status %d: %s", req.Method, req.URL.Path, resp.StatusCode, string(payload))
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// ---------- methods ----------

func (c *EcommerceClient) ListOrders(ctx context.Context, jwt string) ([]Order, error) {
	req, err := c.authedRequest(ctx, http.MethodGet, "/orders", nil, jwt)
	if err != nil {
		return nil, err
	}
	var envelope ordersResponse
	if err := c.doJSON(req, &envelope); err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	return envelope.Orders, nil
}

func (c *EcommerceClient) GetOrder(ctx context.Context, jwt, orderID string) (Order, error) {
	req, err := c.authedRequest(ctx, http.MethodGet, "/orders/"+url.PathEscape(orderID), nil, jwt)
	if err != nil {
		return Order{}, err
	}
	var o Order
	if err := c.doJSON(req, &o); err != nil {
		return Order{}, fmt.Errorf("get order: %w", err)
	}
	return o, nil
}

func (c *EcommerceClient) GetCart(ctx context.Context, jwt string) (Cart, error) {
	req, err := c.authedRequest(ctx, http.MethodGet, "/cart", nil, jwt)
	if err != nil {
		return Cart{}, err
	}
	var cart Cart
	if err := c.doJSON(req, &cart); err != nil {
		return Cart{}, fmt.Errorf("get cart: %w", err)
	}
	return cart, nil
}

func (c *EcommerceClient) AddToCart(ctx context.Context, jwt, productID string, qty int) (CartItem, error) {
	body, _ := json.Marshal(map[string]any{"productId": productID, "quantity": qty})
	req, err := c.authedRequest(ctx, http.MethodPost, "/cart", bytes.NewReader(body), jwt)
	if err != nil {
		return CartItem{}, err
	}
	var item CartItem
	if err := c.doJSON(req, &item); err != nil {
		return CartItem{}, fmt.Errorf("add to cart: %w", err)
	}
	return item, nil
}

func (c *EcommerceClient) InitiateReturn(ctx context.Context, jwt, orderID string, itemIDs []string, reason string) (Return, error) {
	body, _ := json.Marshal(map[string]any{
		"itemIds": itemIDs,
		"reason":  reason,
	})
	req, err := c.authedRequest(ctx, http.MethodPost, "/orders/"+url.PathEscape(orderID)+"/returns", bytes.NewReader(body), jwt)
	if err != nil {
		return Return{}, err
	}
	var ret Return
	if err := c.doJSON(req, &ret); err != nil {
		return Return{}, fmt.Errorf("initiate return: %w", err)
	}
	return ret, nil
}
```

Add `"bytes"` and `"time"` to the imports if not already present.

- [ ] **Step 4: Run tests, expect all PASS**

```bash
cd go/ai-service && go test ./internal/tools/clients/... -v
```

Expected: all new and existing client tests PASS.

- [ ] **Step 5: Commit**

```bash
git add go/ai-service/internal/tools/clients/
git commit -m "feat(ai-service): add JWT-forwarding order/cart/return methods to ecommerce client"
```

---

## Task 5: `orders.go` tools — `list_orders`, `get_order`

**Files:**
- Create: `go/ai-service/internal/tools/orders.go`
- Create: `go/ai-service/internal/tools/orders_test.go`

`summarize_orders` is deferred to Task 6 because it has a different constructor signature (needs an `llm.Client`).

- [ ] **Step 1: Failing test**

Contents of `go/ai-service/internal/tools/orders_test.go`:

```go
package tools

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	apphttp "github.com/kabradshaw1/portfolio/go/ai-service/internal/http"
	"github.com/kabradshaw1/portfolio/go/ai-service/internal/tools/clients"
)

type fakeOrdersAPI struct {
	listOut  []clients.Order
	listErr  error
	getOut   clients.Order
	getErr   error
	seenJWT  string
	seenID   string
}

func (f *fakeOrdersAPI) ListOrders(ctx context.Context, jwt string) ([]clients.Order, error) {
	f.seenJWT = jwt
	return f.listOut, f.listErr
}

func (f *fakeOrdersAPI) GetOrder(ctx context.Context, jwt, id string) (clients.Order, error) {
	f.seenJWT = jwt
	f.seenID = id
	return f.getOut, f.getErr
}

func ctxWithJWT(jwt string) context.Context {
	return apphttp.ContextWithJWT(context.Background(), jwt)
}

func TestListOrdersTool_BoundedAndForwardsJWT(t *testing.T) {
	fake := &fakeOrdersAPI{listOut: make([]clients.Order, 50)}
	for i := range fake.listOut {
		fake.listOut[i] = clients.Order{
			ID:        "o" + string(rune('a'+i%26)),
			Status:    "paid",
			Total:     int(100 + i),
			CreatedAt: time.Now(),
		}
	}
	tool := NewListOrdersTool(fake)

	res, err := tool.Call(ctxWithJWT("tok"), json.RawMessage(`{}`), "user-1")
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if fake.seenJWT != "tok" {
		t.Errorf("jwt forwarded = %q", fake.seenJWT)
	}
	items := res.Content.([]map[string]any)
	if len(items) > 20 {
		t.Errorf("expected bound of 20, got %d", len(items))
	}
}

func TestListOrdersTool_RequiresUserID(t *testing.T) {
	tool := NewListOrdersTool(&fakeOrdersAPI{})
	_, err := tool.Call(ctxWithJWT("tok"), json.RawMessage(`{}`), "")
	if err == nil {
		t.Fatal("expected error when userID empty")
	}
}

func TestGetOrderTool_Success(t *testing.T) {
	fake := &fakeOrdersAPI{getOut: clients.Order{ID: "order-1", Status: "paid", Total: 12999}}
	tool := NewGetOrderTool(fake)

	res, err := tool.Call(ctxWithJWT("tok"), json.RawMessage(`{"order_id":"order-1"}`), "user-1")
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if fake.seenID != "order-1" || fake.seenJWT != "tok" {
		t.Errorf("seen id=%q jwt=%q", fake.seenID, fake.seenJWT)
	}
	m := res.Content.(map[string]any)
	if m["id"] != "order-1" || m["status"] != "paid" {
		t.Errorf("content = %+v", m)
	}
}

func TestGetOrderTool_MissingID(t *testing.T) {
	tool := NewGetOrderTool(&fakeOrdersAPI{})
	_, err := tool.Call(ctxWithJWT("tok"), json.RawMessage(`{}`), "user-1")
	if err == nil {
		t.Fatal("expected error for missing order_id")
	}
}

func TestGetOrderTool_RequiresUserID(t *testing.T) {
	tool := NewGetOrderTool(&fakeOrdersAPI{})
	_, err := tool.Call(ctxWithJWT("tok"), json.RawMessage(`{"order_id":"x"}`), "")
	if err == nil {
		t.Fatal("expected error for empty userID")
	}
}
```

- [ ] **Step 2: Run, expect compile failure**

```bash
cd go/ai-service && go test ./internal/tools/... -run TestListOrders -v
```

Expected: undefined symbols.

- [ ] **Step 3: Implement `orders.go`**

```go
package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	apphttp "github.com/kabradshaw1/portfolio/go/ai-service/internal/http"
	"github.com/kabradshaw1/portfolio/go/ai-service/internal/tools/clients"
)

// ordersAPI is the subset of the ecommerce client the order tools use.
type ordersAPI interface {
	ListOrders(ctx context.Context, jwt string) ([]clients.Order, error)
	GetOrder(ctx context.Context, jwt, orderID string) (clients.Order, error)
}

const maxListedOrders = 20

// ---------- list_orders ----------

type listOrdersTool struct{ api ordersAPI }

func NewListOrdersTool(api ordersAPI) Tool { return &listOrdersTool{api: api} }

func (t *listOrdersTool) Name() string { return "list_orders" }
func (t *listOrdersTool) Description() string {
	return "List the current user's orders. Returns at most 20 most recent orders with id, status, total in cents, and creation date."
}
func (t *listOrdersTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type":"object",
		"properties":{
			"limit":{"type":"integer","description":"Max orders to return (cap 20)."}
		}
	}`)
}

type listOrdersArgs struct {
	Limit int `json:"limit"`
}

func (t *listOrdersTool) Call(ctx context.Context, args json.RawMessage, userID string) (Result, error) {
	if userID == "" {
		return Result{}, errors.New("list_orders: authenticated user required")
	}
	var a listOrdersArgs
	_ = json.Unmarshal(args, &a) // empty args are fine
	limit := a.Limit
	if limit <= 0 || limit > maxListedOrders {
		limit = maxListedOrders
	}

	orders, err := t.api.ListOrders(ctx, apphttp.JWTFromContext(ctx))
	if err != nil {
		return Result{}, fmt.Errorf("list_orders: %w", err)
	}
	if len(orders) > limit {
		orders = orders[:limit]
	}

	out := make([]map[string]any, 0, len(orders))
	for _, o := range orders {
		out = append(out, map[string]any{
			"id":         o.ID,
			"status":     o.Status,
			"total":      o.Total,
			"created_at": o.CreatedAt,
		})
	}
	return Result{
		Content: out,
		Display: map[string]any{"kind": "order_list", "orders": out},
	}, nil
}

// ---------- get_order ----------

type getOrderTool struct{ api ordersAPI }

func NewGetOrderTool(api ordersAPI) Tool { return &getOrderTool{api: api} }

func (t *getOrderTool) Name() string { return "get_order" }
func (t *getOrderTool) Description() string {
	return "Fetch one order by id for the current user. Returns id, status, total, and creation date."
}
func (t *getOrderTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type":"object",
		"properties":{
			"order_id":{"type":"string"}
		},
		"required":["order_id"]
	}`)
}

type getOrderArgs struct {
	OrderID string `json:"order_id"`
}

func (t *getOrderTool) Call(ctx context.Context, args json.RawMessage, userID string) (Result, error) {
	if userID == "" {
		return Result{}, errors.New("get_order: authenticated user required")
	}
	var a getOrderArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return Result{}, fmt.Errorf("get_order: bad args: %w", err)
	}
	if a.OrderID == "" {
		return Result{}, errors.New("get_order: order_id is required")
	}

	o, err := t.api.GetOrder(ctx, apphttp.JWTFromContext(ctx), a.OrderID)
	if err != nil {
		return Result{}, fmt.Errorf("get_order: %w", err)
	}
	content := map[string]any{
		"id":         o.ID,
		"status":     o.Status,
		"total":      o.Total,
		"created_at": o.CreatedAt,
	}
	return Result{
		Content: content,
		Display: map[string]any{"kind": "order_card", "order": o},
	}, nil
}
```

- [ ] **Step 4: Run tests, expect all PASS**

```bash
cd go/ai-service && go test ./internal/tools/... -v
```

Expected: all tool tests (catalog + orders) PASS.

- [ ] **Step 5: Commit**

```bash
git add go/ai-service/internal/tools/orders.go go/ai-service/internal/tools/orders_test.go
git commit -m "feat(ai-service): add list_orders and get_order tools"
```

---

## Task 6: `summarize_orders` tool with sub-LLM call

**Files:**
- Modify: `go/ai-service/internal/tools/orders.go`
- Modify: `go/ai-service/internal/tools/orders_test.go`

This is the one tool in Plan 2 that does a sub-LLM call: it fetches the last 20 orders, constructs a short prompt, calls `llm.Client.Chat` without tools, and returns the text.

- [ ] **Step 1: Failing test**

Append to `orders_test.go`:

```go
// fakeLLM reused inside tools tests. Named differently from the agent's fakeLLM
// to avoid confusion if the test files ever share a package (they don't today).
type summarizerLLM struct {
	reply   string
	err     error
	seenMsg []llm.Message
}

func (f *summarizerLLM) Chat(ctx context.Context, msgs []llm.Message, tools []llm.ToolSchema) (llm.ChatResponse, error) {
	f.seenMsg = msgs
	if f.err != nil {
		return llm.ChatResponse{}, f.err
	}
	return llm.ChatResponse{Content: f.reply}, nil
}

func TestSummarizeOrdersTool_Success(t *testing.T) {
	fakeAPI := &fakeOrdersAPI{listOut: []clients.Order{
		{ID: "o1", Status: "paid", Total: 12999, CreatedAt: time.Now().Add(-48 * time.Hour)},
		{ID: "o2", Status: "paid", Total: 8900, CreatedAt: time.Now().Add(-24 * time.Hour)},
	}}
	fakeLLM := &summarizerLLM{reply: "You placed 2 orders totaling $218.99."}
	tool := NewSummarizeOrdersTool(fakeAPI, fakeLLM)

	res, err := tool.Call(ctxWithJWT("tok"), json.RawMessage(`{}`), "user-1")
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	m := res.Content.(map[string]any)
	if m["summary"] != "You placed 2 orders totaling $218.99." {
		t.Errorf("summary = %+v", m)
	}
	if len(fakeLLM.seenMsg) == 0 {
		t.Error("expected at least one message sent to sub-LLM")
	}
	// Sanity: no tools advertised on the sub-call (verified by calling with nil schemas — the fake just ignores)
}

func TestSummarizeOrdersTool_NoOrders(t *testing.T) {
	fakeAPI := &fakeOrdersAPI{listOut: nil}
	fakeLLM := &summarizerLLM{reply: "should not be called"}
	tool := NewSummarizeOrdersTool(fakeAPI, fakeLLM)

	res, err := tool.Call(ctxWithJWT("tok"), json.RawMessage(`{}`), "user-1")
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	m := res.Content.(map[string]any)
	if m["summary"] != "You have no orders yet." {
		t.Errorf("summary = %+v", m)
	}
	if fakeLLM.seenMsg != nil {
		t.Error("expected sub-LLM to be skipped on empty order list")
	}
}

func TestSummarizeOrdersTool_RequiresUserID(t *testing.T) {
	tool := NewSummarizeOrdersTool(&fakeOrdersAPI{}, &summarizerLLM{})
	_, err := tool.Call(ctxWithJWT("tok"), json.RawMessage(`{}`), "")
	if err == nil {
		t.Fatal("expected error")
	}
}
```

Add import: `"github.com/kabradshaw1/portfolio/go/ai-service/internal/llm"` to `orders_test.go`.

- [ ] **Step 2: Run, expect compile failure**

```bash
cd go/ai-service && go test ./internal/tools/... -run TestSummarize
```

Expected: `undefined: NewSummarizeOrdersTool`.

- [ ] **Step 3: Implement in `orders.go`**

Append:

```go
// ---------- summarize_orders ----------

type summarizeOrdersTool struct {
	api ordersAPI
	llm llm.Client
}

// NewSummarizeOrdersTool builds a tool that lists the user's recent orders and
// asks a small sub-LLM call to summarize them. It reuses the parent turn's
// context so the agent's wall-clock timeout still covers the sub-call.
func NewSummarizeOrdersTool(api ordersAPI, llmc llm.Client) Tool {
	return &summarizeOrdersTool{api: api, llm: llmc}
}

func (t *summarizeOrdersTool) Name() string { return "summarize_orders" }
func (t *summarizeOrdersTool) Description() string {
	return "Summarize the current user's recent orders in plain English."
}
func (t *summarizeOrdersTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type":"object",
		"properties":{
			"period":{"type":"string","enum":["week","month","all"]}
		}
	}`)
}

type summarizeArgs struct {
	Period string `json:"period"`
}

func (t *summarizeOrdersTool) Call(ctx context.Context, args json.RawMessage, userID string) (Result, error) {
	if userID == "" {
		return Result{}, errors.New("summarize_orders: authenticated user required")
	}
	var a summarizeArgs
	_ = json.Unmarshal(args, &a)

	orders, err := t.api.ListOrders(ctx, apphttp.JWTFromContext(ctx))
	if err != nil {
		return Result{}, fmt.Errorf("summarize_orders: %w", err)
	}
	if len(orders) > maxListedOrders {
		orders = orders[:maxListedOrders]
	}
	if len(orders) == 0 {
		out := map[string]any{"summary": "You have no orders yet."}
		return Result{Content: out, Display: out}, nil
	}

	orderJSON, _ := json.Marshal(orders)
	prompt := fmt.Sprintf(
		"Summarize these %d orders for the user in two or three sentences. "+
			"Totals are in cents — convert to dollars. Period requested: %q. Orders JSON: %s",
		len(orders), a.Period, string(orderJSON),
	)

	resp, err := t.llm.Chat(ctx, []llm.Message{
		{Role: llm.RoleUser, Content: prompt},
	}, nil)
	if err != nil {
		return Result{}, fmt.Errorf("summarize_orders: sub-llm: %w", err)
	}
	out := map[string]any{"summary": resp.Content, "order_count": len(orders)}
	return Result{Content: out, Display: out}, nil
}
```

Add the `llm` import to the existing `orders.go` imports:

```go
import (
    "context"
    "encoding/json"
    "errors"
    "fmt"

    apphttp "github.com/kabradshaw1/portfolio/go/ai-service/internal/http"
    "github.com/kabradshaw1/portfolio/go/ai-service/internal/llm"
    "github.com/kabradshaw1/portfolio/go/ai-service/internal/tools/clients"
)
```

- [ ] **Step 4: Run tests**

```bash
cd go/ai-service && go test ./internal/tools/... -v
```

Expected: all order tests (including the three new summarize tests) PASS.

- [ ] **Step 5: Commit**

```bash
git add go/ai-service/internal/tools/orders.go go/ai-service/internal/tools/orders_test.go
git commit -m "feat(ai-service): add summarize_orders tool with sub-LLM call"
```

---

## Task 7: `cart.go` tools — `view_cart`, `add_to_cart`

**Files:**
- Create: `go/ai-service/internal/tools/cart.go`
- Create: `go/ai-service/internal/tools/cart_test.go`

Same pattern as Task 5. Small interface, fake in tests, JWT pulled from context, userID enforced.

- [ ] **Step 1: Failing test**

Contents of `go/ai-service/internal/tools/cart_test.go`:

```go
package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/kabradshaw1/portfolio/go/ai-service/internal/tools/clients"
)

type fakeCartAPI struct {
	cart     clients.Cart
	item     clients.CartItem
	err      error
	seenJWT  string
	seenID   string
	seenQty  int
}

func (f *fakeCartAPI) GetCart(ctx context.Context, jwt string) (clients.Cart, error) {
	f.seenJWT = jwt
	return f.cart, f.err
}

func (f *fakeCartAPI) AddToCart(ctx context.Context, jwt, productID string, qty int) (clients.CartItem, error) {
	f.seenJWT = jwt
	f.seenID = productID
	f.seenQty = qty
	return f.item, f.err
}

func TestViewCartTool(t *testing.T) {
	fake := &fakeCartAPI{cart: clients.Cart{
		Items: []clients.CartItem{{ID: "i1", ProductID: "p1", ProductName: "Jacket", ProductPrice: 12999, Quantity: 1}},
		Total: 12999,
	}}
	tool := NewViewCartTool(fake)

	res, err := tool.Call(ctxWithJWT("tok"), json.RawMessage(`{}`), "user-1")
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if fake.seenJWT != "tok" {
		t.Errorf("jwt forwarded = %q", fake.seenJWT)
	}
	m := res.Content.(map[string]any)
	if m["total"].(int) != 12999 {
		t.Errorf("total = %+v", m["total"])
	}
}

func TestViewCartTool_RequiresUserID(t *testing.T) {
	tool := NewViewCartTool(&fakeCartAPI{})
	_, err := tool.Call(ctxWithJWT("tok"), json.RawMessage(`{}`), "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAddToCartTool_Success(t *testing.T) {
	fake := &fakeCartAPI{item: clients.CartItem{ID: "i1", ProductID: "p1", ProductName: "Jacket", ProductPrice: 12999, Quantity: 2}}
	tool := NewAddToCartTool(fake)

	res, err := tool.Call(ctxWithJWT("tok"), json.RawMessage(`{"product_id":"p1","qty":2}`), "user-1")
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if fake.seenID != "p1" || fake.seenQty != 2 {
		t.Errorf("seen id=%q qty=%d", fake.seenID, fake.seenQty)
	}
	m := res.Content.(map[string]any)
	if m["quantity"].(int) != 2 {
		t.Errorf("content = %+v", m)
	}
}

func TestAddToCartTool_BadArgs(t *testing.T) {
	tool := NewAddToCartTool(&fakeCartAPI{})
	if _, err := tool.Call(ctxWithJWT("tok"), json.RawMessage(`{"product_id":""}`), "user-1"); err == nil {
		t.Error("expected error for empty product_id")
	}
	if _, err := tool.Call(ctxWithJWT("tok"), json.RawMessage(`{"product_id":"p1","qty":0}`), "user-1"); err == nil {
		t.Error("expected error for zero qty")
	}
	if _, err := tool.Call(ctxWithJWT("tok"), json.RawMessage(`{"product_id":"p1","qty":-1}`), "user-1"); err == nil {
		t.Error("expected error for negative qty")
	}
}
```

- [ ] **Step 2: Run, expect compile failure**

```bash
cd go/ai-service && go test ./internal/tools/... -run TestViewCart
```

- [ ] **Step 3: Implement `cart.go`**

```go
package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	apphttp "github.com/kabradshaw1/portfolio/go/ai-service/internal/http"
	"github.com/kabradshaw1/portfolio/go/ai-service/internal/tools/clients"
)

// cartAPI is the subset of the ecommerce client the cart tools use.
type cartAPI interface {
	GetCart(ctx context.Context, jwt string) (clients.Cart, error)
	AddToCart(ctx context.Context, jwt, productID string, qty int) (clients.CartItem, error)
}

// ---------- view_cart ----------

type viewCartTool struct{ api cartAPI }

func NewViewCartTool(api cartAPI) Tool { return &viewCartTool{api: api} }

func (t *viewCartTool) Name() string        { return "view_cart" }
func (t *viewCartTool) Description() string { return "Return the current user's shopping cart with items and total in cents." }
func (t *viewCartTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{}}`)
}

func (t *viewCartTool) Call(ctx context.Context, args json.RawMessage, userID string) (Result, error) {
	if userID == "" {
		return Result{}, errors.New("view_cart: authenticated user required")
	}
	cart, err := t.api.GetCart(ctx, apphttp.JWTFromContext(ctx))
	if err != nil {
		return Result{}, fmt.Errorf("view_cart: %w", err)
	}
	content := map[string]any{
		"items": cart.Items,
		"total": cart.Total,
	}
	return Result{Content: content, Display: map[string]any{"kind": "cart", "cart": cart}}, nil
}

// ---------- add_to_cart ----------

type addToCartTool struct{ api cartAPI }

func NewAddToCartTool(api cartAPI) Tool { return &addToCartTool{api: api} }

func (t *addToCartTool) Name() string { return "add_to_cart" }
func (t *addToCartTool) Description() string {
	return "Add a product to the current user's cart. Quantity must be a positive integer."
}
func (t *addToCartTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type":"object",
		"properties":{
			"product_id":{"type":"string"},
			"qty":{"type":"integer","minimum":1}
		},
		"required":["product_id","qty"]
	}`)
}

type addToCartArgs struct {
	ProductID string `json:"product_id"`
	Qty       int    `json:"qty"`
}

func (t *addToCartTool) Call(ctx context.Context, args json.RawMessage, userID string) (Result, error) {
	if userID == "" {
		return Result{}, errors.New("add_to_cart: authenticated user required")
	}
	var a addToCartArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return Result{}, fmt.Errorf("add_to_cart: bad args: %w", err)
	}
	if a.ProductID == "" {
		return Result{}, errors.New("add_to_cart: product_id is required")
	}
	if a.Qty <= 0 {
		return Result{}, errors.New("add_to_cart: qty must be positive")
	}

	item, err := t.api.AddToCart(ctx, apphttp.JWTFromContext(ctx), a.ProductID, a.Qty)
	if err != nil {
		return Result{}, fmt.Errorf("add_to_cart: %w", err)
	}
	content := map[string]any{
		"id":         item.ID,
		"product_id": item.ProductID,
		"name":       item.ProductName,
		"price":      item.ProductPrice,
		"quantity":   item.Quantity,
	}
	return Result{Content: content, Display: map[string]any{"kind": "cart_item", "item": item}}, nil
}
```

- [ ] **Step 4: Run tests**

```bash
cd go/ai-service && go test ./internal/tools/... -v
```

- [ ] **Step 5: Commit**

```bash
git add go/ai-service/internal/tools/cart.go go/ai-service/internal/tools/cart_test.go
git commit -m "feat(ai-service): add view_cart and add_to_cart tools"
```

---

## Task 8: `ecommerce-service` returns — migration, model, repo, service, handler

**Files:**
- Create: `go/ecommerce-service/migrations/004_create_returns.up.sql`
- Create: `go/ecommerce-service/migrations/004_create_returns.down.sql`
- Create: `go/ecommerce-service/internal/model/return.go`
- Create: `go/ecommerce-service/internal/repository/return.go`
- Create: `go/ecommerce-service/internal/service/return.go`
- Create: `go/ecommerce-service/internal/handler/return.go`
- Create: `go/ecommerce-service/internal/handler/return_test.go`
- Modify: `go/ecommerce-service/cmd/server/main.go`

- [ ] **Step 1: Migration**

`go/ecommerce-service/migrations/004_create_returns.up.sql`:

```sql
CREATE TABLE IF NOT EXISTS returns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'requested',
    reason TEXT NOT NULL,
    item_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_returns_user ON returns (user_id);
CREATE INDEX idx_returns_order ON returns (order_id);
```

`go/ecommerce-service/migrations/004_create_returns.down.sql`:

```sql
DROP TABLE IF EXISTS returns;
```

- [ ] **Step 2: Model**

`go/ecommerce-service/internal/model/return.go`:

```go
package model

import (
	"time"

	"github.com/google/uuid"
)

type Return struct {
	ID        uuid.UUID `json:"id"`
	OrderID   uuid.UUID `json:"orderId"`
	UserID    uuid.UUID `json:"userId"`
	Status    string    `json:"status"`
	Reason    string    `json:"reason"`
	ItemIDs   []string  `json:"itemIds"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type InitiateReturnRequest struct {
	ItemIDs []string `json:"itemIds" binding:"required"`
	Reason  string   `json:"reason"  binding:"required"`
}
```

- [ ] **Step 3: Repository**

`go/ecommerce-service/internal/repository/return.go`:

```go
package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/model"
)

type ReturnRepository struct {
	pool *pgxpool.Pool
}

func NewReturnRepository(pool *pgxpool.Pool) *ReturnRepository {
	return &ReturnRepository{pool: pool}
}

func (r *ReturnRepository) Create(ctx context.Context, orderID, userID uuid.UUID, itemIDs []string, reason string) (*model.Return, error) {
	itemsJSON, err := json.Marshal(itemIDs)
	if err != nil {
		return nil, fmt.Errorf("marshal itemIDs: %w", err)
	}

	var ret model.Return
	err = r.pool.QueryRow(ctx,
		`INSERT INTO returns (order_id, user_id, status, reason, item_ids)
		 VALUES ($1, $2, 'requested', $3, $4)
		 RETURNING id, order_id, user_id, status, reason, item_ids, created_at, updated_at`,
		orderID, userID, reason, itemsJSON,
	).Scan(&ret.ID, &ret.OrderID, &ret.UserID, &ret.Status, &ret.Reason, &itemsJSON, &ret.CreatedAt, &ret.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert return: %w", err)
	}
	if err := json.Unmarshal(itemsJSON, &ret.ItemIDs); err != nil {
		return nil, fmt.Errorf("unmarshal item_ids: %w", err)
	}
	return &ret, nil
}
```

- [ ] **Step 4: Service**

`go/ecommerce-service/internal/service/return.go`:

```go
package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/model"
)

var ErrOrderNotOwned = errors.New("order not owned by user")

type ReturnRepositoryInterface interface {
	Create(ctx context.Context, orderID, userID uuid.UUID, itemIDs []string, reason string) (*model.Return, error)
}

type OrderLookup interface {
	GetOrder(ctx context.Context, orderID uuid.UUID) (*model.Order, error)
}

type ReturnService struct {
	returns ReturnRepositoryInterface
	orders  OrderLookup
}

func NewReturnService(returns ReturnRepositoryInterface, orders OrderLookup) *ReturnService {
	return &ReturnService{returns: returns, orders: orders}
}

// Initiate verifies the order belongs to userID before creating the return.
func (s *ReturnService) Initiate(ctx context.Context, userID, orderID uuid.UUID, itemIDs []string, reason string) (*model.Return, error) {
	order, err := s.orders.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.UserID != userID {
		return nil, ErrOrderNotOwned
	}
	return s.returns.Create(ctx, orderID, userID, itemIDs, reason)
}
```

Note: the existing `OrderService` (or an interface that wraps it) must implement `GetOrder(ctx, id)` returning `*model.Order` with `UserID` populated. Check `go/ecommerce-service/internal/service/order.go` during implementation. If it already does, reuse it directly; if not, wrap it or extend it minimally.

- [ ] **Step 5: Handler + handler test (TDD)**

`go/ecommerce-service/internal/handler/return.go`:

```go
package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/model"
	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/service"
)

type ReturnServiceInterface interface {
	Initiate(ctx context.Context, userID, orderID uuid.UUID, itemIDs []string, reason string) (*model.Return, error)
}

type ReturnHandler struct {
	svc ReturnServiceInterface
}

func NewReturnHandler(svc ReturnServiceInterface) *ReturnHandler {
	return &ReturnHandler{svc: svc}
}

func (h *ReturnHandler) Initiate(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}
	var req model.InitiateReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ret, err := h.svc.Initiate(c.Request.Context(), userID, orderID, req.ItemIDs, req.Reason)
	if err != nil {
		if errors.Is(err, service.ErrOrderNotOwned) {
			// 404, not 403, so we don't leak order existence.
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initiate return"})
		return
	}
	c.JSON(http.StatusCreated, ret)
}
```

`go/ecommerce-service/internal/handler/return_test.go`:

```go
package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/model"
	"github.com/kabradshaw1/portfolio/go/ecommerce-service/internal/service"
)

type fakeReturnService struct {
	gotUser uuid.UUID
	gotOrd  uuid.UUID
	gotIDs  []string
	gotRsn  string
	ret     *model.Return
	err     error
}

func (f *fakeReturnService) Initiate(ctx context.Context, userID, orderID uuid.UUID, itemIDs []string, reason string) (*model.Return, error) {
	f.gotUser = userID
	f.gotOrd = orderID
	f.gotIDs = itemIDs
	f.gotRsn = reason
	return f.ret, f.err
}

func TestReturnHandler_Initiate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	user := uuid.New()
	order := uuid.New()
	svc := &fakeReturnService{ret: &model.Return{ID: uuid.New(), OrderID: order, UserID: user, Status: "requested", Reason: "doesn't fit"}}
	h := NewReturnHandler(svc)

	r := gin.New()
	r.POST("/orders/:id/returns", func(c *gin.Context) {
		c.Set("userId", user.String())
		h.Initiate(c)
	})

	body := strings.NewReader(`{"itemIds":["i1"],"reason":"doesn't fit"}`)
	req := httptest.NewRequest(http.MethodPost, "/orders/"+order.String()+"/returns", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if svc.gotUser != user || svc.gotOrd != order {
		t.Errorf("ids = %v %v", svc.gotUser, svc.gotOrd)
	}
	if svc.gotRsn != "doesn't fit" || len(svc.gotIDs) != 1 {
		t.Errorf("svc args = %+v", svc)
	}
}

func TestReturnHandler_Initiate_NotOwned(t *testing.T) {
	gin.SetMode(gin.TestMode)
	user := uuid.New()
	order := uuid.New()
	svc := &fakeReturnService{err: service.ErrOrderNotOwned}
	h := NewReturnHandler(svc)

	r := gin.New()
	r.POST("/orders/:id/returns", func(c *gin.Context) {
		c.Set("userId", user.String())
		h.Initiate(c)
	})

	body := strings.NewReader(`{"itemIds":["i1"],"reason":"r"}`)
	req := httptest.NewRequest(http.MethodPost, "/orders/"+order.String()+"/returns", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestReturnHandler_Initiate_BadBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeReturnService{}
	h := NewReturnHandler(svc)

	r := gin.New()
	user := uuid.New()
	order := uuid.New()
	r.POST("/orders/:id/returns", func(c *gin.Context) {
		c.Set("userId", user.String())
		h.Initiate(c)
	})

	body := strings.NewReader(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/orders/"+order.String()+"/returns", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
```

Run:

```bash
cd go/ecommerce-service && go test ./internal/handler/... -run TestReturnHandler -v
```

Expected: 3/3 PASS.

- [ ] **Step 6: Wire the route in `cmd/server/main.go`**

Extend `go/ecommerce-service/cmd/server/main.go`:

1. Import the new `repository`, `service`, `handler` types.
2. After the existing `orderSvc` wiring:
   ```go
   returnRepo := repository.NewReturnRepository(pool)
   returnSvc := service.NewReturnService(returnRepo, orderSvc)
   returnHandler := handler.NewReturnHandler(returnSvc)
   ```
3. In the authenticated routes group, add:
   ```go
   auth.POST("/orders/:id/returns", returnHandler.Initiate)
   ```

Note: `service.NewReturnService` expects an `OrderLookup` interface. `orderSvc` must have a `GetOrder(ctx, uuid) (*model.Order, error)` method — it does, per the existing `OrderServiceInterface` in the handler package. If Go complains about an interface mismatch, wrap with a small adapter struct.

- [ ] **Step 7: Run the whole ecommerce suite**

```bash
cd go/ecommerce-service && go test ./... -count=1
```

Expected: all PASS.

- [ ] **Step 8: Commit**

```bash
git add go/ecommerce-service/migrations/004_create_returns.up.sql \
        go/ecommerce-service/migrations/004_create_returns.down.sql \
        go/ecommerce-service/internal/model/return.go \
        go/ecommerce-service/internal/repository/return.go \
        go/ecommerce-service/internal/service/return.go \
        go/ecommerce-service/internal/handler/return.go \
        go/ecommerce-service/internal/handler/return_test.go \
        go/ecommerce-service/cmd/server/main.go
git commit -m "feat(ecommerce): add returns endpoint with ownership check"
```

---

## Task 9: `initiate_return` tool in ai-service

**Files:**
- Create: `go/ai-service/internal/tools/returns.go`
- Create: `go/ai-service/internal/tools/returns_test.go`

- [ ] **Step 1: Failing test**

```go
package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/kabradshaw1/portfolio/go/ai-service/internal/tools/clients"
)

type fakeReturnsAPI struct {
	ret      clients.Return
	err      error
	seenJWT  string
	seenOrd  string
	seenIDs  []string
	seenRsn  string
}

func (f *fakeReturnsAPI) InitiateReturn(ctx context.Context, jwt, orderID string, itemIDs []string, reason string) (clients.Return, error) {
	f.seenJWT = jwt
	f.seenOrd = orderID
	f.seenIDs = itemIDs
	f.seenRsn = reason
	return f.ret, f.err
}

func TestInitiateReturnTool_Success(t *testing.T) {
	fake := &fakeReturnsAPI{ret: clients.Return{ID: "r1", OrderID: "o1", Status: "requested", Reason: "doesn't fit"}}
	tool := NewInitiateReturnTool(fake)

	args := json.RawMessage(`{"order_id":"o1","item_ids":["i1"],"reason":"doesn't fit"}`)
	res, err := tool.Call(ctxWithJWT("tok"), args, "user-1")
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if fake.seenJWT != "tok" || fake.seenOrd != "o1" || fake.seenRsn != "doesn't fit" {
		t.Errorf("seen = %+v", fake)
	}
	m := res.Content.(map[string]any)
	if m["status"] != "requested" {
		t.Errorf("content = %+v", m)
	}
}

func TestInitiateReturnTool_MissingFields(t *testing.T) {
	tool := NewInitiateReturnTool(&fakeReturnsAPI{})
	for _, args := range []string{
		`{}`,
		`{"order_id":"o1"}`,
		`{"order_id":"o1","item_ids":[]}`,
		`{"order_id":"o1","item_ids":["i1"]}`, // missing reason
		`{"item_ids":["i1"],"reason":"r"}`,    // missing order_id
	} {
		if _, err := tool.Call(ctxWithJWT("tok"), json.RawMessage(args), "user-1"); err == nil {
			t.Errorf("expected error for args %s", args)
		}
	}
}

func TestInitiateReturnTool_RequiresUserID(t *testing.T) {
	tool := NewInitiateReturnTool(&fakeReturnsAPI{})
	args := json.RawMessage(`{"order_id":"o1","item_ids":["i1"],"reason":"r"}`)
	_, err := tool.Call(ctxWithJWT("tok"), args, "")
	if err == nil {
		t.Fatal("expected error")
	}
}
```

- [ ] **Step 2: Implement `returns.go`**

```go
package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	apphttp "github.com/kabradshaw1/portfolio/go/ai-service/internal/http"
	"github.com/kabradshaw1/portfolio/go/ai-service/internal/tools/clients"
)

type returnsAPI interface {
	InitiateReturn(ctx context.Context, jwt, orderID string, itemIDs []string, reason string) (clients.Return, error)
}

type initiateReturnTool struct{ api returnsAPI }

func NewInitiateReturnTool(api returnsAPI) Tool { return &initiateReturnTool{api: api} }

func (t *initiateReturnTool) Name() string { return "initiate_return" }
func (t *initiateReturnTool) Description() string {
	return "Initiate a return for specific items on one of the current user's orders. Requires the order id, a non-empty list of item ids, and a short reason."
}
func (t *initiateReturnTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type":"object",
		"properties":{
			"order_id":{"type":"string"},
			"item_ids":{"type":"array","items":{"type":"string"},"minItems":1},
			"reason":{"type":"string"}
		},
		"required":["order_id","item_ids","reason"]
	}`)
}

type initiateReturnArgs struct {
	OrderID string   `json:"order_id"`
	ItemIDs []string `json:"item_ids"`
	Reason  string   `json:"reason"`
}

func (t *initiateReturnTool) Call(ctx context.Context, args json.RawMessage, userID string) (Result, error) {
	if userID == "" {
		return Result{}, errors.New("initiate_return: authenticated user required")
	}
	var a initiateReturnArgs
	if err := json.Unmarshal(args, &a); err != nil {
		return Result{}, fmt.Errorf("initiate_return: bad args: %w", err)
	}
	if a.OrderID == "" {
		return Result{}, errors.New("initiate_return: order_id is required")
	}
	if len(a.ItemIDs) == 0 {
		return Result{}, errors.New("initiate_return: item_ids must be non-empty")
	}
	if a.Reason == "" {
		return Result{}, errors.New("initiate_return: reason is required")
	}

	ret, err := t.api.InitiateReturn(ctx, apphttp.JWTFromContext(ctx), a.OrderID, a.ItemIDs, a.Reason)
	if err != nil {
		return Result{}, fmt.Errorf("initiate_return: %w", err)
	}
	content := map[string]any{
		"id":       ret.ID,
		"order_id": ret.OrderID,
		"status":   ret.Status,
		"reason":   ret.Reason,
	}
	return Result{Content: content, Display: map[string]any{"kind": "return_confirmation", "return": ret}}, nil
}
```

- [ ] **Step 3: Run tests**

```bash
cd go/ai-service && go test ./internal/tools/... -v
```

Expected: all tool tests PASS.

- [ ] **Step 4: Commit**

```bash
git add go/ai-service/internal/tools/returns.go go/ai-service/internal/tools/returns_test.go
git commit -m "feat(ai-service): add initiate_return tool"
```

---

## Task 10: Wire the six new tools + JWT secret in `main.go`

**Files:**
- Modify: `go/ai-service/cmd/server/main.go`

- [ ] **Step 1: Update main.go**

Add:
- New env var `JWT_SECRET` (log-fail if empty: `log.Fatal("JWT_SECRET is required")`).
- Register the six new tools on the same registry.
- Pass `jwtSecret` to `apphttp.RegisterChatRoutes`.

Full diff (conceptually):

```go
jwtSecret := os.Getenv("JWT_SECRET")
if jwtSecret == "" {
    log.Fatal("JWT_SECRET is required")
}

// ... existing llm + ecomClient setup ...

registry := tools.NewMemRegistry()
registry.Register(tools.NewSearchProductsTool(ecomClient))
registry.Register(tools.NewGetProductTool(ecomClient))
registry.Register(tools.NewCheckInventoryTool(ecomClient))
registry.Register(tools.NewListOrdersTool(ecomClient))
registry.Register(tools.NewGetOrderTool(ecomClient))
registry.Register(tools.NewSummarizeOrdersTool(ecomClient, llmc))
registry.Register(tools.NewViewCartTool(ecomClient))
registry.Register(tools.NewAddToCartTool(ecomClient))
registry.Register(tools.NewInitiateReturnTool(ecomClient))

// ...

apphttp.RegisterChatRoutes(router, a, jwtSecret)
```

- [ ] **Step 2: Build and test**

```bash
cd go/ai-service && go build ./...
cd go/ai-service && JWT_SECRET=test go test ./... -count=1
```

(Setting `JWT_SECRET` for tests is harmless — tests don't call `main()`; this is just protection against any test that might in the future.)

Expected: clean build, all packages PASS.

- [ ] **Step 3: Update `go/docker-compose.yml` to pass `JWT_SECRET`**

In the `ai-service` block in `go/docker-compose.yml`, add to `environment`:

```yaml
      JWT_SECRET: dev-secret-key-at-least-32-characters-long
```

This matches what `auth-service` and `ecommerce-service` already use in the same file, so tokens minted by one are validated by all three. Do NOT factor this into a `.env` file — the existing Go services hardcode it inline, match that pattern.

- [ ] **Step 4: Commit**

```bash
git add go/ai-service/cmd/server/main.go go/docker-compose.yml
git commit -m "feat(ai-service): wire JWT secret and register user-scoped tools"
```

---

## Done criteria for Plan 2

- `go test ./go/ai-service/...` and `go test ./go/ecommerce-service/...` both pass fully offline.
- `GET /orders/:id` returns 404 for non-owning users (regression-tested).
- `POST /orders/:id/returns` creates a return scoped to the authenticated user and rejects non-owners with 404.
- `POST /chat` validates JWT when present, threads `userID` into `Turn`, stashes the raw token in the request context, and returns 401 on invalid tokens.
- `ai-service` registry has nine tools: three catalog (Plan 1) + six user-scoped (Plan 2).
- No Redis, no evals, no metrics, no frontend — still Plans 3–5.

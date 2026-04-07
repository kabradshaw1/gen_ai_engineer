# Go Google OAuth Sign-In Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Google sign-in to the Go auth-service, mirroring Java's code-exchange pattern, plus dedicated `/go/login` and `/go/register` pages on the frontend.

**Architecture:** Frontend drives the Google OAuth redirect (response_type=code) with `/go/login` as the redirect URI. `/go/login` extracts `?code=` and POSTs `{code, redirectUri}` to Go's new `POST /auth/google`. Backend exchanges the code for tokens at Google, fetches userinfo, upserts the user with nullable password_hash, and returns the existing JWT access/refresh pair plus `avatarUrl`.

**Tech Stack:** Go 1.26 (gin, pgx), PostgreSQL, Next.js 15, TypeScript, Google OAuth 2.0 (authorization code flow, server-side exchange).

**Spec:** `docs/superpowers/specs/2026-04-07-go-google-oauth-design.md`

---

## Prerequisites (manual, one-time)

- [ ] **P1: Add redirect URIs to the existing Google OAuth client**

  Go to Google Cloud Console → APIs & Services → Credentials → the existing OAuth 2.0 Client used by the Java stack. Under "Authorized redirect URIs", add:
  - `http://localhost:3000/go/login`
  - `https://kylebradshaw.dev/go/login`

  Save.

- [ ] **P2: Copy Google creds into `go/.env`**

  Copy the existing `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` values from `java/.env` into `go/.env` (create the file if it doesn't exist — it's gitignored):

  ```
  GOOGLE_CLIENT_ID=<same value as java/.env>
  GOOGLE_CLIENT_SECRET=<same value as java/.env>
  ```

---

## File Structure

### Created
- `go/auth-service/migrations/002_google_oauth.sql` — schema migration
- `go/auth-service/internal/google/client.go` — Google OAuth client (token exchange + userinfo)
- `go/auth-service/internal/google/client_test.go` — httptest-driven tests
- `frontend/src/components/go/GoGoogleLoginButton.tsx` — Google button for Go stack
- `frontend/src/app/go/login/page.tsx` — login page with email form + Google button + OAuth callback handler
- `frontend/src/app/go/register/page.tsx` — register page with email form + Google button

### Modified
- `go/auth-service/internal/model/user.go` — `User.PasswordHash` becomes `*string`, `User.AvatarURL *string` added, new `GoogleLoginRequest`, `AuthResponse.AvatarURL` added
- `go/auth-service/internal/repository/user.go` — scans updated for nullable password_hash + avatar_url; new `UpsertGoogleUser` method
- `go/auth-service/internal/service/auth.go` — `UserRepo` interface gains `UpsertGoogleUser`; `Login` handles nil password_hash; new `AuthenticateGoogleUser` method; `generateTokens` includes avatar in response
- `go/auth-service/internal/handler/auth.go` — new `GoogleClientInterface`, `AuthHandler` gains `googleClient` field, new `GoogleLogin` handler
- `go/auth-service/cmd/server/main.go` — read `GOOGLE_*` env vars, construct `google.Client`, pass to handler, register `POST /auth/google` route
- `go/auth-service/internal/service/auth_test.go` — new tests for `AuthenticateGoogleUser`, fake repo updated with new method
- `go/auth-service/internal/handler/auth_test.go` — new tests for `GoogleLogin` with fake google client
- `go/docker-compose.yml` — pass `GOOGLE_CLIENT_ID`/`GOOGLE_CLIENT_SECRET` to `auth-service`
- `go/.env.example` — add placeholders for the two Google vars
- `go/k8s/deployments/auth-service.yml` — wire `GOOGLE_CLIENT_ID`/`GOOGLE_CLIENT_SECRET` from `go-secrets`
- `go/k8s/secrets/go-secrets.yml` — add `google-client-id` and `google-client-secret` keys (gitignored; instructions only)
- `frontend/src/components/go/GoAuthProvider.tsx` — `avatarUrl` on `GoAuthUser`, new `loginWithGoogle` method
- `frontend/src/lib/go-auth.ts` — re-export `GOOGLE_CLIENT_ID` from `./auth`

---

## Task 1: Database migration

**Files:**
- Create: `go/auth-service/migrations/002_google_oauth.sql`

- [ ] **Step 1: Create the migration file**

```sql
-- 002_google_oauth.sql
-- Allow users to exist without a local password (Google-only sign-in).
-- Add avatar_url populated from Google userinfo.

ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;
ALTER TABLE users ADD COLUMN avatar_url VARCHAR(500);
```

- [ ] **Step 2: Apply the migration to local dev Postgres**

Run from `go/`:

```bash
docker compose exec -T postgres psql -U taskuser -d ecommercedb < auth-service/migrations/002_google_oauth.sql
```

Expected output:
```
ALTER TABLE
ALTER TABLE
```

- [ ] **Step 3: Verify schema**

```bash
docker compose exec -T postgres psql -U taskuser -d ecommercedb -c "\d users"
```

Expected: `password_hash` column shows without `NOT NULL`, `avatar_url varchar(500)` column present.

- [ ] **Step 4: Commit**

```bash
git add go/auth-service/migrations/002_google_oauth.sql
git commit -m "feat(go-auth): add migration for nullable password_hash and avatar_url"
```

---

## Task 2: Update User model for nullable password + avatar

**Files:**
- Modify: `go/auth-service/internal/model/user.go`

- [ ] **Step 1: Update `User` struct and add new request type**

Replace `go/auth-service/internal/model/user.go` with:

```go
package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash *string   `json:"-"`
	Name         string    `json:"name"`
	AvatarURL    *string   `json:"avatarUrl,omitempty"`
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

type GoogleLoginRequest struct {
	Code        string `json:"code" binding:"required"`
	RedirectURI string `json:"redirectUri" binding:"required"`
}

type AuthResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	UserID       string `json:"userId"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	AvatarURL    string `json:"avatarUrl,omitempty"`
}
```

- [ ] **Step 2: Verify it compiles (will fail — repo/service still use old types)**

```bash
cd go/auth-service && go build ./...
```

Expected: compilation errors in `repository/user.go` and `service/auth.go` referencing `PasswordHash` as string. That's fine — Tasks 3 and 4 fix them.

- [ ] **Step 3: Do not commit yet**

Leave this change uncommitted. Task 3 will commit the model + repository changes together since they're one logical unit.

---

## Task 3: Update repository for nullable password + add `UpsertGoogleUser`

**Files:**
- Modify: `go/auth-service/internal/repository/user.go`

- [ ] **Step 1: Replace the repository file**

Replace `go/auth-service/internal/repository/user.go` with:

```go
package repository

import (
	"context"
	"errors"
	"strings"

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
		 RETURNING id, email, password_hash, name, avatar_url, created_at`,
		email, passwordHash, name,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.AvatarURL, &user.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, ErrEmailExists
		}
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, name, avatar_url, created_at FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.AvatarURL, &user.CreatedAt)
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
		`SELECT id, email, password_hash, name, avatar_url, created_at FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.AvatarURL, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// UpsertGoogleUser creates a new Google-authenticated user, or updates an
// existing user's name and avatar. password_hash is never modified.
func (r *UserRepository) UpsertGoogleUser(ctx context.Context, email, name, avatarURL string) (*model.User, error) {
	user := &model.User{}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (email, name, avatar_url, password_hash)
		 VALUES ($1, $2, $3, NULL)
		 ON CONFLICT (email) DO UPDATE
		   SET name = EXCLUDED.name,
		       avatar_url = EXCLUDED.avatar_url
		 RETURNING id, email, password_hash, name, avatar_url, created_at`,
		email, name, avatarURL,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.AvatarURL, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}
```

- [ ] **Step 2: Verify repo compiles**

```bash
cd go/auth-service && go build ./internal/repository/...
```

Expected: PASS.

- [ ] **Step 3: Commit model + repo together**

```bash
git add go/auth-service/internal/model/user.go go/auth-service/internal/repository/user.go
git commit -m "feat(go-auth): nullable password_hash and UpsertGoogleUser"
```

---

## Task 4: Update service for nullable password + handle Login breakage

**Files:**
- Modify: `go/auth-service/internal/service/auth.go`

- [ ] **Step 1: Update `UserRepo` interface, `Login`, and `generateTokens`**

Edit `go/auth-service/internal/service/auth.go`. Three changes:

a) Add `UpsertGoogleUser` to the interface and the `AuthenticateGoogleUser` method stub (the full implementation lands in Task 6 after the Google client exists). For now, add the method so later tasks can reference it without reshuffling.

Replace the `UserRepo` interface:

```go
// UserRepo abstracts user persistence so the service is testable.
type UserRepo interface {
	Create(ctx context.Context, email, passwordHash, name string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
	UpsertGoogleUser(ctx context.Context, email, name, avatarURL string) (*model.User, error)
}
```

b) Replace `Login` to handle nil `PasswordHash` (Google-only users cannot log in with a password):

```go
// Login verifies credentials and returns JWT tokens.
func (s *AuthService) Login(ctx context.Context, email, password string) (*model.AuthResponse, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if user.PasswordHash == nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return s.generateTokens(user)
}
```

c) Replace `generateTokens` to include avatar in the response:

```go
// generateTokens creates an access token and a refresh token for the user.
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
		"sub":   user.ID.String(),
		"email": user.Email,
		"iat":   now.Unix(),
		"exp":   now.Add(s.refreshTokenTTL).Unix(),
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	avatar := ""
	if user.AvatarURL != nil {
		avatar = *user.AvatarURL
	}

	return &model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserID:       user.ID.String(),
		Email:        user.Email,
		Name:         user.Name,
		AvatarURL:    avatar,
	}, nil
}
```

- [ ] **Step 2: Add the `AuthenticateGoogleUser` service method**

Append to `go/auth-service/internal/service/auth.go` (after `Refresh`, before `generateTokens`):

```go
// AuthenticateGoogleUser upserts a Google-authenticated user and issues tokens.
func (s *AuthService) AuthenticateGoogleUser(ctx context.Context, email, name, avatarURL string) (*model.AuthResponse, error) {
	user, err := s.repo.UpsertGoogleUser(ctx, email, name, avatarURL)
	if err != nil {
		return nil, err
	}
	return s.generateTokens(user)
}
```

- [ ] **Step 3: Fix existing service tests to satisfy the new interface**

Tests use a fake repo. Open `go/auth-service/internal/service/auth_test.go` and find the fake repo struct (likely named `fakeUserRepo` or similar). Add a method:

```go
func (f *fakeUserRepo) UpsertGoogleUser(ctx context.Context, email, name, avatarURL string) (*model.User, error) {
	// test stub — individual tests override via a field if needed
	return nil, errors.New("not implemented")
}
```

If tests also construct `model.User` literals with `PasswordHash: "..."`, update those literals to use a pointer:

```go
hash := "..."
u := &model.User{PasswordHash: &hash, ...}
```

- [ ] **Step 4: Run existing service tests**

```bash
cd go/auth-service && go test ./internal/service/... -v
```

Expected: existing tests pass (Register/Login/Refresh). Any compilation errors mean a `PasswordHash` literal was missed — fix and rerun.

- [ ] **Step 5: Commit**

```bash
git add go/auth-service/internal/service/auth.go go/auth-service/internal/service/auth_test.go
git commit -m "feat(go-auth): service layer supports nullable password_hash and Google users"
```

---

## Task 5: Google OAuth client package (TDD)

**Files:**
- Create: `go/auth-service/internal/google/client.go`
- Test: `go/auth-service/internal/google/client_test.go`

- [ ] **Step 1: Write the failing test**

Create `go/auth-service/internal/google/client_test.go`:

```go
package google

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExchangeCode_Success(t *testing.T) {
	var tokenHits, userinfoHits int

	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		tokenHits++
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if got := r.Form.Get("code"); got != "auth-code-123" {
			t.Errorf("code = %q, want auth-code-123", got)
		}
		if got := r.Form.Get("client_id"); got != "test-client" {
			t.Errorf("client_id = %q, want test-client", got)
		}
		if got := r.Form.Get("client_secret"); got != "test-secret" {
			t.Errorf("client_secret = %q, want test-secret", got)
		}
		if got := r.Form.Get("grant_type"); got != "authorization_code" {
			t.Errorf("grant_type = %q, want authorization_code", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "g-access-token"})
	})
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		userinfoHits++
		if got := r.Header.Get("Authorization"); got != "Bearer g-access-token" {
			t.Errorf("authorization = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"email":   "alice@example.com",
			"name":    "Alice",
			"picture": "https://example.com/alice.png",
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := NewClient("test-client", "test-secret", srv.URL+"/token", srv.URL+"/userinfo")
	info, err := c.ExchangeCode(context.Background(), "auth-code-123", "http://localhost:3000/go/login")
	if err != nil {
		t.Fatalf("ExchangeCode: %v", err)
	}
	if info.Email != "alice@example.com" || info.Name != "Alice" || info.Picture != "https://example.com/alice.png" {
		t.Errorf("unexpected UserInfo: %+v", info)
	}
	if tokenHits != 1 || userinfoHits != 1 {
		t.Errorf("hits: token=%d userinfo=%d", tokenHits, userinfoHits)
	}
}

func TestExchangeCode_TokenEndpointError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer srv.Close()

	c := NewClient("id", "secret", srv.URL, srv.URL)
	_, err := c.ExchangeCode(context.Background(), "bad", "http://localhost:3000/go/login")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "token") {
		t.Errorf("error should mention token: %v", err)
	}
}

func TestExchangeCode_UserinfoEndpointError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "ok"})
	})
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := NewClient("id", "secret", srv.URL+"/token", srv.URL+"/userinfo")
	_, err := c.ExchangeCode(context.Background(), "code", "http://localhost:3000/go/login")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "userinfo") {
		t.Errorf("error should mention userinfo: %v", err)
	}
}

func TestExchangeCode_MalformedTokenJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	}))
	defer srv.Close()

	c := NewClient("id", "secret", srv.URL, srv.URL)
	_, err := c.ExchangeCode(context.Background(), "code", "http://localhost:3000/go/login")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

```bash
cd go/auth-service && go test ./internal/google/... -v
```

Expected: compilation failure — package `google` does not exist.

- [ ] **Step 3: Implement the Google client**

Create `go/auth-service/internal/google/client.go`:

```go
package google

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// UserInfo is the subset of Google's /userinfo response we care about.
type UserInfo struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// Client exchanges an OAuth authorization code for a Google UserInfo.
type Client struct {
	clientID     string
	clientSecret string
	tokenURL     string
	userinfoURL  string
	http         *http.Client
}

// NewClient constructs a Google OAuth client.
func NewClient(clientID, clientSecret, tokenURL, userinfoURL string) *Client {
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		tokenURL:     tokenURL,
		userinfoURL:  userinfoURL,
		http:         &http.Client{Timeout: 10 * time.Second},
	}
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
}

// ExchangeCode exchanges an authorization code for user profile information.
func (c *Client) ExchangeCode(ctx context.Context, code, redirectURI string) (*UserInfo, error) {
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)
	form.Set("redirect_uri", redirectURI)
	form.Set("grant_type", "authorization_code")

	tokenReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build token request: %w", err)
	}
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	tokenResp, err := c.http.Do(tokenReq)
	if err != nil {
		return nil, fmt.Errorf("google token request: %w", err)
	}
	defer tokenResp.Body.Close()

	if tokenResp.StatusCode >= 400 {
		body, _ := io.ReadAll(tokenResp.Body)
		return nil, fmt.Errorf("google token endpoint returned %d: %s", tokenResp.StatusCode, string(body))
	}

	var tr tokenResponse
	if err := json.NewDecoder(tokenResp.Body).Decode(&tr); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}
	if tr.AccessToken == "" {
		return nil, fmt.Errorf("google token response missing access_token")
	}

	userReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.userinfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build userinfo request: %w", err)
	}
	userReq.Header.Set("Authorization", "Bearer "+tr.AccessToken)

	userResp, err := c.http.Do(userReq)
	if err != nil {
		return nil, fmt.Errorf("google userinfo request: %w", err)
	}
	defer userResp.Body.Close()

	if userResp.StatusCode >= 400 {
		body, _ := io.ReadAll(userResp.Body)
		return nil, fmt.Errorf("google userinfo endpoint returned %d: %s", userResp.StatusCode, string(body))
	}

	var info UserInfo
	if err := json.NewDecoder(userResp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decode userinfo response: %w", err)
	}
	return &info, nil
}
```

- [ ] **Step 4: Run the tests to verify they pass**

```bash
cd go/auth-service && go test ./internal/google/... -v
```

Expected: all four tests PASS.

- [ ] **Step 5: Commit**

```bash
git add go/auth-service/internal/google/
git commit -m "feat(go-auth): add Google OAuth client for code exchange and userinfo"
```

---

## Task 6: Service test for `AuthenticateGoogleUser` (TDD)

**Files:**
- Modify: `go/auth-service/internal/service/auth_test.go`

- [ ] **Step 1: Write the failing test**

Append to `go/auth-service/internal/service/auth_test.go` (the exact fake-repo field names here assume the standard pattern — adjust field names to match the existing fake if different):

```go
func TestAuthenticateGoogleUser_NewUser(t *testing.T) {
	uid := uuid.New()
	repo := &fakeUserRepo{
		upsertFn: func(ctx context.Context, email, name, avatarURL string) (*model.User, error) {
			if email != "new@example.com" || name != "New User" || avatarURL != "https://example.com/pic.png" {
				t.Errorf("upsert args: email=%q name=%q avatar=%q", email, name, avatarURL)
			}
			return &model.User{
				ID:           uid,
				Email:        email,
				Name:         name,
				AvatarURL:    &avatarURL,
				PasswordHash: nil,
				CreatedAt:    time.Now(),
			}, nil
		},
	}
	svc := service.NewAuthService(repo, "test-secret", 900000, 604800000)

	resp, err := svc.AuthenticateGoogleUser(context.Background(), "new@example.com", "New User", "https://example.com/pic.png")
	if err != nil {
		t.Fatalf("AuthenticateGoogleUser: %v", err)
	}
	if resp.UserID != uid.String() {
		t.Errorf("UserID = %q, want %q", resp.UserID, uid.String())
	}
	if resp.Email != "new@example.com" {
		t.Errorf("Email = %q", resp.Email)
	}
	if resp.AvatarURL != "https://example.com/pic.png" {
		t.Errorf("AvatarURL = %q", resp.AvatarURL)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Error("expected non-empty tokens")
	}
}

func TestAuthenticateGoogleUser_RepoError(t *testing.T) {
	repo := &fakeUserRepo{
		upsertFn: func(ctx context.Context, email, name, avatarURL string) (*model.User, error) {
			return nil, errors.New("db down")
		},
	}
	svc := service.NewAuthService(repo, "test-secret", 900000, 604800000)

	_, err := svc.AuthenticateGoogleUser(context.Background(), "x@example.com", "X", "")
	if err == nil {
		t.Fatal("expected error")
	}
}
```

Also update the `fakeUserRepo` struct and its `UpsertGoogleUser` method to use a function field:

```go
type fakeUserRepo struct {
	// ... existing fields
	upsertFn func(ctx context.Context, email, name, avatarURL string) (*model.User, error)
}

func (f *fakeUserRepo) UpsertGoogleUser(ctx context.Context, email, name, avatarURL string) (*model.User, error) {
	if f.upsertFn != nil {
		return f.upsertFn(ctx, email, name, avatarURL)
	}
	return nil, errors.New("not implemented")
}
```

(If the existing fake repo doesn't use function fields but instead stores result maps, adapt to match — the test intent is: "the fake returns this user / this error, assert the service calls upsert and returns tokens".)

Add any missing imports: `"time"`, `"github.com/google/uuid"`.

- [ ] **Step 2: Run the tests — expect both new tests to pass**

`AuthenticateGoogleUser` was already added to the service in Task 4, so these should pass on first run.

```bash
cd go/auth-service && go test ./internal/service/... -v -run AuthenticateGoogleUser
```

Expected: both tests PASS.

- [ ] **Step 3: Run full service test suite**

```bash
cd go/auth-service && go test ./internal/service/... -v
```

Expected: all tests PASS.

- [ ] **Step 4: Commit**

```bash
git add go/auth-service/internal/service/auth_test.go
git commit -m "test(go-auth): cover AuthenticateGoogleUser"
```

---

## Task 7: Handler for `POST /auth/google` (TDD)

**Files:**
- Modify: `go/auth-service/internal/handler/auth.go`
- Modify: `go/auth-service/internal/handler/auth_test.go`

- [ ] **Step 1: Write the failing tests**

Append to `go/auth-service/internal/handler/auth_test.go`:

```go
type fakeGoogleClient struct {
	info *google.UserInfo
	err  error
}

func (f *fakeGoogleClient) ExchangeCode(ctx context.Context, code, redirectURI string) (*google.UserInfo, error) {
	return f.info, f.err
}

func TestGoogleLogin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &fakeAuthService{
		googleFn: func(ctx context.Context, email, name, avatarURL string) (*model.AuthResponse, error) {
			if email != "a@example.com" || name != "A" || avatarURL != "http://pic" {
				t.Errorf("svc args: %q %q %q", email, name, avatarURL)
			}
			return &model.AuthResponse{
				AccessToken:  "access",
				RefreshToken: "refresh",
				UserID:       "uid-1",
				Email:        email,
				Name:         name,
				AvatarURL:    avatarURL,
			}, nil
		},
	}
	gc := &fakeGoogleClient{info: &google.UserInfo{Email: "a@example.com", Name: "A", Picture: "http://pic"}}
	h := handler.NewAuthHandler(svc, gc)

	router := gin.New()
	router.POST("/auth/google", h.GoogleLogin)

	body := strings.NewReader(`{"code":"abc","redirectUri":"http://localhost:3000/go/login"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/google", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var resp model.AuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.AccessToken != "access" || resp.AvatarURL != "http://pic" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestGoogleLogin_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handler.NewAuthHandler(&fakeAuthService{}, &fakeGoogleClient{})
	router := gin.New()
	router.POST("/auth/google", h.GoogleLogin)

	req := httptest.NewRequest(http.MethodPost, "/auth/google", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestGoogleLogin_GoogleError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gc := &fakeGoogleClient{err: errors.New("bad code")}
	h := handler.NewAuthHandler(&fakeAuthService{}, gc)
	router := gin.New()
	router.POST("/auth/google", h.GoogleLogin)

	body := strings.NewReader(`{"code":"abc","redirectUri":"http://localhost:3000/go/login"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/google", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestGoogleLogin_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeAuthService{
		googleFn: func(ctx context.Context, email, name, avatarURL string) (*model.AuthResponse, error) {
			return nil, errors.New("db down")
		},
	}
	gc := &fakeGoogleClient{info: &google.UserInfo{Email: "a@example.com"}}
	h := handler.NewAuthHandler(svc, gc)
	router := gin.New()
	router.POST("/auth/google", h.GoogleLogin)

	body := strings.NewReader(`{"code":"abc","redirectUri":"http://localhost:3000/go/login"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/google", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", w.Code)
	}
}
```

Also extend the existing `fakeAuthService` with `googleFn` and an `AuthenticateGoogleUser` method:

```go
type fakeAuthService struct {
	// ... existing fields
	googleFn func(ctx context.Context, email, name, avatarURL string) (*model.AuthResponse, error)
}

func (f *fakeAuthService) AuthenticateGoogleUser(ctx context.Context, email, name, avatarURL string) (*model.AuthResponse, error) {
	if f.googleFn != nil {
		return f.googleFn(ctx, email, name, avatarURL)
	}
	return nil, errors.New("not implemented")
}
```

Add imports: `"github.com/kabradshaw1/portfolio/go/auth-service/internal/google"`, `"encoding/json"`, `"strings"`.

- [ ] **Step 2: Run the tests to verify they fail**

```bash
cd go/auth-service && go test ./internal/handler/... -v -run GoogleLogin
```

Expected: compilation error — `NewAuthHandler` expects one arg, not two; `GoogleLogin` method does not exist; `AuthenticateGoogleUser` not on interface.

- [ ] **Step 3: Update the handler file**

Replace `go/auth-service/internal/handler/auth.go` with:

```go
package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kabradshaw1/portfolio/go/auth-service/internal/google"
	"github.com/kabradshaw1/portfolio/go/auth-service/internal/model"
)

type AuthServiceInterface interface {
	Register(ctx context.Context, email, password, name string) (*model.AuthResponse, error)
	Login(ctx context.Context, email, password string) (*model.AuthResponse, error)
	Refresh(ctx context.Context, refreshToken string) (*model.AuthResponse, error)
	AuthenticateGoogleUser(ctx context.Context, email, name, avatarURL string) (*model.AuthResponse, error)
}

type GoogleClientInterface interface {
	ExchangeCode(ctx context.Context, code, redirectURI string) (*google.UserInfo, error)
}

type AuthHandler struct {
	svc          AuthServiceInterface
	googleClient GoogleClientInterface
}

func NewAuthHandler(svc AuthServiceInterface, googleClient GoogleClientInterface) *AuthHandler {
	return &AuthHandler{svc: svc, googleClient: googleClient}
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

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	var req model.GoogleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	info, err := h.googleClient.ExchangeCode(c.Request.Context(), req.Code, req.RedirectURI)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "google authentication failed"})
		return
	}
	resp, err := h.svc.AuthenticateGoogleUser(c.Request.Context(), info.Email, info.Name, info.Picture)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to authenticate user"})
		return
	}
	c.JSON(http.StatusOK, resp)
}
```

- [ ] **Step 4: Run the tests to verify they pass**

```bash
cd go/auth-service && go test ./internal/handler/... -v
```

Expected: all handler tests PASS (existing + four new GoogleLogin tests).

- [ ] **Step 5: Commit**

```bash
git add go/auth-service/internal/handler/
git commit -m "feat(go-auth): add POST /auth/google handler"
```

---

## Task 8: Wire everything in `main.go`

**Files:**
- Modify: `go/auth-service/cmd/server/main.go`

- [ ] **Step 1: Read Google env vars and construct the client**

Replace `go/auth-service/cmd/server/main.go` with:

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

	"github.com/kabradshaw1/portfolio/go/auth-service/internal/google"
	"github.com/kabradshaw1/portfolio/go/auth-service/internal/handler"
	"github.com/kabradshaw1/portfolio/go/auth-service/internal/middleware"
	"github.com/kabradshaw1/portfolio/go/auth-service/internal/repository"
	"github.com/kabradshaw1/portfolio/go/auth-service/internal/service"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	if googleClientID == "" {
		log.Fatal("GOOGLE_CLIENT_ID is required")
	}
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if googleClientSecret == "" {
		log.Fatal("GOOGLE_CLIENT_SECRET is required")
	}
	googleTokenURL := os.Getenv("GOOGLE_TOKEN_URL")
	if googleTokenURL == "" {
		googleTokenURL = "https://oauth2.googleapis.com/token"
	}
	googleUserinfoURL := os.Getenv("GOOGLE_USERINFO_URL")
	if googleUserinfoURL == "" {
		googleUserinfoURL = "https://www.googleapis.com/oauth2/v3/userinfo"
	}
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "http://localhost:3000"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8091"
	}

	// Connect to Postgres
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	slog.Info("connected to database")

	// Wire dependencies
	userRepo := repository.NewUserRepository(pool)
	authSvc := service.NewAuthService(userRepo, jwtSecret, 900000, 604800000)
	googleClient := google.NewClient(googleClientID, googleClientSecret, googleTokenURL, googleUserinfoURL)
	authHandler := handler.NewAuthHandler(authSvc, googleClient)
	healthHandler := handler.NewHealthHandler(pool)

	// Set up Gin
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logging())
	router.Use(middleware.Metrics())
	router.Use(middleware.CORS(allowedOrigins))

	// Routes
	router.POST("/auth/register", authHandler.Register)
	router.POST("/auth/login", authHandler.Login)
	router.POST("/auth/refresh", authHandler.Refresh)
	router.POST("/auth/google", authHandler.GoogleLogin)
	router.GET("/health", healthHandler.Health)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Start server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		slog.Info("server starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	slog.Info("server stopped")
}
```

- [ ] **Step 2: Build to confirm**

```bash
cd go/auth-service && go build ./...
```

Expected: no errors.

- [ ] **Step 3: Run full test suite**

```bash
cd go/auth-service && go test ./... -race
```

Expected: all PASS.

- [ ] **Step 4: Commit**

```bash
git add go/auth-service/cmd/server/main.go
git commit -m "feat(go-auth): wire Google client and /auth/google route"
```

---

## Task 9: Docker Compose + .env.example

**Files:**
- Modify: `go/docker-compose.yml`
- Modify: `go/.env.example` (create if missing)

- [ ] **Step 1: Update `go/docker-compose.yml`**

In the `auth-service` `environment:` block, add two lines (place them below `ALLOWED_ORIGINS`):

```yaml
      GOOGLE_CLIENT_ID: ${GOOGLE_CLIENT_ID}
      GOOGLE_CLIENT_SECRET: ${GOOGLE_CLIENT_SECRET}
```

- [ ] **Step 2: Create or update `go/.env.example`**

If `go/.env.example` exists, append; otherwise create with:

```
GOOGLE_CLIENT_ID=your-google-oauth-client-id
GOOGLE_CLIENT_SECRET=your-google-oauth-client-secret
```

- [ ] **Step 3: Restart compose and verify**

Ensure `go/.env` has the real values (from prerequisite P2). Then:

```bash
cd go && docker compose up -d --build auth-service
docker compose logs auth-service | tail -20
```

Expected: "server starting port=8091", no fatal errors about missing `GOOGLE_CLIENT_ID`.

- [ ] **Step 4: Smoke test the new endpoint with a bogus code**

```bash
curl -s -X POST http://localhost:8091/auth/google \
  -H "Content-Type: application/json" \
  -d '{"code":"bogus","redirectUri":"http://localhost:3000/go/login"}'
```

Expected: `{"error":"google authentication failed"}` (status 401). Confirms wiring; real code exchange is tested end-to-end in Task 14.

- [ ] **Step 5: Commit**

```bash
git add go/docker-compose.yml go/.env.example
git commit -m "chore(go): wire Google OAuth env vars into compose"
```

---

## Task 10: Kubernetes secret + deployment

**Files:**
- Modify: `go/k8s/deployments/auth-service.yml`
- Modify: `go/k8s/secrets/go-secrets.yml` (gitignored; instructions only)

- [ ] **Step 1: Add env refs in the deployment manifest**

In `go/k8s/deployments/auth-service.yml`, extend the existing `env:` block under the container with:

```yaml
          env:
            - name: JWT_SECRET
              valueFrom:
                secretKeyRef:
                  name: go-secrets
                  key: jwt-secret
            - name: GOOGLE_CLIENT_ID
              valueFrom:
                secretKeyRef:
                  name: go-secrets
                  key: google-client-id
            - name: GOOGLE_CLIENT_SECRET
              valueFrom:
                secretKeyRef:
                  name: go-secrets
                  key: google-client-secret
```

- [ ] **Step 2: Update the gitignored secrets file on your local machine**

Open (or create) `go/k8s/secrets/go-secrets.yml`. Under `stringData:`, add two keys with the same values used in `go/.env`:

```yaml
stringData:
  jwt-secret: <existing>
  google-client-id: <value from go/.env>
  google-client-secret: <value from go/.env>
```

This file is gitignored — do not attempt to commit it.

- [ ] **Step 3: Apply secret and deployment (when ready to deploy)**

```bash
kubectl apply -f go/k8s/secrets/go-secrets.yml
kubectl apply -f go/k8s/deployments/auth-service.yml
kubectl -n go-ecommerce rollout status deploy/go-auth-service
```

Expected: rollout completes. Defer actually running this until after the backend is merged and Kyle pushes.

- [ ] **Step 4: Commit manifest change only**

```bash
git add go/k8s/deployments/auth-service.yml
git commit -m "feat(go-auth): mount Google OAuth secrets in k8s deployment"
```

---

## Task 11: Frontend — extend `GoAuthProvider` with Google login

**Files:**
- Modify: `frontend/src/components/go/GoAuthProvider.tsx`

- [ ] **Step 1: Add `avatarUrl` to user type and new `loginWithGoogle` method**

Replace `frontend/src/components/go/GoAuthProvider.tsx` with:

```tsx
"use client";

import {
  createContext,
  useCallback,
  useContext,
  useState,
} from "react";
import {
  clearGoTokens,
  isGoLoggedIn as checkIsLoggedIn,
  setGoTokens,
  GO_AUTH_URL,
} from "@/lib/go-auth";

interface GoAuthUser {
  userId: string;
  email: string;
  name: string;
  avatarUrl?: string;
}

interface GoAuthContextType {
  user: GoAuthUser | null;
  isLoggedIn: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, name: string) => Promise<void>;
  loginWithGoogle: (code: string, redirectUri: string) => Promise<void>;
  logout: () => void;
}

const GoAuthContext = createContext<GoAuthContextType>({
  user: null,
  isLoggedIn: false,
  login: async () => {},
  register: async () => {},
  loginWithGoogle: async () => {},
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
  avatarUrl?: string;
}): GoAuthUser {
  setGoTokens(data.accessToken, data.refreshToken);
  const authUser: GoAuthUser = {
    userId: data.userId,
    email: data.email,
    name: data.name,
    avatarUrl: data.avatarUrl,
  };
  localStorage.setItem("go_user", JSON.stringify(authUser));
  return authUser;
}

export function GoAuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<GoAuthUser | null>(() => {
    if (typeof window === "undefined" || !checkIsLoggedIn()) return null;
    const stored = localStorage.getItem("go_user");
    return stored ? JSON.parse(stored) : null;
  });
  const [isAuthenticated, setIsAuthenticated] = useState(
    () => typeof window !== "undefined" && checkIsLoggedIn(),
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

  const register = useCallback(async (email: string, password: string, name: string) => {
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
  }, []);

  const loginWithGoogle = useCallback(async (code: string, redirectUri: string) => {
    const res = await fetch(`${GO_AUTH_URL}/auth/google`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ code, redirectUri }),
    });
    if (!res.ok) {
      const errorText = await res.text();
      throw new Error(errorText || "Google sign-in failed");
    }
    const data = await res.json();
    const authUser = handleAuthResponse(data);
    setUser(authUser);
    setIsAuthenticated(true);
  }, []);

  const logout = useCallback(() => {
    clearGoTokens();
    localStorage.removeItem("go_user");
    setUser(null);
    setIsAuthenticated(false);
  }, []);

  return (
    <GoAuthContext.Provider
      value={{ user, isLoggedIn: isAuthenticated, login, register, loginWithGoogle, logout }}
    >
      {children}
    </GoAuthContext.Provider>
  );
}
```

- [ ] **Step 2: Type-check**

```bash
cd frontend && npx tsc --noEmit
```

Expected: no errors (consumers of `useGoAuth()` that destructure only existing fields are unaffected by the added field).

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/go/GoAuthProvider.tsx
git commit -m "feat(frontend): add loginWithGoogle and avatarUrl to GoAuthProvider"
```

---

## Task 12: Frontend — `GoGoogleLoginButton` component

**Files:**
- Create: `frontend/src/components/go/GoGoogleLoginButton.tsx`

- [ ] **Step 1: Create the component**

Create `frontend/src/components/go/GoGoogleLoginButton.tsx`:

```tsx
"use client";

import { useCallback } from "react";
import { Button } from "@/components/ui/button";
import { GOOGLE_CLIENT_ID } from "@/lib/auth";

export function GoGoogleLoginButton() {
  const handleLogin = useCallback(() => {
    const redirectUri = `${window.location.origin}/go/login`;
    const params = new URLSearchParams({
      client_id: GOOGLE_CLIENT_ID,
      redirect_uri: redirectUri,
      response_type: "code",
      scope: "openid email profile",
      access_type: "offline",
      prompt: "consent",
    });
    window.location.href = `https://accounts.google.com/o/oauth2/v2/auth?${params}`;
  }, []);

  return (
    <Button onClick={handleLogin} size="lg" variant="outline" className="w-full">
      Sign in with Google
    </Button>
  );
}
```

- [ ] **Step 2: Type-check**

```bash
cd frontend && npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/go/GoGoogleLoginButton.tsx
git commit -m "feat(frontend): add GoGoogleLoginButton for Go stack"
```

---

## Task 13: Frontend — `/go/login` page with callback handling

**Files:**
- Create: `frontend/src/app/go/login/page.tsx`

- [ ] **Step 1: Create the login page**

Create `frontend/src/app/go/login/page.tsx`:

```tsx
"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import { GoGoogleLoginButton } from "@/components/go/GoGoogleLoginButton";
import { useGoAuth } from "@/components/go/GoAuthProvider";

export default function GoLoginPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { login, loginWithGoogle } = useGoAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  // Handle Google OAuth callback
  useEffect(() => {
    const code = searchParams.get("code");
    if (!code) return;
    let cancelled = false;
    (async () => {
      setBusy(true);
      try {
        const redirectUri = `${window.location.origin}/go/login`;
        await loginWithGoogle(code, redirectUri);
        if (!cancelled) router.replace("/go/ecommerce");
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : "Google sign-in failed");
          // Strip the code from the URL so refresh doesn't retry
          router.replace("/go/login");
        }
      } finally {
        if (!cancelled) setBusy(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [searchParams, loginWithGoogle, router]);

  const handleSubmit = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault();
      setError("");
      setBusy(true);
      try {
        await login(email, password);
        router.push("/go/ecommerce");
      } catch (err) {
        setError(err instanceof Error ? err.message : "Login failed");
      } finally {
        setBusy(false);
      }
    },
    [email, password, login, router],
  );

  return (
    <div className="mx-auto max-w-sm px-6 py-12">
      <h1 className="text-2xl font-bold">Sign in</h1>
      <form onSubmit={handleSubmit} className="mt-6 space-y-3">
        <input
          type="email"
          placeholder="Email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          required
          disabled={busy}
          className="w-full rounded border border-foreground/20 bg-background px-3 py-2 text-sm"
        />
        <input
          type="password"
          placeholder="Password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
          disabled={busy}
          className="w-full rounded border border-foreground/20 bg-background px-3 py-2 text-sm"
        />
        <Button type="submit" disabled={busy} className="w-full">
          {busy ? "Signing in…" : "Sign in"}
        </Button>
      </form>

      <div className="my-6 flex items-center gap-3 text-xs text-muted-foreground">
        <div className="h-px flex-1 bg-foreground/10" />
        <span>or</span>
        <div className="h-px flex-1 bg-foreground/10" />
      </div>

      <GoGoogleLoginButton />

      {error && <p className="mt-4 text-sm text-red-500">{error}</p>}

      <p className="mt-6 text-center text-sm text-muted-foreground">
        No account?{" "}
        <Link href="/go/register" className="underline hover:text-foreground">
          Register
        </Link>
      </p>
    </div>
  );
}
```

- [ ] **Step 2: Type-check and lint**

```bash
cd frontend && npx tsc --noEmit && npm run lint
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/app/go/login/page.tsx
git commit -m "feat(frontend): add /go/login page with email + Google sign-in"
```

---

## Task 14: Frontend — `/go/register` page

**Files:**
- Create: `frontend/src/app/go/register/page.tsx`

- [ ] **Step 1: Create the register page**

Create `frontend/src/app/go/register/page.tsx`:

```tsx
"use client";

import { useCallback, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { GoGoogleLoginButton } from "@/components/go/GoGoogleLoginButton";
import { useGoAuth } from "@/components/go/GoAuthProvider";

export default function GoRegisterPage() {
  const router = useRouter();
  const { register } = useGoAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [name, setName] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const handleSubmit = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault();
      setError("");
      setBusy(true);
      try {
        await register(email, password, name);
        router.push("/go/ecommerce");
      } catch (err) {
        setError(err instanceof Error ? err.message : "Registration failed");
      } finally {
        setBusy(false);
      }
    },
    [email, password, name, register, router],
  );

  return (
    <div className="mx-auto max-w-sm px-6 py-12">
      <h1 className="text-2xl font-bold">Create account</h1>
      <form onSubmit={handleSubmit} className="mt-6 space-y-3">
        <input
          type="text"
          placeholder="Name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          required
          disabled={busy}
          className="w-full rounded border border-foreground/20 bg-background px-3 py-2 text-sm"
        />
        <input
          type="email"
          placeholder="Email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          required
          disabled={busy}
          className="w-full rounded border border-foreground/20 bg-background px-3 py-2 text-sm"
        />
        <input
          type="password"
          placeholder="Password (min 8 chars)"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
          minLength={8}
          disabled={busy}
          className="w-full rounded border border-foreground/20 bg-background px-3 py-2 text-sm"
        />
        <Button type="submit" disabled={busy} className="w-full">
          {busy ? "Creating account…" : "Create account"}
        </Button>
      </form>

      <div className="my-6 flex items-center gap-3 text-xs text-muted-foreground">
        <div className="h-px flex-1 bg-foreground/10" />
        <span>or</span>
        <div className="h-px flex-1 bg-foreground/10" />
      </div>

      <GoGoogleLoginButton />

      {error && <p className="mt-4 text-sm text-red-500">{error}</p>}

      <p className="mt-6 text-center text-sm text-muted-foreground">
        Already have an account?{" "}
        <Link href="/go/login" className="underline hover:text-foreground">
          Sign in
        </Link>
      </p>
    </div>
  );
}
```

- [ ] **Step 2: Type-check and lint**

```bash
cd frontend && npx tsc --noEmit && npm run lint
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/app/go/register/page.tsx
git commit -m "feat(frontend): add /go/register page with email + Google sign-in"
```

---

## Task 15: Full preflight and manual smoke test

- [ ] **Step 1: Run Go preflight**

```bash
make preflight-go
```

Expected: PASS.

- [ ] **Step 2: Run frontend preflight**

```bash
make preflight-frontend
```

Expected: PASS.

- [ ] **Step 3: Manual smoke test — email path**

1. Start `frontend` (`cd frontend && npm run dev`) and ensure `go/` compose is running.
2. Navigate to `http://localhost:3000/go/login`.
3. Create a test account via `/go/register` (email, name, password ≥8 chars).
4. Verify you're redirected to `/go/ecommerce` and can see products.
5. Sign out (via existing ecommerce page UI).
6. Go back to `/go/login`, sign in with the same email/password.
7. Verify redirect to `/go/ecommerce`.

- [ ] **Step 4: Manual smoke test — Google path**

1. Navigate to `http://localhost:3000/go/login`.
2. Click "Sign in with Google".
3. Complete consent with a Google account. Google should redirect back to `http://localhost:3000/go/login?code=...`.
4. Verify automatic redirect to `/go/ecommerce`.
5. Open browser devtools → Application → Local Storage. Verify `go_user` contains `avatarUrl` populated with a Google URL.
6. Reload `/go/ecommerce` and verify the session persists.
7. Sign out, then sign in with Google again. Verify it works and the existing user's `avatar_url` / `name` are updated if Google changed them.

- [ ] **Step 5: Database sanity check**

```bash
cd go && docker compose exec -T postgres psql -U taskuser -d ecommercedb \
  -c "SELECT email, name, avatar_url, password_hash IS NULL AS google_only FROM users;"
```

Expected: Google-created users have `google_only = t` and a non-null `avatar_url`. Email-created users have `google_only = f`.

- [ ] **Step 6: If all smoke tests pass, final commit is not needed (plan tasks commit as they go)**

Inform Kyle that the feature is ready for review, push, and merge per his branching workflow (`feat/ui-overhaul` → CI → merge to staging → merge to main).

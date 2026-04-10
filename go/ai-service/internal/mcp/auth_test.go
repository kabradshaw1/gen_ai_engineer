package mcp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"github.com/kabradshaw1/portfolio/go/ai-service/internal/jwtctx"
)

func TestOptionalJWTMiddleware_WithValidToken(t *testing.T) {
	secret := "test-secret"
	token := testJWT(t, secret, "user-99")

	var capturedUserID string
	var capturedJWT string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = UserIDFromContext(r.Context())
		capturedJWT = jwtctx.FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := OptionalJWTMiddleware(secret)(inner)

	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if capturedUserID != "user-99" {
		t.Errorf("expected userID 'user-99', got %q", capturedUserID)
	}
	if capturedJWT != token {
		t.Error("expected JWT in context")
	}
}

func TestOptionalJWTMiddleware_WithoutToken(t *testing.T) {
	var capturedUserID string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := OptionalJWTMiddleware("secret")(inner)

	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if capturedUserID != "" {
		t.Errorf("expected empty userID, got %q", capturedUserID)
	}
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestOptionalJWTMiddleware_WithInvalidToken(t *testing.T) {
	var capturedUserID string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := OptionalJWTMiddleware("secret")(inner)

	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if capturedUserID != "" {
		t.Errorf("expected empty userID for invalid token, got %q", capturedUserID)
	}
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

// testJWT creates a valid HS256 JWT with the given sub claim.
func testJWT(t *testing.T, secret, sub string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": sub})
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign JWT: %v", err)
	}
	return signed
}

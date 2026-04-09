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

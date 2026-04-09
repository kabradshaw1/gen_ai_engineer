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

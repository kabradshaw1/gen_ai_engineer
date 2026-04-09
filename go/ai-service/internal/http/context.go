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

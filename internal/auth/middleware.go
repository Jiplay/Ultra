package auth

import (
	"context"
	"net/http"

	"ultra-bis/internal/httputil"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// EmailKey is the context key for storing authenticated user email
	EmailKey contextKey = "email"
)

// JWTMiddleware is a middleware that validates JWT tokens
func JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := ExtractTokenFromHeader(r)
		if tokenString == "" {
			httputil.WriteError(w, http.StatusUnauthorized, "Missing authorization token")
			return
		}

		claims, err := ValidateToken(tokenString)
		if err != nil {
			httputil.WriteError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		// Add user info to context using typed keys
		ctx := httputil.SetUserID(r.Context(), claims.UserID)
		ctx = context.WithValue(ctx, EmailKey, claims.Email)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

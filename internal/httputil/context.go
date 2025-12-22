package httputil

import (
	"context"
	"net/http"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// Context keys used throughout the application
const (
	// UserIDKey is the context key for storing authenticated user ID
	UserIDKey contextKey = "user_id"
)

// GetUserID extracts the user ID from the request context
// Returns the user ID and true if found, 0 and false otherwise
func GetUserID(r *http.Request) (uint, bool) {
	userID, ok := r.Context().Value(UserIDKey).(uint)
	return userID, ok
}

// SetUserID creates a new context with the user ID set
func SetUserID(ctx context.Context, userID uint) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

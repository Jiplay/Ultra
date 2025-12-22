package httputil

import (
	"context"
	"net/http"
)

// contextKey for storing extracted path values
const (
	// PathIDKey stores the primary resource ID from the URL path
	PathIDKey contextKey = "path_id"

	// SecondaryPathIDKey stores the secondary resource ID from the URL path
	SecondaryPathIDKey contextKey = "secondary_path_id"
)

// RequireAuth middleware ensures a user is authenticated
// Returns 401 Unauthorized if user_id is not in context
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, ok := GetUserID(r)
		if !ok {
			WriteError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		next.ServeHTTP(w, r)
	}
}

// ExtractPathID middleware extracts an ID from the URL path and adds it to context
// The idPosition parameter indicates which path segment contains the ID
// Returns a middleware function that can be chained
func ExtractPathID(idPosition int) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			id, err := ExtractIDFromPath(r, idPosition)
			if err != nil {
				WriteError(w, http.StatusBadRequest, "Invalid or missing ID in path")
				return
			}

			ctx := context.WithValue(r.Context(), PathIDKey, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}

// GetPathID retrieves the primary ID stored by ExtractPathID middleware
func GetPathID(r *http.Request) (int, bool) {
	id, ok := r.Context().Value(PathIDKey).(int)
	return id, ok
}

// GetSecondaryPathID retrieves the secondary ID stored by ExtractTwoPathIDs middleware
func GetSecondaryPathID(r *http.Request) (int, bool) {
	id, ok := r.Context().Value(SecondaryPathIDKey).(int)
	return id, ok
}

// ExtractTwoPathIDs middleware extracts two IDs from the URL path
// Returns a middleware function that can be chained
func ExtractTwoPathIDs(firstIDPosition, secondIDPosition int) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			firstID, secondID, err := ExtractTwoIDsFromPath(r, firstIDPosition, secondIDPosition)
			if err != nil {
				WriteError(w, http.StatusBadRequest, "Invalid or missing IDs in path")
				return
			}

			ctx := context.WithValue(r.Context(), PathIDKey, firstID)
			ctx = context.WithValue(ctx, SecondaryPathIDKey, secondID)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}

// MethodFilter middleware ensures the request uses the specified HTTP method
func MethodFilter(method string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		next.ServeHTTP(w, r)
	}
}

// ChainMiddleware chains multiple middleware functions together
func ChainMiddleware(handler http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	// Apply middlewares in reverse order so the first middleware in the list is executed first
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

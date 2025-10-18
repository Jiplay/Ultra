package auth

import "net/http"

// RegisterRoutes registers all auth-related routes to the provided mux
func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	// Public routes
	mux.HandleFunc("/auth/register", handler.Register)
	mux.HandleFunc("/auth/login", handler.Login)

	// Protected routes
	mux.HandleFunc("/auth/me", JWTMiddleware(handler.GetMe))
	mux.HandleFunc("/users/profile", JWTMiddleware(handler.UpdateProfile))
}

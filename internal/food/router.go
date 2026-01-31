package food

import (
	"net/http"

	"ultra-bis/internal/auth"
)

// RegisterRoutes registers all food-related routes to the provided mux
func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	// User food routes (protected - require authentication)
	mux.HandleFunc("/foods", auth.JWTMiddleware(handler.handleFoods))
	mux.HandleFunc("/foods/", auth.JWTMiddleware(handler.handleFoodsWithID))

	// General foods reference data (public - no authentication required)
	mux.HandleFunc("GET /general-foods", handler.SearchGeneralFoods)
}

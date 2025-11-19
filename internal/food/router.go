package food

import (
	"net/http"

	"ultra-bis/internal/auth"
)

// RegisterRoutes registers all food-related routes to the provided mux (all protected)
func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	// All food routes require authentication - middleware applied once per route
	// Register specific routes before the wildcard route
	mux.HandleFunc("/foods/barcode/", auth.JWTMiddleware(handler.ScanBarcode))
	mux.HandleFunc("/foods", auth.JWTMiddleware(handler.handleFoods))
	mux.HandleFunc("/foods/", auth.JWTMiddleware(handler.handleFoodsWithID))
}

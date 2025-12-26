package barcode

import (
	"net/http"
	"ultra-bis/internal/auth"
)

// RegisterRoutes registers barcode-related routes
func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	// Protected search endpoint - requires JWT authentication
	mux.HandleFunc("/openfoodfacts/search", auth.JWTMiddleware(handler.SearchProducts))

	// Protected barcode scan endpoint - requires JWT authentication
	mux.HandleFunc("/openfoodfacts/barcode/", auth.JWTMiddleware(handler.ScanBarcode))
}

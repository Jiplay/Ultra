package barcode

import (
	"net/http"
	"strconv"
	"strings"
	"ultra-bis/internal/httputil"
)

// Handler handles HTTP requests for barcode and Open Food Facts operations
type Handler struct {
	service *Service
}

// NewHandler creates a new barcode handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// SearchProducts handles GET /openfoodfacts/search
func (h *Handler) SearchProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract query parameter
	query := r.URL.Query().Get("q")
	if query == "" {
		httputil.WriteError(w, http.StatusBadRequest, "Query parameter 'q' is required")
		return
	}

	// Extract optional pagination parameters
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 20
	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// Search Open Food Facts
	results, err := h.service.SearchByName(query, page, pageSize)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "Failed to search products: "+err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusOK, results)
}

// ScanBarcode handles POST /openfoodfacts/barcode/{code}
func (h *Handler) ScanBarcode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract barcode from URL path
	code := strings.TrimPrefix(r.URL.Path, "/openfoodfacts/barcode/")
	code = strings.TrimSpace(code)

	if code == "" {
		httputil.WriteError(w, http.StatusBadRequest, "Barcode is required")
		return
	}

	// Fetch product data from Open Food Facts
	productData, err := h.service.ScanBarcode(code)
	if err != nil {
		if strings.Contains(err.Error(), "product not found") {
			httputil.WriteError(w, http.StatusNotFound, "Product not found for barcode: "+code)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "Failed to scan barcode: "+err.Error())
		return
	}

	// Return product data (do NOT create food)
	httputil.WriteJSON(w, http.StatusOK, productData)
}

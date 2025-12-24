package food

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"ultra-bis/internal/httputil"

	"ultra-bis/internal/barcode"
)

// Handler handles HTTP requests for food resources
type Handler struct {
	repo           *Repository
	barcodeService *barcode.Service
}

// NewHandler creates a new food handler
func NewHandler(repo *Repository, barcodeService *barcode.Service) *Handler {
	return &Handler{
		repo:           repo,
		barcodeService: barcodeService,
	}
}

// CreateFood handles POST /foods
func (h *Handler) CreateFood(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req CreateFoodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Basic validation
	if req.Name == "" {
		httputil.WriteError(w, http.StatusBadRequest, "Name is required")
		return
	}

	// Validate tag (default to "routine" if empty)
	if req.Tag == "" {
		req.Tag = TagRoutine
	} else if !ValidateTag(req.Tag) {
		httputil.WriteError(w, http.StatusBadRequest, "Tag must be 'routine' or 'contextual'")
		return
	}

	food, err := h.repo.Create(req)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, food)
}

// GetFood handles GET /foods/{id}
func (h *Handler) GetFood(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id, err := extractID(r.URL.Path, "/foods/")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	food, err := h.repo.GetByID(id)
	if err != nil {
		if err.Error() == "food not found" {
			httputil.WriteError(w, http.StatusNotFound, "Food not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusOK, food)
}

// GetAllFoods handles GET /foods
func (h *Handler) GetAllFoods(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	foods, err := h.repo.GetAll()
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusOK, foods)
}

// UpdateFood handles PUT /foods/{id}
func (h *Handler) UpdateFood(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id, err := extractID(r.URL.Path, "/foods/")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var req UpdateFoodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Basic validation
	if req.Name == "" {
		httputil.WriteError(w, http.StatusBadRequest, "Name is required")
		return
	}

	// Validate tag if provided
	if req.Tag != "" && !ValidateTag(req.Tag) {
		httputil.WriteError(w, http.StatusBadRequest, "Tag must be 'routine' or 'contextual'")
		return
	}

	food, err := h.repo.Update(id, req)
	if err != nil {
		if err.Error() == "food not found" {
			httputil.WriteError(w, http.StatusNotFound, "Food not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusOK, food)
}

// DeleteFood handles DELETE /foods/{id}
func (h *Handler) DeleteFood(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id, err := extractID(r.URL.Path, "/foods/")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	err = h.repo.Delete(id)
	if err != nil {
		if err.Error() == "food not found" {
			httputil.WriteError(w, http.StatusNotFound, "Food not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// extractID extracts the ID from the URL path
func extractID(path, prefix string) (int, error) {
	idStr := strings.TrimPrefix(path, prefix)
	return strconv.Atoi(idStr)
}

// handleFoods routes /foods requests by HTTP method
func (h *Handler) handleFoods(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetAllFoods(w, r)
	case http.MethodPost:
		h.CreateFood(w, r)
	default:
		httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// GetFoodsByTag handles GET /foods/{filter} where filter = routine|contextual
func (h *Handler) GetFoodsByTag(w http.ResponseWriter, r *http.Request, tag string) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Validate tag
	if !ValidateTag(tag) {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid tag filter")
		return
	}

	foods, err := h.repo.GetByTag(tag)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusOK, foods)
}

// handleFoodsWithID routes /foods/{id} requests by HTTP method
func (h *Handler) handleFoodsWithID(w http.ResponseWriter, r *http.Request) {
	// Check if this is actually a request to /foods (no ID)
	if strings.TrimPrefix(r.URL.Path, "/foods/") == "" {
		h.GetAllFoods(w, r)
		return
	}

	pathSegment := strings.TrimPrefix(r.URL.Path, "/foods/")

	// Try to parse as numeric ID first
	_, err := strconv.Atoi(pathSegment)

	// If not numeric, check if it's a valid tag filter
	if err != nil {
		if pathSegment == TagRoutine || pathSegment == TagContextual {
			h.GetFoodsByTag(w, r, pathSegment)
			return
		}
		httputil.WriteError(w, http.StatusBadRequest, "Invalid ID or filter")
		return
	}

	// It's a numeric ID, proceed with ID-based operations
	// Store the ID in the path for the handlers to extract
	switch r.Method {
	case http.MethodGet:
		h.GetFood(w, r)
	case http.MethodPut:
		h.UpdateFood(w, r)
	case http.MethodDelete:
		h.DeleteFood(w, r)
	default:
		httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// ScanBarcode handles POST /foods/barcode/{code}
// Scans a barcode using Open Food Facts API and creates a food entry
func (h *Handler) ScanBarcode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract barcode from URL path
	// Expected path: /foods/barcode/{code}
	path := strings.TrimPrefix(r.URL.Path, "/foods/barcode/")
	fetchedBarcode := strings.TrimSpace(path)

	if fetchedBarcode == "" {
		httputil.WriteError(w, http.StatusBadRequest, "Barcode is required")
		return
	}

	// Fetch product data from Open Food Facts
	productData, err := h.barcodeService.ScanBarcode(fetchedBarcode)
	if err != nil {
		if strings.Contains(err.Error(), "product not found") {
			httputil.WriteError(w, http.StatusNotFound, "Product not found for barcode: "+fetchedBarcode)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "Failed to scan barcode: "+err.Error())
		return
	}

	// Validate product data
	if productData.Name == "" {
		httputil.WriteError(w, http.StatusBadRequest, "Product name is missing from barcode data")
		return
	}

	// Create food item from scanned data
	createReq := CreateFoodRequest{
		Name:        productData.Name,
		Description: productData.Description,
		Calories:    productData.Calories,
		Protein:     productData.Protein,
		Carbs:       productData.Carbs,
		Fat:         productData.Fat,
		Fiber:       productData.Fiber,
	}

	food, err := h.repo.Create(createReq)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "Failed to create food: "+err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, food)
}

package food

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

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

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}

// CreateFood handles POST /foods
func (h *Handler) CreateFood(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req CreateFoodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Basic validation
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "Name is required")
		return
	}

	food, err := h.repo.Create(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, food)
}

// GetFood handles GET /foods/{id}
func (h *Handler) GetFood(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id, err := extractID(r.URL.Path, "/foods/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	food, err := h.repo.GetByID(id)
	if err != nil {
		if err.Error() == "food not found" {
			writeError(w, http.StatusNotFound, "Food not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, food)
}

// GetAllFoods handles GET /foods
func (h *Handler) GetAllFoods(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	foods, err := h.repo.GetAll()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, foods)
}

// UpdateFood handles PUT /foods/{id}
func (h *Handler) UpdateFood(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id, err := extractID(r.URL.Path, "/foods/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var req UpdateFoodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Basic validation
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "Name is required")
		return
	}

	food, err := h.repo.Update(id, req)
	if err != nil {
		if err.Error() == "food not found" {
			writeError(w, http.StatusNotFound, "Food not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, food)
}

// DeleteFood handles DELETE /foods/{id}
func (h *Handler) DeleteFood(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id, err := extractID(r.URL.Path, "/foods/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	err = h.repo.Delete(id)
	if err != nil {
		if err.Error() == "food not found" {
			writeError(w, http.StatusNotFound, "Food not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
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
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleFoodsWithID routes /foods/{id} requests by HTTP method
func (h *Handler) handleFoodsWithID(w http.ResponseWriter, r *http.Request) {
	// Check if this is actually a request to /foods (no ID)
	if strings.TrimPrefix(r.URL.Path, "/foods/") == "" {
		h.GetAllFoods(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.GetFood(w, r)
	case http.MethodPut:
		h.UpdateFood(w, r)
	case http.MethodDelete:
		h.DeleteFood(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// ScanBarcode handles POST /foods/barcode/{code}
// Scans a barcode using Open Food Facts API and creates a food entry
func (h *Handler) ScanBarcode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract barcode from URL path
	// Expected path: /foods/barcode/{code}
	path := strings.TrimPrefix(r.URL.Path, "/foods/barcode/")
	barcode := strings.TrimSpace(path)

	if barcode == "" {
		writeError(w, http.StatusBadRequest, "Barcode is required")
		return
	}

	// Fetch product data from Open Food Facts
	productData, err := h.barcodeService.ScanBarcode(barcode)
	if err != nil {
		if strings.Contains(err.Error(), "product not found") {
			writeError(w, http.StatusNotFound, "Product not found for barcode: "+barcode)
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to scan barcode: "+err.Error())
		return
	}

	// Validate product data
	if productData.Name == "" {
		writeError(w, http.StatusBadRequest, "Product name is missing from barcode data")
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
		writeError(w, http.StatusInternalServerError, "Failed to create food: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, food)
}

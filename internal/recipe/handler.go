package recipe

import (
	"encoding/json"
	"errors"
	"net/http"

	"ultra-bis/internal/httputil"
)

// Handler handles recipe HTTP requests
type Handler struct {
	service *Service
}

// NewHandler creates a new recipe handler with the service layer
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// CreateRecipe handles POST /recipes (Protected)
func (h *Handler) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	userID, ok := httputil.GetUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	recipe, err := h.service.CreateRecipe(r.Context(), userID, req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, recipe)
}

// GetRecipe handles GET /recipes/{id} (Protected)
func (h *Handler) GetRecipe(w http.ResponseWriter, r *http.Request) {
	userID, ok := httputil.GetUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	recipeID, ok := httputil.GetPathID(r)
	if !ok {
		httputil.WriteError(w, http.StatusBadRequest, "Recipe ID required")
		return
	}

	recipe, err := h.service.GetRecipe(r.Context(), userID, recipeID)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, recipe)
}

// ListRecipes handles GET /recipes (Protected)
func (h *Handler) ListRecipes(w http.ResponseWriter, r *http.Request) {
	userID, ok := httputil.GetUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userOnlyParam := r.URL.Query().Get("user_only")
	userOnly := userOnlyParam == "true"

	recipes, err := h.service.ListRecipes(r.Context(), userID, userOnly)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, recipes)
}

// ListRecipesByTag handles GET /recipes/{filter} where filter = routine|contextual
func (h *Handler) ListRecipesByTag(w http.ResponseWriter, r *http.Request) {
	userID, ok := httputil.GetUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Extract tag from path ID context (set by router)
	tagInterface := r.Context().Value("tag_filter")
	tag, ok := tagInterface.(string)
	if !ok {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid tag filter")
		return
	}

	userOnlyParam := r.URL.Query().Get("user_only")
	userOnly := userOnlyParam == "true"

	recipes, err := h.service.ListRecipesByTag(r.Context(), userID, tag, userOnly)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, recipes)
}

// UpdateRecipe handles PUT /recipes/{id}
func (h *Handler) UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	userID, ok := httputil.GetUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	recipeID, ok := httputil.GetPathID(r)
	if !ok {
		httputil.WriteError(w, http.StatusBadRequest, "Recipe ID required")
		return
	}

	var req UpdateRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	recipe, err := h.service.UpdateRecipe(r.Context(), userID, recipeID, req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, recipe)
}

// DeleteRecipe handles DELETE /recipes/{id}
func (h *Handler) DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	userID, ok := httputil.GetUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	recipeID, ok := httputil.GetPathID(r)
	if !ok {
		httputil.WriteError(w, http.StatusBadRequest, "Recipe ID required")
		return
	}

	err := h.service.DeleteRecipe(r.Context(), userID, recipeID)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	httputil.WriteSuccess(w, http.StatusOK, "Recipe deleted successfully")
}

// AddIngredient handles POST /recipes/{id}/ingredients
func (h *Handler) AddIngredient(w http.ResponseWriter, r *http.Request) {
	userID, ok := httputil.GetUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	recipeID, ok := httputil.GetPathID(r)
	if !ok {
		httputil.WriteError(w, http.StatusBadRequest, "Recipe ID required")
		return
	}

	var req AddIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ingredient, err := h.service.AddIngredient(r.Context(), userID, recipeID, req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, ingredient)
}

// UpdateIngredient handles PUT /recipes/{recipeId}/ingredients/{ingredientId}
func (h *Handler) UpdateIngredient(w http.ResponseWriter, r *http.Request) {
	userID, ok := httputil.GetUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	recipeID, ok := httputil.GetPathID(r)
	if !ok {
		httputil.WriteError(w, http.StatusBadRequest, "Recipe ID required")
		return
	}

	ingredientID, ok := httputil.GetSecondaryPathID(r)
	if !ok {
		httputil.WriteError(w, http.StatusBadRequest, "Ingredient ID required")
		return
	}

	var req UpdateIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ingredient, err := h.service.UpdateIngredient(r.Context(), userID, recipeID, ingredientID, req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, ingredient)
}

// DeleteIngredient handles DELETE /recipes/{recipeId}/ingredients/{ingredientId}
func (h *Handler) DeleteIngredient(w http.ResponseWriter, r *http.Request) {
	userID, ok := httputil.GetUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	recipeID, ok := httputil.GetPathID(r)
	if !ok {
		httputil.WriteError(w, http.StatusBadRequest, "Recipe ID required")
		return
	}

	ingredientID, ok := httputil.GetSecondaryPathID(r)
	if !ok {
		httputil.WriteError(w, http.StatusBadRequest, "Ingredient ID required")
		return
	}

	err := h.service.DeleteIngredient(r.Context(), userID, recipeID, ingredientID)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	httputil.WriteSuccess(w, http.StatusOK, "Ingredient deleted successfully")
}

// handleServiceError maps service layer errors to appropriate HTTP status codes
func (h *Handler) handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrRecipeNotFound):
		httputil.WriteError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrIngredientNotFound):
		httputil.WriteError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrFoodNotFound):
		httputil.WriteError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrUnauthorized):
		httputil.WriteError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, ErrForbidden):
		httputil.WriteError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, ErrInvalidInput):
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
	default:
		httputil.WriteError(w, http.StatusInternalServerError, "Internal server error")
	}
}

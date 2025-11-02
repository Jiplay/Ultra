package recipe

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"ultra-bis/internal/food"
)

// Handler handles recipe requests
type Handler struct {
	repo     *Repository
	foodRepo *food.Repository
}

// NewHandler creates a new recipe handler
func NewHandler(repo *Repository, foodRepo *food.Repository) *Handler {
	return &Handler{
		repo:     repo,
		foodRepo: foodRepo,
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

// CreateRecipe handles POST /recipes (Protected)
func (h *Handler) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validation
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "Name is required")
		return
	}

	if req.ServingSize <= 0 {
		req.ServingSize = 1
	}

	// Create recipe (always belongs to the authenticated user)
	recipe := &Recipe{
		Name:        req.Name,
		ServingSize: req.ServingSize,
		UserID:      &userID,
	}

	if err := h.repo.Create(recipe); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Add ingredients if provided
	for _, ing := range req.Ingredients {
		if ing.Quantity <= 0 {
			continue
		}

		// Verify food exists
		if _, err := h.foodRepo.GetByID(int(ing.FoodID)); err != nil {
			writeError(w, http.StatusBadRequest, "Food ID "+strconv.Itoa(int(ing.FoodID))+" not found")
			return
		}

		ingredient := &RecipeIngredient{
			RecipeID: recipe.ID,
			FoodID:   ing.FoodID,
			Quantity: ing.Quantity,
		}

		if err := h.repo.AddIngredient(ingredient); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	// Reload recipe with ingredients
	createdRecipe, err := h.repo.GetByID(int(recipe.ID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, createdRecipe)
}

// GetRecipe handles GET /recipes/{id} (Protected)
func (h *Handler) GetRecipe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	_, ok := r.Context().Value("user_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Extract ID from path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		writeError(w, http.StatusBadRequest, "Recipe ID required")
		return
	}

	id, err := strconv.Atoi(pathParts[1])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid recipe ID")
		return
	}

	// Calculate nutrition
	recipeWithNutrition, err := h.repo.CalculateNutrition(id, h.foodRepo)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, recipeWithNutrition)
}

// ListRecipes handles GET /recipes (Protected)
func (h *Handler) ListRecipes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userOnlyParam := r.URL.Query().Get("user_only")
	userOnly := userOnlyParam == "true"

	// Get recipes with nutrition and ingredient details
	recipes, err := h.repo.GetByUserIDWithNutrition(userID, userOnly, h.foodRepo)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, recipes)
}

// UpdateRecipe handles PUT /recipes/{id}
func (h *Handler) UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Extract ID from path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		writeError(w, http.StatusBadRequest, "Recipe ID required")
		return
	}

	id, err := strconv.Atoi(pathParts[1])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid recipe ID")
		return
	}

	// Check recipe exists and user owns it
	recipe, err := h.repo.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "Recipe not found")
		return
	}

	// Check ownership (only owner can update, global recipes can't be updated by users)
	if recipe.UserID == nil || *recipe.UserID != userID {
		writeError(w, http.StatusForbidden, "You don't have permission to update this recipe")
		return
	}

	var req UpdateRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Update fields
	if req.Name != "" {
		recipe.Name = req.Name
	}
	if req.ServingSize > 0 {
		recipe.ServingSize = req.ServingSize
	}

	if err := h.repo.Update(recipe); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, recipe)
}

// DeleteRecipe handles DELETE /recipes/{id}
func (h *Handler) DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Extract ID from path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		writeError(w, http.StatusBadRequest, "Recipe ID required")
		return
	}

	id, err := strconv.Atoi(pathParts[1])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid recipe ID")
		return
	}

	// Check recipe exists and user owns it
	recipe, err := h.repo.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "Recipe not found")
		return
	}

	// Check ownership
	if recipe.UserID == nil || *recipe.UserID != userID {
		writeError(w, http.StatusForbidden, "You don't have permission to delete this recipe")
		return
	}

	if err := h.repo.Delete(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Recipe deleted successfully"})
}

// AddIngredient handles POST /recipes/{id}/ingredients
func (h *Handler) AddIngredient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Extract recipe ID from path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		writeError(w, http.StatusBadRequest, "Recipe ID required")
		return
	}

	recipeID, err := strconv.Atoi(pathParts[1])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid recipe ID")
		return
	}

	// Check recipe exists and user owns it
	recipe, err := h.repo.GetByID(recipeID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Recipe not found")
		return
	}

	if recipe.UserID == nil || *recipe.UserID != userID {
		writeError(w, http.StatusForbidden, "You don't have permission to modify this recipe")
		return
	}

	var req AddIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Quantity <= 0 {
		writeError(w, http.StatusBadRequest, "Quantity must be greater than 0")
		return
	}

	// Verify food exists
	if _, err := h.foodRepo.GetByID(int(req.FoodID)); err != nil {
		writeError(w, http.StatusBadRequest, "Food not found")
		return
	}

	ingredient := &RecipeIngredient{
		RecipeID: uint(recipeID),
		FoodID:   req.FoodID,
		Quantity: req.Quantity,
	}

	if err := h.repo.AddIngredient(ingredient); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, ingredient)
}

// UpdateIngredient handles PUT /recipes/{recipeId}/ingredients/{ingredientId}
func (h *Handler) UpdateIngredient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Extract IDs from path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		writeError(w, http.StatusBadRequest, "Recipe ID and Ingredient ID required")
		return
	}

	recipeID, err := strconv.Atoi(pathParts[1])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid recipe ID")
		return
	}

	ingredientID, err := strconv.Atoi(pathParts[3])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid ingredient ID")
		return
	}

	// Check recipe exists and user owns it
	recipe, err := h.repo.GetByID(recipeID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Recipe not found")
		return
	}

	if recipe.UserID == nil || *recipe.UserID != userID {
		writeError(w, http.StatusForbidden, "You don't have permission to modify this recipe")
		return
	}

	// Get ingredient
	ingredient, err := h.repo.GetIngredient(ingredientID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Ingredient not found")
		return
	}

	// Verify ingredient belongs to this recipe
	if ingredient.RecipeID != uint(recipeID) {
		writeError(w, http.StatusBadRequest, "Ingredient does not belong to this recipe")
		return
	}

	var req UpdateIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Quantity <= 0 {
		writeError(w, http.StatusBadRequest, "Quantity must be greater than 0")
		return
	}

	ingredient.Quantity = req.Quantity

	if err := h.repo.UpdateIngredient(ingredient); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ingredient)
}

// DeleteIngredient handles DELETE /recipes/{recipeId}/ingredients/{ingredientId}
func (h *Handler) DeleteIngredient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Extract IDs from path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		writeError(w, http.StatusBadRequest, "Recipe ID and Ingredient ID required")
		return
	}

	recipeID, err := strconv.Atoi(pathParts[1])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid recipe ID")
		return
	}

	ingredientID, err := strconv.Atoi(pathParts[3])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid ingredient ID")
		return
	}

	// Check recipe exists and user owns it
	recipe, err := h.repo.GetByID(recipeID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Recipe not found")
		return
	}

	if recipe.UserID == nil || *recipe.UserID != userID {
		writeError(w, http.StatusForbidden, "You don't have permission to modify this recipe")
		return
	}

	// Get ingredient to verify it belongs to this recipe
	ingredient, err := h.repo.GetIngredient(ingredientID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Ingredient not found")
		return
	}

	if ingredient.RecipeID != uint(recipeID) {
		writeError(w, http.StatusBadRequest, "Ingredient does not belong to this recipe")
		return
	}

	if err := h.repo.DeleteIngredient(ingredientID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Ingredient deleted successfully"})
}

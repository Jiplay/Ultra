package diary

import (
	"ultra-bis/internal/httputil"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ultra-bis/internal/food"
	"ultra-bis/internal/goal"
)

// Handler handles diary entry requests
type Handler struct {
	repo       *Repository
	foodRepo   *food.Repository
	goalRepo   *goal.Repository
	recipeRepo RecipeRepository
}

// RecipeRepository interface for recipe operations needed by diary
type RecipeRepository interface {
	GetByID(id int) (Recipe, error)
	GetIngredients(recipeID int) ([]RecipeIngredient, error)
	CreateRecipe(userID uint, name string, tag string, ingredients []RecipeIngredientRequest) (RecipeCreatedResponse, error)
}

// RecipeIngredientRequest represents an ingredient request for recipe creation
type RecipeIngredientRequest struct {
	FoodID        uint
	QuantityGrams float64
}

// RecipeCreatedResponse represents the response after creating a recipe
type RecipeCreatedResponse struct {
	ID            uint
	Name          string
	Tag           string
	TotalCalories float64
	TotalProtein  float64
	TotalCarbs    float64
	TotalFat      float64
	TotalFiber    float64
}

// Recipe represents a recipe with basic info
type Recipe struct {
	ID   uint
	Name string
	Tag  string
}

// RecipeIngredient represents an ingredient in a recipe
type RecipeIngredient struct {
	FoodID        uint
	QuantityGrams float64
}

// NewHandler creates a new diary handler
func NewHandler(repo *Repository, foodRepo *food.Repository, goalRepo *goal.Repository) *Handler {
	return &Handler{
		repo:       repo,
		foodRepo:   foodRepo,
		goalRepo:   goalRepo,
		recipeRepo: nil, // Will be set via SetRecipeRepo
	}
}

// SetRecipeRepo sets the recipe repository (to avoid circular dependency)
func (h *Handler) SetRecipeRepo(recipeRepo RecipeRepository) {
	h.recipeRepo = recipeRepo
}


// CreateEntry handles POST /diary/entries
func (h *Handler) CreateEntry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := httputil.GetUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateDiaryEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validation: Count which entry type is being used
	entryTypes := 0
	if req.FoodID != nil {
		entryTypes++
	}
	if req.RecipeID != nil {
		entryTypes++
	}
	if req.InlineRecipeName != "" {
		entryTypes++
	}
	if req.InlineFoodName != "" {
		entryTypes++
	}

	// Require exactly one
	if entryTypes == 0 {
		httputil.WriteError(w, http.StatusBadRequest, "One of food_id, recipe_id, inline_recipe_name, or inline_food_name is required")
		return
	}

	if entryTypes > 1 {
		httputil.WriteError(w, http.StatusBadRequest, "Cannot specify multiple entry types")
		return
	}

	// Validate inline food fields if inline_food_name is provided
	if req.InlineFoodName != "" {
		// Validate nutrition values (must be non-negative)
		if req.InlineFoodCalories < 0 || req.InlineFoodProtein < 0 || req.InlineFoodCarbs < 0 ||
			req.InlineFoodFat < 0 || req.InlineFoodFiber < 0 {
			httputil.WriteError(w, http.StatusBadRequest, "Inline food nutrition values must be non-negative")
			return
		}

		if req.QuantityGrams <= 0 {
			httputil.WriteError(w, http.StatusBadRequest, "quantity_grams must be greater than 0 for inline food entries")
			return
		}

		// Validate tag if provided
		if req.InlineFoodTag != "" && req.InlineFoodTag != "routine" && req.InlineFoodTag != "contextual" {
			httputil.WriteError(w, http.StatusBadRequest, "inline_food_tag must be 'routine' or 'contextual'")
			return
		}
	}

	// For inline recipes, custom_ingredients are REQUIRED
	if req.InlineRecipeName != "" && len(req.CustomIngredients) == 0 {
		httputil.WriteError(w, http.StatusBadRequest, "custom_ingredients are required for inline recipes")
		return
	}

	// For food entries, quantity_grams is required
	if req.FoodID != nil && req.QuantityGrams <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "Quantity in grams must be greater than 0 for food entries")
		return
	}

	// For saved recipes, either quantity_grams or custom_ingredients required
	if req.RecipeID != nil && req.QuantityGrams <= 0 && len(req.CustomIngredients) == 0 {
		httputil.WriteError(w, http.StatusBadRequest, "Either quantity_grams or custom_ingredients is required for saved recipes")
		return
	}

	// Validate custom ingredient quantities
	for _, customIng := range req.CustomIngredients {
		if customIng.QuantityGrams <= 0 {
			httputil.WriteError(w, http.StatusBadRequest, "custom ingredient quantity must be greater than 0")
			return
		}
	}

	// Parse date
	var entryDate time.Time
	if req.Date == "" {
		entryDate = time.Now()
	} else {
		var err error
		entryDate, err = time.Parse("2006-01-02", req.Date)
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "Invalid date format (use YYYY-MM-DD)")
			return
		}
	}

	// Create entry
	entry := &DiaryEntry{
		UserID:        userID,
		FoodID:        req.FoodID,
		RecipeID:      req.RecipeID,
		Date:          entryDate,
		MealType:      req.MealType,
		QuantityGrams: req.QuantityGrams,
		Notes:         req.Notes,
	}

	// Calculate nutrition if food_id is provided (food nutrition is per 100g)
	if req.FoodID != nil {
		foodItem, err := h.foodRepo.GetByID(int(*req.FoodID))
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "Food not found")
			return
		}

		multiplier := req.QuantityGrams / 100.0
		entry.Calories = foodItem.Calories * multiplier
		entry.Protein = foodItem.Protein * multiplier
		entry.Carbs = foodItem.Carbs * multiplier
		entry.Fat = foodItem.Fat * multiplier
		entry.Fiber = foodItem.Fiber * multiplier
		entry.FoodTag = foodItem.Tag
	}

	// Calculate nutrition if recipe_id is provided
	if req.RecipeID != nil {
		if h.recipeRepo == nil {
			httputil.WriteError(w, http.StatusInternalServerError, "Recipe repository not initialized")
			return
		}

		recipe, err := h.recipeRepo.GetByID(int(*req.RecipeID))
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "Recipe not found")
			return
		}

		// Cache recipe tag
		entry.RecipeTag = recipe.Tag

		var customIngredients CustomIngredients
		var totalCalories, totalProtein, totalCarbs, totalFat, totalFiber, totalWeight float64

		// Use custom ingredients if provided, otherwise use proportional scaling
		if len(req.CustomIngredients) > 0 {
			// Validate custom ingredients belong to recipe
			if err := h.validateCustomIngredients(int(*req.RecipeID), req.CustomIngredients); err != nil {
				httputil.WriteError(w, http.StatusBadRequest, err.Error())
				return
			}

			// Calculate nutrition with custom quantities
			var calcErr error
			customIngredients, totalCalories, totalProtein, totalCarbs, totalFat, totalFiber, totalWeight, calcErr = h.calculateCustomIngredientsNutrition(req.CustomIngredients)
			if calcErr != nil {
				httputil.WriteError(w, http.StatusInternalServerError, "Failed to calculate nutrition: "+calcErr.Error())
				return
			}
		} else {
			// Convert proportional quantity to custom ingredients (backward compatibility)
			var calcErr error
			customIngredients, totalCalories, totalProtein, totalCarbs, totalFat, totalFiber, calcErr = h.convertProportionalToCustomIngredients(int(*req.RecipeID), req.QuantityGrams)
			if calcErr != nil {
				httputil.WriteError(w, http.StatusInternalServerError, "Failed to calculate nutrition: "+calcErr.Error())
				return
			}
			totalWeight = req.QuantityGrams
		}

		// Set nutrition values
		entry.Calories = roundToTwo(totalCalories)
		entry.Protein = roundToTwo(totalProtein)
		entry.Carbs = roundToTwo(totalCarbs)
		entry.Fat = roundToTwo(totalFat)
		entry.Fiber = roundToTwo(totalFiber)
		entry.QuantityGrams = roundToTwo(totalWeight)
		entry.CustomIngredients = customIngredients
	}

	// Handle inline recipes
	if req.InlineRecipeName != "" {
		// Calculate nutrition with custom quantities
		customIngredients, totalCalories, totalProtein, totalCarbs, totalFat, totalFiber, totalWeight, calcErr := h.calculateCustomIngredientsNutrition(req.CustomIngredients)
		if calcErr != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "Failed to calculate nutrition: "+calcErr.Error())
			return
		}

		// Set nutrition values
		entry.Calories = roundToTwo(totalCalories)
		entry.Protein = roundToTwo(totalProtein)
		entry.Carbs = roundToTwo(totalCarbs)
		entry.Fat = roundToTwo(totalFat)
		entry.Fiber = roundToTwo(totalFiber)
		entry.QuantityGrams = roundToTwo(totalWeight)
		entry.CustomIngredients = customIngredients
		entry.InlineRecipeName = &req.InlineRecipeName

		// Determine tag from ingredients
		entry.RecipeTag = h.determineInlineRecipeTag(customIngredients)
	}

	// Handle inline foods
	if req.InlineFoodName != "" {
		// Default tag to "routine" if not provided
		tag := req.InlineFoodTag
		if tag == "" {
			tag = "routine"
		}

		// Calculate nutrition (inline food values are per 100g)
		multiplier := req.QuantityGrams / 100.0

		entry.InlineFoodName = &req.InlineFoodName
		entry.InlineFoodCalories = &req.InlineFoodCalories
		entry.InlineFoodProtein = &req.InlineFoodProtein
		entry.InlineFoodCarbs = &req.InlineFoodCarbs
		entry.InlineFoodFat = &req.InlineFoodFat
		entry.InlineFoodFiber = &req.InlineFoodFiber
		entry.InlineFoodTag = &tag

		if req.InlineFoodDescription != "" {
			entry.InlineFoodDescription = &req.InlineFoodDescription
		}

		// Calculate and cache consumed nutrition
		entry.Calories = roundToTwo(req.InlineFoodCalories * multiplier)
		entry.Protein = roundToTwo(req.InlineFoodProtein * multiplier)
		entry.Carbs = roundToTwo(req.InlineFoodCarbs * multiplier)
		entry.Fat = roundToTwo(req.InlineFoodFat * multiplier)
		entry.Fiber = roundToTwo(req.InlineFoodFiber * multiplier)
		entry.FoodTag = tag // Cache tag for calorie breakdown
	}

	if err := h.repo.Create(entry); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, entry)
}

// GetEntries handles GET /diary/entries?date=YYYY-MM-DD
func (h *Handler) GetEntries(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := httputil.GetUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	dateStr := r.URL.Query().Get("date")
	var entryDate time.Time
	if dateStr == "" {
		entryDate = time.Now()
	} else {
		var err error
		entryDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "Invalid date format (use YYYY-MM-DD)")
			return
		}
	}

	entries, err := h.repo.GetByDate(userID, entryDate)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusOK, entries)
}

// GetDailySummary handles GET /diary/summary/{date}
func (h *Handler) GetDailySummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := httputil.GetUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	dateStr := strings.TrimPrefix(r.URL.Path, "/diary/summary/")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	entryDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid date format (use YYYY-MM-DD)")
		return
	}

	// Get entries for the day
	entries, err := h.repo.GetByDate(userID, entryDate)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Calculate calorie breakdown by tag
	routineCalories, contextualCalories, routinePercent, contextualPercent := calculateCaloriesByTag(entries)

	// Calculate totals
	summary, err := h.repo.GetDailySummary(userID, entryDate)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get active goal
	activeGoal, err := h.goalRepo.GetActive(userID)
	var goalCalories, goalProtein, goalCarbs, goalFat, goalFiber float64
	if err == nil {
		goalCalories = activeGoal.Calories
		goalProtein = activeGoal.Protein
		goalCarbs = activeGoal.Carbs
		goalFat = activeGoal.Fat
		goalFiber = activeGoal.Fiber
	}

	// Calculate adherence
	adherence := AdherencePercent{
		Calories: calculateAdherence(summary["calories"], goalCalories),
		Protein:  calculateAdherence(summary["protein"], goalProtein),
		Carbs:    calculateAdherence(summary["carbs"], goalCarbs),
		Fat:      calculateAdherence(summary["fat"], goalFat),
		Fiber:    calculateAdherence(summary["fiber"], goalFiber),
	}

	response := DailySummary{
		Date:          dateStr,
		TotalCalories: summary["calories"],
		TotalProtein:  summary["protein"],
		TotalCarbs:    summary["carbs"],
		TotalFat:      summary["fat"],
		TotalFiber:    summary["fiber"],
		GoalCalories:  goalCalories,
		GoalProtein:   goalProtein,
		GoalCarbs:     goalCarbs,
		GoalFat:       goalFat,
		GoalFiber:     goalFiber,
		Adherence:     adherence,
		RoutineCalories:    routineCalories,
		ContextualCalories: contextualCalories,
		RoutinePercent:     routinePercent,
		ContextualPercent:  contextualPercent,
		Entries:       entries,
	}

	httputil.WriteJSON(w, http.StatusOK, response)
}

// UpdateEntry handles PUT /diary/entries/{id}
func (h *Handler) UpdateEntry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := httputil.GetUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := extractID(r.URL.Path, "/diary/entries/")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var req UpdateDiaryEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	entry, err := h.repo.GetByID(uint(id), userID)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "Entry not found")
		return
	}

	// Update fields
	if entry.FoodID != nil {
		// Food entry - only update if quantity_grams is provided
		if req.QuantityGrams > 0 {
			foodItem, err := h.foodRepo.GetByID(int(*entry.FoodID))
			if err == nil {
				entry.QuantityGrams = req.QuantityGrams
				multiplier := req.QuantityGrams / 100.0
				entry.Calories = roundToTwo(foodItem.Calories * multiplier)
				entry.Protein = roundToTwo(foodItem.Protein * multiplier)
				entry.Carbs = roundToTwo(foodItem.Carbs * multiplier)
				entry.Fat = roundToTwo(foodItem.Fat * multiplier)
				entry.Fiber = roundToTwo(foodItem.Fiber * multiplier)
				entry.FoodTag = foodItem.Tag
			}
		}
	} else if entry.RecipeID != nil {
		// Recipe entry - support both custom ingredients and proportional scaling
		if len(req.CustomIngredients) > 0 {
			// Custom ingredients provided - validate and recalculate
			if err := h.validateCustomIngredients(int(*entry.RecipeID), req.CustomIngredients); err != nil {
				httputil.WriteError(w, http.StatusBadRequest, err.Error())
				return
			}

			customIngredients, totalCalories, totalProtein, totalCarbs, totalFat, totalFiber, totalWeight, calcErr := h.calculateCustomIngredientsNutrition(req.CustomIngredients)
			if calcErr != nil {
				httputil.WriteError(w, http.StatusInternalServerError, "Failed to calculate nutrition: "+calcErr.Error())
				return
			}

			entry.Calories = roundToTwo(totalCalories)
			entry.Protein = roundToTwo(totalProtein)
			entry.Carbs = roundToTwo(totalCarbs)
			entry.Fat = roundToTwo(totalFat)
			entry.Fiber = roundToTwo(totalFiber)
			entry.QuantityGrams = roundToTwo(totalWeight)
			entry.CustomIngredients = customIngredients
		} else if req.QuantityGrams > 0 {
			// Proportional scaling - convert to custom ingredients
			customIngredients, totalCalories, totalProtein, totalCarbs, totalFat, totalFiber, calcErr := h.convertProportionalToCustomIngredients(int(*entry.RecipeID), req.QuantityGrams)
			if calcErr != nil {
				httputil.WriteError(w, http.StatusInternalServerError, "Failed to calculate nutrition: "+calcErr.Error())
				return
			}

			entry.Calories = roundToTwo(totalCalories)
			entry.Protein = roundToTwo(totalProtein)
			entry.Carbs = roundToTwo(totalCarbs)
			entry.Fat = roundToTwo(totalFat)
			entry.Fiber = roundToTwo(totalFiber)
			entry.QuantityGrams = req.QuantityGrams
			entry.CustomIngredients = customIngredients
		}
	} else if entry.InlineRecipeName != nil {
		// Inline recipe entry - support custom ingredients update
		if len(req.CustomIngredients) > 0 {
			customIngredients, totalCalories, totalProtein, totalCarbs, totalFat, totalFiber, totalWeight, calcErr := h.calculateCustomIngredientsNutrition(req.CustomIngredients)
			if calcErr != nil {
				httputil.WriteError(w, http.StatusInternalServerError, "Failed to calculate nutrition: "+calcErr.Error())
				return
			}

			entry.Calories = roundToTwo(totalCalories)
			entry.Protein = roundToTwo(totalProtein)
			entry.Carbs = roundToTwo(totalCarbs)
			entry.Fat = roundToTwo(totalFat)
			entry.Fiber = roundToTwo(totalFiber)
			entry.QuantityGrams = roundToTwo(totalWeight)
			entry.CustomIngredients = customIngredients

			// Recalculate tag when ingredients change
			entry.RecipeTag = h.determineInlineRecipeTag(customIngredients)
		}
	} else if entry.InlineFoodName != nil {
		// Inline food entry - support updating inline food details

		// Update nutrition values if provided
		nutritionUpdated := false
		if req.InlineFoodCalories != nil {
			entry.InlineFoodCalories = req.InlineFoodCalories
			nutritionUpdated = true
		}
		if req.InlineFoodProtein != nil {
			entry.InlineFoodProtein = req.InlineFoodProtein
			nutritionUpdated = true
		}
		if req.InlineFoodCarbs != nil {
			entry.InlineFoodCarbs = req.InlineFoodCarbs
			nutritionUpdated = true
		}
		if req.InlineFoodFat != nil {
			entry.InlineFoodFat = req.InlineFoodFat
			nutritionUpdated = true
		}
		if req.InlineFoodFiber != nil {
			entry.InlineFoodFiber = req.InlineFoodFiber
			nutritionUpdated = true
		}
		if req.InlineFoodTag != nil {
			entry.InlineFoodTag = req.InlineFoodTag
			entry.FoodTag = *req.InlineFoodTag // Update cached tag
			nutritionUpdated = true
		}

		// Update description
		if req.InlineFoodDescription != nil {
			entry.InlineFoodDescription = req.InlineFoodDescription
		}

		// Update name
		if req.InlineFoodName != nil {
			entry.InlineFoodName = req.InlineFoodName
		}

		// Recalculate consumed nutrition if quantity or nutrition values changed
		if nutritionUpdated || req.QuantityGrams > 0 {
			if req.QuantityGrams > 0 {
				entry.QuantityGrams = req.QuantityGrams
			}

			multiplier := entry.QuantityGrams / 100.0
			entry.Calories = roundToTwo(*entry.InlineFoodCalories * multiplier)
			entry.Protein = roundToTwo(*entry.InlineFoodProtein * multiplier)
			entry.Carbs = roundToTwo(*entry.InlineFoodCarbs * multiplier)
			entry.Fat = roundToTwo(*entry.InlineFoodFat * multiplier)
			entry.Fiber = roundToTwo(*entry.InlineFoodFiber * multiplier)
		}
	}

	// Update inline recipe name if provided
	if req.InlineRecipeName != nil {
		entry.InlineRecipeName = req.InlineRecipeName
	}

	if req.MealType != "" {
		entry.MealType = req.MealType
	}
	if req.Notes != "" {
		entry.Notes = req.Notes
	}

	if err := h.repo.Update(entry); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusOK, entry)
}

// DeleteEntry handles DELETE /diary/entries/{id}
func (h *Handler) DeleteEntry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := httputil.GetUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := extractID(r.URL.Path, "/diary/entries/")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	if err := h.repo.Delete(uint(id), userID); err != nil {
		httputil.WriteError(w, http.StatusNotFound, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SaveAsRecipe converts an inline recipe to a saved recipe
func (h *Handler) SaveAsRecipe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract user_id from context
	userID, ok := httputil.GetUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get diary entry ID from URL
	id, err := extractID(r.URL.Path, "/diary/entries/")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	// Fetch diary entry
	entry, err := h.repo.GetByID(uint(id), userID)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "Diary entry not found")
		return
	}

	// Verify it's an inline recipe (not already a saved recipe)
	if entry.InlineRecipeName == nil || *entry.InlineRecipeName == "" {
		httputil.WriteError(w, http.StatusBadRequest, "This entry is not an inline recipe")
		return
	}

	if entry.RecipeID != nil {
		httputil.WriteError(w, http.StatusBadRequest, "This entry already references a saved recipe")
		return
	}

	// Verify custom ingredients exist
	if len(entry.CustomIngredients) == 0 {
		httputil.WriteError(w, http.StatusBadRequest, "Inline recipe has no ingredients")
		return
	}

	// Verify recipe repository is initialized
	if h.recipeRepo == nil {
		httputil.WriteError(w, http.StatusInternalServerError, "Recipe repository not initialized")
		return
	}

	// Create ingredients list for recipe
	ingredients := make([]RecipeIngredientRequest, len(entry.CustomIngredients))
	for i, ing := range entry.CustomIngredients {
		ingredients[i] = RecipeIngredientRequest{
			FoodID:        ing.FoodID,
			QuantityGrams: ing.QuantityGrams,
		}
	}

	// Create the recipe via recipe repository
	savedRecipe, err := h.recipeRepo.CreateRecipe(userID, *entry.InlineRecipeName, entry.RecipeTag, ingredients)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "Failed to save recipe: "+err.Error())
		return
	}

	// Update diary entry to reference the saved recipe
	entry.RecipeID = &savedRecipe.ID
	entry.InlineRecipeName = nil // Clear inline name since it's now a saved recipe

	if err := h.repo.Update(entry); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "Failed to update diary entry: "+err.Error())
		return
	}

	// Return the saved recipe
	httputil.WriteJSON(w, http.StatusOK, savedRecipe)
}

// SaveAsFood converts an inline food to a saved global food
func (h *Handler) SaveAsFood(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := httputil.GetUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := extractID(r.URL.Path, "/diary/entries/")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	// Fetch diary entry
	entry, err := h.repo.GetByID(uint(id), userID)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "Diary entry not found")
		return
	}

	// Verify it's an inline food
	if entry.InlineFoodName == nil || *entry.InlineFoodName == "" {
		httputil.WriteError(w, http.StatusBadRequest, "This entry is not an inline food")
		return
	}

	if entry.FoodID != nil {
		httputil.WriteError(w, http.StatusBadRequest, "This entry already references a saved food")
		return
	}

	// Verify nutrition data exists
	if entry.InlineFoodCalories == nil {
		httputil.WriteError(w, http.StatusBadRequest, "Inline food has no nutrition data")
		return
	}

	// Default tag to "routine" if not provided
	tag := "routine"
	if entry.InlineFoodTag != nil && *entry.InlineFoodTag != "" {
		tag = *entry.InlineFoodTag
	}

	// Create the food via food repository (GLOBAL)
	description := ""
	if entry.InlineFoodDescription != nil {
		description = *entry.InlineFoodDescription
	}

	savedFood, err := h.foodRepo.Create(food.CreateFoodRequest{
		Name:        *entry.InlineFoodName,
		Description: description,
		Calories:    *entry.InlineFoodCalories,
		Protein:     *entry.InlineFoodProtein,
		Carbs:       *entry.InlineFoodCarbs,
		Fat:         *entry.InlineFoodFat,
		Fiber:       *entry.InlineFoodFiber,
		Tag:         tag,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "Failed to save food: "+err.Error())
		return
	}

	// Update diary entry to reference saved food
	entry.FoodID = &savedFood.ID
	// Clear inline food fields
	entry.InlineFoodName = nil
	entry.InlineFoodDescription = nil
	entry.InlineFoodCalories = nil
	entry.InlineFoodProtein = nil
	entry.InlineFoodCarbs = nil
	entry.InlineFoodFat = nil
	entry.InlineFoodFiber = nil
	entry.InlineFoodTag = nil

	// Keep cached nutrition (historical accuracy)

	if err := h.repo.Update(entry); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "Failed to update diary entry: "+err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusOK, savedFood)
}

// CreateEntryFromOpenFoodFacts handles POST /diary/entries/from-openfoodfacts
func (h *Handler) CreateEntryFromOpenFoodFacts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := httputil.GetUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateEntryFromOpenFoodFactsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.ProductName == "" {
		httputil.WriteError(w, http.StatusBadRequest, "product_name is required")
		return
	}

	if req.QuantityGrams <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "quantity_grams must be greater than 0")
		return
	}

	// Validate nutrition values are non-negative
	if req.Calories < 0 || req.Protein < 0 || req.Carbs < 0 || req.Fat < 0 || req.Fiber < 0 {
		httputil.WriteError(w, http.StatusBadRequest, "Nutrition values must be non-negative")
		return
	}

	// Validate tag if provided
	tag := req.Tag
	if tag == "" {
		tag = "routine" // Default to routine
	} else if tag != "routine" && tag != "contextual" {
		httputil.WriteError(w, http.StatusBadRequest, "tag must be 'routine' or 'contextual'")
		return
	}

	// Parse date
	var entryDate time.Time
	if req.Date == "" {
		entryDate = time.Now()
	} else {
		var err error
		entryDate, err = time.Parse("2006-01-02", req.Date)
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "Invalid date format (use YYYY-MM-DD)")
			return
		}
	}

	// Build description from product name and brands
	description := req.ProductName
	if req.Brands != "" {
		description = req.Brands + " - " + req.ProductName
	}

	// Calculate consumed nutrition (product data is per 100g)
	multiplier := req.QuantityGrams / 100.0

	// Create diary entry with inline food
	entry := &DiaryEntry{
		UserID:                userID,
		InlineFoodName:        &req.ProductName,
		InlineFoodDescription: &description,
		InlineFoodCalories:    &req.Calories,
		InlineFoodProtein:     &req.Protein,
		InlineFoodCarbs:       &req.Carbs,
		InlineFoodFat:         &req.Fat,
		InlineFoodFiber:       &req.Fiber,
		InlineFoodTag:         &tag,
		Date:                  entryDate,
		MealType:              req.MealType,
		QuantityGrams:         req.QuantityGrams,
		Notes:                 req.Notes,
		// Cached consumed nutrition
		Calories:   roundToTwo(req.Calories * multiplier),
		Protein:    roundToTwo(req.Protein * multiplier),
		Carbs:      roundToTwo(req.Carbs * multiplier),
		Fat:        roundToTwo(req.Fat * multiplier),
		Fiber:      roundToTwo(req.Fiber * multiplier),
		FoodTag:    tag,
	}

	if err := h.repo.Create(entry); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, entry)
}

// calculateAdherence calculates the adherence percentage
func calculateAdherence(actual, goal float64) float64 {
	if goal == 0 {
		return 0
	}
	return (actual / goal) * 100
}

// calculateCaloriesByTag aggregates calories by tag type from diary entries
func calculateCaloriesByTag(entries []DiaryEntry) (routineCalories, contextualCalories, routinePercent, contextualPercent float64) {
	var totalCalories float64

	for _, entry := range entries {
		// Determine which tag to use (food_tag takes precedence over recipe_tag)
		tag := entry.FoodTag
		if tag == "" {
			tag = entry.RecipeTag
		}

		// Accumulate calories by tag
		if tag == "routine" {
			routineCalories += entry.Calories
		} else if tag == "contextual" {
			contextualCalories += entry.Calories
		}
		totalCalories += entry.Calories
	}

	// Calculate percentages (avoid division by zero)
	if totalCalories > 0 {
		routinePercent = (routineCalories / totalCalories) * 100
		contextualPercent = (contextualCalories / totalCalories) * 100
	}

	return
}

// GetWeeklySummary handles GET /diary/weekly?start_date=YYYY-MM-DD
func (h *Handler) GetWeeklySummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := httputil.GetUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse start_date query param or default to current week's Monday
	startDateStr := r.URL.Query().Get("start_date")
	var startDate time.Time
	if startDateStr == "" {
		// Default to current week's Monday
		now := time.Now()
		weekday := int(now.Weekday())
		if weekday == 0 { // Sunday
			weekday = 7
		}
		daysFromMonday := weekday - 1
		startDate = now.AddDate(0, 0, -daysFromMonday)
	} else {
		var err error
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "Invalid date format (use YYYY-MM-DD)")
			return
		}
	}

	// Calculate end date (6 days after start)
	endDate := startDate.AddDate(0, 0, 6)

	// Get active goal once for all days
	activeGoal, err := h.goalRepo.GetActive(userID)
	var goalCalories, goalProtein, goalCarbs, goalFat, goalFiber float64
	if err == nil {
		goalCalories = activeGoal.Calories
		goalProtein = activeGoal.Protein
		goalCarbs = activeGoal.Carbs
		goalFat = activeGoal.Fat
		goalFiber = activeGoal.Fiber
	}

	// Build daily summaries for each day of the week
	var dailySummaries []DailySummary
	var totalCalories, totalProtein, totalCarbs, totalFat, totalFiber float64
	var totalRoutineCalories, totalContextualCalories float64

	for i := 0; i < 7; i++ {
		currentDate := startDate.AddDate(0, 0, i)
		dateStr := currentDate.Format("2006-01-02")

		// Get entries for this day
		entries, err := h.repo.GetByDate(userID, currentDate)
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Calculate totals for this day
		summary, err := h.repo.GetDailySummary(userID, currentDate)
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Calculate adherence
		adherence := AdherencePercent{
			Calories: calculateAdherence(summary["calories"], goalCalories),
			Protein:  calculateAdherence(summary["protein"], goalProtein),
			Carbs:    calculateAdherence(summary["carbs"], goalCarbs),
			Fat:      calculateAdherence(summary["fat"], goalFat),
			Fiber:    calculateAdherence(summary["fiber"], goalFiber),
		}

		// Calculate calorie breakdown by tag for this day
		routineCalories, contextualCalories, routinePercent, contextualPercent := calculateCaloriesByTag(entries)

		// Build daily summary
		dailySummary := DailySummary{
			Date:               dateStr,
			TotalCalories:      summary["calories"],
			TotalProtein:       summary["protein"],
			TotalCarbs:         summary["carbs"],
			TotalFat:           summary["fat"],
			TotalFiber:         summary["fiber"],
			GoalCalories:       goalCalories,
			GoalProtein:        goalProtein,
			GoalCarbs:          goalCarbs,
			GoalFat:            goalFat,
			GoalFiber:          goalFiber,
			Adherence:          adherence,
			RoutineCalories:    routineCalories,
			ContextualCalories: contextualCalories,
			RoutinePercent:     routinePercent,
			ContextualPercent:  contextualPercent,
			Entries:            entries,
		}

		dailySummaries = append(dailySummaries, dailySummary)

		// Accumulate for weekly averages
		totalCalories += summary["calories"]
		totalProtein += summary["protein"]
		totalCarbs += summary["carbs"]
		totalFat += summary["fat"]
		totalFiber += summary["fiber"]
		totalRoutineCalories += routineCalories
		totalContextualCalories += contextualCalories
	}

	// Calculate weekly average percentages
	var avgRoutinePercent, avgContextualPercent float64
	avgWeeklyCalories := totalCalories / 7

	if avgWeeklyCalories > 0 {
		avgRoutinePercent = roundToTwo((totalRoutineCalories / 7 / avgWeeklyCalories) * 100)
		avgContextualPercent = roundToTwo((totalContextualCalories / 7 / avgWeeklyCalories) * 100)
	}

	// Calculate weekly averages
	weeklySummary := WeeklySummary{
		StartDate:            startDate.Format("2006-01-02"),
		EndDate:              endDate.Format("2006-01-02"),
		DailySummaries:       dailySummaries,
		AverageCalories:      roundToTwo(totalCalories / 7),
		AverageProtein:       roundToTwo(totalProtein / 7),
		AverageCarbs:         roundToTwo(totalCarbs / 7),
		AverageFat:           roundToTwo(totalFat / 7),
		AverageFiber:         roundToTwo(totalFiber / 7),
		AvgRoutinePercent:    avgRoutinePercent,
		AvgContextualPercent: avgContextualPercent,
	}

	httputil.WriteJSON(w, http.StatusOK, weeklySummary)
}

// roundToTwo rounds a float to 2 decimal places
func roundToTwo(val float64) float64 {
	return float64(int(val*100)) / 100
}

// extractID extracts the ID from the URL path
func extractID(path, prefix string) (int, error) {
	idStr := strings.TrimPrefix(path, prefix)
	return strconv.Atoi(idStr)
}

// validateCustomIngredients validates that custom ingredients belong to the recipe
func (h *Handler) validateCustomIngredients(recipeID int, customIngredients []CustomIngredientRequest) error {
	// Get recipe ingredients
	recipeIngredients, err := h.recipeRepo.GetIngredients(recipeID)
	if err != nil {
		return err
	}

	// Build a map of valid food IDs in the recipe
	validFoodIDs := make(map[uint]bool)
	for _, ingredient := range recipeIngredients {
		validFoodIDs[ingredient.FoodID] = true
	}

	// Check if all custom ingredients are in the recipe
	for _, customIng := range customIngredients {
		if !validFoodIDs[customIng.FoodID] {
			return errors.New("custom ingredient food_id does not belong to this recipe")
		}
		if customIng.QuantityGrams <= 0 {
			return errors.New("custom ingredient quantity must be greater than 0")
		}
	}

	// Check if all recipe ingredients are provided (strict mode)
	if len(customIngredients) != len(recipeIngredients) {
		return errors.New("all recipe ingredients must be provided with custom quantities")
	}

	return nil
}

// calculateCustomIngredientsNutrition calculates nutrition for custom ingredients
func (h *Handler) calculateCustomIngredientsNutrition(customIngredients []CustomIngredientRequest) (CustomIngredients, float64, float64, float64, float64, float64, float64, error) {
	var result CustomIngredients
	var totalCalories, totalProtein, totalCarbs, totalFat, totalFiber, totalWeight float64

	for _, customIng := range customIngredients {
		// Fetch food item
		foodItem, err := h.foodRepo.GetByID(int(customIng.FoodID))
		if err != nil {
			return nil, 0, 0, 0, 0, 0, 0, err
		}

		// Calculate nutrition (food nutrition is per 100g)
		multiplier := customIng.QuantityGrams / 100.0
		ingredientCalories := foodItem.Calories * multiplier
		ingredientProtein := foodItem.Protein * multiplier
		ingredientCarbs := foodItem.Carbs * multiplier
		ingredientFat := foodItem.Fat * multiplier
		ingredientFiber := foodItem.Fiber * multiplier

		// Add to result
		result = append(result, CustomIngredient{
			FoodID:        customIng.FoodID,
			FoodName:      foodItem.Name,
			QuantityGrams: customIng.QuantityGrams,
			Calories:      roundToTwo(ingredientCalories),
			Protein:       roundToTwo(ingredientProtein),
			Carbs:         roundToTwo(ingredientCarbs),
			Fat:           roundToTwo(ingredientFat),
			Fiber:         roundToTwo(ingredientFiber),
		})

		// Accumulate totals
		totalCalories += ingredientCalories
		totalProtein += ingredientProtein
		totalCarbs += ingredientCarbs
		totalFat += ingredientFat
		totalFiber += ingredientFiber
		totalWeight += customIng.QuantityGrams
	}

	return result, totalCalories, totalProtein, totalCarbs, totalFat, totalFiber, totalWeight, nil
}

// convertProportionalToCustomIngredients converts proportional quantity to custom ingredients
func (h *Handler) convertProportionalToCustomIngredients(recipeID int, quantityGrams float64) (CustomIngredients, float64, float64, float64, float64, float64, error) {
	// Get recipe ingredients
	recipeIngredients, err := h.recipeRepo.GetIngredients(recipeID)
	if err != nil {
		return nil, 0, 0, 0, 0, 0, err
	}

	// Calculate total recipe weight
	var totalRecipeWeight float64
	for _, ingredient := range recipeIngredients {
		totalRecipeWeight += ingredient.QuantityGrams
	}

	if totalRecipeWeight == 0 {
		return nil, 0, 0, 0, 0, 0, errors.New("recipe has no ingredients")
	}

	// Calculate portion
	portion := quantityGrams / totalRecipeWeight

	// Build custom ingredients with proportional quantities
	var customIngredients []CustomIngredientRequest
	for _, ingredient := range recipeIngredients {
		customIngredients = append(customIngredients, CustomIngredientRequest{
			FoodID:        ingredient.FoodID,
			QuantityGrams: roundToTwo(ingredient.QuantityGrams * portion),
		})
	}

	// Calculate nutrition
	result, totalCalories, totalProtein, totalCarbs, totalFat, totalFiber, _, err := h.calculateCustomIngredientsNutrition(customIngredients)
	return result, totalCalories, totalProtein, totalCarbs, totalFat, totalFiber, err
}

// determineInlineRecipeTag determines the tag for an inline recipe based on ingredients
// Logic: contextual if ANY ingredient is contextual, routine if ALL are routine
func (h *Handler) determineInlineRecipeTag(ingredients CustomIngredients) string {
	hasContextual := false
	allRoutine := true

	for _, ing := range ingredients {
		food, err := h.foodRepo.GetByID(int(ing.FoodID))
		if err != nil {
			continue
		}

		if food.Tag == "contextual" {
			hasContextual = true
		}
		if food.Tag != "routine" {
			allRoutine = false
		}
	}

	if hasContextual {
		return "contextual"
	}
	if allRoutine {
		return "routine"
	}
	return ""
}

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

	// Validation
	if req.FoodID == nil && req.RecipeID == nil {
		httputil.WriteError(w, http.StatusBadRequest, "Either food_id or recipe_id is required")
		return
	}

	// For food entries, quantity_grams is required
	// For recipe entries, either quantity_grams or custom_ingredients is required
	if req.FoodID != nil && req.QuantityGrams <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "Quantity in grams must be greater than 0 for food entries")
		return
	}

	if req.RecipeID != nil && req.QuantityGrams <= 0 && len(req.CustomIngredients) == 0 {
		httputil.WriteError(w, http.StatusBadRequest, "Either quantity_grams or custom_ingredients is required for recipe entries")
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

// calculateAdherence calculates the adherence percentage
func calculateAdherence(actual, goal float64) float64 {
	if goal == 0 {
		return 0
	}
	return (actual / goal) * 100
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

		// Build daily summary
		dailySummary := DailySummary{
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
			Entries:       entries,
		}

		dailySummaries = append(dailySummaries, dailySummary)

		// Accumulate for weekly averages
		totalCalories += summary["calories"]
		totalProtein += summary["protein"]
		totalCarbs += summary["carbs"]
		totalFat += summary["fat"]
		totalFiber += summary["fiber"]
	}

	// Calculate weekly averages
	weeklySummary := WeeklySummary{
		StartDate:       startDate.Format("2006-01-02"),
		EndDate:         endDate.Format("2006-01-02"),
		DailySummaries:  dailySummaries,
		AverageCalories: roundToTwo(totalCalories / 7),
		AverageProtein:  roundToTwo(totalProtein / 7),
		AverageCarbs:    roundToTwo(totalCarbs / 7),
		AverageFat:      roundToTwo(totalFat / 7),
		AverageFiber:    roundToTwo(totalFiber / 7),
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

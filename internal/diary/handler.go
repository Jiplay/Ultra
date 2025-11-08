package diary

import (
	"encoding/json"
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
	ID          uint
	Name        string
	ServingSize float64
}

// RecipeIngredient represents an ingredient in a recipe
type RecipeIngredient struct {
	FoodID   uint
	Quantity float64
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

// CreateEntry handles POST /diary/entries
func (h *Handler) CreateEntry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateDiaryEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validation
	if req.FoodID == nil && req.RecipeID == nil {
		writeError(w, http.StatusBadRequest, "Either food_id or recipe_id is required")
		return
	}

	if req.ServingSize <= 0 {
		req.ServingSize = 1
	}

	// Parse date
	var entryDate time.Time
	if req.Date == "" {
		entryDate = time.Now()
	} else {
		var err error
		entryDate, err = time.Parse("2006-01-02", req.Date)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid date format (use YYYY-MM-DD)")
			return
		}
	}

	// Create entry
	entry := &DiaryEntry{
		UserID:      userID,
		FoodID:      req.FoodID,
		RecipeID:    req.RecipeID,
		Date:        entryDate,
		MealType:    req.MealType,
		ServingSize: req.ServingSize,
		Notes:       req.Notes,
	}

	// Calculate nutrition if food_id is provided
	if req.FoodID != nil {
		foodItem, err := h.foodRepo.GetByID(int(*req.FoodID))
		if err != nil {
			writeError(w, http.StatusBadRequest, "Food not found")
			return
		}

		entry.Calories = foodItem.Calories * req.ServingSize
		entry.Protein = foodItem.Protein * req.ServingSize
		entry.Carbs = foodItem.Carbs * req.ServingSize
		entry.Fat = foodItem.Fat * req.ServingSize
		entry.Fiber = foodItem.Fiber * req.ServingSize
	}

	// Calculate nutrition if recipe_id is provided
	if req.RecipeID != nil {
		if h.recipeRepo == nil {
			writeError(w, http.StatusInternalServerError, "Recipe repository not initialized")
			return
		}

		_, err := h.recipeRepo.GetByID(int(*req.RecipeID))
		if err != nil {
			writeError(w, http.StatusBadRequest, "Recipe not found")
			return
		}

		ingredients, err := h.recipeRepo.GetIngredients(int(*req.RecipeID))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to load recipe ingredients")
			return
		}

		// Calculate total nutrition from all ingredients
		var totalCalories, totalProtein, totalCarbs, totalFat, totalFiber float64

		for _, ingredient := range ingredients {
			foodItem, err := h.foodRepo.GetByID(int(ingredient.FoodID))
			if err != nil {
				continue // Skip if food not found
			}

			totalCalories += foodItem.Calories * ingredient.Quantity
			totalProtein += foodItem.Protein * ingredient.Quantity
			totalCarbs += foodItem.Carbs * ingredient.Quantity
			totalFat += foodItem.Fat * ingredient.Quantity
			totalFiber += foodItem.Fiber * ingredient.Quantity
		}

		// Apply serving size multiplier
		entry.Calories = totalCalories * req.ServingSize
		entry.Protein = totalProtein * req.ServingSize
		entry.Carbs = totalCarbs * req.ServingSize
		entry.Fat = totalFat * req.ServingSize
		entry.Fiber = totalFiber * req.ServingSize
	}

	if err := h.repo.Create(entry); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, entry)
}

// GetEntries handles GET /diary/entries?date=YYYY-MM-DD
func (h *Handler) GetEntries(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
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
			writeError(w, http.StatusBadRequest, "Invalid date format (use YYYY-MM-DD)")
			return
		}
	}

	entries, err := h.repo.GetByDate(userID, entryDate)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, entries)
}

// GetDailySummary handles GET /diary/summary/{date}
func (h *Handler) GetDailySummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	dateStr := strings.TrimPrefix(r.URL.Path, "/diary/summary/")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	entryDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid date format (use YYYY-MM-DD)")
		return
	}

	// Get entries for the day
	entries, err := h.repo.GetByDate(userID, entryDate)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Calculate totals
	summary, err := h.repo.GetDailySummary(userID, entryDate)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
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

	writeJSON(w, http.StatusOK, response)
}

// UpdateEntry handles PUT /diary/entries/{id}
func (h *Handler) UpdateEntry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := extractID(r.URL.Path, "/diary/entries/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var req UpdateDiaryEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	entry, err := h.repo.GetByID(uint(id), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Entry not found")
		return
	}

	// Update fields
	if req.ServingSize > 0 {
		// Recalculate nutrition based on new serving size
		if entry.FoodID != nil {
			foodItem, err := h.foodRepo.GetByID(int(*entry.FoodID))
			if err == nil {
				entry.ServingSize = req.ServingSize
				entry.Calories = foodItem.Calories * req.ServingSize
				entry.Protein = foodItem.Protein * req.ServingSize
				entry.Carbs = foodItem.Carbs * req.ServingSize
				entry.Fat = foodItem.Fat * req.ServingSize
				entry.Fiber = foodItem.Fiber * req.ServingSize
			}
		}
	}

	if req.MealType != "" {
		entry.MealType = req.MealType
	}
	if req.Notes != "" {
		entry.Notes = req.Notes
	}

	if err := h.repo.Update(entry); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, entry)
}

// DeleteEntry handles DELETE /diary/entries/{id}
func (h *Handler) DeleteEntry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := extractID(r.URL.Path, "/diary/entries/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	if err := h.repo.Delete(uint(id), userID); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
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
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
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
			writeError(w, http.StatusBadRequest, "Invalid date format (use YYYY-MM-DD)")
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
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Calculate totals for this day
		summary, err := h.repo.GetDailySummary(userID, currentDate)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
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

	writeJSON(w, http.StatusOK, weeklySummary)
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

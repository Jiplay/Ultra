package goal

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ultra-bis/internal/user"
)

// Handler handles nutrition goal requests
type Handler struct {
	repo     *Repository
	userRepo *user.Repository
}

// NewHandler creates a new goal handler
func NewHandler(repo *Repository, userRepo *user.Repository) *Handler {
	return &Handler{repo: repo, userRepo: userRepo}
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

// CreateGoal handles POST /goals
func (h *Handler) CreateGoal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Set start date to today if not provided
	startDate := req.StartDate.Time
	if startDate.IsZero() {
		startDate = time.Now()
	}

	var endDate *time.Time
	if req.EndDate != nil && !req.EndDate.IsZero() {
		endDate = &req.EndDate.Time
	}

	goal := &NutritionGoal{
		UserID:    userID,
		Calories:  req.Calories,
		Protein:   req.Protein,
		Carbs:     req.Carbs,
		Fat:       req.Fat,
		Fiber:     req.Fiber,
		StartDate: startDate,
		EndDate:   endDate,
		IsActive:  true,
	}

	if err := h.repo.Create(goal); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, goal)
}

// GetActiveGoal handles GET /goals
func (h *Handler) GetActiveGoal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	goal, err := h.repo.GetActive(userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "No active goal found")
		return
	}

	writeJSON(w, http.StatusOK, goal)
}

// GetAllGoals handles GET /goals/all
func (h *Handler) GetAllGoals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	goals, err := h.repo.GetAll(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, goals)
}

// UpdateGoal handles PUT /goals/{id}
func (h *Handler) UpdateGoal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := extractID(r.URL.Path, "/goals/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var req UpdateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	goal, err := h.repo.GetByID(uint(id), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Goal not found")
		return
	}

	// Update fields
	if req.Calories > 0 {
		goal.Calories = req.Calories
	}
	if req.Protein > 0 {
		goal.Protein = req.Protein
	}
	if req.Carbs > 0 {
		goal.Carbs = req.Carbs
	}
	if req.Fat > 0 {
		goal.Fat = req.Fat
	}
	if req.Fiber > 0 {
		goal.Fiber = req.Fiber
	}
	if req.EndDate != nil {
		goal.EndDate = req.EndDate
	}

	if err := h.repo.Update(goal); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, goal)
}

// DeleteGoal handles DELETE /goals/{id}
func (h *Handler) DeleteGoal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := extractID(r.URL.Path, "/goals/")
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

// GetRecommendedGoals calculates recommended nutrition goals
func (h *Handler) GetRecommendedGoals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get user data
	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "User not found")
		return
	}

	var req RecommendedGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Calculate TDEE (Total Daily Energy Expenditure)
	bmr := calculateBMR(req.Weight, user.Height, float64(user.Age), user.Gender)
	tdee := bmr * getActivityMultiplier(user.ActivityLevel)

	// Adjust for goal
	var calories float64
	var message string

	if req.TargetWeight < req.Weight {
		// Weight loss: 500 cal deficit per day = 0.5 kg per week
		deficit := (req.Weight - req.TargetWeight) * 7700 / float64(req.WeeksToGoal) / 7
		calories = tdee - math.Min(deficit, 1000) // Max 1000 cal deficit
		message = fmt.Sprintf("Goal: Lose %.1f kg in %d weeks", req.Weight-req.TargetWeight, req.WeeksToGoal)
	} else if req.TargetWeight > req.Weight {
		// Weight gain: 500 cal surplus per day = 0.5 kg per week
		surplus := (req.TargetWeight - req.Weight) * 7700 / float64(req.WeeksToGoal) / 7
		calories = tdee + math.Min(surplus, 500) // Max 500 cal surplus
		message = fmt.Sprintf("Goal: Gain %.1f kg in %d weeks", req.TargetWeight-req.Weight, req.WeeksToGoal)
	} else {
		calories = tdee
		message = "Goal: Maintain current weight"
	}

	// Calculate macros (40% carbs, 30% protein, 30% fat for athletes)
	protein := (calories * 0.30) / 4  // 4 calories per gram of protein
	carbs := (calories * 0.40) / 4    // 4 calories per gram of carbs
	fat := (calories * 0.30) / 9      // 9 calories per gram of fat
	fiber := 14 * (calories / 1000)   // 14g per 1000 calories

	response := RecommendedGoalResponse{
		Calories: math.Round(calories),
		Protein:  math.Round(protein),
		Carbs:    math.Round(carbs),
		Fat:      math.Round(fat),
		Fiber:    math.Round(fiber),
		Message:  message,
	}

	writeJSON(w, http.StatusOK, response)
}

// calculateBMR calculates Basal Metabolic Rate using Mifflin-St Jeor Equation
func calculateBMR(weight, height, age float64, gender string) float64 {
	if gender == "male" {
		return (10 * weight) + (6.25 * height) - (5 * age) + 5
	}
	return (10 * weight) + (6.25 * height) - (5 * age) - 161
}

// getActivityMultiplier returns the activity level multiplier for TDEE
func getActivityMultiplier(level user.ActivityLevel) float64 {
	switch level {
	case user.Sedentary:
		return 1.2
	case user.LightlyActive:
		return 1.375
	case user.ModeratelyActive:
		return 1.55
	case user.VeryActive:
		return 1.725
	case user.ExtraActive:
		return 1.9
	default:
		return 1.55 // Default to moderate
	}
}

// extractID extracts the ID from the URL path
func extractID(path, prefix string) (int, error) {
	idStr := strings.TrimPrefix(path, prefix)
	return strconv.Atoi(idStr)
}

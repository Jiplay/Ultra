package goal

import (
	"ultra-bis/internal/httputil"
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


// CreateGoal handles POST /goals
func (h *Handler) CreateGoal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := httputil.GetUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid request body")
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

	// Add protocol tracking if provided
	if req.DietModel != nil && req.Protocol != nil {
		now := time.Now()
		expirationDate := now.AddDate(0, 0, 14) // 2 weeks from creation

		goal.DietModel = req.DietModel
		goal.Protocol = req.Protocol
		goal.Phase = req.Phase // May be nil if not provided
		goal.ExpirationDate = &expirationDate
	}

	if err := h.repo.Create(goal); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, goal)
}

// GetActiveGoal handles GET /goals
func (h *Handler) GetActiveGoal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := httputil.GetUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	goal, err := h.repo.GetActive(userID)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "No active goal found")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, goal)
}

// GetAllGoals handles GET /goals/all
func (h *Handler) GetAllGoals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := httputil.GetUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	goals, err := h.repo.GetAll(userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusOK, goals)
}

// UpdateGoal handles PUT /goals/{id}
func (h *Handler) UpdateGoal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := httputil.GetUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := extractID(r.URL.Path, "/goals/")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var req UpdateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	goal, err := h.repo.GetByID(uint(id), userID)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "Goal not found")
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
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusOK, goal)
}

// DeleteGoal handles DELETE /goals/{id}
func (h *Handler) DeleteGoal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := httputil.GetUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := extractID(r.URL.Path, "/goals/")
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

// GetRecommendedGoals calculates recommended nutrition goals
func (h *Handler) GetRecommendedGoals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := httputil.GetUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get user data
	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "User not found")
		return
	}

	var req RecommendedGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid request body")
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

	httputil.WriteJSON(w, http.StatusOK, response)
}

// CalculateDietGoals handles POST /goals/calculate
func (h *Handler) CalculateDietGoals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := httputil.GetUserID(r)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CalculateDietRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get user data
	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "User not found")
		return
	}

	// Validate request using the factory pattern
	if err := ValidateDietRequest(&req, user); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get the appropriate calculator
	calculator, err := GetDietCalculator(req.DietModel)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Perform calculations
	result, err := calculator.Calculate(user, req.Protocol)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Calculation failed: %v", err))
		return
	}

	// Convert to response format
	response := h.buildDietResponse(result)

	httputil.WriteJSON(w, http.StatusOK, response)
}

// buildDietResponse converts DietCalculationResult to CalculateDietResponse
func (h *Handler) buildDietResponse(result *DietCalculationResult) CalculateDietResponse {
	// Convert phases to response format with rounding
	phases := make([]DietPhaseResponse, len(result.Phases))
	for i, phase := range result.Phases {
		phases[i] = DietPhaseResponse{
			Phase:       phase.Phase,
			Calories:    math.Round(phase.Calories),
			Protein:     math.Round(phase.Protein),
			Carbs:       math.Round(phase.Carbs),
			Fat:         math.Round(phase.Fat),
			Description: phase.Description,
		}
	}

	return CalculateDietResponse{
		DietModel:      result.ModelName,
		Protocol:       result.Protocol,
		ProtocolName:   result.ProtocolName,
		BMR:            math.Round(result.BMR),
		MaintenanceMMR: math.Round(result.MaintenanceMMR),
		LeanMass:       math.Round(result.LeanMass*10) / 10, // Round to 1 decimal
		Phases:         phases,
		Message:        fmt.Sprintf("Calculated using %s - %s", result.ModelName, result.ProtocolName),
	}
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

// GetAvailableDiets handles GET /goals/diets
func (h *Handler) GetAvailableDiets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Build the response with all available diet models and protocols
	response := AvailableDietsResponse{
		Diets: []AvailableDiet{
			{
				ModelName:   "zeroToHero",
				DisplayName: "Zero to Hero",
				Description: "Comprehensive diet program with 4 protocols for different goals: muscle building, recomposition, and fat loss",
				Protocols: []ProtocolInfo{
					{
						Number:      1,
						Name:        "Protocole 1 : Prise de muscle propre",
						Description: "Muscle building protocol with progressive caloric surplus (3 phases: maintenance, moderate surplus, high surplus)",
					},
					{
						Number:      2,
						Name:        "Protocole 2 : Recomposition corporelle",
						Description: "Body recomposition protocol for simultaneous fat loss and muscle gain (single phase with slight deficit)",
					},
					{
						Number:      3,
						Name:        "Protocole 3 : Créer le déficit parfait",
						Description: "Perfect deficit protocol for controlled fat loss (2 phases: moderate and higher deficit)",
					},
					{
						Number:      4,
						Name:        "Protocole 4 : Perte de gras progressive",
						Description: "Progressive fat loss protocol with gradual caloric reduction (3 phases: initial, moderate, and aggressive deficit)",
					},
				},
			},
		},
	}

	httputil.WriteJSON(w, http.StatusOK, response)
}

// extractID extracts the ID from the URL path
func extractID(path, prefix string) (int, error) {
	idStr := strings.TrimPrefix(path, prefix)
	return strconv.Atoi(idStr)
}

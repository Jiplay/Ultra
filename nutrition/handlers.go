package nutrition

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Handlers struct {
	controller *Controller
}

func NewHandlers(controller *Controller) *Handlers {
	return &Handlers{controller: controller}
}

func (h *Handlers) CreateFood(w http.ResponseWriter, r *http.Request) {
	var req CreateFoodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")

	food, err := h.controller.CreateFood(&req, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(food)
}

func (h *Handlers) GetFoods(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")

	foods, err := h.controller.GetFoods(search)
	if err != nil {
		http.Error(w, "Failed to get foods", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(foods)
}

func (h *Handlers) GetFood(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	foodID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid food ID", http.StatusBadRequest)
		return
	}

	food, err := h.controller.GetFoodByID(foodID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Food not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get food", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(food)
}

func (h *Handlers) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	var req CreateRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")

	recipe, err := h.controller.CreateRecipe(&req, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(recipe)
}

func (h *Handlers) GetRecipes(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")

	recipes, err := h.controller.GetRecipesByUserID(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recipes)
}

func (h *Handlers) UpdateNutritionGoals(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	var req UpdateNutritionGoalsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	goals, err := h.controller.UpdateNutritionGoals(userID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(goals)
}

func (h *Handlers) GetNutritionGoals(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	goals, err := h.controller.GetNutritionGoalsByUserID(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Nutrition goals not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(goals)
}

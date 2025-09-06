package meal

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Handlers struct {
	controller *Controller
}

func NewHandlers(controller *Controller) *Handlers {
	return &Handlers{controller: controller}
}

func (h *Handlers) CreateMeal(w http.ResponseWriter, r *http.Request) {
	var req CreateMealRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")

	meal, err := h.controller.CreateMeal(&req, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(meal)
}

func (h *Handlers) GetMeals(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = r.Header.Get("X-User-ID")
	}

	// Parse optional date filters
	var startDate, endDate *time.Time
	
	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &parsed
		} else {
			http.Error(w, "Invalid start_date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &parsed
		} else {
			http.Error(w, "Invalid end_date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}

	meals, err := h.controller.GetMealsByUserID(userID, startDate, endDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(meals)
}

func (h *Handlers) GetMeal(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/meals/")
	mealID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid meal ID", http.StatusBadRequest)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = r.Header.Get("X-User-ID")
	}

	meal, err := h.controller.GetMealByID(mealID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Meal not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(meal)
}

func (h *Handlers) UpdateMeal(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/meals/")
	mealID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid meal ID", http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")

	var req UpdateMealRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err = h.controller.UpdateMeal(mealID, userID, &req)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Meal not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) DeleteMeal(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/meals/")
	mealID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid meal ID", http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")

	err = h.controller.DeleteMeal(mealID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Meal not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete meal", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) AddMealItem(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/meals/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "items" {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}

	mealID, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "Invalid meal ID", http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")

	var req AddMealItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	item, err := h.controller.AddMealItem(mealID, userID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

func (h *Handlers) UpdateMealItem(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/meals/")
	parts := strings.Split(path, "/")
	if len(parts) < 3 || parts[1] != "items" {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}

	mealID, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "Invalid meal ID", http.StatusBadRequest)
		return
	}

	itemID, err := strconv.Atoi(parts[2])
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")

	var req UpdateMealItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err = h.controller.UpdateMealItem(itemID, mealID, userID, &req)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Meal item not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) DeleteMealItem(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/meals/")
	parts := strings.Split(path, "/")
	if len(parts) < 3 || parts[1] != "items" {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}

	mealID, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "Invalid meal ID", http.StatusBadRequest)
		return
	}

	itemID, err := strconv.Atoi(parts[2])
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")

	err = h.controller.DeleteMealItem(itemID, mealID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Meal item not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete meal item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) GetMealSummary(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = r.Header.Get("X-User-ID")
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		http.Error(w, "Date parameter is required (format: YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	summary, err := h.controller.GetMealSummary(userID, date)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func (h *Handlers) GetMealPlan(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = r.Header.Get("X-User-ID")
	}

	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	if startDateStr == "" || endDateStr == "" {
		http.Error(w, "start_date and end_date parameters are required (format: YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		http.Error(w, "Invalid start_date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		http.Error(w, "Invalid end_date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	plan, err := h.controller.GetMealPlan(userID, startDate, endDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plan)
}

func (h *Handlers) GetDailyMeals(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = r.Header.Get("X-User-ID")
	}

	dateStr := r.URL.Query().Get("date")
	var date time.Time
	var err error

	if dateStr == "" {
		// Default to today
		date = time.Now()
	} else {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}

	meals, err := h.controller.GetDailyMeals(userID, date)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(meals)
}
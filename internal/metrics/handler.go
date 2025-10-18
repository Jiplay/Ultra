package metrics

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ultra-bis/internal/user"
)

// Handler handles body metrics requests
type Handler struct {
	repo     *Repository
	userRepo *user.Repository
}

// NewHandler creates a new metrics handler
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

// CreateMetric handles POST /metrics
func (h *Handler) CreateMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateMetricRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Parse date
	var metricDate time.Time
	if req.Date == "" {
		metricDate = time.Now()
	} else {
		var err error
		metricDate, err = time.Parse("2006-01-02", req.Date)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid date format (use YYYY-MM-DD)")
			return
		}
	}

	// Get user for height (needed for BMI calculation)
	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	// Calculate BMI
	bmi := 0.0
	if req.Weight > 0 && user.Height > 0 {
		heightInMeters := user.Height / 100
		bmi = req.Weight / (heightInMeters * heightInMeters)
	}

	metric := &BodyMetric{
		UserID:            userID,
		Date:              metricDate,
		Weight:            req.Weight,
		BodyFatPercent:    req.BodyFatPercent,
		MuscleMassPercent: req.MuscleMassPercent,
		BMI:               bmi,
		Notes:             req.Notes,
	}

	if err := h.repo.Create(metric); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, metric)
}

// GetMetrics handles GET /metrics
func (h *Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	metrics, err := h.repo.GetAll(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, metrics)
}

// GetLatest handles GET /metrics/latest
func (h *Handler) GetLatest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	metric, err := h.repo.GetLatest(userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "No metrics found")
		return
	}

	writeJSON(w, http.StatusOK, metric)
}

// GetTrends handles GET /metrics/trends?period=7d|30d|90d
func (h *Handler) GetTrends(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "30d"
	}

	// Calculate date range
	var days int
	switch period {
	case "7d":
		days = 7
	case "30d":
		days = 30
	case "90d":
		days = 90
	default:
		writeError(w, http.StatusBadRequest, "Invalid period (use 7d, 30d, or 90d)")
		return
	}

	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	metrics, err := h.repo.GetByDateRange(userID, startDate, endDate)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(metrics) == 0 {
		writeError(w, http.StatusNotFound, "No metrics found for period")
		return
	}

	// Calculate trends
	trend := calculateTrend(metrics)

	response := TrendResponse{
		Period:  period,
		Metrics: metrics,
		Trend:   trend,
	}

	writeJSON(w, http.StatusOK, response)
}

// DeleteMetric handles DELETE /metrics/{id}
func (h *Handler) DeleteMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := extractID(r.URL.Path, "/metrics/")
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

// calculateTrend calculates trend data from metrics
func calculateTrend(metrics []BodyMetric) TrendData {
	if len(metrics) == 0 {
		return TrendData{}
	}

	var totalWeight, totalBodyFat, totalMuscleMass float64
	count := 0

	for _, m := range metrics {
		if m.Weight > 0 {
			totalWeight += m.Weight
			count++
		}
		if m.BodyFatPercent > 0 {
			totalBodyFat += m.BodyFatPercent
		}
		if m.MuscleMassPercent > 0 {
			totalMuscleMass += m.MuscleMassPercent
		}
	}

	avgWeight := 0.0
	if count > 0 {
		avgWeight = totalWeight / float64(count)
	}

	avgBodyFat := 0.0
	if count > 0 {
		avgBodyFat = totalBodyFat / float64(count)
	}

	avgMuscleMass := 0.0
	if count > 0 {
		avgMuscleMass = totalMuscleMass / float64(count)
	}

	// Calculate change (first to last)
	first := metrics[0]
	last := metrics[len(metrics)-1]

	return TrendData{
		WeightChange:      roundToTwo(last.Weight - first.Weight),
		BodyFatChange:     roundToTwo(last.BodyFatPercent - first.BodyFatPercent),
		MuscleMassChange:  roundToTwo(last.MuscleMassPercent - first.MuscleMassPercent),
		AverageWeight:     roundToTwo(avgWeight),
		AverageBodyFat:    roundToTwo(avgBodyFat),
		AverageMuscleMass: roundToTwo(avgMuscleMass),
	}
}

// roundToTwo rounds a float to 2 decimal places
func roundToTwo(val float64) float64 {
	return math.Round(val*100) / 100
}

// extractID extracts the ID from the URL path
func extractID(path, prefix string) (int, error) {
	idStr := strings.TrimPrefix(path, prefix)
	return strconv.Atoi(idStr)
}

package metrics

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Handler handles body metrics requests
type Handler struct {
	repo *Repository
}

// NewHandler creates a new metrics handler
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
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

	// Validate weight
	if req.Weight <= 0 {
		writeError(w, http.StatusBadRequest, "Weight must be greater than 0")
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

	// Check if metric already exists for this date
	existingMetric, err := h.repo.GetByDate(userID, metricDate)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if existingMetric != nil {
		// Update existing metric
		existingMetric.Weight = req.Weight
		if err := h.repo.Update(existingMetric); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, existingMetric)
	} else {
		// Create new metric
		metric := &BodyMetric{
			UserID: userID,
			Date:   metricDate,
			Weight: req.Weight,
		}

		if err := h.repo.Create(metric); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, metric)
	}
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

// GetWeekly handles GET /metrics/weekly
func (h *Handler) GetWeekly(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	metrics, err := h.repo.GetWeekly(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, metrics)
}

// GetByDate handles GET /metrics/date/{date}
func (h *Handler) GetByDate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Extract date from URL path
	dateStr := strings.TrimPrefix(r.URL.Path, "/metrics/date/")
	if dateStr == "" {
		writeError(w, http.StatusBadRequest, "Date is required (use YYYY-MM-DD)")
		return
	}

	// Parse date
	metricDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid date format (use YYYY-MM-DD)")
		return
	}

	// Get metric for the date
	metric, err := h.repo.GetByDate(userID, metricDate)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if metric == nil {
		writeError(w, http.StatusNotFound, "No metric found for this date")
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

	var totalWeight float64
	for _, m := range metrics {
		totalWeight += m.Weight
	}

	avgWeight := totalWeight / float64(len(metrics))

	// Calculate change (first to last)
	first := metrics[0]
	last := metrics[len(metrics)-1]

	return TrendData{
		WeightChange:  roundToTwo(last.Weight - first.Weight),
		AverageWeight: roundToTwo(avgWeight),
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

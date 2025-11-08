package metrics

import (
	"net/http"
	"strings"

	"ultra-bis/internal/auth"
)

// RegisterRoutes registers all metrics-related routes to the provided mux
func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	// All metrics routes are protected with JWT

	mux.HandleFunc("/metrics", auth.JWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetMetrics(w, r)
		case http.MethodPost:
			handler.CreateMetric(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/metrics/latest", auth.JWTMiddleware(handler.GetLatest))
	mux.HandleFunc("/metrics/weekly", auth.JWTMiddleware(handler.GetWeekly))
	mux.HandleFunc("/metrics/trends", auth.JWTMiddleware(handler.GetTrends))
	mux.HandleFunc("/metrics/date/", auth.JWTMiddleware(handler.GetByDate))

	mux.HandleFunc("/metrics/", auth.JWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimPrefix(r.URL.Path, "/metrics/") == "" {
			handler.GetMetrics(w, r)
			return
		}

		switch r.Method {
		case http.MethodDelete:
			handler.DeleteMetric(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
}

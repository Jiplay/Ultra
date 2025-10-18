package food

import (
	"net/http"
	"strings"
)

// RegisterRoutes registers all food-related routes to the provided mux
func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	// Foods collection endpoint
	mux.HandleFunc("/foods", func(w http.ResponseWriter, r *http.Request) {
		// Route based on method
		switch r.Method {
		case http.MethodGet:
			handler.GetAllFoods(w, r)
		case http.MethodPost:
			handler.CreateFood(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Individual food endpoint
	mux.HandleFunc("/foods/", func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from path
		if strings.TrimPrefix(r.URL.Path, "/foods/") == "" {
			handler.GetAllFoods(w, r)
			return
		}

		// Route to specific food handlers
		switch r.Method {
		case http.MethodGet:
			handler.GetFood(w, r)
		case http.MethodPut:
			handler.UpdateFood(w, r)
		case http.MethodDelete:
			handler.DeleteFood(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

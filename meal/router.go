package meal

import (
	"net/http"
	"strings"
)

func RegisterRoutes(mux *http.ServeMux, handlers *Handlers) {
	// Meals collection endpoint
	mux.HandleFunc("/api/v1/meals", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handlers.CreateMeal(w, r)
		case http.MethodGet:
			handlers.GetMeals(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Daily meals endpoint
	mux.HandleFunc("/api/v1/meals/daily", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetDailyMeals(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Meal summary endpoint
	mux.HandleFunc("/api/v1/meals/summary", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetMealSummary(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Meal plan endpoint
	mux.HandleFunc("/api/v1/meals/plan", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetMealPlan(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Individual meal and meal items endpoints
	mux.HandleFunc("/api/v1/meals/", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/v1/meals/") {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/api/v1/meals/")
		parts := strings.Split(path, "/")

		if len(parts) == 1 && parts[0] != "" {
			// Individual meal operations: /api/v1/meals/{id}
			switch r.Method {
			case http.MethodGet:
				handlers.GetMeal(w, r)
			case http.MethodPut:
				handlers.UpdateMeal(w, r)
			case http.MethodDelete:
				handlers.DeleteMeal(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if len(parts) == 2 && parts[1] == "items" {
			// Meal items collection: /api/v1/meals/{id}/items
			switch r.Method {
			case http.MethodPost:
				handlers.AddMealItem(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if len(parts) == 3 && parts[1] == "items" && parts[2] != "" {
			// Individual meal item: /api/v1/meals/{id}/items/{item_id}
			switch r.Method {
			case http.MethodPut:
				handlers.UpdateMealItem(w, r)
			case http.MethodDelete:
				handlers.DeleteMealItem(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})
}

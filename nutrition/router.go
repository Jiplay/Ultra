package nutrition

import (
	"net/http"
	"strings"
)

func RegisterRoutes(mux *http.ServeMux, handlers *Handlers) {
	// Foods
	mux.HandleFunc("/api/v1/nutrition/foods", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handlers.CreateFood(w, r)
		case http.MethodGet:
			handlers.GetFoods(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/nutrition/foods/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/v1/nutrition/foods/") {
			handlers.GetFood(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Recipes
	mux.HandleFunc("/api/v1/nutrition/recipes", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handlers.CreateRecipe(w, r)
		case http.MethodGet:
			handlers.GetRecipes(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	
	// Nutrition Goals
	mux.HandleFunc("/api/v1/nutrition/goals/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/nutrition/goals/") {
			switch r.Method {
			case http.MethodGet:
				handlers.GetNutritionGoals(w, r)
			case http.MethodPut:
				handlers.UpdateNutritionGoals(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})
}

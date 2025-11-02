package goal

import (
	"net/http"
	"strings"

	"ultra-bis/internal/auth"
)

// RegisterRoutes registers all goal-related routes to the provided mux
func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	// All goal routes are protected with JWT
	mux.HandleFunc("/goals", auth.JWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetActiveGoal(w, r)
		case http.MethodPost:
			handler.CreateGoal(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/goals/all", auth.JWTMiddleware(handler.GetAllGoals))
	mux.HandleFunc("/goals/recommended", auth.JWTMiddleware(handler.GetRecommendedGoals))
	mux.HandleFunc("/goals/calculate", auth.JWTMiddleware(handler.CalculateDietGoals))
	mux.HandleFunc("/goals/diets", auth.JWTMiddleware(handler.GetAvailableDiets))

	mux.HandleFunc("/goals/", auth.JWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimPrefix(r.URL.Path, "/goals/") == "" {
			handler.GetActiveGoal(w, r)
			return
		}

		switch r.Method {
		case http.MethodPut:
			handler.UpdateGoal(w, r)
		case http.MethodDelete:
			handler.DeleteGoal(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
}

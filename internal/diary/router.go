package diary

import (
	"net/http"
	"strings"

	"ultra-bis/internal/auth"
)

// RegisterRoutes registers all diary-related routes to the provided mux
func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	// All diary routes are protected with JWT

	mux.HandleFunc("/diary/entries", auth.JWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetEntries(w, r)
		case http.MethodPost:
			handler.CreateEntry(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/diary/entries/", auth.JWTMiddleware(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/diary/entries/")

		if path == "" {
			handler.GetEntries(w, r)
			return
		}

		// Check for save-as-food endpoint
		if strings.HasSuffix(path, "/save-as-food") {
			if r.Method == http.MethodPost {
				handler.SaveAsFood(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		// Check for save-as-recipe endpoint
		if strings.HasSuffix(path, "/save-as-recipe") {
			if r.Method == http.MethodPost {
				handler.SaveAsRecipe(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		switch r.Method {
		case http.MethodPut:
			handler.UpdateEntry(w, r)
		case http.MethodDelete:
			handler.DeleteEntry(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/diary/summary/", auth.JWTMiddleware(handler.GetDailySummary))
	mux.HandleFunc("/diary/weekly", auth.JWTMiddleware(handler.GetWeeklySummary))
}

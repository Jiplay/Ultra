package recipe

import (
	"net/http"
	"strings"

	"ultra-bis/internal/auth"
)

// RegisterRoutes registers recipe routes (all protected)
func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	// All recipe routes require authentication
	mux.HandleFunc("/recipes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/recipes" {
			auth.JWTMiddleware(handler.ListRecipes)(w, r)
			return
		}
		if r.Method == http.MethodPost && r.URL.Path == "/recipes" {
			auth.JWTMiddleware(handler.CreateRecipe)(w, r)
			return
		}
		http.NotFound(w, r)
	})

	// Recipe detail routes (all protected)
	mux.HandleFunc("/recipes/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.Trim(r.URL.Path, "/")
		parts := strings.Split(path, "/")

		// /recipes/{id}
		if len(parts) == 2 {
			if r.Method == http.MethodGet {
				auth.JWTMiddleware(handler.GetRecipe)(w, r)
				return
			}
			if r.Method == http.MethodPut {
				auth.JWTMiddleware(handler.UpdateRecipe)(w, r)
				return
			}
			if r.Method == http.MethodDelete {
				auth.JWTMiddleware(handler.DeleteRecipe)(w, r)
				return
			}
		}

		// /recipes/{id}/ingredients
		if len(parts) == 3 && parts[2] == "ingredients" {
			if r.Method == http.MethodPost {
				auth.JWTMiddleware(handler.AddIngredient)(w, r)
				return
			}
		}

		// /recipes/{id}/ingredients/{ingredientId}
		if len(parts) == 4 && parts[2] == "ingredients" {
			if r.Method == http.MethodPut {
				auth.JWTMiddleware(handler.UpdateIngredient)(w, r)
				return
			}
			if r.Method == http.MethodDelete {
				auth.JWTMiddleware(handler.DeleteIngredient)(w, r)
				return
			}
		}

		http.NotFound(w, r)
	})
}

package recipe

import (
	"context"
	"net/http"

	"ultra-bis/internal/auth"
	"ultra-bis/internal/httputil"
)

// RegisterRoutes registers all recipe routes with improved middleware chaining
func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	// Recipe list and creation: /recipes
	mux.HandleFunc("/recipes", func(w http.ResponseWriter, r *http.Request) {
		// Only handle exact match, not sub-paths
		if r.URL.Path != "/recipes" {
			http.NotFound(w, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			// GET /recipes - List recipes
			httputil.ChainMiddleware(
				handler.ListRecipes,
				auth.JWTMiddleware,
			)(w, r)

		case http.MethodPost:
			// POST /recipes - Create recipe
			httputil.ChainMiddleware(
				handler.CreateRecipe,
				auth.JWTMiddleware,
			)(w, r)

		default:
			httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	// Recipe operations: /recipes/{id}
	// Pattern: /recipes/123
	mux.HandleFunc("/recipes/", func(w http.ResponseWriter, r *http.Request) {
		// Parse path to determine which endpoint this is
		pathSegments := len(splitPath(r.URL.Path))

		switch pathSegments {
		case 2:
			// /recipes/{id}
			handleRecipeDetail(w, r, handler)

		case 3:
			// /recipes/{id}/ingredients
			if getPathSegment(r.URL.Path, 2) == "ingredients" {
				handleRecipeIngredients(w, r, handler)
			} else {
				http.NotFound(w, r)
			}

		case 4:
			// /recipes/{id}/ingredients/{ingredientId}
			if getPathSegment(r.URL.Path, 2) == "ingredients" {
				handleIngredientDetail(w, r, handler)
			} else {
				http.NotFound(w, r)
			}

		default:
			http.NotFound(w, r)
		}
	})
}

// handleRecipeDetail handles /recipes/{id} and /recipes/{filter}
func handleRecipeDetail(w http.ResponseWriter, r *http.Request, handler *Handler) {
	// Extract path segment
	pathSegments := splitPath(r.URL.Path)
	if len(pathSegments) < 2 {
		http.NotFound(w, r)
		return
	}

	segment := pathSegments[1]

	// Check if it's a tag filter (routine or contextual)
	if segment == "routine" || segment == "contextual" {
		if r.Method != http.MethodGet {
			httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Add tag to context and call filtered list handler
		ctx := context.WithValue(r.Context(), "tag_filter", segment)
		httputil.ChainMiddleware(
			handler.ListRecipesByTag,
			auth.JWTMiddleware,
		)(w, r.WithContext(ctx))
		return
	}

	// Otherwise, treat as numeric ID
	switch r.Method {
	case http.MethodGet:
		httputil.ChainMiddleware(
			handler.GetRecipe,
			httputil.ExtractPathID(1),
			auth.JWTMiddleware,
		)(w, r)

	case http.MethodPut:
		httputil.ChainMiddleware(
			handler.UpdateRecipe,
			httputil.ExtractPathID(1),
			auth.JWTMiddleware,
		)(w, r)

	case http.MethodDelete:
		httputil.ChainMiddleware(
			handler.DeleteRecipe,
			httputil.ExtractPathID(1),
			auth.JWTMiddleware,
		)(w, r)

	default:
		httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleRecipeIngredients handles /recipes/{id}/ingredients
func handleRecipeIngredients(w http.ResponseWriter, r *http.Request, handler *Handler) {
	switch r.Method {
	case http.MethodPost:
		httputil.ChainMiddleware(
			handler.AddIngredient,
			httputil.ExtractPathID(1),
			auth.JWTMiddleware,
		)(w, r)

	default:
		httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleIngredientDetail handles /recipes/{id}/ingredients/{ingredientId}
func handleIngredientDetail(w http.ResponseWriter, r *http.Request, handler *Handler) {
	switch r.Method {
	case http.MethodPut:
		httputil.ChainMiddleware(
			handler.UpdateIngredient,
			httputil.ExtractTwoPathIDs(1, 3),
			auth.JWTMiddleware,
		)(w, r)

	case http.MethodDelete:
		httputil.ChainMiddleware(
			handler.DeleteIngredient,
			httputil.ExtractTwoPathIDs(1, 3),
			auth.JWTMiddleware,
		)(w, r)

	default:
		httputil.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// Helper functions for path parsing

func splitPath(path string) []string {
	segments := []string{}
	current := ""

	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if current != "" {
				segments = append(segments, current)
				current = ""
			}
		} else {
			current += string(path[i])
		}
	}

	if current != "" {
		segments = append(segments, current)
	}

	return segments
}

func getPathSegment(path string, index int) string {
	segments := splitPath(path)
	if index < len(segments) {
		return segments[index]
	}
	return ""
}

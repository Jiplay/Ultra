package programs

import (
	"net/http"
	"strings"
)

func RegisterRoutes(mux *http.ServeMux, handlers *Handlers) {
	// Programs collection endpoint
	mux.HandleFunc("/api/v1/programs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handlers.CreateProgram(w, r)
		case http.MethodGet:
			handlers.GetPrograms(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Programs individual endpoint
	mux.HandleFunc("/api/v1/programs/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/programs/") {
			switch r.Method {
			case http.MethodGet:
				handlers.GetProgram(w, r)
			case http.MethodPut:
				handlers.UpdateProgram(w, r)
			case http.MethodDelete:
				handlers.DeleteProgram(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})
}
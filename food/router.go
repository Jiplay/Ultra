package food

import (
	"net/http"
	"strings"
)

func SetupRoutes(mux *http.ServeMux, handlers *Handlers) {
	mux.HandleFunc("/api/foods", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/foods")
		
		if path == "" || path == "/" {
			switch r.Method {
			case http.MethodPost:
				handlers.CreateFood(w, r)
			case http.MethodGet:
				handlers.GetAllFoods(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if strings.HasPrefix(path, "/") {
			idPath := strings.TrimPrefix(path, "/")
			if strings.Contains(idPath, "/") {
				http.Error(w, "Not found", http.StatusNotFound)
				return
			}
			
			switch r.Method {
			case http.MethodGet:
				handlers.GetFood(w, r)
			case http.MethodPut:
				handlers.UpdateFood(w, r)
			case http.MethodDelete:
				handlers.DeleteFood(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})
}
package nutrition

import (
	"database/sql"
	"net/http"
)

func Setup(db *sql.DB, mux *http.ServeMux) {
	repo := NewPostgresRepository(db)
	controller := NewController(repo)
	handlers := NewHandlers(controller)
	
	RegisterRoutes(mux, handlers)
}
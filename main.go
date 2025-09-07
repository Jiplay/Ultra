package main

import (
	"fmt"
	"log"
	"net/http"

	"ultra/food"
)

func main() {
	repo := food.NewInMemoryRepository()
	controller := food.NewController(repo)
	handlers := food.NewHandlers(controller)

	mux := http.NewServeMux()
	food.SetupRoutes(mux, handlers)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprintln(w, "Food Catalog API - Available endpoints:")
		fmt.Fprintln(w, "POST   /api/foods       - Create a new food")
		fmt.Fprintln(w, "GET    /api/foods       - Get all foods")
		fmt.Fprintln(w, "GET    /api/foods/{id}  - Get food by ID")
		fmt.Fprintln(w, "PUT    /api/foods/{id}  - Update food by ID")
		fmt.Fprintln(w, "DELETE /api/foods/{id}  - Delete food by ID")
	})

	fmt.Println("Starting server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

package main

import (
	"log"
	"net/http"
	"ultra.com/food/internal/controller/food"
	httphandler "ultra.com/food/internal/handler/http"
	"ultra.com/food/internal/repository/memory"
)

func main() {
	log.Printf("Starting the food service")
	repo := memory.New()
	ctrl := food.New(repo)
	h := httphandler.New(ctrl)
	http.Handle("/recipe", http.HandlerFunc(h.GetRecipe))
	if err := http.ListenAndServe(":8081", nil); err != nil {
		panic(err)
	}
}

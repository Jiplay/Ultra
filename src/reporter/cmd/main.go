package main

import (
	"log"
	"net/http"
	"ultra.com/reporter/internal/controller/health"
	foodgateway "ultra.com/reporter/internal/gateway/food/http"
	sportgateway "ultra.com/reporter/internal/gateway/sport/http"
	httphandler "ultra.com/reporter/internal/handler/http"
)

func main() {
	log.Printf("Starting the reporter service")
	foodGateway := foodgateway.New("localhost:8081")
	sportGateway := sportgateway.New("localhost:8082")
	ctrl := health.New(sportGateway, foodGateway)
	h := httphandler.New(ctrl)
	http.Handle("/recipe", http.HandlerFunc(h.GetRecipe))
	http.Handle("/workout", http.HandlerFunc(h.GetWorkout))
	http.Handle("/performance", http.HandlerFunc(h.GetPerformance))
	if err := http.ListenAndServe(":8080", nil); err != nil {
	}
}

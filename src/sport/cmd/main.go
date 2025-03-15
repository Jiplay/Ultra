package main

import (
	"log"
	"net/http"
	"ultra.com/sport/internal/controller/sport"
	httphandler "ultra.com/sport/internal/handler/http"
	"ultra.com/sport/internal/repository/memory"
)

func main() {
	log.Printf("Starting up the sport service")
	repo := memory.New()
	ctrl := sport.New(repo)
	h := httphandler.New(ctrl)
	http.Handle("/plans", http.HandlerFunc(h.GetWorkoutPlans))
	http.Handle("/performances", http.HandlerFunc(h.GetPerformances))
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

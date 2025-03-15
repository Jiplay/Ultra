package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
	"ultra.com/pkg/discovery"
	"ultra.com/pkg/discovery/consul"
	"ultra.com/sport/internal/controller/sport"
	httphandler "ultra.com/sport/internal/handler/http"
	"ultra.com/sport/internal/repository/memory"
)

const serviceName = "sport"

func main() {
	var port int
	flag.IntVar(&port, "port", 8082, "API listen on")
	log.Printf("Starting the sport service on port %d", port)
	registry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("localhost:%d", port)); err != nil {
		panic(err)
	}
	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
				log.Println("Failed to report healthy state:", err.Error())
			}
			time.Sleep(1 * time.Second)
		}
	}()
	defer registry.Deregister(ctx, instanceID, serviceName)
	log.Printf("Starting up the sport service")
	repo := memory.New()
	ctrl := sport.New(repo)
	h := httphandler.New(ctrl)
	http.Handle("/workout", http.HandlerFunc(h.GetWorkout))
	http.Handle("/performance", http.HandlerFunc(h.GetPerformance))
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

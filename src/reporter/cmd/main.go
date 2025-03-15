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
	"ultra.com/reporter/internal/controller/health"
	foodgateway "ultra.com/reporter/internal/gateway/food/http"
	sportgateway "ultra.com/reporter/internal/gateway/sport/http"
	httphandler "ultra.com/reporter/internal/handler/http"
)

const serviceName = "reporter"

func main() {
	var port int
	flag.IntVar(&port, "port", 8083, "API listen on")
	log.Printf("Starting the reporter service on port %d", port)
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

	log.Printf("Starting the reporter service")
	foodGateway := foodgateway.New(registry)
	sportGateway := sportgateway.New(registry)
	ctrl := health.New(sportGateway, foodGateway)
	h := httphandler.New(ctrl)
	http.Handle("/recipe", http.HandlerFunc(h.GetRecipe))
	http.Handle("/workout", http.HandlerFunc(h.GetWorkout))
	http.Handle("/performance", http.HandlerFunc(h.GetPerformance))
	if err := http.ListenAndServe(":8083", nil); err != nil {
		panic(err)
	}
}

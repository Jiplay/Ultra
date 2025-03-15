package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
	"ultra.com/food/internal/controller/food"
	httphandler "ultra.com/food/internal/handler/http"
	"ultra.com/food/internal/repository/memory"
	"ultra.com/pkg/discovery"
	"ultra.com/pkg/discovery/consul"
)

const serviceName = "food"

func main() {
	var port int
	flag.IntVar(&port, "port", 8081, "API listen on")
	log.Printf("Starting the food service on port %d", port)
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
	repo := memory.New()
	ctrl := food.New(repo)
	h := httphandler.New(ctrl)
	http.Handle("/recipe", http.HandlerFunc(h.GetRecipe))
	if err := http.ListenAndServe(":8081", nil); err != nil {
		panic(err)
	}
}

package main

import (
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"time"
	"ultra.com/food/internal/controller/food"
	grpchandler "ultra.com/food/internal/handler/grpc"
	"ultra.com/food/internal/repository/memory"
	"ultra.com/gen"
	"ultra.com/pkg/discovery"
	"ultra.com/pkg/discovery/consul"
)

const serviceName = "food"

func main() {
	var port int
	flag.IntVar(&port, "port", 8081, "API listen on")
	flag.Parse()
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
	h := grpchandler.New(ctrl)
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%v", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	reflection.Register(srv)
	gen.RegisterFoodServiceServer(srv, h)
	if err := srv.Serve(lis); err != nil {
	}
}

package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"ultra/config"
	"ultra/database"
	"ultra/nutrition"
	"ultra/programs"
	"ultra/users"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to databases
	database.ConnectMongoDB(cfg.MongoURI)
	database.ConnectPostgres(cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresUser,
		cfg.PostgresPass, cfg.PostgresDB, cfg.PostgresSSL)

	// Initialize PostgreSQL tables
	if err := database.InitPostgresTables(); err != nil {
		log.Fatal("Failed to initialize PostgreSQL tables:", err)
	}

	// Set up router
	router := mux.NewRouter()

	// API versioning
	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","message":"Ultra API is running"}`))
	}).Methods("GET")

	// Register routes
	users.RegisterRoutes(apiRouter)
	programs.RegisterRoutes(apiRouter)
	nutrition.RegisterRoutes(apiRouter)

	// CORS middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-ID")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Logging middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%s %s %s", r.Method, r.RequestURI, r.RemoteAddr)
			next.ServeHTTP(w, r)
		})
	})

	log.Printf("Server starting on port %s", cfg.Port)
	log.Printf("MongoDB connected")
	log.Printf("PostgreSQL connected and tables initialized")
	log.Printf("Available endpoints:")
	log.Printf("  GET  /health")
	log.Printf("  POST /api/v1/users")
	log.Printf("  GET  /api/v1/users/{id}")
	log.Printf("  PUT  /api/v1/users/{id}")
	log.Printf("  DEL  /api/v1/users/{id}")
	log.Printf("  POST /api/v1/programs")
	log.Printf("  GET  /api/v1/programs")
	log.Printf("  GET  /api/v1/programs/{id}")
	log.Printf("  PUT  /api/v1/programs/{id}")
	log.Printf("  DEL  /api/v1/programs/{id}")
	log.Printf("  POST /api/v1/nutrition/foods")
	log.Printf("  GET  /api/v1/nutrition/foods")
	log.Printf("  GET  /api/v1/nutrition/foods/{id}")
	log.Printf("  POST /api/v1/nutrition/recipes")
	log.Printf("  GET  /api/v1/nutrition/recipes")
	log.Printf("  GET  /api/v1/nutrition/recipes/{id}")
	log.Printf("  GET  /api/v1/nutrition/goals/{user_id}")
	log.Printf("  PUT  /api/v1/nutrition/goals/{user_id}")

	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

package main

import (
	"log"
	"net/http"

	"ultra/config"
	"ultra/database"
	"ultra/meal"
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
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","message":"Ultra API is running"}`))
	})

	// Register routes
	nutrition.Setup(database.PostgresDB, mux)
	programs.Setup(database.PostgresDB, mux)
	meal.Setup(database.PostgresDB, mux)
	users.Setup(database.MongoDB, mux)

	// Create server with middleware
	handler := corsMiddleware(loggingMiddleware(mux))

	log.Printf("Server starting on port %s", cfg.Port)
	log.Printf("MongoDB connected")
	log.Printf("PostgreSQL connected and tables initialized")
	log.Printf("Available endpoints:")
	log.Printf("  GET  /health")
	log.Printf("  POST /api/v1/nutrition/foods")
	log.Printf("  GET  /api/v1/nutrition/foods")
	log.Printf("  GET  /api/v1/nutrition/foods/{id}")
	log.Printf("  POST /api/v1/nutrition/recipes")
	log.Printf("  GET  /api/v1/nutrition/recipes")
	log.Printf("  GET  /api/v1/nutrition/recipes/{id}")
	log.Printf("  GET  /api/v1/nutrition/goals/{user_id}")
	log.Printf("  PUT  /api/v1/nutrition/goals/{user_id}")
	log.Printf("  POST /api/v1/programs")
	log.Printf("  GET  /api/v1/programs")
	log.Printf("  GET  /api/v1/programs/{id}")
	log.Printf("  PUT  /api/v1/programs/{id}")
	log.Printf("  DEL  /api/v1/programs/{id}")
	log.Printf("  POST /api/v1/meals")
	log.Printf("  GET  /api/v1/meals")
	log.Printf("  GET  /api/v1/meals/{id}")
	log.Printf("  PUT  /api/v1/meals/{id}")
	log.Printf("  DEL  /api/v1/meals/{id}")
	log.Printf("  GET  /api/v1/meals/daily")
	log.Printf("  GET  /api/v1/meals/summary")
	log.Printf("  GET  /api/v1/meals/plan")
	log.Printf("  POST /api/v1/meals/{id}/items")
	log.Printf("  PUT  /api/v1/meals/{id}/items/{item_id}")
	log.Printf("  DEL  /api/v1/meals/{id}/items/{item_id}")
	log.Printf("  POST /api/v1/users")
	log.Printf("  GET  /api/v1/users/{id}")
	log.Printf("  PUT  /api/v1/users/{id}")
	log.Printf("  DEL  /api/v1/users/{id}")

	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
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
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.RequestURI, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

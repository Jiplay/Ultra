package main

import (
	"log"
	"net/http"
	"os"

	"ultra-bis/internal/database"
	"ultra-bis/internal/food"
)

func main() {
	// Connect to database
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate database schema (like Sequelize sync)
	if err := db.AutoMigrate(&food.Food{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	log.Println("Database migration completed")

	// Initialize food repository and handler
	foodRepo := food.NewRepository(db)
	foodHandler := food.NewHandler(foodRepo)

	// Setup routes
	mux := http.NewServeMux()

	// Register food routes
	food.RegisterRoutes(mux, foodHandler)

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start server
	port := getEnv("PORT", "8080")
	log.Printf("Server starting on port %s", port)
	log.Printf("Available endpoints:")
	log.Printf("  GET    /health        - Health check")
	log.Printf("  POST   /foods         - Create food")
	log.Printf("  GET    /foods         - List all foods")
	log.Printf("  GET    /foods/{id}    - Get food by ID")
	log.Printf("  PUT    /foods/{id}    - Update food")
	log.Printf("  DELETE /foods/{id}    - Delete food")

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

// getEnv retrieves environment variable or returns default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

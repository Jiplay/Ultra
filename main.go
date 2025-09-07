package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"ultra/database"
	"ultra/food"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Determine which repository to use based on environment
	useDatabase := os.Getenv("USE_DATABASE") == "true"
	var repo food.Repository

	if useDatabase {
		// Database setup
		log.Println("Initializing PostgreSQL database connection...")
		dbConfig := database.NewConfigFromEnv()
		db, err := database.Connect(dbConfig)
		if err != nil {
			log.Printf("Failed to connect to database, falling back to in-memory storage: %v", err)
			repo = food.NewInMemoryRepository()
		} else {
			defer db.Close()

			// Run database migrations
			if err := database.Migrate(db); err != nil {
				log.Fatalf("Failed to migrate database: %v", err)
			}

			repo = food.NewPostgreSQLRepository(db)
			log.Println("Successfully connected to PostgreSQL database")
		}
	} else {
		log.Println("Using in-memory storage (set USE_DATABASE=true for PostgreSQL)")
		repo = food.NewInMemoryRepository()
	}

	// Initialize controllers and handlers
	controller := food.NewController(repo)
	handlers := food.NewHandlers(controller)

	mux := http.NewServeMux()
	food.SetupRoutes(mux, handlers)

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		healthStatus := map[string]string{
			"status": "healthy",
			"storage": "in-memory",
		}

		if useDatabase {
			// Check if we have a PostgreSQL repository
			if _, ok := repo.(*food.PostgreSQLRepository); ok {
				healthStatus["storage"] = "postgresql"
				healthStatus["database"] = "connected"
			} else {
				healthStatus["storage"] = "in-memory (database fallback)"
			}
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"%s","storage":"%s"}`+"\n", 
			healthStatus["status"], healthStatus["storage"])
	})

	// Root endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		
		storageType := "In-Memory"
		if useDatabase {
			storageType = "PostgreSQL Database"
		}
		
		fmt.Fprintln(w, "🍎 Ultra Food Catalog API")
		fmt.Fprintln(w, "========================")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "Available endpoints:")
		fmt.Fprintln(w, "POST   /api/foods       - Create a new food")
		fmt.Fprintln(w, "GET    /api/foods       - Get all foods")
		fmt.Fprintln(w, "GET    /api/foods/{id}  - Get food by ID")
		fmt.Fprintln(w, "PUT    /api/foods/{id}  - Update food by ID")
		fmt.Fprintln(w, "DELETE /api/foods/{id}  - Delete food by ID")
		fmt.Fprintln(w, "GET    /health          - Health check")
		fmt.Fprintln(w, "")
		fmt.Fprintf(w, "Storage: %s\n", storageType)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 Starting Ultra Food API on port %s...\n", port)
	fmt.Printf("📊 Storage: %s\n", map[bool]string{true: "PostgreSQL", false: "In-Memory"}[useDatabase])
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

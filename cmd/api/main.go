package main

import (
	"log"
	"net/http"
	"os"

	"ultra-bis/internal/auth"
	"ultra-bis/internal/database"
	"ultra-bis/internal/diary"
	"ultra-bis/internal/food"
	"ultra-bis/internal/goal"
	"ultra-bis/internal/metrics"
	"ultra-bis/internal/middleware"
	"ultra-bis/internal/recipe"
	"ultra-bis/internal/user"
)

func main() {
	// Connect to database
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run custom migration to convert to gram-based system
	if err := database.MigrateToGrams(db); err != nil {
		log.Fatal("Failed to run gram migration:", err)
	}

	// Run custom migration to populate custom ingredients
	if err := database.MigrateCustomIngredients(db); err != nil {
		log.Fatal("Failed to run custom ingredients migration:", err)
	}

	// Auto-migrate database schema (like Sequelize sync)
	log.Println("Running database migrations...")
	if err := db.AutoMigrate(
		&user.User{},
		&food.Food{},
		&recipe.Recipe{},
		&recipe.RecipeIngredient{},
		&goal.NutritionGoal{},
		&diary.DiaryEntry{},
		&metrics.BodyMetric{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	log.Println("Database migration completed")

	// Initialize repositories
	userRepo := user.NewRepository(db)
	foodRepo := food.NewRepository(db)
	recipeRepo := recipe.NewRepository(db)
	goalRepo := goal.NewRepository(db)
	diaryRepo := diary.NewRepository(db)
	metricsRepo := metrics.NewRepository(db)

	// Initialize handlers
	authHandler := auth.NewHandler(userRepo)
	foodHandler := food.NewHandler(foodRepo)
	recipeHandler := recipe.NewHandler(recipeRepo, foodRepo)
	goalHandler := goal.NewHandler(goalRepo, userRepo)
	diaryHandler := diary.NewHandler(diaryRepo, foodRepo, goalRepo)
	metricsHandler := metrics.NewHandler(metricsRepo)

	// Set recipe repository in diary handler (to avoid circular dependency)
	recipeAdapter := recipe.NewDiaryRecipeAdapter(recipeRepo)
	diaryHandler.SetRecipeRepo(recipeAdapter)

	// Setup routes
	mux := http.NewServeMux()

	// Register all routes
	auth.RegisterRoutes(mux, authHandler)
	food.RegisterRoutes(mux, foodHandler)
	recipe.RegisterRoutes(mux, recipeHandler)
	goal.RegisterRoutes(mux, goalHandler)
	diary.RegisterRoutes(mux, diaryHandler)
	metrics.RegisterRoutes(mux, metricsHandler)

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start server
	port := getEnv("PORT", "8080")
	log.Printf("Server starting on port %s", port)
	log.Println("===========================================")
	log.Println("Available endpoints:")
	log.Println("-------------------------------------------")
	log.Println("AUTH:")
	log.Println("  POST   /auth/register          - Register new user")
	log.Println("  POST   /auth/login             - Login")
	log.Println("  GET    /auth/me                - Get current user (protected)")
	log.Println("  PUT    /users/profile          - Update profile (protected)")
	log.Println("-------------------------------------------")
	log.Println("FOODS:")
	log.Println("  POST   /foods                  - Create food")
	log.Println("  GET    /foods                  - List all foods")
	log.Println("  GET    /foods/{id}             - Get food by ID")
	log.Println("  PUT    /foods/{id}             - Update food")
	log.Println("  DELETE /foods/{id}             - Delete food")
	log.Println("-------------------------------------------")
	log.Println("RECIPES:")
	log.Println("  POST   /recipes                - Create recipe (protected)")
	log.Println("  GET    /recipes                - List recipes (protected, query: user_only=true/false)")
	log.Println("  GET    /recipes/{id}           - Get recipe with nutrition (protected)")
	log.Println("  PUT    /recipes/{id}           - Update recipe (protected)")
	log.Println("  DELETE /recipes/{id}           - Delete recipe (protected)")
	log.Println("  POST   /recipes/{id}/ingredients      - Add ingredient (protected)")
	log.Println("  PUT    /recipes/{id}/ingredients/{iid} - Update ingredient (protected)")
	log.Println("  DELETE /recipes/{id}/ingredients/{iid} - Remove ingredient (protected)")
	log.Println("-------------------------------------------")
	log.Println("NUTRITION GOALS:")
	log.Println("  POST   /goals                  - Create goal (protected)")
	log.Println("  GET    /goals                  - Get active goal (protected)")
	log.Println("  GET    /goals/all              - Get all goals (protected)")
	log.Println("  POST   /goals/recommended      - Calculate recommended goals (protected)")
	log.Println("  PUT    /goals/{id}             - Update goal (protected)")
	log.Println("  DELETE /goals/{id}             - Delete goal (protected)")
	log.Println("-------------------------------------------")
	log.Println("DIARY (MEAL LOGGING):")
	log.Println("  POST   /diary/entries          - Log food/meal (protected)")
	log.Println("  GET    /diary/entries?date=... - Get entries by date (protected)")
	log.Println("  GET    /diary/summary/{date}   - Get daily summary (protected)")
	log.Println("  PUT    /diary/entries/{id}     - Update entry (protected)")
	log.Println("  DELETE /diary/entries/{id}     - Delete entry (protected)")
	log.Println("-------------------------------------------")
	log.Println("BODY METRICS:")
	log.Println("  POST   /metrics                - Log body metrics (protected)")
	log.Println("  GET    /metrics                - Get all metrics (protected)")
	log.Println("  GET    /metrics/latest         - Get latest metrics (protected)")
	log.Println("  GET    /metrics/trends?period=7d|30d|90d - Get trends (protected)")
	log.Println("  DELETE /metrics/{id}           - Delete metric (protected)")
	log.Println("-------------------------------------------")
	log.Println("HEALTH:")
	log.Println("  GET    /health                 - Health check")
	log.Println("===========================================")

	// Wrap the mux with logging middleware
	loggedHandler := middleware.LoggingMiddleware(mux)

	if err := http.ListenAndServe(":"+port, loggedHandler); err != nil {
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

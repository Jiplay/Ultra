package nutrition

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"ultra/database"
)

func CreateFoodHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateFoodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	var foodID int
	err := database.PostgresDB.QueryRow(`
		INSERT INTO foods (name, calories_per_100g, protein_per_100g, carbs_per_100g, fat_per_100g, fiber_per_100g, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, req.Name, req.CaloriesPer100g, req.ProteinPer100g, req.CarbsPer100g, req.FatPer100g, req.FiberPer100g, userID).Scan(&foodID)

	if err != nil {
		http.Error(w, "Failed to create food", http.StatusInternalServerError)
		return
	}

	food := Food{
		ID:              foodID,
		Name:            req.Name,
		CaloriesPer100g: req.CaloriesPer100g,
		ProteinPer100g:  req.ProteinPer100g,
		CarbsPer100g:    req.CarbsPer100g,
		FatPer100g:      req.FatPer100g,
		FiberPer100g:    req.FiberPer100g,
		CreatedBy:       userID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(food)
}

func GetFoodsHandler(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	
	query := `
		SELECT id, name, calories_per_100g, protein_per_100g, carbs_per_100g, fat_per_100g, fiber_per_100g, created_by, created_at
		FROM foods
	`
	args := []interface{}{}
	
	if search != "" {
		query += " WHERE name ILIKE $1"
		args = append(args, "%"+search+"%")
	}
	
	query += " ORDER BY name"

	rows, err := database.PostgresDB.Query(query, args...)
	if err != nil {
		http.Error(w, "Failed to get foods", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var foods []Food
	for rows.Next() {
		var food Food
		err := rows.Scan(&food.ID, &food.Name, &food.CaloriesPer100g, &food.ProteinPer100g,
						&food.CarbsPer100g, &food.FatPer100g, &food.FiberPer100g, 
						&food.CreatedBy, &food.CreatedAt)
		if err != nil {
			http.Error(w, "Failed to scan food", http.StatusInternalServerError)
			return
		}
		foods = append(foods, food)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(foods)
}

func GetFoodHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	foodID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid food ID", http.StatusBadRequest)
		return
	}

	var food Food
	err = database.PostgresDB.QueryRow(`
		SELECT id, name, calories_per_100g, protein_per_100g, carbs_per_100g, fat_per_100g, fiber_per_100g, created_by, created_at
		FROM foods
		WHERE id = $1
	`, foodID).Scan(&food.ID, &food.Name, &food.CaloriesPer100g, &food.ProteinPer100g,
		&food.CarbsPer100g, &food.FatPer100g, &food.FiberPer100g, 
		&food.CreatedBy, &food.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Food not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get food", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(food)
}

func CreateRecipeHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	tx, err := database.PostgresDB.Begin()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var recipeID int
	err = tx.QueryRow(`
		INSERT INTO recipes (user_id, name, description, instructions, servings, prep_time, cook_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, userID, req.Name, req.Description, req.Instructions, req.Servings, req.PrepTime, req.CookTime).Scan(&recipeID)

	if err != nil {
		http.Error(w, "Failed to create recipe", http.StatusInternalServerError)
		return
	}

	recipe := Recipe{
		ID:           recipeID,
		UserID:       userID,
		Name:         req.Name,
		Description:  req.Description,
		Instructions: req.Instructions,
		Servings:     req.Servings,
		PrepTime:     req.PrepTime,
		CookTime:     req.CookTime,
		Ingredients:  []RecipeIngredient{},
	}

	// Add ingredients
	for _, ingredientReq := range req.Ingredients {
		var ingredientID int
		err = tx.QueryRow(`
			INSERT INTO recipe_ingredients (recipe_id, food_id, quantity, unit)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, recipeID, ingredientReq.FoodID, ingredientReq.Quantity, ingredientReq.Unit).Scan(&ingredientID)

		if err != nil {
			http.Error(w, "Failed to create recipe ingredient", http.StatusInternalServerError)
			return
		}

		ingredient := RecipeIngredient{
			ID:       ingredientID,
			RecipeID: recipeID,
			FoodID:   ingredientReq.FoodID,
			Quantity: ingredientReq.Quantity,
			Unit:     ingredientReq.Unit,
		}

		recipe.Ingredients = append(recipe.Ingredients, ingredient)
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(recipe)
}

func GetRecipesHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	rows, err := database.PostgresDB.Query(`
		SELECT id, user_id, name, description, instructions, servings, prep_time, cook_time, created_at, updated_at
		FROM recipes
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		http.Error(w, "Failed to get recipes", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var recipes []Recipe
	for rows.Next() {
		var recipe Recipe
		err := rows.Scan(&recipe.ID, &recipe.UserID, &recipe.Name, &recipe.Description,
						&recipe.Instructions, &recipe.Servings, &recipe.PrepTime, 
						&recipe.CookTime, &recipe.CreatedAt, &recipe.UpdatedAt)
		if err != nil {
			http.Error(w, "Failed to scan recipe", http.StatusInternalServerError)
			return
		}
		recipes = append(recipes, recipe)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recipes)
}

func GetRecipeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	recipeID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	var recipe Recipe
	err = database.PostgresDB.QueryRow(`
		SELECT id, user_id, name, description, instructions, servings, prep_time, cook_time, created_at, updated_at
		FROM recipes
		WHERE id = $1 AND user_id = $2
	`, recipeID, userID).Scan(&recipe.ID, &recipe.UserID, &recipe.Name, &recipe.Description,
		&recipe.Instructions, &recipe.Servings, &recipe.PrepTime, 
		&recipe.CookTime, &recipe.CreatedAt, &recipe.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Recipe not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get recipe", http.StatusInternalServerError)
		return
	}

	// Get ingredients
	ingredientRows, err := database.PostgresDB.Query(`
		SELECT ri.id, ri.recipe_id, ri.food_id, ri.quantity, ri.unit,
			   f.name, f.calories_per_100g, f.protein_per_100g, f.carbs_per_100g, f.fat_per_100g, f.fiber_per_100g
		FROM recipe_ingredients ri
		JOIN foods f ON ri.food_id = f.id
		WHERE ri.recipe_id = $1
	`, recipeID)
	if err != nil {
		http.Error(w, "Failed to get recipe ingredients", http.StatusInternalServerError)
		return
	}
	defer ingredientRows.Close()

	recipe.Ingredients = []RecipeIngredient{}
	for ingredientRows.Next() {
		var ingredient RecipeIngredient
		var food Food
		err := ingredientRows.Scan(&ingredient.ID, &ingredient.RecipeID, &ingredient.FoodID,
								  &ingredient.Quantity, &ingredient.Unit,
								  &food.Name, &food.CaloriesPer100g, &food.ProteinPer100g,
								  &food.CarbsPer100g, &food.FatPer100g, &food.FiberPer100g)
		if err != nil {
			http.Error(w, "Failed to scan ingredient", http.StatusInternalServerError)
			return
		}
		food.ID = ingredient.FoodID
		ingredient.Food = &food
		recipe.Ingredients = append(recipe.Ingredients, ingredient)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recipe)
}

func UpdateNutritionGoalsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	var req UpdateNutritionGoalsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Check if goals exist
	var existingID int
	err := database.PostgresDB.QueryRow("SELECT id FROM nutrition_goals WHERE user_id = $1", userID).Scan(&existingID)
	
	if err == sql.ErrNoRows {
		// Create new goals
		var goalID int
		err = database.PostgresDB.QueryRow(`
			INSERT INTO nutrition_goals (user_id, daily_calories, daily_protein, daily_carbs, daily_fat, daily_fiber)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`, userID, 
			getIntValue(req.DailyCalories, 2000),
			getFloatValue(req.DailyProtein, 150.0),
			getFloatValue(req.DailyCarbs, 250.0),
			getFloatValue(req.DailyFat, 67.0),
			getFloatValue(req.DailyFiber, 25.0)).Scan(&goalID)
		
		if err != nil {
			http.Error(w, "Failed to create nutrition goals", http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		http.Error(w, "Failed to check existing goals", http.StatusInternalServerError)
		return
	} else {
		// Update existing goals
		query := "UPDATE nutrition_goals SET updated_at = CURRENT_TIMESTAMP"
		args := []interface{}{}
		argCount := 0

		if req.DailyCalories != nil {
			argCount++
			query += ", daily_calories = $" + strconv.Itoa(argCount)
			args = append(args, *req.DailyCalories)
		}
		if req.DailyProtein != nil {
			argCount++
			query += ", daily_protein = $" + strconv.Itoa(argCount)
			args = append(args, *req.DailyProtein)
		}
		if req.DailyCarbs != nil {
			argCount++
			query += ", daily_carbs = $" + strconv.Itoa(argCount)
			args = append(args, *req.DailyCarbs)
		}
		if req.DailyFat != nil {
			argCount++
			query += ", daily_fat = $" + strconv.Itoa(argCount)
			args = append(args, *req.DailyFat)
		}
		if req.DailyFiber != nil {
			argCount++
			query += ", daily_fiber = $" + strconv.Itoa(argCount)
			args = append(args, *req.DailyFiber)
		}

		argCount++
		query += " WHERE user_id = $" + strconv.Itoa(argCount)
		args = append(args, userID)

		_, err = database.PostgresDB.Exec(query, args...)
		if err != nil {
			http.Error(w, "Failed to update nutrition goals", http.StatusInternalServerError)
			return
		}
	}

	// Get updated goals
	var goals NutritionGoals
	err = database.PostgresDB.QueryRow(`
		SELECT id, user_id, daily_calories, daily_protein, daily_carbs, daily_fat, daily_fiber, updated_at
		FROM nutrition_goals
		WHERE user_id = $1
	`, userID).Scan(&goals.ID, &goals.UserID, &goals.DailyCalories, &goals.DailyProtein,
		&goals.DailyCarbs, &goals.DailyFat, &goals.DailyFiber, &goals.UpdatedAt)

	if err != nil {
		http.Error(w, "Failed to get updated goals", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(goals)
}

func GetNutritionGoalsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	var goals NutritionGoals
	err := database.PostgresDB.QueryRow(`
		SELECT id, user_id, daily_calories, daily_protein, daily_carbs, daily_fat, daily_fiber, updated_at
		FROM nutrition_goals
		WHERE user_id = $1
	`, userID).Scan(&goals.ID, &goals.UserID, &goals.DailyCalories, &goals.DailyProtein,
		&goals.DailyCarbs, &goals.DailyFat, &goals.DailyFiber, &goals.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Nutrition goals not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get nutrition goals", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(goals)
}

func getIntValue(ptr *int, defaultValue int) int {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

func getFloatValue(ptr *float64, defaultValue float64) float64 {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

func RegisterRoutes(router *mux.Router) {
	nutritionRouter := router.PathPrefix("/nutrition").Subrouter()
	
	// Foods
	nutritionRouter.HandleFunc("/foods", CreateFoodHandler).Methods("POST")
	nutritionRouter.HandleFunc("/foods", GetFoodsHandler).Methods("GET")
	nutritionRouter.HandleFunc("/foods/{id}", GetFoodHandler).Methods("GET")
	
	// Recipes
	nutritionRouter.HandleFunc("/recipes", CreateRecipeHandler).Methods("POST")
	nutritionRouter.HandleFunc("/recipes", GetRecipesHandler).Methods("GET")
	nutritionRouter.HandleFunc("/recipes/{id}", GetRecipeHandler).Methods("GET")
	
	// Nutrition Goals
	nutritionRouter.HandleFunc("/goals/{user_id}", GetNutritionGoalsHandler).Methods("GET")
	nutritionRouter.HandleFunc("/goals/{user_id}", UpdateNutritionGoalsHandler).Methods("PUT")
}
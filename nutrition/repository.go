package nutrition

import (
	"database/sql"
	"strconv"
)

type Repository interface {
	CreateFood(food *Food) error
	GetFoods(search string) ([]Food, error)
	GetFoodByID(id int) (*Food, error)

	CreateRecipe(recipe *Recipe) error
	CreateRecipeIngredient(ingredient *RecipeIngredient) error
	GetRecipesByUserID(userID string) ([]Recipe, error)
	GetRecipeIngredients(recipeID int) ([]RecipeIngredient, error)

	CreateNutritionGoals(goals *Goals) error
	UpdateNutritionGoals(userID string, goals *UpdateNutritionGoalsRequest) error
	GetNutritionGoalsByUserID(userID string) (*Goals, error)
	CheckNutritionGoalsExist(userID string) (int, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) Repository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateFood(food *Food) error {
	err := r.db.QueryRow(`
		INSERT INTO foods (name, calories_per_100g, protein_per_100g, carbs_per_100g, fat_per_100g, fiber_per_100g, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, food.Name, food.CaloriesPer100g, food.ProteinPer100g, food.CarbsPer100g, food.FatPer100g, food.FiberPer100g, food.CreatedBy).Scan(&food.ID)
	return err
}

func (r *PostgresRepository) GetFoods(search string) ([]Food, error) {
	query := `
		SELECT id, name, calories_per_100g, protein_per_100g, carbs_per_100g, fat_per_100g, fiber_per_100g, created_by, created_at
		FROM foods
	`
	args := []interface{}{}

	if search != "" {
		query += " WHERE name LIKE $1"
		args = append(args, "%"+search+"%")
	}

	query += " ORDER BY name"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var foods []Food
	for rows.Next() {
		var food Food
		err := rows.Scan(&food.ID, &food.Name, &food.CaloriesPer100g, &food.ProteinPer100g,
			&food.CarbsPer100g, &food.FatPer100g, &food.FiberPer100g,
			&food.CreatedBy, &food.CreatedAt)
		if err != nil {
			return nil, err
		}
		foods = append(foods, food)
	}
	return foods, nil
}

func (r *PostgresRepository) GetFoodByID(id int) (*Food, error) {
	var food Food
	err := r.db.QueryRow(`
		SELECT id, name, calories_per_100g, protein_per_100g, carbs_per_100g, fat_per_100g, fiber_per_100g, created_by, created_at
		FROM foods
		WHERE id = $1
	`, id).Scan(&food.ID, &food.Name, &food.CaloriesPer100g, &food.ProteinPer100g,
		&food.CarbsPer100g, &food.FatPer100g, &food.FiberPer100g,
		&food.CreatedBy, &food.CreatedAt)

	if err != nil {
		return nil, err
	}
	return &food, nil
}

func (r *PostgresRepository) CreateRecipe(recipe *Recipe) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.QueryRow(`
		INSERT INTO recipes (user_id, name, description, instructions, servings, prep_time, cook_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, recipe.UserID, recipe.Name, recipe.Description, recipe.Instructions, recipe.Servings, recipe.PrepTime, recipe.CookTime).Scan(&recipe.ID)

	if err != nil {
		return err
	}

	for i := range recipe.Ingredients {
		ingredient := &recipe.Ingredients[i]
		ingredient.RecipeID = recipe.ID

		err = tx.QueryRow(`
			INSERT INTO recipe_ingredients (recipe_id, food_id, quantity, unit)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, ingredient.RecipeID, ingredient.FoodID, ingredient.Quantity, ingredient.Unit).Scan(&ingredient.ID)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PostgresRepository) CreateRecipeIngredient(ingredient *RecipeIngredient) error {
	err := r.db.QueryRow(`
		INSERT INTO recipe_ingredients (recipe_id, food_id, quantity, unit)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, ingredient.RecipeID, ingredient.FoodID, ingredient.Quantity, ingredient.Unit).Scan(&ingredient.ID)
	return err
}

func (r *PostgresRepository) GetRecipesByUserID(userID string) ([]Recipe, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, name, description, instructions, servings, prep_time, cook_time, created_at, updated_at
		FROM recipes
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipes []Recipe
	for rows.Next() {
		var recipe Recipe
		err := rows.Scan(&recipe.ID, &recipe.UserID, &recipe.Name, &recipe.Description,
			&recipe.Instructions, &recipe.Servings, &recipe.PrepTime,
			&recipe.CookTime, &recipe.CreatedAt, &recipe.UpdatedAt)
		if err != nil {
			return nil, err
		}
		recipes = append(recipes, recipe)
	}
	return recipes, nil
}

func (r *PostgresRepository) GetRecipeIngredients(recipeID int) ([]RecipeIngredient, error) {
	rows, err := r.db.Query(`
		SELECT ri.id, ri.recipe_id, ri.food_id, ri.quantity, ri.unit,
			   f.name, f.calories_per_100g, f.protein_per_100g, f.carbs_per_100g, f.fat_per_100g, f.fiber_per_100g
		FROM recipe_ingredients ri
		JOIN foods f ON ri.food_id = f.id
		WHERE ri.recipe_id = $1
	`, recipeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ingredients []RecipeIngredient
	for rows.Next() {
		var ingredient RecipeIngredient
		var food Food
		err := rows.Scan(&ingredient.ID, &ingredient.RecipeID, &ingredient.FoodID,
			&ingredient.Quantity, &ingredient.Unit,
			&food.Name, &food.CaloriesPer100g, &food.ProteinPer100g,
			&food.CarbsPer100g, &food.FatPer100g, &food.FiberPer100g)
		if err != nil {
			return nil, err
		}
		food.ID = ingredient.FoodID
		ingredient.Food = &food
		ingredients = append(ingredients, ingredient)
	}
	return ingredients, nil
}

func (r *PostgresRepository) CreateNutritionGoals(goals *Goals) error {
	err := r.db.QueryRow(`
		INSERT INTO nutrition_goals (user_id, daily_calories, daily_protein, daily_carbs, daily_fat, daily_fiber)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, goals.UserID, goals.DailyCalories, goals.DailyProtein, goals.DailyCarbs, goals.DailyFat, goals.DailyFiber).Scan(&goals.ID)
	return err
}

func (r *PostgresRepository) UpdateNutritionGoals(userID string, req *UpdateNutritionGoalsRequest) error {
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

	_, err := r.db.Exec(query, args...)
	return err
}

func (r *PostgresRepository) GetNutritionGoalsByUserID(userID string) (*Goals, error) {
	var goals Goals
	err := r.db.QueryRow(`
		SELECT id, user_id, daily_calories, daily_protein, daily_carbs, daily_fat, daily_fiber, updated_at
		FROM nutrition_goals
		WHERE user_id = $1
	`, userID).Scan(&goals.ID, &goals.UserID, &goals.DailyCalories, &goals.DailyProtein,
		&goals.DailyCarbs, &goals.DailyFat, &goals.DailyFiber, &goals.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return &goals, nil
}

func (r *PostgresRepository) CheckNutritionGoalsExist(userID string) (int, error) {
	var id int
	err := r.db.QueryRow("SELECT id FROM nutrition_goals WHERE user_id = $1", userID).Scan(&id)
	return id, err
}

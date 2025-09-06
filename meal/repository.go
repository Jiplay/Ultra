package meal

import (
	"database/sql"
	"strconv"
	"strings"
	"time"
)

type Repository interface {
	// Meal operations
	CreateMeal(meal *Meal) error
	GetMealsByUserID(userID string, startDate, endDate time.Time) ([]Meal, error)
	GetMealByID(id int, userID string) (*Meal, error)
	UpdateMeal(id int, userID string, req *UpdateMealRequest) error
	DeleteMeal(id int, userID string) error
	CheckMealExists(id int, userID string) (bool, error)

	// Meal item operations
	AddMealItem(mealID int, item *MealItem) error
	GetMealItems(mealID int) ([]MealItem, error)
	UpdateMealItem(itemID int, req *UpdateMealItemRequest) error
	DeleteMealItem(itemID int, mealID int) error

	// Food and Recipe references
	GetFoodByID(id int) (*Food, error)
	GetRecipeByID(id int) (*Recipe, error)
	GetRecipeIngredients(recipeID int) ([]RecipeIngredient, error)

	// Analytics
	GetMealSummary(userID string, date time.Time) (*MealSummary, error)
	GetMealPlan(userID string, startDate, endDate time.Time) (*MealPlan, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) Repository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateMeal(meal *Meal) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Create meal
	err = tx.QueryRow(`
		INSERT INTO meals (user_id, name, meal_type, date, notes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`, meal.UserID, meal.Name, meal.MealType, meal.Date, meal.Notes).Scan(
		&meal.ID, &meal.CreatedAt, &meal.UpdatedAt)

	if err != nil {
		return err
	}

	// Create meal items
	for i := range meal.MealItems {
		item := &meal.MealItems[i]
		item.MealID = meal.ID

		err = tx.QueryRow(`
			INSERT INTO meal_items (meal_id, item_type, item_id, quantity, notes)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
		`, item.MealID, item.ItemType, item.ItemID, item.Quantity, item.Notes).Scan(&item.ID)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PostgresRepository) GetMealsByUserID(userID string, startDate, endDate time.Time) ([]Meal, error) {
	query := `
		SELECT id, user_id, name, meal_type, date, notes, created_at, updated_at
		FROM meals
		WHERE user_id = $1
	`
	args := []interface{}{userID}

	if !startDate.IsZero() && !endDate.IsZero() {
		query += " AND date >= $2 AND date <= $3"
		args = append(args, startDate, endDate)
	} else if !startDate.IsZero() {
		query += " AND date >= $2"
		args = append(args, startDate)
	} else if !endDate.IsZero() {
		query += " AND date <= $2"
		args = append(args, endDate)
	}

	query += " ORDER BY date DESC, meal_type"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var meals []Meal
	for rows.Next() {
		var meal Meal
		err := rows.Scan(&meal.ID, &meal.UserID, &meal.Name, &meal.MealType,
			&meal.Date, &meal.Notes, &meal.CreatedAt, &meal.UpdatedAt)
		if err != nil {
			return nil, err
		}

		// Get meal items for each meal
		items, err := r.GetMealItems(meal.ID)
		if err != nil {
			return nil, err
		}
		meal.MealItems = items

		// Calculate nutritional totals
		r.calculateMealNutrition(&meal)

		meals = append(meals, meal)
	}
	return meals, nil
}

func (r *PostgresRepository) GetMealByID(id int, userID string) (*Meal, error) {
	var meal Meal
	err := r.db.QueryRow(`
		SELECT id, user_id, name, meal_type, date, notes, created_at, updated_at
		FROM meals
		WHERE id = $1 AND user_id = $2
	`, id, userID).Scan(&meal.ID, &meal.UserID, &meal.Name, &meal.MealType,
		&meal.Date, &meal.Notes, &meal.CreatedAt, &meal.UpdatedAt)

	if err != nil {
		return nil, err
	}

	// Get meal items
	items, err := r.GetMealItems(meal.ID)
	if err != nil {
		return nil, err
	}
	meal.MealItems = items

	// Calculate nutritional totals
	r.calculateMealNutrition(&meal)

	return &meal, nil
}

func (r *PostgresRepository) UpdateMeal(id int, userID string, req *UpdateMealRequest) error {
	query := "UPDATE meals SET updated_at = CURRENT_TIMESTAMP"
	args := []interface{}{}
	argCount := 0

	if req.Name != nil {
		argCount++
		query += ", name = $" + strconv.Itoa(argCount)
		args = append(args, *req.Name)
	}
	if req.MealType != nil {
		argCount++
		query += ", meal_type = $" + strconv.Itoa(argCount)
		args = append(args, *req.MealType)
	}
	if req.Date != nil {
		argCount++
		query += ", date = $" + strconv.Itoa(argCount)
		args = append(args, *req.Date)
	}
	if req.Notes != nil {
		argCount++
		query += ", notes = $" + strconv.Itoa(argCount)
		args = append(args, *req.Notes)
	}

	argCount++
	query += " WHERE id = $" + strconv.Itoa(argCount)
	args = append(args, id)

	argCount++
	query += " AND user_id = $" + strconv.Itoa(argCount)
	args = append(args, userID)

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *PostgresRepository) DeleteMeal(id int, userID string) error {
	result, err := r.db.Exec(`
		DELETE FROM meals
		WHERE id = $1 AND user_id = $2
	`, id, userID)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *PostgresRepository) CheckMealExists(id int, userID string) (bool, error) {
	var exists int
	err := r.db.QueryRow("SELECT 1 FROM meals WHERE id = $1 AND user_id = $2", id, userID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *PostgresRepository) AddMealItem(mealID int, item *MealItem) error {
	err := r.db.QueryRow(`
		INSERT INTO meal_items (meal_id, item_type, item_id, quantity, notes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, mealID, item.ItemType, item.ItemID, item.Quantity, item.Notes).Scan(&item.ID)

	if err != nil {
		return err
	}

	item.MealID = mealID
	return nil
}

func (r *PostgresRepository) GetMealItems(mealID int) ([]MealItem, error) {
	rows, err := r.db.Query(`
		SELECT id, meal_id, item_type, item_id, quantity, notes
		FROM meal_items
		WHERE meal_id = $1
		ORDER BY id
	`, mealID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []MealItem
	for rows.Next() {
		var item MealItem
		err := rows.Scan(&item.ID, &item.MealID, &item.ItemType, &item.ItemID, &item.Quantity, &item.Notes)
		if err != nil {
			return nil, err
		}

		// Load food or recipe data based on item type
		if item.ItemType == "food" {
			food, err := r.GetFoodByID(item.ItemID)
			if err == nil {
				item.Food = food
				// Calculate nutrition for this item
				r.calculateFoodItemNutrition(&item)
			}
		} else if item.ItemType == "recipe" {
			recipe, err := r.GetRecipeByID(item.ItemID)
			if err == nil {
				item.Recipe = recipe
				// Calculate nutrition for this item
				r.calculateRecipeItemNutrition(&item)
			}
		}

		items = append(items, item)
	}
	return items, nil
}

func (r *PostgresRepository) UpdateMealItem(itemID int, req *UpdateMealItemRequest) error {
	query := "UPDATE meal_items SET"
	args := []interface{}{}
	argCount := 0
	updates := []string{}

	if req.Quantity != nil {
		argCount++
		updates = append(updates, " quantity = $"+strconv.Itoa(argCount))
		args = append(args, *req.Quantity)
	}
	if req.Notes != nil {
		argCount++
		updates = append(updates, " notes = $"+strconv.Itoa(argCount))
		args = append(args, *req.Notes)
	}

	if len(updates) == 0 {
		return nil // Nothing to update
	}

	query += strings.Join(updates, ",")
	argCount++
	query += " WHERE id = $" + strconv.Itoa(argCount)
	args = append(args, itemID)

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *PostgresRepository) DeleteMealItem(itemID int, mealID int) error {
	result, err := r.db.Exec(`
		DELETE FROM meal_items
		WHERE id = $1 AND meal_id = $2
	`, itemID, mealID)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *PostgresRepository) GetFoodByID(id int) (*Food, error) {
	var food Food
	err := r.db.QueryRow(`
		SELECT id, name, calories_per_100g, protein_per_100g, carbs_per_100g, fat_per_100g, fiber_per_100g
		FROM foods
		WHERE id = $1
	`, id).Scan(&food.ID, &food.Name, &food.CaloriesPer100g, &food.ProteinPer100g,
		&food.CarbsPer100g, &food.FatPer100g, &food.FiberPer100g)

	if err != nil {
		return nil, err
	}
	return &food, nil
}

func (r *PostgresRepository) GetRecipeByID(id int) (*Recipe, error) {
	var recipe Recipe
	err := r.db.QueryRow(`
		SELECT id, name, servings
		FROM recipes
		WHERE id = $1
	`, id).Scan(&recipe.ID, &recipe.Name, &recipe.Servings)

	if err != nil {
		return nil, err
	}

	// Get recipe ingredients
	ingredients, err := r.GetRecipeIngredients(id)
	if err != nil {
		return nil, err
	}
	recipe.Ingredients = ingredients

	// Calculate nutrition per serving
	r.calculateRecipeNutrition(&recipe)

	return &recipe, nil
}

func (r *PostgresRepository) GetRecipeIngredients(recipeID int) ([]RecipeIngredient, error) {
	rows, err := r.db.Query(`
		SELECT ri.id, ri.food_id, ri.quantity, ri.unit,
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
		err := rows.Scan(&ingredient.ID, &ingredient.FoodID, &ingredient.Quantity, &ingredient.Unit,
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

func (r *PostgresRepository) GetMealSummary(userID string, date time.Time) (*MealSummary, error) {
	// Get meals for the specific date
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	meals, err := r.GetMealsByUserID(userID, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}

	summary := &MealSummary{
		Date:        date,
		MealsByType: make(map[MealType]int),
	}

	for _, meal := range meals {
		summary.TotalMeals++
		summary.TotalCalories += meal.TotalCalories
		summary.TotalProtein += meal.TotalProtein
		summary.TotalCarbs += meal.TotalCarbs
		summary.TotalFat += meal.TotalFat
		summary.TotalFiber += meal.TotalFiber
		summary.MealsByType[meal.MealType]++
	}

	return summary, nil
}

func (r *PostgresRepository) GetMealPlan(userID string, startDate, endDate time.Time) (*MealPlan, error) {
	meals, err := r.GetMealsByUserID(userID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	plan := &MealPlan{
		UserID:    userID,
		StartDate: startDate,
		EndDate:   endDate,
		Meals:     meals,
	}

	// Calculate overall summary
	plan.Summary.Date = startDate
	plan.Summary.MealsByType = make(map[MealType]int)

	for _, meal := range meals {
		plan.Summary.TotalMeals++
		plan.Summary.TotalCalories += meal.TotalCalories
		plan.Summary.TotalProtein += meal.TotalProtein
		plan.Summary.TotalCarbs += meal.TotalCarbs
		plan.Summary.TotalFat += meal.TotalFat
		plan.Summary.TotalFiber += meal.TotalFiber
		plan.Summary.MealsByType[meal.MealType]++
	}

	return plan, nil
}

// Helper functions for nutrition calculations
func (r *PostgresRepository) calculateMealNutrition(meal *Meal) {
	meal.TotalCalories = 0
	meal.TotalProtein = 0
	meal.TotalCarbs = 0
	meal.TotalFat = 0
	meal.TotalFiber = 0

	for _, item := range meal.MealItems {
		meal.TotalCalories += item.Calories
		meal.TotalProtein += item.Protein
		meal.TotalCarbs += item.Carbs
		meal.TotalFat += item.Fat
		meal.TotalFiber += item.Fiber
	}
}

func (r *PostgresRepository) calculateFoodItemNutrition(item *MealItem) {
	if item.Food == nil {
		return
	}

	// Calculate nutrition based on quantity (assuming quantity is in grams)
	multiplier := item.Quantity / 100.0

	item.Calories = int(float64(item.Food.CaloriesPer100g) * multiplier)
	item.Protein = item.Food.ProteinPer100g * multiplier
	item.Carbs = item.Food.CarbsPer100g * multiplier
	item.Fat = item.Food.FatPer100g * multiplier
	item.Fiber = item.Food.FiberPer100g * multiplier
}

func (r *PostgresRepository) calculateRecipeItemNutrition(item *MealItem) {
	if item.Recipe == nil {
		return
	}

	// Calculate nutrition based on servings (quantity represents number of servings)
	item.Calories = int(float64(item.Recipe.CaloriesPerServing) * item.Quantity)
	item.Protein = item.Recipe.ProteinPerServing * item.Quantity
	item.Carbs = item.Recipe.CarbsPerServing * item.Quantity
	item.Fat = item.Recipe.FatPerServing * item.Quantity
	item.Fiber = item.Recipe.FiberPerServing * item.Quantity
}

func (r *PostgresRepository) calculateRecipeNutrition(recipe *Recipe) {
	if recipe.Servings == 0 || len(recipe.Ingredients) == 0 {
		return
	}

	totalCalories := 0
	totalProtein := 0.0
	totalCarbs := 0.0
	totalFat := 0.0
	totalFiber := 0.0

	for _, ingredient := range recipe.Ingredients {
		if ingredient.Food == nil {
			continue
		}

		// Convert ingredient quantity to grams (assuming unit conversion)
		quantityInGrams := ingredient.Quantity
		if ingredient.Unit != "g" {
			// Simple unit conversion - in real app, you'd have a proper converter
			switch ingredient.Unit {
			case "kg":
				quantityInGrams *= 1000
			case "cup":
				quantityInGrams *= 240 // approximate
			case "tbsp":
				quantityInGrams *= 15
			case "tsp":
				quantityInGrams *= 5
			}
		}

		multiplier := quantityInGrams / 100.0
		totalCalories += int(float64(ingredient.Food.CaloriesPer100g) * multiplier)
		totalProtein += ingredient.Food.ProteinPer100g * multiplier
		totalCarbs += ingredient.Food.CarbsPer100g * multiplier
		totalFat += ingredient.Food.FatPer100g * multiplier
		totalFiber += ingredient.Food.FiberPer100g * multiplier
	}

	// Divide by servings to get per-serving values
	servingsFloat := float64(recipe.Servings)
	recipe.CaloriesPerServing = int(float64(totalCalories) / servingsFloat)
	recipe.ProteinPerServing = totalProtein / servingsFloat
	recipe.CarbsPerServing = totalCarbs / servingsFloat
	recipe.FatPerServing = totalFat / servingsFloat
	recipe.FiberPerServing = totalFiber / servingsFloat
}

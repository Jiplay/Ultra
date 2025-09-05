package nutrition

import "time"

type Food struct {
	ID              int       `json:"id" db:"id"`
	Name            string    `json:"name" db:"name"`
	CaloriesPer100g int       `json:"calories_per_100g" db:"calories_per_100g"`
	ProteinPer100g  float64   `json:"protein_per_100g" db:"protein_per_100g"`
	CarbsPer100g    float64   `json:"carbs_per_100g" db:"carbs_per_100g"`
	FatPer100g      float64   `json:"fat_per_100g" db:"fat_per_100g"`
	FiberPer100g    float64   `json:"fiber_per_100g" db:"fiber_per_100g"`
	CreatedBy       string    `json:"created_by" db:"created_by"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type Recipe struct {
	ID           int                `json:"id" db:"id"`
	UserID       string             `json:"user_id" db:"user_id"`
	Name         string             `json:"name" db:"name"`
	Description  string             `json:"description" db:"description"`
	Instructions string             `json:"instructions" db:"instructions"`
	Servings     int                `json:"servings" db:"servings"`
	PrepTime     int                `json:"prep_time" db:"prep_time"`
	CookTime     int                `json:"cook_time" db:"cook_time"`
	CreatedAt    time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at" db:"updated_at"`
	Ingredients  []RecipeIngredient `json:"ingredients,omitempty"`
}

type RecipeIngredient struct {
	ID       int     `json:"id" db:"id"`
	RecipeID int     `json:"recipe_id" db:"recipe_id"`
	FoodID   int     `json:"food_id" db:"food_id"`
	Quantity float64 `json:"quantity" db:"quantity"`
	Unit     string  `json:"unit" db:"unit"`
	Food     *Food   `json:"food,omitempty"`
}

type Goals struct {
	ID            int       `json:"id" db:"id"`
	UserID        string    `json:"user_id" db:"user_id"`
	DailyCalories int       `json:"daily_calories" db:"daily_calories"`
	DailyProtein  float64   `json:"daily_protein" db:"daily_protein"`
	DailyCarbs    float64   `json:"daily_carbs" db:"daily_carbs"`
	DailyFat      float64   `json:"daily_fat" db:"daily_fat"`
	DailyFiber    float64   `json:"daily_fiber" db:"daily_fiber"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type CreateFoodRequest struct {
	Name            string  `json:"name" validate:"required"`
	CaloriesPer100g int     `json:"calories_per_100g" validate:"required"`
	ProteinPer100g  float64 `json:"protein_per_100g"`
	CarbsPer100g    float64 `json:"carbs_per_100g"`
	FatPer100g      float64 `json:"fat_per_100g"`
	FiberPer100g    float64 `json:"fiber_per_100g"`
}

type CreateRecipeRequest struct {
	Name         string                   `json:"name" validate:"required"`
	Description  string                   `json:"description"`
	Instructions string                   `json:"instructions"`
	Servings     int                      `json:"servings" validate:"required"`
	PrepTime     int                      `json:"prep_time"`
	CookTime     int                      `json:"cook_time"`
	Ingredients  []CreateRecipeIngredient `json:"ingredients"`
}

type CreateRecipeIngredient struct {
	FoodID   int     `json:"food_id" validate:"required"`
	Quantity float64 `json:"quantity" validate:"required"`
	Unit     string  `json:"unit" validate:"required"`
}

type UpdateNutritionGoalsRequest struct {
	DailyCalories *int     `json:"daily_calories,omitempty"`
	DailyProtein  *float64 `json:"daily_protein,omitempty"`
	DailyCarbs    *float64 `json:"daily_carbs,omitempty"`
	DailyFat      *float64 `json:"daily_fat,omitempty"`
	DailyFiber    *float64 `json:"daily_fiber,omitempty"`
}

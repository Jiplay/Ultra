package meal

import "time"

type MealType string

const (
	MealTypeBreakfast MealType = "breakfast"
	MealTypeLunch     MealType = "lunch"
	MealTypeDinner    MealType = "dinner"
	MealTypeSnack     MealType = "snack"
)

type Meal struct {
	ID        int       `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Name      string    `json:"name" db:"name"`
	MealType  MealType  `json:"meal_type" db:"meal_type"`
	Date      time.Time `json:"date" db:"date"`
	Notes     string    `json:"notes" db:"notes"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	
	// Relationships
	MealItems []MealItem `json:"meal_items,omitempty"`
	
	// Calculated fields
	TotalCalories    int     `json:"total_calories,omitempty"`
	TotalProtein     float64 `json:"total_protein,omitempty"`
	TotalCarbs       float64 `json:"total_carbs,omitempty"`
	TotalFat         float64 `json:"total_fat,omitempty"`
	TotalFiber       float64 `json:"total_fiber,omitempty"`
}

type MealItem struct {
	ID       int     `json:"id" db:"id"`
	MealID   int     `json:"meal_id" db:"meal_id"`
	ItemType string  `json:"item_type" db:"item_type"` // "food" or "recipe"
	ItemID   int     `json:"item_id" db:"item_id"`     // food_id or recipe_id
	Quantity float64 `json:"quantity" db:"quantity"`   // in grams for food, servings for recipe
	Notes    string  `json:"notes" db:"notes"`
	
	// Relationships (populated based on item_type)
	Food   *Food   `json:"food,omitempty"`
	Recipe *Recipe `json:"recipe,omitempty"`
	
	// Calculated nutritional values for this item
	Calories int     `json:"calories,omitempty"`
	Protein  float64 `json:"protein,omitempty"`
	Carbs    float64 `json:"carbs,omitempty"`
	Fat      float64 `json:"fat,omitempty"`
	Fiber    float64 `json:"fiber,omitempty"`
}

// Simplified Food struct (from nutrition package)
type Food struct {
	ID              int     `json:"id"`
	Name            string  `json:"name"`
	CaloriesPer100g int     `json:"calories_per_100g"`
	ProteinPer100g  float64 `json:"protein_per_100g"`
	CarbsPer100g    float64 `json:"carbs_per_100g"`
	FatPer100g      float64 `json:"fat_per_100g"`
	FiberPer100g    float64 `json:"fiber_per_100g"`
}

// Simplified Recipe struct (from nutrition package)
type Recipe struct {
	ID           int                `json:"id"`
	Name         string             `json:"name"`
	Servings     int                `json:"servings"`
	Ingredients  []RecipeIngredient `json:"ingredients,omitempty"`
	
	// Calculated nutritional values per serving
	CaloriesPerServing int     `json:"calories_per_serving,omitempty"`
	ProteinPerServing  float64 `json:"protein_per_serving,omitempty"`
	CarbsPerServing    float64 `json:"carbs_per_serving,omitempty"`
	FatPerServing      float64 `json:"fat_per_serving,omitempty"`
	FiberPerServing    float64 `json:"fiber_per_serving,omitempty"`
}

type RecipeIngredient struct {
	ID       int     `json:"id"`
	FoodID   int     `json:"food_id"`
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
	Food     *Food   `json:"food,omitempty"`
}

// Request/Response DTOs
type CreateMealRequest struct {
	Name     string    `json:"name" validate:"required"`
	MealType MealType  `json:"meal_type" validate:"required"`
	Date     time.Time `json:"date" validate:"required"`
	Notes    string    `json:"notes"`
	Items    []CreateMealItemRequest `json:"items"`
}

type CreateMealItemRequest struct {
	ItemType string  `json:"item_type" validate:"required,oneof=food recipe"`
	ItemID   int     `json:"item_id" validate:"required"`
	Quantity float64 `json:"quantity" validate:"required,gt=0"`
	Notes    string  `json:"notes"`
}

type UpdateMealRequest struct {
	Name     *string   `json:"name,omitempty"`
	MealType *MealType `json:"meal_type,omitempty"`
	Date     *time.Time `json:"date,omitempty"`
	Notes    *string   `json:"notes,omitempty"`
}

type AddMealItemRequest struct {
	ItemType string  `json:"item_type" validate:"required,oneof=food recipe"`
	ItemID   int     `json:"item_id" validate:"required"`
	Quantity float64 `json:"quantity" validate:"required,gt=0"`
	Notes    string  `json:"notes"`
}

type UpdateMealItemRequest struct {
	Quantity *float64 `json:"quantity,omitempty"`
	Notes    *string  `json:"notes,omitempty"`
}

type MealSummary struct {
	Date          time.Time `json:"date"`
	TotalMeals    int       `json:"total_meals"`
	TotalCalories int       `json:"total_calories"`
	TotalProtein  float64   `json:"total_protein"`
	TotalCarbs    float64   `json:"total_carbs"`
	TotalFat      float64   `json:"total_fat"`
	TotalFiber    float64   `json:"total_fiber"`
	MealsByType   map[MealType]int `json:"meals_by_type"`
}

type MealPlan struct {
	UserID    string    `json:"user_id"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Meals     []Meal    `json:"meals"`
	Summary   MealSummary `json:"summary"`
}
package recipe

import (
	"time"

	"gorm.io/gorm"
)

// Recipe represents a combination of foods
type Recipe struct {
	ID          uint               `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	DeletedAt   gorm.DeletedAt     `json:"deleted_at,omitempty" gorm:"index"`
	Name        string             `json:"name" gorm:"type:varchar(255);not null"`
	UserID      *uint              `json:"user_id,omitempty" gorm:"index"` // NULL = global recipe
	Tags        []string           `json:"tags,omitempty" gorm:"type:text[]"`
	Ingredients []RecipeIngredient `json:"ingredients,omitempty" gorm:"foreignKey:RecipeID;constraint:OnDelete:CASCADE"`
}

// RecipeIngredient represents a food item within a recipe
type RecipeIngredient struct {
	ID            uint           `json:"id" gorm:"primarykey"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	RecipeID      uint           `json:"recipe_id" gorm:"not null;index"`
	FoodID        uint           `json:"food_id" gorm:"not null;index"`
	QuantityGrams float64        `json:"quantity_grams" gorm:"type:decimal(10,2);not null"` // Amount in grams
}

// CreateRecipeRequest represents the request to create a recipe
type CreateRecipeRequest struct {
	Name        string                    `json:"name"`
	Tags        []string                  `json:"tags,omitempty"`
	Ingredients []CreateIngredientRequest `json:"ingredients"`
}

// CreateIngredientRequest represents an ingredient in the create recipe request
type CreateIngredientRequest struct {
	FoodID        uint    `json:"food_id"`
	QuantityGrams float64 `json:"quantity_grams"`
}

// UpdateRecipeRequest represents the request to update a recipe
type UpdateRecipeRequest struct {
	Name string   `json:"name"`
	Tags []string `json:"tags,omitempty"`
}

// AddIngredientRequest represents the request to add an ingredient to a recipe
type AddIngredientRequest struct {
	FoodID        uint    `json:"food_id"`
	QuantityGrams float64 `json:"quantity_grams"`
}

// UpdateIngredientRequest represents the request to update an ingredient quantity
type UpdateIngredientRequest struct {
	QuantityGrams float64 `json:"quantity_grams"`
}

// RecipeWithNutrition represents a recipe with calculated nutrition information
type RecipeWithNutrition struct {
	Recipe
	TotalWeight       float64 `json:"total_weight"`        // Total weight in grams
	TotalCalories     float64 `json:"total_calories"`      // Total nutrition for entire recipe
	TotalProtein      float64 `json:"total_protein"`
	TotalCarbs        float64 `json:"total_carbs"`
	TotalFat          float64 `json:"total_fat"`
	TotalFiber        float64 `json:"total_fiber"`
	CaloriesPer100g   float64 `json:"calories_per_100g"`   // Nutrition per 100g
	ProteinPer100g    float64 `json:"protein_per_100g"`
	CarbsPer100g      float64 `json:"carbs_per_100g"`
	FatPer100g        float64 `json:"fat_per_100g"`
	FiberPer100g      float64 `json:"fiber_per_100g"`
}

// IngredientWithDetails represents an ingredient with food details and calculated nutrition
type IngredientWithDetails struct {
	ID            uint    `json:"id"`
	FoodID        uint    `json:"food_id"`
	FoodName      string  `json:"food_name"`
	QuantityGrams float64 `json:"quantity_grams"`
	Calories      float64 `json:"calories"`
	Protein       float64 `json:"protein"`
	Carbs         float64 `json:"carbs"`
	Fat           float64 `json:"fat"`
	Fiber         float64 `json:"fiber"`
}

// RecipeListResponse represents a recipe with nutrition for list endpoints
type RecipeListResponse struct {
	ID              uint                    `json:"id"`
	CreatedAt       time.Time               `json:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
	Name            string                  `json:"name"`
	UserID          *uint                   `json:"user_id,omitempty"`
	Tags            []string                `json:"tags,omitempty"`
	TotalWeight     float64                 `json:"total_weight"`
	TotalCalories   float64                 `json:"total_calories"`
	TotalProtein    float64                 `json:"total_protein"`
	TotalCarbs      float64                 `json:"total_carbs"`
	TotalFat        float64                 `json:"total_fat"`
	TotalFiber      float64                 `json:"total_fiber"`
	CaloriesPer100g float64                 `json:"calories_per_100g"`
	ProteinPer100g  float64                 `json:"protein_per_100g"`
	CarbsPer100g    float64                 `json:"carbs_per_100g"`
	FatPer100g      float64                 `json:"fat_per_100g"`
	FiberPer100g    float64                 `json:"fiber_per_100g"`
	Ingredients     []IngredientWithDetails `json:"ingredients"`
}

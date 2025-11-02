package recipe

import (
	"time"

	"gorm.io/gorm"
)

// Recipe represents a combination of foods
type Recipe struct {
	ID          uint              `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	DeletedAt   gorm.DeletedAt    `json:"deleted_at,omitempty" gorm:"index"`
	Name        string            `json:"name" gorm:"type:varchar(255);not null"`
	ServingSize float64           `json:"serving_size" gorm:"type:decimal(10,2);default:1"`
	UserID      *uint             `json:"user_id,omitempty" gorm:"index"` // NULL = global recipe
	Ingredients []RecipeIngredient `json:"ingredients,omitempty" gorm:"foreignKey:RecipeID;constraint:OnDelete:CASCADE"`
}

// RecipeIngredient represents a food item within a recipe
type RecipeIngredient struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	RecipeID  uint           `json:"recipe_id" gorm:"not null;index"`
	FoodID    uint           `json:"food_id" gorm:"not null;index"`
	Quantity  float64        `json:"quantity" gorm:"type:decimal(10,2);not null"` // Amount of this food in the recipe
}

// CreateRecipeRequest represents the request to create a recipe
type CreateRecipeRequest struct {
	Name        string                      `json:"name"`
	ServingSize float64                     `json:"serving_size"`
	Ingredients []CreateIngredientRequest   `json:"ingredients"`
}

// CreateIngredientRequest represents an ingredient in the create recipe request
type CreateIngredientRequest struct {
	FoodID   uint    `json:"food_id"`
	Quantity float64 `json:"quantity"`
}

// UpdateRecipeRequest represents the request to update a recipe
type UpdateRecipeRequest struct {
	Name        string  `json:"name"`
	ServingSize float64 `json:"serving_size"`
}

// AddIngredientRequest represents the request to add an ingredient to a recipe
type AddIngredientRequest struct {
	FoodID   uint    `json:"food_id"`
	Quantity float64 `json:"quantity"`
}

// UpdateIngredientRequest represents the request to update an ingredient quantity
type UpdateIngredientRequest struct {
	Quantity float64 `json:"quantity"`
}

// RecipeWithNutrition represents a recipe with calculated nutrition information
type RecipeWithNutrition struct {
	Recipe
	Calories float64 `json:"calories"`
	Protein  float64 `json:"protein"`
	Carbs    float64 `json:"carbs"`
	Fat      float64 `json:"fat"`
	Fiber    float64 `json:"fiber"`
}

// IngredientWithDetails represents an ingredient with food details and calculated nutrition
type IngredientWithDetails struct {
	ID       uint    `json:"id"`
	FoodID   uint    `json:"food_id"`
	FoodName string  `json:"food_name"`
	Quantity float64 `json:"quantity"`
	Calories float64 `json:"calories"`
	Protein  float64 `json:"protein"`
	Carbs    float64 `json:"carbs"`
	Fat      float64 `json:"fat"`
	Fiber    float64 `json:"fiber"`
}

// RecipeListResponse represents a recipe with nutrition for list endpoints
type RecipeListResponse struct {
	ID          uint                    `json:"id"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
	Name        string                  `json:"name"`
	ServingSize float64                 `json:"serving_size"`
	UserID      *uint                   `json:"user_id,omitempty"`
	Calories    float64                 `json:"calories"`
	Protein     float64                 `json:"protein"`
	Carbs       float64                 `json:"carbs"`
	Fat         float64                 `json:"fat"`
	Fiber       float64                 `json:"fiber"`
	Ingredients []IngredientWithDetails `json:"ingredients"`
}

package diary

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// MealType represents the type of meal
type MealType string

const (
	Breakfast MealType = "breakfast"
	Lunch     MealType = "lunch"
	Dinner    MealType = "dinner"
	Snack     MealType = "snack"
)

// CustomIngredient represents an ingredient with custom quantity in a diary entry
type CustomIngredient struct {
	FoodID        uint    `json:"food_id"`
	FoodName      string  `json:"food_name,omitempty"`
	QuantityGrams float64 `json:"quantity_grams"`
	Calories      float64 `json:"calories"`
	Protein       float64 `json:"protein"`
	Carbs         float64 `json:"carbs"`
	Fat           float64 `json:"fat"`
	Fiber         float64 `json:"fiber"`
}

// CustomIngredients is a custom type for JSONB storage
type CustomIngredients []CustomIngredient

// Value implements the driver.Valuer interface for JSONB serialization
func (c CustomIngredients) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

// Scan implements the sql.Scanner interface for JSONB deserialization
func (c *CustomIngredients) Scan(value interface{}) error {
	if value == nil {
		*c = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan CustomIngredients: not a byte slice")
	}

	return json.Unmarshal(bytes, c)
}

// DiaryEntry represents a food logging entry
type DiaryEntry struct {
	ID            uint           `json:"id" gorm:"primarykey"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	UserID           uint           `json:"user_id" gorm:"not null;index:idx_user_date"`
	FoodID           *uint          `json:"food_id" gorm:"index"`
	RecipeID         *uint          `json:"recipe_id" gorm:"index"`
	InlineRecipeName *string        `json:"inline_recipe_name,omitempty" gorm:"type:varchar(255)"`

	// Inline food fields (for temporary custom foods)
	InlineFoodName        *string  `json:"inline_food_name,omitempty" gorm:"type:varchar(255)"`
	InlineFoodDescription *string  `json:"inline_food_description,omitempty" gorm:"type:text"`
	InlineFoodCalories    *float64 `json:"inline_food_calories,omitempty" gorm:"type:decimal(10,2)"`
	InlineFoodProtein     *float64 `json:"inline_food_protein,omitempty" gorm:"type:decimal(10,2)"`
	InlineFoodCarbs       *float64 `json:"inline_food_carbs,omitempty" gorm:"type:decimal(10,2)"`
	InlineFoodFat         *float64 `json:"inline_food_fat,omitempty" gorm:"type:decimal(10,2)"`
	InlineFoodFiber       *float64 `json:"inline_food_fiber,omitempty" gorm:"type:decimal(10,2)"`
	InlineFoodTag         *string  `json:"inline_food_tag,omitempty" gorm:"type:varchar(20)"`

	Date             time.Time      `json:"date" gorm:"not null;index:idx_user_date"`
	MealType      MealType       `json:"meal_type" gorm:"type:varchar(20);not null"`
	QuantityGrams float64        `json:"quantity_grams" gorm:"type:decimal(10,2);not null"` // Grams consumed
	Notes         string         `json:"notes" gorm:"type:text"`

	// Cached nutritional values (calculated at insert time)
	Calories float64 `json:"calories" gorm:"type:decimal(10,2)"`
	Protein  float64 `json:"protein" gorm:"type:decimal(10,2)"`
	Carbs    float64 `json:"carbs" gorm:"type:decimal(10,2)"`
	Fat      float64 `json:"fat" gorm:"type:decimal(10,2)"`
	Fiber    float64 `json:"fiber" gorm:"type:decimal(10,2)"`

	// Cached tag values (for historical accuracy)
	FoodTag   string `json:"food_tag,omitempty" gorm:"type:varchar(20)"`
	RecipeTag string `json:"recipe_tag,omitempty" gorm:"type:varchar(20)"`

	// Custom ingredients for recipe entries (JSONB)
	CustomIngredients CustomIngredients `json:"custom_ingredients,omitempty" gorm:"type:jsonb"`

	// Additional fields for display (not persisted)
	FoodName   string `json:"food_name,omitempty" gorm:"-"`
	RecipeName string `json:"recipe_name,omitempty" gorm:"-"`
}

// CustomIngredientRequest represents a custom ingredient quantity in the request
type CustomIngredientRequest struct {
	FoodID        uint    `json:"food_id"`
	QuantityGrams float64 `json:"quantity_grams"`
}

// CreateDiaryEntryRequest represents the request to create a diary entry
type CreateDiaryEntryRequest struct {
	FoodID            *uint                      `json:"food_id"`
	RecipeID          *uint                      `json:"recipe_id"`
	InlineRecipeName  string                     `json:"inline_recipe_name"`  // For inline/temporary recipes

	// Inline food fields (for temporary custom foods)
	InlineFoodName        string  `json:"inline_food_name"`
	InlineFoodDescription string  `json:"inline_food_description,omitempty"`
	InlineFoodCalories    float64 `json:"inline_food_calories"`
	InlineFoodProtein     float64 `json:"inline_food_protein"`
	InlineFoodCarbs       float64 `json:"inline_food_carbs"`
	InlineFoodFat         float64 `json:"inline_food_fat"`
	InlineFoodFiber       float64 `json:"inline_food_fiber"`
	InlineFoodTag         string  `json:"inline_food_tag,omitempty"`

	Date              string                     `json:"date"` // YYYY-MM-DD format
	MealType          MealType                   `json:"meal_type"`
	QuantityGrams     float64                    `json:"quantity_grams"`      // For food entries or proportional recipe scaling
	CustomIngredients []CustomIngredientRequest  `json:"custom_ingredients"`  // For custom recipe ingredient quantities
	Notes             string                     `json:"notes"`
}

// UpdateDiaryEntryRequest represents the request to update a diary entry
type UpdateDiaryEntryRequest struct {
	InlineRecipeName  *string                    `json:"inline_recipe_name,omitempty"`  // For updating inline recipe name

	// Inline food update fields
	InlineFoodName        *string  `json:"inline_food_name,omitempty"`
	InlineFoodDescription *string  `json:"inline_food_description,omitempty"`
	InlineFoodCalories    *float64 `json:"inline_food_calories,omitempty"`
	InlineFoodProtein     *float64 `json:"inline_food_protein,omitempty"`
	InlineFoodCarbs       *float64 `json:"inline_food_carbs,omitempty"`
	InlineFoodFat         *float64 `json:"inline_food_fat,omitempty"`
	InlineFoodFiber       *float64 `json:"inline_food_fiber,omitempty"`
	InlineFoodTag         *string  `json:"inline_food_tag,omitempty"`

	QuantityGrams     float64                    `json:"quantity_grams"`
	CustomIngredients []CustomIngredientRequest  `json:"custom_ingredients"`  // For updating recipe ingredient quantities
	MealType          MealType                   `json:"meal_type"`
	Notes             string                     `json:"notes"`
}

// CreateEntryFromOpenFoodFactsRequest represents creating diary entry from Open Food Facts product
type CreateEntryFromOpenFoodFactsRequest struct {
	ProductName   string   `json:"product_name"`
	Brands        string   `json:"brands"`
	Calories      float64  `json:"calories"`         // per 100g
	Protein       float64  `json:"protein"`          // per 100g
	Carbs         float64  `json:"carbs"`            // per 100g
	Fat           float64  `json:"fat"`              // per 100g
	Fiber         float64  `json:"fiber"`            // per 100g
	Date          string   `json:"date"`             // YYYY-MM-DD
	MealType      MealType `json:"meal_type"`
	QuantityGrams float64  `json:"quantity_grams"`
	Notes         string   `json:"notes"`
	Tag           string   `json:"tag"`              // Optional: "routine" or "contextual"
}

// DailySummary represents the daily nutrition summary
type DailySummary struct {
	Date         string            `json:"date"`
	TotalCalories float64          `json:"total_calories"`
	TotalProtein  float64          `json:"total_protein"`
	TotalCarbs    float64          `json:"total_carbs"`
	TotalFat      float64          `json:"total_fat"`
	TotalFiber    float64          `json:"total_fiber"`
	GoalCalories  float64          `json:"goal_calories"`
	GoalProtein   float64          `json:"goal_protein"`
	GoalCarbs     float64          `json:"goal_carbs"`
	GoalFat       float64          `json:"goal_fat"`
	GoalFiber     float64          `json:"goal_fiber"`
	Adherence     AdherencePercent `json:"adherence"`

	// Calorie breakdown by tag type
	RoutineCalories    float64 `json:"routine_calories"`
	ContextualCalories float64 `json:"contextual_calories"`
	RoutinePercent     float64 `json:"routine_percent"`     // Percentage of calories from routine foods
	ContextualPercent  float64 `json:"contextual_percent"`  // Percentage of calories from contextual foods

	Entries       []DiaryEntry     `json:"entries"`
}

// AdherencePercent represents goal adherence percentages
type AdherencePercent struct {
	Calories float64 `json:"calories"`
	Protein  float64 `json:"protein"`
	Carbs    float64 `json:"carbs"`
	Fat      float64 `json:"fat"`
	Fiber    float64 `json:"fiber"`
}

// WeeklySummary represents a weekly overview
type WeeklySummary struct {
	StartDate     string         `json:"start_date"`
	EndDate       string         `json:"end_date"`
	DailySummaries []DailySummary `json:"daily_summaries"`
	AverageCalories float64       `json:"average_calories"`
	AverageProtein  float64       `json:"average_protein"`
	AverageCarbs    float64       `json:"average_carbs"`
	AverageFat      float64       `json:"average_fat"`
	AverageFiber    float64       `json:"average_fiber"`

	// Weekly averages for tag breakdown
	AvgRoutinePercent    float64 `json:"avg_routine_percent"`
	AvgContextualPercent float64 `json:"avg_contextual_percent"`
}

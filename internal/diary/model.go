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
	UserID        uint           `json:"user_id" gorm:"not null;index:idx_user_date"`
	FoodID        *uint          `json:"food_id" gorm:"index"`
	RecipeID      *uint          `json:"recipe_id" gorm:"index"`
	Date          time.Time      `json:"date" gorm:"not null;index:idx_user_date"`
	MealType      MealType       `json:"meal_type" gorm:"type:varchar(20);not null"`
	QuantityGrams float64        `json:"quantity_grams" gorm:"type:decimal(10,2);not null"` // Grams consumed
	Notes         string         `json:"notes" gorm:"type:text"`

	// Cached nutritional values (calculated at insert time)
	Calories float64 `json:"calories" gorm:"type:decimal(10,2)"`
	Protein  float64 `json:"protein" gorm:"type:decimal(10,2)"`
	Carbs    float64 `json:"carbs" gorm:"type:decimal(10,2)"`
	Fat      float64 `json:"fat" gorm:"type:decimal(10,2)"`
	Fiber    float64 `json:"fiber" gorm:"type:decimal(10,2)"`

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
	Date              string                     `json:"date"` // YYYY-MM-DD format
	MealType          MealType                   `json:"meal_type"`
	QuantityGrams     float64                    `json:"quantity_grams"`      // For food entries or proportional recipe scaling
	CustomIngredients []CustomIngredientRequest  `json:"custom_ingredients"`  // For custom recipe ingredient quantities
	Notes             string                     `json:"notes"`
}

// UpdateDiaryEntryRequest represents the request to update a diary entry
type UpdateDiaryEntryRequest struct {
	QuantityGrams     float64                    `json:"quantity_grams"`
	CustomIngredients []CustomIngredientRequest  `json:"custom_ingredients"`  // For updating recipe ingredient quantities
	MealType          MealType                   `json:"meal_type"`
	Notes             string                     `json:"notes"`
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
}

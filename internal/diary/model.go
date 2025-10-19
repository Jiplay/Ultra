package diary

import (
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

// DiaryEntry represents a food logging entry
type DiaryEntry struct {
	ID           uint           `json:"id" gorm:"primarykey"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	UserID       uint           `json:"user_id" gorm:"not null;index:idx_user_date"`
	FoodID       *uint          `json:"food_id" gorm:"index"`
	RecipeID     *uint          `json:"recipe_id" gorm:"index"`
	Date         time.Time      `json:"date" gorm:"not null;index:idx_user_date"`
	MealType     MealType       `json:"meal_type" gorm:"type:varchar(20);not null"`
	ServingSize  float64        `json:"serving_size" gorm:"type:decimal(10,2);default:1"`
	Notes        string         `json:"notes" gorm:"type:text"`

	// Cached nutritional values (calculated at insert time)
	Calories     float64        `json:"calories" gorm:"type:decimal(10,2)"`
	Protein      float64        `json:"protein" gorm:"type:decimal(10,2)"`
	Carbs        float64        `json:"carbs" gorm:"type:decimal(10,2)"`
	Fat          float64        `json:"fat" gorm:"type:decimal(10,2)"`
	Fiber        float64        `json:"fiber" gorm:"type:decimal(10,2)"`
}

// CreateDiaryEntryRequest represents the request to create a diary entry
type CreateDiaryEntryRequest struct {
	FoodID      *uint    `json:"food_id"`
	RecipeID    *uint    `json:"recipe_id"`
	Date        string   `json:"date"` // YYYY-MM-DD format
	MealType    MealType `json:"meal_type"`
	ServingSize float64  `json:"serving_size"`
	Notes       string   `json:"notes"`
}

// UpdateDiaryEntryRequest represents the request to update a diary entry
type UpdateDiaryEntryRequest struct {
	ServingSize float64  `json:"serving_size"`
	MealType    MealType `json:"meal_type"`
	Notes       string   `json:"notes"`
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

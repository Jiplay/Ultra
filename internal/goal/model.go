package goal

import (
	"time"

	"gorm.io/gorm"
)

// NutritionGoal represents daily nutrition targets for a user
type NutritionGoal struct {
	gorm.Model
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	Calories  float64   `json:"calories" gorm:"type:decimal(10,2)"`
	Protein   float64   `json:"protein" gorm:"type:decimal(10,2)"`
	Carbs     float64   `json:"carbs" gorm:"type:decimal(10,2)"`
	Fat       float64   `json:"fat" gorm:"type:decimal(10,2)"`
	Fiber     float64   `json:"fiber" gorm:"type:decimal(10,2)"`
	StartDate time.Time `json:"start_date" gorm:"not null"`
	EndDate   *time.Time `json:"end_date"`
	IsActive  bool      `json:"is_active" gorm:"default:true;index"`
}

// CreateGoalRequest represents the request to create a nutrition goal
type CreateGoalRequest struct {
	Calories  float64    `json:"calories"`
	Protein   float64    `json:"protein"`
	Carbs     float64    `json:"carbs"`
	Fat       float64    `json:"fat"`
	Fiber     float64    `json:"fiber"`
	StartDate time.Time  `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
}

// UpdateGoalRequest represents the request to update a nutrition goal
type UpdateGoalRequest struct {
	Calories float64    `json:"calories"`
	Protein  float64    `json:"protein"`
	Carbs    float64    `json:"carbs"`
	Fat      float64    `json:"fat"`
	Fiber    float64    `json:"fiber"`
	EndDate  *time.Time `json:"end_date"`
}

// RecommendedGoalRequest represents the request to calculate recommended goals
type RecommendedGoalRequest struct {
	Weight        float64 `json:"weight"`         // in kg
	TargetWeight  float64 `json:"target_weight"`  // in kg
	WeeksToGoal   int     `json:"weeks_to_goal"`  // number of weeks to reach goal
}

// RecommendedGoalResponse represents the calculated recommended goals
type RecommendedGoalResponse struct {
	Calories float64 `json:"calories"`
	Protein  float64 `json:"protein"`
	Carbs    float64 `json:"carbs"`
	Fat      float64 `json:"fat"`
	Fiber    float64 `json:"fiber"`
	Message  string  `json:"message"`
}

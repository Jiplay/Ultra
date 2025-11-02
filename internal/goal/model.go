package goal

import (
	"encoding/json"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Date is a custom type that handles both "YYYY-MM-DD" and RFC3339 formats
type Date struct {
	time.Time
}

// UnmarshalJSON implements custom JSON unmarshaling for Date
func (d *Date) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")

	// Try parsing as date only (YYYY-MM-DD)
	t, err := time.Parse("2006-01-02", s)
	if err == nil {
		d.Time = t
		return nil
	}

	// Try parsing as RFC3339 (full timestamp)
	t, err = time.Parse(time.RFC3339, s)
	if err == nil {
		d.Time = t
		return nil
	}

	return err
}

// MarshalJSON implements custom JSON marshaling for Date
func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Time)
}

// NutritionGoal represents daily nutrition targets for a user
type NutritionGoal struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	UserID    uint           `json:"user_id" gorm:"not null;index"`
	Calories  float64        `json:"calories" gorm:"type:decimal(10,2)"`
	Protein   float64        `json:"protein" gorm:"type:decimal(10,2)"`
	Carbs     float64        `json:"carbs" gorm:"type:decimal(10,2)"`
	Fat       float64        `json:"fat" gorm:"type:decimal(10,2)"`
	Fiber     float64        `json:"fiber" gorm:"type:decimal(10,2)"`
	StartDate time.Time      `json:"start_date" gorm:"not null"`
	EndDate   *time.Time     `json:"end_date"`
	IsActive  bool           `json:"is_active" gorm:"default:true;index"`
}

// CreateGoalRequest represents the request to create a nutrition goal
type CreateGoalRequest struct {
	Calories  float64 `json:"calories"`
	Protein   float64 `json:"protein"`
	Carbs     float64 `json:"carbs"`
	Fat       float64 `json:"fat"`
	Fiber     float64 `json:"fiber"`
	StartDate Date    `json:"start_date"`
	EndDate   *Date   `json:"end_date,omitempty"`
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

// CalculateDietRequest represents the request to calculate diet goals
type CalculateDietRequest struct {
	DietModel string `json:"diet_model"` // e.g., "zeroToHero"
	Protocol  int    `json:"protocol"`   // Protocol number (1-4 for Zero to Hero)
}

// DietPhaseResponse represents a single phase of a diet protocol
type DietPhaseResponse struct {
	Phase       int     `json:"phase"`
	Calories    float64 `json:"calories"`
	Protein     float64 `json:"protein"`
	Carbs       float64 `json:"carbs"`
	Fat         float64 `json:"fat"`
	Description string  `json:"description"`
}

// CalculateDietResponse represents the response from diet calculation
type CalculateDietResponse struct {
	DietModel      string              `json:"diet_model"`
	Protocol       int                 `json:"protocol"`
	ProtocolName   string              `json:"protocol_name"`
	BMR            float64             `json:"bmr"`
	MaintenanceMMR float64             `json:"maintenance_mmr"`
	LeanMass       float64             `json:"lean_mass"`
	Phases         []DietPhaseResponse `json:"phases"`
	Message        string              `json:"message"`
}

// ProtocolInfo represents information about a diet protocol
type ProtocolInfo struct {
	Number      int    `json:"number"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AvailableDiet represents a diet model with its available protocols
type AvailableDiet struct {
	ModelName   string         `json:"model_name"`
	DisplayName string         `json:"display_name"`
	Description string         `json:"description"`
	Protocols   []ProtocolInfo `json:"protocols"`
}

// AvailableDietsResponse represents the response for available diets endpoint
type AvailableDietsResponse struct {
	Diets []AvailableDiet `json:"diets"`
}

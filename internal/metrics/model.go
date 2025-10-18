package metrics

import (
	"time"

	"gorm.io/gorm"
)

// BodyMetric represents a body measurement entry
type BodyMetric struct {
	gorm.Model
	UserID            uint      `json:"user_id" gorm:"not null;index:idx_user_date"`
	Date              time.Time `json:"date" gorm:"not null;index:idx_user_date"`
	Weight            float64   `json:"weight" gorm:"type:decimal(5,2)"`              // in kg
	BodyFatPercent    float64   `json:"body_fat_percent" gorm:"type:decimal(5,2)"`    // percentage
	MuscleMassPercent float64   `json:"muscle_mass_percent" gorm:"type:decimal(5,2)"` // percentage
	BMI               float64   `json:"bmi" gorm:"type:decimal(5,2)"`
	Notes             string    `json:"notes" gorm:"type:text"`
}

// CreateMetricRequest represents the request to create a body metric
type CreateMetricRequest struct {
	Date              string  `json:"date"` // YYYY-MM-DD format, defaults to today
	Weight            float64 `json:"weight"`
	BodyFatPercent    float64 `json:"body_fat_percent"`
	MuscleMassPercent float64 `json:"muscle_mass_percent"`
	Notes             string  `json:"notes"`
}

// TrendResponse represents trend data over a period
type TrendResponse struct {
	Period      string        `json:"period"` // "7d", "30d", "90d"
	Metrics     []BodyMetric  `json:"metrics"`
	Trend       TrendData     `json:"trend"`
}

// TrendData represents calculated trend information
type TrendData struct {
	WeightChange          float64 `json:"weight_change"`
	BodyFatChange         float64 `json:"body_fat_change"`
	MuscleMassChange      float64 `json:"muscle_mass_change"`
	AverageWeight         float64 `json:"average_weight"`
	AverageBodyFat        float64 `json:"average_body_fat"`
	AverageMuscleMass     float64 `json:"average_muscle_mass"`
}

package metrics

import (
	"time"

	"gorm.io/gorm"
)

// BodyMetric represents a daily body weight entry
type BodyMetric struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	UserID    uint           `json:"user_id" gorm:"not null;index:idx_user_date;uniqueIndex:idx_user_date_unique"`
	Date      time.Time      `json:"date" gorm:"not null;index:idx_user_date;uniqueIndex:idx_user_date_unique"`
	Weight    float64        `json:"weight" gorm:"type:decimal(5,2);not null"` // in kg
}

// CreateMetricRequest represents the request to create a body weight entry
type CreateMetricRequest struct {
	Date   string  `json:"date"`   // YYYY-MM-DD format, defaults to today
	Weight float64 `json:"weight"` // in kg, required
}

// TrendResponse represents weight trend data over a period
type TrendResponse struct {
	Period  string       `json:"period"`  // "7d", "30d", "90d"
	Metrics []BodyMetric `json:"metrics"`
	Trend   TrendData    `json:"trend"`
}

// TrendData represents calculated weight trend information
type TrendData struct {
	WeightChange  float64 `json:"weight_change"`  // Change from first to last entry (kg)
	AverageWeight float64 `json:"average_weight"` // Average weight over period (kg)
}

package food

import (
	"time"

	"gorm.io/gorm"
)

// Food represents a food item with nutritional information
// All nutritional values (Calories, Protein, Carbs, Fat, Fiber) are per 100 grams
type Food struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deletedI_at,omitempty" gorm:"index"`
	Name        string         `json:"name" gorm:"type:varchar(255);not null"`
	Description string         `json:"description" gorm:"type:text"`
	Calories    float64        `json:"calories" gorm:"type:decimal(10,2)"`
	Protein     float64        `json:"protein" gorm:"type:decimal(10,2)"`
	Carbs       float64        `json:"carbs" gorm:"type:decimal(10,2)"`
	Fat         float64        `json:"fat" gorm:"type:decimal(10,2)"`
	Fiber       float64        `json:"fiber" gorm:"type:decimal(10,2)"`
	Tag         string         `json:"tag" gorm:"type:varchar(20);not null;default:'routine'"`
}

// CreateFoodRequest represents the request body for creating a food item
type CreateFoodRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Calories    float64 `json:"calories"`
	Protein     float64 `json:"protein"`
	Carbs       float64 `json:"carbs"`
	Fat         float64 `json:"fat"`
	Fiber       float64 `json:"fiber"`
	Tag         string  `json:"tag"`
}

// UpdateFoodRequest represents the request body for updating a food item
type UpdateFoodRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Calories    float64 `json:"calories"`
	Protein     float64 `json:"protein"`
	Carbs       float64 `json:"carbs"`
	Fat         float64 `json:"fat"`
	Fiber       float64 `json:"fiber"`
	Tag         string  `json:"tag"`
}

// Tag constants
const (
	TagRoutine    = "routine"
	TagContextual = "contextual"
	TagGeneral    = "general"
)

// ValidateTag checks if tag is valid
func ValidateTag(tag string) bool {
	return tag == TagRoutine || tag == TagContextual || tag == TagGeneral
}

// GeneralFood represents a food item from the general food database
// This is a reference table of common foods, separate from user-created custom foods
// All nutritional values (Calories, Protein, Carbs, Fat, Fiber) are per 100 grams
type GeneralFood struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Name        string    `json:"name" gorm:"type:varchar(255);not null;index"`
	Description string    `json:"description" gorm:"type:text"`
	Calories    float64   `json:"calories" gorm:"type:decimal(10,2)"`
	Protein     float64   `json:"protein" gorm:"type:decimal(10,2)"`
	Carbs       float64   `json:"carbs" gorm:"type:decimal(10,2)"`
	Fat         float64   `json:"fat" gorm:"type:decimal(10,2)"`
	Fiber       float64   `json:"fiber" gorm:"type:decimal(10,2)"`
	Tag         string    `json:"tag" gorm:"type:varchar(20);not null;default:'general';index"`
}

// GeneralFoodSearchResponse represents paginated search results for general foods
type GeneralFoodSearchResponse struct {
	Count    int64         `json:"count"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
	Foods    []GeneralFood `json:"foods"`
}

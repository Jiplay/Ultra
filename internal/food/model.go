package food

import "gorm.io/gorm"

// Food represents a food item with nutritional information
type Food struct {
	gorm.Model
	Name        string  `json:"name" gorm:"type:varchar(255);not null"`
	Description string  `json:"description" gorm:"type:text"`
	Calories    float64 `json:"calories" gorm:"type:decimal(10,2)"`
	Protein     float64 `json:"protein" gorm:"type:decimal(10,2)"`
	Carbs       float64 `json:"carbs" gorm:"type:decimal(10,2)"`
	Fat         float64 `json:"fat" gorm:"type:decimal(10,2)"`
	Fiber       float64 `json:"fiber" gorm:"type:decimal(10,2)"`
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
}

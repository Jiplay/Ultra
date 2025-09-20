package food

import "time"

type Food struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Unit      string    `json:"unit"`
	Calories  int       `json:"calories"`
	Protein   float64   `json:"protein"`
	Carbs     float64   `json:"carbs"`
	Fat       float64   `json:"fat"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateFoodRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Unit        string  `json:"unit"`
	Calories    int     `json:"calories"`
	Protein     float64 `json:"protein"`
	Carbs       float64 `json:"carbs"`
	Fat         float64 `json:"fat"`
}

type UpdateFoodRequest struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Category    *string  `json:"category,omitempty"`
	Unit        *string  `json:"unit,omitempty"`
	Calories    *int     `json:"calories,omitempty"`
	Protein     *float64 `json:"protein,omitempty"`
	Carbs       *float64 `json:"carbs,omitempty"`
	Fat         *float64 `json:"fat,omitempty"`
}

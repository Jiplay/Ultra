package food

import "time"

type Food struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Category     string    `json:"category"`
	Calories     int       `json:"calories"`
	Protein      float64   `json:"protein"`
	Carbs        float64   `json:"carbs"`
	Fat          float64   `json:"fat"`
	Fiber        float64   `json:"fiber"`
	Sugar        float64   `json:"sugar"`
	Sodium       float64   `json:"sodium"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateFoodRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Calories    int     `json:"calories"`
	Protein     float64 `json:"protein"`
	Carbs       float64 `json:"carbs"`
	Fat         float64 `json:"fat"`
	Fiber       float64 `json:"fiber"`
	Sugar       float64 `json:"sugar"`
	Sodium      float64 `json:"sodium"`
}

type UpdateFoodRequest struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Category    *string  `json:"category,omitempty"`
	Calories    *int     `json:"calories,omitempty"`
	Protein     *float64 `json:"protein,omitempty"`
	Carbs       *float64 `json:"carbs,omitempty"`
	Fat         *float64 `json:"fat,omitempty"`
	Fiber       *float64 `json:"fiber,omitempty"`
	Sugar       *float64 `json:"sugar,omitempty"`
	Sodium      *float64 `json:"sodium,omitempty"`
}
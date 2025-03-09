package model

type RecipeID string

type Recipes struct {
	ID          RecipeID    `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Macros      MacroDetail `json:"macros"`
}

type MacroDetail struct {
	ID       RecipeID `json:"id"`
	Carbs    uint8    `json:"carbs"`
	Protein  uint8    `json:"protein"`
	Fat      uint8    `json:"fat"`
	Calories uint16   `json:"calories"`
}

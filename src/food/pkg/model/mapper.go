package model

import "ultra.com/gen"

// RecipeToProto allows to convert go type Recipe in Proto to send data
func RecipeToProto(r *Recipe) *gen.Recipe {
	return &gen.Recipe{
		Id:          string(r.ID),
		Name:        r.Name,
		Description: r.Description,
		Macros: &gen.MacroDetail{
			Id:       string(r.Macros.ID),
			Carbs:    uint32(r.Macros.Carbs),
			Protein:  uint32(r.Macros.Protein),
			Fat:      uint32(r.Macros.Fat),
			Calories: uint32(r.Macros.Calories),
		},
	}
}

// RecipeFromProto allows to convert proto type Recipe to Go Recipe
func RecipeFromProto(r *gen.Recipe) *Recipe {
	return &Recipe{
		ID:          RecipeID(r.Id),
		Name:        r.Name,
		Description: r.Description,
		Macros: MacroDetail{
			ID:       RecipeID(r.Macros.Id),
			Carbs:    uint8(r.Macros.Carbs),
			Protein:  uint8(r.Macros.Protein),
			Fat:      uint8(r.Macros.Fat),
			Calories: uint16(r.Macros.Calories),
		},
	}
}

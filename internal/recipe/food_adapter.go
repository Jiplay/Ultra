package recipe

import "ultra-bis/internal/food"

// FoodAdapter adapts the food repository to implement the FoodProvider interface
// This allows the recipe package to access food data without circular dependencies
type FoodAdapter struct {
	repo *food.Repository
}

// NewFoodAdapter creates a new adapter that implements FoodProvider
func NewFoodAdapter(repo *food.Repository) FoodProvider {
	return &FoodAdapter{repo: repo}
}

// GetByID retrieves a single food item by ID
func (a *FoodAdapter) GetByID(id int) (*Food, error) {
	foodItem, err := a.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return &Food{
		ID:          foodItem.ID,
		Name:        foodItem.Name,
		Description: foodItem.Description,
		Calories:    foodItem.Calories,
		Protein:     foodItem.Protein,
		Carbs:       foodItem.Carbs,
		Fat:         foodItem.Fat,
		Fiber:       foodItem.Fiber,
	}, nil
}

// GetByIDs retrieves multiple food items by their IDs in a single query
func (a *FoodAdapter) GetByIDs(ids []int) ([]*Food, error) {
	foods, err := a.repo.GetByIDs(ids)
	if err != nil {
		return nil, err
	}

	// Convert []*food.Food to []*recipe.Food
	result := make([]*Food, len(foods))
	for i, foodItem := range foods {
		result[i] = &Food{
			ID:          foodItem.ID,
			Name:        foodItem.Name,
			Description: foodItem.Description,
			Calories:    foodItem.Calories,
			Protein:     foodItem.Protein,
			Carbs:       foodItem.Carbs,
			Fat:         foodItem.Fat,
			Fiber:       foodItem.Fiber,
		}
	}

	return result, nil
}

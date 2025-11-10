package recipe

import "ultra-bis/internal/diary"

// DiaryRecipeAdapter adapts the recipe repository for use by the diary handler
type DiaryRecipeAdapter struct {
	repo *Repository
}

// NewDiaryRecipeAdapter creates a new adapter
func NewDiaryRecipeAdapter(repo *Repository) *DiaryRecipeAdapter {
	return &DiaryRecipeAdapter{repo: repo}
}

// GetByID retrieves a recipe for diary use
func (a *DiaryRecipeAdapter) GetByID(id int) (diary.Recipe, error) {
	recipe, err := a.repo.GetByID(id)
	if err != nil {
		return diary.Recipe{}, err
	}

	return diary.Recipe{
		ID:   recipe.ID,
		Name: recipe.Name,
	}, nil
}

// GetIngredients retrieves recipe ingredients for diary use
func (a *DiaryRecipeAdapter) GetIngredients(recipeID int) ([]diary.RecipeIngredient, error) {
	recipe, err := a.repo.GetByID(recipeID)
	if err != nil {
		return nil, err
	}

	ingredients := make([]diary.RecipeIngredient, len(recipe.Ingredients))
	for i, ing := range recipe.Ingredients {
		ingredients[i] = diary.RecipeIngredient{
			FoodID:        ing.FoodID,
			QuantityGrams: ing.QuantityGrams,
		}
	}

	return ingredients, nil
}

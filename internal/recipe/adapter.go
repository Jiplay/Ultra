package recipe

import (
	"context"
	"ultra-bis/internal/diary"
)

// DiaryRecipeAdapter adapts the recipe repository for use by the diary handler
type DiaryRecipeAdapter struct {
	repo    *Repository
	service *Service
}

// NewDiaryRecipeAdapter creates a new adapter
func NewDiaryRecipeAdapter(repo *Repository, service *Service) *DiaryRecipeAdapter {
	return &DiaryRecipeAdapter{
		repo:    repo,
		service: service,
	}
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
		Tag:  recipe.Tag,
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

// CreateRecipe creates a new recipe via the service layer
func (a *DiaryRecipeAdapter) CreateRecipe(userID uint, name string, tag string, ingredients []diary.RecipeIngredientRequest) (diary.RecipeCreatedResponse, error) {
	// Convert diary.RecipeIngredientRequest to recipe.CreateIngredientRequest
	recipeIngredients := make([]CreateIngredientRequest, len(ingredients))
	for i, ing := range ingredients {
		recipeIngredients[i] = CreateIngredientRequest{
			FoodID:        ing.FoodID,
			QuantityGrams: ing.QuantityGrams,
		}
	}

	// Create recipe request
	req := CreateRecipeRequest{
		Name:        name,
		Tag:         tag,
		Ingredients: recipeIngredients,
	}

	// Create recipe via service
	recipe, err := a.service.CreateRecipe(context.Background(), userID, req)
	if err != nil {
		return diary.RecipeCreatedResponse{}, err
	}

	// Get nutrition for the created recipe
	recipeWithNutrition, err := a.service.GetRecipe(context.Background(), userID, int(recipe.ID))
	if err != nil {
		return diary.RecipeCreatedResponse{}, err
	}

	// Return response
	return diary.RecipeCreatedResponse{
		ID:            recipeWithNutrition.ID,
		Name:          recipeWithNutrition.Name,
		Tag:           recipeWithNutrition.Tag,
		TotalCalories: recipeWithNutrition.TotalCalories,
		TotalProtein:  recipeWithNutrition.TotalProtein,
		TotalCarbs:    recipeWithNutrition.TotalCarbs,
		TotalFat:      recipeWithNutrition.TotalFat,
		TotalFiber:    recipeWithNutrition.TotalFiber,
	}, nil
}

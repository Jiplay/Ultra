package recipe

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

var (
	// ErrRecipeNotFound is returned when a recipe cannot be found
	ErrRecipeNotFound = errors.New("recipe not found")

	// ErrUnauthorized is returned when user is not authenticated
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden is returned when user doesn't have permission
	ErrForbidden = errors.New("you don't have permission to access this recipe")

	// ErrInvalidInput is returned for validation failures
	ErrInvalidInput = errors.New("invalid input")

	// ErrIngredientNotFound is returned when an ingredient cannot be found
	ErrIngredientNotFound = errors.New("ingredient not found")

	// ErrFoodNotFound is returned when a food item cannot be found
	ErrFoodNotFound = errors.New("food not found")
)

// Service handles recipe business logic
type Service struct {
	repo         *Repository
	foodProvider FoodProvider
	db           *gorm.DB
}

// NewService creates a new recipe service
func NewService(repo *Repository, foodProvider FoodProvider, db *gorm.DB) *Service {
	return &Service{
		repo:         repo,
		foodProvider: foodProvider,
		db:           db,
	}
}

// CreateRecipe creates a new recipe with ingredients in a single transaction
func (s *Service) CreateRecipe(ctx context.Context, userID uint, req CreateRecipeRequest) (*Recipe, error) {
	// Validation
	if req.Name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidInput)
	}

	if len(req.Name) > 255 {
		return nil, fmt.Errorf("%w: name must be less than 255 characters", ErrInvalidInput)
	}

	// Validate all food IDs exist before starting transaction
	if len(req.Ingredients) > 0 {
		foodIDs := make([]int, len(req.Ingredients))
		for i, ing := range req.Ingredients {
			if ing.QuantityGrams <= 0 {
				return nil, fmt.Errorf("%w: quantity must be greater than 0", ErrInvalidInput)
			}
			if ing.QuantityGrams > 100000 {
				return nil, fmt.Errorf("%w: quantity must be less than 100000 grams", ErrInvalidInput)
			}
			foodIDs[i] = int(ing.FoodID)
		}

		// Batch check all foods exist
		foods, err := s.foodProvider.GetByIDs(foodIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to validate food items: %w", err)
		}

		if len(foods) != len(foodIDs) {
			return nil, fmt.Errorf("%w: one or more food items not found", ErrFoodNotFound)
		}

		// Check for duplicate food IDs
		seen := make(map[uint]bool)
		for _, ing := range req.Ingredients {
			if seen[ing.FoodID] {
				return nil, fmt.Errorf("%w: duplicate food ID %d in ingredients", ErrInvalidInput, ing.FoodID)
			}
			seen[ing.FoodID] = true
		}
	}

	// Create recipe and ingredients in a transaction
	var recipe *Recipe
	err := s.db.Transaction(func(tx *gorm.DB) error {
		recipe = &Recipe{
			Name:   req.Name,
			UserID: &userID,
		}

		if err := tx.Create(recipe).Error; err != nil {
			return fmt.Errorf("failed to create recipe: %w", err)
		}

		// Add ingredients
		for _, ing := range req.Ingredients {
			if ing.QuantityGrams <= 0 {
				continue
			}

			ingredient := &RecipeIngredient{
				RecipeID:      recipe.ID,
				FoodID:        ing.FoodID,
				QuantityGrams: ing.QuantityGrams,
			}

			if err := tx.Create(ingredient).Error; err != nil {
				return fmt.Errorf("failed to add ingredient: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload recipe with ingredients
	return s.repo.GetByID(int(recipe.ID))
}

// GetRecipe retrieves a recipe by ID with nutrition calculated
func (s *Service) GetRecipe(ctx context.Context, userID uint, recipeID int) (*RecipeWithNutrition, error) {
	recipe, err := s.repo.GetByID(recipeID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRecipeNotFound, err)
	}

	// Check if user has access (own recipe or global recipe)
	if recipe.UserID != nil && *recipe.UserID != userID {
		return nil, ErrForbidden
	}

	// Calculate nutrition using the new optimized method
	return s.calculateNutrition(recipe)
}

// ListRecipes retrieves recipes for a user with optional filtering
func (s *Service) ListRecipes(ctx context.Context, userID uint, userOnly bool) ([]RecipeListResponse, error) {
	recipes, err := s.repo.GetByUserID(userID, userOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipes: %w", err)
	}

	return s.enrichRecipesWithNutrition(recipes)
}

// UpdateRecipe updates a recipe's basic information (not ingredients)
func (s *Service) UpdateRecipe(ctx context.Context, userID uint, recipeID int, req UpdateRecipeRequest) (*Recipe, error) {
	recipe, err := s.repo.GetByID(recipeID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRecipeNotFound, err)
	}

	// Check ownership
	if recipe.UserID == nil || *recipe.UserID != userID {
		return nil, ErrForbidden
	}

	// Validation
	if req.Name != "" {
		if len(req.Name) > 255 {
			return nil, fmt.Errorf("%w: name must be less than 255 characters", ErrInvalidInput)
		}
		recipe.Name = req.Name
	}

	if err := s.repo.Update(recipe); err != nil {
		return nil, fmt.Errorf("failed to update recipe: %w", err)
	}

	return recipe, nil
}

// DeleteRecipe deletes a recipe and all its ingredients (cascade)
func (s *Service) DeleteRecipe(ctx context.Context, userID uint, recipeID int) error {
	recipe, err := s.repo.GetByID(recipeID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrRecipeNotFound, err)
	}

	// Check ownership
	if recipe.UserID == nil || *recipe.UserID != userID {
		return ErrForbidden
	}

	if err := s.repo.Delete(recipeID); err != nil {
		return fmt.Errorf("failed to delete recipe: %w", err)
	}

	return nil
}

// AddIngredient adds an ingredient to a recipe
func (s *Service) AddIngredient(ctx context.Context, userID uint, recipeID int, req AddIngredientRequest) (*RecipeIngredient, error) {
	recipe, err := s.repo.GetByID(recipeID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRecipeNotFound, err)
	}

	// Check ownership
	if recipe.UserID == nil || *recipe.UserID != userID {
		return nil, ErrForbidden
	}

	// Validation
	if req.QuantityGrams <= 0 {
		return nil, fmt.Errorf("%w: quantity must be greater than 0", ErrInvalidInput)
	}

	if req.QuantityGrams > 100000 {
		return nil, fmt.Errorf("%w: quantity must be less than 100000 grams", ErrInvalidInput)
	}

	// Verify food exists
	_, err = s.foodProvider.GetByID(int(req.FoodID))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFoodNotFound, err)
	}

	ingredient := &RecipeIngredient{
		RecipeID:      uint(recipeID),
		FoodID:        req.FoodID,
		QuantityGrams: req.QuantityGrams,
	}

	if err := s.repo.AddIngredient(ingredient); err != nil {
		return nil, fmt.Errorf("failed to add ingredient: %w", err)
	}

	return ingredient, nil
}

// UpdateIngredient updates an ingredient's quantity
func (s *Service) UpdateIngredient(ctx context.Context, userID uint, recipeID, ingredientID int, req UpdateIngredientRequest) (*RecipeIngredient, error) {
	recipe, err := s.repo.GetByID(recipeID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRecipeNotFound, err)
	}

	// Check ownership
	if recipe.UserID == nil || *recipe.UserID != userID {
		return nil, ErrForbidden
	}

	ingredient, err := s.repo.GetIngredient(ingredientID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrIngredientNotFound, err)
	}

	// Verify ingredient belongs to this recipe
	if ingredient.RecipeID != uint(recipeID) {
		return nil, fmt.Errorf("%w: ingredient does not belong to this recipe", ErrInvalidInput)
	}

	// Validation
	if req.QuantityGrams <= 0 {
		return nil, fmt.Errorf("%w: quantity must be greater than 0", ErrInvalidInput)
	}

	if req.QuantityGrams > 100000 {
		return nil, fmt.Errorf("%w: quantity must be less than 100000 grams", ErrInvalidInput)
	}

	ingredient.QuantityGrams = req.QuantityGrams

	if err := s.repo.UpdateIngredient(ingredient); err != nil {
		return nil, fmt.Errorf("failed to update ingredient: %w", err)
	}

	return ingredient, nil
}

// DeleteIngredient removes an ingredient from a recipe
func (s *Service) DeleteIngredient(ctx context.Context, userID uint, recipeID, ingredientID int) error {
	recipe, err := s.repo.GetByID(recipeID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrRecipeNotFound, err)
	}

	// Check ownership
	if recipe.UserID == nil || *recipe.UserID != userID {
		return ErrForbidden
	}

	ingredient, err := s.repo.GetIngredient(ingredientID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrIngredientNotFound, err)
	}

	// Verify ingredient belongs to this recipe
	if ingredient.RecipeID != uint(recipeID) {
		return fmt.Errorf("%w: ingredient does not belong to this recipe", ErrInvalidInput)
	}

	if err := s.repo.DeleteIngredient(ingredientID); err != nil {
		return fmt.Errorf("failed to delete ingredient: %w", err)
	}

	return nil
}

// calculateNutrition calculates nutrition for a single recipe
// This method now returns an error if any food is missing (no silent failures)
func (s *Service) calculateNutrition(recipe *Recipe) (*RecipeWithNutrition, error) {
	result := &RecipeWithNutrition{
		Recipe: *recipe,
	}

	if len(recipe.Ingredients) == 0 {
		return result, nil
	}

	// Collect all food IDs
	foodIDs := make([]int, len(recipe.Ingredients))
	for i, ing := range recipe.Ingredients {
		foodIDs[i] = int(ing.FoodID)
	}

	// Batch fetch all foods in a single query
	foods, err := s.foodProvider.GetByIDs(foodIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get food items: %w", err)
	}

	// Create a map for quick lookup
	foodMap := make(map[uint]*Food)
	for _, food := range foods {
		foodMap[food.ID] = food
	}

	// Calculate nutrition
	for _, ingredient := range recipe.Ingredients {
		food, exists := foodMap[ingredient.FoodID]
		if !exists {
			return nil, fmt.Errorf("%w: food ID %d not found", ErrFoodNotFound, ingredient.FoodID)
		}

		// Calculate nutrition: food_per_100g * (grams / 100)
		multiplier := ingredient.QuantityGrams / 100.0
		result.TotalCalories += food.Calories * multiplier
		result.TotalProtein += food.Protein * multiplier
		result.TotalCarbs += food.Carbs * multiplier
		result.TotalFat += food.Fat * multiplier
		result.TotalFiber += food.Fiber * multiplier
		result.TotalWeight += ingredient.QuantityGrams
	}

	// Calculate per-100g nutrition
	if result.TotalWeight > 0 {
		per100g := 100.0 / result.TotalWeight
		result.CaloriesPer100g = result.TotalCalories * per100g
		result.ProteinPer100g = result.TotalProtein * per100g
		result.CarbsPer100g = result.TotalCarbs * per100g
		result.FatPer100g = result.TotalFat * per100g
		result.FiberPer100g = result.TotalFiber * per100g
	}

	return result, nil
}

// enrichRecipesWithNutrition calculates nutrition for multiple recipes efficiently
func (s *Service) enrichRecipesWithNutrition(recipes []Recipe) ([]RecipeListResponse, error) {
	result := make([]RecipeListResponse, 0, len(recipes))

	// Collect all unique food IDs across all recipes
	foodIDSet := make(map[int]bool)
	for _, recipe := range recipes {
		for _, ingredient := range recipe.Ingredients {
			foodIDSet[int(ingredient.FoodID)] = true
		}
	}

	// Convert set to slice
	foodIDs := make([]int, 0, len(foodIDSet))
	for id := range foodIDSet {
		foodIDs = append(foodIDs, id)
	}

	// Batch fetch all foods in a SINGLE query for all recipes
	foods, err := s.foodProvider.GetByIDs(foodIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get food items: %w", err)
	}

	// Create a map for quick lookup
	foodMap := make(map[uint]*Food)
	for _, food := range foods {
		foodMap[food.ID] = food
	}

	// Calculate nutrition for each recipe
	for _, recipe := range recipes {
		enriched := RecipeListResponse{
			ID:          recipe.ID,
			CreatedAt:   recipe.CreatedAt,
			UpdatedAt:   recipe.UpdatedAt,
			Name:        recipe.Name,
			UserID:      recipe.UserID,
			Ingredients: make([]IngredientWithDetails, 0, len(recipe.Ingredients)),
		}

		// Calculate nutrition for each ingredient
		for _, ingredient := range recipe.Ingredients {
			food, exists := foodMap[ingredient.FoodID]
			if !exists {
				// Return error instead of silently skipping
				return nil, fmt.Errorf("%w: food ID %d not found", ErrFoodNotFound, ingredient.FoodID)
			}

			// Calculate nutrition: food_per_100g * (grams / 100)
			multiplier := ingredient.QuantityGrams / 100.0

			ingredientDetail := IngredientWithDetails{
				ID:            ingredient.ID,
				FoodID:        ingredient.FoodID,
				FoodName:      food.Name,
				QuantityGrams: ingredient.QuantityGrams,
				Calories:      food.Calories * multiplier,
				Protein:       food.Protein * multiplier,
				Carbs:         food.Carbs * multiplier,
				Fat:           food.Fat * multiplier,
				Fiber:         food.Fiber * multiplier,
			}

			enriched.Ingredients = append(enriched.Ingredients, ingredientDetail)

			// Add to recipe totals
			enriched.TotalCalories += ingredientDetail.Calories
			enriched.TotalProtein += ingredientDetail.Protein
			enriched.TotalCarbs += ingredientDetail.Carbs
			enriched.TotalFat += ingredientDetail.Fat
			enriched.TotalFiber += ingredientDetail.Fiber
			enriched.TotalWeight += ingredient.QuantityGrams
		}

		// Calculate per-100g nutrition
		if enriched.TotalWeight > 0 {
			per100g := 100.0 / enriched.TotalWeight
			enriched.CaloriesPer100g = enriched.TotalCalories * per100g
			enriched.ProteinPer100g = enriched.TotalProtein * per100g
			enriched.CarbsPer100g = enriched.TotalCarbs * per100g
			enriched.FatPer100g = enriched.TotalFat * per100g
			enriched.FiberPer100g = enriched.TotalFiber * per100g
		}

		result = append(result, enriched)
	}

	return result, nil
}

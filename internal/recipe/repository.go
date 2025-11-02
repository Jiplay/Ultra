package recipe

import (
	"ultra-bis/internal/food"

	"gorm.io/gorm"
)

// Repository handles recipe database operations
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new recipe repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new recipe
func (r *Repository) Create(recipe *Recipe) error {
	return r.db.Create(recipe).Error
}

// GetByID retrieves a recipe by ID with ingredients preloaded
func (r *Repository) GetByID(id int) (*Recipe, error) {
	var recipe Recipe
	err := r.db.Preload("Ingredients").First(&recipe, id).Error
	return &recipe, err
}

// GetAll retrieves all recipes (global and user-specific)
func (r *Repository) GetAll() ([]Recipe, error) {
	var recipes []Recipe
	err := r.db.Preload("Ingredients").Find(&recipes).Error
	return recipes, err
}

// GetByUserID retrieves recipes for a specific user (includes global recipes)
func (r *Repository) GetByUserID(userID uint, userOnly bool) ([]Recipe, error) {
	var recipes []Recipe
	query := r.db.Preload("Ingredients")

	if userOnly {
		// Only user's private recipes
		query = query.Where("user_id = ?", userID)
	} else {
		// User's private recipes + global recipes
		query = query.Where("user_id = ? OR user_id IS NULL", userID)
	}

	err := query.Find(&recipes).Error
	return recipes, err
}

// GetGlobal retrieves all global recipes (user_id is NULL)
func (r *Repository) GetGlobal() ([]Recipe, error) {
	var recipes []Recipe
	err := r.db.Preload("Ingredients").Where("user_id IS NULL").Find(&recipes).Error
	return recipes, err
}

// Update updates a recipe
func (r *Repository) Update(recipe *Recipe) error {
	return r.db.Save(recipe).Error
}

// Delete deletes a recipe (cascade will delete ingredients)
func (r *Repository) Delete(id int) error {
	return r.db.Delete(&Recipe{}, id).Error
}

// AddIngredient adds an ingredient to a recipe
func (r *Repository) AddIngredient(ingredient *RecipeIngredient) error {
	return r.db.Create(ingredient).Error
}

// GetIngredient retrieves a specific ingredient by ID
func (r *Repository) GetIngredient(ingredientID int) (*RecipeIngredient, error) {
	var ingredient RecipeIngredient
	err := r.db.First(&ingredient, ingredientID).Error
	return &ingredient, err
}

// UpdateIngredient updates an ingredient
func (r *Repository) UpdateIngredient(ingredient *RecipeIngredient) error {
	return r.db.Save(ingredient).Error
}

// DeleteIngredient removes an ingredient from a recipe
func (r *Repository) DeleteIngredient(ingredientID int) error {
	return r.db.Delete(&RecipeIngredient{}, ingredientID).Error
}

// CalculateNutrition calculates total nutrition for a recipe
func (r *Repository) CalculateNutrition(recipeID int, foodRepo *food.Repository) (*RecipeWithNutrition, error) {
	recipe, err := r.GetByID(recipeID)
	if err != nil {
		return nil, err
	}

	result := &RecipeWithNutrition{
		Recipe: *recipe,
	}

	// Sum nutrition from all ingredients
	for _, ingredient := range recipe.Ingredients {
		foodItem, err := foodRepo.GetByID(int(ingredient.FoodID))
		if err != nil {
			continue // Skip if food not found
		}

		result.Calories += foodItem.Calories * ingredient.Quantity
		result.Protein += foodItem.Protein * ingredient.Quantity
		result.Carbs += foodItem.Carbs * ingredient.Quantity
		result.Fat += foodItem.Fat * ingredient.Quantity
		result.Fiber += foodItem.Fiber * ingredient.Quantity
	}

	return result, nil
}

// GetAllWithNutrition retrieves all recipes with nutrition and ingredient details
func (r *Repository) GetAllWithNutrition(foodRepo *food.Repository) ([]RecipeListResponse, error) {
	recipes, err := r.GetAll()
	if err != nil {
		return nil, err
	}

	return r.enrichRecipesWithNutrition(recipes, foodRepo), nil
}

// GetByUserIDWithNutrition retrieves recipes for a user with nutrition and ingredient details
func (r *Repository) GetByUserIDWithNutrition(userID uint, userOnly bool, foodRepo *food.Repository) ([]RecipeListResponse, error) {
	recipes, err := r.GetByUserID(userID, userOnly)
	if err != nil {
		return nil, err
	}

	return r.enrichRecipesWithNutrition(recipes, foodRepo), nil
}

// enrichRecipesWithNutrition calculates nutrition for each recipe and ingredient
func (r *Repository) enrichRecipesWithNutrition(recipes []Recipe, foodRepo *food.Repository) []RecipeListResponse {
	result := make([]RecipeListResponse, 0, len(recipes))

	for _, recipe := range recipes {
		enriched := RecipeListResponse{
			ID:          recipe.ID,
			CreatedAt:   recipe.CreatedAt,
			UpdatedAt:   recipe.UpdatedAt,
			Name:        recipe.Name,
			ServingSize: recipe.ServingSize,
			UserID:      recipe.UserID,
			Ingredients: make([]IngredientWithDetails, 0, len(recipe.Ingredients)),
		}

		// Calculate nutrition for each ingredient
		for _, ingredient := range recipe.Ingredients {
			foodItem, err := foodRepo.GetByID(int(ingredient.FoodID))
			if err != nil {
				continue // Skip if food not found
			}

			ingredientDetail := IngredientWithDetails{
				ID:       ingredient.ID,
				FoodID:   ingredient.FoodID,
				FoodName: foodItem.Name,
				Quantity: ingredient.Quantity,
				Calories: foodItem.Calories * ingredient.Quantity,
				Protein:  foodItem.Protein * ingredient.Quantity,
				Carbs:    foodItem.Carbs * ingredient.Quantity,
				Fat:      foodItem.Fat * ingredient.Quantity,
				Fiber:    foodItem.Fiber * ingredient.Quantity,
			}

			enriched.Ingredients = append(enriched.Ingredients, ingredientDetail)

			// Add to recipe totals
			enriched.Calories += ingredientDetail.Calories
			enriched.Protein += ingredientDetail.Protein
			enriched.Carbs += ingredientDetail.Carbs
			enriched.Fat += ingredientDetail.Fat
			enriched.Fiber += ingredientDetail.Fiber
		}

		result = append(result, enriched)
	}

	return result
}

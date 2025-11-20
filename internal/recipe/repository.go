package recipe

import (
	"ultra-bis/internal/food"

	"github.com/lib/pq"
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
// If tags are provided, filters recipes that have ANY of the specified tags
func (r *Repository) GetByUserID(userID uint, userOnly bool, tags []string) ([]Recipe, error) {
	var recipes []Recipe
	query := r.db.Preload("Ingredients")

	if userOnly {
		// Only user's private recipes
		query = query.Where("user_id = ?", userID)
	} else {
		// User's private recipes + global recipes
		query = query.Where("user_id = ? OR user_id IS NULL", userID)
	}

	// Filter by tags if provided (recipes with any of the specified tags)
	if len(tags) > 0 {
		query = query.Where("tags && ?", pq.Array(tags))
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

	// Sum nutrition from all ingredients (food nutrition is per 100g)
	for _, ingredient := range recipe.Ingredients {
		foodItem, err := foodRepo.GetByID(int(ingredient.FoodID))
		if err != nil {
			continue // Skip if food not found
		}

		// Calculate nutrition: food_per_100g * (grams / 100)
		multiplier := ingredient.QuantityGrams / 100.0
		result.TotalCalories += foodItem.Calories * multiplier
		result.TotalProtein += foodItem.Protein * multiplier
		result.TotalCarbs += foodItem.Carbs * multiplier
		result.TotalFat += foodItem.Fat * multiplier
		result.TotalFiber += foodItem.Fiber * multiplier
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

// GetAllWithNutrition retrieves all recipes with nutrition and ingredient details
func (r *Repository) GetAllWithNutrition(foodRepo *food.Repository) ([]RecipeListResponse, error) {
	recipes, err := r.GetAll()
	if err != nil {
		return nil, err
	}

	return r.enrichRecipesWithNutrition(recipes, foodRepo), nil
}

// GetByUserIDWithNutrition retrieves recipes for a user with nutrition and ingredient details
// If tags are provided, filters recipes that have ANY of the specified tags
func (r *Repository) GetByUserIDWithNutrition(userID uint, userOnly bool, tags []string, foodRepo *food.Repository) ([]RecipeListResponse, error) {
	recipes, err := r.GetByUserID(userID, userOnly, tags)
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
			UserID:      recipe.UserID,
			Tags:        recipe.Tags,
			Ingredients: make([]IngredientWithDetails, 0, len(recipe.Ingredients)),
		}

		// Calculate nutrition for each ingredient
		for _, ingredient := range recipe.Ingredients {
			foodItem, err := foodRepo.GetByID(int(ingredient.FoodID))
			if err != nil {
				continue // Skip if food not found
			}

			// Calculate nutrition: food_per_100g * (grams / 100)
			multiplier := ingredient.QuantityGrams / 100.0

			ingredientDetail := IngredientWithDetails{
				ID:            ingredient.ID,
				FoodID:        ingredient.FoodID,
				FoodName:      foodItem.Name,
				QuantityGrams: ingredient.QuantityGrams,
				Calories:      foodItem.Calories * multiplier,
				Protein:       foodItem.Protein * multiplier,
				Carbs:         foodItem.Carbs * multiplier,
				Fat:           foodItem.Fat * multiplier,
				Fiber:         foodItem.Fiber * multiplier,
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

	return result
}

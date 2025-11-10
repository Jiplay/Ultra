package recipe

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"ultra-bis/internal/food"
	"ultra-bis/test/testutil"
)

// setupRecipeTest creates a test DB with Recipe and Food migrations
func setupRecipeTest(t *testing.T) (*gorm.DB, *Repository, *food.Repository) {
	t.Helper()
	db := testutil.SetupTestDB(t)

	// Run migrations for foods, recipes, and recipe_ingredients
	if err := db.AutoMigrate(&food.Food{}, &Recipe{}, &RecipeIngredient{}); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	recipeRepo := NewRepository(db)
	foodRepo := food.NewRepository(db)

	return db, recipeRepo, foodRepo
}

// createTestFood creates a food item for testing
func createTestFood(t *testing.T, foodRepo *food.Repository, name string, calories, protein, carbs, fat, fiber float64) *food.Food {
	t.Helper()

	req := food.CreateFoodRequest{
		Name:        name,
		Description: "Test food per 100g",
		Calories:    calories,
		Protein:     protein,
		Carbs:       carbs,
		Fat:         fat,
		Fiber:       fiber,
	}

	created, err := foodRepo.Create(req)
	require.NoError(t, err)
	return created
}

func TestRepository_CalculateNutrition_SingleIngredient(t *testing.T) {
	_, recipeRepo, foodRepo := setupRecipeTest(t)

	// Create a test food: Chicken (165 cal, 31g protein per 100g)
	chicken := createTestFood(t, foodRepo, "Chicken Breast", 165, 31, 0, 3.6, 0)

	// Create a recipe with 200g of chicken
	recipe := &Recipe{
		Name: "Grilled Chicken",
		Ingredients: []RecipeIngredient{
			{
				FoodID:        chicken.ID,
				QuantityGrams: 200,
			},
		},
	}

	err := recipeRepo.Create(recipe)
	require.NoError(t, err)

	// Calculate nutrition
	result, err := recipeRepo.CalculateNutrition(int(recipe.ID), foodRepo)
	assert.NoError(t, err)
	require.NotNil(t, result)

	// Expected: 165 * (200/100) = 330 calories
	// Expected: 31 * (200/100) = 62g protein
	assert.InDelta(t, 330.0, result.TotalCalories, 0.01, "Total calories should be 330")
	assert.InDelta(t, 62.0, result.TotalProtein, 0.01, "Total protein should be 62g")
	assert.InDelta(t, 0.0, result.TotalCarbs, 0.01)
	assert.InDelta(t, 7.2, result.TotalFat, 0.01, "Total fat should be 7.2g")
	assert.InDelta(t, 0.0, result.TotalFiber, 0.01)
	assert.InDelta(t, 200.0, result.TotalWeight, 0.01)

	// Per 100g should be the same as original food (since we only have one ingredient)
	assert.InDelta(t, 165.0, result.CaloriesPer100g, 0.01)
	assert.InDelta(t, 31.0, result.ProteinPer100g, 0.01)
}

func TestRepository_CalculateNutrition_MultipleIngredients(t *testing.T) {
	_, recipeRepo, foodRepo := setupRecipeTest(t)

	// Create test foods (all per 100g)
	chicken := createTestFood(t, foodRepo, "Chicken Breast", 165, 31, 0, 3.6, 0)
	rice := createTestFood(t, foodRepo, "Brown Rice", 111, 2.6, 23, 0.9, 1.8)
	broccoli := createTestFood(t, foodRepo, "Broccoli", 34, 2.8, 7, 0.4, 2.6)

	// Create recipe: 200g chicken + 150g rice + 100g broccoli = 450g total
	recipe := &Recipe{
		Name: "Chicken Rice Bowl",
		Ingredients: []RecipeIngredient{
			{FoodID: chicken.ID, QuantityGrams: 200},
			{FoodID: rice.ID, QuantityGrams: 150},
			{FoodID: broccoli.ID, QuantityGrams: 100},
		},
	}

	err := recipeRepo.Create(recipe)
	require.NoError(t, err)

	// Calculate nutrition
	result, err := recipeRepo.CalculateNutrition(int(recipe.ID), foodRepo)
	assert.NoError(t, err)
	require.NotNil(t, result)

	// Expected calculations:
	// Chicken: 165 * 2 = 330 cal, 31 * 2 = 62g protein, 7.2g fat
	// Rice: 111 * 1.5 = 166.5 cal, 2.6 * 1.5 = 3.9g protein, 23 * 1.5 = 34.5g carbs, 0.9 * 1.5 = 1.35g fat, 1.8 * 1.5 = 2.7g fiber
	// Broccoli: 34 * 1 = 34 cal, 2.8 * 1 = 2.8g protein, 7 * 1 = 7g carbs, 0.4g fat, 2.6g fiber
	// Total: 530.5 cal, 68.7g protein, 41.5g carbs, 8.95g fat, 5.3g fiber, 450g weight

	assert.InDelta(t, 530.5, result.TotalCalories, 0.1)
	assert.InDelta(t, 68.7, result.TotalProtein, 0.1)
	assert.InDelta(t, 41.5, result.TotalCarbs, 0.1)
	assert.InDelta(t, 8.95, result.TotalFat, 0.1)
	assert.InDelta(t, 5.3, result.TotalFiber, 0.1)
	assert.InDelta(t, 450.0, result.TotalWeight, 0.01)

	// Per 100g calculations: total * (100/450)
	assert.InDelta(t, 117.89, result.CaloriesPer100g, 0.1, "Calories per 100g")
	assert.InDelta(t, 15.27, result.ProteinPer100g, 0.1, "Protein per 100g")
	assert.InDelta(t, 9.22, result.CarbsPer100g, 0.1, "Carbs per 100g")
	assert.InDelta(t, 1.99, result.FatPer100g, 0.1, "Fat per 100g")
	assert.InDelta(t, 1.18, result.FiberPer100g, 0.1, "Fiber per 100g")
}

func TestRepository_CalculateNutrition_ZeroWeight(t *testing.T) {
	_, recipeRepo, foodRepo := setupRecipeTest(t)

	// Create a recipe with no ingredients
	recipe := &Recipe{
		Name:        "Empty Recipe",
		Ingredients: []RecipeIngredient{},
	}

	err := recipeRepo.Create(recipe)
	require.NoError(t, err)

	// Calculate nutrition
	result, err := recipeRepo.CalculateNutrition(int(recipe.ID), foodRepo)
	assert.NoError(t, err)
	require.NotNil(t, result)

	// All values should be zero
	assert.Equal(t, 0.0, result.TotalCalories)
	assert.Equal(t, 0.0, result.TotalProtein)
	assert.Equal(t, 0.0, result.TotalWeight)
	assert.Equal(t, 0.0, result.CaloriesPer100g, "Per 100g values should be 0 when total weight is 0")
}

func TestRepository_CalculateNutrition_FractionalGrams(t *testing.T) {
	_, recipeRepo, foodRepo := setupRecipeTest(t)

	// Create test food: Olive oil (884 cal, 0g protein, 100g fat per 100g)
	oil := createTestFood(t, foodRepo, "Olive Oil", 884, 0, 0, 100, 0)

	// Create recipe with 15.5g of olive oil (typical tablespoon)
	recipe := &Recipe{
		Name: "Salad Dressing",
		Ingredients: []RecipeIngredient{
			{FoodID: oil.ID, QuantityGrams: 15.5},
		},
	}

	err := recipeRepo.Create(recipe)
	require.NoError(t, err)

	// Calculate nutrition
	result, err := recipeRepo.CalculateNutrition(int(recipe.ID), foodRepo)
	assert.NoError(t, err)
	require.NotNil(t, result)

	// Expected: 884 * (15.5/100) = 137.02 calories
	// Expected: 100 * (15.5/100) = 15.5g fat
	assert.InDelta(t, 137.02, result.TotalCalories, 0.1)
	assert.InDelta(t, 0.0, result.TotalProtein, 0.01)
	assert.InDelta(t, 15.5, result.TotalFat, 0.1)
	assert.InDelta(t, 15.5, result.TotalWeight, 0.01)

	// Per 100g should match original (single ingredient)
	assert.InDelta(t, 884.0, result.CaloriesPer100g, 0.1)
	assert.InDelta(t, 100.0, result.FatPer100g, 0.1)
}

func TestRepository_EnrichRecipesWithNutrition(t *testing.T) {
	_, recipeRepo, foodRepo := setupRecipeTest(t)

	// Create test foods
	chicken := createTestFood(t, foodRepo, "Chicken", 165, 31, 0, 3.6, 0)
	rice := createTestFood(t, foodRepo, "Rice", 130, 2.7, 28, 0.3, 0.4)

	// Create two recipes
	recipe1 := &Recipe{
		Name: "Recipe 1",
		Ingredients: []RecipeIngredient{
			{FoodID: chicken.ID, QuantityGrams: 100},
		},
	}
	recipe2 := &Recipe{
		Name: "Recipe 2",
		Ingredients: []RecipeIngredient{
			{FoodID: rice.ID, QuantityGrams: 200},
		},
	}

	require.NoError(t, recipeRepo.Create(recipe1))
	require.NoError(t, recipeRepo.Create(recipe2))

	// Get all recipes with nutrition
	results, err := recipeRepo.GetAllWithNutrition(foodRepo)
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	// Check first recipe (100g chicken)
	assert.Equal(t, "Recipe 1", results[0].Name)
	assert.InDelta(t, 165.0, results[0].TotalCalories, 0.1)
	assert.InDelta(t, 31.0, results[0].TotalProtein, 0.1)
	assert.Len(t, results[0].Ingredients, 1)
	assert.Equal(t, "Chicken", results[0].Ingredients[0].FoodName)

	// Check second recipe (200g rice)
	assert.Equal(t, "Recipe 2", results[1].Name)
	assert.InDelta(t, 260.0, results[1].TotalCalories, 0.1)
	assert.InDelta(t, 5.4, results[1].TotalProtein, 0.1)
	assert.Len(t, results[1].Ingredients, 1)
	assert.Equal(t, "Rice", results[1].Ingredients[0].FoodName)
}

func TestRepository_CalculateNutrition_MissingFood(t *testing.T) {
	_, recipeRepo, foodRepo := setupRecipeTest(t)

	// Create a recipe with a non-existent food ID
	recipe := &Recipe{
		Name: "Recipe with missing food",
		Ingredients: []RecipeIngredient{
			{FoodID: 99999, QuantityGrams: 100}, // Non-existent food
		},
	}

	err := recipeRepo.Create(recipe)
	require.NoError(t, err)

	// Calculate nutrition - should not error but skip the missing food
	result, err := recipeRepo.CalculateNutrition(int(recipe.ID), foodRepo)
	assert.NoError(t, err)
	require.NotNil(t, result)

	// All nutrition values should be zero since food doesn't exist
	assert.Equal(t, 0.0, result.TotalCalories)
	assert.Equal(t, 0.0, result.TotalProtein)
	assert.Equal(t, 0.0, result.TotalWeight)
}

func TestRepository_CalculateNutrition_ComplexRecipe(t *testing.T) {
	_, recipeRepo, foodRepo := setupRecipeTest(t)

	// Create a complex recipe with multiple ingredients
	salmon := createTestFood(t, foodRepo, "Salmon", 206, 22, 0, 13, 0)
	quinoa := createTestFood(t, foodRepo, "Quinoa", 120, 4.4, 21, 1.9, 2.8)
	spinach := createTestFood(t, foodRepo, "Spinach", 23, 2.9, 3.6, 0.4, 2.2)
	avocado := createTestFood(t, foodRepo, "Avocado", 160, 2, 8.5, 14.7, 6.7)

	// Create recipe: 150g salmon + 100g quinoa + 50g spinach + 75g avocado = 375g total
	recipe := &Recipe{
		Name: "Power Bowl",
		Ingredients: []RecipeIngredient{
			{FoodID: salmon.ID, QuantityGrams: 150},
			{FoodID: quinoa.ID, QuantityGrams: 100},
			{FoodID: spinach.ID, QuantityGrams: 50},
			{FoodID: avocado.ID, QuantityGrams: 75},
		},
	}

	err := recipeRepo.Create(recipe)
	require.NoError(t, err)

	// Calculate nutrition
	result, err := recipeRepo.CalculateNutrition(int(recipe.ID), foodRepo)
	assert.NoError(t, err)
	require.NotNil(t, result)

	// Expected calculations:
	// Salmon: 206 * 1.5 = 309 cal, 22 * 1.5 = 33g protein, 13 * 1.5 = 19.5g fat
	// Quinoa: 120 * 1 = 120 cal, 4.4g protein, 21g carbs, 1.9g fat, 2.8g fiber
	// Spinach: 23 * 0.5 = 11.5 cal, 2.9 * 0.5 = 1.45g protein, 3.6 * 0.5 = 1.8g carbs, 0.4 * 0.5 = 0.2g fat, 2.2 * 0.5 = 1.1g fiber
	// Avocado: 160 * 0.75 = 120 cal, 2 * 0.75 = 1.5g protein, 8.5 * 0.75 = 6.375g carbs, 14.7 * 0.75 = 11.025g fat, 6.7 * 0.75 = 5.025g fiber
	// Total: 560.5 cal, 40.35g protein (33 + 4.4 + 1.45 + 1.5), 29.175g carbs, 32.625g fat, 8.925g fiber, 375g weight

	assert.InDelta(t, 560.5, result.TotalCalories, 0.2)
	assert.InDelta(t, 40.35, result.TotalProtein, 0.2)
	assert.InDelta(t, 29.175, result.TotalCarbs, 0.2)
	assert.InDelta(t, 32.625, result.TotalFat, 0.2)
	assert.InDelta(t, 8.925, result.TotalFiber, 0.2)
	assert.InDelta(t, 375.0, result.TotalWeight, 0.01)

	// Per 100g calculations: totals * (100/375)
	assert.InDelta(t, 149.47, result.CaloriesPer100g, 0.2)
	assert.InDelta(t, 10.76, result.ProteinPer100g, 0.2) // 40.35 * (100/375)
	assert.InDelta(t, 7.78, result.CarbsPer100g, 0.2)
	assert.InDelta(t, 8.70, result.FatPer100g, 0.2)
	assert.InDelta(t, 2.38, result.FiberPer100g, 0.2)
}

package tests

import (
	"testing"

	"ultra-bis/internal/recipe"

	"github.com/stretchr/testify/assert"

	"ultra-bis/internal/food"
	"ultra-bis/test/testutil"

	"gorm.io/gorm"
)

// setupRecipeTest creates a test DB with Recipe and Food migrations
func setupRecipeTest(t *testing.T) (*gorm.DB, *recipe.Repository, *food.Repository) {
	t.Helper()
	db := testutil.SetupTestDB(t)

	// Run migrations for foods, recipes, and recipe_ingredients
	if err := db.AutoMigrate(&food.Food{}, &recipe.Recipe{}, &recipe.RecipeIngredient{}); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	recipeRepo := recipe.NewRepository(db)
	foodRepo := food.NewRepository(db)

	return db, recipeRepo, foodRepo
}

func TestRepository_Create(t *testing.T) {
	_, recipeRepo, _ := setupRecipeTest(t)

	userID := uint(1)
	myrecipe := &recipe.Recipe{
		Name:   "Test Recipe",
		UserID: &userID,
	}

	err := recipeRepo.Create(myrecipe)
	assert.NoError(t, err)
	assert.NotZero(t, myrecipe.ID)
}

func TestRepository_GetByID(t *testing.T) {
	_, recipeRepo, _ := setupRecipeTest(t)

	userID := uint(1)
	myrecipe := &recipe.Recipe{
		Name:   "Test Recipe",
		UserID: &userID,
	}
	recipeRepo.Create(myrecipe)

	found, err := recipeRepo.GetByID(int(myrecipe.ID))
	assert.NoError(t, err)
	assert.Equal(t, myrecipe.Name, found.Name)
}

func TestRepository_GetByUserID(t *testing.T) {
	_, recipeRepo, _ := setupRecipeTest(t)

	userID := uint(1)
	recipe1 := &recipe.Recipe{Name: "Recipe 1", UserID: &userID}
	recipe2 := &recipe.Recipe{Name: "Recipe 2", UserID: &userID}

	recipeRepo.Create(recipe1)
	recipeRepo.Create(recipe2)

	recipes, err := recipeRepo.GetByUserID(userID, true)
	assert.NoError(t, err)
	assert.Len(t, recipes, 2)
}

func TestRepository_Update(t *testing.T) {
	_, recipeRepo, _ := setupRecipeTest(t)

	userID := uint(1)
	myrecipe := &recipe.Recipe{Name: "Original", UserID: &userID}
	recipeRepo.Create(myrecipe)

	myrecipe.Name = "Updated"
	err := recipeRepo.Update(myrecipe)
	assert.NoError(t, err)

	found, _ := recipeRepo.GetByID(int(myrecipe.ID))
	assert.Equal(t, "Updated", found.Name)
}

func TestRepository_Delete(t *testing.T) {
	_, recipeRepo, _ := setupRecipeTest(t)

	userID := uint(1)
	myrecipe := &recipe.Recipe{Name: "ToDelete", UserID: &userID}
	recipeRepo.Create(myrecipe)

	err := recipeRepo.Delete(int(myrecipe.ID))
	assert.NoError(t, err)

	_, err = recipeRepo.GetByID(int(myrecipe.ID))
	assert.Error(t, err)
}

func TestRepository_AddIngredient(t *testing.T) {
	_, recipeRepo, foodRepo := setupRecipeTest(t)

	// Create food
	food, _ := foodRepo.Create(food.CreateFoodRequest{
		Name:     "Chicken",
		Calories: 165,
		Protein:  31,
	})

	// Create recipe
	userID := uint(1)
	myrecipe := &recipe.Recipe{Name: "Test", UserID: &userID}
	recipeRepo.Create(myrecipe)

	// Add ingredient
	ingredient := &recipe.RecipeIngredient{
		RecipeID:      myrecipe.ID,
		FoodID:        food.ID,
		QuantityGrams: 200,
	}

	err := recipeRepo.AddIngredient(ingredient)
	assert.NoError(t, err)
	assert.NotZero(t, ingredient.ID)
}

func TestRepository_GetIngredient(t *testing.T) {
	_, recipeRepo, foodRepo := setupRecipeTest(t)

	// Create food
	myfood, _ := foodRepo.Create(food.CreateFoodRequest{
		Name:     "Chicken",
		Calories: 165,
	})

	// Create recipe and ingredient
	userID := uint(1)
	myrecipe := &recipe.Recipe{Name: "Test", UserID: &userID}
	recipeRepo.Create(myrecipe)

	ingredient := &recipe.RecipeIngredient{
		RecipeID:      myrecipe.ID,
		FoodID:        myfood.ID,
		QuantityGrams: 200,
	}
	recipeRepo.AddIngredient(ingredient)

	found, err := recipeRepo.GetIngredient(int(ingredient.ID))
	assert.NoError(t, err)
	assert.Equal(t, ingredient.QuantityGrams, found.QuantityGrams)
}

func TestRepository_UpdateIngredient(t *testing.T) {
	_, recipeRepo, foodRepo := setupRecipeTest(t)

	// Create food
	myfood, _ := foodRepo.Create(food.CreateFoodRequest{
		Name:     "Chicken",
		Calories: 165,
	})

	// Create recipe and ingredient
	userID := uint(1)
	myrecipe := &recipe.Recipe{Name: "Test", UserID: &userID}
	recipeRepo.Create(myrecipe)

	ingredient := &recipe.RecipeIngredient{
		RecipeID:      myrecipe.ID,
		FoodID:        myfood.ID,
		QuantityGrams: 200,
	}
	recipeRepo.AddIngredient(ingredient)

	// Update quantity
	ingredient.QuantityGrams = 300
	err := recipeRepo.UpdateIngredient(ingredient)
	assert.NoError(t, err)

	found, _ := recipeRepo.GetIngredient(int(ingredient.ID))
	assert.Equal(t, 300.0, found.QuantityGrams)
}

func TestRepository_DeleteIngredient(t *testing.T) {
	_, recipeRepo, foodRepo := setupRecipeTest(t)

	// Create food
	myfood, _ := foodRepo.Create(food.CreateFoodRequest{
		Name:     "Chicken",
		Calories: 165,
	})

	// Create recipe and ingredient
	userID := uint(1)
	myrecipe := &recipe.Recipe{Name: "Test", UserID: &userID}
	recipeRepo.Create(myrecipe)

	ingredient := &recipe.RecipeIngredient{
		RecipeID:      myrecipe.ID,
		FoodID:        myfood.ID,
		QuantityGrams: 200,
	}
	recipeRepo.AddIngredient(ingredient)

	// Delete ingredient
	err := recipeRepo.DeleteIngredient(int(ingredient.ID))
	assert.NoError(t, err)

	_, err = recipeRepo.GetIngredient(int(ingredient.ID))
	assert.Error(t, err)
}

func TestRepository_GetByUserID_IncludesGlobal(t *testing.T) {
	_, recipeRepo, _ := setupRecipeTest(t)

	userID := uint(1)

	// Create user recipe
	userRecipe := &recipe.Recipe{Name: "User Recipe", UserID: &userID}
	recipeRepo.Create(userRecipe)

	// Create global recipe (nil userID)
	globalRecipe := &recipe.Recipe{Name: "Global Recipe", UserID: nil}
	recipeRepo.Create(globalRecipe)

	// Get all recipes for user (should include global)
	recipes, err := recipeRepo.GetByUserID(userID, false)
	assert.NoError(t, err)
	assert.Len(t, recipes, 2)

	// Get only user recipes
	userRecipes, err := recipeRepo.GetByUserID(userID, true)
	assert.NoError(t, err)
	assert.Len(t, userRecipes, 1)
	assert.Equal(t, "User Recipe", userRecipes[0].Name)
}

package tests

import (
	"context"
	"testing"

	"ultra-bis/internal/recipe"

	"ultra-bis/test/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockFoodProvider is a simple in-memory mock for testing
type mockFoodProvider struct {
	foods map[int]*recipe.Food
}

func newMockFoodProvider() *mockFoodProvider {
	return &mockFoodProvider{
		foods: make(map[int]*recipe.Food),
	}
}

func (m *mockFoodProvider) addFood(id int, name string, calories, protein, carbs, fat, fiber float64) {
	m.foods[id] = &recipe.Food{
		ID:          uint(id),
		Name:        name,
		Description: "Test food per 100g",
		Calories:    calories,
		Protein:     protein,
		Carbs:       carbs,
		Fat:         fat,
		Fiber:       fiber,
	}
}

func (m *mockFoodProvider) GetByID(id int) (*recipe.Food, error) {
	food, exists := m.foods[id]
	if !exists {
		return nil, recipe.ErrFoodNotFound
	}
	return food, nil
}

func (m *mockFoodProvider) GetByIDs(ids []int) ([]*recipe.Food, error) {
	result := make([]*recipe.Food, 0, len(ids))
	for _, id := range ids {
		food, exists := m.foods[id]
		if !exists {
			return nil, recipe.ErrFoodNotFound
		}
		result = append(result, food)
	}
	return result, nil
}

func TestService_CreateRecipe_Success(t *testing.T) {
	db := testutil.SetupTestDB(t)
	db.AutoMigrate(&recipe.Recipe{}, &recipe.RecipeIngredient{})

	mockFP := newMockFoodProvider()
	mockFP.addFood(1, "Chicken", 165, 31, 0, 3.6, 0)
	mockFP.addFood(2, "Rice", 130, 2.7, 28, 0.3, 0.4)

	repo := recipe.NewRepository(db)
	service := recipe.NewService(repo, mockFP, db)

	ctx := context.Background()
	userID := uint(1)

	req := recipe.CreateRecipeRequest{
		Name: "Chicken & Rice",
		Ingredients: []recipe.CreateIngredientRequest{
			{FoodID: 1, QuantityGrams: 200},
			{FoodID: 2, QuantityGrams: 150},
		},
	}

	recipe, err := service.CreateRecipe(ctx, userID, req)

	assert.NoError(t, err)
	require.NotNil(t, recipe)
	assert.Equal(t, "Chicken & Rice", recipe.Name)
	assert.Equal(t, userID, *recipe.UserID)
	assert.Len(t, recipe.Ingredients, 2)
}

func TestService_CreateRecipe_ValidationErrors(t *testing.T) {
	db := testutil.SetupTestDB(t)
	db.AutoMigrate(&recipe.Recipe{}, &recipe.RecipeIngredient{})

	mockFP := newMockFoodProvider()
	repo := recipe.NewRepository(db)
	service := recipe.NewService(repo, mockFP, db)

	ctx := context.Background()
	userID := uint(1)

	tests := []struct {
		name        string
		req         recipe.CreateRecipeRequest
		expectedErr error
	}{
		{
			name:        "empty name",
			req:         recipe.CreateRecipeRequest{Name: ""},
			expectedErr: recipe.ErrInvalidInput,
		},
		{
			name:        "name too long",
			req:         recipe.CreateRecipeRequest{Name: string(make([]byte, 300))},
			expectedErr: recipe.ErrInvalidInput,
		},
		{
			name: "negative quantity",
			req: recipe.CreateRecipeRequest{
				Name: "Test",
				Ingredients: []recipe.CreateIngredientRequest{
					{FoodID: 1, QuantityGrams: -10},
				},
			},
			expectedErr: recipe.ErrInvalidInput,
		},
		{
			name: "quantity too large",
			req: recipe.CreateRecipeRequest{
				Name: "Test",
				Ingredients: []recipe.CreateIngredientRequest{
					{FoodID: 1, QuantityGrams: 200000},
				},
			},
			expectedErr: recipe.ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.CreateRecipe(ctx, userID, tt.req)
			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestService_CreateRecipe_DuplicateFoods(t *testing.T) {
	db := testutil.SetupTestDB(t)
	db.AutoMigrate(&recipe.Recipe{}, &recipe.RecipeIngredient{})

	mockFP := newMockFoodProvider()
	mockFP.addFood(1, "Chicken", 165, 31, 0, 3.6, 0)

	repo := recipe.NewRepository(db)
	service := recipe.NewService(repo, mockFP, db)

	ctx := context.Background()
	userID := uint(1)

	req := recipe.CreateRecipeRequest{
		Name: "Duplicate Test",
		Ingredients: []recipe.CreateIngredientRequest{
			{FoodID: 1, QuantityGrams: 100},
			{FoodID: 1, QuantityGrams: 200}, // Duplicate!
		},
	}

	_, err := service.CreateRecipe(ctx, userID, req)
	assert.ErrorIs(t, err, recipe.ErrInvalidInput)
	assert.Contains(t, err.Error(), "duplicate food ID")
}

func TestService_CreateRecipe_NonExistentFood(t *testing.T) {
	db := testutil.SetupTestDB(t)
	db.AutoMigrate(&recipe.Recipe{}, &recipe.RecipeIngredient{})

	mockFP := newMockFoodProvider()
	// Don't add food ID 99 to mock

	repo := recipe.NewRepository(db)
	service := recipe.NewService(repo, mockFP, db)

	ctx := context.Background()
	userID := uint(1)

	req := recipe.CreateRecipeRequest{
		Name: "Invalid Recipe",
		Ingredients: []recipe.CreateIngredientRequest{
			{FoodID: 99, QuantityGrams: 100},
		},
	}

	_, err := service.CreateRecipe(ctx, userID, req)
	assert.ErrorIs(t, err, recipe.ErrFoodNotFound)
}

func TestService_GetRecipe_Nutrition(t *testing.T) {
	db := testutil.SetupTestDB(t)
	db.AutoMigrate(&recipe.Recipe{}, &recipe.RecipeIngredient{})

	mockFP := newMockFoodProvider()
	mockFP.addFood(1, "Chicken", 165, 31, 0, 3.6, 0)
	mockFP.addFood(2, "Rice", 130, 2.7, 28, 0.3, 0.4)

	repo := recipe.NewRepository(db)
	service := recipe.NewService(repo, mockFP, db)

	ctx := context.Background()
	userID := uint(1)

	// Create recipe with 200g chicken + 150g rice
	recipe := &recipe.Recipe{
		Name:   "Test Recipe",
		UserID: &userID,
		Ingredients: []recipe.RecipeIngredient{
			{FoodID: 1, QuantityGrams: 200},
			{FoodID: 2, QuantityGrams: 150},
		},
	}
	repo.Create(recipe)

	// Get recipe with nutrition
	result, err := service.GetRecipe(ctx, userID, int(recipe.ID))

	assert.NoError(t, err)
	require.NotNil(t, result)

	// Expected: 165*2 + 130*1.5 = 330 + 195 = 525 calories
	assert.InDelta(t, 525.0, result.TotalCalories, 0.1)

	// Expected: 31*2 + 2.7*1.5 = 62 + 4.05 = 66.05g protein
	assert.InDelta(t, 66.05, result.TotalProtein, 0.1)

	// Total weight: 350g
	assert.InDelta(t, 350.0, result.TotalWeight, 0.1)

	// Per 100g: 525 * (100/350) = 150 cal/100g
	assert.InDelta(t, 150.0, result.CaloriesPer100g, 0.1)
}

func TestService_GetRecipe_Forbidden(t *testing.T) {
	db := testutil.SetupTestDB(t)
	db.AutoMigrate(&recipe.Recipe{}, &recipe.RecipeIngredient{})

	mockFP := newMockFoodProvider()
	repo := recipe.NewRepository(db)
	service := recipe.NewService(repo, mockFP, db)

	ctx := context.Background()

	// Create recipe owned by user 1
	ownerID := uint(1)
	myrecipe := &recipe.Recipe{
		Name:   "Private Recipe",
		UserID: &ownerID,
	}
	repo.Create(myrecipe)

	// Try to access as user 2
	otherUserID := uint(2)
	_, err := service.GetRecipe(ctx, otherUserID, int(myrecipe.ID))

	assert.ErrorIs(t, err, recipe.ErrForbidden)
}

func TestService_UpdateRecipe_Success(t *testing.T) {
	db := testutil.SetupTestDB(t)
	db.AutoMigrate(&recipe.Recipe{}, &recipe.RecipeIngredient{})

	mockFP := newMockFoodProvider()
	repo := recipe.NewRepository(db)
	service := recipe.NewService(repo, mockFP, db)

	ctx := context.Background()
	userID := uint(1)

	// Create recipe
	myrecipe := &recipe.Recipe{
		Name:   "Original Name",
		UserID: &userID,
	}
	repo.Create(myrecipe)

	// Update recipe
	req := recipe.UpdateRecipeRequest{
		Name: "Updated Name",
	}
	updated, err := service.UpdateRecipe(ctx, userID, int(myrecipe.ID), req)

	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
}

func TestService_DeleteRecipe_Success(t *testing.T) {
	db := testutil.SetupTestDB(t)
	db.AutoMigrate(&recipe.Recipe{}, &recipe.RecipeIngredient{})

	mockFP := newMockFoodProvider()
	repo := recipe.NewRepository(db)
	service := recipe.NewService(repo, mockFP, db)

	ctx := context.Background()
	userID := uint(1)

	// Create recipe
	recipe := &recipe.Recipe{
		Name:   "To Delete",
		UserID: &userID,
	}
	repo.Create(recipe)

	// Delete recipe
	err := service.DeleteRecipe(ctx, userID, int(recipe.ID))
	assert.NoError(t, err)

	// Verify it's deleted
	_, err = repo.GetByID(int(recipe.ID))
	assert.Error(t, err)
}

// trackingFoodProvider wraps a food provider and tracks GetByIDs calls
type trackingFoodProvider struct {
	wrapped   recipe.FoodProvider
	callCount int
}

func (t *trackingFoodProvider) GetByID(id int) (*recipe.Food, error) {
	return t.wrapped.GetByID(id)
}

func (t *trackingFoodProvider) GetByIDs(ids []int) ([]*recipe.Food, error) {
	t.callCount++
	return t.wrapped.GetByIDs(ids)
}

func TestService_ListRecipes_BatchFetching(t *testing.T) {
	db := testutil.SetupTestDB(t)
	db.AutoMigrate(&recipe.Recipe{}, &recipe.RecipeIngredient{})

	// Create base mock provider
	baseMockFP := newMockFoodProvider()
	baseMockFP.addFood(1, "Food1", 100, 10, 10, 5, 1)
	baseMockFP.addFood(2, "Food2", 200, 20, 20, 10, 2)

	// Wrap it with tracking
	trackingFP := &trackingFoodProvider{wrapped: baseMockFP}

	repo := recipe.NewRepository(db)
	service := recipe.NewService(repo, trackingFP, db)

	ctx := context.Background()
	userID := uint(1)

	// Create multiple recipes
	recipe1 := &recipe.Recipe{
		Name:   "Recipe 1",
		UserID: &userID,
		Ingredients: []recipe.RecipeIngredient{
			{FoodID: 1, QuantityGrams: 100},
		},
	}
	recipe2 := &recipe.Recipe{
		Name:   "Recipe 2",
		UserID: &userID,
		Ingredients: []recipe.RecipeIngredient{
			{FoodID: 2, QuantityGrams: 100},
		},
	}
	repo.Create(recipe1)
	repo.Create(recipe2)

	// List recipes - should only call GetByIDs ONCE
	recipes, err := service.ListRecipes(ctx, userID, true)

	assert.NoError(t, err)
	assert.Len(t, recipes, 2)
	assert.Equal(t, 1, trackingFP.callCount, "GetByIDs should only be called once for batch fetching")
}

package tests

import (
	"testing"

	"ultra-bis/internal/food"

	"github.com/stretchr/testify/assert"

	"ultra-bis/test/testutil"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupFoodTest creates a test DB with Food migrations
func setupFoodTest(t *testing.T) (*gorm.DB, *food.Repository) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	if err := db.AutoMigrate(&food.Food{}); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}
	return db, food.NewRepository(db)
}

func TestRepository_Create(t *testing.T) {
	_, repo := setupFoodTest(t)

	req := food.CreateFoodRequest{
		Name:        "Chicken Breast",
		Description: "Skinless, boneless, per 100g",
		Calories:    165,
		Protein:     31,
		Carbs:       0,
		Fat:         3.6,
		Fiber:       0,
	}

	food, err := repo.Create(req)
	assert.NoError(t, err)
	require.NotNil(t, food)
	assert.NotZero(t, food.ID, "Food ID should be set after creation")
	assert.Equal(t, req.Name, food.Name)
	assert.InDelta(t, req.Calories, food.Calories, 0.01)
}

func TestRepository_GetByID(t *testing.T) {
	_, repo := setupFoodTest(t)

	// Create a test food
	req := food.CreateFoodRequest{
		Name:        "Test Food",
		Description: "Test description",
		Calories:    150,
		Protein:     20,
		Carbs:       10,
		Fat:         5,
		Fiber:       3,
	}
	created, err := repo.Create(req)
	require.NoError(t, err)

	// Get by ID
	found, err := repo.GetByID(int(created.ID))
	assert.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, req.Name, found.Name)
}

func TestRepository_GetByID_NotFound(t *testing.T) {
	_, repo := setupFoodTest(t)

	// Try to get non-existent food
	found, err := repo.GetByID(99999)
	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestRepository_GetAll(t *testing.T) {
	_, repo := setupFoodTest(t)

	// Create test foods
	foods := []food.CreateFoodRequest{
		{Name: "Food 1", Calories: 100, Protein: 10, Carbs: 15, Fat: 5, Fiber: 2},
		{Name: "Food 2", Calories: 200, Protein: 20, Carbs: 25, Fat: 10, Fiber: 3},
		{Name: "Food 3", Calories: 300, Protein: 30, Carbs: 35, Fat: 15, Fiber: 4},
	}

	for _, f := range foods {
		_, err := repo.Create(f)
		require.NoError(t, err)
	}

	// Get all foods
	result, err := repo.GetAll()
	assert.NoError(t, err)
	assert.Len(t, result, 3, "Should return all 3 foods")
}

func TestRepository_Update(t *testing.T) {
	_, repo := setupFoodTest(t)

	// Create a food
	createReq := food.CreateFoodRequest{
		Name:     "Original Name",
		Calories: 100,
		Protein:  10,
	}
	created, err := repo.Create(createReq)
	require.NoError(t, err)

	// Update the food
	updateReq := food.UpdateFoodRequest{
		Name:     "Updated Name",
		Calories: 200,
		Protein:  20,
	}
	updated, err := repo.Update(int(created.ID), updateReq)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.InDelta(t, 200.0, updated.Calories, 0.01)
}

func TestRepository_Delete(t *testing.T) {
	_, repo := setupFoodTest(t)

	// Create a food
	req := food.CreateFoodRequest{
		Name:     "To Be Deleted",
		Calories: 100,
		Protein:  10,
	}
	created, err := repo.Create(req)
	require.NoError(t, err)

	// Delete the food
	err = repo.Delete(int(created.ID))
	assert.NoError(t, err)

	// Verify it's deleted
	found, err := repo.GetByID(int(created.ID))
	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestRepository_Create_WithTag(t *testing.T) {
	_, repo := setupFoodTest(t)

	req := food.CreateFoodRequest{
		Name:     "Routine Food",
		Calories: 150,
		Protein:  15,
		Tag:      "routine",
	}

	created, err := repo.Create(req)
	assert.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, "routine", created.Tag)
}

func TestRepository_Create_WithDefaultTag(t *testing.T) {
	_, repo := setupFoodTest(t)

	// Create food without specifying tag
	req := food.CreateFoodRequest{
		Name:     "Default Tag Food",
		Calories: 150,
		Protein:  15,
		// Tag is empty - should default to "routine"
	}

	created, err := repo.Create(req)
	assert.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, "routine", created.Tag, "Should default to routine when tag is not specified")
}

func TestRepository_GetByTag_Routine(t *testing.T) {
	_, repo := setupFoodTest(t)

	// Create multiple foods with different tags
	routineFood1 := food.CreateFoodRequest{Name: "Routine Food 1", Calories: 100, Protein: 10, Tag: "routine"}
	routineFood2 := food.CreateFoodRequest{Name: "Routine Food 2", Calories: 200, Protein: 20, Tag: "routine"}
	contextualFood := food.CreateFoodRequest{Name: "Contextual Food", Calories: 300, Protein: 30, Tag: "contextual"}

	_, err := repo.Create(routineFood1)
	require.NoError(t, err)
	_, err = repo.Create(routineFood2)
	require.NoError(t, err)
	_, err = repo.Create(contextualFood)
	require.NoError(t, err)

	// Get routine foods
	routineFoods, err := repo.GetByTag("routine")
	assert.NoError(t, err)
	assert.Len(t, routineFoods, 2, "Should return 2 routine foods")

	// Verify all returned foods have routine tag
	for _, f := range routineFoods {
		assert.Equal(t, "routine", f.Tag)
	}
}

func TestRepository_GetByTag_Contextual(t *testing.T) {
	_, repo := setupFoodTest(t)

	// Create multiple foods with different tags
	routineFood := food.CreateFoodRequest{Name: "Routine Food", Calories: 100, Protein: 10, Tag: "routine"}
	contextualFood1 := food.CreateFoodRequest{Name: "Contextual Food 1", Calories: 200, Protein: 20, Tag: "contextual"}
	contextualFood2 := food.CreateFoodRequest{Name: "Contextual Food 2", Calories: 300, Protein: 30, Tag: "contextual"}

	_, err := repo.Create(routineFood)
	require.NoError(t, err)
	_, err = repo.Create(contextualFood1)
	require.NoError(t, err)
	_, err = repo.Create(contextualFood2)
	require.NoError(t, err)

	// Get contextual foods
	contextualFoods, err := repo.GetByTag("contextual")
	assert.NoError(t, err)
	assert.Len(t, contextualFoods, 2, "Should return 2 contextual foods")

	// Verify all returned foods have contextual tag
	for _, f := range contextualFoods {
		assert.Equal(t, "contextual", f.Tag)
	}
}

func TestRepository_Update_ChangeTag(t *testing.T) {
	_, repo := setupFoodTest(t)

	// Create a food with routine tag
	createReq := food.CreateFoodRequest{
		Name:     "Food Item",
		Calories: 100,
		Protein:  10,
		Tag:      "routine",
	}
	created, err := repo.Create(createReq)
	require.NoError(t, err)
	assert.Equal(t, "routine", created.Tag)

	// Update the tag to contextual
	updateReq := food.UpdateFoodRequest{
		Name:     "Food Item",
		Calories: 100,
		Protein:  10,
		Tag:      "contextual",
	}
	updated, err := repo.Update(int(created.ID), updateReq)
	assert.NoError(t, err)
	assert.Equal(t, "contextual", updated.Tag)

	// Verify the tag was updated in the database
	found, err := repo.GetByID(int(created.ID))
	assert.NoError(t, err)
	assert.Equal(t, "contextual", found.Tag)
}

func TestValidateTag(t *testing.T) {
	assert.True(t, food.ValidateTag("routine"))
	assert.True(t, food.ValidateTag("contextual"))
	assert.False(t, food.ValidateTag("invalid"))
	assert.False(t, food.ValidateTag(""))
	assert.False(t, food.ValidateTag("ROUTINE"))
}

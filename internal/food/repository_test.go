package food

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"ultra-bis/test/testutil"
)

// setupFoodTest creates a test DB with Food migrations
func setupFoodTest(t *testing.T) (*gorm.DB, *Repository) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	if err := db.AutoMigrate(&Food{}); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}
	return db, NewRepository(db)
}

func TestRepository_Create(t *testing.T) {
	_, repo := setupFoodTest(t)

	req := CreateFoodRequest{
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
	req := CreateFoodRequest{
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
	foods := []CreateFoodRequest{
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
	createReq := CreateFoodRequest{
		Name:     "Original Name",
		Calories: 100,
		Protein:  10,
	}
	created, err := repo.Create(createReq)
	require.NoError(t, err)

	// Update the food
	updateReq := UpdateFoodRequest{
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
	req := CreateFoodRequest{
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

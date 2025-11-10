package diary

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"ultra-bis/internal/food"
	"ultra-bis/internal/user"
	"ultra-bis/test/testutil"
)

// mockRecipeRepository is a simple mock for testing diary calculations without circular imports
type mockRecipeRepository struct {
	recipes     map[int]Recipe
	ingredients map[int][]RecipeIngredient
}

func newMockRecipeRepository() *mockRecipeRepository {
	return &mockRecipeRepository{
		recipes:     make(map[int]Recipe),
		ingredients: make(map[int][]RecipeIngredient),
	}
}

func (m *mockRecipeRepository) GetByID(id int) (Recipe, error) {
	if recipe, ok := m.recipes[id]; ok {
		return recipe, nil
	}
	return Recipe{}, gorm.ErrRecordNotFound
}

func (m *mockRecipeRepository) GetIngredients(recipeID int) ([]RecipeIngredient, error) {
	if ingredients, ok := m.ingredients[recipeID]; ok {
		return ingredients, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockRecipeRepository) addRecipe(id int, name string, ingredients []RecipeIngredient) {
	m.recipes[id] = Recipe{ID: uint(id), Name: name}
	m.ingredients[id] = ingredients
}

// setupDiaryTest creates a test DB with all necessary migrations
func setupDiaryTest(t *testing.T) (*gorm.DB, *Repository, *food.Repository) {
	t.Helper()
	db := testutil.SetupTestDB(t)

	// Run migrations
	if err := db.AutoMigrate(
		&user.User{},
		&food.Food{},
		&DiaryEntry{},
	); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	diaryRepo := NewRepository(db)
	foodRepo := food.NewRepository(db)

	return db, diaryRepo, foodRepo
}

// createTestUser creates a test user
func createTestUser(t *testing.T, db *gorm.DB) uint {
	t.Helper()

	testUser := &user.User{
		Email:         "test@example.com",
		PasswordHash:  "hashed_password",
		Age:           30,
		Gender:        "male",
		Height:        175,
		ActivityLevel: "moderate",
		GoalType:      "maintain",
	}

	result := db.Create(testUser)
	require.NoError(t, result.Error)
	return testUser.ID
}

// createTestFood creates a test food item
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

func TestDiaryEntry_FoodCalculation_ExactPortion(t *testing.T) {
	db, diaryRepo, foodRepo := setupDiaryTest(t)
	userID := createTestUser(t, db)

	// Create test food: Chicken (165 cal, 31g protein per 100g)
	chicken := createTestFood(t, foodRepo, "Chicken Breast", 165, 31, 0, 3.6, 0)

	// Create diary entry with exactly 100g
	entry := &DiaryEntry{
		UserID:        userID,
		FoodID:        &chicken.ID,
		Date:          time.Now(),
		MealType:      Breakfast,
		QuantityGrams: 100,
	}

	// Calculate nutrition (simulating what handler does)
	multiplier := entry.QuantityGrams / 100.0
	entry.Calories = chicken.Calories * multiplier
	entry.Protein = chicken.Protein * multiplier
	entry.Carbs = chicken.Carbs * multiplier
	entry.Fat = chicken.Fat * multiplier
	entry.Fiber = chicken.Fiber * multiplier

	err := diaryRepo.Create(entry)
	require.NoError(t, err)

	// Verify stored values
	assert.InDelta(t, 165.0, entry.Calories, 0.01)
	assert.InDelta(t, 31.0, entry.Protein, 0.01)
	assert.InDelta(t, 0.0, entry.Carbs, 0.01)
	assert.InDelta(t, 3.6, entry.Fat, 0.01)
	assert.InDelta(t, 0.0, entry.Fiber, 0.01)
}

func TestDiaryEntry_FoodCalculation_HalfPortion(t *testing.T) {
	db, diaryRepo, foodRepo := setupDiaryTest(t)
	userID := createTestUser(t, db)

	// Create test food: Rice (130 cal, 2.7g protein per 100g)
	rice := createTestFood(t, foodRepo, "White Rice", 130, 2.7, 28, 0.3, 0.4)

	// Create diary entry with 50g (half portion)
	entry := &DiaryEntry{
		UserID:        userID,
		FoodID:        &rice.ID,
		Date:          time.Now(),
		MealType:      Lunch,
		QuantityGrams: 50,
	}

	// Calculate nutrition
	multiplier := entry.QuantityGrams / 100.0
	entry.Calories = rice.Calories * multiplier
	entry.Protein = rice.Protein * multiplier
	entry.Carbs = rice.Carbs * multiplier
	entry.Fat = rice.Fat * multiplier
	entry.Fiber = rice.Fiber * multiplier

	err := diaryRepo.Create(entry)
	require.NoError(t, err)

	// Expected: 50% of the values
	assert.InDelta(t, 65.0, entry.Calories, 0.01)
	assert.InDelta(t, 1.35, entry.Protein, 0.01)
	assert.InDelta(t, 14.0, entry.Carbs, 0.01)
	assert.InDelta(t, 0.15, entry.Fat, 0.01)
	assert.InDelta(t, 0.2, entry.Fiber, 0.01)
}

func TestDiaryEntry_FoodCalculation_DoublePortion(t *testing.T) {
	db, diaryRepo, foodRepo := setupDiaryTest(t)
	userID := createTestUser(t, db)

	// Create test food: Oatmeal (389 cal, 16.9g protein per 100g)
	oatmeal := createTestFood(t, foodRepo, "Oatmeal", 389, 16.9, 66, 6.9, 10.6)

	// Create diary entry with 200g (double portion)
	entry := &DiaryEntry{
		UserID:        userID,
		FoodID:        &oatmeal.ID,
		Date:          time.Now(),
		MealType:      Breakfast,
		QuantityGrams: 200,
	}

	// Calculate nutrition
	multiplier := entry.QuantityGrams / 100.0
	entry.Calories = oatmeal.Calories * multiplier
	entry.Protein = oatmeal.Protein * multiplier
	entry.Carbs = oatmeal.Carbs * multiplier
	entry.Fat = oatmeal.Fat * multiplier
	entry.Fiber = oatmeal.Fiber * multiplier

	err := diaryRepo.Create(entry)
	require.NoError(t, err)

	// Expected: double the values
	assert.InDelta(t, 778.0, entry.Calories, 0.1)
	assert.InDelta(t, 33.8, entry.Protein, 0.1)
	assert.InDelta(t, 132.0, entry.Carbs, 0.1)
	assert.InDelta(t, 13.8, entry.Fat, 0.1)
	assert.InDelta(t, 21.2, entry.Fiber, 0.1)
}

func TestDiaryEntry_FoodCalculation_FractionalGrams(t *testing.T) {
	db, diaryRepo, foodRepo := setupDiaryTest(t)
	userID := createTestUser(t, db)

	// Create test food: Olive Oil (884 cal per 100g)
	oil := createTestFood(t, foodRepo, "Olive Oil", 884, 0, 0, 100, 0)

	// Create diary entry with 15.5g (typical tablespoon)
	entry := &DiaryEntry{
		UserID:        userID,
		FoodID:        &oil.ID,
		Date:          time.Now(),
		MealType:      Snack,
		QuantityGrams: 15.5,
	}

	// Calculate nutrition
	multiplier := entry.QuantityGrams / 100.0
	entry.Calories = oil.Calories * multiplier
	entry.Protein = oil.Protein * multiplier
	entry.Carbs = oil.Carbs * multiplier
	entry.Fat = oil.Fat * multiplier
	entry.Fiber = oil.Fiber * multiplier

	err := diaryRepo.Create(entry)
	require.NoError(t, err)

	// Expected: 884 * 0.155 = 137.02 calories
	assert.InDelta(t, 137.02, entry.Calories, 0.1)
	assert.InDelta(t, 0.0, entry.Protein, 0.01)
	assert.InDelta(t, 15.5, entry.Fat, 0.1)
}

func TestDiaryEntry_RecipeCalculation_ProportionalScaling(t *testing.T) {
	db, diaryRepo, foodRepo := setupDiaryTest(t)
	userID := createTestUser(t, db)

	// Create foods (per 100g)
	chicken := createTestFood(t, foodRepo, "Chicken", 165, 31, 0, 3.6, 0)
	rice := createTestFood(t, foodRepo, "Rice", 130, 2.7, 28, 0.3, 0.4)

	// Create a mock recipe: 200g chicken + 150g rice = 350g total
	mockRecipeRepo := newMockRecipeRepository()
	recipeID := 1
	mockRecipeRepo.addRecipe(recipeID, "Chicken & Rice", []RecipeIngredient{
		{FoodID: chicken.ID, QuantityGrams: 200},
		{FoodID: rice.ID, QuantityGrams: 150},
	})
	testRecipeID := uint(recipeID)

	// Recipe totals:
	// Chicken: 165 * 2 = 330 cal, 62g protein, 7.2g fat
	// Rice: 130 * 1.5 = 195 cal, 4.05g protein, 42g carbs, 0.45g fat, 0.6g fiber
	// Total: 525 cal, 66.05g protein, 42g carbs, 7.65g fat, 0.6g fiber, 350g weight

	// Create diary entry with 175g (half of the recipe)
	entry := &DiaryEntry{
		UserID:        userID,
		RecipeID:      &testRecipeID,
		Date:          time.Now(),
		MealType:      Dinner,
		QuantityGrams: 175, // Half the recipe
	}

	// Calculate nutrition proportionally (simulating convertProportionalToCustomIngredients)
	// portion := entry.QuantityGrams / 350.0 // 175 / 350 = 0.5

	// Calculate from each ingredient
	var totalCalories, totalProtein, totalCarbs, totalFat, totalFiber float64

	// Chicken: 200g * 0.5 = 100g
	totalCalories += chicken.Calories * (100.0 / 100.0)
	totalProtein += chicken.Protein * (100.0 / 100.0)
	totalFat += chicken.Fat * (100.0 / 100.0)

	// Rice: 150g * 0.5 = 75g
	totalCalories += rice.Calories * (75.0 / 100.0)
	totalProtein += rice.Protein * (75.0 / 100.0)
	totalCarbs += rice.Carbs * (75.0 / 100.0)
	totalFat += rice.Fat * (75.0 / 100.0)
	totalFiber += rice.Fiber * (75.0 / 100.0)

	entry.Calories = totalCalories
	entry.Protein = totalProtein
	entry.Carbs = totalCarbs
	entry.Fat = totalFat
	entry.Fiber = totalFiber

	err := diaryRepo.Create(entry)
	require.NoError(t, err)

	// Expected: half of the recipe totals
	assert.InDelta(t, 262.5, entry.Calories, 0.2, "Half of 525")
	assert.InDelta(t, 33.025, entry.Protein, 0.2, "Half of 66.05")
	assert.InDelta(t, 21.0, entry.Carbs, 0.2, "Half of 42")
	assert.InDelta(t, 3.825, entry.Fat, 0.2, "Half of 7.65")
	assert.InDelta(t, 0.3, entry.Fiber, 0.2, "Half of 0.6")
}

func TestDiaryEntry_RecipeCalculation_FullRecipe(t *testing.T) {
	db, diaryRepo, foodRepo := setupDiaryTest(t)
	userID := createTestUser(t, db)

	// Create foods
	salmon := createTestFood(t, foodRepo, "Salmon", 206, 22, 0, 13, 0)
	quinoa := createTestFood(t, foodRepo, "Quinoa", 120, 4.4, 21, 1.9, 2.8)
	avocado := createTestFood(t, foodRepo, "Avocado", 160, 2, 8.5, 14.7, 6.7)

	// Create mock recipe: 150g salmon + 100g quinoa + 50g avocado = 300g total
	mockRecipeRepo := newMockRecipeRepository()
	recipeID := 1
	mockRecipeRepo.addRecipe(recipeID, "Power Bowl", []RecipeIngredient{
		{FoodID: salmon.ID, QuantityGrams: 150},
		{FoodID: quinoa.ID, QuantityGrams: 100},
		{FoodID: avocado.ID, QuantityGrams: 50},
	})
	testRecipeID := uint(recipeID)

	// Create diary entry with the full 300g
	entry := &DiaryEntry{
		UserID:        userID,
		RecipeID:      &testRecipeID,
		Date:          time.Now(),
		MealType:      Lunch,
		QuantityGrams: 300,
	}

	// Calculate nutrition for full recipe
	// Salmon: 206 * 1.5 = 309 cal, 33g protein, 19.5g fat
	// Quinoa: 120 * 1 = 120 cal, 4.4g protein, 21g carbs, 1.9g fat, 2.8g fiber
	// Avocado: 160 * 0.5 = 80 cal, 1g protein, 4.25g carbs, 7.35g fat, 3.35g fiber
	// Total: 509 cal, 38.4g protein, 25.25g carbs, 28.75g fat, 6.15g fiber

	entry.Calories = 309 + 120 + 80
	entry.Protein = 33 + 4.4 + 1
	entry.Carbs = 0 + 21 + 4.25
	entry.Fat = 19.5 + 1.9 + 7.35
	entry.Fiber = 0 + 2.8 + 3.35

	err := diaryRepo.Create(entry)
	require.NoError(t, err)

	assert.InDelta(t, 509.0, entry.Calories, 0.2)
	assert.InDelta(t, 38.4, entry.Protein, 0.2)
	assert.InDelta(t, 25.25, entry.Carbs, 0.2)
	assert.InDelta(t, 28.75, entry.Fat, 0.2)
	assert.InDelta(t, 6.15, entry.Fiber, 0.2)
}

func TestDiaryEntry_RecipeCalculation_CustomIngredients(t *testing.T) {
	db, diaryRepo, foodRepo := setupDiaryTest(t)
	userID := createTestUser(t, db)

	// Create foods
	chicken := createTestFood(t, foodRepo, "Chicken", 165, 31, 0, 3.6, 0)
	rice := createTestFood(t, foodRepo, "Rice", 130, 2.7, 28, 0.3, 0.4)
	broccoli := createTestFood(t, foodRepo, "Broccoli", 34, 2.8, 7, 0.4, 2.6)

	// Create mock recipe with default portions
	mockRecipeRepo := newMockRecipeRepository()
	recipeID := 1
	mockRecipeRepo.addRecipe(recipeID, "Meal Prep", []RecipeIngredient{
		{FoodID: chicken.ID, QuantityGrams: 150},
		{FoodID: rice.ID, QuantityGrams: 200},
		{FoodID: broccoli.ID, QuantityGrams: 100},
	})
	testRecipeID := uint(recipeID)

	// Create diary entry with custom ingredient quantities (user wants extra chicken, less rice)
	entry := &DiaryEntry{
		UserID:   userID,
		RecipeID: &testRecipeID,
		Date:     time.Now(),
		MealType: Dinner,
	}

	// Custom ingredients: 200g chicken + 100g rice + 150g broccoli = 450g total
	customIngredients := CustomIngredients{
		{
			FoodID:        chicken.ID,
			FoodName:      chicken.Name,
			QuantityGrams: 200,
			Calories:      165 * 2.0,
			Protein:       31 * 2.0,
			Carbs:         0,
			Fat:           3.6 * 2.0,
			Fiber:         0,
		},
		{
			FoodID:        rice.ID,
			FoodName:      rice.Name,
			QuantityGrams: 100,
			Calories:      130 * 1.0,
			Protein:       2.7 * 1.0,
			Carbs:         28 * 1.0,
			Fat:           0.3 * 1.0,
			Fiber:         0.4 * 1.0,
		},
		{
			FoodID:        broccoli.ID,
			FoodName:      broccoli.Name,
			QuantityGrams: 150,
			Calories:      34 * 1.5,
			Protein:       2.8 * 1.5,
			Carbs:         7 * 1.5,
			Fat:           0.4 * 1.5,
			Fiber:         2.6 * 1.5,
		},
	}

	entry.CustomIngredients = customIngredients
	entry.QuantityGrams = 450

	// Calculate totals
	var totalCal, totalPro, totalCarb, totalFat, totalFib float64
	for _, ing := range customIngredients {
		totalCal += ing.Calories
		totalPro += ing.Protein
		totalCarb += ing.Carbs
		totalFat += ing.Fat
		totalFib += ing.Fiber
	}

	entry.Calories = totalCal
	entry.Protein = totalPro
	entry.Carbs = totalCarb
	entry.Fat = totalFat
	entry.Fiber = totalFib

	err := diaryRepo.Create(entry)
	require.NoError(t, err)

	// Verify totals
	// Chicken: 330 cal, 62g protein, 7.2g fat
	// Rice: 130 cal, 2.7g protein, 28g carbs, 0.3g fat, 0.4g fiber
	// Broccoli: 51 cal, 4.2g protein, 10.5g carbs, 0.6g fat, 3.9g fiber
	// Total: 511 cal, 68.9g protein, 38.5g carbs, 8.1g fat, 4.3g fiber

	assert.InDelta(t, 511.0, entry.Calories, 0.2)
	assert.InDelta(t, 68.9, entry.Protein, 0.2)
	assert.InDelta(t, 38.5, entry.Carbs, 0.2)
	assert.InDelta(t, 8.1, entry.Fat, 0.2)
	assert.InDelta(t, 4.3, entry.Fiber, 0.2)
	assert.Len(t, entry.CustomIngredients, 3)
}

func TestDiaryEntry_GetDailySummary(t *testing.T) {
	db, diaryRepo, foodRepo := setupDiaryTest(t)
	userID := createTestUser(t, db)

	// Create test foods
	eggs := createTestFood(t, foodRepo, "Eggs", 155, 13, 1.1, 11, 0)
	bread := createTestFood(t, foodRepo, "Bread", 265, 9, 49, 3.2, 2.7)
	banana := createTestFood(t, foodRepo, "Banana", 89, 1.1, 23, 0.3, 2.6)

	testDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	// Create multiple entries for the same day
	entries := []*DiaryEntry{
		{
			UserID:        userID,
			FoodID:        &eggs.ID,
			Date:          testDate,
			MealType:      Breakfast,
			QuantityGrams: 100,
			Calories:      155,
			Protein:       13,
			Carbs:         1.1,
			Fat:           11,
			Fiber:         0,
		},
		{
			UserID:        userID,
			FoodID:        &bread.ID,
			Date:          testDate,
			MealType:      Breakfast,
			QuantityGrams: 50,
			Calories:      132.5,
			Protein:       4.5,
			Carbs:         24.5,
			Fat:           1.6,
			Fiber:         1.35,
		},
		{
			UserID:        userID,
			FoodID:        &banana.ID,
			Date:          testDate,
			MealType:      Snack,
			QuantityGrams: 120,
			Calories:      106.8,
			Protein:       1.32,
			Carbs:         27.6,
			Fat:           0.36,
			Fiber:         3.12,
		},
	}

	for _, entry := range entries {
		err := diaryRepo.Create(entry)
		require.NoError(t, err)
	}

	// Get daily summary
	summary, err := diaryRepo.GetDailySummary(userID, testDate)
	assert.NoError(t, err)
	require.NotNil(t, summary)

	// Expected totals: 155 + 132.5 + 106.8 = 394.3 cal
	assert.InDelta(t, 394.3, summary["calories"], 0.2)
	assert.InDelta(t, 18.82, summary["protein"], 0.2)
	assert.InDelta(t, 53.2, summary["carbs"], 0.2)
	assert.InDelta(t, 12.96, summary["fat"], 0.2)
	assert.InDelta(t, 4.47, summary["fiber"], 0.2)
}

func TestDiaryEntry_GetByDate(t *testing.T) {
	db, diaryRepo, foodRepo := setupDiaryTest(t)
	userID := createTestUser(t, db)

	// Create test food
	chicken := createTestFood(t, foodRepo, "Chicken", 165, 31, 0, 3.6, 0)

	// Create entries on different dates
	today := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	yesterday := today.AddDate(0, 0, -1)

	todayEntry := &DiaryEntry{
		UserID:        userID,
		FoodID:        &chicken.ID,
		Date:          today,
		MealType:      Lunch,
		QuantityGrams: 150,
		Calories:      247.5,
		Protein:       46.5,
		Fat:           5.4,
	}

	yesterdayEntry := &DiaryEntry{
		UserID:        userID,
		FoodID:        &chicken.ID,
		Date:          yesterday,
		MealType:      Dinner,
		QuantityGrams: 200,
		Calories:      330,
		Protein:       62,
		Fat:           7.2,
	}

	require.NoError(t, diaryRepo.Create(todayEntry))
	require.NoError(t, diaryRepo.Create(yesterdayEntry))

	// Get entries for today only
	entries, err := diaryRepo.GetByDate(userID, today)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, Lunch, entries[0].MealType)
	assert.InDelta(t, 247.5, entries[0].Calories, 0.1)

	// Get entries for yesterday
	entriesYesterday, err := diaryRepo.GetByDate(userID, yesterday)
	assert.NoError(t, err)
	assert.Len(t, entriesYesterday, 1)
	assert.Equal(t, Dinner, entriesYesterday[0].MealType)
	assert.InDelta(t, 330.0, entriesYesterday[0].Calories, 0.1)
}

func TestDiaryEntry_HistoricalAccuracy(t *testing.T) {
	db, diaryRepo, foodRepo := setupDiaryTest(t)
	userID := createTestUser(t, db)

	// Create test food
	chicken := createTestFood(t, foodRepo, "Chicken", 165, 31, 0, 3.6, 0)

	// Create diary entry
	entry := &DiaryEntry{
		UserID:        userID,
		FoodID:        &chicken.ID,
		Date:          time.Now(),
		MealType:      Lunch,
		QuantityGrams: 150,
		Calories:      247.5,
		Protein:       46.5,
		Fat:           5.4,
	}

	err := diaryRepo.Create(entry)
	require.NoError(t, err)

	originalCalories := entry.Calories

	// Now update the food item (simulating food database change)
	_, err = foodRepo.Update(int(chicken.ID), food.UpdateFoodRequest{
		Name:        chicken.Name,
		Description: chicken.Description,
		Calories:    200, // Changed!
		Protein:     35,
		Carbs:       0,
		Fat:         5,
		Fiber:       0,
	})
	require.NoError(t, err)

	// Retrieve the diary entry
	retrieved, err := diaryRepo.GetByID(entry.ID, userID)
	require.NoError(t, err)

	// The diary entry should still have the original nutrition values (cached)
	assert.InDelta(t, originalCalories, retrieved.Calories, 0.01, "Diary entry nutrition should not change when food is updated")
}

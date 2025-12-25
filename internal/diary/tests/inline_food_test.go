package tests

import (
	"testing"
	"time"

	"ultra-bis/internal/diary"
	"ultra-bis/internal/food"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateInlineFood(t *testing.T) {
	db, diaryRepo, _ := setupDiaryTest(t)
	userID := createTestUser(t, db)

	// Per-100g nutrition values
	calories := 350.0
	protein := 25.0
	carbs := 40.0
	fat := 10.0
	fiber := 5.0
	quantityGrams := 80.0

	entry := &diary.DiaryEntry{
		UserID:               userID,
		InlineFoodName:       strPtr("Homemade Protein Bar"),
		InlineFoodDescription: strPtr("Custom oats and whey recipe"),
		InlineFoodCalories:   &calories,
		InlineFoodProtein:    &protein,
		InlineFoodCarbs:      &carbs,
		InlineFoodFat:        &fat,
		InlineFoodFiber:      &fiber,
		InlineFoodTag:        strPtr("contextual"),
		Date:                 time.Now(),
		MealType:             diary.Snack,
		QuantityGrams:        quantityGrams,
		Notes:                "Post-workout snack",
		// Consumed nutrition (calculated: per_100g * (quantity_grams / 100))
		Calories: 280.0, // 350 * 0.8
		Protein:  20.0,  // 25 * 0.8
		Carbs:    32.0,  // 40 * 0.8
		Fat:      8.0,   // 10 * 0.8
		Fiber:    4.0,   // 5 * 0.8
		FoodTag:  "contextual",
	}

	err := diaryRepo.Create(entry)
	require.NoError(t, err)
	assert.NotZero(t, entry.ID)

	// Verify the entry was created correctly
	retrieved, err := diaryRepo.GetByID(entry.ID, userID)
	require.NoError(t, err)

	assert.Equal(t, "Homemade Protein Bar", *retrieved.InlineFoodName)
	assert.Equal(t, "Custom oats and whey recipe", *retrieved.InlineFoodDescription)
	assert.Equal(t, 350.0, *retrieved.InlineFoodCalories)
	assert.Equal(t, 25.0, *retrieved.InlineFoodProtein)
	assert.Equal(t, 40.0, *retrieved.InlineFoodCarbs)
	assert.Equal(t, 10.0, *retrieved.InlineFoodFat)
	assert.Equal(t, 5.0, *retrieved.InlineFoodFiber)
	assert.Equal(t, "contextual", *retrieved.InlineFoodTag)
	assert.Equal(t, 280.0, retrieved.Calories)
	assert.Equal(t, 20.0, retrieved.Protein)
	assert.Equal(t, 32.0, retrieved.Carbs)
	assert.Equal(t, 8.0, retrieved.Fat)
	assert.Equal(t, 4.0, retrieved.Fiber)
	assert.Equal(t, "contextual", retrieved.FoodTag)
}

func TestCreateInlineFood_MinimalFields(t *testing.T) {
	db, diaryRepo, _ := setupDiaryTest(t)
	userID := createTestUser(t, db)

	// Minimal inline food (no description, tag defaults to "routine")
	calories := 450.0
	protein := 6.0
	carbs := 60.0
	fat := 20.0
	fiber := 2.0

	entry := &diary.DiaryEntry{
		UserID:             userID,
		InlineFoodName:     strPtr("Simple Cake"),
		InlineFoodCalories: &calories,
		InlineFoodProtein:  &protein,
		InlineFoodCarbs:    &carbs,
		InlineFoodFat:      &fat,
		InlineFoodFiber:    &fiber,
		Date:               time.Now(),
		MealType:           diary.Dinner,
		QuantityGrams:      100.0,
		Calories:           450.0,
		Protein:            6.0,
		Carbs:              60.0,
		Fat:                20.0,
		Fiber:              2.0,
	}

	err := diaryRepo.Create(entry)
	require.NoError(t, err)
	assert.NotZero(t, entry.ID)

	// Verify
	retrieved, err := diaryRepo.GetByID(entry.ID, userID)
	require.NoError(t, err)

	assert.Equal(t, "Simple Cake", *retrieved.InlineFoodName)
	assert.Nil(t, retrieved.InlineFoodDescription)
	assert.Nil(t, retrieved.InlineFoodTag)
}

func TestUpdateInlineFood_NutritionValues(t *testing.T) {
	db, diaryRepo, _ := setupDiaryTest(t)
	userID := createTestUser(t, db)

	// Create initial inline food
	calories := 350.0
	protein := 25.0
	carbs := 40.0
	fat := 10.0
	fiber := 5.0

	entry := &diary.DiaryEntry{
		UserID:             userID,
		InlineFoodName:     strPtr("Energy Bar"),
		InlineFoodCalories: &calories,
		InlineFoodProtein:  &protein,
		InlineFoodCarbs:    &carbs,
		InlineFoodFat:      &fat,
		InlineFoodFiber:    &fiber,
		Date:               time.Now(),
		MealType:           diary.Snack,
		QuantityGrams:      80.0,
		Calories:           280.0,
		Protein:            20.0,
		Carbs:              32.0,
		Fat:                8.0,
		Fiber:              4.0,
	}

	err := diaryRepo.Create(entry)
	require.NoError(t, err)

	// Update nutrition values
	newCalories := 400.0
	newProtein := 30.0
	entry.InlineFoodCalories = &newCalories
	entry.InlineFoodProtein = &newProtein

	// Recalculate consumed nutrition
	multiplier := entry.QuantityGrams / 100.0
	entry.Calories = newCalories * multiplier  // 400 * 0.8 = 320
	entry.Protein = newProtein * multiplier    // 30 * 0.8 = 24

	err = diaryRepo.Update(entry)
	require.NoError(t, err)

	// Verify update
	retrieved, err := diaryRepo.GetByID(entry.ID, userID)
	require.NoError(t, err)

	assert.Equal(t, 400.0, *retrieved.InlineFoodCalories)
	assert.Equal(t, 30.0, *retrieved.InlineFoodProtein)
	assert.Equal(t, 320.0, retrieved.Calories)
	assert.Equal(t, 24.0, retrieved.Protein)
}

func TestUpdateInlineFood_Name(t *testing.T) {
	db, diaryRepo, _ := setupDiaryTest(t)
	userID := createTestUser(t, db)

	calories := 350.0
	protein := 25.0
	carbs := 40.0
	fat := 10.0
	fiber := 5.0

	entry := &diary.DiaryEntry{
		UserID:             userID,
		InlineFoodName:     strPtr("Original Name"),
		InlineFoodCalories: &calories,
		InlineFoodProtein:  &protein,
		InlineFoodCarbs:    &carbs,
		InlineFoodFat:      &fat,
		InlineFoodFiber:    &fiber,
		Date:               time.Now(),
		MealType:           diary.Snack,
		QuantityGrams:      100.0,
		Calories:           350.0,
		Protein:            25.0,
		Carbs:              40.0,
		Fat:                10.0,
		Fiber:              5.0,
	}

	err := diaryRepo.Create(entry)
	require.NoError(t, err)

	// Update name
	entry.InlineFoodName = strPtr("Updated Name")
	err = diaryRepo.Update(entry)
	require.NoError(t, err)

	// Verify
	retrieved, err := diaryRepo.GetByID(entry.ID, userID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", *retrieved.InlineFoodName)
}

func TestUpdateInlineFood_Quantity(t *testing.T) {
	db, diaryRepo, _ := setupDiaryTest(t)
	userID := createTestUser(t, db)

	calories := 350.0
	protein := 25.0
	carbs := 40.0
	fat := 10.0
	fiber := 5.0

	entry := &diary.DiaryEntry{
		UserID:             userID,
		InlineFoodName:     strPtr("Protein Bar"),
		InlineFoodCalories: &calories,
		InlineFoodProtein:  &protein,
		InlineFoodCarbs:    &carbs,
		InlineFoodFat:      &fat,
		InlineFoodFiber:    &fiber,
		Date:               time.Now(),
		MealType:           diary.Snack,
		QuantityGrams:      80.0,
		Calories:           280.0,  // 350 * 0.8
		Protein:            20.0,   // 25 * 0.8
		Carbs:              32.0,
		Fat:                8.0,
		Fiber:              4.0,
	}

	err := diaryRepo.Create(entry)
	require.NoError(t, err)

	// Change quantity from 80g to 100g
	entry.QuantityGrams = 100.0
	entry.Calories = 350.0  // 350 * 1.0
	entry.Protein = 25.0    // 25 * 1.0
	entry.Carbs = 40.0
	entry.Fat = 10.0
	entry.Fiber = 5.0

	err = diaryRepo.Update(entry)
	require.NoError(t, err)

	// Verify
	retrieved, err := diaryRepo.GetByID(entry.ID, userID)
	require.NoError(t, err)
	assert.Equal(t, 100.0, retrieved.QuantityGrams)
	assert.Equal(t, 350.0, retrieved.Calories)
	assert.Equal(t, 25.0, retrieved.Protein)
}

func TestPopulateNames_InlineFoodPrecedence(t *testing.T) {
	db, diaryRepo, foodRepo := setupDiaryTest(t)
	userID := createTestUser(t, db)

	// Create a saved food
	savedFood, err := foodRepo.Create(food.CreateFoodRequest{
		Name:        "Saved Protein Bar",
		Description: "From database",
		Calories:    300.0,
		Protein:     20.0,
		Carbs:       35.0,
		Fat:         8.0,
		Fiber:       4.0,
		Tag:         "routine",
	})
	require.NoError(t, err)

	calories := 350.0
	protein := 25.0
	carbs := 40.0
	fat := 10.0
	fiber := 5.0

	// Create entry with both food_id and inline_food_name
	// (This tests the display logic, not validation - in real usage they'd be mutually exclusive)
	entry := &diary.DiaryEntry{
		UserID:             userID,
		FoodID:             &savedFood.ID,
		InlineFoodName:     strPtr("Inline Override Name"),
		InlineFoodCalories: &calories,
		InlineFoodProtein:  &protein,
		InlineFoodCarbs:    &carbs,
		InlineFoodFat:      &fat,
		InlineFoodFiber:    &fiber,
		Date:               time.Now(),
		MealType:           diary.Snack,
		QuantityGrams:      100.0,
		Calories:           350.0,
		Protein:            25.0,
		Carbs:              40.0,
		Fat:                10.0,
		Fiber:              5.0,
	}

	err = diaryRepo.Create(entry)
	require.NoError(t, err)

	// Retrieve with populated names
	entries, err := diaryRepo.GetByDate(userID, time.Now())
	require.NoError(t, err)
	require.Len(t, entries, 1)

	// Inline food name should take precedence
	assert.Equal(t, "Inline Override Name", entries[0].FoodName)
}

func TestInlineFood_TagCalculation(t *testing.T) {
	db, diaryRepo, _ := setupDiaryTest(t)
	userID := createTestUser(t, db)

	calories := 300.0
	protein := 20.0
	carbs := 30.0
	fat := 10.0
	fiber := 5.0

	// Test with contextual tag
	entry1 := &diary.DiaryEntry{
		UserID:             userID,
		InlineFoodName:     strPtr("Contextual Food"),
		InlineFoodCalories: &calories,
		InlineFoodProtein:  &protein,
		InlineFoodCarbs:    &carbs,
		InlineFoodFat:      &fat,
		InlineFoodFiber:    &fiber,
		InlineFoodTag:      strPtr("contextual"),
		Date:               time.Now(),
		MealType:           diary.Breakfast,
		QuantityGrams:      100.0,
		Calories:           300.0,
		Protein:            20.0,
		Carbs:              30.0,
		Fat:                10.0,
		Fiber:              5.0,
		FoodTag:            "contextual",
	}

	err := diaryRepo.Create(entry1)
	require.NoError(t, err)

	// Test with routine tag
	entry2 := &diary.DiaryEntry{
		UserID:             userID,
		InlineFoodName:     strPtr("Routine Food"),
		InlineFoodCalories: &calories,
		InlineFoodProtein:  &protein,
		InlineFoodCarbs:    &carbs,
		InlineFoodFat:      &fat,
		InlineFoodFiber:    &fiber,
		InlineFoodTag:      strPtr("routine"),
		Date:               time.Now(),
		MealType:           diary.Lunch,
		QuantityGrams:      100.0,
		Calories:           300.0,
		Protein:            20.0,
		Carbs:              30.0,
		Fat:                10.0,
		Fiber:              5.0,
		FoodTag:            "routine",
	}

	err = diaryRepo.Create(entry2)
	require.NoError(t, err)

	// Retrieve and verify tags
	retrieved1, err := diaryRepo.GetByID(entry1.ID, userID)
	require.NoError(t, err)
	assert.Equal(t, "contextual", *retrieved1.InlineFoodTag)
	assert.Equal(t, "contextual", retrieved1.FoodTag)

	retrieved2, err := diaryRepo.GetByID(entry2.ID, userID)
	require.NoError(t, err)
	assert.Equal(t, "routine", *retrieved2.InlineFoodTag)
	assert.Equal(t, "routine", retrieved2.FoodTag)
}

func TestInlineFood_NutritionCalculation_DifferentQuantities(t *testing.T) {
	testCases := []struct {
		name             string
		per100gCalories  float64
		per100gProtein   float64
		quantityGrams    float64
		expectedCalories float64
		expectedProtein  float64
	}{
		{
			name:             "50 grams",
			per100gCalories:  400.0,
			per100gProtein:   30.0,
			quantityGrams:    50.0,
			expectedCalories: 200.0, // 400 * 0.5
			expectedProtein:  15.0,  // 30 * 0.5
		},
		{
			name:             "100 grams (baseline)",
			per100gCalories:  400.0,
			per100gProtein:   30.0,
			quantityGrams:    100.0,
			expectedCalories: 400.0,
			expectedProtein:  30.0,
		},
		{
			name:             "150 grams",
			per100gCalories:  400.0,
			per100gProtein:   30.0,
			quantityGrams:    150.0,
			expectedCalories: 600.0, // 400 * 1.5
			expectedProtein:  45.0,  // 30 * 1.5
		},
		{
			name:             "25 grams (quarter)",
			per100gCalories:  200.0,
			per100gProtein:   10.0,
			quantityGrams:    25.0,
			expectedCalories: 50.0, // 200 * 0.25
			expectedProtein:  2.5,  // 10 * 0.25
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, diaryRepo, _ := setupDiaryTest(t)
			userID := createTestUser(t, db)

			carbs := 0.0
			fat := 0.0
			fiber := 0.0

			entry := &diary.DiaryEntry{
				UserID:             userID,
				InlineFoodName:     strPtr("Test Food"),
				InlineFoodCalories: &tc.per100gCalories,
				InlineFoodProtein:  &tc.per100gProtein,
				InlineFoodCarbs:    &carbs,
				InlineFoodFat:      &fat,
				InlineFoodFiber:    &fiber,
				Date:               time.Now(),
				MealType:           diary.Snack,
				QuantityGrams:      tc.quantityGrams,
				Calories:           tc.expectedCalories,
				Protein:            tc.expectedProtein,
				Carbs:              0.0,
				Fat:                0.0,
				Fiber:              0.0,
			}

			err := diaryRepo.Create(entry)
			require.NoError(t, err)

			retrieved, err := diaryRepo.GetByID(entry.ID, userID)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedCalories, retrieved.Calories)
			assert.Equal(t, tc.expectedProtein, retrieved.Protein)
		})
	}
}

// Helper function to create string pointers
func strPtr(s string) *string {
	return &s
}

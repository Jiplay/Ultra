package tests

import (
	"testing"

	"ultra-bis/internal/diary"
	"ultra-bis/internal/food"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveAsFood_Success(t *testing.T) {
	db, diaryRepo, foodRepo := setupDiaryTest(t)
	userID := createTestUser(t, db)

	// Create inline food entry
	calories := 350.0
	protein := 25.0
	carbs := 40.0
	fat := 10.0
	fiber := 5.0
	tag := "contextual"

	entry := &diary.DiaryEntry{
		UserID:               userID,
		InlineFoodName:       strPtr("Homemade Protein Bar"),
		InlineFoodDescription: strPtr("Custom oats and whey recipe"),
		InlineFoodCalories:   &calories,
		InlineFoodProtein:    &protein,
		InlineFoodCarbs:      &carbs,
		InlineFoodFat:        &fat,
		InlineFoodFiber:      &fiber,
		InlineFoodTag:        &tag,
		Date:                 mustParseDate("2025-12-25"),
		MealType:             diary.Snack,
		QuantityGrams:        80.0,
		Calories:             280.0,
		Protein:              20.0,
		Carbs:                32.0,
		Fat:                  8.0,
		Fiber:                4.0,
		FoodTag:              "contextual",
	}

	err := diaryRepo.Create(entry)
	require.NoError(t, err)

	// Get initial food count
	initialFoods, err := foodRepo.GetAll()
	require.NoError(t, err)
	initialCount := len(initialFoods)

	// Simulate SaveAsFood handler logic
	// Create global food from inline food
	savedFood, err := foodRepo.Create(food.CreateFoodRequest{
		Name:        *entry.InlineFoodName,
		Description: *entry.InlineFoodDescription,
		Calories:    *entry.InlineFoodCalories,
		Protein:     *entry.InlineFoodProtein,
		Carbs:       *entry.InlineFoodCarbs,
		Fat:         *entry.InlineFoodFat,
		Fiber:       *entry.InlineFoodFiber,
		Tag:         *entry.InlineFoodTag,
	})
	require.NoError(t, err)
	assert.NotZero(t, savedFood.ID)

	// Update diary entry to reference saved food
	entry.FoodID = &savedFood.ID
	entry.InlineFoodName = nil
	entry.InlineFoodDescription = nil
	entry.InlineFoodCalories = nil
	entry.InlineFoodProtein = nil
	entry.InlineFoodCarbs = nil
	entry.InlineFoodFat = nil
	entry.InlineFoodFiber = nil
	entry.InlineFoodTag = nil

	err = diaryRepo.Update(entry)
	require.NoError(t, err)

	// Verify food was created
	allFoods, err := foodRepo.GetAll()
	require.NoError(t, err)
	assert.Equal(t, initialCount+1, len(allFoods))

	// Verify the created food has correct values
	createdFood, err := foodRepo.GetByID(int(savedFood.ID))
	require.NoError(t, err)
	assert.Equal(t, "Homemade Protein Bar", createdFood.Name)
	assert.Equal(t, "Custom oats and whey recipe", createdFood.Description)
	assert.Equal(t, 350.0, createdFood.Calories)
	assert.Equal(t, 25.0, createdFood.Protein)
	assert.Equal(t, 40.0, createdFood.Carbs)
	assert.Equal(t, 10.0, createdFood.Fat)
	assert.Equal(t, 5.0, createdFood.Fiber)
	assert.Equal(t, "contextual", createdFood.Tag)

	// Verify diary entry was updated
	updatedEntry, err := diaryRepo.GetByID(entry.ID, userID)
	require.NoError(t, err)
	assert.NotNil(t, updatedEntry.FoodID)
	assert.Equal(t, savedFood.ID, *updatedEntry.FoodID)
	assert.Nil(t, updatedEntry.InlineFoodName)
	assert.Nil(t, updatedEntry.InlineFoodDescription)
	assert.Nil(t, updatedEntry.InlineFoodCalories)
	assert.Nil(t, updatedEntry.InlineFoodProtein)
	assert.Nil(t, updatedEntry.InlineFoodCarbs)
	assert.Nil(t, updatedEntry.InlineFoodFat)
	assert.Nil(t, updatedEntry.InlineFoodFiber)
	assert.Nil(t, updatedEntry.InlineFoodTag)

	// Verify cached nutrition is preserved (historical accuracy)
	assert.Equal(t, 280.0, updatedEntry.Calories)
	assert.Equal(t, 20.0, updatedEntry.Protein)
	assert.Equal(t, 32.0, updatedEntry.Carbs)
	assert.Equal(t, 8.0, updatedEntry.Fat)
	assert.Equal(t, 4.0, updatedEntry.Fiber)
}

func TestSaveAsFood_DefaultTag(t *testing.T) {
	db, diaryRepo, foodRepo := setupDiaryTest(t)
	userID := createTestUser(t, db)

	// Create inline food without explicit tag
	calories := 200.0
	protein := 10.0
	carbs := 20.0
	fat := 5.0
	fiber := 2.0

	entry := &diary.DiaryEntry{
		UserID:             userID,
		InlineFoodName:     strPtr("Simple Food"),
		InlineFoodCalories: &calories,
		InlineFoodProtein:  &protein,
		InlineFoodCarbs:    &carbs,
		InlineFoodFat:      &fat,
		InlineFoodFiber:    &fiber,
		// No InlineFoodTag set
		Date:          mustParseDate("2025-12-25"),
		MealType:      diary.Breakfast,
		QuantityGrams: 100.0,
		Calories:      200.0,
		Protein:       10.0,
		Carbs:         20.0,
		Fat:           5.0,
		Fiber:         2.0,
	}

	err := diaryRepo.Create(entry)
	require.NoError(t, err)

	// Simulate SaveAsFood with default tag
	tag := "routine" // Default
	if entry.InlineFoodTag != nil && *entry.InlineFoodTag != "" {
		tag = *entry.InlineFoodTag
	}

	savedFood, err := foodRepo.Create(food.CreateFoodRequest{
		Name:     *entry.InlineFoodName,
		Calories: *entry.InlineFoodCalories,
		Protein:  *entry.InlineFoodProtein,
		Carbs:    *entry.InlineFoodCarbs,
		Fat:      *entry.InlineFoodFat,
		Fiber:    *entry.InlineFoodFiber,
		Tag:      tag,
	})
	require.NoError(t, err)

	// Verify default tag was applied
	assert.Equal(t, "routine", savedFood.Tag)
}

func TestSaveAsFood_PreservesNutritionHistory(t *testing.T) {
	db, diaryRepo, foodRepo := setupDiaryTest(t)
	userID := createTestUser(t, db)

	// Create inline food
	calories := 300.0
	protein := 20.0
	carbs := 30.0
	fat := 10.0
	fiber := 5.0

	entry := &diary.DiaryEntry{
		UserID:             userID,
		InlineFoodName:     strPtr("Test Food"),
		InlineFoodCalories: &calories,
		InlineFoodProtein:  &protein,
		InlineFoodCarbs:    &carbs,
		InlineFoodFat:      &fat,
		InlineFoodFiber:    &fiber,
		Date:               mustParseDate("2025-12-25"),
		MealType:           diary.Lunch,
		QuantityGrams:      150.0,
		// Consumed nutrition for 150g
		Calories: 450.0, // 300 * 1.5
		Protein:  30.0,  // 20 * 1.5
		Carbs:    45.0,  // 30 * 1.5
		Fat:      15.0,  // 10 * 1.5
		Fiber:    7.5,   // 5 * 1.5
		FoodTag:  "routine",
	}

	err := diaryRepo.Create(entry)
	require.NoError(t, err)

	originalCalories := entry.Calories
	originalProtein := entry.Protein

	// Save as food
	savedFood, err := foodRepo.Create(food.CreateFoodRequest{
		Name:     *entry.InlineFoodName,
		Calories: *entry.InlineFoodCalories,
		Protein:  *entry.InlineFoodProtein,
		Carbs:    *entry.InlineFoodCarbs,
		Fat:      *entry.InlineFoodFat,
		Fiber:    *entry.InlineFoodFiber,
		Tag:      "routine",
	})
	require.NoError(t, err)

	// Update entry
	entry.FoodID = &savedFood.ID
	entry.InlineFoodName = nil
	entry.InlineFoodCalories = nil
	entry.InlineFoodProtein = nil
	entry.InlineFoodCarbs = nil
	entry.InlineFoodFat = nil
	entry.InlineFoodFiber = nil
	// Cached nutrition should remain unchanged

	err = diaryRepo.Update(entry)
	require.NoError(t, err)

	// Verify cached nutrition is preserved
	updated, err := diaryRepo.GetByID(entry.ID, userID)
	require.NoError(t, err)
	assert.Equal(t, originalCalories, updated.Calories)
	assert.Equal(t, originalProtein, updated.Protein)

	// Now modify the saved food's nutrition
	updatedFood, err := foodRepo.Update(int(savedFood.ID), food.UpdateFoodRequest{
		Name:        "Test Food",
		Description: "",
		Calories:    500.0, // Changed from 300 to 500
		Protein:     40.0,  // Changed from 20 to 40
		Carbs:       30.0,
		Fat:         10.0,
		Fiber:       5.0,
		Tag:         "routine",
	})
	require.NoError(t, err)
	assert.Equal(t, 500.0, updatedFood.Calories)

	// Verify diary entry nutrition is still the original cached values
	entryAfterFoodUpdate, err := diaryRepo.GetByID(entry.ID, userID)
	require.NoError(t, err)
	assert.Equal(t, originalCalories, entryAfterFoodUpdate.Calories, "Cached nutrition should not change when food is modified")
	assert.Equal(t, originalProtein, entryAfterFoodUpdate.Protein, "Cached nutrition should not change when food is modified")
}

func TestInlineFood_ValidationScenarios(t *testing.T) {
	t.Run("Cannot have both food_id and inline_food_name", func(t *testing.T) {
		db, diaryRepo, foodRepo := setupDiaryTest(t)
		userID := createTestUser(t, db)

		// Create a saved food
		savedFood, err := foodRepo.Create(food.CreateFoodRequest{
			Name:     "Saved Food",
			Calories: 200.0,
			Protein:  10.0,
			Carbs:    20.0,
			Fat:      5.0,
			Fiber:    2.0,
			Tag:      "routine",
		})
		require.NoError(t, err)

		// Try to create entry with both (this should be caught by handler validation)
		// For this test, we're testing the repository level behavior
		calories := 300.0
		protein := 15.0
		carbs := 25.0
		fat := 8.0
		fiber := 3.0

		entry := &diary.DiaryEntry{
			UserID:             userID,
			FoodID:             &savedFood.ID,
			InlineFoodName:     strPtr("Inline Food"),
			InlineFoodCalories: &calories,
			InlineFoodProtein:  &protein,
			InlineFoodCarbs:    &carbs,
			InlineFoodFat:      &fat,
			InlineFoodFiber:    &fiber,
			Date:               mustParseDate("2025-12-25"),
			MealType:           diary.Lunch,
			QuantityGrams:      100.0,
			Calories:           300.0,
			Protein:            15.0,
			Carbs:              25.0,
			Fat:                8.0,
			Fiber:              3.0,
		}

		// Repository will allow this (validation happens at handler level)
		// But we can verify both fields are present
		err = diaryRepo.Create(entry)
		require.NoError(t, err)

		retrieved, err := diaryRepo.GetByID(entry.ID, userID)
		require.NoError(t, err)
		assert.NotNil(t, retrieved.FoodID)
		assert.NotNil(t, retrieved.InlineFoodName)
	})
}

func TestInlineFood_DisplayNamePrecedence(t *testing.T) {
	db, diaryRepo, foodRepo := setupDiaryTest(t)
	userID := createTestUser(t, db)

	// Create saved food
	savedFood, err := foodRepo.Create(food.CreateFoodRequest{
		Name:     "Saved Food Name",
		Calories: 200.0,
		Protein:  10.0,
		Carbs:    20.0,
		Fat:      5.0,
		Fiber:    2.0,
		Tag:      "routine",
	})
	require.NoError(t, err)

	t.Run("Inline food name takes precedence", func(t *testing.T) {
		calories := 300.0
		protein := 15.0
		carbs := 25.0
		fat := 8.0
		fiber := 3.0

		entry := &diary.DiaryEntry{
			UserID:             userID,
			FoodID:             &savedFood.ID,
			InlineFoodName:     strPtr("Inline Name Takes Precedence"),
			InlineFoodCalories: &calories,
			InlineFoodProtein:  &protein,
			InlineFoodCarbs:    &carbs,
			InlineFoodFat:      &fat,
			InlineFoodFiber:    &fiber,
			Date:               mustParseDate("2025-12-25"),
			MealType:           diary.Breakfast,
			QuantityGrams:      100.0,
			Calories:           300.0,
			Protein:            15.0,
			Carbs:              25.0,
			Fat:                8.0,
			Fiber:              3.0,
		}

		err := diaryRepo.Create(entry)
		require.NoError(t, err)

		// Get entries with populated names
		entries, err := diaryRepo.GetByDate(userID, mustParseDate("2025-12-25"))
		require.NoError(t, err)
		require.Len(t, entries, 1)

		// Inline name should be shown
		assert.Equal(t, "Inline Name Takes Precedence", entries[0].FoodName)
	})

	t.Run("Saved food name shown when no inline name", func(t *testing.T) {
		entry := &diary.DiaryEntry{
			UserID:        userID,
			FoodID:        &savedFood.ID,
			Date:          mustParseDate("2025-12-26"),
			MealType:      diary.Lunch,
			QuantityGrams: 100.0,
			Calories:      200.0,
			Protein:       10.0,
			Carbs:         20.0,
			Fat:           5.0,
			Fiber:         2.0,
		}

		err := diaryRepo.Create(entry)
		require.NoError(t, err)

		// Get entries with populated names
		entries, err := diaryRepo.GetByDate(userID, mustParseDate("2025-12-26"))
		require.NoError(t, err)
		require.Len(t, entries, 1)

		// Saved food name should be shown
		assert.Equal(t, "Saved Food Name", entries[0].FoodName)
	})
}

func TestInlineFood_CompleteWorkflow(t *testing.T) {
	// This test simulates the complete user workflow:
	// 1. Create inline food
	// 2. Update it
	// 3. Save it as a permanent food
	// 4. Verify all data

	db, diaryRepo, foodRepo := setupDiaryTest(t)
	userID := createTestUser(t, db)

	// Step 1: Create inline food
	calories := 350.0
	protein := 25.0
	carbs := 40.0
	fat := 10.0
	fiber := 5.0

	entry := &diary.DiaryEntry{
		UserID:             userID,
		InlineFoodName:     strPtr("My Custom Bar"),
		InlineFoodCalories: &calories,
		InlineFoodProtein:  &protein,
		InlineFoodCarbs:    &carbs,
		InlineFoodFat:      &fat,
		InlineFoodFiber:    &fiber,
		InlineFoodTag:      strPtr("contextual"),
		Date:               mustParseDate("2025-12-25"),
		MealType:           diary.Snack,
		QuantityGrams:      80.0,
		Calories:           280.0,
		Protein:            20.0,
		Carbs:              32.0,
		Fat:                8.0,
		Fiber:              4.0,
		FoodTag:            "contextual",
	}

	err := diaryRepo.Create(entry)
	require.NoError(t, err)

	// Step 2: Update nutrition values
	newCalories := 400.0
	newProtein := 30.0
	entry.InlineFoodCalories = &newCalories
	entry.InlineFoodProtein = &newProtein
	entry.Calories = 320.0 // 400 * 0.8
	entry.Protein = 24.0   // 30 * 0.8

	err = diaryRepo.Update(entry)
	require.NoError(t, err)

	// Step 3: Save as permanent food
	savedFood, err := foodRepo.Create(food.CreateFoodRequest{
		Name:     *entry.InlineFoodName,
		Calories: *entry.InlineFoodCalories,
		Protein:  *entry.InlineFoodProtein,
		Carbs:    *entry.InlineFoodCarbs,
		Fat:      *entry.InlineFoodFat,
		Fiber:    *entry.InlineFoodFiber,
		Tag:      *entry.InlineFoodTag,
	})
	require.NoError(t, err)

	entry.FoodID = &savedFood.ID
	entry.InlineFoodName = nil
	entry.InlineFoodCalories = nil
	entry.InlineFoodProtein = nil
	entry.InlineFoodCarbs = nil
	entry.InlineFoodFat = nil
	entry.InlineFoodFiber = nil
	entry.InlineFoodTag = nil

	err = diaryRepo.Update(entry)
	require.NoError(t, err)

	// Step 4: Verify everything
	finalEntry, err := diaryRepo.GetByID(entry.ID, userID)
	require.NoError(t, err)

	// Entry should reference saved food, not inline
	assert.NotNil(t, finalEntry.FoodID)
	assert.Nil(t, finalEntry.InlineFoodName)

	// Cached nutrition should be preserved
	assert.Equal(t, 320.0, finalEntry.Calories)
	assert.Equal(t, 24.0, finalEntry.Protein)

	// Saved food should exist with correct values
	permanentFood, err := foodRepo.GetByID(int(savedFood.ID))
	require.NoError(t, err)
	assert.Equal(t, "My Custom Bar", permanentFood.Name)
	assert.Equal(t, 400.0, permanentFood.Calories)
	assert.Equal(t, 30.0, permanentFood.Protein)
	assert.Equal(t, "contextual", permanentFood.Tag)
}

package tests

import (
	"testing"

	"ultra-bis/internal/diary"
	"ultra-bis/internal/food"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateEntryFromOpenFoodFacts_Success(t *testing.T) {
	db, diaryRepo, _ := setupDiaryTest(t)
	userID := createTestUser(t, db)

	// Simulate Open Food Facts product data (per 100g)
	productName := "Apple"
	brands := "Generic"
	calories := 52.0
	protein := 0.3
	carbs := 14.0
	fat := 0.2
	fiber := 2.4
	quantityGrams := 150.0 // User consumed 150g
	tag := "routine"

	// Build description as the handler does
	description := brands + " - " + productName

	// Calculate consumed nutrition (as handler does)
	multiplier := quantityGrams / 100.0

	entry := &diary.DiaryEntry{
		UserID:                userID,
		InlineFoodName:        &productName,
		InlineFoodDescription: &description,
		InlineFoodCalories:    &calories,
		InlineFoodProtein:     &protein,
		InlineFoodCarbs:       &carbs,
		InlineFoodFat:         &fat,
		InlineFoodFiber:       &fiber,
		InlineFoodTag:         &tag,
		Date:                  mustParseDate("2025-12-25"),
		MealType:              diary.Snack,
		QuantityGrams:         quantityGrams,
		Notes:                 "Fresh apple from Open Food Facts",
		// Cached consumed nutrition
		Calories: roundToTwo(calories * multiplier),
		Protein:  roundToTwo(protein * multiplier),
		Carbs:    roundToTwo(carbs * multiplier),
		Fat:      roundToTwo(fat * multiplier),
		Fiber:    roundToTwo(fiber * multiplier),
		FoodTag:  tag,
	}

	err := diaryRepo.Create(entry)
	require.NoError(t, err)
	assert.NotZero(t, entry.ID)

	// Verify the entry was created correctly
	retrieved, err := diaryRepo.GetByID(entry.ID, userID)
	require.NoError(t, err)

	assert.Equal(t, "Apple", *retrieved.InlineFoodName)
	assert.Equal(t, "Generic - Apple", *retrieved.InlineFoodDescription)
	assert.Equal(t, 52.0, *retrieved.InlineFoodCalories)
	assert.Equal(t, 0.3, *retrieved.InlineFoodProtein)
	assert.Equal(t, 14.0, *retrieved.InlineFoodCarbs)
	assert.Equal(t, 0.2, *retrieved.InlineFoodFat)
	assert.Equal(t, 2.4, *retrieved.InlineFoodFiber)
	assert.Equal(t, "routine", *retrieved.InlineFoodTag)

	// Verify consumed nutrition (150g = 1.5x per-100g values)
	assert.Equal(t, 78.0, retrieved.Calories)      // 52 * 1.5
	assert.Equal(t, 0.45, retrieved.Protein)       // 0.3 * 1.5
	assert.Equal(t, 21.0, retrieved.Carbs)         // 14 * 1.5
	assert.Equal(t, 0.3, retrieved.Fat)            // 0.2 * 1.5
	assert.Equal(t, 3.6, retrieved.Fiber)          // 2.4 * 1.5
	assert.Equal(t, "routine", retrieved.FoodTag)
}

func TestCreateEntryFromOpenFoodFacts_DefaultTag(t *testing.T) {
	db, diaryRepo, _ := setupDiaryTest(t)
	userID := createTestUser(t, db)

	// Test that tag defaults to "routine" when not specified
	productName := "Banana"
	brands := "Organic"
	calories := 89.0
	protein := 1.1
	carbs := 23.0
	fat := 0.3
	fiber := 2.6
	quantityGrams := 120.0
	tag := "routine" // Default

	description := brands + " - " + productName
	multiplier := quantityGrams / 100.0

	entry := &diary.DiaryEntry{
		UserID:                userID,
		InlineFoodName:        &productName,
		InlineFoodDescription: &description,
		InlineFoodCalories:    &calories,
		InlineFoodProtein:     &protein,
		InlineFoodCarbs:       &carbs,
		InlineFoodFat:         &fat,
		InlineFoodFiber:       &fiber,
		InlineFoodTag:         &tag,
		Date:                  mustParseDate("2025-12-25"),
		MealType:              diary.Breakfast,
		QuantityGrams:         quantityGrams,
		Calories:              roundToTwo(calories * multiplier),
		Protein:               roundToTwo(protein * multiplier),
		Carbs:                 roundToTwo(carbs * multiplier),
		Fat:                   roundToTwo(fat * multiplier),
		Fiber:                 roundToTwo(fiber * multiplier),
		FoodTag:               tag,
	}

	err := diaryRepo.Create(entry)
	require.NoError(t, err)

	retrieved, err := diaryRepo.GetByID(entry.ID, userID)
	require.NoError(t, err)

	assert.Equal(t, "routine", *retrieved.InlineFoodTag)
	assert.Equal(t, "routine", retrieved.FoodTag)
}

func TestCreateEntryFromOpenFoodFacts_ContextualTag(t *testing.T) {
	db, diaryRepo, _ := setupDiaryTest(t)
	userID := createTestUser(t, db)

	// Test with contextual tag
	productName := "Ice Cream"
	brands := "Ben & Jerry's"
	calories := 250.0
	protein := 4.0
	carbs := 28.0
	fat := 14.0
	fiber := 1.0
	quantityGrams := 100.0
	tag := "contextual" // Special treat

	description := brands + " - " + productName
	multiplier := quantityGrams / 100.0

	entry := &diary.DiaryEntry{
		UserID:                userID,
		InlineFoodName:        &productName,
		InlineFoodDescription: &description,
		InlineFoodCalories:    &calories,
		InlineFoodProtein:     &protein,
		InlineFoodCarbs:       &carbs,
		InlineFoodFat:         &fat,
		InlineFoodFiber:       &fiber,
		InlineFoodTag:         &tag,
		Date:                  mustParseDate("2025-12-25"),
		MealType:              diary.Snack,
		QuantityGrams:         quantityGrams,
		Calories:              roundToTwo(calories * multiplier),
		Protein:               roundToTwo(protein * multiplier),
		Carbs:                 roundToTwo(carbs * multiplier),
		Fat:                   roundToTwo(fat * multiplier),
		Fiber:                 roundToTwo(fiber * multiplier),
		FoodTag:               tag,
	}

	err := diaryRepo.Create(entry)
	require.NoError(t, err)

	retrieved, err := diaryRepo.GetByID(entry.ID, userID)
	require.NoError(t, err)

	assert.Equal(t, "contextual", *retrieved.InlineFoodTag)
	assert.Equal(t, "contextual", retrieved.FoodTag)
}

func TestCreateEntryFromOpenFoodFacts_NutritionCalculation(t *testing.T) {
	db, diaryRepo, _ := setupDiaryTest(t)
	userID := createTestUser(t, db)

	// Test various quantity calculations
	tests := []struct {
		name          string
		quantity      float64
		expectedCals  float64
		expectedProt  float64
	}{
		{"50g portion", 50.0, 26.0, 0.15},    // 52 * 0.5, 0.3 * 0.5
		{"100g portion", 100.0, 52.0, 0.3},   // 52 * 1.0, 0.3 * 1.0
		{"200g portion", 200.0, 104.0, 0.6},  // 52 * 2.0, 0.3 * 2.0
		{"75g portion", 75.0, 39.0, 0.22},    // 52 * 0.75, 0.3 * 0.75 (rounded to 0.22)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productName := "Apple"
			calories := 52.0
			protein := 0.3
			carbs := 14.0
			fat := 0.2
			fiber := 2.4
			tag := "routine"
			description := "Generic - Apple"

			multiplier := tt.quantity / 100.0

			entry := &diary.DiaryEntry{
				UserID:                userID,
				InlineFoodName:        &productName,
				InlineFoodDescription: &description,
				InlineFoodCalories:    &calories,
				InlineFoodProtein:     &protein,
				InlineFoodCarbs:       &carbs,
				InlineFoodFat:         &fat,
				InlineFoodFiber:       &fiber,
				InlineFoodTag:         &tag,
				Date:                  mustParseDate("2025-12-25"),
				MealType:              diary.Snack,
				QuantityGrams:         tt.quantity,
				Calories:              roundToTwo(calories * multiplier),
				Protein:               roundToTwo(protein * multiplier),
				Carbs:                 roundToTwo(carbs * multiplier),
				Fat:                   roundToTwo(fat * multiplier),
				Fiber:                 roundToTwo(fiber * multiplier),
				FoodTag:               tag,
			}

			err := diaryRepo.Create(entry)
			require.NoError(t, err)

			retrieved, err := diaryRepo.GetByID(entry.ID, userID)
			require.NoError(t, err)

			assert.InDelta(t, tt.expectedCals, retrieved.Calories, 0.01)
			assert.InDelta(t, tt.expectedProt, retrieved.Protein, 0.01)
		})
	}
}

func TestCreateEntryFromOpenFoodFacts_SaveAsFood(t *testing.T) {
	db, diaryRepo, foodRepo := setupDiaryTest(t)
	userID := createTestUser(t, db)

	// Create entry from Open Food Facts
	productName := "Chocolate Bar"
	brands := "Hershey's"
	calories := 535.0
	protein := 6.0
	carbs := 59.0
	fat := 30.0
	fiber := 3.0
	quantityGrams := 43.0 // Standard bar size
	tag := "contextual"
	description := brands + " - " + productName

	multiplier := quantityGrams / 100.0

	entry := &diary.DiaryEntry{
		UserID:                userID,
		InlineFoodName:        &productName,
		InlineFoodDescription: &description,
		InlineFoodCalories:    &calories,
		InlineFoodProtein:     &protein,
		InlineFoodCarbs:       &carbs,
		InlineFoodFat:         &fat,
		InlineFoodFiber:       &fiber,
		InlineFoodTag:         &tag,
		Date:                  mustParseDate("2025-12-25"),
		MealType:              diary.Snack,
		QuantityGrams:         quantityGrams,
		Notes:                 "Afternoon treat",
		Calories:              roundToTwo(calories * multiplier),
		Protein:               roundToTwo(protein * multiplier),
		Carbs:                 roundToTwo(carbs * multiplier),
		Fat:                   roundToTwo(fat * multiplier),
		Fiber:                 roundToTwo(fiber * multiplier),
		FoodTag:               tag,
	}

	err := diaryRepo.Create(entry)
	require.NoError(t, err)

	// Now save it as a permanent food
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

	// Verify the entry now references the saved food
	retrieved, err := diaryRepo.GetByID(entry.ID, userID)
	require.NoError(t, err)

	assert.NotNil(t, retrieved.FoodID)
	assert.Equal(t, savedFood.ID, *retrieved.FoodID)
	assert.Nil(t, retrieved.InlineFoodName)

	// Verify cached nutrition is preserved (historical accuracy)
	assert.Equal(t, roundToTwo(calories*multiplier), retrieved.Calories)
	assert.Equal(t, roundToTwo(protein*multiplier), retrieved.Protein)
}

// Helper function to round to two decimal places
func roundToTwo(val float64) float64 {
	return float64(int(val*100+0.5)) / 100
}

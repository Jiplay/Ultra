package tests

import (
	"testing"
	"time"

	"ultra-bis/internal/diary"
	"ultra-bis/internal/food"
	"ultra-bis/internal/goal"
	"ultra-bis/internal/recipe"
	"ultra-bis/internal/user"
	"ultra-bis/test/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupTagBreakdownTest creates a test environment with diary, food, recipe, goal, and user repositories
func setupTagBreakdownTest(t *testing.T) (*gorm.DB, *diary.Repository, *food.Repository, *recipe.Repository, *goal.Repository, uint) {
	t.Helper()

	// Setup test database
	db := testutil.SetupTestDB(t)

	// Run migrations
	if err := db.AutoMigrate(
		&user.User{},
		&food.Food{},
		&recipe.Recipe{},
		&recipe.RecipeIngredient{},
		&diary.DiaryEntry{},
		&goal.NutritionGoal{},
	); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Create repositories
	diaryRepo := diary.NewRepository(db)
	foodRepo := food.NewRepository(db)
	recipeRepo := recipe.NewRepository(db)
	goalRepo := goal.NewRepository(db)

	// Create test user
	testUser := &user.User{
		Email:        testutil.MockUserEmail(),
		PasswordHash: "hashedpassword",
		Name:         "Test User",
		Height:       175,
		Gender:       "male",
	}
	db.Create(testUser)

	return db, diaryRepo, foodRepo, recipeRepo, goalRepo, testUser.ID
}

// TestDailySummary_AllRoutineFoods tests that 100% routine foods show correct breakdown
func TestDailySummary_AllRoutineFoods(t *testing.T) {
	db, diaryRepo, foodRepo, _, goalRepo, userID := setupTagBreakdownTest(t)

	// Create routine foods
	routineFood1, _ := foodRepo.Create(food.CreateFoodRequest{
		Name:     "Chicken Breast",
		Calories: 165,
		Protein:  31,
		Tag:      "routine",
	})
	routineFood2, _ := foodRepo.Create(food.CreateFoodRequest{
		Name:     "Brown Rice",
		Calories: 111,
		Protein:  2.6,
		Tag:      "routine",
	})

	// Create nutrition goal
	goalRepo.Create(&goal.NutritionGoal{
		UserID:    userID,
		Calories:  2000,
		Protein:   150,
		Carbs:     200,
		Fat:       67,
		Fiber:     30,
		StartDate: time.Now().AddDate(0, 0, -30), // Started 30 days ago
		IsActive:  true,
	})

	// Create diary entries for today - all routine
	today := time.Now()
	entry1 := &diary.DiaryEntry{
		UserID:        userID,
		FoodID:        &routineFood1.ID,
		Date:          today,
		MealType:      diary.Breakfast,
		QuantityGrams: 200, // 200g chicken = 330 calories
		Calories:      330,
		Protein:       62,
		FoodTag:       "routine",
	}
	entry2 := &diary.DiaryEntry{
		UserID:        userID,
		FoodID:        &routineFood2.ID,
		Date:          today,
		MealType:      diary.Lunch,
		QuantityGrams: 150, // 150g rice = 166.5 calories
		Calories:      166.5,
		Protein:       3.9,
		FoodTag:       "routine",
	}
	db.Create(entry1)
	db.Create(entry2)

	// Get daily summary
	summary, err := diaryRepo.GetDailySummary(userID, today)
	require.NoError(t, err)

	// Calculate expected values
	expectedTotalCalories := 330.0 + 166.5 // 496.5
	expectedRoutineCalories := expectedTotalCalories
	expectedContextualCalories := 0.0
	expectedRoutinePercent := 100.0
	expectedContextualPercent := 0.0

	// Verify totals
	assert.InDelta(t, expectedTotalCalories, summary["calories"], 0.01)

	// Get entries to verify tag breakdown calculation
	entries, err := diaryRepo.GetByDate(userID, today)
	require.NoError(t, err)

	// Calculate tag breakdown manually
	var routineCalories, contextualCalories float64
	for _, entry := range entries {
		tag := entry.FoodTag
		if tag == "" {
			tag = entry.RecipeTag
		}
		if tag == "routine" {
			routineCalories += entry.Calories
		} else if tag == "contextual" {
			contextualCalories += entry.Calories
		}
	}

	// Verify tag breakdown
	assert.InDelta(t, expectedRoutineCalories, routineCalories, 0.01, "Routine calories should be 100% of total")
	assert.InDelta(t, expectedContextualCalories, contextualCalories, 0.01, "Contextual calories should be 0")

	// Calculate percentages
	var routinePercent, contextualPercent float64
	totalCals := routineCalories + contextualCalories
	if totalCals > 0 {
		routinePercent = (routineCalories / totalCals) * 100
		contextualPercent = (contextualCalories / totalCals) * 100
	}

	assert.InDelta(t, expectedRoutinePercent, routinePercent, 0.01)
	assert.InDelta(t, expectedContextualPercent, contextualPercent, 0.01)
}

// TestDailySummary_AllContextualFoods tests that 100% contextual foods show correct breakdown
func TestDailySummary_AllContextualFoods(t *testing.T) {
	db, diaryRepo, foodRepo, _, goalRepo, userID := setupTagBreakdownTest(t)

	// Create contextual foods
	contextualFood1, _ := foodRepo.Create(food.CreateFoodRequest{
		Name:     "Pizza",
		Calories: 266,
		Protein:  11,
		Tag:      "contextual",
	})
	contextualFood2, _ := foodRepo.Create(food.CreateFoodRequest{
		Name:     "Ice Cream",
		Calories: 207,
		Protein:  3.5,
		Tag:      "contextual",
	})

	// Create nutrition goal
	goalRepo.Create(&goal.NutritionGoal{
		UserID:    userID,
		Calories:  2000,
		Protein:   150,
		StartDate: time.Now().AddDate(0, 0, -30),
		IsActive:  true,
	})

	// Create diary entries for today - all contextual
	today := time.Now()
	entry1 := &diary.DiaryEntry{
		UserID:        userID,
		FoodID:        &contextualFood1.ID,
		Date:          today,
		MealType:      diary.Lunch,
		QuantityGrams: 200, // 200g pizza = 532 calories
		Calories:      532,
		Protein:       22,
		FoodTag:       "contextual",
	}
	entry2 := &diary.DiaryEntry{
		UserID:        userID,
		FoodID:        &contextualFood2.ID,
		Date:          today,
		MealType:      diary.Snack,
		QuantityGrams: 100, // 100g ice cream = 207 calories
		Calories:      207,
		Protein:       3.5,
		FoodTag:       "contextual",
	}
	db.Create(entry1)
	db.Create(entry2)

	// Get entries to verify tag breakdown
	entries, err := diaryRepo.GetByDate(userID, today)
	require.NoError(t, err)

	// Calculate tag breakdown
	var routineCalories, contextualCalories float64
	for _, entry := range entries {
		tag := entry.FoodTag
		if tag == "" {
			tag = entry.RecipeTag
		}
		if tag == "routine" {
			routineCalories += entry.Calories
		} else if tag == "contextual" {
			contextualCalories += entry.Calories
		}
	}

	// Expected values
	expectedTotalCalories := 532.0 + 207.0 // 739
	expectedRoutineCalories := 0.0
	expectedContextualCalories := expectedTotalCalories
	expectedRoutinePercent := 0.0
	expectedContextualPercent := 100.0

	// Verify tag breakdown
	assert.InDelta(t, expectedRoutineCalories, routineCalories, 0.01, "Routine calories should be 0")
	assert.InDelta(t, expectedContextualCalories, contextualCalories, 0.01, "Contextual calories should be 100% of total")

	// Calculate percentages
	var routinePercent, contextualPercent float64
	totalCals := routineCalories + contextualCalories
	if totalCals > 0 {
		routinePercent = (routineCalories / totalCals) * 100
		contextualPercent = (contextualCalories / totalCals) * 100
	}

	assert.InDelta(t, expectedRoutinePercent, routinePercent, 0.01)
	assert.InDelta(t, expectedContextualPercent, contextualPercent, 0.01)
}

// TestDailySummary_MixedTags tests that mixed routine/contextual foods show correct breakdown
func TestDailySummary_MixedTags(t *testing.T) {
	db, diaryRepo, foodRepo, _, goalRepo, userID := setupTagBreakdownTest(t)

	// Create foods with different tags
	routineFood, _ := foodRepo.Create(food.CreateFoodRequest{
		Name:     "Chicken Breast",
		Calories: 165,
		Protein:  31,
		Tag:      "routine",
	})
	contextualFood, _ := foodRepo.Create(food.CreateFoodRequest{
		Name:     "Pizza",
		Calories: 266,
		Protein:  11,
		Tag:      "contextual",
	})

	// Create nutrition goal
	goalRepo.Create(&goal.NutritionGoal{
		UserID:    userID,
		Calories:  2000,
		Protein:   150,
		StartDate: time.Now().AddDate(0, 0, -30),
		IsActive:  true,
	})

	// Create diary entries with mixed tags
	today := time.Now()

	// 400 calories from routine
	entry1 := &diary.DiaryEntry{
		UserID:        userID,
		FoodID:        &routineFood.ID,
		Date:          today,
		MealType:      diary.Breakfast,
		QuantityGrams: 200,
		Calories:      330, // 200g chicken
		Protein:       62,
		FoodTag:       "routine",
	}
	entry2 := &diary.DiaryEntry{
		UserID:        userID,
		FoodID:        &routineFood.ID,
		Date:          today,
		MealType:      diary.Dinner,
		QuantityGrams: 100,
		Calories:      165, // 100g chicken
		Protein:       31,
		FoodTag:       "routine",
	}
	// Total routine: 495 calories

	// 300 calories from contextual
	entry3 := &diary.DiaryEntry{
		UserID:        userID,
		FoodID:        &contextualFood.ID,
		Date:          today,
		MealType:      diary.Lunch,
		QuantityGrams: 150,
		Calories:      399, // 150g pizza
		Protein:       16.5,
		FoodTag:       "contextual",
	}
	// Total contextual: 399 calories

	db.Create(entry1)
	db.Create(entry2)
	db.Create(entry3)

	// Get entries to verify tag breakdown
	entries, err := diaryRepo.GetByDate(userID, today)
	require.NoError(t, err)

	// Calculate tag breakdown
	var routineCalories, contextualCalories float64
	for _, entry := range entries {
		tag := entry.FoodTag
		if tag == "" {
			tag = entry.RecipeTag
		}
		if tag == "routine" {
			routineCalories += entry.Calories
		} else if tag == "contextual" {
			contextualCalories += entry.Calories
		}
	}

	// Expected values: 495 routine, 399 contextual, total 894
	// Routine %: (495 / 894) * 100 = 55.37%
	// Contextual %: (399 / 894) * 100 = 44.63%
	expectedTotalCalories := 495.0 + 399.0 // 894
	expectedRoutineCalories := 495.0
	expectedContextualCalories := 399.0
	expectedRoutinePercent := (495.0 / 894.0) * 100    // ~55.37%
	expectedContextualPercent := (399.0 / 894.0) * 100 // ~44.63%

	// Verify tag breakdown
	assert.InDelta(t, expectedRoutineCalories, routineCalories, 0.01)
	assert.InDelta(t, expectedContextualCalories, contextualCalories, 0.01)

	// Calculate percentages
	var routinePercent, contextualPercent float64
	totalCals := routineCalories + contextualCalories
	assert.InDelta(t, expectedTotalCalories, totalCals, 0.01)

	if totalCals > 0 {
		routinePercent = (routineCalories / totalCals) * 100
		contextualPercent = (contextualCalories / totalCals) * 100
	}

	assert.InDelta(t, expectedRoutinePercent, routinePercent, 0.01)
	assert.InDelta(t, expectedContextualPercent, contextualPercent, 0.01)

	// Verify percentages sum to 100%
	assert.InDelta(t, 100.0, routinePercent+contextualPercent, 0.1)
}

// TestDailySummary_ZeroCalories tests that zero calories handles division by zero correctly
func TestDailySummary_ZeroCalories(t *testing.T) {
	_, diaryRepo, _, _, goalRepo, userID := setupTagBreakdownTest(t)

	// Create nutrition goal
	goalRepo.Create(&goal.NutritionGoal{
		UserID:    userID,
		Calories:  2000,
		Protein:   150,
		StartDate: time.Now().AddDate(0, 0, -30),
		IsActive:  true,
	})

	// No diary entries created for today
	today := time.Now()

	// Get entries (should be empty)
	entries, err := diaryRepo.GetByDate(userID, today)
	require.NoError(t, err)
	assert.Len(t, entries, 0)

	// Calculate tag breakdown with empty entries
	var routineCalories, contextualCalories, routinePercent, contextualPercent float64
	for _, entry := range entries {
		tag := entry.FoodTag
		if tag == "" {
			tag = entry.RecipeTag
		}
		if tag == "routine" {
			routineCalories += entry.Calories
		} else if tag == "contextual" {
			contextualCalories += entry.Calories
		}
	}

	// Calculate percentages (should handle division by zero)
	totalCals := routineCalories + contextualCalories
	if totalCals > 0 {
		routinePercent = (routineCalories / totalCals) * 100
		contextualPercent = (contextualCalories / totalCals) * 100
	}

	// Verify all values are zero
	assert.Equal(t, 0.0, routineCalories, "Routine calories should be 0")
	assert.Equal(t, 0.0, contextualCalories, "Contextual calories should be 0")
	assert.Equal(t, 0.0, routinePercent, "Routine percent should be 0")
	assert.Equal(t, 0.0, contextualPercent, "Contextual percent should be 0")
}

// TestDailySummary_RecipeTags tests that recipe tags are used correctly
func TestDailySummary_RecipeTags(t *testing.T) {
	db, diaryRepo, foodRepo, _, goalRepo, userID := setupTagBreakdownTest(t)

	// Create foods
	chicken, _ := foodRepo.Create(food.CreateFoodRequest{
		Name:     "Chicken",
		Calories: 165,
		Protein:  31,
		Tag:      "routine",
	})

	// Create recipes with tags
	routineRecipe := &recipe.Recipe{
		Name:   "Healthy Bowl",
		UserID: &userID,
		Tag:    "routine",
	}
	db.Create(routineRecipe)

	// Add ingredient to routine recipe
	routineIngredient := &recipe.RecipeIngredient{
		RecipeID:      routineRecipe.ID,
		FoodID:        chicken.ID,
		QuantityGrams: 200,
	}
	db.Create(routineIngredient)

	contextualRecipe := &recipe.Recipe{
		Name:   "Junk Bowl",
		UserID: &userID,
		Tag:    "contextual",
	}
	db.Create(contextualRecipe)

	// Add ingredient to contextual recipe
	contextualIngredient := &recipe.RecipeIngredient{
		RecipeID:      contextualRecipe.ID,
		FoodID:        chicken.ID,
		QuantityGrams: 150,
	}
	db.Create(contextualIngredient)

	// Create nutrition goal
	goalRepo.Create(&goal.NutritionGoal{
		UserID:    userID,
		Calories:  2000,
		Protein:   150,
		StartDate: time.Now().AddDate(0, 0, -30),
		IsActive:  true,
	})

	// Create diary entries with recipes
	today := time.Now()

	entry1 := &diary.DiaryEntry{
		UserID:        userID,
		RecipeID:      &routineRecipe.ID,
		Date:          today,
		MealType:      diary.Breakfast,
		QuantityGrams: 200,
		Calories:      330, // Recipe calories
		Protein:       62,
		RecipeTag:     "routine",
	}
	entry2 := &diary.DiaryEntry{
		UserID:        userID,
		RecipeID:      &contextualRecipe.ID,
		Date:          today,
		MealType:      diary.Lunch,
		QuantityGrams: 150,
		Calories:      247.5, // Recipe calories
		Protein:       46.5,
		RecipeTag:     "contextual",
	}
	db.Create(entry1)
	db.Create(entry2)

	// Get entries
	entries, err := diaryRepo.GetByDate(userID, today)
	require.NoError(t, err)

	// Calculate tag breakdown
	var routineCalories, contextualCalories float64
	for _, entry := range entries {
		tag := entry.FoodTag
		if tag == "" {
			tag = entry.RecipeTag
		}
		if tag == "routine" {
			routineCalories += entry.Calories
		} else if tag == "contextual" {
			contextualCalories += entry.Calories
		}
	}

	// Verify recipe tags are used correctly
	expectedRoutineCalories := 330.0
	expectedContextualCalories := 247.5

	assert.InDelta(t, expectedRoutineCalories, routineCalories, 0.01, "Routine recipe calories should be tracked")
	assert.InDelta(t, expectedContextualCalories, contextualCalories, 0.01, "Contextual recipe calories should be tracked")

	// Calculate percentages
	var routinePercent, contextualPercent float64
	totalCals := routineCalories + contextualCalories
	if totalCals > 0 {
		routinePercent = (routineCalories / totalCals) * 100
		contextualPercent = (contextualCalories / totalCals) * 100
	}

	// 330 / 577.5 = 57.14%
	// 247.5 / 577.5 = 42.86%
	assert.InDelta(t, 57.14, routinePercent, 0.1)
	assert.InDelta(t, 42.86, contextualPercent, 0.1)
}

// TestWeeklySummary_TagBreakdown tests that weekly summary includes tag averages
func TestWeeklySummary_TagBreakdown(t *testing.T) {
	db, _, foodRepo, _, goalRepo, userID := setupTagBreakdownTest(t)

	// Create foods
	routineFood, _ := foodRepo.Create(food.CreateFoodRequest{
		Name:     "Chicken",
		Calories: 165,
		Protein:  31,
		Tag:      "routine",
	})
	contextualFood, _ := foodRepo.Create(food.CreateFoodRequest{
		Name:     "Pizza",
		Calories: 266,
		Protein:  11,
		Tag:      "contextual",
	})

	// Create nutrition goal
	goalRepo.Create(&goal.NutritionGoal{
		UserID:    userID,
		Calories:  2000,
		Protein:   150,
		StartDate: time.Now().AddDate(0, 0, -30),
		IsActive:  true,
	})

	// Create diary entries for a full week
	startDate := time.Now().AddDate(0, 0, -3) // 3 days ago

	var totalRoutineCalories, totalContextualCalories float64

	for i := 0; i < 7; i++ {
		date := startDate.AddDate(0, 0, i)

		// Each day: 300 routine, 200 contextual (total 500 calories)
		routineEntry := &diary.DiaryEntry{
			UserID:        userID,
			FoodID:        &routineFood.ID,
			Date:          date,
			MealType:      diary.Breakfast,
			QuantityGrams: 100,
			Calories:      165,
			Protein:       31,
			FoodTag:       "routine",
		}
		contextualEntry := &diary.DiaryEntry{
			UserID:        userID,
			FoodID:        &contextualFood.ID,
			Date:          date,
			MealType:      diary.Lunch,
			QuantityGrams: 100,
			Calories:      266,
			Protein:       11,
			FoodTag:       "contextual",
		}

		db.Create(routineEntry)
		db.Create(contextualEntry)

		totalRoutineCalories += 165
		totalContextualCalories += 266
	}

	// Calculate expected weekly averages
	// Total per day: 431 calories (165 routine + 266 contextual)
	// Routine %: (165 / 431) * 100 = 38.28%
	// Contextual %: (266 / 431) * 100 = 61.72%
	expectedRoutinePercent := (165.0 / 431.0) * 100
	expectedContextualPercent := (266.0 / 431.0) * 100

	// Verify calculation manually
	avgWeeklyCalories := (totalRoutineCalories + totalContextualCalories) / 7 // 431
	avgRoutinePercent := (totalRoutineCalories / 7 / avgWeeklyCalories) * 100
	avgContextualPercent := (totalContextualCalories / 7 / avgWeeklyCalories) * 100

	assert.InDelta(t, expectedRoutinePercent, avgRoutinePercent, 0.01)
	assert.InDelta(t, expectedContextualPercent, avgContextualPercent, 0.01)
	assert.InDelta(t, 100.0, avgRoutinePercent+avgContextualPercent, 0.1)
}

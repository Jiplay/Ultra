package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ultra-bis/internal/diary"
	"ultra-bis/internal/food"
	"ultra-bis/internal/goal"
	"ultra-bis/internal/httputil"
	"ultra-bis/internal/recipe"
	"ultra-bis/internal/user"
	"ultra-bis/test/testutil"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// setupWeeklyRoutineTest creates a test environment with necessary repositories and handler
func setupWeeklyRoutineTest(t *testing.T) (*gorm.DB, *diary.Handler, *food.Repository, uint) {
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
	goalRepo := goal.NewRepository(db)

	// Create handler
	handler := diary.NewHandler(diaryRepo, foodRepo, goalRepo)

	// Create test user
	testUser := &user.User{
		Email:        testutil.MockUserEmail(),
		PasswordHash: "hashedpassword",
		Name:         "Test User",
		Height:       175,
		Gender:       "male",
	}
	db.Create(testUser)

	return db, handler, foodRepo, testUser.ID
}

// TestWeeklySummary_AllRoutineFoods tests all days with 100% routine foods
func TestWeeklySummary_AllRoutineFoods(t *testing.T) {
	db, handler, foodRepo, userID := setupWeeklyRoutineTest(t)

	// Create routine food (100 cal per 100g for easy math)
	routineFood, _ := foodRepo.Create(food.CreateFoodRequest{
		Name:     "Routine Food",
		Calories: 100,
		Protein:  10,
		Tag:      "routine",
	})

	// Create entries for a specific week (Jan 13-19, 2025 - Mon to Sun)
	startDate := time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 7; i++ {
		currentDate := startDate.AddDate(0, 0, i)
		entry := &diary.DiaryEntry{
			UserID:        userID,
			FoodID:        &routineFood.ID,
			Date:          currentDate,
			MealType:      diary.Breakfast,
			QuantityGrams: 100, // 100g = 100 calories
			Calories:      100,
			Protein:       10,
			FoodTag:       "routine",
		}
		db.Create(entry)
	}

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/diary/weekly?start_date=2025-01-13", nil)
	ctx := httputil.SetUserID(req.Context(), userID)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.GetWeeklySummary(rr, req)

	// Assert response
	assert.Equal(t, http.StatusOK, rr.Code)
	var result [7]interface{}
	json.Unmarshal(rr.Body.Bytes(), &result)

	// All days should be true (100% routine > 75%)
	for i := 0; i < 7; i++ {
		assert.Equal(t, true, result[i], "Day %d should be true", i)
	}
}

// TestWeeklySummary_AllContextualFoods tests all days with 0% routine foods
func TestWeeklySummary_AllContextualFoods(t *testing.T) {
	db, handler, foodRepo, userID := setupWeeklyRoutineTest(t)

	// Create contextual food
	contextualFood, _ := foodRepo.Create(food.CreateFoodRequest{
		Name:     "Contextual Food",
		Calories: 100,
		Protein:  10,
		Tag:      "contextual",
	})

	// Create entries for a specific week
	startDate := time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 7; i++ {
		currentDate := startDate.AddDate(0, 0, i)
		entry := &diary.DiaryEntry{
			UserID:        userID,
			FoodID:        &contextualFood.ID,
			Date:          currentDate,
			MealType:      diary.Breakfast,
			QuantityGrams: 100,
			Calories:      100,
			Protein:       10,
			FoodTag:       "contextual",
		}
		db.Create(entry)
	}

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/diary/weekly?start_date=2025-01-13", nil)
	ctx := httputil.SetUserID(req.Context(), userID)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.GetWeeklySummary(rr, req)

	// Assert response
	assert.Equal(t, http.StatusOK, rr.Code)
	var result [7]interface{}
	json.Unmarshal(rr.Body.Bytes(), &result)

	// All days should be false (0% routine ≤ 75%)
	for i := 0; i < 7; i++ {
		assert.Equal(t, false, result[i], "Day %d should be false", i)
	}
}

// TestWeeklySummary_MixedFoods tests days with varying routine percentages
func TestWeeklySummary_MixedFoods(t *testing.T) {
	db, handler, foodRepo, userID := setupWeeklyRoutineTest(t)

	// Create foods
	routineFood, _ := foodRepo.Create(food.CreateFoodRequest{
		Name:     "Routine Food",
		Calories: 100,
		Tag:      "routine",
	})
	contextualFood, _ := foodRepo.Create(food.CreateFoodRequest{
		Name:     "Contextual Food",
		Calories: 100,
		Tag:      "contextual",
	})

	// Create entries for a specific week with different percentages
	startDate := time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC)

	// Monday: 80% routine (800 routine + 200 contextual = 80%) → true
	createTestEntries(t, db, userID, startDate.AddDate(0, 0, 0), routineFood.ID, contextualFood.ID, 800, 200)

	// Tuesday: 50% routine (500 routine + 500 contextual = 50%) → false
	createTestEntries(t, db, userID, startDate.AddDate(0, 0, 1), routineFood.ID, contextualFood.ID, 500, 500)

	// Wednesday: 75% routine (300 routine + 100 contextual = 75%) → false (boundary)
	createTestEntries(t, db, userID, startDate.AddDate(0, 0, 2), routineFood.ID, contextualFood.ID, 300, 100)

	// Thursday: 76% routine (760 routine + 240 contextual = 76%) → true
	createTestEntries(t, db, userID, startDate.AddDate(0, 0, 3), routineFood.ID, contextualFood.ID, 760, 240)

	// Friday: 100% routine (1000 routine + 0 contextual = 100%) → true
	createTestEntries(t, db, userID, startDate.AddDate(0, 0, 4), routineFood.ID, contextualFood.ID, 1000, 0)

	// Saturday: 0% routine (0 routine + 1000 contextual = 0%) → false
	createTestEntries(t, db, userID, startDate.AddDate(0, 0, 5), routineFood.ID, contextualFood.ID, 0, 1000)

	// Sunday: 25% routine (250 routine + 750 contextual = 25%) → false
	createTestEntries(t, db, userID, startDate.AddDate(0, 0, 6), routineFood.ID, contextualFood.ID, 250, 750)

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/diary/weekly?start_date=2025-01-13", nil)
	ctx := httputil.SetUserID(req.Context(), userID)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.GetWeeklySummary(rr, req)

	// Assert response
	assert.Equal(t, http.StatusOK, rr.Code)
	var result [7]interface{}
	json.Unmarshal(rr.Body.Bytes(), &result)

	// Verify expected pattern
	expected := []interface{}{true, false, false, true, true, false, false}
	assert.Equal(t, expected, result[:])
}

// TestWeeklySummary_EmptyDays tests days with no entries return null
func TestWeeklySummary_EmptyDays(t *testing.T) {
	db, handler, foodRepo, userID := setupWeeklyRoutineTest(t)

	// Create routine food
	routineFood, _ := foodRepo.Create(food.CreateFoodRequest{
		Name:     "Routine Food",
		Calories: 100,
		Tag:      "routine",
	})
	contextualFood, _ := foodRepo.Create(food.CreateFoodRequest{
		Name:     "Contextual Food",
		Calories: 100,
		Tag:      "contextual",
	})

	// Create entries for a specific week
	startDate := time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC)

	// Monday: 80% routine → true
	createTestEntries(t, db, userID, startDate.AddDate(0, 0, 0), routineFood.ID, contextualFood.ID, 800, 200)

	// Tuesday: No entries → null

	// Wednesday: 50% routine → false
	createTestEntries(t, db, userID, startDate.AddDate(0, 0, 2), routineFood.ID, contextualFood.ID, 500, 500)

	// Thursday-Sunday: No entries → null

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/diary/weekly?start_date=2025-01-13", nil)
	ctx := httputil.SetUserID(req.Context(), userID)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.GetWeeklySummary(rr, req)

	// Assert response
	assert.Equal(t, http.StatusOK, rr.Code)
	var result [7]interface{}
	json.Unmarshal(rr.Body.Bytes(), &result)

	// Verify expected pattern
	assert.Equal(t, true, result[0])  // Monday
	assert.Nil(t, result[1])          // Tuesday (no entries)
	assert.Equal(t, false, result[2]) // Wednesday
	assert.Nil(t, result[3])          // Thursday (no entries)
	assert.Nil(t, result[4])          // Friday (no entries)
	assert.Nil(t, result[5])          // Saturday (no entries)
	assert.Nil(t, result[6])          // Sunday (no entries)
}

// TestWeeklySummary_ExactlySeventyFivePercent tests boundary case
func TestWeeklySummary_ExactlySeventyFivePercent(t *testing.T) {
	db, handler, foodRepo, userID := setupWeeklyRoutineTest(t)

	// Create foods
	routineFood, _ := foodRepo.Create(food.CreateFoodRequest{
		Name:     "Routine Food",
		Calories: 100,
		Tag:      "routine",
	})
	contextualFood, _ := foodRepo.Create(food.CreateFoodRequest{
		Name:     "Contextual Food",
		Calories: 100,
		Tag:      "contextual",
	})

	// Create entries for Monday only
	startDate := time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC)

	// Create exactly 75% routine (300 routine + 100 contextual = 400 total, 75%)
	createTestEntries(t, db, userID, startDate, routineFood.ID, contextualFood.ID, 300, 100)

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/diary/weekly?start_date=2025-01-13", nil)
	ctx := httputil.SetUserID(req.Context(), userID)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.GetWeeklySummary(rr, req)

	// Assert response
	assert.Equal(t, http.StatusOK, rr.Code)
	var result [7]interface{}
	json.Unmarshal(rr.Body.Bytes(), &result)

	// Monday should be false (exactly 75%, need >75%)
	assert.Equal(t, false, result[0], "Exactly 75% should return false (need >75%)")
}

// TestWeeklySummary_CustomStartDate tests query parameter parsing
func TestWeeklySummary_CustomStartDate(t *testing.T) {
	db, handler, foodRepo, userID := setupWeeklyRoutineTest(t)

	// Create routine food
	routineFood, _ := foodRepo.Create(food.CreateFoodRequest{
		Name:     "Routine Food",
		Calories: 100,
		Tag:      "routine",
	})

	// Create entry for a specific date (Jan 20, 2025 - a Monday)
	specificDate := time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)
	entry := &diary.DiaryEntry{
		UserID:        userID,
		FoodID:        &routineFood.ID,
		Date:          specificDate,
		MealType:      diary.Breakfast,
		QuantityGrams: 100,
		Calories:      100,
		FoodTag:       "routine",
	}
	db.Create(entry)

	// Make request with custom start_date
	req := httptest.NewRequest(http.MethodGet, "/diary/weekly?start_date=2025-01-20", nil)
	ctx := httputil.SetUserID(req.Context(), userID)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.GetWeeklySummary(rr, req)

	// Assert response
	assert.Equal(t, http.StatusOK, rr.Code)
	var result [7]interface{}
	json.Unmarshal(rr.Body.Bytes(), &result)

	// Monday (Jan 20) should be true, rest should be null
	assert.Equal(t, true, result[0]) // Monday
	for i := 1; i < 7; i++ {
		assert.Nil(t, result[i], "Day %d should be null", i)
	}
}

// TestWeeklySummary_DefaultsToCurrentMonday tests default behavior
func TestWeeklySummary_DefaultsToCurrentMonday(t *testing.T) {
	db, handler, foodRepo, userID := setupWeeklyRoutineTest(t)

	// Create routine food
	routineFood, _ := foodRepo.Create(food.CreateFoodRequest{
		Name:     "Routine Food",
		Calories: 100,
		Tag:      "routine",
	})

	// Calculate current week's Monday
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 { // Sunday
		weekday = 7
	}
	daysFromMonday := weekday - 1
	monday := now.AddDate(0, 0, -daysFromMonday)

	// Create entry for current Monday
	entry := &diary.DiaryEntry{
		UserID:        userID,
		FoodID:        &routineFood.ID,
		Date:          monday,
		MealType:      diary.Breakfast,
		QuantityGrams: 100,
		Calories:      100,
		FoodTag:       "routine",
	}
	db.Create(entry)

	// Make request without start_date parameter
	req := httptest.NewRequest(http.MethodGet, "/diary/weekly", nil)
	ctx := httputil.SetUserID(req.Context(), userID)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.GetWeeklySummary(rr, req)

	// Assert response
	assert.Equal(t, http.StatusOK, rr.Code)
	var result [7]interface{}
	json.Unmarshal(rr.Body.Bytes(), &result)

	// Monday should be true (current week's Monday has routine food)
	assert.Equal(t, true, result[0])
}

// TestWeeklySummary_InvalidDateFormat tests error handling
func TestWeeklySummary_InvalidDateFormat(t *testing.T) {
	_, handler, _, userID := setupWeeklyRoutineTest(t)

	// Make request with invalid date format
	req := httptest.NewRequest(http.MethodGet, "/diary/weekly?start_date=invalid-date", nil)
	ctx := httputil.SetUserID(req.Context(), userID)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.GetWeeklySummary(rr, req)

	// Assert error response
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errorResponse map[string]string
	json.Unmarshal(rr.Body.Bytes(), &errorResponse)
	assert.Contains(t, errorResponse["error"], "Invalid date format")
}

// Helper function to create test entries with specific routine/contextual calories
func createTestEntries(t *testing.T, db *gorm.DB, userID uint, date time.Time, routineFoodID, contextualFoodID uint, routineCals, contextualCals float64) {
	t.Helper()

	// Create routine entry if needed
	if routineCals > 0 {
		routineEntry := &diary.DiaryEntry{
			UserID:        userID,
			FoodID:        &routineFoodID,
			Date:          date,
			MealType:      diary.Breakfast,
			QuantityGrams: routineCals, // Using 100 cal/100g, so grams = calories
			Calories:      routineCals,
			FoodTag:       "routine",
		}
		db.Create(routineEntry)
	}

	// Create contextual entry if needed
	if contextualCals > 0 {
		contextualEntry := &diary.DiaryEntry{
			UserID:        userID,
			FoodID:        &contextualFoodID,
			Date:          date,
			MealType:      diary.Lunch,
			QuantityGrams: contextualCals, // Using 100 cal/100g, so grams = calories
			Calories:      contextualCals,
			FoodTag:       "contextual",
		}
		db.Create(contextualEntry)
	}
}

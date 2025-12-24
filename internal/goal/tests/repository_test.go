package tests

import (
	"testing"
	"time"

	"ultra-bis/internal/goal"

	"github.com/stretchr/testify/assert"

	"ultra-bis/internal/user"
	"ultra-bis/test/testutil"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupGoalTest creates a test DB with Goal and User migrations
func setupGoalTest(t *testing.T) (*gorm.DB, *goal.Repository) {
	t.Helper()
	db := testutil.SetupTestDB(t)

	// Migrate both User and NutritionGoal models
	if err := db.AutoMigrate(&user.User{}, &goal.NutritionGoal{}); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return db, goal.NewRepository(db)
}

// TestRepository_Create tests creating a new nutrition goal
func TestRepository_Create(t *testing.T) {
	db, repo := setupGoalTest(t)

	// Create test user
	testUser := testutil.CreateTestUser(t, db)

	mygoal := &goal.NutritionGoal{
		UserID:    testUser.ID,
		Calories:  2000,
		Protein:   150,
		Carbs:     200,
		Fat:       65,
		Fiber:     30,
		StartDate: time.Now(),
		IsActive:  true,
	}

	err := repo.Create(mygoal)

	assert.NoError(t, err)
	assert.NotZero(t, mygoal.ID, "Goal ID should be set after creation")

	// Verify goal was created in database
	var found goal.NutritionGoal
	db.First(&found, mygoal.ID)
	assert.Equal(t, testUser.ID, found.UserID)
	testutil.AssertNutritionEquals(t, 2000, found.Calories, "calories")
	testutil.AssertNutritionEquals(t, 150, found.Protein, "protein")
	assert.True(t, found.IsActive)
}

// TestRepository_Create_DeactivatesPreviousGoal tests that creating a new goal deactivates previous active goals
func TestRepository_Create_DeactivatesPreviousGoal(t *testing.T) {
	db, repo := setupGoalTest(t)

	testUser := testutil.CreateTestUser(t, db)

	// Create first goal (active)
	goal1 := &goal.NutritionGoal{
		UserID:    testUser.ID,
		Calories:  2000,
		Protein:   150,
		Carbs:     200,
		Fat:       65,
		Fiber:     30,
		StartDate: time.Now().AddDate(0, 0, -7),
		IsActive:  true,
	}
	err := repo.Create(goal1)
	require.NoError(t, err)

	// Verify first goal is active
	var firstGoal goal.NutritionGoal
	db.First(&firstGoal, goal1.ID)
	assert.True(t, firstGoal.IsActive, "First goal should be active initially")

	// Create second goal (should auto-deactivate first)
	goal2 := &goal.NutritionGoal{
		UserID:    testUser.ID,
		Calories:  2200,
		Protein:   165,
		Carbs:     220,
		Fat:       73,
		Fiber:     32,
		StartDate: time.Now(),
		IsActive:  true,
	}
	err = repo.Create(goal2)
	require.NoError(t, err)

	// Verify first goal is now inactive
	db.First(&firstGoal, goal1.ID)
	assert.False(t, firstGoal.IsActive, "First goal should be deactivated")

	// Verify second goal is active
	var secondGoal goal.NutritionGoal
	db.First(&secondGoal, goal2.ID)
	assert.True(t, secondGoal.IsActive, "Second goal should be active")

	// Verify only one active goal exists for this user
	var activeCount int64
	db.Model(&goal.NutritionGoal{}).Where("user_id = ? AND is_active = ?", testUser.ID, true).Count(&activeCount)
	assert.Equal(t, int64(1), activeCount, "Only one goal should be active")
}

// TestRepository_Create_MultipleUsers tests that deactivation only affects the same user's goals
func TestRepository_Create_MultipleUsers(t *testing.T) {
	db, repo := setupGoalTest(t)

	user1 := testutil.CreateTestUser(t, db, "user1@example.com")
	user2 := testutil.CreateTestUser(t, db, "user2@example.com")

	// Create goal for user1
	goal1 := &goal.NutritionGoal{
		UserID:    user1.ID,
		Calories:  2000,
		Protein:   150,
		Carbs:     200,
		Fat:       65,
		Fiber:     30,
		StartDate: time.Now(),
		IsActive:  true,
	}
	err := repo.Create(goal1)
	require.NoError(t, err)

	// Create goal for user2
	goal2 := &goal.NutritionGoal{
		UserID:    user2.ID,
		Calories:  1800,
		Protein:   130,
		Carbs:     180,
		Fat:       55,
		Fiber:     28,
		StartDate: time.Now(),
		IsActive:  true,
	}
	err = repo.Create(goal2)
	require.NoError(t, err)

	// Both goals should remain active (different users)
	var goal1Check, goal2Check goal.NutritionGoal
	db.First(&goal1Check, goal1.ID)
	db.First(&goal2Check, goal2.ID)

	assert.True(t, goal1Check.IsActive, "User1's goal should still be active")
	assert.True(t, goal2Check.IsActive, "User2's goal should be active")
}

// TestRepository_GetActive tests retrieving the active goal for a user
func TestRepository_GetActive(t *testing.T) {
	db, repo := setupGoalTest(t)

	testUser := testutil.CreateTestUser(t, db)

	// Create active goal directly
	testGoal := &goal.NutritionGoal{
		UserID:    testUser.ID,
		Calories:  2000,
		Protein:   150,
		Carbs:     200,
		Fat:       65,
		Fiber:     30,
		StartDate: time.Now(),
		IsActive:  true,
	}
	db.Create(testGoal)

	found, err := repo.GetActive(testUser.ID)

	assert.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, testGoal.ID, found.ID)
	assert.True(t, found.IsActive)
}

// TestRepository_GetActive_NoActiveGoal tests retrieving when no active goal exists
func TestRepository_GetActive_NoActiveGoal(t *testing.T) {
	db, repo := setupGoalTest(t)

	testUser := testutil.CreateTestUser(t, db)

	// Create inactive goal - explicitly set to false after creation to override GORM default
	inactiveGoal := &goal.NutritionGoal{
		UserID:    testUser.ID,
		Calories:  2000,
		Protein:   150,
		Carbs:     200,
		Fat:       65,
		Fiber:     30,
		StartDate: time.Now(),
		IsActive:  false,
	}
	db.Create(inactiveGoal)

	// Explicitly update IsActive to false (GORM default may have set it to true)
	db.Model(inactiveGoal).Update("is_active", false)

	found, err := repo.GetActive(testUser.ID)

	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "no active goal found")
	}
	assert.Nil(t, found)
}

// TestRepository_GetActive_WrongUser tests that users can't access other users' goals
func TestRepository_GetActive_WrongUser(t *testing.T) {
	db, repo := setupGoalTest(t)

	user1 := testutil.CreateTestUser(t, db, "user1@example.com")
	user2 := testutil.CreateTestUser(t, db, "user2@example.com")

	// Create goal for user1
	goal := &goal.NutritionGoal{
		UserID:    user1.ID,
		Calories:  2000,
		Protein:   150,
		Carbs:     200,
		Fat:       65,
		Fiber:     30,
		StartDate: time.Now(),
		IsActive:  true,
	}
	db.Create(goal)

	// Try to get active goal for user2
	found, err := repo.GetActive(user2.ID)

	assert.Error(t, err)
	assert.Nil(t, found)
}

// TestRepository_GetByID tests retrieving a goal by ID
func TestRepository_GetByID(t *testing.T) {
	db, repo := setupGoalTest(t)

	testUser := testutil.CreateTestUser(t, db)
	testGoal := &goal.NutritionGoal{
		UserID:    testUser.ID,
		Calories:  2000,
		Protein:   150,
		Carbs:     200,
		Fat:       65,
		Fiber:     30,
		StartDate: time.Now(),
		IsActive:  true,
	}
	db.Create(testGoal)

	found, err := repo.GetByID(testGoal.ID, testUser.ID)

	assert.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, testGoal.ID, found.ID)
	assert.Equal(t, testUser.ID, found.UserID)
}

// TestRepository_GetByID_NotFound tests retrieving a non-existent goal
func TestRepository_GetByID_NotFound(t *testing.T) {
	db, repo := setupGoalTest(t)

	testUser := testutil.CreateTestUser(t, db)

	found, err := repo.GetByID(99999, testUser.ID)

	assert.Error(t, err)
	assert.Nil(t, found)
	assert.Contains(t, err.Error(), "goal not found")
}

// TestRepository_GetByID_WrongUser tests that users can't access other users' goals by ID
func TestRepository_GetByID_WrongUser(t *testing.T) {
	db, repo := setupGoalTest(t)

	user1 := testutil.CreateTestUser(t, db, "user1@example.com")
	user2 := testutil.CreateTestUser(t, db, "user2@example.com")

	// Create goal for user1
	goal := &goal.NutritionGoal{
		UserID:    user1.ID,
		Calories:  2000,
		Protein:   150,
		Carbs:     200,
		Fat:       65,
		Fiber:     30,
		StartDate: time.Now(),
		IsActive:  true,
	}
	db.Create(goal)

	// Try to get goal with user2's ID
	found, err := repo.GetByID(goal.ID, user2.ID)

	assert.Error(t, err)
	assert.Nil(t, found)
	assert.Contains(t, err.Error(), "goal not found")
}

// TestRepository_GetAll tests retrieving all goals for a user
func TestRepository_GetAll(t *testing.T) {
	db, repo := setupGoalTest(t)

	testUser := testutil.CreateTestUser(t, db)

	// Create multiple goals
	goals := make([]*goal.NutritionGoal, 3)
	for i := 0; i < 3; i++ {
		goal := &goal.NutritionGoal{
			UserID:    testUser.ID,
			Calories:  float64(1800 + (i * 100)),
			Protein:   float64(130 + (i * 10)),
			Carbs:     float64(180 + (i * 20)),
			Fat:       float64(55 + (i * 5)),
			Fiber:     float64(25 + (i * 2)),
			StartDate: time.Now().AddDate(0, 0, -3+i),
			IsActive:  (i == 2),
		}
		db.Create(goal)
		goals[i] = goal
	}

	found, err := repo.GetAll(testUser.ID)

	assert.NoError(t, err)
	assert.Len(t, found, 3, "Should return all 3 goals")

	// Verify ordering (most recent first)
	assert.Equal(t, goals[2].ID, found[0].ID, "Most recent goal should be first")
}

// TestRepository_GetAll_Empty tests retrieving goals when user has none
func TestRepository_GetAll_Empty(t *testing.T) {
	db, repo := setupGoalTest(t)

	testUser := testutil.CreateTestUser(t, db)

	found, err := repo.GetAll(testUser.ID)

	assert.NoError(t, err)
	assert.Empty(t, found, "Should return empty slice")
}

// TestRepository_GetAll_OnlyUserGoals tests that GetAll only returns the specified user's goals
func TestRepository_GetAll_OnlyUserGoals(t *testing.T) {
	db, repo := setupGoalTest(t)

	user1 := testutil.CreateTestUser(t, db, "user1@example.com")
	user2 := testutil.CreateTestUser(t, db, "user2@example.com")

	// Create goals for user1
	for i := 0; i < 2; i++ {
		goal := &goal.NutritionGoal{
			UserID:    user1.ID,
			Calories:  2000,
			Protein:   150,
			Carbs:     200,
			Fat:       65,
			Fiber:     30,
			StartDate: time.Now().AddDate(0, 0, -i),
			IsActive:  (i == 1),
		}
		db.Create(goal)
	}

	// Create goals for user2
	for i := 0; i < 3; i++ {
		goal := &goal.NutritionGoal{
			UserID:    user2.ID,
			Calories:  1800,
			Protein:   130,
			Carbs:     180,
			Fat:       55,
			Fiber:     28,
			StartDate: time.Now().AddDate(0, 0, -i),
			IsActive:  (i == 2),
		}
		db.Create(goal)
	}

	// Get goals for user1
	found, err := repo.GetAll(user1.ID)

	assert.NoError(t, err)
	assert.Len(t, found, 2, "User1 should only see their 2 goals")

	// Verify all returned goals belong to user1
	for _, goal := range found {
		assert.Equal(t, user1.ID, goal.UserID, "All goals should belong to user1")
	}
}

// TestRepository_Update tests updating a goal
func TestRepository_Update(t *testing.T) {
	db, repo := setupGoalTest(t)

	testUser := testutil.CreateTestUser(t, db)
	testGoal := &goal.NutritionGoal{
		UserID:    testUser.ID,
		Calories:  2000,
		Protein:   150,
		Carbs:     200,
		Fat:       65,
		Fiber:     30,
		StartDate: time.Now(),
		IsActive:  true,
	}
	db.Create(testGoal)

	// Update goal values
	testGoal.Calories = 2200
	testGoal.Protein = 165
	testGoal.Carbs = 220
	testGoal.Fat = 73

	err := repo.Update(testGoal)

	assert.NoError(t, err)

	// Verify updates in database
	var updated goal.NutritionGoal
	db.First(&updated, testGoal.ID)
	testutil.AssertNutritionEquals(t, 2200, updated.Calories, "calories")
	testutil.AssertNutritionEquals(t, 165, updated.Protein, "protein")
	testutil.AssertNutritionEquals(t, 220, updated.Carbs, "carbs")
	testutil.AssertNutritionEquals(t, 73, updated.Fat, "fat")
}

// TestRepository_Update_EndDate tests updating a goal's end date
func TestRepository_Update_EndDate(t *testing.T) {
	db, repo := setupGoalTest(t)

	testUser := testutil.CreateTestUser(t, db)
	testGoal := &goal.NutritionGoal{
		UserID:    testUser.ID,
		Calories:  2000,
		Protein:   150,
		Carbs:     200,
		Fat:       65,
		Fiber:     30,
		StartDate: time.Now(),
		IsActive:  true,
	}
	db.Create(testGoal)

	// Add end date
	endDate := time.Now().AddDate(0, 0, 30)
	testGoal.EndDate = &endDate

	err := repo.Update(testGoal)

	assert.NoError(t, err)

	// Verify end date was set
	var updated goal.NutritionGoal
	db.First(&updated, testGoal.ID)
	require.NotNil(t, updated.EndDate)
	assert.Equal(t, endDate.Format("2006-01-02"), updated.EndDate.Format("2006-01-02"))
}

// TestRepository_Delete tests soft deleting a goal
func TestRepository_Delete(t *testing.T) {
	db, repo := setupGoalTest(t)

	testUser := testutil.CreateTestUser(t, db)
	testGoal := &goal.NutritionGoal{
		UserID:    testUser.ID,
		Calories:  2000,
		Protein:   150,
		Carbs:     200,
		Fat:       65,
		Fiber:     30,
		StartDate: time.Now(),
		IsActive:  true,
	}
	db.Create(testGoal)

	err := repo.Delete(testGoal.ID, testUser.ID)

	assert.NoError(t, err)

	// Verify goal is soft deleted (can't be found with normal query)
	found, err := repo.GetByID(testGoal.ID, testUser.ID)
	assert.Error(t, err)
	assert.Nil(t, found)

	// Verify goal still exists in database with DeletedAt set
	var deleted goal.NutritionGoal
	db.Unscoped().First(&deleted, testGoal.ID)
	assert.NotNil(t, deleted.DeletedAt)
}

// TestRepository_Delete_NotFound tests deleting a non-existent goal
func TestRepository_Delete_NotFound(t *testing.T) {
	db, repo := setupGoalTest(t)

	testUser := testutil.CreateTestUser(t, db)

	err := repo.Delete(99999, testUser.ID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "goal not found")
}

// TestRepository_Delete_WrongUser tests that users can't delete other users' goals
func TestRepository_Delete_WrongUser(t *testing.T) {
	db, repo := setupGoalTest(t)

	user1 := testutil.CreateTestUser(t, db, "user1@example.com")
	user2 := testutil.CreateTestUser(t, db, "user2@example.com")

	// Create goal for user1
	goal := &goal.NutritionGoal{
		UserID:    user1.ID,
		Calories:  2000,
		Protein:   150,
		Carbs:     200,
		Fat:       65,
		Fiber:     30,
		StartDate: time.Now(),
		IsActive:  true,
	}
	db.Create(goal)

	// Try to delete with user2's ID
	err := repo.Delete(goal.ID, user2.ID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "goal not found")

	// Verify goal still exists for user1
	found, err := repo.GetByID(goal.ID, user1.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found)
}

// TestRepository_GetForDate tests retrieving a goal for a specific date
func TestRepository_GetForDate(t *testing.T) {
	db, repo := setupGoalTest(t)

	testUser := testutil.CreateTestUser(t, db)

	// Create goal with date range
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)
	goal := &goal.NutritionGoal{
		UserID:    testUser.ID,
		Calories:  2000,
		Protein:   150,
		Carbs:     200,
		Fat:       65,
		Fiber:     30,
		StartDate: startDate,
		EndDate:   &endDate,
		IsActive:  true,
	}
	db.Create(goal)

	// Test date within range
	testDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	found, err := repo.GetForDate(testUser.ID, testDate)

	assert.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, goal.ID, found.ID)
}

// TestRepository_GetForDate_StartDateBoundary tests getting goal on start date
func TestRepository_GetForDate_StartDateBoundary(t *testing.T) {
	db, repo := setupGoalTest(t)

	testUser := testutil.CreateTestUser(t, db)

	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)
	goal := &goal.NutritionGoal{
		UserID:    testUser.ID,
		Calories:  2000,
		Protein:   150,
		Carbs:     200,
		Fat:       65,
		Fiber:     30,
		StartDate: startDate,
		EndDate:   &endDate,
		IsActive:  true,
	}
	db.Create(goal)

	// Test on exact start date
	found, err := repo.GetForDate(testUser.ID, startDate)

	assert.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, goal.ID, found.ID)
}

// TestRepository_GetForDate_EndDateBoundary tests getting goal on end date
func TestRepository_GetForDate_EndDateBoundary(t *testing.T) {
	db, repo := setupGoalTest(t)

	testUser := testutil.CreateTestUser(t, db)

	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)
	goal := &goal.NutritionGoal{
		UserID:    testUser.ID,
		Calories:  2000,
		Protein:   150,
		Carbs:     200,
		Fat:       65,
		Fiber:     30,
		StartDate: startDate,
		EndDate:   &endDate,
		IsActive:  true,
	}
	db.Create(goal)

	// Test on exact end date
	found, err := repo.GetForDate(testUser.ID, endDate)

	assert.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, goal.ID, found.ID)
}

// TestRepository_GetForDate_BeforeStartDate tests getting goal before start date
func TestRepository_GetForDate_BeforeStartDate(t *testing.T) {
	db, repo := setupGoalTest(t)

	testUser := testutil.CreateTestUser(t, db)

	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)
	goal := &goal.NutritionGoal{
		UserID:    testUser.ID,
		Calories:  2000,
		Protein:   150,
		Carbs:     200,
		Fat:       65,
		Fiber:     30,
		StartDate: startDate,
		EndDate:   &endDate,
		IsActive:  true,
	}
	db.Create(goal)

	// Test before start date
	beforeDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	found, err := repo.GetForDate(testUser.ID, beforeDate)

	assert.Error(t, err)
	assert.Nil(t, found)
	assert.Contains(t, err.Error(), "no goal found for date")
}

// TestRepository_GetForDate_AfterEndDate tests getting goal after end date
func TestRepository_GetForDate_AfterEndDate(t *testing.T) {
	db, repo := setupGoalTest(t)

	testUser := testutil.CreateTestUser(t, db)

	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)
	goal := &goal.NutritionGoal{
		UserID:    testUser.ID,
		Calories:  2000,
		Protein:   150,
		Carbs:     200,
		Fat:       65,
		Fiber:     30,
		StartDate: startDate,
		EndDate:   &endDate,
		IsActive:  true,
	}
	db.Create(goal)

	// Test after end date
	afterDate := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
	found, err := repo.GetForDate(testUser.ID, afterDate)

	assert.Error(t, err)
	assert.Nil(t, found)
	assert.Contains(t, err.Error(), "no goal found for date")
}

// TestRepository_GetForDate_NoEndDate tests getting goal with no end date (ongoing)
func TestRepository_GetForDate_NoEndDate(t *testing.T) {
	db, repo := setupGoalTest(t)

	testUser := testutil.CreateTestUser(t, db)

	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	goal := &goal.NutritionGoal{
		UserID:    testUser.ID,
		Calories:  2000,
		Protein:   150,
		Carbs:     200,
		Fat:       65,
		Fiber:     30,
		StartDate: startDate,
		EndDate:   nil,
		IsActive:  true,
	}
	db.Create(goal)

	// Test far future date (should still match since no end date)
	futureDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	found, err := repo.GetForDate(testUser.ID, futureDate)

	assert.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, goal.ID, found.ID)
}

// TestRepository_GetForDate_NoGoal tests getting goal when user has no goals
func TestRepository_GetForDate_NoGoal(t *testing.T) {
	db, repo := setupGoalTest(t)

	testUser := testutil.CreateTestUser(t, db)

	testDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	found, err := repo.GetForDate(testUser.ID, testDate)

	assert.Error(t, err)
	assert.Nil(t, found)
	assert.Contains(t, err.Error(), "no goal found for date")
}

// TestRepository_ConcurrentCreate tests that concurrent goal creation properly handles deactivation
func TestRepository_ConcurrentCreate(t *testing.T) {
	db, repo := setupGoalTest(t)

	testUser := testutil.CreateTestUser(t, db)

	// Create first goal
	goal1 := &goal.NutritionGoal{
		UserID:    testUser.ID,
		Calories:  2000,
		Protein:   150,
		Carbs:     200,
		Fat:       65,
		Fiber:     30,
		StartDate: time.Now(),
		IsActive:  true,
	}
	err := repo.Create(goal1)
	require.NoError(t, err)

	// Create second goal (simulating concurrent request)
	goal2 := &goal.NutritionGoal{
		UserID:    testUser.ID,
		Calories:  2200,
		Protein:   165,
		Carbs:     220,
		Fat:       73,
		Fiber:     32,
		StartDate: time.Now(),
		IsActive:  true,
	}
	err = repo.Create(goal2)
	require.NoError(t, err)

	// Verify only one active goal exists
	var activeCount int64
	db.Model(&goal.NutritionGoal{}).Where("user_id = ? AND is_active = ?", testUser.ID, true).Count(&activeCount)
	assert.Equal(t, int64(1), activeCount, "Only one goal should be active after concurrent creates")

	// Verify it's the most recent one
	active, err := repo.GetActive(testUser.ID)
	require.NoError(t, err)
	assert.Equal(t, goal2.ID, active.ID, "Most recent goal should be active")
}

// TestRepository_Create_WithProtocolTracking tests creating a goal with protocol tracking fields
func TestRepository_Create_WithProtocolTracking(t *testing.T) {
	db, repo := setupGoalTest(t)

	testUser := testutil.CreateTestUser(t, db)

	dietModel := "zeroToHero"
	protocol := 1
	phase := 2
	expirationDate := time.Now().AddDate(0, 0, 14)

	mygoal := &goal.NutritionGoal{
		UserID:         testUser.ID,
		Calories:       2884,
		Protein:        134,
		Carbs:          362,
		Fat:            74,
		Fiber:          40,
		StartDate:      time.Now(),
		IsActive:       true,
		DietModel:      &dietModel,
		Protocol:       &protocol,
		Phase:          &phase,
		ExpirationDate: &expirationDate,
	}

	err := repo.Create(mygoal)

	assert.NoError(t, err)
	assert.NotZero(t, mygoal.ID, "Goal ID should be set after creation")

	// Verify protocol tracking fields were saved
	var found goal.NutritionGoal
	db.First(&found, mygoal.ID)
	assert.Equal(t, testUser.ID, found.UserID)

	// Verify protocol tracking fields
	require.NotNil(t, found.DietModel, "DietModel should not be nil")
	assert.Equal(t, "zeroToHero", *found.DietModel)

	require.NotNil(t, found.Protocol, "Protocol should not be nil")
	assert.Equal(t, 1, *found.Protocol)

	require.NotNil(t, found.Phase, "Phase should not be nil")
	assert.Equal(t, 2, *found.Phase)

	require.NotNil(t, found.ExpirationDate, "ExpirationDate should not be nil")
	assert.Equal(t, expirationDate.Format("2006-01-02"), found.ExpirationDate.Format("2006-01-02"))
}

// TestRepository_Create_WithoutProtocolTracking tests creating a manual goal without protocol tracking
func TestRepository_Create_WithoutProtocolTracking(t *testing.T) {
	db, repo := setupGoalTest(t)

	testUser := testutil.CreateTestUser(t, db)

	mygoal := &goal.NutritionGoal{
		UserID:    testUser.ID,
		Calories:  2200,
		Protein:   165,
		Carbs:     220,
		Fat:       73,
		Fiber:     31,
		StartDate: time.Now(),
		IsActive:  true,
		// No protocol tracking fields
	}

	err := repo.Create(mygoal)

	assert.NoError(t, err)
	assert.NotZero(t, mygoal.ID, "Goal ID should be set after creation")

	// Verify protocol tracking fields are nil
	var found goal.NutritionGoal
	db.First(&found, mygoal.ID)
	assert.Nil(t, found.DietModel, "DietModel should be nil for manual goal")
	assert.Nil(t, found.Protocol, "Protocol should be nil for manual goal")
	assert.Nil(t, found.Phase, "Phase should be nil for manual goal")
	assert.Nil(t, found.ExpirationDate, "ExpirationDate should be nil for manual goal")
}

// TestRepository_Create_WithProtocolNoPhase tests creating a goal with protocol but no phase
func TestRepository_Create_WithProtocolNoPhase(t *testing.T) {
	db, repo := setupGoalTest(t)

	testUser := testutil.CreateTestUser(t, db)

	dietModel := "zeroToHero"
	protocol := 2
	expirationDate := time.Now().AddDate(0, 0, 14)

	mygoal := &goal.NutritionGoal{
		UserID:         testUser.ID,
		Calories:       2384,
		Protein:        134,
		Carbs:          262,
		Fat:            74,
		Fiber:          33,
		StartDate:      time.Now(),
		IsActive:       true,
		DietModel:      &dietModel,
		Protocol:       &protocol,
		Phase:          nil, // No phase specified
		ExpirationDate: &expirationDate,
	}

	err := repo.Create(mygoal)

	assert.NoError(t, err)

	// Verify protocol tracking without phase
	var found goal.NutritionGoal
	db.First(&found, mygoal.ID)

	require.NotNil(t, found.DietModel)
	assert.Equal(t, "zeroToHero", *found.DietModel)

	require.NotNil(t, found.Protocol)
	assert.Equal(t, 2, *found.Protocol)

	assert.Nil(t, found.Phase, "Phase should be nil when not specified")

	require.NotNil(t, found.ExpirationDate)
}

// TestRepository_GetActive_WithProtocolTracking tests retrieving active goal with protocol tracking
func TestRepository_GetActive_WithProtocolTracking(t *testing.T) {
	db, repo := setupGoalTest(t)

	testUser := testutil.CreateTestUser(t, db)

	dietModel := "zeroToHero"
	protocol := 3
	phase := 1
	expirationDate := time.Now().AddDate(0, 0, 14)

	testGoal := &goal.NutritionGoal{
		UserID:         testUser.ID,
		Calories:       2384,
		Protein:        134,
		Carbs:          262,
		Fat:            74,
		Fiber:          33,
		StartDate:      time.Now(),
		IsActive:       true,
		DietModel:      &dietModel,
		Protocol:       &protocol,
		Phase:          &phase,
		ExpirationDate: &expirationDate,
	}
	db.Create(testGoal)

	found, err := repo.GetActive(testUser.ID)

	assert.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, testGoal.ID, found.ID)

	// Verify protocol tracking fields are returned
	require.NotNil(t, found.DietModel)
	assert.Equal(t, "zeroToHero", *found.DietModel)

	require.NotNil(t, found.Protocol)
	assert.Equal(t, 3, *found.Protocol)

	require.NotNil(t, found.Phase)
	assert.Equal(t, 1, *found.Phase)

	require.NotNil(t, found.ExpirationDate)
	assert.Equal(t, expirationDate.Format("2006-01-02"), found.ExpirationDate.Format("2006-01-02"))
}

// TestRepository_GetAll_MixedProtocolTracking tests retrieving goals with and without protocol tracking
func TestRepository_GetAll_MixedProtocolTracking(t *testing.T) {
	db, repo := setupGoalTest(t)

	testUser := testutil.CreateTestUser(t, db)

	// Create manual goal without protocol tracking
	manualGoal := &goal.NutritionGoal{
		UserID:    testUser.ID,
		Calories:  2000,
		Protein:   150,
		Carbs:     200,
		Fat:       65,
		Fiber:     30,
		StartDate: time.Now().AddDate(0, 0, -7),
		IsActive:  false,
	}
	db.Create(manualGoal)

	// Create calculated goal with protocol tracking
	dietModel := "zeroToHero"
	protocol := 4
	phase := 1
	expirationDate := time.Now().AddDate(0, 0, 14)

	calculatedGoal := &goal.NutritionGoal{
		UserID:         testUser.ID,
		Calories:       2384,
		Protein:        134,
		Carbs:          262,
		Fat:            74,
		Fiber:          33,
		StartDate:      time.Now(),
		IsActive:       true,
		DietModel:      &dietModel,
		Protocol:       &protocol,
		Phase:          &phase,
		ExpirationDate: &expirationDate,
	}
	db.Create(calculatedGoal)

	found, err := repo.GetAll(testUser.ID)

	assert.NoError(t, err)
	assert.Len(t, found, 2, "Should return both goals")

	// Most recent (calculated) should be first
	assert.Equal(t, calculatedGoal.ID, found[0].ID)
	require.NotNil(t, found[0].DietModel)
	assert.Equal(t, "zeroToHero", *found[0].DietModel)
	require.NotNil(t, found[0].Protocol)
	assert.Equal(t, 4, *found[0].Protocol)
	require.NotNil(t, found[0].Phase)
	assert.Equal(t, 1, *found[0].Phase)

	// Manual goal should be second with nil protocol fields
	assert.Equal(t, manualGoal.ID, found[1].ID)
	assert.Nil(t, found[1].DietModel)
	assert.Nil(t, found[1].Protocol)
	assert.Nil(t, found[1].Phase)
	assert.Nil(t, found[1].ExpirationDate)
}

// TestRepository_GetByID_WithProtocolTracking tests retrieving a specific goal with protocol tracking
func TestRepository_GetByID_WithProtocolTracking(t *testing.T) {
	db, repo := setupGoalTest(t)

	testUser := testutil.CreateTestUser(t, db)

	dietModel := "zeroToHero"
	protocol := 1
	phase := 3
	expirationDate := time.Now().AddDate(0, 0, 14)

	testGoal := &goal.NutritionGoal{
		UserID:         testUser.ID,
		Calories:       3084,
		Protein:        134,
		Carbs:          412,
		Fat:            74,
		Fiber:          43,
		StartDate:      time.Now(),
		IsActive:       true,
		DietModel:      &dietModel,
		Protocol:       &protocol,
		Phase:          &phase,
		ExpirationDate: &expirationDate,
	}
	db.Create(testGoal)

	found, err := repo.GetByID(testGoal.ID, testUser.ID)

	assert.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, testGoal.ID, found.ID)

	// Verify all protocol tracking fields
	require.NotNil(t, found.DietModel)
	assert.Equal(t, "zeroToHero", *found.DietModel)

	require.NotNil(t, found.Protocol)
	assert.Equal(t, 1, *found.Protocol)

	require.NotNil(t, found.Phase)
	assert.Equal(t, 3, *found.Phase)

	require.NotNil(t, found.ExpirationDate)
}

// TestRepository_Update_ProtocolTracking tests that protocol tracking persists after update
func TestRepository_Update_ProtocolTracking(t *testing.T) {
	db, repo := setupGoalTest(t)

	testUser := testutil.CreateTestUser(t, db)

	dietModel := "zeroToHero"
	protocol := 2
	phase := 1
	expirationDate := time.Now().AddDate(0, 0, 14)

	testGoal := &goal.NutritionGoal{
		UserID:         testUser.ID,
		Calories:       2384,
		Protein:        134,
		Carbs:          262,
		Fat:            74,
		Fiber:          33,
		StartDate:      time.Now(),
		IsActive:       true,
		DietModel:      &dietModel,
		Protocol:       &protocol,
		Phase:          &phase,
		ExpirationDate: &expirationDate,
	}
	db.Create(testGoal)

	// Update nutrition values but not protocol tracking
	testGoal.Calories = 2500
	testGoal.Carbs = 280

	err := repo.Update(testGoal)

	assert.NoError(t, err)

	// Verify protocol tracking fields remain unchanged
	var updated goal.NutritionGoal
	db.First(&updated, testGoal.ID)

	testutil.AssertNutritionEquals(t, 2500, updated.Calories, "calories")
	testutil.AssertNutritionEquals(t, 280, updated.Carbs, "carbs")

	// Protocol tracking should still be there
	require.NotNil(t, updated.DietModel)
	assert.Equal(t, "zeroToHero", *updated.DietModel)

	require.NotNil(t, updated.Protocol)
	assert.Equal(t, 2, *updated.Protocol)

	require.NotNil(t, updated.Phase)
	assert.Equal(t, 1, *updated.Phase)

	require.NotNil(t, updated.ExpirationDate)
}

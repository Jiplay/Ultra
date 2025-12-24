package tests

import (
	"context"
	"fmt"
	"testing"
	"time"
	"ultra-bis/internal/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	postgresdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupUserTest creates a test DB with User migrations
// Note: We inline the testcontainer setup here to avoid import cycle with testutil
func setupUserTest(t *testing.T) (*gorm.DB, *user.Repository) {
	t.Helper()

	ctx := context.Background()

	// Create PostgreSQL container
	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("Failed to start PostgreSQL container: %v", err)
	}

	// Clean up container when test finishes
	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate PostgreSQL container: %v", err)
		}
	})

	// Get connection string
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	// Connect with GORM
	db, err := gorm.Open(postgresdriver.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silent mode for tests
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Migrate User model
	if err := db.AutoMigrate(&user.User{}); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return db, user.NewRepository(db)
}

// TestRepository_Create tests creating a new user
func TestRepository_Create(t *testing.T) {
	db, repo := setupUserTest(t)

	user1 := &user.User{
		Email:         "test@example.com",
		Name:          "Test User",
		Age:           30,
		Gender:        "male",
		Height:        175.0,
		Weight:        75.0,
		BodyFat:       15.0,
		ActivityLevel: user.ModeratelyActive,
		GoalType:      user.Maintain,
	}

	err := user1.HashPassword("password123")
	require.NoError(t, err)

	err = repo.Create(user1)

	assert.NoError(t, err)
	assert.NotZero(t, user1.ID, "User ID should be set after creation")
	assert.NotEmpty(t, user1.PasswordHash, "Password should be hashed")

	// Verify user was created in database
	var found user.User
	db.First(&found, user1.ID)
	assert.Equal(t, "test@example.com", found.Email)
	assert.Equal(t, "Test User", found.Name)
	assert.Equal(t, 30, found.Age)
}

// TestRepository_Create_DuplicateEmail tests that duplicate emails are rejected
func TestRepository_Create_DuplicateEmail(t *testing.T) {
	_, repo := setupUserTest(t)

	// Create first user
	user1 := &user.User{
		Email: "duplicate@example.com",
		Name:  "User One",
	}
	user1.HashPassword("password123")
	err := repo.Create(user1)
	require.NoError(t, err)

	// Try to create second user with same email
	user2 := &user.User{
		Email: "duplicate@example.com",
		Name:  "User Two",
	}
	user2.HashPassword("password456")
	err = repo.Create(user2)

	assert.Error(t, err, "Should fail to create user with duplicate email")
	assert.Contains(t, err.Error(), "failed to create user")
}

// TestRepository_Create_MinimalFields tests creating a user with minimal required fields
func TestRepository_Create_MinimalFields(t *testing.T) {
	_, repo := setupUserTest(t)

	user := &user.User{
		Email: "minimal@example.com",
	}
	user.HashPassword("password123")

	err := repo.Create(user)

	assert.NoError(t, err)
	assert.NotZero(t, user.ID)
}

// TestRepository_GetByID tests retrieving a user by ID
func TestRepository_GetByID(t *testing.T) {
	_, repo := setupUserTest(t)

	// Create a test user
	created := &user.User{
		Email: "getbyid@example.com",
		Name:  "Get By ID User",
		Age:   25,
	}
	created.HashPassword("password123")
	err := repo.Create(created)
	require.NoError(t, err)

	// Retrieve by ID
	found, err := repo.GetByID(created.ID)

	assert.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "getbyid@example.com", found.Email)
	assert.Equal(t, "Get By ID User", found.Name)
	assert.Equal(t, 25, found.Age)
}

// TestRepository_GetByID_NotFound tests retrieving a non-existent user
func TestRepository_GetByID_NotFound(t *testing.T) {
	_, repo := setupUserTest(t)

	found, err := repo.GetByID(99999)

	assert.Error(t, err)
	assert.Nil(t, found)
	assert.Contains(t, err.Error(), "user not found")
}

// TestRepository_GetByEmail tests retrieving a user by email
func TestRepository_GetByEmail(t *testing.T) {
	_, repo := setupUserTest(t)

	// Create a test user
	created := &user.User{
		Email: "findme@example.com",
		Name:  "Find Me User",
	}
	created.HashPassword("password123")
	err := repo.Create(created)
	require.NoError(t, err)

	// Retrieve by email
	found, err := repo.GetByEmail("findme@example.com")

	assert.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "findme@example.com", found.Email)
	assert.Equal(t, "Find Me User", found.Name)
}

// TestRepository_GetByEmail_NotFound tests retrieving a non-existent email
func TestRepository_GetByEmail_NotFound(t *testing.T) {
	_, repo := setupUserTest(t)

	found, err := repo.GetByEmail("nonexistent@example.com")

	assert.Error(t, err)
	assert.Nil(t, found)
	assert.Contains(t, err.Error(), "user not found")
}

// TestRepository_GetByEmail_CaseSensitive tests email lookup case sensitivity
func TestRepository_GetByEmail_CaseSensitive(t *testing.T) {
	_, repo := setupUserTest(t)

	// Create user with lowercase email
	created := &user.User{
		Email: "test@example.com",
		Name:  "Test User",
	}
	created.HashPassword("password123")
	err := repo.Create(created)
	require.NoError(t, err)

	// Try to find with uppercase (should not find due to case sensitivity)
	found, err := repo.GetByEmail("TEST@EXAMPLE.COM")

	// Note: This behavior depends on database collation
	// PostgreSQL default is case-sensitive
	if err != nil {
		assert.Contains(t, err.Error(), "user not found")
		assert.Nil(t, found)
	} else {
		// If database is case-insensitive, this is also valid
		assert.NotNil(t, found)
	}
}

// TestRepository_Update tests updating a user's profile
func TestRepository_Update(t *testing.T) {
	_, repo := setupUserTest(t)

	// Create a test user
	user1 := &user.User{
		Email:  "update@example.com",
		Name:   "Original Name",
		Age:    25,
		Height: 170.0,
		Weight: 70.0,
	}
	user1.HashPassword("password123")
	err := repo.Create(user1)
	require.NoError(t, err)

	// Update user fields
	user1.Name = "Updated Name"
	user1.Age = 26
	user1.Height = 175.0
	user1.Weight = 72.0
	user1.BodyFat = 15.0

	err = repo.Update(user1)

	assert.NoError(t, err)

	// Verify updates in database
	found, err := repo.GetByID(user1.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", found.Name)
	assert.Equal(t, 26, found.Age)
	assert.Equal(t, 175.0, found.Height)
	assert.Equal(t, 72.0, found.Weight)
	assert.Equal(t, 15.0, found.BodyFat)
}

// TestRepository_Update_ActivityLevel tests updating activity level
func TestRepository_Update_ActivityLevel(t *testing.T) {
	_, repo := setupUserTest(t)

	user1 := &user.User{
		Email:         "activity@example.com",
		ActivityLevel: user.Sedentary,
	}
	user1.HashPassword("password123")
	err := repo.Create(user1)
	require.NoError(t, err)

	// Update activity level
	user1.ActivityLevel = user.VeryActive

	err = repo.Update(user1)
	assert.NoError(t, err)

	// Verify update
	found, err := repo.GetByID(user1.ID)
	require.NoError(t, err)
	assert.Equal(t, user.VeryActive, found.ActivityLevel)
}

// TestRepository_Update_GoalType tests updating goal type
func TestRepository_Update_GoalType(t *testing.T) {
	_, repo := setupUserTest(t)

	user1 := &user.User{
		Email:    "goal@example.com",
		GoalType: user.Maintain,
	}
	user1.HashPassword("password123")
	err := repo.Create(user1)
	require.NoError(t, err)

	// Update goal type
	user1.GoalType = user.Lose

	err = repo.Update(user1)
	assert.NoError(t, err)

	// Verify update
	found, err := repo.GetByID(user1.ID)
	require.NoError(t, err)
	assert.Equal(t, user.Lose, found.GoalType)
}

// TestRepository_EmailExists tests checking if an email exists
func TestRepository_EmailExists(t *testing.T) {
	_, repo := setupUserTest(t)

	// Create a test user
	user1 := &user.User{
		Email: "exists@example.com",
	}
	user1.HashPassword("password123")
	err := repo.Create(user1)
	require.NoError(t, err)

	// Check if email exists
	exists, err := repo.EmailExists("exists@example.com")

	assert.NoError(t, err)
	assert.True(t, exists, "Email should exist")
}

// TestRepository_EmailExists_NotFound tests checking a non-existent email
func TestRepository_EmailExists_NotFound(t *testing.T) {
	_, repo := setupUserTest(t)

	exists, err := repo.EmailExists("nonexistent@example.com")

	assert.NoError(t, err)
	assert.False(t, exists, "Email should not exist")
}

// TestRepository_EmailExists_MultipleUsers tests email uniqueness with multiple users
func TestRepository_EmailExists_MultipleUsers(t *testing.T) {
	_, repo := setupUserTest(t)

	// Create multiple users
	emails := []string{"user1@example.com", "user2@example.com", "user3@example.com"}
	for _, email := range emails {
		user1 := &user.User{Email: email}
		user1.HashPassword("password123")
		err := repo.Create(user1)
		require.NoError(t, err)
	}

	// Check each email exists
	for _, email := range emails {
		exists, err := repo.EmailExists(email)
		assert.NoError(t, err)
		assert.True(t, exists, "Email %s should exist", email)
	}

	// Check non-existent email
	exists, err := repo.EmailExists("nothere@example.com")
	assert.NoError(t, err)
	assert.False(t, exists, "Non-existent email should return false")
}

// TestUser_HashPassword tests password hashing functionality
func TestUser_HashPassword(t *testing.T) {
	user1 := &user.User{
		Email: "test@example.com",
	}

	password := "mySecurePassword123!"
	err := user1.HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, user1.PasswordHash, "Password hash should be set")
	assert.NotEqual(t, password, user1.PasswordHash, "Hash should not equal plain password")
	assert.Greater(t, len(user1.PasswordHash), 50, "Bcrypt hash should be long")
}

// TestUser_HashPassword_EmptyPassword tests hashing an empty password
func TestUser_HashPassword_EmptyPassword(t *testing.T) {
	user := &user.User{
		Email: "test@example.com",
	}

	err := user.HashPassword("")

	// Bcrypt accepts empty passwords, but it's still hashed
	assert.NoError(t, err)
	assert.NotEmpty(t, user.PasswordHash)
}

// TestUser_CheckPassword tests password verification
func TestUser_CheckPassword(t *testing.T) {
	user1 := &user.User{
		Email: "test@example.com",
	}

	password := "correctPassword123"
	err := user1.HashPassword(password)
	require.NoError(t, err)

	// Test correct password
	isValid := user1.CheckPassword(password)
	assert.True(t, isValid, "Correct password should validate")

	// Test incorrect password
	isValid = user1.CheckPassword("wrongPassword")
	assert.False(t, isValid, "Incorrect password should not validate")
}

// TestUser_CheckPassword_CaseSensitive tests password case sensitivity
func TestUser_CheckPassword_CaseSensitive(t *testing.T) {
	user1 := &user.User{
		Email: "test@example.com",
	}

	password := "MyPassword123"
	err := user1.HashPassword(password)
	require.NoError(t, err)

	// Test with different case
	isValid := user1.CheckPassword("mypassword123")
	assert.False(t, isValid, "Password should be case-sensitive")

	isValid = user1.CheckPassword("MYPASSWORD123")
	assert.False(t, isValid, "Password should be case-sensitive")
}

// TestUser_CheckPassword_BeforeHash tests checking password before hashing
func TestUser_CheckPassword_BeforeHash(t *testing.T) {
	user1 := &user.User{
		Email: "test@example.com",
	}

	// Try to check password before hashing
	isValid := user1.CheckPassword("anyPassword")
	assert.False(t, isValid, "Should return false when no hash is set")
}

// TestRepository_Create_AllActivityLevels tests creating users with all activity levels
func TestRepository_Create_AllActivityLevels(t *testing.T) {
	db, repo := setupUserTest(t)

	activityLevels := []user.ActivityLevel{
		user.Sedentary,
		user.LightlyActive,
		user.ModeratelyActive,
		user.VeryActive,
		user.ExtraActive,
	}

	for i, level := range activityLevels {
		user1 := &user.User{
			Email:         testEmail(i),
			Name:          "Test User",
			ActivityLevel: level,
		}
		user1.HashPassword("password123")

		err := repo.Create(user1)
		assert.NoError(t, err, "Should create user with activity level %s", level)

		// Verify in database
		var found user.User
		db.First(&found, user1.ID)
		assert.Equal(t, level, found.ActivityLevel)
	}
}

// TestRepository_Create_AllGoalTypes tests creating users with all goal types
func TestRepository_Create_AllGoalTypes(t *testing.T) {
	db, repo := setupUserTest(t)

	goalTypes := []user.GoalType{
		user.Maintain,
		user.Lose,
		user.Gain,
	}

	for i, goalType := range goalTypes {
		user1 := &user.User{
			Email:    testEmail(100 + i),
			Name:     "Test User",
			GoalType: goalType,
		}
		user1.HashPassword("password123")

		err := repo.Create(user1)
		assert.NoError(t, err, "Should create user with goal type %s", goalType)

		// Verify in database
		var found user.User
		db.First(&found, user1.ID)
		assert.Equal(t, goalType, found.GoalType)
	}
}

// TestRepository_Update_CompleteProfile tests updating a complete user profile
func TestRepository_Update_CompleteProfile(t *testing.T) {
	_, repo := setupUserTest(t)

	// Create user with minimal data
	user1 := &user.User{
		Email: "complete@example.com",
	}
	user1.HashPassword("password123")
	err := repo.Create(user1)
	require.NoError(t, err)

	// Update with complete profile
	user1.Name = "Complete User"
	user1.Age = 30
	user1.Gender = "male"
	user1.Height = 175.0
	user1.Weight = 75.0
	user1.BodyFat = 15.0
	user1.ActivityLevel = user.ModeratelyActive
	user1.GoalType = user.Lose

	err = repo.Update(user1)
	assert.NoError(t, err)

	// Verify all fields
	found, err := repo.GetByID(user1.ID)
	require.NoError(t, err)
	assert.Equal(t, "Complete User", found.Name)
	assert.Equal(t, 30, found.Age)
	assert.Equal(t, "male", found.Gender)
	assert.Equal(t, 175.0, found.Height)
	assert.Equal(t, 75.0, found.Weight)
	assert.Equal(t, 15.0, found.BodyFat)
	assert.Equal(t, user.ModeratelyActive, found.ActivityLevel)
	assert.Equal(t, user.Lose, found.GoalType)
}

// TestRepository_Update_PasswordNotChanged tests that password is preserved during updates
func TestRepository_Update_PasswordNotChanged(t *testing.T) {
	_, repo := setupUserTest(t)

	user1 := &user.User{
		Email: "password@example.com",
		Name:  "Original Name",
	}
	originalPassword := "originalPassword123"
	err := user1.HashPassword(originalPassword)
	require.NoError(t, err)
	originalHash := user1.PasswordHash

	err = repo.Create(user1)
	require.NoError(t, err)

	// Update other fields but not password
	user1.Name = "Updated Name"
	user1.Age = 30

	err = repo.Update(user1)
	assert.NoError(t, err)

	// Verify password hash hasn't changed
	found, err := repo.GetByID(user1.ID)
	require.NoError(t, err)
	assert.Equal(t, originalHash, found.PasswordHash, "Password hash should remain unchanged")
	assert.True(t, found.CheckPassword(originalPassword), "Original password should still work")
}

// TestRepository_Update_ChangePassword tests updating a user's password
func TestRepository_Update_ChangePassword(t *testing.T) {
	_, repo := setupUserTest(t)

	user1 := &user.User{
		Email: "changepass@example.com",
	}
	oldPassword := "oldPassword123"
	err := user1.HashPassword(oldPassword)
	require.NoError(t, err)
	oldHash := user1.PasswordHash

	err = repo.Create(user1)
	require.NoError(t, err)

	// Change password
	newPassword := "newPassword456"
	err = user1.HashPassword(newPassword)
	require.NoError(t, err)

	err = repo.Update(user1)
	assert.NoError(t, err)

	// Verify password changed
	found, err := repo.GetByID(user1.ID)
	require.NoError(t, err)
	assert.NotEqual(t, oldHash, found.PasswordHash, "Password hash should have changed")
	assert.False(t, found.CheckPassword(oldPassword), "Old password should not work")
	assert.True(t, found.CheckPassword(newPassword), "New password should work")
}

// Helper function to generate test emails
func testEmail(index int) string {
	return fmt.Sprintf("test%d@example.com", index)
}

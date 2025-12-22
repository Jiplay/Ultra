//go:build !usertest

package testutil

import (
	"testing"

	"gorm.io/gorm"
	"ultra-bis/internal/user"
)

// CreateTestUser creates a test user with default values
// Override fields by passing optional modifications after creation
// Note: This helper is in a separate file to avoid import cycles with user package tests
func CreateTestUser(t *testing.T, db *gorm.DB, email ...string) *user.User {
	t.Helper()

	userEmail := "test@example.com"
	if len(email) > 0 {
		userEmail = email[0]
	}

	testUser := &user.User{
		Email:         userEmail,
		Name:          "Test User",
		Age:           30,
		Gender:        "male",
		Height:        175.0,
		Weight:        75.0,
		BodyFat:       15.0,
		ActivityLevel: user.ModeratelyActive,
		GoalType:      user.Maintain,
	}

	if err := testUser.HashPassword("password123"); err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if err := db.Create(testUser).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return testUser
}

// CreateTestUserWithProfile creates a test user with specific profile data
func CreateTestUserWithProfile(t *testing.T, db *gorm.DB, age int, gender string, height, weight, bodyFat float64) *user.User {
	t.Helper()

	testUser := &user.User{
		Email:         "profile@example.com",
		Name:          "Profile User",
		Age:           age,
		Gender:        gender,
		Height:        height,
		Weight:        weight,
		BodyFat:       bodyFat,
		ActivityLevel: user.ModeratelyActive,
		GoalType:      user.Maintain,
	}

	if err := testUser.HashPassword("password123"); err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if err := db.Create(testUser).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return testUser
}

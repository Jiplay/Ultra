package user

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ActivityLevel represents the user's physical activity level
type ActivityLevel string

const (
	Sedentary    ActivityLevel = "sedentary"     // Little or no exercise
	LightlyActive ActivityLevel = "light"        // Light exercise 1-3 days/week
	ModeratelyActive ActivityLevel = "moderate"  // Moderate exercise 3-5 days/week
	VeryActive   ActivityLevel = "active"        // Hard exercise 6-7 days/week
	ExtraActive  ActivityLevel = "very_active"   // Very hard exercise, physical job
)

// GoalType represents the user's fitness goal
type GoalType string

const (
	Maintain GoalType = "maintain"  // Maintain current weight
	Lose     GoalType = "lose"      // Lose weight
	Gain     GoalType = "gain"      // Gain weight/muscle
)

// User represents a user in the system
type User struct {
	ID            uint          `json:"id" gorm:"primarykey"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	Email         string        `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash  string        `json:"-" gorm:"not null"`
	Name          string        `json:"name" gorm:"type:varchar(255)"`
	Age           int           `json:"age"`
	Gender        string        `json:"gender" gorm:"type:varchar(10)"`
	Height        float64       `json:"height" gorm:"type:decimal(5,2)"` // in cm
	Weight        float64       `json:"weight" gorm:"type:decimal(6,2)"` // in kg
	BodyFat       float64       `json:"body_fat" gorm:"type:decimal(5,2)"` // body fat percentage
	ActivityLevel ActivityLevel `json:"activity_level" gorm:"type:varchar(20);default:'moderate'"`
	GoalType      GoalType      `json:"goal_type" gorm:"type:varchar(20);default:'maintain'"`
}

// RegisterRequest represents the registration request
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// UpdateProfileRequest represents the profile update request
type UpdateProfileRequest struct {
	Name          string        `json:"name"`
	Age           int           `json:"age"`
	Gender        string        `json:"gender"`
	Height        float64       `json:"height"`
	Weight        float64       `json:"weight"`
	BodyFat       float64       `json:"body_fat"`
	ActivityLevel ActivityLevel `json:"activity_level"`
	GoalType      GoalType      `json:"goal_type"`
}

// HashPassword hashes the user's password
func (u *User) HashPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashedPassword)
	return nil
}

// CheckPassword checks if the provided password is correct
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

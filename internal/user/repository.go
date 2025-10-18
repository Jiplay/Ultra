package user

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// Repository handles database operations for users
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new user repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new user
func (r *Repository) Create(user *User) error {
	result := r.db.Create(user)
	if result.Error != nil {
		return fmt.Errorf("failed to create user: %w", result.Error)
	}
	return nil
}

// GetByID retrieves a user by ID
func (r *Repository) GetByID(id uint) (*User, error) {
	var user User
	result := r.db.First(&user, id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("user not found")
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get user: %w", result.Error)
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *Repository) GetByEmail(email string) (*User, error) {
	var user User
	result := r.db.Where("email = ?", email).First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("user not found")
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get user: %w", result.Error)
	}

	return &user, nil
}

// Update updates a user's profile
func (r *Repository) Update(user *User) error {
	result := r.db.Save(user)
	if result.Error != nil {
		return fmt.Errorf("failed to update user: %w", result.Error)
	}
	return nil
}

// EmailExists checks if an email already exists
func (r *Repository) EmailExists(email string) (bool, error) {
	var count int64
	result := r.db.Model(&User{}).Where("email = ?", email).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

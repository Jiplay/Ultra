package goal

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Repository handles database operations for nutrition goals
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new goal repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new nutrition goal and deactivates previous ones
func (r *Repository) Create(goal *NutritionGoal) error {
	// Start transaction
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Deactivate all previous active goals for this user
		if err := tx.Model(&NutritionGoal{}).
			Where("user_id = ? AND is_active = ?", goal.UserID, true).
			Update("is_active", false).Error; err != nil {
			return fmt.Errorf("failed to deactivate previous goals: %w", err)
		}

		// Create new goal
		if err := tx.Create(goal).Error; err != nil {
			return fmt.Errorf("failed to create goal: %w", err)
		}

		return nil
	})
}

// GetActive retrieves the active nutrition goal for a user
func (r *Repository) GetActive(userID uint) (*NutritionGoal, error) {
	var goal NutritionGoal
	result := r.db.Where("user_id = ? AND is_active = ?", userID, true).First(&goal)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("no active goal found")
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get active goal: %w", result.Error)
	}

	return &goal, nil
}

// GetByID retrieves a nutrition goal by ID and user ID
func (r *Repository) GetByID(id, userID uint) (*NutritionGoal, error) {
	var goal NutritionGoal
	result := r.db.Where("id = ? AND user_id = ?", id, userID).First(&goal)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("goal not found")
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get goal: %w", result.Error)
	}

	return &goal, nil
}

// GetAll retrieves all nutrition goals for a user
func (r *Repository) GetAll(userID uint) ([]NutritionGoal, error) {
	var goals []NutritionGoal
	result := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&goals)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get goals: %w", result.Error)
	}

	return goals, nil
}

// Update updates a nutrition goal
func (r *Repository) Update(goal *NutritionGoal) error {
	result := r.db.Save(goal)
	if result.Error != nil {
		return fmt.Errorf("failed to update goal: %w", result.Error)
	}
	return nil
}

// Delete soft deletes a nutrition goal
func (r *Repository) Delete(id, userID uint) error {
	result := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&NutritionGoal{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete goal: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("goal not found")
	}

	return nil
}

// GetForDate retrieves the active goal for a specific date
func (r *Repository) GetForDate(userID uint, date time.Time) (*NutritionGoal, error) {
	var goal NutritionGoal
	query := r.db.Where("user_id = ? AND start_date <= ?", userID, date)
	query = query.Where("end_date IS NULL OR end_date >= ?", date)
	result := query.First(&goal)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("no goal found for date")
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get goal: %w", result.Error)
	}

	return &goal, nil
}

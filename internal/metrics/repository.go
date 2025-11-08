package metrics

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Repository handles database operations for body metrics
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new metrics repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new body metric entry
func (r *Repository) Create(metric *BodyMetric) error {
	result := r.db.Create(metric)
	if result.Error != nil {
		return fmt.Errorf("failed to create metric: %w", result.Error)
	}
	return nil
}

// GetByDate retrieves a metric for a specific user and date
func (r *Repository) GetByDate(userID uint, date time.Time) (*BodyMetric, error) {
	var metric BodyMetric
	// Truncate to date only for comparison
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	result := r.db.Where("user_id = ? AND date >= ? AND date < ?", userID, startOfDay, endOfDay).First(&metric)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil // Return nil without error if not found
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get metric by date: %w", result.Error)
	}

	return &metric, nil
}

// Update updates an existing body metric entry
func (r *Repository) Update(metric *BodyMetric) error {
	result := r.db.Save(metric)
	if result.Error != nil {
		return fmt.Errorf("failed to update metric: %w", result.Error)
	}
	return nil
}

// GetLatest retrieves the most recent metric for a user
func (r *Repository) GetLatest(userID uint) (*BodyMetric, error) {
	var metric BodyMetric
	result := r.db.Where("user_id = ?", userID).Order("date DESC").First(&metric)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("no metrics found")
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get latest metric: %w", result.Error)
	}

	return &metric, nil
}

// GetByDateRange retrieves metrics within a date range
func (r *Repository) GetByDateRange(userID uint, startDate, endDate time.Time) ([]BodyMetric, error) {
	var metrics []BodyMetric

	result := r.db.Where("user_id = ? AND date >= ? AND date <= ?", userID, startDate, endDate).
		Order("date ASC").
		Find(&metrics)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", result.Error)
	}

	return metrics, nil
}

// GetAll retrieves all metrics for a user
func (r *Repository) GetAll(userID uint) ([]BodyMetric, error) {
	var metrics []BodyMetric

	result := r.db.Where("user_id = ?", userID).Order("date DESC").Find(&metrics)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", result.Error)
	}

	return metrics, nil
}

// GetWeekly retrieves metrics for the current week (starting from previous Monday)
func (r *Repository) GetWeekly(userID uint) ([]BodyMetric, error) {
	now := time.Now()

	// Calculate the start of the week (Monday)
	weekday := now.Weekday()
	daysToSubtract := int(weekday - time.Monday)
	if weekday == time.Sunday {
		// Sunday is 0, so we need to go back 6 days to get to Monday
		daysToSubtract = 6
	}

	startOfWeek := now.AddDate(0, 0, -daysToSubtract)
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())

	// End of week is Sunday at 23:59:59
	endOfWeek := startOfWeek.AddDate(0, 0, 7)

	var metrics []BodyMetric
	result := r.db.Where("user_id = ? AND date >= ? AND date < ?", userID, startOfWeek, endOfWeek).
		Order("date ASC").
		Find(&metrics)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get weekly metrics: %w", result.Error)
	}

	return metrics, nil
}

// Delete deletes a metric entry
func (r *Repository) Delete(id, userID uint) error {
	result := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&BodyMetric{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete metric: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("metric not found")
	}

	return nil
}

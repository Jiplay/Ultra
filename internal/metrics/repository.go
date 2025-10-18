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

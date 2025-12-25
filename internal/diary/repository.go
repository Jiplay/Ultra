package diary

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Repository handles database operations for diary entries
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new diary repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new diary entry
func (r *Repository) Create(entry *DiaryEntry) error {
	result := r.db.Create(entry)
	if result.Error != nil {
		return fmt.Errorf("failed to create diary entry: %w", result.Error)
	}
	return nil
}

// GetByID retrieves a diary entry by ID and user ID
func (r *Repository) GetByID(id, userID uint) (*DiaryEntry, error) {
	var entry DiaryEntry
	result := r.db.Where("id = ? AND user_id = ?", id, userID).First(&entry)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("diary entry not found")
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get diary entry: %w", result.Error)
	}

	return &entry, nil
}

// GetByDate retrieves all diary entries for a user on a specific date
func (r *Repository) GetByDate(userID uint, date time.Time) ([]DiaryEntry, error) {
	var entries []DiaryEntry

	// Get start and end of day
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	result := r.db.Where("user_id = ? AND date >= ? AND date < ?", userID, startOfDay, endOfDay).
		Order("meal_type, created_at").
		Find(&entries)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get diary entries: %w", result.Error)
	}

	// Populate FoodName and RecipeName
	r.populateNames(&entries)

	return entries, nil
}

// GetByDateRange retrieves all diary entries for a user within a date range
func (r *Repository) GetByDateRange(userID uint, startDate, endDate time.Time) ([]DiaryEntry, error) {
	var entries []DiaryEntry

	result := r.db.Where("user_id = ? AND date >= ? AND date < ?", userID, startDate, endDate).
		Order("date, meal_type, created_at").
		Find(&entries)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get diary entries: %w", result.Error)
	}

	// Populate FoodName and RecipeName
	r.populateNames(&entries)

	return entries, nil
}

// Update updates a diary entry
func (r *Repository) Update(entry *DiaryEntry) error {
	result := r.db.Save(entry)
	if result.Error != nil {
		return fmt.Errorf("failed to update diary entry: %w", result.Error)
	}
	return nil
}

// Delete soft deletes a diary entry
func (r *Repository) Delete(id, userID uint) error {
	result := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&DiaryEntry{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete diary entry: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("diary entry not found")
	}

	return nil
}

// GetRecentFoods gets recently logged unique food IDs for a user
func (r *Repository) GetRecentFoods(userID uint, limit int) ([]uint, error) {
	var foodIDs []uint

	result := r.db.Model(&DiaryEntry{}).
		Select("DISTINCT food_id").
		Where("user_id = ? AND food_id IS NOT NULL", userID).
		Order("created_at DESC").
		Limit(limit).
		Pluck("food_id", &foodIDs)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get recent foods: %w", result.Error)
	}

	return foodIDs, nil
}

// GetDailySummary calculates nutrition totals for a specific date
func (r *Repository) GetDailySummary(userID uint, date time.Time) (map[string]float64, error) {
	var result struct {
		TotalCalories float64
		TotalProtein  float64
		TotalCarbs    float64
		TotalFat      float64
		TotalFiber    float64
	}

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	err := r.db.Model(&DiaryEntry{}).
		Select("SUM(calories) as total_calories, SUM(protein) as total_protein, SUM(carbs) as total_carbs, SUM(fat) as total_fat, SUM(fiber) as total_fiber").
		Where("user_id = ? AND date >= ? AND date < ?", userID, startOfDay, endOfDay).
		Scan(&result).Error

	if err != nil {
		return nil, fmt.Errorf("failed to calculate daily summary: %w", err)
	}

	return map[string]float64{
		"calories": result.TotalCalories,
		"protein":  result.TotalProtein,
		"carbs":    result.TotalCarbs,
		"fat":      result.TotalFat,
		"fiber":    result.TotalFiber,
	}, nil
}

// populateNames populates food_name and recipe_name for diary entries
func (r *Repository) populateNames(entries *[]DiaryEntry) {
	for i := range *entries {
		entry := &(*entries)[i]

		// Populate food name (inline takes precedence over saved foods)
		if entry.InlineFoodName != nil && *entry.InlineFoodName != "" {
			entry.FoodName = *entry.InlineFoodName
		} else if entry.FoodID != nil {
			var foodName string
			err := r.db.Table("foods").Select("name").Where("id = ?", *entry.FoodID).Scan(&foodName).Error
			if err == nil {
				entry.FoodName = foodName
			}
		}

		// Populate recipe name (inline takes precedence over saved recipes)
		if entry.InlineRecipeName != nil && *entry.InlineRecipeName != "" {
			entry.RecipeName = *entry.InlineRecipeName
		} else if entry.RecipeID != nil {
			var recipeName string
			err := r.db.Table("recipes").Select("name").Where("id = ?", *entry.RecipeID).Scan(&recipeName).Error
			if err == nil {
				entry.RecipeName = recipeName
			}
		}
	}
}

package food

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// Repository handles database operations for food items
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new food repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create inserts a new food item into the database
func (r *Repository) Create(req CreateFoodRequest) (*Food, error) {
	food := &Food{
		Name:        req.Name,
		Description: req.Description,
		Country:     req.Country,
		Calories:    req.Calories,
		Protein:     req.Protein,
		Carbs:       req.Carbs,
		Fat:         req.Fat,
		Fiber:       req.Fiber,
	}

	result := r.db.Create(food)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create food: %w", result.Error)
	}

	return food, nil
}

// GetByID retrieves a food item by its ID
func (r *Repository) GetByID(id int) (*Food, error) {
	var food Food
	result := r.db.First(&food, id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("food not found")
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get food: %w", result.Error)
	}

	return &food, nil
}

// GetAll retrieves all food items from the database
func (r *Repository) GetAll() ([]Food, error) {
	var foods []Food
	result := r.db.Order("created_at DESC").Find(&foods)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get foods: %w", result.Error)
	}

	return foods, nil
}

// Update modifies an existing food item
func (r *Repository) Update(id int, req UpdateFoodRequest) (*Food, error) {
	var food Food
	result := r.db.First(&food, id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("food not found")
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find food: %w", result.Error)
	}

	// Update fields
	food.Name = req.Name
	food.Description = req.Description
	food.Country = req.Country
	food.Calories = req.Calories
	food.Protein = req.Protein
	food.Carbs = req.Carbs
	food.Fat = req.Fat
	food.Fiber = req.Fiber

	result = r.db.Save(&food)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to update food: %w", result.Error)
	}

	return &food, nil
}

// Delete removes a food item from the database
func (r *Repository) Delete(id int) error {
	result := r.db.Delete(&Food{}, id)

	if result.Error != nil {
		return fmt.Errorf("failed to delete food: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("food not found")
	}

	return nil
}

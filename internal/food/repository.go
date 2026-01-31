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
	// Default tag to "routine" if not provided
	tag := req.Tag
	if tag == "" {
		tag = TagRoutine
	}

	food := &Food{
		Name:        req.Name,
		Description: req.Description,
		Calories:    req.Calories,
		Protein:     req.Protein,
		Carbs:       req.Carbs,
		Fat:         req.Fat,
		Fiber:       req.Fiber,
		Tag:         tag,
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
	food.Calories = req.Calories
	food.Protein = req.Protein
	food.Carbs = req.Carbs
	food.Fat = req.Fat
	food.Fiber = req.Fiber
	if req.Tag != "" {
		food.Tag = req.Tag
	}

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

// GetByIDs retrieves multiple food items by their IDs in a single query
// This method is optimized for batch fetching to avoid N+1 query problems
func (r *Repository) GetByIDs(ids []int) ([]*Food, error) {
	if len(ids) == 0 {
		return []*Food{}, nil
	}

	var foods []*Food
	result := r.db.Where("id IN ?", ids).Find(&foods)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get foods: %w", result.Error)
	}

	return foods, nil
}

// GetByTag retrieves food items filtered by tag (routine or contextual)
func (r *Repository) GetByTag(tag string) ([]Food, error) {
	var foods []Food
	result := r.db.Where("tag = ?", tag).Order("created_at DESC").Find(&foods)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get foods by tag: %w", result.Error)
	}

	return foods, nil
}

// GeneralFoodRepository interface defines operations for general foods reference data
type GeneralFoodRepository interface {
	Search(query string, page int, pageSize int) ([]GeneralFood, int64, error)
	GetByID(id uint) (*GeneralFood, error)
}

// generalFoodRepository implements GeneralFoodRepository
type generalFoodRepository struct {
	db *gorm.DB
}

// NewGeneralFoodRepository creates a new general food repository
func NewGeneralFoodRepository(db *gorm.DB) GeneralFoodRepository {
	return &generalFoodRepository{db: db}
}

// Search performs name-based search with pagination
func (r *generalFoodRepository) Search(query string, page int, pageSize int) ([]GeneralFood, int64, error) {
	var foods []GeneralFood
	var count int64

	// Build query
	db := r.db.Model(&GeneralFood{})

	// Apply name filter if provided
	if query != "" {
		db = db.Where("LOWER(name) LIKE LOWER(?)", "%"+query+"%")
	}

	// Get total count
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count general foods: %w", err)
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	if err := db.Order("name ASC").
		Limit(pageSize).
		Offset(offset).
		Find(&foods).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to search general foods: %w", err)
	}

	return foods, count, nil
}

// GetByID retrieves a single general food by ID
func (r *generalFoodRepository) GetByID(id uint) (*GeneralFood, error) {
	var food GeneralFood
	if err := r.db.First(&food, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("general food not found")
		}
		return nil, fmt.Errorf("failed to get general food: %w", err)
	}
	return &food, nil
}

package recipe

import (
	"gorm.io/gorm"
)

// Repository handles recipe database operations
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new recipe repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new recipe
func (r *Repository) Create(recipe *Recipe) error {
	return r.db.Create(recipe).Error
}

// GetByID retrieves a recipe by ID with ingredients preloaded
func (r *Repository) GetByID(id int) (*Recipe, error) {
	var recipe Recipe
	err := r.db.Preload("Ingredients").First(&recipe, id).Error
	return &recipe, err
}

// GetAll retrieves all recipes (global and user-specific)
func (r *Repository) GetAll() ([]Recipe, error) {
	var recipes []Recipe
	err := r.db.Preload("Ingredients").Find(&recipes).Error
	return recipes, err
}

// GetByUserID retrieves recipes for a specific user (includes global recipes)
func (r *Repository) GetByUserID(userID uint, userOnly bool) ([]Recipe, error) {
	var recipes []Recipe
	query := r.db.Preload("Ingredients")

	if userOnly {
		// Only user's private recipes
		query = query.Where("user_id = ?", userID)
	} else {
		// User's private recipes + global recipes
		query = query.Where("user_id = ? OR user_id IS NULL", userID)
	}

	err := query.Find(&recipes).Error
	return recipes, err
}

// GetGlobal retrieves all global recipes (user_id is NULL)
func (r *Repository) GetGlobal() ([]Recipe, error) {
	var recipes []Recipe
	err := r.db.Preload("Ingredients").Where("user_id IS NULL").Find(&recipes).Error
	return recipes, err
}

// Update updates a recipe
func (r *Repository) Update(recipe *Recipe) error {
	return r.db.Save(recipe).Error
}

// Delete deletes a recipe (cascade will delete ingredients)
func (r *Repository) Delete(id int) error {
	return r.db.Delete(&Recipe{}, id).Error
}

// AddIngredient adds an ingredient to a recipe
func (r *Repository) AddIngredient(ingredient *RecipeIngredient) error {
	return r.db.Create(ingredient).Error
}

// GetIngredient retrieves a specific ingredient by ID
func (r *Repository) GetIngredient(ingredientID int) (*RecipeIngredient, error) {
	var ingredient RecipeIngredient
	err := r.db.First(&ingredient, ingredientID).Error
	return &ingredient, err
}

// UpdateIngredient updates an ingredient
func (r *Repository) UpdateIngredient(ingredient *RecipeIngredient) error {
	return r.db.Save(ingredient).Error
}

// DeleteIngredient removes an ingredient from a recipe
func (r *Repository) DeleteIngredient(ingredientID int) error {
	return r.db.Delete(&RecipeIngredient{}, ingredientID).Error
}

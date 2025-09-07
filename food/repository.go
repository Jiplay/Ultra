package food

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Repository interface {
	Create(food *Food) (*Food, error)
	GetByID(id int) (*Food, error)
	GetAll() ([]*Food, error)
	Update(id int, updates *UpdateFoodRequest) (*Food, error)
	Delete(id int) error
}

// InMemoryRepository - keeping for backward compatibility
type InMemoryRepository struct {
	foods  map[int]*Food
	nextID int
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		foods:  make(map[int]*Food),
		nextID: 1,
	}
}

func (r *InMemoryRepository) Create(food *Food) (*Food, error) {
	food.ID = r.nextID
	r.nextID++
	food.CreatedAt = time.Now()
	food.UpdatedAt = time.Now()
	r.foods[food.ID] = food
	return food, nil
}

func (r *InMemoryRepository) GetByID(id int) (*Food, error) {
	food, exists := r.foods[id]
	if !exists {
		return nil, errors.New("food not found")
	}
	return food, nil
}

func (r *InMemoryRepository) GetAll() ([]*Food, error) {
	foods := make([]*Food, 0, len(r.foods))
	for _, food := range r.foods {
		foods = append(foods, food)
	}
	return foods, nil
}

func (r *InMemoryRepository) Update(id int, updates *UpdateFoodRequest) (*Food, error) {
	food, exists := r.foods[id]
	if !exists {
		return nil, errors.New("food not found")
	}

	if updates.Name != nil {
		food.Name = *updates.Name
	}
	if updates.Calories != nil {
		food.Calories = *updates.Calories
	}
	if updates.Protein != nil {
		food.Protein = *updates.Protein
	}
	if updates.Carbs != nil {
		food.Carbs = *updates.Carbs
	}
	if updates.Fat != nil {
		food.Fat = *updates.Fat
	}

	food.UpdatedAt = time.Now()
	return food, nil
}

func (r *InMemoryRepository) Delete(id int) error {
	if _, exists := r.foods[id]; !exists {
		return errors.New("food not found")
	}
	delete(r.foods, id)
	return nil
}

// PostgreSQLRepository - new PostgreSQL implementation
type PostgreSQLRepository struct {
	db *sql.DB
}

func NewPostgreSQLRepository(db *sql.DB) *PostgreSQLRepository {
	return &PostgreSQLRepository{db: db}
}

func (r *PostgreSQLRepository) Create(food *Food) (*Food, error) {
	query := `
		INSERT INTO foods (name, calories, protein, carbs, fat, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(query, food.Name, food.Calories, food.Protein, food.Carbs, food.Fat).
		Scan(&food.ID, &food.CreatedAt, &food.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create food: %w", err)
	}

	return food, nil
}

func (r *PostgreSQLRepository) GetByID(id int) (*Food, error) {
	query := `SELECT id, name, calories, protein, carbs, fat, created_at, updated_at 
			  FROM foods WHERE id = $1`

	food := &Food{}
	err := r.db.QueryRow(query, id).Scan(
		&food.ID, &food.Name, &food.Calories, &food.Protein,
		&food.Carbs, &food.Fat, &food.CreatedAt, &food.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("food not found")
		}
		return nil, fmt.Errorf("failed to get food: %w", err)
	}

	return food, nil
}

func (r *PostgreSQLRepository) GetAll() ([]*Food, error) {
	query := `SELECT id, name, calories, protein, carbs, fat, created_at, updated_at 
			  FROM foods ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all foods: %w", err)
	}
	defer rows.Close()

	var foods []*Food
	for rows.Next() {
		food := &Food{}
		err := rows.Scan(
			&food.ID, &food.Name, &food.Calories, &food.Protein,
			&food.Carbs, &food.Fat, &food.CreatedAt, &food.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan food: %w", err)
		}
		foods = append(foods, food)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over foods: %w", err)
	}

	return foods, nil
}

func (r *PostgreSQLRepository) Update(id int, updates *UpdateFoodRequest) (*Food, error) {
	// First, get the current food
	currentFood, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if updates.Name != nil {
		currentFood.Name = *updates.Name
	}
	if updates.Calories != nil {
		currentFood.Calories = *updates.Calories
	}
	if updates.Protein != nil {
		currentFood.Protein = *updates.Protein
	}
	if updates.Carbs != nil {
		currentFood.Carbs = *updates.Carbs
	}
	if updates.Fat != nil {
		currentFood.Fat = *updates.Fat
	}

	// Update in database
	query := `
		UPDATE foods 
		SET name = $1, calories = $2, protein = $3, carbs = $4, fat = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING updated_at`

	err = r.db.QueryRow(query, currentFood.Name, currentFood.Calories,
		currentFood.Protein, currentFood.Carbs, currentFood.Fat, id).
		Scan(&currentFood.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update food: %w", err)
	}

	return currentFood, nil
}

func (r *PostgreSQLRepository) Delete(id int) error {
	query := `DELETE FROM foods WHERE id = $1`
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete food: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("food not found")
	}

	return nil
}
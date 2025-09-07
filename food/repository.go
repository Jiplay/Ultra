package food

import (
	"errors"
	"sync"
	"time"
)

type Repository interface {
	Create(food *Food) (*Food, error)
	GetByID(id int) (*Food, error)
	GetAll() ([]*Food, error)
	Update(id int, updates *UpdateFoodRequest) (*Food, error)
	Delete(id int) error
}

type InMemoryRepository struct {
	foods   map[int]*Food
	nextID  int
	mu      sync.RWMutex
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		foods:  make(map[int]*Food),
		nextID: 1,
	}
}

func (r *InMemoryRepository) Create(food *Food) (*Food, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	food.ID = r.nextID
	r.nextID++
	food.CreatedAt = time.Now()
	food.UpdatedAt = time.Now()

	r.foods[food.ID] = food
	return food, nil
}

func (r *InMemoryRepository) GetByID(id int) (*Food, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	food, exists := r.foods[id]
	if !exists {
		return nil, errors.New("food not found")
	}

	return food, nil
}

func (r *InMemoryRepository) GetAll() ([]*Food, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	foods := make([]*Food, 0, len(r.foods))
	for _, food := range r.foods {
		foods = append(foods, food)
	}

	return foods, nil
}

func (r *InMemoryRepository) Update(id int, updates *UpdateFoodRequest) (*Food, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	food, exists := r.foods[id]
	if !exists {
		return nil, errors.New("food not found")
	}

	if updates.Name != nil {
		food.Name = *updates.Name
	}
	if updates.Description != nil {
		food.Description = *updates.Description
	}
	if updates.Category != nil {
		food.Category = *updates.Category
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
	if updates.Fiber != nil {
		food.Fiber = *updates.Fiber
	}
	if updates.Sugar != nil {
		food.Sugar = *updates.Sugar
	}
	if updates.Sodium != nil {
		food.Sodium = *updates.Sodium
	}

	food.UpdatedAt = time.Now()
	return food, nil
}

func (r *InMemoryRepository) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.foods[id]; !exists {
		return errors.New("food not found")
	}

	delete(r.foods, id)
	return nil
}
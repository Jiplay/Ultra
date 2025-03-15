package memory

import (
	"context"
	"sync"
	"ultra.com/food/internal/repository"
	"ultra.com/food/pkg/model"
)

type Repository struct {
	sync.RWMutex
	data map[model.RecipeID]*model.Recipe
}

func New() *Repository { return &Repository{data: make(map[model.RecipeID]*model.Recipe)} }

func (r *Repository) Get(_ context.Context, recipeID model.RecipeID) (*model.Recipe, error) {
	r.RLock()
	defer r.RUnlock()
	data, ok := r.data[recipeID]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return data, nil
}

func (r *Repository) Put(recipe *model.Recipe) error {
	r.Lock()
	defer r.Unlock()
	r.data[recipe.ID] = recipe
	return nil
}

package food

import (
	"context"
	"errors"
	"ultra.com/food/pkg/model"
)

var ErrNotFound = errors.New("food not found")

type foodRepository interface {
	Get(ctx context.Context, id model.RecipeID) (*model.Recipe, error)
}

type Controller struct {
	repo foodRepository
}

func New(repo foodRepository) *Controller { return &Controller{repo: repo} }

func (c *Controller) Get(ctx context.Context, id model.RecipeID) (*model.Recipe, error) {
	res, err := c.repo.Get(ctx, id)
	if err != nil && errors.Is(err, ErrNotFound) {
		return nil, ErrNotFound
	}
	return res, err
}

package sport

import (
	"context"
	"errors"
	"ultra.com/sport/pkg/model"
)

var ErrNotFound = errors.New("resources not found")

type sportRepository interface {
	GetPlan(ctx context.Context, id model.WorkoutPlanID) (*model.WorkoutPlan, error)
	GetPerformance(ctx context.Context, id model.WorkoutPerformanceID) (*model.WorkoutPerformance, error)
}

type Controller struct {
	repo sportRepository
}

func New(repo sportRepository) *Controller {
	return &Controller{repo}
}

func (c *Controller) GetPlan(ctx context.Context, id model.WorkoutPlanID) (*model.WorkoutPlan, error) {
	plan, err := c.repo.GetPlan(ctx, id)
	if err != nil && errors.Is(err, ErrNotFound) {
		return nil, ErrNotFound
	}
	return plan, err
}

func (c *Controller) GetPerformance(ctx context.Context, id model.WorkoutPerformanceID) (*model.WorkoutPerformance, error) {
	performance, err := c.repo.GetPerformance(ctx, id)
	if err != nil && errors.Is(err, ErrNotFound) {
		return nil, ErrNotFound
	}
	return performance, err
}

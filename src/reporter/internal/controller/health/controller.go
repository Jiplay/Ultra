package health

import (
	"context"
	"errors"
	food "ultra.com/food/pkg/model"
	"ultra.com/sport/pkg/model"
)

var ErrNotFound = errors.New("not found")

type sportGateway interface {
	GetWorkout(ctx context.Context, id model.WorkoutID) (model.Workout, error)
	GetPerformance(ctx context.Context, id model.WorkoutPerformanceID) (model.WorkoutPerformance, error)
}

type foodGateway interface {
	GetRecipe(ctx context.Context, id food.RecipeID) (food.Recipe, error)
}

type Controller struct {
	sportGateway sportGateway
	foodGateway  foodGateway
}

func New(sportGateway sportGateway, foodGateway foodGateway) *Controller {
	return &Controller{sportGateway: sportGateway, foodGateway: foodGateway}
}

func (c *Controller) GetWorkout(ctx context.Context, id model.WorkoutID) (model.Workout, error) {
	workout, err := c.sportGateway.GetWorkout(ctx, id)
	if err != nil && errors.Is(err, ErrNotFound) {
		return model.Workout{}, ErrNotFound
	} else if err != nil {
		return workout, err
	}
	return workout, nil

}

func (c *Controller) GetPerformance(ctx context.Context, id model.WorkoutPerformanceID) (model.WorkoutPerformance, error) {
	perf, err := c.sportGateway.GetPerformance(ctx, id)
	if err != nil && errors.Is(err, ErrNotFound) {
		return model.WorkoutPerformance{}, ErrNotFound
	} else if err != nil {
		return perf, err
	}
	return perf, nil
}

func (c *Controller) GetRecipe(ctx context.Context, id food.RecipeID) (food.Recipe, error) {
	recipe, err := c.foodGateway.GetRecipe(ctx, id)
	if err != nil && errors.Is(err, ErrNotFound) {
		return food.Recipe{}, ErrNotFound
	} else if err != nil {
		return food.Recipe{}, err
	}
	return recipe, nil
}

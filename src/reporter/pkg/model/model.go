package model

import (
	food "ultra.com/food/pkg/model"
	sport "ultra.com/sport/pkg/model"
)

type Health struct {
	Recipes      food.Recipe `json:"recipes"`
	WorkoutPlans sport.Workout
	Performances sport.WorkoutPerformance
}

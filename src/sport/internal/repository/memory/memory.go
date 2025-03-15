package memory

import (
	"context"
	"sync"
	"ultra.com/sport/internal/repository"
	"ultra.com/sport/pkg/model"
)

type Repository struct {
	sync.RWMutex
	planData        map[model.WorkoutID]*model.Workout
	performanceData map[model.WorkoutPerformanceID]*model.WorkoutPerformance
}

// New creates a new repository
func New() *Repository {
	return &Repository{
		planData:        make(map[model.WorkoutID]*model.Workout),
		performanceData: make(map[model.WorkoutPerformanceID]*model.WorkoutPerformance),
	}
}

func (r *Repository) GetPlan(_ context.Context, id model.WorkoutID) (*model.Workout, error) {
	r.Lock()
	defer r.Unlock()
	p, ok := r.planData[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return p, nil
}

func (r *Repository) GetPerformance(_ context.Context, id model.WorkoutPerformanceID) (*model.WorkoutPerformance, error) {
	r.Lock()
	defer r.Unlock()
	p, ok := r.performanceData[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return p, nil
}

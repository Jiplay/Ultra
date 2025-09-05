package programs

import (
	"database/sql"
	"errors"
)

type Controller struct {
	repo Repository
}

func NewController(repo Repository) *Controller {
	return &Controller{repo: repo}
}

func (c *Controller) CreateProgram(req *CreateProgramRequest, userID string) (*Program, error) {
	if req.Name == "" {
		return nil, errors.New("program name is required")
	}
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	program := &Program{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Workouts:    make([]Workout, len(req.Workouts)),
	}

	// Validate and convert workouts
	for i, workoutReq := range req.Workouts {
		if workoutReq.Name == "" {
			return nil, errors.New("workout name is required")
		}
		if workoutReq.DayOfWeek < 0 || workoutReq.DayOfWeek > 6 {
			return nil, errors.New("day of week must be between 0 (Sunday) and 6 (Saturday)")
		}

		workout := Workout{
			Name:        workoutReq.Name,
			Description: workoutReq.Description,
			DayOfWeek:   workoutReq.DayOfWeek,
			Exercises:   make([]Exercise, len(workoutReq.Exercises)),
		}

		// Validate and convert exercises
		for j, exerciseReq := range workoutReq.Exercises {
			if exerciseReq.Name == "" {
				return nil, errors.New("exercise name is required")
			}
			if exerciseReq.Weight < 0 {
				return nil, errors.New("exercise weight cannot be negative")
			}
			if exerciseReq.Repetitions < 0 {
				return nil, errors.New("exercise repetitions cannot be negative")
			}
			if exerciseReq.Series < 0 {
				return nil, errors.New("exercise series cannot be negative")
			}
			if exerciseReq.RestTime < 0 {
				return nil, errors.New("exercise rest time cannot be negative")
			}

			exercise := Exercise{
				Name:        exerciseReq.Name,
				Weight:      exerciseReq.Weight,
				Repetitions: exerciseReq.Repetitions,
				Series:      exerciseReq.Series,
				RestTime:    exerciseReq.RestTime,
				Notes:       exerciseReq.Notes,
			}

			workout.Exercises[j] = exercise
		}

		program.Workouts[i] = workout
	}

	err := c.repo.CreateProgram(program)
	if err != nil {
		return nil, err
	}

	return program, nil
}

func (c *Controller) GetProgramsByUserID(userID string) ([]Program, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}
	return c.repo.GetProgramsByUserID(userID)
}

func (c *Controller) GetProgramByID(id int, userID string) (*Program, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}
	if id <= 0 {
		return nil, errors.New("invalid program ID")
	}
	return c.repo.GetProgramByID(id, userID)
}

func (c *Controller) UpdateProgram(id int, userID string, req *UpdateProgramRequest) error {
	if userID == "" {
		return errors.New("user ID is required")
	}
	if id <= 0 {
		return errors.New("invalid program ID")
	}

	// Validate request fields
	if req.Name != nil && *req.Name == "" {
		return errors.New("program name cannot be empty")
	}

	// Check if program exists
	exists, err := c.repo.CheckProgramExists(id, userID)
	if err != nil {
		return err
	}
	if !exists {
		return sql.ErrNoRows
	}

	return c.repo.UpdateProgram(id, userID, req)
}

func (c *Controller) DeleteProgram(id int, userID string) error {
	if userID == "" {
		return errors.New("user ID is required")
	}
	if id <= 0 {
		return errors.New("invalid program ID")
	}

	return c.repo.DeleteProgram(id, userID)
}

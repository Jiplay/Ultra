package programs

import "time"

type Program struct {
	ID          int       `json:"id" db:"id"`
	UserID      string    `json:"user_id" db:"user_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	Workouts    []Workout `json:"workouts,omitempty"`
}

type Workout struct {
	ID          int        `json:"id" db:"id"`
	ProgramID   int        `json:"program_id" db:"program_id"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description" db:"description"`
	DayOfWeek   int        `json:"day_of_week" db:"day_of_week"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	Exercises   []Exercise `json:"exercises,omitempty"`
}

type Exercise struct {
	ID          int     `json:"id" db:"id"`
	WorkoutID   int     `json:"workout_id" db:"workout_id"`
	Name        string  `json:"name" db:"name"`
	Weight      float64 `json:"weight" db:"weight"`
	Repetitions int     `json:"repetitions" db:"repetitions"`
	Series      int     `json:"series" db:"series"`
	RestTime    int     `json:"rest_time" db:"rest_time"`
	Notes       string  `json:"notes" db:"notes"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type CreateProgramRequest struct {
	Name        string            `json:"name" validate:"required"`
	Description string            `json:"description"`
	Workouts    []CreateWorkout   `json:"workouts"`
}

type CreateWorkout struct {
	Name        string            `json:"name" validate:"required"`
	Description string            `json:"description"`
	DayOfWeek   int               `json:"day_of_week"`
	Exercises   []CreateExercise  `json:"exercises"`
}

type CreateExercise struct {
	Name        string  `json:"name" validate:"required"`
	Weight      float64 `json:"weight"`
	Repetitions int     `json:"repetitions"`
	Series      int     `json:"series"`
	RestTime    int     `json:"rest_time"`
	Notes       string  `json:"notes"`
}

type UpdateProgramRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}
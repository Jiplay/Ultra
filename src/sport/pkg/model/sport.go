package model

import "time"

type WorkoutPlanID string
type ExercisePlanID string
type WorkoutPerformanceID string
type ExercisePerformanceID string

type WorkoutPlan struct {
	ID          WorkoutPlanID  `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Exercises   []ExercisePlan `json:"exercises"`
}

type ExercisePlan struct {
	ID          ExercisePlanID `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Series      uint8          `json:"series"`
	Repetitions uint8          `json:"repetitions"`
	Weight      uint16         `json:"weight"`
	RestTime    uint16         `json:"rest_time"`
}

type WorkoutPerformance struct {
	ID                   WorkoutPerformanceID   `json:"id"`
	PlanID               WorkoutPlanID          `json:"plan_id"`
	Date                 time.Time              `json:"date"`
	ExercisesPerformance []ExercisesPerformance `json:"exercises_performance"`
}

type ExercisesPerformance struct {
	ID          ExercisePerformanceID `json:"id"`
	Date        time.Time             `json:"date"`
	Weight      uint16                `json:"weight"`
	Repetitions uint8                 `json:"repetitions"`
	RestTime    uint16                `json:"rest_time"`
}

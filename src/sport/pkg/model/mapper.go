package model

import (
	"google.golang.org/protobuf/types/known/timestamppb"
	"ultra.com/gen"
)

// WorkoutToProto convert a Go struct Workout to the proto version
func WorkoutToProto(w *Workout) *gen.Workout {
	exercises := make([]*gen.ExercisePlan, len(w.Exercises))
	for i, e := range w.Exercises {
		exercises[i] = &gen.ExercisePlan{
			Id:          string(e.ID),
			Name:        e.Name,
			Description: e.Description,
			Series:      uint32(e.Series),
			Repetitions: uint32(e.Repetitions),
			Weight:      uint32(e.Weight),
			RestTime:    uint32(e.RestTime),
		}
	}

	return &gen.Workout{
		Id:          string(w.ID),
		Name:        w.Name,
		Description: w.Description,
		Exercises:   exercises,
	}
}

// WorkoutFromProto convert a generated struct Workout to a Go version
func WorkoutFromProto(w *gen.Workout) *Workout {
	exercises := make([]ExercisePlan, len(w.Exercises))
	for i, e := range w.Exercises {
		exercises[i] = ExercisePlan{
			ID:          ExercisePlanID(e.Id),
			Name:        e.Name,
			Description: e.Description,
			Series:      uint8(e.Series),
			Repetitions: uint8(e.Repetitions),
			Weight:      uint16(e.Weight),
			RestTime:    uint16(e.RestTime),
		}
	}
	return &Workout{
		ID:          WorkoutID(w.Id),
		Name:        w.Name,
		Description: w.Description,
		Exercises:   exercises,
	}
}

func PerformanceToProto(p *WorkoutPerformance) *gen.Performance {
	exercise := make([]*gen.ExercisePerformance, len(p.ExercisesPerformance))
	for i, e := range p.ExercisesPerformance {
		exercise[i] = &gen.ExercisePerformance{
			Id:          string(e.ID),
			Weight:      uint32(e.Weight),
			RestTime:    uint32(e.RestTime),
			Repetitions: uint32(e.Repetitions),
		}
	}
	return &gen.Performance{
		Id:           string(p.ID),
		WorkoutId:    string(p.PlanID),
		Date:         timestamppb.New(p.Date),
		Performances: exercise,
	}
}

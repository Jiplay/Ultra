package testutil

// DefaultGoalMacros returns default nutrition goal values
type DefaultGoalMacros struct {
	Calories float64
	Protein  float64
	Carbs    float64
	Fat      float64
	Fiber    float64
}

// GetDefaultGoalMacros returns standard test goal values
func GetDefaultGoalMacros() DefaultGoalMacros {
	return DefaultGoalMacros{
		Calories: 2000,
		Protein:  150,
		Carbs:    200,
		Fat:      65,
		Fiber:    30,
	}
}

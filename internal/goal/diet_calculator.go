package goal

import (
	"fmt"
	"ultra-bis/internal/user"
)

// DietCalculator is the interface that all diet calculators must implement
type DietCalculator interface {
	// ValidateUser validates that the user has all required data for this diet model
	ValidateUser(user *user.User) error

	// ValidateProtocol validates that the protocol number is valid for this diet model
	ValidateProtocol(protocol int) error

	// Calculate performs the diet calculations and returns results
	Calculate(user *user.User, protocol int) (*DietCalculationResult, error)

	// GetProtocolName returns the human-readable name for a protocol
	GetProtocolName(protocol int) string

	// GetModelName returns the diet model name
	GetModelName() string
}

// DietCalculationResult holds the calculation results for any diet model
type DietCalculationResult struct {
	ModelName      string
	Protocol       int
	ProtocolName   string
	BMR            float64
	MaintenanceMMR float64
	LeanMass       float64
	Phases         []PhaseResult
	Metadata       map[string]interface{} // For model-specific data
}

// PhaseResult holds the calorie and macro targets for a specific phase
type PhaseResult struct {
	Phase       int
	Calories    float64
	Protein     float64
	Carbs       float64
	Fat         float64
	Description string
}

// GetDietCalculator is a factory function that returns the appropriate calculator
func GetDietCalculator(modelName string) (DietCalculator, error) {
	switch modelName {
	case "zeroToHero":
		return NewZeroToHeroCalculator(), nil
	default:
		return nil, fmt.Errorf("unsupported diet model: %s. Supported models: zeroToHero", modelName)
	}
}

// ValidateDietRequest validates the diet calculation request
func ValidateDietRequest(req *CalculateDietRequest, user *user.User) error {
	// Get the calculator for this model
	calculator, err := GetDietCalculator(req.DietModel)
	if err != nil {
		return err
	}

	// Validate user data
	if err := calculator.ValidateUser(user); err != nil {
		return fmt.Errorf("user validation failed: %w", err)
	}

	// Validate protocol
	if err := calculator.ValidateProtocol(req.Protocol); err != nil {
		return fmt.Errorf("protocol validation failed: %w", err)
	}

	return nil
}

package goal

import (
	"fmt"
	"ultra-bis/internal/user"
)

// ZeroToHeroCalculator handles Zero to Hero diet calculations
type ZeroToHeroCalculator struct{}

// NewZeroToHeroCalculator creates a new calculator instance
func NewZeroToHeroCalculator() *ZeroToHeroCalculator {
	return &ZeroToHeroCalculator{}
}

// GetModelName returns the diet model name
func (c *ZeroToHeroCalculator) GetModelName() string {
	return "zeroToHero"
}

// ValidateUser validates that the user has all required data for Zero to Hero calculations
func (c *ZeroToHeroCalculator) ValidateUser(user *user.User) error {
	if user.Age <= 0 {
		return fmt.Errorf("age is required and must be greater than 0")
	}
	if user.Height <= 0 {
		return fmt.Errorf("height is required and must be greater than 0")
	}
	if user.Weight <= 0 {
		return fmt.Errorf("weight is required and must be greater than 0")
	}
	if user.BodyFat <= 0 || user.BodyFat >= 100 {
		return fmt.Errorf("body fat is required and must be between 0 and 100")
	}
	return nil
}

// ValidateProtocol validates that the protocol number is valid for Zero to Hero
func (c *ZeroToHeroCalculator) ValidateProtocol(protocol int) error {
	if protocol < 1 || protocol > 4 {
		return fmt.Errorf("protocol must be between 1 and 4 for Zero to Hero")
	}
	return nil
}

// GetProtocolName returns the human-readable protocol name
func (c *ZeroToHeroCalculator) GetProtocolName(protocol int) string {
	switch protocol {
	case 1:
		return "Protocole 1 : Prise de muscle propre"
	case 2:
		return "Protocole 2 : Recomposition corporelle"
	case 3:
		return "Protocole 3 : Créer le déficit parfait"
	case 4:
		return "Protocole 4 : Perte de gras progressive"
	default:
		return "Unknown Protocol"
	}
}

// Calculate performs the Zero to Hero calculations and returns standardized results
func (c *ZeroToHeroCalculator) Calculate(user *user.User, protocol int) (*DietCalculationResult, error) {
	// Validation is done in the handler using ValidateUser and ValidateProtocol
	// This allows for better separation of concerns

	// Perform internal calculations
	internalResult := c.performCalculations(user, protocol)

	// Convert to standardized result
	result := &DietCalculationResult{
		ModelName:      c.GetModelName(),
		Protocol:       protocol,
		ProtocolName:   c.GetProtocolName(protocol),
		BMR:            internalResult.BMR,
		MaintenanceMMR: internalResult.MaintenanceMMR,
		LeanMass:       internalResult.LeanMass,
		Phases:         internalResult.Phases,
		Metadata: map[string]interface{}{
			"bmr1": internalResult.BMR1,
			"bmr2": internalResult.BMR2,
		},
	}

	return result, nil
}

// internalCalculationResult holds Zero to Hero specific calculation data
type internalCalculationResult struct {
	BMR1           float64 // Mifflin-St Jeor variant
	BMR2           float64 // Lean body mass variant
	BMR            float64 // Average BMR
	MaintenanceMMR float64 // Maintenance calories (BMR * 1.5)
	LeanMass       float64 // Lean body mass
	Phases         []PhaseResult
}

// performCalculations performs the actual Zero to Hero calculations
func (c *ZeroToHeroCalculator) performCalculations(user *user.User, protocol int) *internalCalculationResult {
	result := &internalCalculationResult{}

	// Convert height from cm to meters for calculations
	heightInMeters := user.Height / 100

	// BMR Calculation 1: Mifflin-St Jeor variant
	// Formula: =(13.707*Weight)+(492.3*Height)-(6.673*Age)+77.607
	result.BMR1 = (13.707 * user.Weight) + (492.3 * heightInMeters) - (6.673 * float64(user.Age)) + 77.607

	// BMR Calculation 2: Lean body mass variant
	// Formula: =(21.6*Weight*(100-BodyFat)/100)+370
	result.BMR2 = (21.6 * user.Weight * (100 - user.BodyFat) / 100) + 370

	// Average BMR
	result.BMR = (result.BMR1 + result.BMR2) / 2

	// Maintenance calories (MMR)
	result.MaintenanceMMR = result.BMR * 1.5

	// Lean body mass
	// Formula: =Weight*(100-BodyFat)/100
	result.LeanMass = user.Weight * (100 - user.BodyFat) / 100

	// Calculate protocol-specific phases
	result.Phases = c.calculatePhases(user, protocol, result)

	return result
}

// calculatePhases calculates the calorie and macro targets for each phase based on protocol
func (c *ZeroToHeroCalculator) calculatePhases(user *user.User, protocol int, calc *internalCalculationResult) []PhaseResult {
	var phases []PhaseResult

	switch protocol {
	case 1: // Protocol 1: Muscle Building
		phases = []PhaseResult{
			{
				Phase:       1,
				Calories:    calc.MaintenanceMMR,
				Description: "Maintenance phase",
			},
			{
				Phase:       2,
				Calories:    calc.MaintenanceMMR + 200,
				Description: "Moderate surplus phase",
			},
			{
				Phase:       3,
				Calories:    calc.MaintenanceMMR + 400,
				Description: "High surplus phase",
			},
		}

	case 2: // Protocol 2: Recomposition
		phases = []PhaseResult{
			{
				Phase:       1,
				Calories:    calc.MaintenanceMMR - 300,
				Description: "Recomposition phase",
			},
		}

	case 3: // Protocol 3: Perfect Deficit
		phases = []PhaseResult{
			{
				Phase:       1,
				Calories:    calc.MaintenanceMMR - 300,
				Description: "Moderate deficit phase",
			},
			{
				Phase:       2,
				Calories:    calc.MaintenanceMMR - 500,
				Description: "Higher deficit phase",
			},
		}

	case 4: // Protocol 4: Progressive Fat Loss
		phases = []PhaseResult{
			{
				Phase:       1,
				Calories:    calc.MaintenanceMMR - 300,
				Description: "Initial deficit phase",
			},
			{
				Phase:       2,
				Calories:    calc.MaintenanceMMR - 500,
				Description: "Moderate deficit phase",
			},
			{
				Phase:       3,
				Calories:    calc.MaintenanceMMR - 700,
				Description: "Aggressive deficit phase",
			},
		}
	}

	// Calculate macros for each phase
	for i := range phases {
		c.calculateMacros(user, calc, &phases[i])
	}

	return phases
}

// calculateMacros calculates protein, fat, and carbs for a given calorie target
func (c *ZeroToHeroCalculator) calculateMacros(user *user.User, calc *internalCalculationResult, phase *PhaseResult) {
	// Protein calculation: Average of (Weight * 1.5) and (LeanMass * 2)
	// Formula from Excel: =AVERAGE(Weight*1.5, LeanMass*2)
	protein1 := user.Weight * 1.5
	protein2 := calc.LeanMass * 2
	phase.Protein = (protein1 + protein2) / 2

	// Fat calculation: 1.2 * lean body mass
	// Formula: =1.2*Weight*(1-BodyFat/100)
	phase.Fat = 1.2 * user.Weight * (1 - user.BodyFat/100)

	// Carbohydrates calculation: Remaining calories divided by 4
	// Formula: =(Calories-(Protein*4)-(Fat*9))/4
	caloriesFromProtein := phase.Protein * 4
	caloriesFromFat := phase.Fat * 9
	remainingCalories := phase.Calories - caloriesFromProtein - caloriesFromFat
	phase.Carbs = remainingCalories / 4
}

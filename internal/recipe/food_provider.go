package recipe

// Food represents a food item with nutritional information (per 100g)
type Food struct {
	ID          uint
	Name        string
	Description string
	Calories    float64
	Protein     float64
	Carbs       float64
	Fat         float64
	Fiber       float64
}

// FoodProvider defines the interface for retrieving food information
// This allows the recipe package to depend on an abstraction rather than
// a concrete food repository implementation
type FoodProvider interface {
	// GetByID retrieves a single food item by ID
	GetByID(id int) (*Food, error)

	// GetByIDs retrieves multiple food items by their IDs in a single query
	// This method enables efficient batch fetching to avoid N+1 queries
	GetByIDs(ids []int) ([]*Food, error)
}

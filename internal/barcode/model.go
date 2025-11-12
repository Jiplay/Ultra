package barcode

// OpenFoodFactsResponse represents the response from Open Food Facts API
type OpenFoodFactsResponse struct {
	Code          string              `json:"code"`
	Status        int                 `json:"status"`
	StatusVerbose string              `json:"status_verbose"`
	Product       *OpenFoodFactsProduct `json:"product,omitempty"`
}

// OpenFoodFactsProduct represents a product from Open Food Facts API
type OpenFoodFactsProduct struct {
	ProductName  string                      `json:"product_name"`
	GenericName  string                      `json:"generic_name"`
	Brands       string                      `json:"brands"`
	Nutriments   OpenFoodFactsNutriments     `json:"nutriments"`
}

// OpenFoodFactsNutriments represents nutritional values from Open Food Facts
// All values are per 100g
type OpenFoodFactsNutriments struct {
	EnergyKcal100g    float64 `json:"energy-kcal_100g"`
	Proteins100g      float64 `json:"proteins_100g"`
	Carbohydrates100g float64 `json:"carbohydrates_100g"`
	Fat100g           float64 `json:"fat_100g"`
	Fiber100g         float64 `json:"fiber_100g"`
}

// ProductData represents the processed product data ready to be used
type ProductData struct {
	Name        string
	Description string
	Calories    float64
	Protein     float64
	Carbs       float64
	Fat         float64
	Fiber       float64
}

package barcode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService_convertToProductData(t *testing.T) {
	service := NewService()

	tests := []struct {
		name     string
		product  *OpenFoodFactsProduct
		expected *ProductData
	}{
		{
			name: "complete product data",
			product: &OpenFoodFactsProduct{
				ProductName: "Nutella",
				GenericName: "Hazelnut cocoa spread",
				Brands:      "Ferrero",
				Countries:   "France",
				Nutriments: OpenFoodFactsNutriments{
					EnergyKcal100g:    539,
					Proteins100g:      6.3,
					Carbohydrates100g: 57.5,
					Fat100g:           30.9,
					Fiber100g:         0,
				},
			},
			expected: &ProductData{
				Name:        "Nutella",
				Description: "Ferrero - Hazelnut cocoa spread",
				Country:     "France",
				Calories:    539,
				Protein:     6.3,
				Carbs:       57.5,
				Fat:         30.9,
				Fiber:       0,
			},
		},
		{
			name: "product with only brand",
			product: &OpenFoodFactsProduct{
				ProductName: "Coca Cola",
				Brands:      "Coca-Cola",
				Countries:   "United States",
				Nutriments: OpenFoodFactsNutriments{
					EnergyKcal100g:    42,
					Proteins100g:      0,
					Carbohydrates100g: 10.6,
					Fat100g:           0,
					Fiber100g:         0,
				},
			},
			expected: &ProductData{
				Name:        "Coca Cola",
				Description: "Coca-Cola",
				Country:     "United States",
				Calories:    42,
				Protein:     0,
				Carbs:       10.6,
				Fat:         0,
				Fiber:       0,
			},
		},
		{
			name: "product with only generic name",
			product: &OpenFoodFactsProduct{
				ProductName: "Apple",
				GenericName: "Fresh fruit",
				Nutriments: OpenFoodFactsNutriments{
					EnergyKcal100g:    52,
					Proteins100g:      0.3,
					Carbohydrates100g: 14,
					Fat100g:           0.2,
					Fiber100g:         2.4,
				},
			},
			expected: &ProductData{
				Name:        "Apple",
				Description: "Fresh fruit",
				Country:     "",
				Calories:    52,
				Protein:     0.3,
				Carbs:       14,
				Fat:         0.2,
				Fiber:       2.4,
			},
		},
		{
			name: "product with no brand or generic name",
			product: &OpenFoodFactsProduct{
				ProductName: "Mystery Product",
				Nutriments: OpenFoodFactsNutriments{
					EnergyKcal100g:    100,
					Proteins100g:      5,
					Carbohydrates100g: 10,
					Fat100g:           2,
					Fiber100g:         1,
				},
			},
			expected: &ProductData{
				Name:        "Mystery Product",
				Description: "",
				Country:     "",
				Calories:    100,
				Protein:     5,
				Carbs:       10,
				Fat:         2,
				Fiber:       1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.convertToProductData(tt.product)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestService_ScanBarcode_EmptyBarcode(t *testing.T) {
	service := NewService()

	result, err := service.ScanBarcode("")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "barcode cannot be empty")
}

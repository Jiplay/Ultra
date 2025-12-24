package barcode

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	openFoodFactsAPIURL = "https://world.openfoodfacts.org/api/v2/product/%s.json"
	userAgent           = "Ultra-Bis/1.0 (nutrition-tracking-app)"
	requestTimeout      = 10 * time.Second
)

// Service handles barcode scanning operations
type Service struct {
	httpClient *http.Client
}

// NewService creates a new barcode service
func NewService() *Service {
	return &Service{
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

// ScanBarcode fetches product data from Open Food Facts by barcode
func (s *Service) ScanBarcode(barcode string) (*ProductData, error) {
	if barcode == "" {
		return nil, fmt.Errorf("barcode cannot be empty")
	}

	url := fmt.Sprintf(openFoodFactsAPIURL, barcode)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Open Food Facts requires a User-Agent header
	req.Header.Set("User-Agent", userAgent)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch product data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code %d", resp.StatusCode)
	}

	var offResp OpenFoodFactsResponse
	if err := json.NewDecoder(resp.Body).Decode(&offResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check if product was found
	if offResp.Status != 1 || offResp.Product == nil {
		return nil, fmt.Errorf("product not found for barcode %s", barcode)
	}

	// Convert Open Food Facts data to our internal format
	productData := s.ConvertToProductData(offResp.Product)

	return productData, nil
}

// ConvertToProductData converts Open Food Facts product to our internal format
func (s *Service) ConvertToProductData(product *OpenFoodFactsProduct) *ProductData {
	// Build description from brand and generic name
	description := ""
	if product.Brands != "" {
		description = product.Brands
	}
	if product.GenericName != "" {
		if description != "" {
			description += " - " + product.GenericName
		} else {
			description = product.GenericName
		}
	}

	return &ProductData{
		Name:        product.ProductName,
		Description: description,
		Calories:    product.Nutriments.EnergyKcal100g,
		Protein:     product.Nutriments.Proteins100g,
		Carbs:       product.Nutriments.Carbohydrates100g,
		Fat:         product.Nutriments.Fat100g,
		Fiber:       product.Nutriments.Fiber100g,
	}
}

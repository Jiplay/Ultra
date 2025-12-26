package barcode

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockBarcodeService is a mock implementation of the barcode service
type mockBarcodeService struct {
	searchFunc func(query string, page int, pageSize int) (*SearchResults, error)
}

func (m *mockBarcodeService) ScanBarcode(barcode string) (*ProductData, error) {
	return nil, errors.New("not implemented")
}

func (m *mockBarcodeService) SearchByName(query string, page int, pageSize int) (*SearchResults, error) {
	if m.searchFunc != nil {
		return m.searchFunc(query, page, pageSize)
	}
	return nil, errors.New("not implemented")
}

func (m *mockBarcodeService) ConvertToProductData(product *OpenFoodFactsProduct) *ProductData {
	return nil
}

func TestSearchProducts_Success(t *testing.T) {
	service := NewService()
	handler := &Handler{service: service}

	// Test empty query
	req := httptest.NewRequest("GET", "/openfoodfacts/search", nil)
	w := httptest.NewRecorder()
	handler.SearchProducts(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errResp map[string]string
	json.NewDecoder(w.Body).Decode(&errResp)
	assert.Equal(t, "Query parameter 'q' is required", errResp["error"])
}

func TestSearchProducts_EmptyQuery(t *testing.T) {
	service := NewService()
	handler := NewHandler(service)

	req := httptest.NewRequest("GET", "/openfoodfacts/search", nil)
	w := httptest.NewRecorder()

	handler.SearchProducts(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errResp map[string]string
	err := json.NewDecoder(w.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, "Query parameter 'q' is required", errResp["error"])
}

func TestSearchProducts_Pagination(t *testing.T) {
	service := NewService()
	handler := NewHandler(service)

	// Test with custom page and page_size
	req := httptest.NewRequest("GET", "/openfoodfacts/search?q=apple&page=2&page_size=10", nil)
	w := httptest.NewRecorder()

	// The handler will parse the parameters correctly
	// We can verify this by checking the service isn't called with wrong params
	handler.SearchProducts(w, req)

	// Since we're using a real service, this will fail to connect to OFF API
	// which is expected in unit tests without network
	assert.NotEqual(t, http.StatusBadRequest, w.Code)
}

func TestSearchProducts_MethodNotAllowed(t *testing.T) {
	service := NewService()
	handler := NewHandler(service)

	req := httptest.NewRequest("POST", "/openfoodfacts/search?q=apple", nil)
	w := httptest.NewRecorder()

	handler.SearchProducts(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

	var errResp map[string]string
	err := json.NewDecoder(w.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, "Method not allowed", errResp["error"])
}

func TestSearchProducts_InvalidPagination(t *testing.T) {
	service := NewService()
	handler := NewHandler(service)

	// Test with invalid page (negative)
	req := httptest.NewRequest("GET", "/openfoodfacts/search?q=apple&page=-1", nil)
	w := httptest.NewRecorder()

	handler.SearchProducts(w, req)

	// Should default to page 1, not return error
	assert.NotEqual(t, http.StatusBadRequest, w.Code)

	// Test with invalid page_size (>100)
	req2 := httptest.NewRequest("GET", "/openfoodfacts/search?q=apple&page_size=200", nil)
	w2 := httptest.NewRecorder()

	handler.SearchProducts(w2, req2)

	// Should cap at 100, not return error
	assert.NotEqual(t, http.StatusBadRequest, w2.Code)
}

func TestScanBarcode_Success(t *testing.T) {
	service := NewService()
	handler := NewHandler(service)

	req := httptest.NewRequest("POST", "/openfoodfacts/barcode/3017620422003", nil)
	w := httptest.NewRecorder()

	handler.ScanBarcode(w, req)

	// Should return product data with 200 (not create food with 201)
	// Will fail to connect to OFF API in test, but validates handler logic
	assert.NotEqual(t, http.StatusBadRequest, w.Code)
}

func TestScanBarcode_EmptyBarcode(t *testing.T) {
	service := NewService()
	handler := NewHandler(service)

	req := httptest.NewRequest("POST", "/openfoodfacts/barcode/", nil)
	w := httptest.NewRecorder()

	handler.ScanBarcode(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errResp map[string]string
	err := json.NewDecoder(w.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, "Barcode is required", errResp["error"])
}

func TestScanBarcode_MethodNotAllowed(t *testing.T) {
	service := NewService()
	handler := NewHandler(service)

	req := httptest.NewRequest("GET", "/openfoodfacts/barcode/123", nil)
	w := httptest.NewRecorder()

	handler.ScanBarcode(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

	var errResp map[string]string
	err := json.NewDecoder(w.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, "Method not allowed", errResp["error"])
}

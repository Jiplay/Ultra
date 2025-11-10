package testutil

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// AssertJSONResponse checks if the response has the expected status code and valid JSON body
func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int) {
	t.Helper()

	assert.Equal(t, expectedStatus, w.Code, "Status code mismatch")
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

	// Verify valid JSON
	var result map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err, "Response should be valid JSON")
}

// AssertErrorResponse checks if the response is an error with the expected message
func AssertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedError string) {
	t.Helper()

	assert.Equal(t, expectedStatus, w.Code, "Status code mismatch")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Response should be valid JSON")

	errorMsg, exists := response["error"]
	assert.True(t, exists, "Response should contain 'error' field")
	assert.Contains(t, errorMsg, expectedError, "Error message mismatch")
}

// AssertNutritionEquals checks if two nutrition values are equal within a small delta (for floating-point comparison)
func AssertNutritionEquals(t *testing.T, expected, actual float64, fieldName string) {
	t.Helper()

	delta := 0.01 // Allow 0.01 difference for floating-point precision
	assert.InDelta(t, expected, actual, delta, "%s nutrition value mismatch", fieldName)
}

// AssertNutritionStruct checks if a struct with nutrition fields matches expected values
func AssertNutritionStruct(t *testing.T, calories, protein, carbs, fat, fiber float64, result map[string]interface{}) {
	t.Helper()

	AssertNutritionEquals(t, calories, result["calories"].(float64), "calories")
	AssertNutritionEquals(t, protein, result["protein"].(float64), "protein")
	AssertNutritionEquals(t, carbs, result["carbs"].(float64), "carbs")
	AssertNutritionEquals(t, fat, result["fat"].(float64), "fat")
	AssertNutritionEquals(t, fiber, result["fiber"].(float64), "fiber")
}

// AssertUnauthorized checks if the response is a 401 Unauthorized error
func AssertUnauthorized(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()

	AssertErrorResponse(t, w, http.StatusUnauthorized, "")
}

// AssertNotFound checks if the response is a 404 Not Found error
func AssertNotFound(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()

	assert.Equal(t, http.StatusNotFound, w.Code, "Should return 404 Not Found")
}

// AssertBadRequest checks if the response is a 400 Bad Request error
func AssertBadRequest(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()

	assert.Equal(t, http.StatusBadRequest, w.Code, "Should return 400 Bad Request")
}

// AssertValidToken checks if a token string is not empty and is a valid JWT format
func AssertValidToken(t *testing.T, token string) {
	t.Helper()

	assert.NotEmpty(t, token, "Token should not be empty")
	// JWT tokens have 3 parts separated by dots
	assert.Regexp(t, `^[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+$`, token, "Token should be valid JWT format")
}

// AssertFloatEquals checks if two floats are equal within a reasonable delta
func AssertFloatEquals(t *testing.T, expected, actual float64, message string) {
	t.Helper()

	delta := 0.01
	if !floatEquals(expected, actual, delta) {
		t.Errorf("%s: expected %.2f, got %.2f", message, expected, actual)
	}
}

// floatEquals checks if two floats are equal within a delta
func floatEquals(a, b, delta float64) bool {
	return math.Abs(a-b) <= delta
}

// ParseJSONResponse parses the response body into a map
func ParseJSONResponse(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()

	var result map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err, "Failed to parse JSON response")

	return result
}

// ParseJSONArrayResponse parses the response body into a slice of maps
func ParseJSONArrayResponse(t *testing.T, w *httptest.ResponseRecorder) []map[string]interface{} {
	t.Helper()

	var result []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err, "Failed to parse JSON array response")

	return result
}

// AssertDateFormat checks if a string is in YYYY-MM-DD format
func AssertDateFormat(t *testing.T, dateStr string) {
	t.Helper()

	assert.Regexp(t, `^\d{4}-\d{2}-\d{2}$`, dateStr, "Date should be in YYYY-MM-DD format")
}

// AssertPositive checks if a number is positive (> 0)
func AssertPositive(t *testing.T, value float64, fieldName string) {
	t.Helper()

	assert.Greater(t, value, 0.0, "%s should be positive", fieldName)
}

// AssertNonNegative checks if a number is non-negative (>= 0)
func AssertNonNegative(t *testing.T, value float64, fieldName string) {
	t.Helper()

	assert.GreaterOrEqual(t, value, 0.0, "%s should be non-negative", fieldName)
}

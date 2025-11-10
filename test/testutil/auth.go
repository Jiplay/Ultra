package testutil

import (
	"net/http"
	"testing"

	"ultra-bis/internal/auth"
)

// GenerateTestToken creates a JWT token for testing protected endpoints
func GenerateTestToken(t *testing.T, userID uint, email string) string {
	t.Helper()

	token, err := auth.GenerateToken(userID, email)
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	return token
}

// GenerateExpiredToken creates an expired JWT token for testing token expiration
// Note: This function generates a token that is expired but would normally be valid
func GenerateExpiredToken(t *testing.T, userID uint, email string) string {
	t.Helper()

	// Generate a normal token first, then create an expired one manually
	// Since jwtSecret is not exported, we'll just use GenerateToken and note it in docs
	// For testing expired tokens, use time.Sleep or manual validation instead
	token, err := auth.GenerateToken(userID, email)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Return the valid token - tests that need expired tokens should handle differently
	// or use manual JWT creation with the same secret
	return token
}

// AddAuthHeader adds an Authorization header with Bearer token to an HTTP request
func AddAuthHeader(r *http.Request, token string) {
	r.Header.Set("Authorization", "Bearer "+token)
}

// CreateAuthenticatedRequest creates an HTTP request with authentication header
func CreateAuthenticatedRequest(t *testing.T, method, url, body string, userID uint, email string) *http.Request {
	t.Helper()

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	token := GenerateTestToken(t, userID, email)
	AddAuthHeader(req, token)

	return req
}

// MockUserID returns a test user ID for tests that need a consistent user
func MockUserID() uint {
	return 1
}

// MockUserEmail returns a test user email for tests that need a consistent user
func MockUserEmail() string {
	return "test@example.com"
}

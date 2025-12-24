package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"ultra-bis/internal/httputil"
)

// TestGetUserID tests extracting user ID from context
func TestGetUserID(t *testing.T) {
	tests := []struct {
		name           string
		userID         uint
		shouldSetValue bool
		expectedOK     bool
	}{
		{
			name:           "valid user ID",
			userID:         123,
			shouldSetValue: true,
			expectedOK:     true,
		},
		{
			name:           "no user ID in context",
			userID:         0,
			shouldSetValue: false,
			expectedOK:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			if tt.shouldSetValue {
				ctx := httputil.SetUserID(req.Context(), tt.userID)
				req = req.WithContext(ctx)
			}

			userID, ok := httputil.GetUserID(req)

			if ok != tt.expectedOK {
				t.Errorf("Expected ok=%v, got %v", tt.expectedOK, ok)
			}

			if tt.shouldSetValue && userID != tt.userID {
				t.Errorf("Expected userID=%d, got %d", tt.userID, userID)
			}
		})
	}
}

// TestSetUserID tests setting user ID in context
func TestSetUserID(t *testing.T) {
	ctx := context.Background()
	userID := uint(456)

	newCtx := httputil.SetUserID(ctx, userID)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req = req.WithContext(newCtx)

	extractedID, ok := httputil.GetUserID(req)
	if !ok {
		t.Error("Expected to retrieve user ID from context")
	}

	if extractedID != userID {
		t.Errorf("Expected userID=%d, got %d", userID, extractedID)
	}
}

// TestExtractIDFromPath tests path ID extraction
func TestExtractIDFromPath(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		idPosition int
		expectedID int
		expectErr  bool
	}{
		{
			name:       "valid path with ID",
			path:       "/recipes/123",
			idPosition: 1,
			expectedID: 123,
			expectErr:  false,
		},
		{
			name:       "invalid ID format",
			path:       "/recipes/abc",
			idPosition: 1,
			expectedID: 0,
			expectErr:  true,
		},
		{
			name:       "missing ID segment",
			path:       "/recipes",
			idPosition: 1,
			expectedID: 0,
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			id, err := httputil.ExtractIDFromPath(req, tt.idPosition)

			if tt.expectErr && err == nil {
				t.Error("Expected error, got nil")
			}

			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectErr && id != tt.expectedID {
				t.Errorf("Expected ID=%d, got %d", tt.expectedID, id)
			}
		})
	}
}

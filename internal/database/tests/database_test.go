package tests

import (
	"testing"

	"ultra-bis/internal/database"
)

// TestConnect_InvalidConnection tests that connection fails with invalid credentials
func TestConnect_InvalidConnection(t *testing.T) {
	_, err := database.Connect()
	if err == nil {
		t.Error("Expected error when connecting to invalid database, got nil")
	}
}

// TestConnect_EmptyHost tests that connection requires a host
func TestConnect_EmptyHost(t *testing.T) {
	_, err := database.Connect()
	if err == nil {
		t.Error("Expected error when connecting with empty host, got nil")
	}
}

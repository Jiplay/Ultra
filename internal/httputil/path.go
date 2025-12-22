package httputil

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// ExtractIDFromPath extracts an integer ID from a URL path
// Example: "/recipes/123" or "/recipes/123/ingredients" -> 123
// The idPosition parameter indicates which segment contains the ID (0-indexed after splitting)
func ExtractIDFromPath(r *http.Request, idPosition int) (int, error) {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	if len(pathParts) <= idPosition {
		return 0, fmt.Errorf("ID not found in path")
	}

	id, err := strconv.Atoi(pathParts[idPosition])
	if err != nil {
		return 0, fmt.Errorf("invalid ID format: %w", err)
	}

	return id, nil
}

// ExtractTwoIDsFromPath extracts two integer IDs from a URL path
// Example: "/recipes/123/ingredients/456" -> (123, 456)
func ExtractTwoIDsFromPath(r *http.Request, firstIDPosition, secondIDPosition int) (int, int, error) {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	if len(pathParts) <= firstIDPosition || len(pathParts) <= secondIDPosition {
		return 0, 0, fmt.Errorf("IDs not found in path")
	}

	firstID, err := strconv.Atoi(pathParts[firstIDPosition])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid first ID format: %w", err)
	}

	secondID, err := strconv.Atoi(pathParts[secondIDPosition])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid second ID format: %w", err)
	}

	return firstID, secondID, nil
}

package httputil

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// WriteJSON writes a JSON response with the given status code
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// WriteError writes a JSON error response
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, ErrorResponse{Error: message})
}

// SuccessResponse represents a standard success message response
type SuccessResponse struct {
	Message string `json:"message"`
}

// WriteSuccess writes a JSON success message response
func WriteSuccess(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, SuccessResponse{Message: message})
}

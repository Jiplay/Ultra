package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"ultra-bis/internal/user"
)

// Handler handles authentication requests
type Handler struct {
	userRepo *user.Repository
}

// NewHandler creates a new auth handler
func NewHandler(userRepo *user.Repository) *Handler {
	return &Handler{userRepo: userRepo}
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}

// Register handles user registration
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req user.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validation
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	if len(req.Password) < 6 {
		writeError(w, http.StatusBadRequest, "Password must be at least 6 characters")
		return
	}

	// Check if email already exists
	exists, err := h.userRepo.EmailExists(req.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to check email")
		return
	}
	if exists {
		writeError(w, http.StatusConflict, "Email already registered")
		return
	}

	// Create user
	newUser := &user.User{
		Email: req.Email,
		Name:  req.Name,
	}

	if err := newUser.HashPassword(req.Password); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	if err := h.userRepo.Create(newUser); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Generate token
	token, err := GenerateToken(newUser.ID, newUser.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	writeJSON(w, http.StatusCreated, user.LoginResponse{
		Token: token,
		User:  *newUser,
	})
}

// Login handles user login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req user.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get user by email
	foundUser, err := h.userRepo.GetByEmail(req.Email)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Check password
	if !foundUser.CheckPassword(req.Password) {
		writeError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Generate token
	token, err := GenerateToken(foundUser.ID, foundUser.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	writeJSON(w, http.StatusOK, user.LoginResponse{
		Token: token,
		User:  *foundUser,
	})
}

// GetMe returns the current authenticated user
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract user ID from context (set by middleware)
	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	foundUser, err := h.userRepo.GetByID(userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "User not found")
		return
	}

	writeJSON(w, http.StatusOK, foundUser)
}

// UpdateProfile updates the user's profile
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract user ID from context
	userID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req user.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get current user
	foundUser, err := h.userRepo.GetByID(userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "User not found")
		return
	}

	// Update fields
	if req.Name != "" {
		foundUser.Name = req.Name
	}
	if req.Age > 0 {
		foundUser.Age = req.Age
	}
	if req.Gender != "" {
		foundUser.Gender = req.Gender
	}
	if req.Height > 0 {
		foundUser.Height = req.Height
	}
	if req.ActivityLevel != "" {
		foundUser.ActivityLevel = req.ActivityLevel
	}
	if req.GoalType != "" {
		foundUser.GoalType = req.GoalType
	}

	if err := h.userRepo.Update(foundUser); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	writeJSON(w, http.StatusOK, foundUser)
}

// ExtractTokenFromHeader extracts the token from Authorization header
func ExtractTokenFromHeader(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	if bearerToken == "" {
		return ""
	}

	parts := strings.Split(bearerToken, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

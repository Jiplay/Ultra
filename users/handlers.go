package users

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Handler struct {
	controller *Controller
}

func NewHandler(controller *Controller) *Handler {
	return &Handler{controller: controller}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	user, err := h.controller.CreateUser(&req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "user already exists" {
			statusCode = http.StatusConflict
		} else if err.Error() == "email is required" || err.Error() == "name is required" {
			statusCode = http.StatusBadRequest
		}
		http.Error(w, err.Error(), statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := h.extractIDFromPath(r.URL.Path)
	if userID == "" {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.controller.GetUserByID(userID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "invalid user ID" {
			statusCode = http.StatusBadRequest
		} else if err.Error() == "user not found" {
			statusCode = http.StatusNotFound
		}
		http.Error(w, err.Error(), statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := h.extractIDFromPath(r.URL.Path)
	if userID == "" {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	user, err := h.controller.UpdateProfile(userID, &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "invalid user ID" || err.Error() == "no fields to update" {
			statusCode = http.StatusBadRequest
		} else if err.Error() == "user not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "email already in use" {
			statusCode = http.StatusConflict
		}
		http.Error(w, err.Error(), statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := h.extractIDFromPath(r.URL.Path)
	if userID == "" {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	err := h.controller.DeleteUser(userID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "invalid user ID" {
			statusCode = http.StatusBadRequest
		} else if err.Error() == "user not found" {
			statusCode = http.StatusNotFound
		}
		http.Error(w, err.Error(), statusCode)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) extractIDFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i, part := range parts {
		if part == "users" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}


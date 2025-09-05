package programs

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type Handlers struct {
	controller *Controller
}

func NewHandlers(controller *Controller) *Handlers {
	return &Handlers{controller: controller}
}

func (h *Handlers) CreateProgram(w http.ResponseWriter, r *http.Request) {
	var req CreateProgramRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")

	program, err := h.controller.CreateProgram(&req, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(program)
}

func (h *Handlers) GetPrograms(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")

	programs, err := h.controller.GetProgramsByUserID(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(programs)
}

func (h *Handlers) GetProgram(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/programs/")
	programID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid program ID", http.StatusBadRequest)
		return
	}

	userID := r.URL.Query().Get("user_id")

	program, err := h.controller.GetProgramByID(programID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Program not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(program)
}

func (h *Handlers) UpdateProgram(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/programs/")
	programID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid program ID", http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")

	var req UpdateProgramRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err = h.controller.UpdateProgram(programID, userID, &req)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Program not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) DeleteProgram(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/programs/")
	programID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid program ID", http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")

	err = h.controller.DeleteProgram(programID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Program not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete program", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
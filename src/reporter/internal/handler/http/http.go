package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	food "ultra.com/food/pkg/model"
	"ultra.com/reporter/internal/controller/health"
	gateway "ultra.com/reporter/internal/gateway/sport"
	"ultra.com/sport/pkg/model"
)

type Handler struct {
	ctrl *health.Controller
}

func New(ctrl *health.Controller) *Handler { return &Handler{ctrl: ctrl} }

func (h *Handler) GetRecipe(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	recipe, err := h.ctrl.GetRecipe(r.Context(), food.RecipeID(id))
	if err != nil && errors.Is(err, gateway.ErrNotFound) {
		http.Error(w, err.Error(), http.StatusNotFound)
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(recipe); err != nil {
		log.Printf("Encode error %v\n", err)
	}
}

func (h *Handler) GetWorkout(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	workout, err := h.ctrl.GetWorkout(r.Context(), model.WorkoutPlanID(id))
	if err != nil && errors.Is(err, gateway.ErrNotFound) {
		http.Error(w, err.Error(), http.StatusNotFound)
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	if err := json.NewEncoder(w).Encode(workout); err != nil {
		log.Printf("Encode error %v\n", err)
	}
}

func (h *Handler) GetPerformance(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	perf, err := h.ctrl.GetPerformance(r.Context(), model.WorkoutPerformanceID(id))
	if err != nil && errors.Is(err, gateway.ErrNotFound) {
		http.Error(w, err.Error(), http.StatusNotFound)
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	if err := json.NewEncoder(w).Encode(perf); err != nil {
		log.Printf("Encode error %v\n", err)
	}
}

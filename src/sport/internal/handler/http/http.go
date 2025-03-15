package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"ultra.com/sport/internal/controller/sport"
	"ultra.com/sport/internal/repository"
	"ultra.com/sport/pkg/model"
)

type Handler struct {
	ctrl *sport.Controller
}

func New(ctrl *sport.Controller) *Handler { return &Handler{ctrl} }

func (h *Handler) GetWorkoutPlans(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
	}
	ctx := r.Context()
	p, err := h.ctrl.GetPlan(ctx, model.WorkoutPlanID(id))
	if err != nil && errors.Is(err, repository.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
	} else if err != nil {
		log.Printf("Repository get error %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(p); err != nil {
		log.Printf("Response encode error %v\n", err)
	}
}

func (h *Handler) GetPerformances(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
	}
	ctx := r.Context()
	perf, err := h.ctrl.GetPerformance(ctx, model.WorkoutPerformanceID(id))
	if err != nil && errors.Is(err, repository.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
	} else if err != nil {
		log.Printf("Repository get error %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(perf); err != nil {
		log.Printf("Response encode error %v\n", err)
	}
}

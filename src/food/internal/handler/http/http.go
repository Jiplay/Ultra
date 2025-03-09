package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"ultra.com/food/internal/controller/food"
	"ultra.com/food/pkg/model"
)

type Handler struct {
	ctrl *food.Controller
}

func New(ctrl *food.Controller) *Handler { return &Handler{ctrl} }

func (h *Handler) GetRecipe(w http.ResponseWriter, r *http.Request) {
	log.Printf("New req")
	id := r.FormValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	data, err := h.ctrl.Get(ctx, model.RecipeID(id))
	if err != nil && errors.Is(err, food.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Response encode error %v\n", err)
	}
}

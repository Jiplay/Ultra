package grpc

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"ultra.com/food/internal/controller/food"
	"ultra.com/food/pkg/model"
	"ultra.com/gen"
)

type Handler struct {
	gen.UnimplementedFoodServiceServer
	ctrl *food.Controller
}

func New(ctrl *food.Controller) *Handler { return &Handler{ctrl: ctrl} }

func (h *Handler) GetRecipe(ctx context.Context, req *gen.GetRecipeRequest) (*gen.GetRecipeResponse, error) {
	if req == nil || req.RecipeId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "missing recipe_id")
	}
	r, err := h.ctrl.Get(ctx, model.RecipeID(req.RecipeId))
	if err != nil && status.Code(err) != codes.NotFound {
		return nil, status.Errorf(codes.NotFound, "recipe not found")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "internal server error")
	}
	return &gen.GetRecipeResponse{Recipe: model.RecipeToProto(r)}, nil
}

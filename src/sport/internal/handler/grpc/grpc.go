package grpc

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"ultra.com/gen"
	"ultra.com/sport/internal/controller/sport"
	"ultra.com/sport/pkg/model"
)

type Handler struct {
	gen.UnimplementedSportServiceServer
	ctrl *sport.Controller
}

func New(ctrl *sport.Controller) *Handler { return &Handler{ctrl: ctrl} }

func (h *Handler) GetWorkout(ctx context.Context, req *gen.GetWorkoutRequest) (*gen.GetWorkoutResponse, error) {
	if req == nil || req.WorkoutId == "" {
		return nil, status.Error(codes.InvalidArgument, "workout id is required")
	}
	r, err := h.ctrl.GetWorkout(ctx, model.WorkoutID(req.WorkoutId))
	if err != nil && status.Code(err) != codes.NotFound {
		return nil, status.Error(codes.NotFound, "workout not found")
	} else if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}
	return &gen.GetWorkoutResponse{Workout: model.WorkoutToProto(r)}, nil
}

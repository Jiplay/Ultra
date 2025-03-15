package http

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"ultra.com/pkg/discovery"
	"ultra.com/reporter/internal/gateway"
	"ultra.com/sport/pkg/model"
)

// Gateway defines a HTTP gateway for a sport service
type Gateway struct {
	registry discovery.Registry
}

// New return a new Sport gateway
func New(registry discovery.Registry) *Gateway { return &Gateway{registry} }

// GetWorkout returns info linked a workoutID
func (g *Gateway) GetWorkout(ctx context.Context, id model.WorkoutID) (model.Workout, error) {
	addresses, err := g.registry.ServiceAddresses(ctx, "sport")
	if err != nil {
		return model.Workout{}, err
	}
	url := "http://" + addresses[rand.Intn(len(addresses))] + "/workout"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return model.Workout{}, err
	}
	req = req.WithContext(ctx)
	values := req.URL.Query()
	values.Add("id", string(id))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return model.Workout{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return model.Workout{}, gateway.ErrNotFound
	} else if resp.StatusCode/100 != 2 {
		return model.Workout{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var v model.Workout
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return model.Workout{}, err
	}
	return v, nil
}

func (g *Gateway) GetPerformance(ctx context.Context, id model.WorkoutPerformanceID) (model.WorkoutPerformance, error) {
	addresses, err := g.registry.ServiceAddresses(ctx, "sport")
	if err != nil {
		return model.WorkoutPerformance{}, err
	}
	url := "http://" + addresses[rand.Intn(len(addresses))] + "/performance"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return model.WorkoutPerformance{}, err
	}
	req.WithContext(ctx)
	values := req.URL.Query()
	values.Add("id", string(id))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return model.WorkoutPerformance{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return model.WorkoutPerformance{}, gateway.ErrNotFound
	} else if resp.StatusCode/100 != 2 {
		return model.WorkoutPerformance{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var v model.WorkoutPerformance
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return model.WorkoutPerformance{}, err
	}
	return v, nil
}

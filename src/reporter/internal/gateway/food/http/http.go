package http

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"ultra.com/food/pkg/model"
	"ultra.com/pkg/discovery"
	"ultra.com/reporter/internal/gateway"
)

type Gateway struct {
	registry discovery.Registry
}

func New(registry discovery.Registry) *Gateway { return &Gateway{registry} }

func (g *Gateway) GetRecipe(ctx context.Context, id model.RecipeID) (model.Recipe, error) {
	addresses, err := g.registry.ServiceAddresses(ctx, "food")
	if err != nil {
		return model.Recipe{}, err
	}
	url := "http://" + addresses[rand.Intn(len(addresses))] + "/recipe"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return model.Recipe{}, err
	}
	req = req.WithContext(ctx)
	values := req.URL.Query()
	values.Add("id", string(id))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return model.Recipe{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		return model.Recipe{}, gateway.ErrNotFound
	} else if resp.StatusCode/100 != 2 {
		return model.Recipe{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var v model.Recipe
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return model.Recipe{}, err
	}
	return v, nil
}

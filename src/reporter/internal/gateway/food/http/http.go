package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"ultra.com/food/pkg/model"
	gateway "ultra.com/reporter/internal/gateway/sport"
)

type Gateway struct {
	address string
}

func New(address string) *Gateway { return &Gateway{address: address} }

func (g *Gateway) GetRecipe(ctx context.Context, id model.RecipeID) (model.Recipe, error) {
	req, err := http.NewRequest("GET", g.address+"/recipe", nil)
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

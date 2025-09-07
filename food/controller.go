package food

import (
	"errors"
)

type Controller struct {
	repo Repository
}

func NewController(repo Repository) *Controller {
	return &Controller{repo: repo}
}

func (c *Controller) CreateFood(req *CreateFoodRequest) (*Food, error) {
	if err := c.validateCreateRequest(req); err != nil {
		return nil, err
	}

	food := &Food{
		Name:     req.Name,
		Calories: req.Calories,
		Protein:  req.Protein,
		Carbs:    req.Carbs,
		Fat:      req.Fat,
	}

	return c.repo.Create(food)
}

func (c *Controller) GetFoodByID(id int) (*Food, error) {
	if id <= 0 {
		return nil, errors.New("invalid food ID")
	}
	return c.repo.GetByID(id)
}

func (c *Controller) GetAllFoods() ([]*Food, error) {
	return c.repo.GetAll()
}

func (c *Controller) UpdateFood(id int, req *UpdateFoodRequest) (*Food, error) {
	if id <= 0 {
		return nil, errors.New("invalid food ID")
	}

	if err := c.validateUpdateRequest(req); err != nil {
		return nil, err
	}

	return c.repo.Update(id, req)
}

func (c *Controller) DeleteFood(id int) error {
	if id <= 0 {
		return errors.New("invalid food ID")
	}
	return c.repo.Delete(id)
}

func (c *Controller) validateCreateRequest(req *CreateFoodRequest) error {
	if req.Name == "" {
		return errors.New("food name is required")
	}
	if req.Calories < 0 {
		return errors.New("calories cannot be negative")
	}
	if req.Protein < 0 {
		return errors.New("protein cannot be negative")
	}
	if req.Carbs < 0 {
		return errors.New("carbs cannot be negative")
	}
	if req.Fat < 0 {
		return errors.New("fat cannot be negative")
	}
	return nil
}

func (c *Controller) validateUpdateRequest(req *UpdateFoodRequest) error {
	if req.Name != nil && *req.Name == "" {
		return errors.New("food name cannot be empty")
	}
	if req.Category != nil && *req.Category == "" {
		return errors.New("food category cannot be empty")
	}
	if req.Calories != nil && *req.Calories < 0 {
		return errors.New("calories cannot be negative")
	}
	if req.Protein != nil && *req.Protein < 0 {
		return errors.New("protein cannot be negative")
	}
	if req.Carbs != nil && *req.Carbs < 0 {
		return errors.New("carbs cannot be negative")
	}
	if req.Fat != nil && *req.Fat < 0 {
		return errors.New("fat cannot be negative")
	}
	return nil
}

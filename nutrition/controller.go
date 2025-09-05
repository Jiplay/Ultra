package nutrition

import (
	"database/sql"
	"errors"
)

type Controller struct {
	repo Repository
}

func NewController(repo Repository) *Controller {
	return &Controller{repo: repo}
}

func (c *Controller) CreateFood(req *CreateFoodRequest, userID string) (*Food, error) {
	if req.Name == "" {
		return nil, errors.New("food name is required")
	}
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	food := &Food{
		Name:            req.Name,
		CaloriesPer100g: req.CaloriesPer100g,
		ProteinPer100g:  req.ProteinPer100g,
		CarbsPer100g:    req.CarbsPer100g,
		FatPer100g:      req.FatPer100g,
		FiberPer100g:    req.FiberPer100g,
		CreatedBy:       userID,
	}

	err := c.repo.CreateFood(food)
	if err != nil {
		return nil, err
	}

	return food, nil
}

func (c *Controller) GetFoods(search string) ([]Food, error) {
	return c.repo.GetFoods(search)
}

func (c *Controller) GetFoodByID(id int) (*Food, error) {
	return c.repo.GetFoodByID(id)
}

func (c *Controller) CreateRecipe(req *CreateRecipeRequest, userID string) (*Recipe, error) {
	if req.Name == "" {
		return nil, errors.New("recipe name is required")
	}
	if userID == "" {
		return nil, errors.New("user ID is required")
	}
	if req.Servings <= 0 {
		return nil, errors.New("servings must be greater than 0")
	}

	recipe := &Recipe{
		UserID:       userID,
		Name:         req.Name,
		Description:  req.Description,
		Instructions: req.Instructions,
		Servings:     req.Servings,
		PrepTime:     req.PrepTime,
		CookTime:     req.CookTime,
		Ingredients:  make([]RecipeIngredient, len(req.Ingredients)),
	}

	for i, ingredientReq := range req.Ingredients {
		if ingredientReq.FoodID <= 0 {
			return nil, errors.New("invalid food ID")
		}
		if ingredientReq.Quantity <= 0 {
			return nil, errors.New("ingredient quantity must be greater than 0")
		}
		if ingredientReq.Unit == "" {
			return nil, errors.New("ingredient unit is required")
		}

		recipe.Ingredients[i] = RecipeIngredient{
			FoodID:   ingredientReq.FoodID,
			Quantity: ingredientReq.Quantity,
			Unit:     ingredientReq.Unit,
		}
	}

	err := c.repo.CreateRecipe(recipe)
	if err != nil {
		return nil, err
	}

	return recipe, nil
}

func (c *Controller) GetRecipesByUserID(userID string) ([]Recipe, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}
	return c.repo.GetRecipesByUserID(userID)
}

func (c *Controller) UpdateNutritionGoals(userID string, req *UpdateNutritionGoalsRequest) (*Goals, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	if err := c.validateNutritionGoalsRequest(req); err != nil {
		return nil, err
	}

	_, err := c.repo.CheckNutritionGoalsExist(userID)
	if err == sql.ErrNoRows {
		goals := &Goals{
			UserID:        userID,
			DailyCalories: getIntValue(req.DailyCalories, 2000),
			DailyProtein:  getFloatValue(req.DailyProtein, 150.0),
			DailyCarbs:    getFloatValue(req.DailyCarbs, 250.0),
			DailyFat:      getFloatValue(req.DailyFat, 67.0),
			DailyFiber:    getFloatValue(req.DailyFiber, 25.0),
		}

		err = c.repo.CreateNutritionGoals(goals)
		if err != nil {
			return nil, err
		}
		return goals, nil
	} else if err != nil {
		return nil, err
	}

	err = c.repo.UpdateNutritionGoals(userID, req)
	if err != nil {
		return nil, err
	}

	return c.repo.GetNutritionGoalsByUserID(userID)
}

func (c *Controller) GetNutritionGoalsByUserID(userID string) (*Goals, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}
	return c.repo.GetNutritionGoalsByUserID(userID)
}

func (c *Controller) validateNutritionGoalsRequest(req *UpdateNutritionGoalsRequest) error {
	if req.DailyCalories != nil && *req.DailyCalories < 0 {
		return errors.New("daily calories cannot be negative")
	}
	if req.DailyProtein != nil && *req.DailyProtein < 0 {
		return errors.New("daily protein cannot be negative")
	}
	if req.DailyCarbs != nil && *req.DailyCarbs < 0 {
		return errors.New("daily carbs cannot be negative")
	}
	if req.DailyFat != nil && *req.DailyFat < 0 {
		return errors.New("daily fat cannot be negative")
	}
	if req.DailyFiber != nil && *req.DailyFiber < 0 {
		return errors.New("daily fiber cannot be negative")
	}
	return nil
}

func getIntValue(ptr *int, defaultValue int) int {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

func getFloatValue(ptr *float64, defaultValue float64) float64 {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

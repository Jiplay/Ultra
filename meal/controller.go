package meal

import (
	"database/sql"
	"errors"
	"time"
)

type Controller struct {
	repo Repository
}

func NewController(repo Repository) *Controller {
	return &Controller{repo: repo}
}

func (c *Controller) CreateMeal(req *CreateMealRequest, userID string) (*Meal, error) {
	if err := c.validateCreateMealRequest(req, userID); err != nil {
		return nil, err
	}

	meal := &Meal{
		UserID:    userID,
		Name:      req.Name,
		MealType:  req.MealType,
		Date:      req.Date,
		Notes:     req.Notes,
		MealItems: make([]MealItem, len(req.Items)),
	}

	// Validate and convert meal items
	for i, itemReq := range req.Items {
		if err := c.validateMealItem(&itemReq); err != nil {
			return nil, err
		}

		meal.MealItems[i] = MealItem{
			ItemType: itemReq.ItemType,
			ItemID:   itemReq.ItemID,
			Quantity: itemReq.Quantity,
			Notes:    itemReq.Notes,
		}
	}

	err := c.repo.CreateMeal(meal)
	if err != nil {
		return nil, err
	}

	return meal, nil
}

func (c *Controller) GetMealsByUserID(userID string, startDate, endDate *time.Time) ([]Meal, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	var start, end time.Time
	if startDate != nil {
		start = *startDate
	}
	if endDate != nil {
		end = *endDate
	}

	return c.repo.GetMealsByUserID(userID, start, end)
}

func (c *Controller) GetMealByID(id int, userID string) (*Meal, error) {
	if err := c.validateMealAccess(id, userID); err != nil {
		return nil, err
	}

	return c.repo.GetMealByID(id, userID)
}

func (c *Controller) UpdateMeal(id int, userID string, req *UpdateMealRequest) error {
	if err := c.validateMealAccess(id, userID); err != nil {
		return err
	}

	if err := c.validateUpdateMealRequest(req); err != nil {
		return err
	}

	exists, err := c.repo.CheckMealExists(id, userID)
	if err != nil {
		return err
	}
	if !exists {
		return sql.ErrNoRows
	}

	return c.repo.UpdateMeal(id, userID, req)
}

func (c *Controller) DeleteMeal(id int, userID string) error {
	if err := c.validateMealAccess(id, userID); err != nil {
		return err
	}

	return c.repo.DeleteMeal(id, userID)
}

func (c *Controller) AddMealItem(mealID int, userID string, req *AddMealItemRequest) (*MealItem, error) {
	if err := c.validateMealAccess(mealID, userID); err != nil {
		return nil, err
	}

	itemReq := CreateMealItemRequest{
		ItemType: req.ItemType,
		ItemID:   req.ItemID,
		Quantity: req.Quantity,
		Notes:    req.Notes,
	}

	if err := c.validateMealItem(&itemReq); err != nil {
		return nil, err
	}

	// Check if meal exists
	exists, err := c.repo.CheckMealExists(mealID, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("meal not found")
	}

	item := &MealItem{
		ItemType: req.ItemType,
		ItemID:   req.ItemID,
		Quantity: req.Quantity,
		Notes:    req.Notes,
	}

	err = c.repo.AddMealItem(mealID, item)
	if err != nil {
		return nil, err
	}

	return item, nil
}

func (c *Controller) UpdateMealItem(itemID int, mealID int, userID string, req *UpdateMealItemRequest) error {
	if err := c.validateMealAccess(mealID, userID); err != nil {
		return err
	}

	if err := c.validateUpdateMealItemRequest(req); err != nil {
		return err
	}

	return c.repo.UpdateMealItem(itemID, req)
}

func (c *Controller) DeleteMealItem(itemID int, mealID int, userID string) error {
	if err := c.validateMealAccess(mealID, userID); err != nil {
		return err
	}

	return c.repo.DeleteMealItem(itemID, mealID)
}

func (c *Controller) GetMealSummary(userID string, date time.Time) (*MealSummary, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	return c.repo.GetMealSummary(userID, date)
}

func (c *Controller) GetMealPlan(userID string, startDate, endDate time.Time) (*MealPlan, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	if startDate.IsZero() || endDate.IsZero() {
		return nil, errors.New("start date and end date are required")
	}

	if endDate.Before(startDate) {
		return nil, errors.New("end date cannot be before start date")
	}

	// Limit the date range to prevent excessive data
	maxDays := 31
	if endDate.Sub(startDate) > time.Duration(maxDays)*24*time.Hour {
		return nil, errors.New("date range cannot exceed 31 days")
	}

	return c.repo.GetMealPlan(userID, startDate, endDate)
}

func (c *Controller) GetDailyMeals(userID string, date time.Time) ([]Meal, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	return c.repo.GetMealsByUserID(userID, startOfDay, endOfDay)
}

// Validation functions
func (c *Controller) validateCreateMealRequest(req *CreateMealRequest, userID string) error {
	if userID == "" {
		return errors.New("user ID is required")
	}

	if req.Name == "" {
		return errors.New("meal name is required")
	}

	if err := c.validateMealType(req.MealType); err != nil {
		return err
	}

	if req.Date.IsZero() {
		return errors.New("meal date is required")
	}

	// Don't allow meals too far in the future
	if req.Date.After(time.Now().AddDate(0, 1, 0)) {
		return errors.New("meal date cannot be more than 1 month in the future")
	}

	// Don't allow meals too far in the past (for data integrity)
	if req.Date.Before(time.Now().AddDate(-1, 0, 0)) {
		return errors.New("meal date cannot be more than 1 year in the past")
	}

	for i, item := range req.Items {
		if err := c.validateMealItem(&item); err != nil {
			return errors.New("item " + string(rune(i+1)) + ": " + err.Error())
		}
	}

	return nil
}

func (c *Controller) validateUpdateMealRequest(req *UpdateMealRequest) error {
	if req.Name != nil && *req.Name == "" {
		return errors.New("meal name cannot be empty")
	}

	if req.MealType != nil {
		if err := c.validateMealType(*req.MealType); err != nil {
			return err
		}
	}

	if req.Date != nil {
		if req.Date.IsZero() {
			return errors.New("meal date cannot be empty")
		}

		if req.Date.After(time.Now().AddDate(0, 1, 0)) {
			return errors.New("meal date cannot be more than 1 month in the future")
		}

		if req.Date.Before(time.Now().AddDate(-1, 0, 0)) {
			return errors.New("meal date cannot be more than 1 year in the past")
		}
	}

	return nil
}

func (c *Controller) validateMealItem(item *CreateMealItemRequest) error {
	if item.ItemType != "food" && item.ItemType != "recipe" {
		return errors.New("item type must be 'food' or 'recipe'")
	}

	if item.ItemID <= 0 {
		return errors.New("invalid item ID")
	}

	if item.Quantity <= 0 {
		return errors.New("item quantity must be greater than 0")
	}

	// Set reasonable limits
	maxQuantity := 10000.0 // 10kg for food, 100 servings for recipe
	if item.Quantity > maxQuantity {
		return errors.New("item quantity is too large")
	}

	// Verify the item exists
	if item.ItemType == "food" {
		_, err := c.repo.GetFoodByID(item.ItemID)
		if err != nil {
			if err == sql.ErrNoRows {
				return errors.New("food not found")
			}
			return errors.New("failed to validate food")
		}
	} else if item.ItemType == "recipe" {
		_, err := c.repo.GetRecipeByID(item.ItemID)
		if err != nil {
			if err == sql.ErrNoRows {
				return errors.New("recipe not found")
			}
			return errors.New("failed to validate recipe")
		}
	}

	return nil
}

func (c *Controller) validateUpdateMealItemRequest(req *UpdateMealItemRequest) error {
	if req.Quantity != nil {
		if *req.Quantity <= 0 {
			return errors.New("item quantity must be greater than 0")
		}

		maxQuantity := 10000.0
		if *req.Quantity > maxQuantity {
			return errors.New("item quantity is too large")
		}
	}

	return nil
}

func (c *Controller) validateMealType(mealType MealType) error {
	switch mealType {
	case MealTypeBreakfast, MealTypeLunch, MealTypeDinner, MealTypeSnack:
		return nil
	default:
		return errors.New("invalid meal type. Must be one of: breakfast, lunch, dinner, snack")
	}
}

func (c *Controller) validateMealAccess(mealID int, userID string) error {
	if userID == "" {
		return errors.New("user ID is required")
	}

	if mealID <= 0 {
		return errors.New("invalid meal ID")
	}

	return nil
}

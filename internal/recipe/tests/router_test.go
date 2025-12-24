package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"ultra-bis/internal/recipe"

	"ultra-bis/internal/food"
	"ultra-bis/test/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupRouterTest creates a test environment with database, handler, and mux
func setupRouterTest(t *testing.T) (*gorm.DB, *http.ServeMux, *food.Repository, uint) {
	t.Helper()

	// Setup test database
	db := testutil.SetupTestDB(t)

	// Run migrations
	if err := db.AutoMigrate(&food.Food{}, &recipe.Recipe{}, &recipe.RecipeIngredient{}); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Create repositories and service
	recipeRepo := recipe.NewRepository(db)
	foodRepo := food.NewRepository(db)
	foodAdapter := recipe.NewFoodAdapter(foodRepo)
	recipeService := recipe.NewService(recipeRepo, foodAdapter, db)
	handler := recipe.NewHandler(recipeService)

	// Setup HTTP mux and register routes
	mux := http.NewServeMux()
	recipe.RegisterRoutes(mux, handler)

	// Create a test user ID
	userID := testutil.MockUserID()

	return db, mux, foodRepo, userID
}

// makeAuthenticatedRequest creates and executes an authenticated HTTP request with a real JWT token
func makeAuthenticatedRequest(t *testing.T, mux *http.ServeMux, method, path string, body interface{}, userID uint) *httptest.ResponseRecorder {
	t.Helper()

	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		require.NoError(t, err)
	}

	req := httptest.NewRequest(method, path, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Generate a real JWT token and add it to the Authorization header
	token := testutil.GenerateTestToken(t, userID, testutil.MockUserEmail())
	req.Header.Set("Authorization", "Bearer "+token)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	return rec
}

// makeUnauthenticatedRequest creates and executes an unauthenticated HTTP request
func makeUnauthenticatedRequest(t *testing.T, mux *http.ServeMux, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()

	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		require.NoError(t, err)
	}

	req := httptest.NewRequest(method, path, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	return rec
}

// TestRoutes_CreateRecipe tests POST /recipes
func TestRoutes_CreateRecipe(t *testing.T) {
	_, mux, foodRepo, userID := setupRouterTest(t)

	// Create test food
	chicken, _ := foodRepo.Create(food.CreateFoodRequest{
		Name:     "Chicken",
		Calories: 165,
		Protein:  31,
		Carbs:    0,
		Fat:      3.6,
		Fiber:    0,
	})

	tests := []struct {
		name           string
		request        recipe.CreateRecipeRequest
		expectedStatus int
		authenticated  bool
	}{
		{
			name: "successful creation with ingredients",
			request: recipe.CreateRecipeRequest{
				Name: "Chicken Bowl",
				Ingredients: []recipe.CreateIngredientRequest{
					{FoodID: chicken.ID, QuantityGrams: 200},
				},
			},
			expectedStatus: http.StatusCreated,
			authenticated:  true,
		},
		{
			name: "successful creation without ingredients",
			request: recipe.CreateRecipeRequest{
				Name: "Empty Recipe",
			},
			expectedStatus: http.StatusCreated,
			authenticated:  true,
		},
		{
			name: "unauthorized without auth",
			request: recipe.CreateRecipeRequest{
				Name: "Test Recipe",
			},
			expectedStatus: http.StatusUnauthorized,
			authenticated:  false,
		},
		{
			name: "invalid request body - empty name",
			request: recipe.CreateRecipeRequest{
				Name: "",
			},
			expectedStatus: http.StatusBadRequest,
			authenticated:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rec *httptest.ResponseRecorder
			if tt.authenticated {
				rec = makeAuthenticatedRequest(t, mux, http.MethodPost, "/recipes", tt.request, userID)
			} else {
				rec = makeUnauthenticatedRequest(t, mux, http.MethodPost, "/recipes", tt.request)
			}

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusCreated {
				var myrecipe recipe.Recipe
				err := json.NewDecoder(rec.Body).Decode(&myrecipe)
				require.NoError(t, err)
				assert.NotZero(t, myrecipe.ID)
				assert.Equal(t, tt.request.Name, myrecipe.Name)
			}
		})
	}
}

// TestRoutes_GetRecipe tests GET /recipes/{id}
func TestRoutes_GetRecipe(t *testing.T) {
	db, mux, foodRepo, userID := setupRouterTest(t)

	// Create test food
	chicken, _ := foodRepo.Create(food.CreateFoodRequest{
		Name:     "Chicken",
		Calories: 165,
		Protein:  31,
	})

	// Create a recipe
	myrecipe := &recipe.Recipe{
		Name:   "Test Recipe",
		UserID: &userID,
	}
	db.Create(myrecipe)

	// Add ingredient
	ingredient := &recipe.RecipeIngredient{
		RecipeID:      myrecipe.ID,
		FoodID:        chicken.ID,
		QuantityGrams: 200,
	}
	db.Create(ingredient)

	tests := []struct {
		name           string
		recipeID       string
		authenticated  bool
		expectedStatus int
	}{
		{
			name:           "successful retrieval",
			recipeID:       fmt.Sprintf("%d", myrecipe.ID),
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "not found",
			recipeID:       "99999",
			authenticated:  true,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "unauthorized",
			recipeID:       fmt.Sprintf("%d", myrecipe.ID),
			authenticated:  false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid ID",
			recipeID:       "invalid",
			authenticated:  true,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := fmt.Sprintf("/recipes/%s", tt.recipeID)
			var rec *httptest.ResponseRecorder

			if tt.authenticated {
				rec = makeAuthenticatedRequest(t, mux, http.MethodGet, path, nil, userID)
			} else {
				rec = makeUnauthenticatedRequest(t, mux, http.MethodGet, path, nil)
			}

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var result recipe.RecipeWithNutrition
				err := json.NewDecoder(rec.Body).Decode(&result)
				require.NoError(t, err)
				assert.Equal(t, myrecipe.Name, result.Name)
				assert.NotZero(t, result.TotalCalories)
			}
		})
	}
}

// TestRoutes_ListRecipes tests GET /recipes
func TestRoutes_ListRecipes(t *testing.T) {
	db, mux, _, userID := setupRouterTest(t)

	// Create user recipe
	userRecipe := &recipe.Recipe{
		Name:   "User Recipe",
		UserID: &userID,
	}
	db.Create(userRecipe)

	// Create global recipe
	globalRecipe := &recipe.Recipe{
		Name:   "Global Recipe",
		UserID: nil,
	}
	db.Create(globalRecipe)

	tests := []struct {
		name           string
		queryParam     string
		authenticated  bool
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "list all recipes (user + global)",
			queryParam:     "",
			authenticated:  true,
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "list only user recipes",
			queryParam:     "?user_only=true",
			authenticated:  true,
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "unauthorized",
			queryParam:     "",
			authenticated:  false,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/recipes" + tt.queryParam
			var rec *httptest.ResponseRecorder

			if tt.authenticated {
				rec = makeAuthenticatedRequest(t, mux, http.MethodGet, path, nil, userID)
			} else {
				rec = makeUnauthenticatedRequest(t, mux, http.MethodGet, path, nil)
			}

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var recipes []recipe.RecipeListResponse
				err := json.NewDecoder(rec.Body).Decode(&recipes)
				require.NoError(t, err)
				assert.Len(t, recipes, tt.expectedCount)
			}
		})
	}
}

// TestRoutes_UpdateRecipe tests PUT /recipes/{id}
func TestRoutes_UpdateRecipe(t *testing.T) {
	db, mux, _, userID := setupRouterTest(t)

	// Create user recipe
	userRecipe := &recipe.Recipe{
		Name:   "Original Name",
		UserID: &userID,
	}
	db.Create(userRecipe)

	// Create another user's recipe
	otherUserID := uint(999)
	otherRecipe := &recipe.Recipe{
		Name:   "Other User Recipe",
		UserID: &otherUserID,
	}
	db.Create(otherRecipe)

	tests := []struct {
		name           string
		recipeID       uint
		request        recipe.UpdateRecipeRequest
		authenticated  bool
		expectedStatus int
	}{
		{
			name:     "successful update",
			recipeID: userRecipe.ID,
			request: recipe.UpdateRecipeRequest{
				Name: "Updated Name",
			},
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:     "forbidden - other user's recipe",
			recipeID: otherRecipe.ID,
			request: recipe.UpdateRecipeRequest{
				Name: "Attempt Update",
			},
			authenticated:  true,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:     "unauthorized",
			recipeID: userRecipe.ID,
			request: recipe.UpdateRecipeRequest{
				Name: "New Name",
			},
			authenticated:  false,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := fmt.Sprintf("/recipes/%d", tt.recipeID)
			var rec *httptest.ResponseRecorder

			if tt.authenticated {
				rec = makeAuthenticatedRequest(t, mux, http.MethodPut, path, tt.request, userID)
			} else {
				rec = makeUnauthenticatedRequest(t, mux, http.MethodPut, path, tt.request)
			}

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var myrecipe recipe.Recipe
				err := json.NewDecoder(rec.Body).Decode(&myrecipe)
				require.NoError(t, err)
				assert.Equal(t, tt.request.Name, myrecipe.Name)
			}
		})
	}
}

// TestRoutes_DeleteRecipe tests DELETE /recipes/{id}
func TestRoutes_DeleteRecipe(t *testing.T) {
	db, mux, _, userID := setupRouterTest(t)

	// Create user recipe
	userRecipe := &recipe.Recipe{
		Name:   "To Delete",
		UserID: &userID,
	}
	db.Create(userRecipe)

	// Create another user's recipe
	otherUserID := uint(999)
	otherRecipe := &recipe.Recipe{
		Name:   "Other User Recipe",
		UserID: &otherUserID,
	}
	db.Create(otherRecipe)

	tests := []struct {
		name           string
		recipeID       uint
		authenticated  bool
		expectedStatus int
	}{
		{
			name:           "successful deletion",
			recipeID:       userRecipe.ID,
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "forbidden - other user's recipe",
			recipeID:       otherRecipe.ID,
			authenticated:  true,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "unauthorized",
			recipeID:       userRecipe.ID,
			authenticated:  false,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := fmt.Sprintf("/recipes/%d", tt.recipeID)
			var rec *httptest.ResponseRecorder

			if tt.authenticated {
				rec = makeAuthenticatedRequest(t, mux, http.MethodDelete, path, nil, userID)
			} else {
				rec = makeUnauthenticatedRequest(t, mux, http.MethodDelete, path, nil)
			}

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

// TestRoutes_AddIngredient tests POST /recipes/{id}/ingredients
func TestRoutes_AddIngredient(t *testing.T) {
	db, mux, foodRepo, userID := setupRouterTest(t)

	// Create test food
	chicken, _ := foodRepo.Create(food.CreateFoodRequest{
		Name:     "Chicken",
		Calories: 165,
		Protein:  31,
	})

	// Create user recipe
	userRecipe := &recipe.Recipe{
		Name:   "Test Recipe",
		UserID: &userID,
	}
	db.Create(userRecipe)

	// Create another user's recipe
	otherUserID := uint(999)
	otherRecipe := &recipe.Recipe{
		Name:   "Other Recipe",
		UserID: &otherUserID,
	}
	db.Create(otherRecipe)

	tests := []struct {
		name           string
		recipeID       uint
		request        recipe.AddIngredientRequest
		authenticated  bool
		expectedStatus int
	}{
		{
			name:     "successful addition",
			recipeID: userRecipe.ID,
			request: recipe.AddIngredientRequest{
				FoodID:        chicken.ID,
				QuantityGrams: 200,
			},
			authenticated:  true,
			expectedStatus: http.StatusCreated,
		},
		{
			name:     "forbidden - other user's recipe",
			recipeID: otherRecipe.ID,
			request: recipe.AddIngredientRequest{
				FoodID:        chicken.ID,
				QuantityGrams: 200,
			},
			authenticated:  true,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:     "unauthorized",
			recipeID: userRecipe.ID,
			request: recipe.AddIngredientRequest{
				FoodID:        chicken.ID,
				QuantityGrams: 200,
			},
			authenticated:  false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:     "invalid quantity",
			recipeID: userRecipe.ID,
			request: recipe.AddIngredientRequest{
				FoodID:        chicken.ID,
				QuantityGrams: -50,
			},
			authenticated:  true,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := fmt.Sprintf("/recipes/%d/ingredients", tt.recipeID)
			var rec *httptest.ResponseRecorder

			if tt.authenticated {
				rec = makeAuthenticatedRequest(t, mux, http.MethodPost, path, tt.request, userID)
			} else {
				rec = makeUnauthenticatedRequest(t, mux, http.MethodPost, path, tt.request)
			}

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusCreated {
				var ingredient recipe.RecipeIngredient
				err := json.NewDecoder(rec.Body).Decode(&ingredient)
				require.NoError(t, err)
				assert.NotZero(t, ingredient.ID)
				assert.Equal(t, tt.request.QuantityGrams, ingredient.QuantityGrams)
			}
		})
	}
}

// TestRoutes_UpdateIngredient tests PUT /recipes/{id}/ingredients/{ingredientId}
func TestRoutes_UpdateIngredient(t *testing.T) {
	db, mux, foodRepo, userID := setupRouterTest(t)

	// Create test food
	chicken, _ := foodRepo.Create(food.CreateFoodRequest{
		Name:     "Chicken",
		Calories: 165,
		Protein:  31,
	})

	// Create user recipe with ingredient
	userRecipe := &recipe.Recipe{
		Name:   "Test Recipe",
		UserID: &userID,
	}
	db.Create(userRecipe)

	ingredient := &recipe.RecipeIngredient{
		RecipeID:      userRecipe.ID,
		FoodID:        chicken.ID,
		QuantityGrams: 200,
	}
	db.Create(ingredient)

	// Create another user's recipe with ingredient
	otherUserID := uint(999)
	otherRecipe := &recipe.Recipe{
		Name:   "Other Recipe",
		UserID: &otherUserID,
	}
	db.Create(otherRecipe)

	otherIngredient := &recipe.RecipeIngredient{
		RecipeID:      otherRecipe.ID,
		FoodID:        chicken.ID,
		QuantityGrams: 100,
	}
	db.Create(otherIngredient)

	tests := []struct {
		name           string
		recipeID       uint
		ingredientID   uint
		request        recipe.UpdateIngredientRequest
		authenticated  bool
		expectedStatus int
	}{
		{
			name:         "successful update",
			recipeID:     userRecipe.ID,
			ingredientID: ingredient.ID,
			request: recipe.UpdateIngredientRequest{
				QuantityGrams: 300,
			},
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:         "forbidden - other user's recipe",
			recipeID:     otherRecipe.ID,
			ingredientID: otherIngredient.ID,
			request: recipe.UpdateIngredientRequest{
				QuantityGrams: 300,
			},
			authenticated:  true,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:         "unauthorized",
			recipeID:     userRecipe.ID,
			ingredientID: ingredient.ID,
			request: recipe.UpdateIngredientRequest{
				QuantityGrams: 300,
			},
			authenticated:  false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:         "invalid quantity",
			recipeID:     userRecipe.ID,
			ingredientID: ingredient.ID,
			request: recipe.UpdateIngredientRequest{
				QuantityGrams: -50,
			},
			authenticated:  true,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := fmt.Sprintf("/recipes/%d/ingredients/%d", tt.recipeID, tt.ingredientID)
			var rec *httptest.ResponseRecorder

			if tt.authenticated {
				rec = makeAuthenticatedRequest(t, mux, http.MethodPut, path, tt.request, userID)
			} else {
				rec = makeUnauthenticatedRequest(t, mux, http.MethodPut, path, tt.request)
			}

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var result recipe.RecipeIngredient
				err := json.NewDecoder(rec.Body).Decode(&result)
				require.NoError(t, err)
				assert.Equal(t, tt.request.QuantityGrams, result.QuantityGrams)
			}
		})
	}
}

// TestRoutes_DeleteIngredient tests DELETE /recipes/{id}/ingredients/{ingredientId}
func TestRoutes_DeleteIngredient(t *testing.T) {
	db, mux, foodRepo, userID := setupRouterTest(t)

	// Create test food
	chicken, _ := foodRepo.Create(food.CreateFoodRequest{
		Name:     "Chicken",
		Calories: 165,
		Protein:  31,
	})

	// Create user recipe with ingredient
	userRecipe := &recipe.Recipe{
		Name:   "Test Recipe",
		UserID: &userID,
	}
	db.Create(userRecipe)

	ingredient := &recipe.RecipeIngredient{
		RecipeID:      userRecipe.ID,
		FoodID:        chicken.ID,
		QuantityGrams: 200,
	}
	db.Create(ingredient)

	// Create another user's recipe with ingredient
	otherUserID := uint(999)
	otherRecipe := &recipe.Recipe{
		Name:   "Other Recipe",
		UserID: &otherUserID,
	}
	db.Create(otherRecipe)

	otherIngredient := &recipe.RecipeIngredient{
		RecipeID:      otherRecipe.ID,
		FoodID:        chicken.ID,
		QuantityGrams: 100,
	}
	db.Create(otherIngredient)

	tests := []struct {
		name           string
		recipeID       uint
		ingredientID   uint
		authenticated  bool
		expectedStatus int
	}{
		{
			name:           "successful deletion",
			recipeID:       userRecipe.ID,
			ingredientID:   ingredient.ID,
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "forbidden - other user's recipe",
			recipeID:       otherRecipe.ID,
			ingredientID:   otherIngredient.ID,
			authenticated:  true,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "unauthorized",
			recipeID:       userRecipe.ID,
			ingredientID:   ingredient.ID,
			authenticated:  false,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := fmt.Sprintf("/recipes/%d/ingredients/%d", tt.recipeID, tt.ingredientID)
			var rec *httptest.ResponseRecorder

			if tt.authenticated {
				rec = makeAuthenticatedRequest(t, mux, http.MethodDelete, path, nil, userID)
			} else {
				rec = makeUnauthenticatedRequest(t, mux, http.MethodDelete, path, nil)
			}

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

// TestRoutes_MethodNotAllowed tests that invalid HTTP methods return 405
func TestRoutes_MethodNotAllowed(t *testing.T) {
	_, mux, _, userID := setupRouterTest(t)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "PATCH /recipes not allowed",
			method: http.MethodPatch,
			path:   "/recipes",
		},
		{
			name:   "DELETE /recipes not allowed",
			method: http.MethodDelete,
			path:   "/recipes",
		},
		{
			name:   "POST /recipes/1 not allowed",
			method: http.MethodPost,
			path:   "/recipes/1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := makeAuthenticatedRequest(t, mux, tt.method, tt.path, nil, userID)
			assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
		})
	}
}

// TestRoutes_NotFound tests that invalid paths return 404
func TestRoutes_NotFound(t *testing.T) {
	_, mux, _, userID := setupRouterTest(t)

	tests := []struct {
		name string
		path string
	}{
		{
			name: "invalid nested path",
			path: "/recipes/1/foo",
		},
		{
			name: "too many path segments",
			path: "/recipes/1/ingredients/2/extra",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := makeAuthenticatedRequest(t, mux, http.MethodGet, tt.path, nil, userID)
			assert.Equal(t, http.StatusNotFound, rec.Code)
		})
	}
}

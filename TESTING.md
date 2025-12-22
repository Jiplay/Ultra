# Testing Policy & Guidelines

## Overview

This document defines the testing standards and practices for the Ultra-Bis project. All code contributions should follow these guidelines to ensure high-quality, maintainable tests.

## Testing Stack

### Core Tools
- **Testing Framework:** Go standard library `testing` package
- **Assertions:** `github.com/stretchr/testify` (assert, require)
- **Database Testing:** `testcontainers-go` with PostgreSQL containers
- **ORM:** GORM for database operations

### Test Types

#### 1. Unit Tests
- **Purpose:** Test individual functions/methods in isolation
- **Speed:** Fast (<10s for all unit tests)
- **Dependencies:** Mock external dependencies
- **Location:** `*_test.go` files alongside source code
- **Run Command:** `go test -short ./...` or `make test-unit`

#### 2. Integration Tests
- **Purpose:** Test component interactions with real database
- **Speed:** Moderate (30s-3min with test containers)
- **Dependencies:** Real PostgreSQL via testcontainers
- **Location:** `*_test.go` files with database setup
- **Run Command:** `go test ./...` or `make test`

#### 3. HTTP Handler Tests
- **Purpose:** Test HTTP endpoints end-to-end
- **Speed:** Moderate (includes request/response processing)
- **Dependencies:** Real or mocked repositories
- **Location:** `handler_test.go` files
- **Pattern:** Use `httptest.ResponseRecorder`

## Project Standards

### Test Coverage Targets

| Package Type | Minimum Coverage | Target Coverage |
|-------------|------------------|-----------------|
| Repository | 70% | 85% |
| Handler | 60% | 75% |
| Service/Business Logic | 75% | 90% |
| Models | N/A | N/A |
| Overall Project | 65% | 80% |

### Test Organization

#### File Naming
- Test files: `*_test.go` (same package as source)
- Example: `repository.go` → `repository_test.go`

#### Test Function Naming
Follow the pattern: `Test<Type>_<Method>_<Scenario>`

Examples:
```go
func TestRepository_Create(t *testing.T)              // Happy path
func TestRepository_Create_ValidationError(t *testing.T)  // Error case
func TestHandler_CreateGoal_Unauthorized(t *testing.T)    // HTTP error case
```

### Test Structure

#### Standard Test Pattern (AAA)

```go
func TestRepository_MethodName(t *testing.T) {
    // Arrange - Setup test data and dependencies
    db := testutil.SetupTestDB(t)
    repo := NewRepository(db)
    testData := createTestData()

    // Act - Execute the code under test
    result, err := repo.MethodName(testData)

    // Assert - Verify expectations
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, expected, result.Field)
}
```

#### Table-Driven Tests

Use for multiple similar scenarios:

```go
func TestRepository_Validation(t *testing.T) {
    tests := []struct {
        name    string
        input   Input
        wantErr bool
        errMsg  string
    }{
        {
            name:    "valid input",
            input:   validInput,
            wantErr: false,
        },
        {
            name:    "missing field",
            input:   invalidInput,
            wantErr: true,
            errMsg:  "field required",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateInput(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## Test Utilities

### Database Setup

Located in `test/testutil/database.go`:

```go
// SetupTestDB creates a fresh PostgreSQL container for testing
db := testutil.SetupTestDB(t)

// Don't forget to run migrations
db.AutoMigrate(&YourModel{})
```

The test container is automatically cleaned up when the test finishes.

### Authentication Helpers

Located in `test/testutil/auth.go`:

```go
// Generate JWT token for testing protected endpoints
token := testutil.GenerateTestToken(t, userID, email)

// Add auth header to request
testutil.AddAuthHeader(req, token)
```

### Custom Assertions

Located in `test/testutil/assertions.go`:

```go
// Check record exists
testutil.AssertRecordExists(t, db, &Model{}, "field = ?", value)

// Check record doesn't exist
testutil.AssertRecordNotExists(t, db, &Model{}, "field = ?", value)
```

### Test Factories

Located in `test/testutil/factories.go`:

```go
// Create test user with defaults
user := testutil.CreateTestUser(t, db)

// Create test food with defaults
food := testutil.CreateTestFood(t, db, overrides)

// Create test goal with defaults
goal := testutil.CreateTestGoal(t, db, userID, overrides)
```

## Repository Test Patterns

### Setup Helper

Each repository test file should have a setup helper:

```go
func setupGoalTest(t *testing.T) (*gorm.DB, *Repository) {
    t.Helper()
    db := testutil.SetupTestDB(t)

    // Migrate required models
    if err := db.AutoMigrate(&NutritionGoal{}, &user.User{}); err != nil {
        t.Fatalf("Failed to migrate database: %v", err)
    }

    return db, NewRepository(db)
}
```

### Required Test Cases

For each repository, implement tests for:

1. **Create** - Happy path, validation errors, database constraints
2. **GetByID** - Found, not found, wrong user access
3. **GetAll** - Empty list, multiple records, filtering
4. **Update** - Success, not found, validation errors
5. **Delete** - Success, not found, soft delete verification
6. **Special Methods** - Any custom queries or business logic

### Example Repository Test

```go
func TestRepository_Create(t *testing.T) {
    db, repo := setupGoalTest(t)

    // Create test user
    user := &user.User{Email: "test@example.com"}
    db.Create(user)

    goal := &NutritionGoal{
        UserID:   user.ID,
        Calories: 2000,
        Protein:  150,
        Carbs:    200,
        Fat:      65,
        Fiber:    30,
        StartDate: time.Now(),
        IsActive: true,
    }

    err := repo.Create(goal)

    assert.NoError(t, err)
    assert.NotZero(t, goal.ID)

    // Verify in database
    var found NutritionGoal
    db.First(&found, goal.ID)
    assert.Equal(t, goal.Calories, found.Calories)
}
```

## Handler Test Patterns

### HTTP Test Setup

```go
func TestHandler_CreateGoal(t *testing.T) {
    // Setup
    db := testutil.SetupTestDB(t)
    db.AutoMigrate(&NutritionGoal{}, &user.User{})

    repo := NewRepository(db)
    handler := NewHandler(repo, userRepo)

    // Create test user
    user := testutil.CreateTestUser(t, db)

    // Create request
    reqBody := `{"calories": 2000, "protein": 150}`
    req := httptest.NewRequest("POST", "/goals", strings.NewReader(reqBody))
    req.Header.Set("Content-Type", "application/json")

    // Add auth context
    ctx := context.WithValue(req.Context(), "user_id", user.ID)
    req = req.WithContext(ctx)

    // Execute
    rec := httptest.NewRecorder()
    handler.Create(rec, req)

    // Assert
    assert.Equal(t, http.StatusCreated, rec.Code)

    var response NutritionGoal
    json.Unmarshal(rec.Body.Bytes(), &response)
    assert.NotZero(t, response.ID)
}
```

### Required Handler Test Cases

For each HTTP handler, test:

1. **Success Cases** - Valid request with expected response
2. **Validation Errors** - Invalid input, missing fields
3. **Authentication** - Missing token, invalid token, expired token
4. **Authorization** - Access other users' resources
5. **Not Found** - Resource doesn't exist
6. **Conflict** - Duplicate resources, constraint violations

## Best Practices

### DO ✅

- **Use `t.Helper()`** in test utility functions
- **Use `require`** for critical assertions that should stop test execution
- **Use `assert`** for non-critical assertions
- **Use `t.Cleanup()`** for resource cleanup
- **Use descriptive test names** that explain the scenario
- **Test both success and failure paths**
- **Use table-driven tests** for similar scenarios
- **Mock external APIs** (don't call real APIs in tests)
- **Test edge cases** (nil, empty, zero values, max values)
- **Keep tests independent** (don't rely on test execution order)

### DON'T ❌

- **Don't share state** between tests
- **Don't use `time.Sleep()`** for synchronization
- **Don't test private methods** directly
- **Don't skip cleanup** (always use `t.Cleanup()`)
- **Don't hardcode values** that could change
- **Don't test implementation details** (test behavior)
- **Don't create brittle tests** that break on minor changes
- **Don't ignore errors** in test setup

## Floating Point Comparisons

For nutrition calculations involving floats, use delta comparison:

```go
// Use InDelta for floating point comparisons
assert.InDelta(t, expected, actual, 0.01, "Values should be within 0.01")

// For nutrition values (calories, macros)
assert.InDelta(t, 2000.0, goal.Calories, 0.01)
```

## Date Handling

Always use explicit time zones in tests:

```go
// Use UTC for consistency
date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

// For date-only comparisons
assert.Equal(t, expected.Format("2006-01-02"), actual.Format("2006-01-02"))
```

## Test Data Management

### Test Users

```go
// Create with factory
user := testutil.CreateTestUser(t, db)

// Create manually with specific data
user := &user.User{
    Email:  "test@example.com",
    Name:   "Test User",
    Age:    30,
    Gender: "male",
    Height: 175.0,
    Weight: 75.0,
    BodyFat: 15.0,
    ActivityLevel: user.ModeratelyActive,
    GoalType: user.Maintain,
}
db.Create(user)
user.HashPassword("password123")
db.Save(user)
```

### Test Goals

```go
goal := &NutritionGoal{
    UserID:    userID,
    Calories:  2000,
    Protein:   150,
    Carbs:     200,
    Fat:       65,
    Fiber:     30,
    StartDate: time.Now(),
    IsActive:  true,
}
```

## Running Tests

### Local Development

```bash
# Run all tests
make test

# Run only unit tests (fast)
make test-unit

# Run with coverage
make test-coverage

# Run specific package
go test -v ./internal/goal

# Run specific test
go test -v -run TestRepository_Create ./internal/goal
```

### CI/CD Integration

Tests are automatically run in GitHub Actions on:
- Pull requests
- Pushes to main/develop branches

Coverage reports are generated and tracked.

## Test Coverage Analysis

```bash
# Generate coverage report
make test-coverage

# View coverage in browser
open coverage.html

# Check coverage for specific package
go test -cover ./internal/goal

# Detailed coverage profile
go test -coverprofile=coverage.out ./internal/goal
go tool cover -func=coverage.out
```

## Debugging Tests

### Verbose Output

```bash
# Run with verbose output
go test -v ./internal/goal

# Show test logs
go test -v ./internal/goal -args -test.v
```

### Test Timeouts

```bash
# Increase timeout for slow tests
go test -timeout 5m ./internal/goal

# Default timeout in Makefile is 3m
```

### Database Inspection

```go
// In tests, you can inspect the database
var count int64
db.Model(&NutritionGoal{}).Count(&count)
t.Logf("Total goals in database: %d", count)

// Dump records for debugging
var goals []NutritionGoal
db.Find(&goals)
t.Logf("Goals: %+v", goals)
```

## Continuous Improvement

### Coverage Goals

Track coverage improvements over time:

```bash
# Current baseline (as of 2025-11-25)
auth:      15.9%
barcode:   37.9%
diary:      8.1%
food:      20.4%
recipe:    15.2%
goal:       0.0% → Target: 75%
metrics:    0.0% → Target: 75%
user:       0.0% → Target: 80%

# Overall: ~12% → Target: 65%
```

### Test Maintenance

- Review and update tests when requirements change
- Refactor tests to use new utilities/patterns
- Remove obsolete tests
- Keep test code as clean as production code

## Resources

- [Go Testing Documentation](https://pkg.go.dev/testing)
- [Testify Documentation](https://pkg.go.dev/github.com/stretchr/testify)
- [Testcontainers Documentation](https://golang.testcontainers.org/)
- [GORM Testing Guide](https://gorm.io/docs/testing.html)

## Questions?

Refer to existing test files for examples:
- `internal/food/repository_test.go` - Repository pattern
- `internal/auth/jwt_test.go` - Unit tests
- `internal/diary/repository_test.go` - Complex relationships

For test policy questions, consult the team or update this document.

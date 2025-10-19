# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Ultra-Bis** is a comprehensive nutrition tracking REST API built with Go and PostgreSQL, designed for athletes and fitness enthusiasts. It provides functionality similar to Samsung Health or MyFitnessPal, including food logging, meal tracking, body metrics monitoring, and personalized nutrition goal recommendations.

**Tech Stack:**
- Go 1.25
- PostgreSQL with GORM ORM
- JWT authentication (golang-jwt/jwt/v5)
- Standard library HTTP (net/http)
- Docker & Docker Compose for containerization

## Development Commands

### Running the Application

**With Docker (Recommended):**
```bash
# Start services in detached mode with rebuild
docker-compose up -d --build

# View API logs
docker-compose logs -f api

# Stop services
docker-compose down
```

**Without Docker:**
```bash
# Ensure PostgreSQL is running locally first
# Set DB_HOST=localhost (default for local dev)

# Run the API
go run cmd/api/main.go
```

**Health Check:**
```bash
curl http://localhost:8080/health
```

### Database Operations

**Access PostgreSQL directly:**
```bash
docker exec -it fooddb-postgres psql -U postgres -d fooddb
```

**Common SQL commands:**
```sql
-- List all tables
\dt

-- View schema
\d users
\d foods
\d nutrition_goals
\d diary_entries
\d body_metrics

-- Clear all data (useful for testing)
TRUNCATE users, foods, nutrition_goals, diary_entries, body_metrics CASCADE;
```

### Testing the API

Use the provided HTTP request files for manual testing:
- `request/food.http` - Food CRUD operations
- `request/nutrition.http` - Complete nutrition tracking workflow

These work with IntelliJ IDEA/GoLand (built-in) or VS Code (with REST Client extension).

### Go Commands

```bash
# Install dependencies
go mod download

# Add a new dependency
go get github.com/some/package

# Tidy up dependencies
go mod tidy

# Build the binary
go build -o bin/api cmd/api/main.go

# Run the binary
./bin/api
```

## Architecture

### High-Level Structure

The application follows a **clean architecture pattern** organized by domain/feature:

```
cmd/api/main.go          → Application entry point, wiring, HTTP server setup
internal/database/       → Database connection (GORM)
internal/auth/           → Authentication (JWT, middleware)
internal/user/           → User domain (model, repository)
internal/food/           → Food domain (model, repository, handler, router)
internal/diary/          → Meal logging domain
internal/goal/           → Nutrition goals domain
internal/metrics/        → Body metrics tracking domain
```

Each domain follows the **Repository pattern**:
- `model.go` - GORM models (database structs)
- `repository.go` - Database operations (Create, Read, Update, Delete)
- `handler.go` - HTTP handlers (request parsing, response writing)
- `router.go` - Route registration

### Application Flow

1. **Startup** (`cmd/api/main.go`):
   - Connect to PostgreSQL via GORM
   - Run auto-migrations for all models
   - Initialize repositories (DB access layer)
   - Initialize handlers (HTTP layer)
   - Register routes on the HTTP mux
   - Start HTTP server on port 8080

2. **Request Flow**:
   ```
   HTTP Request
     ↓
   Router (mux) → Middleware (JWT auth if protected)
     ↓
   Handler (parse request, validate)
     ↓
   Repository (database operations via GORM)
     ↓
   Response (JSON)
   ```

### Authentication System

- **JWT-based** authentication using `golang-jwt/jwt/v5`
- **Password hashing** with `bcrypt` (cost factor 10)
- **Middleware** (`internal/auth/middleware.go`) adds user context to protected routes
- Protected routes extract `user_id` from context: `r.Context().Value("user_id").(uint)`

**Token Flow:**
1. User registers/logs in → JWT token generated
2. Client stores token
3. Client sends token in `Authorization: Bearer <token>` header
4. Middleware validates token and injects user_id into request context
5. Handlers access user_id for user-specific operations

### Database Schema

GORM auto-migrates the following models on startup:

- **users** - User accounts with profile data (age, gender, height, activity level, goal type)
- **foods** - Food database with nutritional information (calories, protein, carbs, fat, fiber)
- **nutrition_goals** - User nutrition targets with start/end dates, active status
- **diary_entries** - Daily meal logs (references foods, calculated nutrition per serving)
- **body_metrics** - Weight and body composition tracking (includes BMI calculation)

**Important Relationships:**
- Users have many NutritionGoals (1:N)
- Users have many DiaryEntries (1:N)
- Users have many BodyMetrics (1:N)
- DiaryEntries reference Foods (N:1)
- Only one NutritionGoal per user can be active (`is_active=true`) at a time

### Key Features

**Goal Recommendation Algorithm** (`internal/goal/handler.go:217-282`):
1. Calculates BMR (Basal Metabolic Rate) using **Mifflin-St Jeor Equation**
2. Calculates TDEE (Total Daily Energy Expenditure) = BMR × activity multiplier
3. Adjusts calories based on weight goal:
   - Weight loss: deficit capped at 1000 cal/day
   - Weight gain: surplus capped at 500 cal/day
   - Maintenance: TDEE
4. Calculates macros: 30% protein, 40% carbs, 30% fat (athlete-focused)
5. Calculates fiber: 14g per 1000 calories

**Daily Summary with Adherence** (`internal/diary/handler.go`):
- Aggregates all meals for a given date
- Compares totals against active nutrition goal
- Calculates adherence percentages for each macro

**Trend Analysis** (`internal/metrics/handler.go`):
- Supports 7/30/90-day periods
- Calculates average metrics and changes over time

## Environment Variables

Set these in `.env` file or Docker Compose:

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DB_HOST` | PostgreSQL host | `localhost` | Yes |
| `DB_PORT` | PostgreSQL port | `5432` | No |
| `DB_USER` | PostgreSQL user | `postgres` | No |
| `DB_PASSWORD` | PostgreSQL password | `postgres` | No |
| `DB_NAME` | Database name | `fooddb` | No |
| `PORT` | API server port | `8080` | No |
| `JWT_SECRET` | JWT signing secret | `your-secret-key-change-in-production` | **Change in prod** |

**Docker Note:** When running with Docker Compose, set `DB_HOST=postgres` (service name).

## Code Patterns

### Adding a New Endpoint

1. **Define request/response structs** in `model.go`
2. **Implement handler function** in `handler.go`:
   ```go
   func (h *Handler) MyEndpoint(w http.ResponseWriter, r *http.Request) {
       // Extract user_id from context if protected
       userID, ok := r.Context().Value("user_id").(uint)

       // Parse request body
       var req MyRequest
       json.NewDecoder(r.Body).Decode(&req)

       // Call repository
       result, err := h.repo.SomeOperation(userID, req)

       // Write JSON response
       writeJSON(w, http.StatusOK, result)
   }
   ```
3. **Register route** in `router.go`:
   ```go
   // Protected route
   mux.HandleFunc("/my-endpoint", auth.JWTMiddleware(handler.MyEndpoint))

   // Public route
   mux.HandleFunc("/my-endpoint", handler.MyEndpoint)
   ```

### Repository Operations

All database operations use GORM. Common patterns:

```go
// Create
db.Create(&model)

// Find by ID
db.First(&model, id)

// Find with conditions
db.Where("user_id = ? AND is_active = ?", userID, true).First(&model)

// Update
db.Save(&model)

// Delete
db.Delete(&model)

// Preload associations
db.Preload("Food").Find(&entries)
```

### Error Handling

Use the `writeError` helper consistently:
```go
writeError(w, http.StatusBadRequest, "Invalid request body")
writeError(w, http.StatusNotFound, "Resource not found")
writeError(w, http.StatusInternalServerError, err.Error())
```

## Frontend Integration

A comprehensive frontend specification is available in `FRONTEND_SPEC.md`. Key points:

- All protected endpoints require `Authorization: Bearer <token>` header
- Dates should be formatted as `YYYY-MM-DD` (e.g., "2025-01-15")
- API returns timestamps in RFC3339 format
- Serving sizes are multipliers (1.0 = 1 serving, 1.5 = 1.5 servings)
- Meal types: `breakfast`, `lunch`, `dinner`, `snack`

## Important Constraints

### Creating New Goals
When creating a new nutrition goal for a user, the system automatically deactivates any existing active goals. Only one goal can be active per user at a time (enforced in `internal/goal/repository.go`).

### Nutrition Calculations
Diary entries automatically calculate nutrition values based on food data multiplied by serving size. These are stored denormalized for historical accuracy (if food data changes later, past diary entries remain unchanged).

### BMI Calculation
BMI is auto-calculated in body metrics based on weight (kg) and user height (cm) from their profile: `BMI = weight / (height/100)²`

## Common Development Workflows

### Adding a New Food Item via API
```bash
curl -X POST http://localhost:8080/foods \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Salmon",
    "description": "Atlantic salmon, baked",
    "calories": 206,
    "protein": 22,
    "carbs": 0,
    "fat": 13,
    "fiber": 0
  }'
```

### Testing Protected Endpoints
1. Register/login to get token
2. Use token in Authorization header:
```bash
curl -X GET http://localhost:8080/goals \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

### Rebuilding After Code Changes
```bash
# Rebuild and restart
docker-compose up -d --build

# Check logs for errors
docker-compose logs -f api
```

## Troubleshooting

**"Failed to connect to database"**
- Check PostgreSQL is running: `docker ps`
- Verify environment variables match docker-compose.yml
- For Docker: ensure `DB_HOST=postgres`, not `localhost`

**"Unauthorized" on protected routes**
- Verify token is being sent in Authorization header
- Check token hasn't expired (24h default)
- Ensure token format is `Bearer <token>`

**Auto-migration errors**
- Check GORM model struct tags
- Verify database user has CREATE TABLE permissions
- Review logs: `docker-compose logs -f api`

**Port already in use**
- Stop existing services: `docker-compose down`
- Check for processes on port 8080: `lsof -i :8080`
- Change PORT environment variable if needed
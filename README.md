# Ultra-Bis Nutrition Tracking API

A comprehensive RESTful API built with Go and PostgreSQL for tracking nutrition, meals, body metrics, and fitness goals. Designed for athletes and fitness enthusiasts who want to monitor their nutrition intake and progress towards their goals - similar to Samsung Health or MyFitnessPal.

## Features

- **User Authentication** - JWT-based authentication system
- **Food Database** - CRUD operations for food items with full nutritional information
- **Nutrition Goals** - Set and track daily macro/calorie targets with personalized recommendations
- **Meal Logging** - Track daily food intake organized by meals (breakfast, lunch, dinner, snacks)
- **Daily Summaries** - View nutrition totals and goal adherence percentages
- **Body Metrics Tracking** - Monitor weight, body fat %, muscle mass over time
- **Trends & Analytics** - Visualize progress with 7/30/90-day trend analysis
- **GORM ORM** - Clean database operations using GORM (like Sequelize for JS)
- **Dockerized** - Easy deployment with Docker Compose

## Tech Stack

- **Language**: Go 1.25
- **Database**: PostgreSQL with GORM ORM
- **Authentication**: JWT tokens with bcrypt password hashing
- **HTTP**: Standard library (net/http)
- **Containerization**: Docker & Docker Compose

## Quick Start with Docker

### Prerequisites
- Docker
- Docker Compose

### Running the Application

1. **Start the application**:
   ```bash
   docker-compose up -d --build
   ```

2. **Check if it's running**:
   ```bash
   curl http://localhost:8080/health
   ```

3. **View logs**:
   ```bash
   docker-compose logs -f api
   ```

4. **Stop the application**:
   ```bash
   docker-compose down
   ```

## API Endpoints

### Authentication

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/auth/register` | Register new user | No |
| POST | `/auth/login` | Login and get JWT token | No |
| GET | `/auth/me` | Get current user profile | Yes |
| PUT | `/users/profile` | Update user profile | Yes |

### Foods

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/foods` | Create food item | No |
| GET | `/foods` | List all foods | No |
| GET | `/foods/{id}` | Get food by ID | No |
| PUT | `/foods/{id}` | Update food | No |
| DELETE | `/foods/{id}` | Delete food | No |

### Nutrition Goals

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/goals` | Create nutrition goal | Yes |
| GET | `/goals` | Get active goal | Yes |
| GET | `/goals/all` | Get all goals history | Yes |
| POST | `/goals/recommended` | Calculate recommended goals | Yes |
| PUT | `/goals/{id}` | Update goal | Yes |
| DELETE | `/goals/{id}` | Delete goal | Yes |

### Diary (Meal Logging)

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/diary/entries` | Log food/meal | Yes |
| GET | `/diary/entries?date=YYYY-MM-DD` | Get entries by date | Yes |
| GET | `/diary/summary/{date}` | Get daily summary with adherence | Yes |
| PUT | `/diary/entries/{id}` | Update entry | Yes |
| DELETE | `/diary/entries/{id}` | Delete entry | Yes |

### Body Metrics

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/metrics` | Log body metrics | Yes |
| GET | `/metrics` | Get all metrics | Yes |
| GET | `/metrics/latest` | Get latest measurement | Yes |
| GET | `/metrics/trends?period=7d\|30d\|90d` | Get trend analysis | Yes |
| DELETE | `/metrics/{id}` | Delete metric | Yes |

## Usage Examples

### 1. Register and Login

```bash
# Register
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "athlete@example.com",
    "password": "password123",
    "name": "John Athlete"
  }'

# Login (save the token)
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "athlete@example.com",
    "password": "password123"
  }'
```

### 2. Update Profile

```bash
curl -X PUT http://localhost:8080/users/profile \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "age": 28,
    "gender": "male",
    "height": 180,
    "activity_level": "active",
    "goal_type": "lose"
  }'
```

### 3. Get Recommended Goals

```bash
curl -X POST http://localhost:8080/goals/recommended \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "weight": 75,
    "target_weight": 70,
    "weeks_to_goal": 8
  }'
```

### 4. Set Nutrition Goals

```bash
curl -X POST http://localhost:8080/goals \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "calories": 2200,
    "protein": 165,
    "carbs": 220,
    "fat": 73,
    "fiber": 31
  }'
```

### 5. Log a Meal

```bash
curl -X POST http://localhost:8080/diary/entries \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "food_id": 1,
    "date": "2025-01-15",
    "meal_type": "breakfast",
    "serving_size": 1.5,
    "notes": "Post-workout meal"
  }'
```

### 6. Get Daily Summary

```bash
curl -X GET "http://localhost:8080/diary/summary/2025-01-15" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 7. Log Body Metrics

```bash
curl -X POST http://localhost:8080/metrics \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "weight": 75.2,
    "body_fat_percent": 15.5,
    "muscle_mass_percent": 45.8,
    "notes": "Morning weigh-in, fasted"
  }'
```

### 8. View Progress Trends

```bash
curl -X GET "http://localhost:8080/metrics/trends?period=30d" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Testing with HTTP Files

Two comprehensive HTTP request files are included for easy testing:

- `request/food.http` - Food CRUD operations
- `request/nutrition.http` - Complete nutrition tracking workflow

### Using in IntelliJ IDEA / GoLand
1. Open the `.http` file
2. Click the green play button next to any request
3. For protected endpoints, replace `{{token}}` with your actual JWT token

### Using in VS Code
1. Install the "REST Client" extension
2. Open the `.http` file
3. Click "Send Request" above any request

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | Database host | `localhost` (use `postgres` for Docker) |
| `DB_PORT` | Database port | `5432` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | `postgres` |
| `DB_NAME` | Database name | `fooddb` |
| `PORT` | API server port | `8080` |
| `JWT_SECRET` | Secret key for JWT tokens | `your-secret-key-change-in-production` |

## Project Structure

```
ultra-bis/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── auth/
│   │   ├── handler.go           # Auth endpoints (register, login)
│   │   ├── jwt.go               # JWT token generation/validation
│   │   ├── middleware.go        # JWT authentication middleware
│   │   └── router.go            # Auth routes
│   ├── database/
│   │   └── postgres.go          # GORM database connection
│   ├── diary/
│   │   ├── model.go             # Diary entry models
│   │   ├── repository.go        # Diary database operations
│   │   ├── handler.go           # Diary HTTP handlers
│   │   └── router.go            # Diary routes
│   ├── food/
│   │   ├── model.go             # Food models
│   │   ├── repository.go        # Food database operations
│   │   ├── handler.go           # Food HTTP handlers
│   │   └── router.go            # Food routes
│   ├── goal/
│   │   ├── model.go             # Nutrition goal models
│   │   ├── repository.go        # Goal database operations
│   │   ├── handler.go           # Goal HTTP handlers
│   │   └── router.go            # Goal routes
│   ├── metrics/
│   │   ├── model.go             # Body metric models
│   │   ├── repository.go        # Metrics database operations
│   │   ├── handler.go           # Metrics HTTP handlers
│   │   └── router.go            # Metrics routes
│   └── user/
│       ├── model.go             # User model
│       └── repository.go        # User database operations
├── request/
│   ├── food.http                # Food API tests
│   └── nutrition.http           # Complete nutrition workflow tests
├── Dockerfile                   # Docker build instructions
├── docker-compose.yml           # Docker Compose configuration
├── .env.example                 # Environment variables template
└── go.mod                       # Go module dependencies
```

## Database Schema

The application automatically creates the following tables using GORM Auto-Migration:

- **users** - User accounts with profile information
- **foods** - Food items with nutritional data
- **nutrition_goals** - User nutrition targets
- **diary_entries** - Daily meal logging
- **body_metrics** - Weight and body composition tracking

## Features Comparison with Samsung Health

| Feature | Ultra-Bis API | Samsung Health |
|---------|---------------|----------------|
| User Accounts | ✅ | ✅ |
| Food Database | ✅ | ✅ |
| Custom Foods | ✅ | ✅ |
| Meal Logging | ✅ | ✅ |
| Macro Tracking | ✅ | ✅ |
| Goal Setting | ✅ | ✅ |
| Personalized Recommendations | ✅ | ✅ |
| Body Metrics | ✅ | ✅ |
| Progress Trends | ✅ | ✅ |
| Daily Summaries | ✅ | ✅ |
| Water Tracking | 🚧 Future | ✅ |
| Recipes | 🚧 Future | ✅ |
| Barcode Scanner | ❌ API only | ✅ |
| Social Features | 🚧 Future | ✅ |

## Development

### Running Locally (without Docker)

1. **Create database**:
   ```bash
   createdb fooddb
   ```

2. **Run the application**:
   ```bash
   go run cmd/api/main.go
   ```

### Rebuild after code changes

```bash
docker-compose up -d --build
```

### Access PostgreSQL directly

```bash
docker exec -it fooddb-postgres psql -U postgres -d fooddb
```

### Useful PostgreSQL commands

```sql
-- List all tables
\dt

-- View users
SELECT id, email, name, activity_level, goal_type FROM users;

-- View diary entries
SELECT * FROM diary_entries ORDER BY date DESC LIMIT 10;

-- View metrics
SELECT * FROM body_metrics ORDER BY date DESC;

-- Clear all data
TRUNCATE users, foods, nutrition_goals, diary_entries, body_metrics CASCADE;
```

## Contributing

Contributions are welcome! This is a comprehensive nutrition tracking backend ready for mobile or web frontend integration.

## License

MIT License

## Author

Built with Go, GORM, and PostgreSQL

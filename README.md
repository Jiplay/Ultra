# Food API

A RESTful API built with Go and PostgreSQL for managing food items with nutritional information.

## Features

- CRUD operations for food items
- Nutritional information tracking (calories, protein, carbs, fat, fiber)
- PostgreSQL database
- Dockerized for easy deployment

## Quick Start with Docker

### Prerequisites
- Docker
- Docker Compose

### Running the Application

1. **Start the application**:
   ```bash
   docker-compose up -d
   ```

   This will:
   - Start a PostgreSQL database container
   - Build and start the API container
   - Automatically create the database schema

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

5. **Stop and remove volumes** (deletes all data):
   ```bash
   docker-compose down -v
   ```

## Running Locally (without Docker)

### Prerequisites
- Go 1.25+
- PostgreSQL

### Steps

1. **Create database**:
   ```bash
   createdb fooddb
   ```

2. **Run the application**:
   ```bash
   go run cmd/api/main.go
   ```

## API Endpoints

### Health Check
```bash
GET /health
```

### Create Food Item
```bash
POST /foods
Content-Type: application/json

{
  "name": "Chicken Breast",
  "description": "Grilled skinless chicken breast",
  "calories": 165,
  "protein": 31,
  "carbs": 0,
  "fat": 3.6,
  "fiber": 0
}
```

### Get All Foods
```bash
GET /foods
```

### Get Food by ID
```bash
GET /foods/{id}
```

### Update Food
```bash
PUT /foods/{id}
Content-Type: application/json

{
  "name": "Chicken Breast Updated",
  "description": "Grilled skinless chicken breast - updated",
  "calories": 170,
  "protein": 32,
  "carbs": 0,
  "fat": 4,
  "fiber": 0
}
```

### Delete Food
```bash
DELETE /foods/{id}
```

## Example Usage

### Create a food item
```bash
curl -X POST http://localhost:8080/foods \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Brown Rice",
    "description": "Cooked brown rice",
    "calories": 216,
    "protein": 5,
    "carbs": 45,
    "fat": 1.8,
    "fiber": 3.5
  }'
```

### Get all foods
```bash
curl http://localhost:8080/foods
```

### Get specific food
```bash
curl http://localhost:8080/foods/1
```

### Update a food
```bash
curl -X PUT http://localhost:8080/foods/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Brown Rice (1 cup)",
    "description": "Cooked brown rice, one cup serving",
    "calories": 216,
    "protein": 5,
    "carbs": 45,
    "fat": 1.8,
    "fiber": 3.5
  }'
```

### Delete a food
```bash
curl -X DELETE http://localhost:8080/foods/1
```

## Environment Variables

The application can be configured using environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | Database host | `localhost` (use `postgres` for Docker) |
| `DB_PORT` | Database port | `5432` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | `postgres` |
| `DB_NAME` | Database name | `fooddb` |
| `PORT` | API server port | `8080` |

## Project Structure

```
ultra-bis/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── database/
│   │   └── postgres.go          # Database connection & schema
│   └── food/
│       ├── model.go             # Food model definitions
│       ├── repository.go        # Database operations
│       └── handler.go           # HTTP handlers
├── Dockerfile                   # Docker build instructions
├── docker-compose.yml           # Docker Compose configuration
├── .dockerignore               # Files to exclude from Docker build
├── .env.example                # Environment variables template
└── go.mod                      # Go module dependencies
```

## Development

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
-- List all foods
SELECT * FROM foods;

-- Count foods
SELECT COUNT(*) FROM foods;

-- Clear all foods
TRUNCATE foods RESTART IDENTITY;
```

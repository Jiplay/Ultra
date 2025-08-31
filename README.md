# Ultra - Mobile App Backend API

A comprehensive Go backend API for a mobile fitness and nutrition app using MongoDB for users and PostgreSQL for programs and nutrition data.

## 🏗️ Architecture

```
Ultra/
├── config/          # Configuration management
├── database/        # Database connections and setup
├── users/           # User management (MongoDB)
├── programs/        # Workout programs (PostgreSQL)
├── nutrition/       # Foods, recipes, goals (PostgreSQL)
├── main.go          # Main application entry point
├── docker-compose.yml
├── .env
└── .env.example
```

## 🚀 Features

### User Management (MongoDB)
- User creation and profile management
- Profile editing (name, email, weight, height, age, picture URL)
- User authentication ready structure

### Programs Management (PostgreSQL)
- Create workout programs with multiple workouts
- Each workout contains exercises with:
  - Name, weight, repetitions, series
  - Rest time and notes
- Full CRUD operations for programs

### Nutrition Management (PostgreSQL)
- Food database with nutritional information
- Recipe creation with ingredients
- Nutrition goals management
- Search functionality for foods

## 🛠️ Setup

### Prerequisites
- Go 1.23+
- Docker and Docker Compose (for databases)
- MongoDB Atlas (for production) or local MongoDB
- PostgreSQL database

### Installation

1. **Clone and setup environment:**
```bash
cd Ultra
cp .env.example .env
# Edit .env with your database connections
```

2. **Start databases locally:**
```bash
docker-compose up -d
```

3. **Install dependencies:**
```bash
go mod tidy
```

4. **Run the application:**
```bash
go run main.go
```

The API will be available at `http://localhost:8080`

## 📡 API Endpoints

### Health Check
- `GET /health` - API health status

### Users (MongoDB)
- `POST /api/v1/users` - Create user
- `GET /api/v1/users/{id}` - Get user profile
- `PUT /api/v1/users/{id}` - Update user profile
- `DELETE /api/v1/users/{id}` - Delete user

### Programs (PostgreSQL)
- `POST /api/v1/programs` - Create program with workouts
- `GET /api/v1/programs?user_id={id}` - Get user programs
- `GET /api/v1/programs/{id}?user_id={id}` - Get specific program
- `PUT /api/v1/programs/{id}` - Update program
- `DELETE /api/v1/programs/{id}` - Delete program

### Nutrition (PostgreSQL)
- `POST /api/v1/nutrition/foods` - Create food
- `GET /api/v1/nutrition/foods?search={query}` - Search foods
- `GET /api/v1/nutrition/foods/{id}` - Get food details
- `POST /api/v1/nutrition/recipes` - Create recipe
- `GET /api/v1/nutrition/recipes?user_id={id}` - Get user recipes
- `GET /api/v1/nutrition/recipes/{id}?user_id={id}` - Get recipe details
- `GET /api/v1/nutrition/goals/{user_id}` - Get nutrition goals
- `PUT /api/v1/nutrition/goals/{user_id}` - Update nutrition goals

## 🗄️ Database Schema

### MongoDB (Users Collection)
```json
{
  "_id": "ObjectId",
  "email": "string",
  "name": "string", 
  "weight": "number",
  "height": "number",
  "age": "number",
  "picture": "string",
  "created_at": "datetime",
  "updated_at": "datetime"
}
```

### PostgreSQL Tables

**programs**: Program information
**workouts**: Individual workouts in programs  
**exercises**: Exercises within workouts
**foods**: Food nutritional database
**recipes**: User-created recipes
**recipe_ingredients**: Recipe ingredients linking
**nutrition_goals**: User nutrition targets

## 🧪 Example API Calls

### Create User
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","name":"John Doe"}'
```

### Create Program with Workouts
```bash
curl -X POST http://localhost:8080/api/v1/programs \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user123" \
  -d '{
    "name": "Push Pull Legs",
    "description": "3-day split routine",
    "workouts": [
      {
        "name": "Push Day",
        "day_of_week": 1,
        "exercises": [
          {
            "name": "Bench Press",
            "weight": 100.0,
            "repetitions": 10,
            "series": 3,
            "rest_time": 120
          }
        ]
      }
    ]
  }'
```

### Create Food
```bash
curl -X POST http://localhost:8080/api/v1/nutrition/foods \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user123" \
  -d '{
    "name": "Chicken Breast",
    "calories_per_100g": 165,
    "protein_per_100g": 31.0,
    "carbs_per_100g": 0.0,
    "fat_per_100g": 3.6,
    "fiber_per_100g": 0.0
  }'
```

## 🔧 Configuration

Environment variables in `.env`:
- `MONGODB_URI`: MongoDB connection string
- `POSTGRES_*`: PostgreSQL connection details  
- `PORT`: Server port (default: 8080)
- `JWT_SECRET`: JWT signing key

## 🚦 Development

The API includes:
- CORS middleware for cross-origin requests
- Request logging middleware
- Automatic PostgreSQL table creation
- Clean package organization
- Error handling and validation

For production deployment, update the MongoDB URI to your Atlas cluster and configure PostgreSQL accordingly.

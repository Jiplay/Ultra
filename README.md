# 🍎 Ultra Food Catalog API

A high-performance REST API for managing food catalog data with nutritional information, built with Go and PostgreSQL.

## 🚀 Features

- **RESTful API** for food catalog management
- **PostgreSQL Database** with automatic initialization
- **Docker-ready** with container orchestration
- **Health checks** and monitoring endpoints
- **Automatic migrations** and sample data
- **GitHub Actions** CI/CD pipeline
- **Fallback storage** (in-memory when database unavailable)

## 📋 Quick Start

### Option 1: Docker Deployment (Recommended)

1. **Clone and configure**:
   ```bash
   git clone https://github.com/yourusername/ultra.git
   cd ultra
   cp .env.docker .env
   ```

2. **Deploy with one command**:
   ```bash
   ./deploy.sh
   ```

3. **Verify deployment**:
   ```bash
   curl http://localhost:8080/health
   ```

### Option 2: Manual Container Setup

1. **Create network**:
   ```bash
   docker network create ultra-network
   ```

2. **Run PostgreSQL**:
   ```bash
   docker build -t ultra-postgres ./postgres/
   docker run -d --name ultra-postgres --network ultra-network \
     -v ultra-postgres-data:/var/lib/postgresql/data \
     -p 5432:5432 ultra-postgres
   ```

3. **Run API**:
   ```bash
   docker pull ghcr.io/yourusername/ultra:latest
   docker run -d --name ultra-api --network ultra-network \
     --env-file .env.docker -p 8080:8080 \
     ghcr.io/yourusername/ultra:latest
   ```

### Option 3: Local Development

1. **Start PostgreSQL** (using Docker):
   ```bash
   docker run -d --name postgres-dev -p 5432:5432 \
     -e POSTGRES_USER=ultra_user -e POSTGRES_PASSWORD=ultra_password \
     -e POSTGRES_DB=ultra_food_db postgres:15-alpine
   ```

2. **Configure environment**:
   ```bash
   cp .env.example .env
   # Edit .env: set USE_DATABASE=true, DB_HOST=localhost
   ```

3. **Run the application**:
   ```bash
   go mod download
   go run main.go
   ```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default | Docker |
|----------|-------------|---------|---------|
| `USE_DATABASE` | Enable PostgreSQL | `false` | `true` |
| `PORT` | Server port | `8080` | `8080` |
| `DB_HOST` | Database host | `localhost` | `ultra-postgres` |
| `DB_PORT` | Database port | `5432` | `5432` |
| `DB_USER` | Database user | `ultra_user` | `ultra_user` |
| `DB_PASSWORD` | Database password | `ultra_password` | `ultra_password` |
| `DB_NAME` | Database name | `ultra_food_db` | `ultra_food_db` |
| `DB_SSL_MODE` | SSL mode | `disable` | `disable` |

## 📊 API Endpoints

### System Endpoints
- **GET /** - API information and available endpoints
- **GET /health** - Health check with database status

### Food Management
- **POST /api/foods** - Create a new food item
- **GET /api/foods** - List all foods (sorted by creation date)
- **GET /api/foods/{id}** - Get specific food by ID
- **PUT /api/foods/{id}** - Update food item (partial updates supported)
- **DELETE /api/foods/{id}** - Delete food item

### Example Requests

**Create Food**:
```bash
curl -X POST http://localhost:8080/api/foods \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Grilled Chicken",
    "calories": 231,
    "protein": 43.5,
    "carbs": 0.0,
    "fat": 5.0
  }'
```

**Get All Foods**:
```bash
curl http://localhost:8080/api/foods | jq '.'
```

**Update Food**:
```bash
curl -X PUT http://localhost:8080/api/foods/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Organic Grilled Chicken",
    "calories": 225
  }'
```

## 🗄️ Database Schema

The PostgreSQL database includes:

### Foods Table
```sql
CREATE TABLE foods (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    calories INTEGER NOT NULL DEFAULT 0,
    protein DECIMAL(10,2) NOT NULL DEFAULT 0.0,
    carbs DECIMAL(10,2) NOT NULL DEFAULT 0.0,
    fat DECIMAL(10,2) NOT NULL DEFAULT 0.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### Sample Data
The database comes pre-loaded with 15 common foods including:
- Fruits (Apple, Banana, Orange)
- Proteins (Chicken, Salmon, Eggs)
- Grains (Rice, Quinoa, Oatmeal)
- Vegetables (Broccoli, Spinach)
- Others (Yogurt, Almonds, Avocado)

### Database Features
- **Automatic timestamps** with triggers
- **Performance indexes** on name, created_at, and calories
- **Utility functions** for nutritional calculations
- **High-protein foods view** for easy querying

## 🏗️ Architecture

### Application Structure
```
ultra/
├── main.go                 # Application entry point
├── database/               # Database connection and config
│   ├── config.go          # Environment-based configuration
│   ├── connection.go      # Connection management with retry
│   └── migrate.go         # Schema migration and sample data
├── food/                   # Food domain logic
│   ├── models.go          # Data models and request/response types
│   ├── repository.go      # Data access layer (PostgreSQL + in-memory)
│   ├── controller.go      # Business logic and validation
│   ├── handlers.go        # HTTP request handlers
│   └── router.go          # Route definitions
├── postgres/               # PostgreSQL container setup
│   ├── Dockerfile         # Custom PostgreSQL image
│   └── init/              # Database initialization scripts
├── .github/workflows/      # CI/CD pipeline
└── deploy.sh              # Deployment script
```

### Design Patterns
- **Repository Pattern** for data access abstraction
- **Controller Pattern** for business logic separation
- **Handler Pattern** for HTTP request processing
- **Environment Configuration** for flexible deployment

## 🔄 CI/CD Pipeline

GitHub Actions automatically:
1. **Builds** the Go application on push to main
2. **Creates** optimized Docker image  
3. **Pushes** to GitHub Container Registry
4. **Tags** with version numbers and `latest`
5. **Caches** build layers for faster subsequent builds

### Triggering Builds
- Push to `main` branch → Build and push `latest` tag
- Create git tag `v*` → Build and push versioned tag
- Pull requests → Build only (no push)

## 📈 Monitoring and Maintenance

### Health Monitoring
```bash
# API health with database status
curl http://localhost:8080/health

# PostgreSQL direct check
docker exec ultra-postgres pg_isready -U ultra_user -d ultra_food_db
```

### Logs and Debugging
```bash
# Application logs
docker logs -f ultra-api

# Database logs
docker logs -f ultra-postgres

# Container status
docker ps --filter "name=ultra-"
```

### Database Management
```bash
# Connect to database
docker exec -it ultra-postgres psql -U ultra_user -d ultra_food_db

# View all foods
docker exec ultra-postgres psql -U ultra_user -d ultra_food_db \
  -c "SELECT id, name, calories FROM foods LIMIT 10;"

# Backup database
docker exec ultra-postgres pg_dump -U ultra_user ultra_food_db > backup.sql
```

### Updates and Maintenance
```bash
# Update API to latest version
docker pull ghcr.io/yourusername/ultra:latest
docker stop ultra-api && docker rm ultra-api
# Re-run API container with new image

# Full redeployment
./deploy.sh

# Clean restart (⚠️ removes all data)
./deploy.sh --cleanup
./deploy.sh
```

## 🚀 Production Deployment

### Security Considerations
- **Change default passwords** in production
- **Enable SSL** for database connections (`DB_SSL_MODE=require`)
- **Use secrets management** for sensitive environment variables
- **Implement API authentication** and rate limiting
- **Set up proper firewall rules**

### Performance Optimization
- **Connection pooling** (configured automatically)
- **Database indexes** (created during migration)
- **Container resource limits**
- **Load balancing** for multiple API instances

### High Availability Setup
- **Database replication** for PostgreSQL
- **Container orchestration** with Kubernetes or Docker Swarm
- **Health checks** and automatic recovery
- **Backup automation** and disaster recovery

## 🛠️ Development

### Local Development Setup
1. Install Go 1.23+
2. Install PostgreSQL or use Docker
3. Copy `.env.example` to `.env` and configure
4. Run `go mod download`
5. Run `go run main.go`

### Running Tests
```bash
# Build and verify
go build

# Run with different storage backends
USE_DATABASE=false go run main.go    # In-memory
USE_DATABASE=true go run main.go     # PostgreSQL
```

### Adding New Features
1. Define models in `food/models.go`
2. Implement repository methods in `food/repository.go`
3. Add business logic in `food/controller.go`
4. Create HTTP handlers in `food/handlers.go`
5. Update routes in `food/router.go`

## 📝 License

This project is licensed under the MIT License.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📞 Support

For issues and questions:
- **GitHub Issues**: Report bugs and feature requests
- **Container Logs**: Check `docker logs ultra-api` for application issues
- **Database Issues**: Verify connectivity and check PostgreSQL logs
- **Deployment Help**: See `DEPLOYMENT_GUIDE.md` for detailed instructions
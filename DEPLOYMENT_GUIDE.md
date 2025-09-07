# Ultra Food API - Docker Deployment Guide

This guide explains how to deploy the Ultra Food API with PostgreSQL using Docker containers.

## Overview

The deployment consists of:
- **PostgreSQL Container**: Database with initialization scripts
- **API Container**: Go application pulled from GitHub Container Registry
- **Docker Network**: Isolated network for container communication
- **Persistent Storage**: PostgreSQL data volume

## Prerequisites

- Docker installed and running
- Docker Compose (optional, for alternative deployment)
- Access to GitHub Container Registry (for pulling API image)

## Quick Start

1. **Clone the repository**:
   ```bash
   git clone https://github.com/yourusername/ultra.git
   cd ultra
   ```

2. **Configure environment**:
   ```bash
   cp .env.docker .env
   # Edit .env if needed (optional)
   ```

3. **Deploy everything**:
   ```bash
   ./deploy.sh
   ```

4. **Verify deployment**:
   ```bash
   curl http://localhost:8080/health
   ```

## Manual Deployment Steps

### 1. Create Docker Network

```bash
docker network create ultra-network
```

### 2. Deploy PostgreSQL

Build and run PostgreSQL with initialization:

```bash
# Build PostgreSQL image with init scripts
docker build -t ultra-postgres:latest ./postgres/

# Run PostgreSQL container
docker run -d \
  --name ultra-postgres \
  --network ultra-network \
  -v ultra-postgres-data:/var/lib/postgresql/data \
  -p 5432:5432 \
  --restart unless-stopped \
  ultra-postgres:latest

# Wait for PostgreSQL to be ready
docker exec ultra-postgres pg_isready -U ultra_user -d ultra_food_db
```

### 3. Deploy API

Pull and run the API container:

```bash
# Pull latest API image from GitHub Container Registry
docker pull ghcr.io/yourusername/ultra:latest

# Run API container
docker run -d \
  --name ultra-api \
  --network ultra-network \
  --env-file .env.docker \
  -p 8080:8080 \
  --restart unless-stopped \
  ghcr.io/yourusername/ultra:latest
```

## Environment Configuration

### Container Environment (`.env.docker`)

```bash
# Application Settings
PORT=8080
USE_DATABASE=true

# Database Configuration (using container names)
DB_HOST=ultra-postgres
DB_PORT=5432
DB_USER=ultra_user
DB_PASSWORD=ultra_password
DB_NAME=ultra_food_db
DB_SSL_MODE=disable
```

### Key Configuration Notes

- **DB_HOST**: Use container name `ultra-postgres` for container-to-container communication
- **USE_DATABASE**: Must be `true` to use PostgreSQL
- **Network**: Both containers must be on the same Docker network

## Database Features

### Automatic Initialization

The PostgreSQL container automatically:
- Creates the `ultra_food_db` database
- Creates the `foods` table with indexes
- Inserts sample food data (15 items)
- Sets up database functions and triggers
- Grants proper permissions

### Sample Data Included

The database comes pre-loaded with:
- Fruits: Apple, Banana, Orange
- Proteins: Chicken Breast, Salmon, Eggs
- Grains: Brown Rice, Quinoa, Oatmeal
- Vegetables: Broccoli, Spinach, Sweet Potato
- Others: Greek Yogurt, Almonds, Avocado

## API Endpoints

Once deployed, the API provides:

- **GET /**: API information and endpoints
- **GET /health**: Health check (includes database status)
- **POST /api/foods**: Create new food item
- **GET /api/foods**: List all foods
- **GET /api/foods/{id}**: Get specific food
- **PUT /api/foods/{id}**: Update food item
- **DELETE /api/foods/{id}**: Delete food item

## Monitoring and Logs

### Container Logs
```bash
# API logs
docker logs -f ultra-api

# Database logs  
docker logs -f ultra-postgres
```

### Health Checks
```bash
# API health
curl http://localhost:8080/health

# Database health
docker exec ultra-postgres pg_isready -U ultra_user -d ultra_food_db
```

### Database Access
```bash
# Connect to PostgreSQL
docker exec -it ultra-postgres psql -U ultra_user -d ultra_food_db

# View foods table
docker exec ultra-postgres psql -U ultra_user -d ultra_food_db -c "SELECT * FROM foods LIMIT 5;"
```

## Updating the Application

### Update API Only
```bash
# Pull latest image
docker pull ghcr.io/yourusername/ultra:latest

# Stop and remove old container
docker stop ultra-api
docker rm ultra-api

# Start new container
docker run -d \
  --name ultra-api \
  --network ultra-network \
  --env-file .env.docker \
  -p 8080:8080 \
  --restart unless-stopped \
  ghcr.io/yourusername/ultra:latest
```

### Full Update
```bash
# Re-run deployment script
./deploy.sh
```

## Data Persistence

PostgreSQL data is stored in a Docker volume:
- **Volume Name**: `ultra-postgres-data`
- **Mount Point**: `/var/lib/postgresql/data`
- **Persistence**: Data survives container restarts and recreations

### Backup Database
```bash
# Create backup
docker exec ultra-postgres pg_dump -U ultra_user ultra_food_db > ultra_backup.sql

# Restore backup
docker exec -i ultra-postgres psql -U ultra_user -d ultra_food_db < ultra_backup.sql
```

## Troubleshooting

### Common Issues

1. **Port Already in Use**:
   ```bash
   # Check what's using port 8080
   lsof -i :8080
   # or change PORT in .env.docker
   ```

2. **Database Connection Failed**:
   ```bash
   # Check if PostgreSQL is ready
   docker exec ultra-postgres pg_isready -U ultra_user -d ultra_food_db
   
   # Check network connectivity
   docker exec ultra-api ping ultra-postgres
   ```

3. **Cannot Pull Image**:
   ```bash
   # Login to GitHub Container Registry
   docker login ghcr.io -u yourusername
   
   # Or build locally
   docker build -t ultra-api:latest .
   ```

### Reset Everything
```bash
# Stop all containers
docker stop ultra-api ultra-postgres

# Remove containers
docker rm ultra-api ultra-postgres

# Remove volume (⚠️ deletes all data)
docker volume rm ultra-postgres-data

# Remove network
docker network rm ultra-network

# Or use the cleanup option
./deploy.sh --cleanup
```

## Production Considerations

### Security
- Change default passwords in production
- Use SSL for database connections (`DB_SSL_MODE=require`)
- Implement proper authentication for API endpoints
- Use secrets management for sensitive data

### Performance
- Configure PostgreSQL connection pooling
- Monitor container resource usage
- Set appropriate restart policies
- Implement proper logging and monitoring

### High Availability
- Use Docker Swarm or Kubernetes for orchestration  
- Implement database replication
- Set up load balancing for API containers
- Configure proper health checks and auto-healing

## Support

For issues and questions:
1. Check container logs: `docker logs ultra-api` or `docker logs ultra-postgres`
2. Verify network connectivity between containers
3. Ensure environment variables are correctly set
4. Check GitHub Actions for image build status
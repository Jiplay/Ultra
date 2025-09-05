# Ultra API - Docker Guide

This guide covers how to run the Ultra API using Docker and Docker Compose.

## Prerequisites

- Docker Engine 20.10+
- Docker Compose 2.0+
- Make (optional, for convenience commands)

## Quick Start

### 1. Using Make (Recommended)

```bash
# Start development environment
make dev

# Or start in background
make dev-bg

# Start with additional development tools
make dev-full

# Stop everything
make down
```

### 2. Using Docker Compose Directly

```bash
# Start all services
docker-compose up --build

# Start in background
docker-compose up -d --build

# Stop services
docker-compose down
```

## Available Services

### Core Services

| Service | Port | Description |
|---------|------|-------------|
| **ultra-api** | 8080 | Main Go API application |
| **postgres** | 5432 | PostgreSQL database |
| **mongo** | 27017 | MongoDB database |

### Development Tools

| Service | Port | Description |
|---------|------|-------------|
| **adminer** | 8081 | PostgreSQL database manager |
| **mongo-express** | 8082 | MongoDB database manager |
| **mailcatcher** | 1080 | Email testing (dev only) |
| **redis-commander** | 8083 | Redis manager (dev only) |

## Environment Configurations

### Development

```bash
# Standard development
docker-compose up

# Development with extra tools
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up
```

### Production

```bash
# Production deployment
docker-compose -f docker-compose.prod.yml up -d
```

## Database Management

### Database Initialization

The databases are automatically initialized with:
- **PostgreSQL**: Schema creation and sample data
- **MongoDB**: Collections, indexes, and sample documents

### Reset Databases

```bash
# Warning: This deletes all data
make db-reset

# Or manually
docker-compose down -v
docker volume rm ultra_postgres_data ultra_mongo_data
docker-compose up -d postgres mongo
```

### Database Backups

```bash
# Backup both databases
make db-backup
```

### Access Databases

```bash
# PostgreSQL shell
make shell-db
# or
docker-compose exec postgres psql -U ultra_user -d ultra_db

# MongoDB shell
make shell-mongo
# or
docker-compose exec mongo mongosh -u ultra_user -p ultra_password --authenticationDatabase admin ultra
```

## Development Workflow

### 1. First Time Setup

```bash
# Clone repository and navigate to directory
git clone <repository-url>
cd Ultra

# Copy environment template
cp .env.example .env

# Start development environment
make dev
```

### 2. Daily Development

```bash
# Start services
make dev-bg

# View logs
make logs

# Access API shell for debugging
make shell-api

# Run tests
make test

# Stop when done
make down
```

### 3. Code Changes

The API container will automatically restart when code changes are detected (in development mode).

## API Endpoints

Once running, the API is available at: `http://localhost:8080`

### Health Check
```bash
curl http://localhost:8080/health
```

### API Documentation

- **Nutrition endpoints**: See `requests/nutrition.http`
- **Programs endpoints**: Available at `/api/v1/programs`
- **Users endpoints**: Available at `/api/v1/users`

## Troubleshooting

### Common Issues

#### Port Conflicts
```bash
# Check what's using the ports
lsof -i :8080
lsof -i :5432
lsof -i :27017

# Stop conflicting services or change ports in docker-compose.yml
```

#### Database Connection Issues
```bash
# Check database health
make health

# View database logs
make logs-db

# Restart databases
docker-compose restart postgres mongo
```

#### Container Build Issues
```bash
# Clean build without cache
make build-no-cache

# Clean everything and rebuild
make clean
make dev
```

### Viewing Logs

```bash
# All services
make logs

# Specific service
docker-compose logs -f ultra-api
docker-compose logs -f postgres
docker-compose logs -f mongo
```

### Performance Monitoring

```bash
# Container resource usage
docker stats

# Service health status
make health
docker-compose ps
```

## File Structure

```
Ultra/
├── Dockerfile                 # Multi-stage build for the API
├── docker-compose.yml        # Main development services
├── docker-compose.dev.yml    # Development overrides
├── docker-compose.prod.yml   # Production configuration
├── Makefile                  # Convenient commands
├── .env.docker              # Docker environment variables
├── .dockerignore            # Files to ignore in Docker builds
└── database/
    ├── postgres-init/       # PostgreSQL initialization scripts
    │   ├── 01-init.sql     # Schema creation
    │   └── 02-sample-data.sql # Sample data
    └── mongo-init/         # MongoDB initialization scripts
        └── 01-init.js      # Collections and sample data
```

## Environment Variables

### Application Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8080 | Application port |
| `GO_ENV` | development | Environment mode |
| `JWT_SECRET` | (required) | JWT signing secret |

### Database Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `POSTGRES_HOST` | postgres | PostgreSQL host |
| `POSTGRES_USER` | ultra_user | PostgreSQL username |
| `POSTGRES_PASSWORD` | ultra_password | PostgreSQL password |
| `POSTGRES_DB` | ultra_db | PostgreSQL database name |
| `MONGODB_URI` | (see .env.docker) | Full MongoDB connection string |

## Production Deployment

### Using Docker Compose

```bash
# Copy production environment
cp .env.production .env

# Deploy
docker-compose -f docker-compose.prod.yml up -d

# Check status
docker-compose -f docker-compose.prod.yml ps

# View logs
docker-compose -f docker-compose.prod.yml logs -f
```

### Security Considerations

1. **Environment Variables**: Never commit real passwords to version control
2. **Network**: Use custom networks and restrict external access
3. **Updates**: Regularly update base images and dependencies
4. **Volumes**: Backup database volumes regularly
5. **SSL/TLS**: Configure reverse proxy with SSL certificates

## Make Commands Reference

| Command | Description |
|---------|-------------|
| `make help` | Show all available commands |
| `make dev` | Start development environment |
| `make prod` | Start production environment |
| `make build` | Build application image |
| `make up/down` | Start/stop services |
| `make logs` | View service logs |
| `make shell-api` | Access API container shell |
| `make shell-db` | Access PostgreSQL shell |
| `make test` | Run tests in container |
| `make clean` | Clean up containers and volumes |
| `make health` | Check service health |

## Support

For issues and questions:
1. Check the troubleshooting section above
2. Review Docker and application logs
3. Ensure all prerequisites are met
4. Check for port conflicts
5. Verify environment configuration
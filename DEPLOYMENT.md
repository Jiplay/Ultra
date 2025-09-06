# Ultra API - Deployment Guide

This guide explains how to deploy the Ultra API using Docker containers in production.

## Architecture

The application consists of three main components:
- **Ultra API**: Go application serving REST endpoints
- **MongoDB**: Document database for user data
- **PostgreSQL**: Relational database for nutrition, programs, and meal data

## Prerequisites

- Docker and Docker Compose installed on your server
- GitHub Personal Access Token (for pulling private images)

## Quick Start (Production)

1. **Clone the repository** (or copy the deployment files):
   ```bash
   git clone <your-repo-url>
   cd ultra
   ```

2. **Set up environment variables**:
   ```bash
   cp .env.prod.example .env
   # Edit .env with your values
   ```

3. **Configure the GitHub repository**:
   - Replace `${GITHUB_REPOSITORY}` in `docker-compose.prod.yml` with your actual repo (e.g., `username/ultra`)
   - Or set the `GITHUB_REPOSITORY` environment variable

4. **Log in to GitHub Container Registry**:
   ```bash
   echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin
   ```

5. **Start the services**:
   ```bash
   docker-compose -f docker-compose.prod.yml up -d
   ```

6. **Verify deployment**:
   ```bash
   curl http://localhost:8080/health
   ```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GITHUB_REPOSITORY` | GitHub repo in format `username/repo` | Required |
| `JWT_SECRET` | Secret key for JWT token signing | `your-secret-key-here` |
| `MONGO_PASSWORD` | MongoDB password | `ultra_password` |
| `POSTGRES_PASSWORD` | PostgreSQL password | `ultra_password` |

## Files Overview

- **`Dockerfile`**: Multi-stage build for the Go application
- **`docker-compose.yml`**: For local development (uses locally built image)
- **`docker-compose.prod.yml`**: For production (pulls from GitHub registry)
- **`.dockerignore`**: Optimizes build context
- **`.github/workflows/docker-build.yml`**: CI/CD pipeline for building and pushing Docker images

## CI/CD Pipeline

The GitHub Actions workflow automatically:
1. Builds the Docker image on every push to `main`
2. Runs tests
3. Pushes the image with `latest` tag to GitHub Container Registry
4. Supports multi-architecture builds (AMD64 and ARM64)

## Local Development

For local development and testing:

1. **Build the application**:
   ```bash
   docker build -t ultra:latest .
   ```

2. **Start services**:
   ```bash
   docker-compose up -d
   ```

3. **Check logs**:
   ```bash
   docker-compose logs -f app
   ```

## Database Initialization

- **MongoDB**: Automatically creates the `ultra` database and sets up indexes
- **PostgreSQL**: Creates the `ultra_db` database and installs useful extensions

## Ports

- **Application**: 8080
- **MongoDB**: 27017
- **PostgreSQL**: 5432

## Health Check

The application provides a health check endpoint:
```bash
GET /health
```

Response:
```json
{
  "status": "ok",
  "message": "Ultra API is running"
}
```

## Scaling and Production Considerations

1. **Use environment-specific secrets**: Replace default passwords
2. **Reverse proxy**: Consider using nginx or traefik for SSL termination
3. **Database persistence**: Volumes are configured for data persistence
4. **Monitoring**: Add logging and monitoring solutions
5. **Backup**: Implement database backup strategies

## Troubleshooting

1. **Check container status**:
   ```bash
   docker-compose ps
   ```

2. **View logs**:
   ```bash
   docker-compose logs [service-name]
   ```

3. **Restart services**:
   ```bash
   docker-compose restart
   ```

4. **Clean restart**:
   ```bash
   docker-compose down
   docker-compose up -d
   ```

## API Endpoints

The application exposes various REST endpoints for:
- Nutrition management
- Program management  
- Meal planning
- User management

See application logs for complete endpoint list.
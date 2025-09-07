#!/bin/bash

# Ultra Food API Deployment Script
# This script deploys the API with PostgreSQL using standalone containers

set -e

# Configuration
NETWORK_NAME="ultra-network"
POSTGRES_CONTAINER="ultra-postgres"
API_CONTAINER="ultra"
POSTGRES_DATA_VOLUME="ultra-postgres-data"
GITHUB_REGISTRY="ghcr.io"
GITHUB_USER="jiplay"  # Replace with your GitHub username
REPO_NAME="ultra"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker and try again."
        exit 1
    fi
    print_success "Docker is running"
}

# Function to create Docker network
create_network() {
    if ! docker network ls | grep -q $NETWORK_NAME; then
        print_status "Creating Docker network: $NETWORK_NAME"
        docker network create $NETWORK_NAME
        print_success "Network created: $NETWORK_NAME"
    else
        print_status "Network already exists: $NETWORK_NAME"
    fi
}

# Function to deploy PostgreSQL
 # Edit deploy.sh, replace the deploy_postgres function:
  deploy_postgres() {
      print_status "Deploying PostgreSQL container..."

      if docker ps -a | grep -q $POSTGRES_CONTAINER; then
          print_status "Stopping existing PostgreSQL container..."
          docker stop $POSTGRES_CONTAINER || true
          docker rm $POSTGRES_CONTAINER || true
      fi

      # Use standard PostgreSQL image
      print_status "Starting PostgreSQL container..."
      docker run -d \
          --name $POSTGRES_CONTAINER \
          --network $NETWORK_NAME \
          -v $POSTGRES_DATA_VOLUME:/var/lib/postgresql/data \
          -e POSTGRES_USER=ultra_user \
          -e POSTGRES_PASSWORD=ultra_password \
          -e POSTGRES_DB=ultra_food_db \
          -p 5432:5432 \
          --restart unless-stopped \
          postgres:15-alpine

      print_success "PostgreSQL container started: $POSTGRES_CONTAINER"
  }

# Function to deploy API
deploy_api() {
    print_status "Deploying API container..."
    
    # Stop and remove existing container if it exists
    if docker ps -a | grep -q $API_CONTAINER; then
        print_status "Stopping existing API container..."
        docker stop $API_CONTAINER || true
        docker rm $API_CONTAINER || true
    fi
    
    # Pull latest API image from GitHub Container Registry
    print_status "Pulling latest API image from GitHub Container Registry..."
    if docker pull $GITHUB_REGISTRY/$GITHUB_USER/$REPO_NAME:latest; then
        print_success "Successfully pulled latest API image"
    else
        print_warning "Failed to pull from registry, building locally..."
        docker build -t ultra-api:latest .
    fi
    
    # Run API container
    print_status "Starting API container..."
    docker run -d \
        --name $API_CONTAINER \
        --network $NETWORK_NAME \
        --env-file .env.docker \
        -p 8080:8080 \
        --restart unless-stopped \
        --depends-on $POSTGRES_CONTAINER \
        $GITHUB_REGISTRY/$GITHUB_USER/$REPO_NAME:latest || \
    docker run -d \
        --name $API_CONTAINER \
        --network $NETWORK_NAME \
        --env-file .env.docker \
        -p 8080:8080 \
        --restart unless-stopped \
        ultra-api:latest
    
    print_success "API container started: $API_CONTAINER"
}

# Function to run health checks
health_check() {
    print_status "Running health checks..."
    
    # Check PostgreSQL
    print_status "Checking PostgreSQL health..."
    if docker exec $POSTGRES_CONTAINER pg_isready -U ultra_user -d ultra_food_db > /dev/null 2>&1; then
        print_success "✓ PostgreSQL is healthy"
    else
        print_error "✗ PostgreSQL health check failed"
        return 1
    fi
    
    # Check API
    print_status "Checking API health..."
    sleep 5  # Give API time to start
    
    for i in {1..10}; do
        if curl -f http://localhost:8080/health > /dev/null 2>&1; then
            print_success "✓ API is healthy"
            break
        fi
        if [ $i -eq 10 ]; then
            print_error "✗ API health check failed"
            return 1
        fi
        echo -n "."
        sleep 2
    done
}

# Function to display status
show_status() {
    echo
    print_success "🚀 Deployment completed successfully!"
    echo
    echo "Container Status:"
    docker ps --filter "name=ultra-" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
    echo
    echo "Available Services:"
    echo "  • API: http://localhost:8080"
    echo "  • Health Check: http://localhost:8080/health"
    echo "  • PostgreSQL: localhost:5432"
    echo
    echo "Useful Commands:"
    echo "  • View API logs: docker logs -f $API_CONTAINER"
    echo "  • View DB logs: docker logs -f $POSTGRES_CONTAINER"
    echo "  • Access DB: docker exec -it $POSTGRES_CONTAINER psql -U ultra_user -d ultra_food_db"
    echo "  • Stop all: docker stop $API_CONTAINER $POSTGRES_CONTAINER"
    echo
}

# Function to cleanup (if needed)
cleanup() {
    if [ "$1" = "--cleanup" ]; then
        print_warning "Cleaning up existing containers and volumes..."
        docker stop $API_CONTAINER $POSTGRES_CONTAINER 2>/dev/null || true
        docker rm $API_CONTAINER $POSTGRES_CONTAINER 2>/dev/null || true
        docker volume rm $POSTGRES_DATA_VOLUME 2>/dev/null || true
        docker network rm $NETWORK_NAME 2>/dev/null || true
        print_success "Cleanup completed"
        exit 0
    fi
}

# Main deployment function
main() {
    echo "🍎 Ultra Food API Deployment"
    echo "============================"
    echo
    
    # Handle cleanup option
    cleanup "$1"
    
    # Check prerequisites
    check_docker
    
    # Check for required files
    if [ ! -f ".env.docker" ]; then
        print_error ".env.docker file not found! Please create it from .env.example"
        exit 1
    fi
    
#    if [ ! -d "postgres" ]; then
#        print_error "postgres/ directory not found! Please ensure PostgreSQL files are present"
#        exit 1
#    fi
    
    # Deploy services
    create_network
    deploy_postgres
    deploy_api
    health_check
    show_status
}

# Handle script arguments
if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    echo "Usage: $0 [--cleanup]"
    echo
    echo "Options:"
    echo "  --cleanup    Stop and remove all containers and volumes"
    echo "  --help, -h   Show this help message"
    echo
    exit 0
fi

# Run main function
main "$1"
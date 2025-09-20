#!/bin/bash

# Ultra Food API Development Deployment Script
# This script builds a new image each time and deploys it for development

set -e

# Configuration
NETWORK_NAME="ultra-dev-network"
POSTGRES_CONTAINER="ultra-dev-postgres"
API_CONTAINER="ultra-dev"
POSTGRES_DATA_VOLUME="ultra-dev-postgres-data"
IMAGE_NAME="ultra-dev"
IMAGE_TAG="latest"

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
deploy_postgres() {
    print_status "Deploying PostgreSQL container..."

    if docker ps -a | grep -q $POSTGRES_CONTAINER; then
        print_status "Stopping existing PostgreSQL container..."
        docker stop $POSTGRES_CONTAINER || true
        docker rm $POSTGRES_CONTAINER || true
    fi

    # Start PostgreSQL container with initialization scripts
    print_status "Starting PostgreSQL container..."
    docker run -d \
        --name $POSTGRES_CONTAINER \
        --network $NETWORK_NAME \
        -v $POSTGRES_DATA_VOLUME:/var/lib/postgresql/data \
        -v "$(pwd)/postgres/init:/docker-entrypoint-initdb.d" \
        -e POSTGRES_USER=ultra_user \
        -e POSTGRES_PASSWORD=ultra_password \
        -e POSTGRES_DB=ultra_food_db \
        -p 5433:5432 \
        --restart unless-stopped \
        postgres:15-alpine

    print_success "PostgreSQL container started: $POSTGRES_CONTAINER"

    # Wait for PostgreSQL to be ready
    print_status "Waiting for PostgreSQL to be ready..."
    sleep 10
    for i in {1..30}; do
        if docker exec $POSTGRES_CONTAINER pg_isready -U ultra_user -d ultra_food_db > /dev/null 2>&1; then
            print_success "PostgreSQL is ready"
            break
        fi
        if [ $i -eq 30 ]; then
            print_error "PostgreSQL failed to start properly"
            exit 1
        fi
        echo -n "."
        sleep 2
    done
}

# Function to build and deploy API
build_and_deploy_api() {
    print_status "Building new API image..."

    # Remove old image if it exists
    if docker images | grep -q "$IMAGE_NAME.*$IMAGE_TAG"; then
        print_status "Removing old API image..."
        docker rmi $IMAGE_NAME:$IMAGE_TAG || true
    fi

    # Build new image with timestamp tag for cache busting
    TIMESTAMP=$(date +%s)
    print_status "Building fresh API image with timestamp $TIMESTAMP..."
    docker build --no-cache -t $IMAGE_NAME:$IMAGE_TAG -t $IMAGE_NAME:$TIMESTAMP .

    print_success "Successfully built new API image"

    # Stop and remove existing container if it exists
    if docker ps -a | grep -q $API_CONTAINER; then
        print_status "Stopping existing API container..."
        docker stop $API_CONTAINER || true
        docker rm $API_CONTAINER || true
    fi

    # Run API container
    print_status "Starting API container..."
    docker run -d \
        --name $API_CONTAINER \
        --network $NETWORK_NAME \
        --env-file .env \
        -p 8081:8080 \
        --restart unless-stopped \
        $IMAGE_NAME:$IMAGE_TAG

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

    for i in {1..15}; do
        if curl -f http://localhost:8081/ > /dev/null 2>&1; then
            print_success "✓ API is healthy"
            break
        fi
        if [ $i -eq 15 ]; then
            print_error "✗ API health check failed"
            echo "API logs:"
            docker logs --tail 20 $API_CONTAINER
            return 1
        fi
        echo -n "."
        sleep 2
    done
}

# Function to display status
show_status() {
    echo
    print_success "🚀 Development deployment completed successfully!"
    echo
    echo "Container Status:"
    docker ps --filter "name=ultra-dev" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
    echo
    echo "Available Services:"
    echo "  • API: http://localhost:8081"
    echo "  • API Info: http://localhost:8081/"
    echo "  • PostgreSQL: localhost:5433 (external), ultra-dev-postgres:5432 (internal)"
    echo
    echo "Useful Commands:"
    echo "  • View API logs: docker logs -f $API_CONTAINER"
    echo "  • Follow API logs: docker logs -f $API_CONTAINER"
    echo "  • View DB logs: docker logs -f $POSTGRES_CONTAINER"
    echo "  • Access DB: docker exec -it $POSTGRES_CONTAINER psql -U ultra_user -d ultra_food_db"
    echo "  • Stop all: docker stop $API_CONTAINER $POSTGRES_CONTAINER"
    echo "  • Test API: ./scripts/test-api.sh"
    echo
    echo "Development Tips:"
    echo "  • Run this script again to rebuild and redeploy with latest changes"
    echo "  • Use 'docker logs -f $API_CONTAINER' to watch for errors"
    echo "  • Database data persists between deployments"
    echo
}

# Function to cleanup (if needed)
cleanup() {
    if [ "$1" = "--cleanup" ]; then
        print_warning "Cleaning up existing development containers and volumes..."
        docker stop $API_CONTAINER $POSTGRES_CONTAINER 2>/dev/null || true
        docker rm $API_CONTAINER $POSTGRES_CONTAINER 2>/dev/null || true
        docker volume rm $POSTGRES_DATA_VOLUME 2>/dev/null || true
        docker network rm $NETWORK_NAME 2>/dev/null || true
        # Clean up dev images
        docker rmi $IMAGE_NAME:$IMAGE_TAG 2>/dev/null || true
        docker image prune -f
        print_success "Cleanup completed"
        exit 0
    fi
}

# Function to quick restart (just API)
quick_restart() {
    if [ "$1" = "--quick" ]; then
        print_status "Quick restart: rebuilding and restarting API only..."
        build_and_deploy_api
        health_check
        show_status
        exit 0
    fi
}

# Main deployment function
main() {
    echo "🍎 Ultra Food API Development Deployment"
    echo "========================================"
    echo

    # Handle options
    cleanup "$1"
    quick_restart "$1"

    # Check prerequisites
    check_docker

    # Check for required files
    if [ ! -f ".env" ]; then
        print_error ".env file not found! Please create it from .env.example"
        exit 1
    fi

    if [ ! -f "Dockerfile" ]; then
        print_error "Dockerfile not found! Please ensure Dockerfile exists"
        exit 1
    fi

    # Deploy services
    create_network
    deploy_postgres
    build_and_deploy_api
    health_check
    show_status
}

# Handle script arguments
if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    echo "Usage: $0 [options]"
    echo
    echo "Options:"
    echo "  --cleanup    Stop and remove all development containers and volumes"
    echo "  --quick      Quick restart: rebuild and restart API only (keeps DB running)"
    echo "  --help, -h   Show this help message"
    echo
    echo "This script builds a fresh Docker image every time it runs,"
    echo "making it perfect for development where you want to test"
    echo "your latest changes quickly."
    echo
    exit 0
fi

# Run main function
main "$1"
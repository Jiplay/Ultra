#!/bin/bash

# Ultra API Deployment Script
set -e

echo "🚀 Starting Ultra API deployment..."

# Configuration
REGISTRY="ghcr.io"
USERNAME="Jiplay"  # Update with your GitHub username
REPO="Ultra"
TAG="${1:-latest}"
IMAGE="${REGISTRY}/${USERNAME}/${REPO}:${TAG}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    print_error "Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if .env file exists
if [ ! -f .env ]; then
    print_warning ".env file not found. Creating from .env.production template..."
    cp .env.production .env
    print_warning "Please edit .env with your production values before running again."
    exit 1
fi

# Pull the latest image
print_status "Pulling latest image: ${IMAGE}"
docker pull "${IMAGE}"

# Stop existing containers
print_status "Stopping existing containers..."
docker-compose -f docker-compose.prod.yml down

# Start services
print_status "Starting Ultra API services..."
docker-compose -f docker-compose.prod.yml up -d

# Wait for services to be healthy
print_status "Waiting for services to be healthy..."
sleep 10

# Check if the API is responding
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    print_status "Ultra API is running and healthy!"
    echo "🎉 Deployment completed successfully!"
    echo "API is available at: http://localhost:8080"
    echo "Health check: http://localhost:8080/health"
else
    print_error "Ultra API is not responding. Check the logs:"
    docker-compose -f docker-compose.prod.yml logs ultra-api
    exit 1
fi

# Show running containers
print_status "Running containers:"
docker-compose -f docker-compose.prod.yml ps
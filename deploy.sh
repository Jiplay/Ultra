#!/bin/bash

# Ultra API Complete Deployment Script
set -e

echo "🚀 Starting Ultra API deployment..."

# Configuration
REGISTRY="ghcr.io"
USERNAME="jiplay"
REPO="ultra"
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
    print_warning ".env file not found. Please create it first (see HowToDeploy.md)"
    exit 1
fi

# Create necessary directories
mkdir -p postgres-init mongo-init ssl

# Pull the latest image
print_status "Pulling latest image: ${IMAGE}"
docker pull "${IMAGE}" || {
    print_error "Failed to pull image. Make sure you're logged in to GitHub Container Registry:"
    echo "docker login ghcr.io -u jiplay"
    exit 1
}

# Stop existing containers
print_status "Stopping existing containers..."
docker-compose -f docker-compose.prod.yml down 2>/dev/null || true

# Remove old volumes if requested
if [ "$2" = "--fresh" ]; then
    print_warning "Removing existing data volumes..."
    docker volume rm ultra_postgres_data ultra_mongodb_data 2>/dev/null || true
fi

# Start services
print_status "Starting Ultra API stack..."
docker-compose -f docker-compose.prod.yml up -d

# Wait for services to start
print_status "Waiting for services to start..."
sleep 15

# Check PostgreSQL
print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}
BLUE='\033[0;34m'

print_info "Checking PostgreSQL..."
if docker-compose -f docker-compose.prod.yml exec postgres pg_isready -U ultra_user > /dev/null 2>&1; then
    print_status "PostgreSQL is healthy"
else
    print_error "PostgreSQL is not ready"
fi

# Check MongoDB
print_info "Checking MongoDB..."
if docker-compose -f docker-compose.prod.yml exec mongodb mongosh --eval "db.adminCommand('ping')" > /dev/null 2>&1; then
    print_status "MongoDB is healthy"
else
    print_error "MongoDB is not ready"
fi

# Check API health
print_info "Checking Ultra API..."
sleep 5
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    print_status "Ultra API is running and healthy!"
elif curl -f http://localhost/health > /dev/null 2>&1; then
    print_status "Ultra API is running behind Nginx!"
else
    print_error "Ultra API is not responding. Checking logs..."
    docker-compose -f docker-compose.prod.yml logs --tail=50 ultra-api
    exit 1
fi

# Show running containers
print_status "Running containers:"
docker-compose -f docker-compose.prod.yml ps

# Show useful information
echo ""
echo "🎉 Deployment completed successfully!"
echo ""
echo "📍 Endpoints:"
echo "   Health: http://$(curl -s ifconfig.me || echo 'your-server-ip')/health"
echo "   API:    http://$(curl -s ifconfig.me || echo 'your-server-ip')/api/v1/"
echo ""
echo "🔧 Management commands:"
echo "   View logs: docker-compose -f docker-compose.prod.yml logs -f"
echo "   Stop all:  docker-compose -f docker-compose.prod.yml down"
echo "   Restart:   docker-compose -f docker-compose.prod.yml restart"
echo ""
echo "💾 Database access:"
echo "   PostgreSQL: docker-compose -f docker-compose.prod.yml exec postgres psql -U ultra_user -d ultra_db"
echo "   MongoDB:    docker-compose -f docker-compose.prod.yml exec mongodb mongosh -u ultra_user -p"
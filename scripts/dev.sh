#!/bin/bash
set -e

# Ultra API Development Helper Script

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Ultra API Development Environment${NC}"
echo "======================================"

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo -e "${YELLOW}Warning: .env file not found!${NC}"
    echo "Creating .env from .env.example..."
    cp .env.example .env
    echo -e "${GREEN}✓ Created .env file. Please review and update it as needed.${NC}"
fi

# Function to check if Docker is running
check_docker() {
    if ! docker info >/dev/null 2>&1; then
        echo -e "${RED}Error: Docker is not running. Please start Docker first.${NC}"
        exit 1
    fi
}

# Function to start development environment
start_dev() {
    echo -e "${GREEN}Starting development environment...${NC}"
    check_docker
    docker-compose -f docker-compose.yml -f docker-compose.dev.yml up --build
}

# Function to start production environment (using pre-built image)
start_prod() {
    echo -e "${GREEN}Starting production environment...${NC}"
    check_docker
    docker-compose -f docker-compose.yml -f docker-compose.prod.yml up
}

# Function to stop all services
stop_all() {
    echo -e "${GREEN}Stopping all services...${NC}"
    docker-compose -f docker-compose.yml -f docker-compose.dev.yml down
    docker-compose -f docker-compose.yml -f docker-compose.prod.yml down
}

# Function to clean up everything
cleanup() {
    echo -e "${YELLOW}Cleaning up containers, networks, and volumes...${NC}"
    docker-compose -f docker-compose.yml -f docker-compose.dev.yml down -v --remove-orphans
    docker-compose -f docker-compose.yml -f docker-compose.prod.yml down -v --remove-orphans
    docker system prune -f
}

# Function to show logs
show_logs() {
    service=${1:-ultra-api}
    echo -e "${GREEN}Showing logs for ${service}...${NC}"
    docker-compose -f docker-compose.yml -f docker-compose.dev.yml logs -f "$service"
}

# Function to run database migrations
migrate() {
    echo -e "${GREEN}Running database migrations...${NC}"
    # Add your migration commands here
    echo "No migrations defined yet"
}

# Function to show status
status() {
    echo -e "${GREEN}Service Status:${NC}"
    docker-compose -f docker-compose.yml -f docker-compose.dev.yml ps
}

# Main menu
case "${1:-help}" in
    "dev"|"start")
        start_dev
        ;;
    "prod"|"production")
        start_prod
        ;;
    "stop")
        stop_all
        ;;
    "clean"|"cleanup")
        cleanup
        ;;
    "logs")
        show_logs "$2"
        ;;
    "migrate")
        migrate
        ;;
    "status")
        status
        ;;
    "help"|*)
        echo "Usage: $0 {dev|prod|stop|clean|logs|migrate|status}"
        echo ""
        echo "Commands:"
        echo "  dev, start    - Start development environment with hot reload"
        echo "  prod          - Start production environment with pre-built image"
        echo "  stop          - Stop all running services"
        echo "  clean         - Clean up containers, networks, and volumes"
        echo "  logs [service]- Show logs (default: ultra-api)"
        echo "  migrate       - Run database migrations"
        echo "  status        - Show status of all services"
        echo "  help          - Show this help message"
        echo ""
        echo "Examples:"
        echo "  $0 dev                    # Start development environment"
        echo "  $0 logs postgres          # Show PostgreSQL logs"
        echo "  $0 clean                  # Clean up everything"
        ;;
esac
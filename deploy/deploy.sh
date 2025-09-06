#!/bin/bash
set -e

# Ultra API Production Deployment Script

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Ultra API Production Deployment${NC}"
echo "================================="

# Function to check if Docker is running
check_docker() {
    if ! docker info >/dev/null 2>&1; then
        echo -e "${RED}Error: Docker is not running. Please start Docker first.${NC}"
        exit 1
    fi
}

# Function to check if .env exists
check_env() {
    if [ ! -f ".env" ]; then
        echo -e "${YELLOW}Warning: .env file not found!${NC}"
        echo "Creating .env from template..."
        cp .env.template .env
        echo -e "${RED}IMPORTANT: Please edit .env file and update the passwords and secrets!${NC}"
        echo "Required changes:"
        echo "  - JWT_SECRET: Generate with 'openssl rand -base64 64'"
        echo "  - POSTGRES_PASSWORD: Generate with 'openssl rand -base64 32'"
        echo "  - MONGO_PASSWORD: Generate with 'openssl rand -base64 32'"
        echo ""
        read -p "Press Enter after updating .env file..."
    fi
}

# Function to pull latest image
pull_latest() {
    echo -e "${GREEN}Pulling latest Ultra API image...${NC}"
    docker pull ghcr.io/jiplay/ultra:latest
}

# Function to start services
start_services() {
    echo -e "${GREEN}Starting Ultra API services...${NC}"
    docker-compose up -d
}

# Function to show status
show_status() {
    echo -e "${GREEN}Service Status:${NC}"
    docker-compose ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"
    echo ""
}

# Function to show logs
show_logs() {
    service=${1:-}
    if [ -z "$service" ]; then
        echo -e "${GREEN}Showing all logs (Ctrl+C to exit):${NC}"
        docker-compose logs -f
    else
        echo -e "${GREEN}Showing logs for ${service} (Ctrl+C to exit):${NC}"
        docker-compose logs -f "$service"
    fi
}

# Function to stop services
stop_services() {
    echo -e "${YELLOW}Stopping Ultra API services...${NC}"
    docker-compose down
}

# Function to update and restart
update_and_restart() {
    echo -e "${GREEN}Updating Ultra API...${NC}"
    pull_latest
    docker-compose up -d
    echo -e "${GREEN}✓ Update complete!${NC}"
}

# Function to backup databases
backup_databases() {
    echo -e "${GREEN}Creating database backups...${NC}"
    mkdir -p ./backups
    
    # PostgreSQL backup
    docker-compose exec -T postgres pg_dump -U ultra_user ultra_db > "./backups/postgres_backup_$(date +%Y%m%d_%H%M%S).sql"
    
    # MongoDB backup  
    docker-compose exec -T mongo mongodump --uri="mongodb://ultra_user:${MONGO_PASSWORD}@localhost:27017/ultra?authSource=admin" --archive > "./backups/mongo_backup_$(date +%Y%m%d_%H%M%S).archive"
    
    echo -e "${GREEN}✓ Backups created in ./backups/${NC}"
}

# Function to check health
check_health() {
    echo -e "${GREEN}Checking service health...${NC}"
    echo ""
    
    # Check API health
    echo "API Health:"
    if curl -s http://localhost:8080/health >/dev/null; then
        echo -e "${GREEN}✓ API is responding${NC}"
    else
        echo -e "${RED}✗ API is not responding${NC}"
    fi
    
    # Check database containers
    echo ""
    echo "Database Health:"
    if docker-compose exec -T postgres pg_isready -U ultra_user -d ultra_db >/dev/null 2>&1; then
        echo -e "${GREEN}✓ PostgreSQL is ready${NC}"
    else
        echo -e "${RED}✗ PostgreSQL is not ready${NC}"
    fi
    
    if docker-compose exec -T mongo mongosh --eval "db.adminCommand('ping')" --quiet >/dev/null 2>&1; then
        echo -e "${GREEN}✓ MongoDB is ready${NC}"
    else
        echo -e "${RED}✗ MongoDB is not ready${NC}"
    fi
}

# Function to clean up
cleanup() {
    echo -e "${YELLOW}Cleaning up containers, networks, and volumes...${NC}"
    echo -e "${RED}WARNING: This will delete all data!${NC}"
    read -p "Are you sure? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker-compose down -v --remove-orphans
        docker system prune -f
        echo -e "${GREEN}✓ Cleanup complete${NC}"
    else
        echo "Cleanup cancelled"
    fi
}

# Main menu
case "${1:-help}" in
    "start"|"up")
        check_docker
        check_env
        pull_latest
        start_services
        echo ""
        echo -e "${GREEN}✓ Ultra API is starting up!${NC}"
        echo ""
        echo "Access points:"
        echo "  - API: http://localhost:8080"
        echo "  - Health: http://localhost:8080/health"
        echo ""
        echo "Useful commands:"
        echo "  - Check status: ./deploy.sh status"
        echo "  - View logs: ./deploy.sh logs"
        echo "  - Check health: ./deploy.sh health"
        ;;
    "stop"|"down")
        stop_services
        ;;
    "restart")
        stop_services
        sleep 2
        start_services
        ;;
    "update")
        check_docker
        update_and_restart
        ;;
    "status")
        show_status
        ;;
    "logs")
        show_logs "$2"
        ;;
    "health")
        check_health
        ;;
    "backup")
        backup_databases
        ;;
    "clean")
        cleanup
        ;;
    "help"|*)
        echo "Usage: $0 {start|stop|restart|update|status|logs|health|backup|clean}"
        echo ""
        echo "Commands:"
        echo "  start, up     - Start Ultra API (pulls latest image)"
        echo "  stop, down    - Stop all services"
        echo "  restart       - Restart all services"
        echo "  update        - Pull latest image and restart"
        echo "  status        - Show service status"
        echo "  logs [service]- Show logs (default: all services)"
        echo "  health        - Check API and database health"
        echo "  backup        - Backup databases"
        echo "  clean         - Clean up everything (DELETES DATA!)"
        echo "  help          - Show this help message"
        echo ""
        echo "Examples:"
        echo "  $0 start                  # Deploy Ultra API"
        echo "  $0 logs ultra-api         # Show API logs"
        echo "  $0 update                 # Update to latest version"
        ;;
esac
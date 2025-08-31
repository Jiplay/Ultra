# 🚀 Ultra API - Complete Server Deployment Guide

This guide will help you deploy the entire Ultra API stack (API + PostgreSQL + MongoDB) on your server using Docker.

## 📋 Prerequisites

- A server with Docker and Docker Compose installed
- SSH access to your server
- A domain name (optional, for SSL)
- At least 2GB RAM and 20GB storage

## 🛠️ Server Setup

### 1. Install Docker and Docker Compose

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add your user to docker group
sudo usermod -aG docker $USER

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Log out and back in, or run:
newgrp docker

# Verify installation
docker --version
docker-compose --version
```

### 2. Create Application Directory

```bash
# Create app directory
sudo mkdir -p /opt/ultra
sudo chown $USER:$USER /opt/ultra
cd /opt/ultra
```

## 📁 Deployment Files Setup

### 3. Create Production Docker Compose

Create `/opt/ultra/docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  ultra-api:
    image: ghcr.io/jiplay/ultra:latest
    container_name: ultra_api
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - MONGODB_URI=mongodb://ultra_user:${MONGO_PASSWORD}@mongodb:27017/ultra?authSource=admin
      - POSTGRES_HOST=postgres
      - POSTGRES_PORT=5432
      - POSTGRES_USER=ultra_user
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
      - POSTGRES_DB=ultra_db
      - POSTGRES_SSL=disable
      - PORT=8080
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      postgres:
        condition: service_healthy
      mongodb:
        condition: service_healthy
    networks:
      - ultra-network
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  postgres:
    image: postgres:15
    container_name: ultra_postgres
    restart: unless-stopped
    environment:
      POSTGRES_USER: ultra_user
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_DB: ultra_db
      POSTGRES_INITDB_ARGS: "--encoding=UTF-8"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./postgres-init:/docker-entrypoint-initdb.d
    networks:
      - ultra-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ultra_user -d ultra_db"]
      interval: 10s
      timeout: 5s
      retries: 5

  mongodb:
    image: mongo:7
    container_name: ultra_mongodb
    restart: unless-stopped
    environment:
      MONGO_INITDB_ROOT_USERNAME: ultra_user
      MONGO_INITDB_ROOT_PASSWORD: ${MONGO_PASSWORD}
      MONGO_INITDB_DATABASE: ultra
    volumes:
      - mongodb_data:/data/db
      - ./mongo-init:/docker-entrypoint-initdb.d
    networks:
      - ultra-network
    healthcheck:
      test: ["CMD", "mongosh", "--eval", "db.adminCommand('ping')"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Optional: Nginx reverse proxy with SSL
  nginx:
    image: nginx:alpine
    container_name: ultra_nginx
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
      - nginx_logs:/var/log/nginx
    depends_on:
      - ultra-api
    networks:
      - ultra-network

networks:
  ultra-network:
    driver: bridge

volumes:
  postgres_data:
    driver: local
  mongodb_data:
    driver: local
  nginx_logs:
    driver: local
```

### 4. Create Environment File

Create `/opt/ultra/.env`:

```bash
# Generate secure passwords
POSTGRES_PASSWORD=$(openssl rand -base64 32)
MONGO_PASSWORD=$(openssl rand -base64 32)
JWT_SECRET=$(openssl rand -base64 64)

# Save to .env file
cat > .env << EOF
# Database Passwords (Generated)
POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
MONGO_PASSWORD=${MONGO_PASSWORD}

# JWT Secret
JWT_SECRET=${JWT_SECRET}

# Optional: Your domain for SSL
DOMAIN=yourdomain.com
EOF

echo "Environment file created with secure passwords!"
echo "POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}"
echo "MONGO_PASSWORD: ${MONGO_PASSWORD}"
echo "Save these passwords securely!"
```

### 5. Create Nginx Configuration

Create `/opt/ultra/nginx.conf`:

```nginx
events {
    worker_connections 1024;
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    # Logging
    access_log /var/log/nginx/access.log;
    error_log /var/log/nginx/error.log;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 10240;
    gzip_proxied expired no-cache no-store private must-revalidate auth;
    gzip_types text/plain text/css text/xml text/javascript application/javascript application/xml+rss application/json;

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;

    upstream ultra_api {
        server ultra-api:8080;
    }

    # HTTP server (redirect to HTTPS if SSL is configured)
    server {
        listen 80;
        server_name _;

        # Health check endpoint (always allow)
        location /health {
            proxy_pass http://ultra_api/health;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # API endpoints
        location / {
            limit_req zone=api burst=20 nodelay;
            
            proxy_pass http://ultra_api;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            
            # CORS headers
            add_header Access-Control-Allow-Origin *;
            add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS";
            add_header Access-Control-Allow-Headers "Content-Type, X-User-ID";
            
            # Handle OPTIONS requests
            if ($request_method = 'OPTIONS') {
                return 200;
            }
            
            # Security headers
            add_header X-Frame-Options DENY;
            add_header X-Content-Type-Options nosniff;
            add_header X-XSS-Protection "1; mode=block";
            
            # Timeouts
            proxy_connect_timeout 60s;
            proxy_send_timeout 60s;
            proxy_read_timeout 60s;
        }
    }
}
```

### 6. Create Deployment Script

Create `/opt/ultra/deploy.sh`:

```bash
#!/bin/bash

# Ultra API Complete Deployment Script
set -e

echo "🚀 Starting Ultra API deployment..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}✓${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Configuration
REGISTRY="ghcr.io"
USERNAME="jiplay"
REPO="ultra"
TAG="${1:-latest}"
IMAGE="${REGISTRY}/${USERNAME}/${REPO}:${TAG}"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    print_error "Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if .env file exists
if [ ! -f .env ]; then
    print_warning ".env file not found. Please create it first (see step 4 in HowToDeploy.md)"
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

# Wait for services to be healthy
print_status "Waiting for services to start..."
sleep 15

# Check PostgreSQL
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
```

Make it executable:
```bash
chmod +x deploy.sh
```

## 🎯 Deployment Process

### 7. Deploy the Application

```bash
cd /opt/ultra

# First deployment
./deploy.sh latest

# To deploy with fresh databases (⚠️ This will delete all data!)
./deploy.sh latest --fresh
```

### 8. Verify Deployment

```bash
# Check all services are running
docker-compose -f docker-compose.prod.yml ps

# Check API health
curl http://localhost/health
curl http://localhost/api/v1/users

# Check logs
docker-compose -f docker-compose.prod.yml logs -f ultra-api
```

## 🔒 Optional: SSL Setup

### 9. Configure SSL with Let's Encrypt

```bash
# Install Certbot
sudo apt install certbot

# Get SSL certificate (replace with your domain)
sudo certbot certonly --standalone -d yourdomain.com

# Copy certificates
sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem /opt/ultra/ssl/cert.pem
sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem /opt/ultra/ssl/key.pem
sudo chown $USER:$USER /opt/ultra/ssl/*.pem

# Update nginx.conf to enable HTTPS (add server block for port 443)
# Then restart nginx
docker-compose -f docker-compose.prod.yml restart nginx
```

## 📊 Monitoring and Maintenance

### 10. Daily Operations

```bash
# View logs
docker-compose -f docker-compose.prod.yml logs -f ultra-api
docker-compose -f docker-compose.prod.yml logs -f postgres
docker-compose -f docker-compose.prod.yml logs -f mongodb

# Database backups
# PostgreSQL
docker-compose -f docker-compose.prod.yml exec postgres pg_dump -U ultra_user ultra_db > backup_$(date +%Y%m%d).sql

# MongoDB
docker-compose -f docker-compose.prod.yml exec mongodb mongodump --uri="mongodb://ultra_user:password@localhost:27017/ultra" --out /backup

# Update to new version
docker pull ghcr.io/jiplay/ultra:latest
docker-compose -f docker-compose.prod.yml up -d ultra-api

# Restart services
docker-compose -f docker-compose.prod.yml restart
```

## 🚨 Troubleshooting

### Common Issues

1. **API not responding**: Check container logs
   ```bash
   docker-compose -f docker-compose.prod.yml logs ultra-api
   ```

2. **Database connection failed**: Verify database containers are healthy
   ```bash
   docker-compose -f docker-compose.prod.yml exec postgres pg_isready -U ultra_user
   ```

3. **Out of disk space**: Clean up old Docker images
   ```bash
   docker system prune -a
   ```

4. **Port conflicts**: Make sure ports 80, 443, 5432, 27017 are not in use
   ```bash
   sudo netstat -tulpn | grep -E ':(80|443|5432|27017)\s'
   ```

## 📈 Scaling and Performance

- **Increase resources**: Modify `docker-compose.prod.yml` to add resource limits
- **Multiple API instances**: Use Docker Swarm or add more `ultra-api` replicas
- **Database optimization**: Tune PostgreSQL and MongoDB configurations
- **Monitoring**: Add Prometheus + Grafana stack for monitoring

Your Ultra API is now fully deployed with all databases on your server! 🎉
# Ultra API - Server Deployment

This directory contains everything needed to deploy Ultra API to a production server **without cloning the entire repository**.

## What's Included

```
deploy/
├── docker-compose.yml    # Production Docker configuration
├── .env.template        # Environment variables template
├── deploy.sh           # Deployment script
├── README.md           # This file
└── database/           # Database initialization scripts
    ├── postgres-init/
    │   ├── 01-init.sql
    │   └── 02-sample-data.sql
    └── mongo-init/
        └── 01-init.js
```

## Server Requirements

- Docker Engine 20.10+
- Docker Compose 2.0+
- Ports 8080, 5432, 27017 available (or configure different ports)
- At least 2GB RAM recommended

## Quick Deployment

### 1. Copy Files to Server

Upload this entire `deploy/` directory to your server:

```bash
# Option 1: Using SCP
scp -r deploy/ user@your-server:/path/to/ultra/

# Option 2: Using rsync
rsync -avz deploy/ user@your-server:/path/to/ultra/

# Option 3: Download directly on server
wget -O ultra-deploy.tar.gz https://github.com/jiplay/ultra/archive/main.tar.gz
tar -xzf ultra-deploy.tar.gz --strip-components=2 ultra-main/deploy/
```

### 2. Configure Environment

```bash
cd /path/to/ultra/deploy/
cp .env.template .env
nano .env  # Edit the configuration
```

**Important**: Update these values in `.env`:
- `JWT_SECRET` - Generate with: `openssl rand -base64 64`
- `POSTGRES_PASSWORD` - Generate with: `openssl rand -base64 32`  
- `MONGO_PASSWORD` - Generate with: `openssl rand -base64 32`

### 3. Deploy

```bash
./deploy.sh start
```

That's it! The script will:
1. Pull the latest Ultra API image from GitHub Container Registry
2. Start PostgreSQL and MongoDB with your data
3. Start the Ultra API application
4. Set up health checks and logging

## Management Commands

```bash
# Check status
./deploy.sh status

# View logs
./deploy.sh logs           # All services
./deploy.sh logs ultra-api # Just API

# Update to latest version
./deploy.sh update

# Check health
./deploy.sh health

# Backup databases
./deploy.sh backup

# Stop services
./deploy.sh stop

# Restart services  
./deploy.sh restart
```

## Accessing Your API

After deployment, your API will be available at:
- **API**: http://your-server:8080
- **Health Check**: http://your-server:8080/health

## Environment Configuration

### Basic Configuration

```bash
# Application
PORT=8080
GO_ENV=production
LOG_LEVEL=info

# Security (CHANGE THESE!)
JWT_SECRET=your-secure-jwt-secret
POSTGRES_PASSWORD=your-secure-postgres-password
MONGO_PASSWORD=your-secure-mongo-password
```

### Using External Databases

If you prefer to use managed databases instead of Docker containers:

```bash
# For external MongoDB (e.g., MongoDB Atlas)
MONGODB_URI=mongodb+srv://user:pass@cluster.mongodb.net/ultra?retryWrites=true&w=majority

# For external PostgreSQL (e.g., AWS RDS)
POSTGRES_HOST=your-postgres-server.amazonaws.com
POSTGRES_SSL=require
```

Then remove the database services from `docker-compose.yml`.

## SSL/HTTPS Setup

### Option 1: Reverse Proxy (Recommended)

Use nginx or Caddy as a reverse proxy:

```nginx
server {
    listen 80;
    server_name your-domain.com;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Option 2: Direct HTTPS

Modify the docker-compose.yml to:
1. Mount SSL certificates
2. Change the application to serve HTTPS
3. Update health check URLs

## Monitoring and Maintenance

### Log Management

Logs are automatically rotated:
- Max size: 50MB per file
- Max files: 5 files per service
- Location: Docker's logging driver

### Database Backups

```bash
# Create backup
./deploy.sh backup

# Backups are stored in ./backups/ directory
ls -la backups/
```

### Health Monitoring

```bash
# Check if everything is running
./deploy.sh health

# Monitor resource usage
docker stats
```

### Updates

When new versions are released:

```bash
./deploy.sh update
```

This automatically:
1. Pulls the latest image
2. Recreates containers
3. Maintains data persistence

## Data Persistence

Your data is stored in Docker volumes:
- `ultra_postgres_data` - PostgreSQL data
- `ultra_mongo_data` - MongoDB data
- `ultra_mongo_config` - MongoDB configuration

These persist even when containers are recreated.

## Troubleshooting

### Common Issues

1. **Port conflicts**: Change ports in docker-compose.yml
2. **Permission errors**: Ensure user can run Docker
3. **Memory issues**: Increase server RAM or reduce resource limits
4. **Database connection fails**: Check health with `./deploy.sh health`

### Getting Logs

```bash
# All services
./deploy.sh logs

# Specific service
./deploy.sh logs ultra-api
./deploy.sh logs postgres
./deploy.sh logs mongo

# Follow logs in real-time
docker-compose logs -f
```

### Reset Everything

```bash
# WARNING: This deletes all data!
./deploy.sh clean
```

## Security Checklist

- [ ] Changed default passwords in `.env`
- [ ] Generated secure JWT secret  
- [ ] Set up firewall rules
- [ ] Configured SSL/TLS
- [ ] Set up log monitoring
- [ ] Configured automated backups
- [ ] Restricted Docker daemon access

## Support

If you encounter issues:

1. Check logs: `./deploy.sh logs`
2. Check health: `./deploy.sh health`  
3. Verify environment: `cat .env`
4. Check Docker: `docker-compose ps`
5. Review this README

For application-specific issues, check the main repository documentation.
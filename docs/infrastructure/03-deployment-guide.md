# Production Deployment Guide

This guide covers the complete process for deploying the Alignment application to a production environment on Azure.

## Prerequisites

- Azure CLI installed and authenticated (`az login`)
- Docker installed locally (for testing)
- A GitHub repository with the Alignment code
- `jq` installed for JSON parsing

## Quick Start

1. **Configure Environment**
   ```bash
   cp .env.example .env
   # Edit .env with your specific values
   ```

2. **Provision Infrastructure**
   ```bash
   ./scripts/provision.sh
   ```

3. **Deploy Application**
   ```bash
   ./scripts/deploy.sh
   ```

4. **Access Application**
   - Frontend: `http://<VM_IP>`
   - Backend API: `http://<VM_IP>:8080`
   - Grafana: `http://<VM_IP>:3000`
   - Prometheus: `http://<VM_IP>:9090`

5. **Clean Up (when done)**
   ```bash
   ./scripts/destroy.sh
   ```

## Environment Configuration

The `.env` file contains all necessary configuration variables:

### Required Variables
- `UNIQUE_NAME`: A short, unique identifier for your deployment
- `RESOURCE_GROUP`: Azure resource group name
- `VM_USERNAME`: Username for the VM
- `REPO_URL`: Your GitHub repository URL

### Security Variables
- `REDIS_PASSWORD`: Password for Redis
- `JWT_SECRET`: Secret for JWT token signing
- `ADMIN_API_KEY`: API key for admin endpoints
- `GF_SECURITY_ADMIN_PASSWORD`: Grafana admin password

### Database Variables (PostgreSQL)
- `POSTGRES_DB`: Database name
- `POSTGRES_USER`: Database username
- `POSTGRES_PASSWORD`: Database password

## Architecture

The production deployment includes:

### Application Services
- **Backend**: Go application in optimized Alpine container (~20MB)
- **Frontend**: React application served by Nginx
- **Redis**: In-memory data store with persistence
- **PostgreSQL**: Primary database for user data

### Monitoring Stack
- **Prometheus**: Metrics collection and storage
- **Grafana**: Visualization and dashboards
- **Loki**: Log aggregation
- **Promtail**: Log collection agent

## Security Features

- Non-root user execution in containers
- Security headers in Nginx configuration
- Health checks for all services
- Isolated network for inter-service communication
- Password-protected Redis and PostgreSQL

## Monitoring

Access Grafana at `http://<VM_IP>:3000` with:
- Username: `admin`
- Password: Value from `GF_SECURITY_ADMIN_PASSWORD`

Available data sources:
- **Prometheus**: Application and infrastructure metrics
- **Loki**: Centralized logging from all containers

## Troubleshooting

### Check Service Status
```bash
ssh azureuser@<VM_IP>
cd alignment
docker-compose -f docker-compose.prod.yml ps
```

### View Logs
```bash
# All services
docker-compose -f docker-compose.prod.yml logs

# Specific service
docker-compose -f docker-compose.prod.yml logs backend

# Follow logs
docker-compose -f docker-compose.prod.yml logs -f
```

### Restart Services
```bash
docker-compose -f docker-compose.prod.yml restart
```

## Scaling Considerations

This single-VM deployment is suitable for:
- Development and staging environments
- Small to medium production loads
- MVP and proof-of-concept deployments

For larger deployments, consider:
- Kubernetes orchestration
- Separate database servers
- CDN for static assets
- Load balancing across multiple instances

## Backup Strategy

### Database Backup
```bash
docker-compose -f docker-compose.prod.yml exec postgres pg_dump -U alignment_user alignment > backup.sql
```

### Redis Backup
Redis persistence is enabled by default with RDB snapshots.

### Application Data
All persistent data is stored in Docker volumes:
- `postgres_data`
- `redis_data`
- `grafana_data`
- `prometheus_data`
- `loki_data`
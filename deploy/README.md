# Deployment

## Local Development

### Using Docker Compose

```bash
# From project root
make dev

# Or directly
docker-compose -f deploy/docker-compose.yml up --build
```

This starts:
- **API**: http://localhost:8080
- **PostgreSQL**: localhost:5432

### Verify

```bash
# Health check
curl http://localhost:8080/health

# Expected: {"status":"ok","timestamp":"...","version":"1.0.0"}
```

### Stop

```bash
make stop

# Or
docker-compose -f deploy/docker-compose.yml down

# With data cleanup
docker-compose -f deploy/docker-compose.yml down -v
```

## Environment Variables

### API Service

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| PORT | No | 8080 | HTTP port |
| DEBUG | No | false | Enable debug mode |
| DATABASE_URL | Yes | - | PostgreSQL connection string |
| JWT_SECRET | Yes | - | JWT signing secret |

### Example .env file

```bash
PORT=8080
DEBUG=true
DATABASE_URL=postgres://habitflow:habitflow@localhost:5432/habitflow?sslmode=disable
JWT_SECRET=your-secret-key-min-32-chars
```

## Production Deployment

### Option 1: Railway

1. Connect GitHub repo
2. Set environment variables
3. Deploy from `main` branch

### Option 2: DigitalOcean App Platform

1. Create new app
2. Connect GitHub repo
3. Configure:
   - Build command: `cd backend && go build -o bin/api ./cmd/api`
   - Run command: `./bin/api`
4. Add PostgreSQL database
5. Set environment variables

### Option 3: Kubernetes

See `/deploy/k8s/` for manifests (coming soon).

```bash
# Apply manifests
kubectl apply -f deploy/k8s/

# Check status
kubectl get pods -l app=habitflow
```

## Database Migrations

```bash
# Install migrate tool
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Apply migrations
make migrate-up

# Rollback
make migrate-down
```

## Monitoring

### Health Check

```
GET /health

Response:
{
  "status": "ok",
  "timestamp": "2024-01-15T12:00:00Z",
  "version": "1.0.0"
}
```

### Logs

```bash
# Docker
docker-compose -f deploy/docker-compose.yml logs -f api

# Kubernetes
kubectl logs -f deployment/habitflow-api
```

## Security Checklist

- [ ] Change default passwords
- [ ] Set strong JWT_SECRET (min 32 chars)
- [ ] Enable HTTPS in production
- [ ] Configure CORS properly
- [ ] Enable rate limiting
- [ ] Set up database backups

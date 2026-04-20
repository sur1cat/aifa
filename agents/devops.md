# DevOps Agent

## Role
DevOps Engineer для Atoma — инфраструктура, деплой, мониторинг.

## Responsibilities
- Настройка и поддержка серверов
- CI/CD pipelines
- Docker containers
- SSL/TLS certificates
- Мониторинг и логирование
- Backup и disaster recovery

## Context
```
Infrastructure:
- Server: Hetzner VPS (46.62.141.47)
- OS: Ubuntu/Debian
- Domain: api.azamatbigali.online
- SSL: Let's Encrypt (nginx)
- Container: Docker
- Database: PostgreSQL 16 (Docker)
- Network: Docker bridge network
```

## Prompt Template
```
Ты DevOps Engineer проекта Atoma.

Инфраструктура:
- Hetzner VPS: 46.62.141.47
- Domain: api.azamatbigali.online
- Docker containers: habitflow-api, habitflow-db
- Network: deploy_habitflow-network
- Nginx reverse proxy с SSL

Принципы:
- Infrastructure as Code
- Immutable deployments
- Zero-downtime deploys
- Security first
- Automated backups

При работе с инфраструктурой:
1. Документируй все изменения
2. Тестируй на staging перед production
3. Имей план отката
4. Мониторь после деплоя
```

## Deployment Commands

### Sync and Deploy
```bash
# Sync backend to server
rsync -avz backend/ root@46.62.141.47:/root/habitflow/backend/

# SSH to server
ssh root@46.62.141.47

# Rebuild and restart
cd /root/habitflow/backend
docker build -t habitflow-api .
docker stop habitflow-api && docker rm habitflow-api
docker run -d --name habitflow-api \
  --network deploy_habitflow-network \
  -p 8080:8080 \
  -e DATABASE_URL='postgres://habitflow:PASSWORD@habitflow-db:5432/habitflow?sslmode=disable' \
  -e JWT_SECRET='your-secret' \
  habitflow-api
```

### Database Operations
```bash
# Run migration
docker exec habitflow-db psql -U habitflow -d habitflow -f /path/to/migration.sql

# Backup database
docker exec habitflow-db pg_dump -U habitflow habitflow > backup_$(date +%Y%m%d).sql

# Restore database
cat backup.sql | docker exec -i habitflow-db psql -U habitflow habitflow
```

### Monitoring
```bash
# Check container status
docker ps

# View logs
docker logs -f habitflow-api
docker logs -f habitflow-db

# Check nginx
sudo nginx -t
sudo systemctl status nginx
```

### SSL Certificate
```bash
# Renew Let's Encrypt
sudo certbot renew
sudo systemctl reload nginx
```

## Docker Configuration

### Dockerfile (backend)
```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o api ./cmd/api

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/api /api
EXPOSE 8080
CMD ["/api"]
```

### docker-compose.yml
```yaml
version: '3.8'
services:
  api:
    build: ./backend
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://habitflow:password@db:5432/habitflow?sslmode=disable
      - JWT_SECRET=secret
    depends_on:
      - db
    networks:
      - habitflow-network

  db:
    image: postgres:16-alpine
    environment:
      - POSTGRES_USER=habitflow
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=habitflow
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - habitflow-network

volumes:
  postgres_data:

networks:
  habitflow-network:
```

## Artifacts
- `/deploy/` — Docker, nginx configs
- `/backend/Dockerfile` — API Dockerfile
- `/backend/migrations/` — Database migrations

## Security Checklist
- [ ] SSH key-only authentication
- [ ] Firewall configured (ufw)
- [ ] SSL/TLS enabled
- [ ] Database not exposed publicly
- [ ] Secrets in environment variables
- [ ] Regular security updates

## Collaboration
- **Architect**: согласует инфраструктурные решения
- **Developer**: деплоит код, получает доступы

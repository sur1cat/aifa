# Atoma Skills & Commands

Быстрые команды и скиллы для работы с проектом.

---

## Deploy

### Backend Deploy (Full)
```bash
# 1. Sync code to server
rsync -avz --exclude '.git' --exclude 'tmp' backend/ root@46.62.141.47:/root/habitflow/backend/

# 2. Build and restart container
ssh root@46.62.141.47 "cd /root/habitflow/backend && \
  docker build -t habitflow-api . && \
  docker stop habitflow-api && docker rm habitflow-api && \
  docker run -d --name habitflow-api \
    --network deploy_habitflow-network \
    -p 8080:8080 \
    -e DATABASE_URL='postgres://habitflow:Z9LmuoZgnxY73FFc4nSKJF6ipN5s4O@habitflow-db:5432/habitflow?sslmode=disable' \
    -e JWT_SECRET='your-super-secret-jwt-key-change-in-production' \
    -e DEBUG=false \
    habitflow-api"

# 3. Verify
curl https://api.azamatbigali.online/api/v1/health
```

### Backend Deploy (Quick restart)
```bash
ssh root@46.62.141.47 "docker restart habitflow-api && sleep 2 && docker logs habitflow-api --tail 10"
```

### Run Migration
```bash
# Copy migration file
scp backend/migrations/XXX.up.sql root@46.62.141.47:/tmp/

# Execute
ssh root@46.62.141.47 "docker exec -i habitflow-db psql -U habitflow -d habitflow < /tmp/XXX.up.sql"
```

---

## iOS Build

### Increment Build Number
```bash
cd ios/HabitFlow
agvtool next-version -all
```

### Archive for TestFlight
```bash
cd ios/HabitFlow
xcodebuild -scheme HabitFlow \
  -destination 'generic/platform=iOS' \
  -configuration Release \
  -archivePath build/HabitFlow.xcarchive \
  archive
```

### Check Current Version
```bash
cd ios/HabitFlow
agvtool what-marketing-version  # App version (1.0.0)
agvtool what-version            # Build number (42)
```

---

## Testing

### Backend Tests
```bash
cd backend
go test ./...                    # All tests
go test ./internal/handler/...   # Handler tests only
go test -v ./...                 # Verbose
go test -cover ./...             # With coverage
```

### API Testing
```bash
# Health check
curl -s https://api.azamatbigali.online/api/v1/health | jq .

# Auth (get token via OTP)
curl -X POST https://api.azamatbigali.online/api/v1/auth/otp/send \
  -H "Content-Type: application/json" \
  -d '{"phone": "+77001234567"}'

# Authenticated request
curl https://api.azamatbigali.online/api/v1/habits \
  -H "Authorization: Bearer TOKEN"

# Create resource
curl -X POST https://api.azamatbigali.online/api/v1/habits \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title": "Test", "icon": "🎯", "color": "green"}'
```

---

## Database

### Connect to Production DB
```bash
ssh root@46.62.141.47 "docker exec -it habitflow-db psql -U habitflow -d habitflow"
```

### Quick Queries
```bash
# List tables
ssh root@46.62.141.47 "docker exec habitflow-db psql -U habitflow -d habitflow -c '\dt'"

# Count users
ssh root@46.62.141.47 "docker exec habitflow-db psql -U habitflow -d habitflow -c 'SELECT COUNT(*) FROM users'"

# Recent habits
ssh root@46.62.141.47 "docker exec habitflow-db psql -U habitflow -d habitflow -c 'SELECT id, title FROM habits ORDER BY created_at DESC LIMIT 5'"
```

### Backup
```bash
ssh root@46.62.141.47 "docker exec habitflow-db pg_dump -U habitflow habitflow > /root/backup_$(date +%Y%m%d).sql"
```

---

## Monitoring

### Check Server Status
```bash
# Container status
ssh root@46.62.141.47 "docker ps"

# API logs
ssh root@46.62.141.47 "docker logs habitflow-api --tail 50"

# Follow logs
ssh root@46.62.141.47 "docker logs -f habitflow-api"

# Container stats
ssh root@46.62.141.47 "docker stats --no-stream"
```

### Check Disk Space
```bash
ssh root@46.62.141.47 "df -h"
```

---

## Git Workflow

### Commit with Claude signature
```bash
git add .
git commit -m "$(cat <<'EOF'
Description of changes

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
git push origin main
```

### Quick Status
```bash
git status
git log --oneline -5
git diff --stat
```

---

## Local Development

### Run Backend Locally
```bash
cd backend
export DATABASE_URL='postgres://habitflow:password@localhost:5432/habitflow?sslmode=disable'
export JWT_SECRET='dev-secret'
export DEBUG=true
go run ./cmd/api
```

### Local DB (Docker)
```bash
docker run -d --name habitflow-db-local \
  -e POSTGRES_USER=habitflow \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=habitflow \
  -p 5432:5432 \
  postgres:16-alpine
```

---

## Troubleshooting

### Container won't start
```bash
# Check logs
ssh root@46.62.141.47 "docker logs habitflow-api"

# Check if port is in use
ssh root@46.62.141.47 "lsof -i :8080"

# Force recreate
ssh root@46.62.141.47 "docker rm -f habitflow-api"
```

### Database connection issues
```bash
# Test DB connection
ssh root@46.62.141.47 "docker exec habitflow-db pg_isready -U habitflow"

# Check DB logs
ssh root@46.62.141.47 "docker logs habitflow-db --tail 20"
```

### SSL issues
```bash
# Check nginx
ssh root@46.62.141.47 "nginx -t"
ssh root@46.62.141.47 "systemctl status nginx"

# Renew certificates
ssh root@46.62.141.47 "certbot renew"
```

---

## Quick Reference

| Task | Command |
|------|---------|
| Deploy backend | See "Backend Deploy" |
| Check API health | `curl https://api.azamatbigali.online/api/v1/health` |
| View logs | `ssh root@46.62.141.47 "docker logs habitflow-api --tail 50"` |
| Run tests | `cd backend && go test ./...` |
| iOS build | `agvtool next-version -all` |
| DB shell | `ssh ... "docker exec -it habitflow-db psql -U habitflow -d habitflow"` |

---
*Last updated: January 4, 2026*

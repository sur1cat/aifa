# Aifa Backend

Go API for Aifa application.

## Tech Stack

- Go 1.22
- Gin web framework
- PostgreSQL 16
- JWT authentication

## Project Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go          # Entry point
├── internal/
│   ├── domain/              # Business entities
│   ├── usecase/             # Business logic
│   ├── repository/          # Data access
│   ├── handler/             # HTTP handlers
│   └── middleware/          # HTTP middleware
├── pkg/                     # Shared packages
├── migrations/              # Database migrations
├── go.mod
└── Dockerfile
```

## Development

### Prerequisites

- Go 1.22+
- Docker (for PostgreSQL)

### Run Locally

```bash
# From project root
make dev

# Or manually
cd backend
go run ./cmd/api
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| PORT | 8080 | Server port |
| DEBUG | false | Enable debug mode |
| DATABASE_URL | - | PostgreSQL connection string |
| JWT_SECRET | - | JWT signing secret |

### API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /health | Health check |
| GET | /api/v1/health | API health check |

## Testing

```bash
go test -v ./...
```

## Building

```bash
# Local build
go build -o bin/api ./cmd/api

# Docker build
docker build -t aifa-api .
```

## Migrations

```bash
# Apply migrations
make migrate-up

# Rollback
make migrate-down

# Create new migration
make migrate-create name=add_users
```

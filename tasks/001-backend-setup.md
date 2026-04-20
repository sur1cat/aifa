# Task: [TASK-001] Backend Setup

## Metadata

| Field | Value |
|-------|-------|
| Status | 🔴 Not Started |
| Priority | P0 |
| Assignee | - |
| Estimate | 2 hours |
| Created | 2024-12-26 |

## Goal

Set up a working Go backend skeleton with health endpoint that can be run via Docker.

## Context

This is the foundational task for the backend. We need a working API server before implementing any features.

References:
- [Architecture Overview](/docs/architecture/OVERVIEW.md)
- [CLAUDE.md](/CLAUDE.md) for conventions

## Acceptance Criteria

- [ ] Go project structure created following Clean Architecture
- [ ] Gin router configured
- [ ] Health endpoint returns `{"status": "ok", "timestamp": "..."}`
- [ ] Graceful shutdown implemented
- [ ] Config loaded from environment variables
- [ ] Dockerfile builds successfully
- [ ] docker-compose starts API + PostgreSQL
- [ ] `curl localhost:8080/health` returns 200

## Implementation Steps

### 1. Create Go module

```bash
cd backend
go mod init habitflow
```

### 2. Create main.go

File: `backend/cmd/api/main.go`

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gin-gonic/gin"
    "habitflow/internal/handler"
)

func main() {
    // Config
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    // Router
    r := gin.Default()

    // Health check
    r.GET("/health", handler.Health)

    // Server
    srv := &http.Server{
        Addr:    ":" + port,
        Handler: r,
    }

    // Start server
    go func() {
        log.Printf("Starting server on :%s", port)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Failed to start server: %v", err)
        }
    }()

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    log.Println("Shutting down server...")

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }

    log.Println("Server exited")
}
```

### 3. Create health handler

File: `backend/internal/handler/health.go`

```go
package handler

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
)

func Health(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status":    "ok",
        "timestamp": time.Now().UTC().Format(time.RFC3339),
    })
}
```

### 4. Add dependencies

```bash
cd backend
go get github.com/gin-gonic/gin
go mod tidy
```

### 5. Create Dockerfile

File: `backend/Dockerfile`

```dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api

FROM alpine:3.19

RUN apk add --no-cache ca-certificates

COPY --from=builder /api /api

EXPOSE 8080

CMD ["/api"]
```

### 6. Update docker-compose.yml

File: `deploy/docker-compose.yml`

```yaml
services:
  api:
    build:
      context: ../backend
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - DATABASE_URL=postgres://habitflow:habitflow@postgres:5432/habitflow?sslmode=disable
    depends_on:
      postgres:
        condition: service_healthy

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: habitflow
      POSTGRES_PASSWORD: habitflow
      POSTGRES_DB: habitflow
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U habitflow"]
      interval: 5s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
```

## Testing

```bash
# Start services
make dev

# Wait for startup, then test
curl http://localhost:8080/health

# Expected output:
# {"status":"ok","timestamp":"2024-12-26T12:00:00Z"}

# Stop services
make stop
```

## Definition of Done

- [ ] `make dev` starts both services
- [ ] Health endpoint responds with 200
- [ ] `make stop` cleanly shuts down
- [ ] No errors in logs
- [ ] Code follows conventions in CLAUDE.md

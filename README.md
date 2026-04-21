# Aifa Backend

REST API for **Aifa** — a minimalist iOS app for habit tracking, task management, and budget control.

## Tech Stack

- **Go 1.23** with Gin web framework
- **PostgreSQL 16** with pgx driver and connection pooling
- **JWT** authentication (Google & Apple Sign-In)
- **OpenAI** integration for AI-powered insights and chat
- **APNs** for push notifications

## Project Structure

```
aifa/
├── cmd/api/                  # Application entry point
├── internal/
│   ├── domain/               # Business entities (User, Habit, Task, Goal, Transaction)
│   ├── handler/              # HTTP handlers
│   ├── middleware/            # Auth & rate limiting middleware
│   └── repository/           # PostgreSQL data access layer
├── pkg/
│   ├── ai/                   # OpenAI client
│   ├── auth/                 # JWT, Google & Apple token verification
│   ├── config/               # Environment-based configuration
│   ├── database/             # Connection pool setup
│   └── push/                 # APNs push notifications
├── migrations/               # Sequential SQL migrations (001-017)
├── docs/
│   └── openapi.yaml          # Full OpenAPI 3.1 specification
├── Dockerfile                # Multi-stage production build
├── docker-compose.yml        # Local dev environment (API + PostgreSQL)
└── Makefile                  # Build, test, migrate commands
```

## Getting Started

### Prerequisites

- Go 1.23+
- PostgreSQL 16 (or Docker)
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI (for migrations)

### Setup

```bash
# Clone and configure
cp .env.example .env
# Edit .env with your settings

# Start PostgreSQL (via Docker)
make docker-up

# Run migrations
make migrate-up

# Start the API
make run
```

### Environment Variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | Server port |
| `DEBUG` | `false` | Debug mode (relaxes JWT check, adds CORS origins) |
| `DATABASE_URL` | `postgres://habitflow:...` | PostgreSQL connection string |
| `JWT_SECRET` | — | **Required in production** (min 32 chars) |
| `JWT_ACCESS_TTL_DAYS` | `30` | Access token lifetime |
| `JWT_REFRESH_TTL_DAYS` | `365` | Refresh token lifetime |
| `OPENAI_API_KEY` | — | OpenAI API key (for AI features) |
| `OPENAI_MODEL` | `gpt-4o-mini` | OpenAI model |

## API Overview

Full specification: [`docs/openapi.yaml`](docs/openapi.yaml)

### Authentication

All protected endpoints require `Authorization: Bearer <access_token>`.

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/auth/google` | Google Sign-In |
| POST | `/api/v1/auth/apple` | Apple Sign-In |
| POST | `/api/v1/auth/refresh` | Refresh tokens |
| GET | `/api/v1/auth/me` | Current user profile |
| POST | `/api/v1/auth/logout` | Invalidate tokens |
| DELETE | `/api/v1/auth/account` | Delete account |

### Habits

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/habits` | List habits |
| POST | `/api/v1/habits` | Create habit |
| GET/PUT/DELETE | `/api/v1/habits/:id` | Get/update/delete habit |
| POST | `/api/v1/habits/:id/toggle` | Toggle daily completion |

### Tasks

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/tasks` | List tasks (filter by `?date=`) |
| POST | `/api/v1/tasks` | Create task |
| GET/PUT/DELETE | `/api/v1/tasks/:id` | Get/update/delete task |
| POST | `/api/v1/tasks/:id/toggle` | Toggle completion |

### Goals

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/goals` | List goals |
| POST | `/api/v1/goals` | Create goal |
| GET/PUT/DELETE | `/api/v1/goals/:id` | Get/update/delete goal |

### Transactions & Budget

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/transactions` | List transactions (filter by `?year=&month=`) |
| POST | `/api/v1/transactions` | Create transaction |
| GET | `/api/v1/transactions/summary` | Monthly income/expense summary |
| GET/PUT/DELETE | `/api/v1/transactions/:id` | Get/update/delete transaction |
| GET | `/api/v1/recurring-transactions` | List recurring transactions |
| POST | `/api/v1/recurring-transactions` | Create recurring transaction |
| GET | `/api/v1/recurring-transactions/projection` | Monthly projection |
| POST | `/api/v1/recurring-transactions/process` | Process due payments |
| GET/PUT/DELETE | `/api/v1/recurring-transactions/:id` | Get/update/delete |
| GET/POST/DELETE | `/api/v1/savings-goal` | Savings goal CRUD |

### AI

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/ai/chat` | Chat with AI agents |
| POST | `/api/v1/ai/insights` | Generate insights (habits/tasks/budget/weekly) |
| POST | `/api/v1/ai/expense-analysis` | Detailed expense analysis |
| POST | `/api/v1/ai/goal-to-habits` | Generate habits from a goal |
| POST | `/api/v1/ai/goal-clarify` | Clarifying questions for goals |

### Push Notifications

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/push/register` | Register device token |
| POST | `/api/v1/push/unregister` | Unregister device token |

## Development

```bash
make test              # Run all tests
make test-cover        # Tests with coverage report
make test-integration  # Integration tests
make lint              # Run golangci-lint
```

## Deployment

```bash
# Docker build
make docker-build

# Or full stack via docker-compose
docker compose up -d
```

The Docker image uses a multi-stage build with a non-root user, health checks, and stripped binaries.

## Database Migrations

```bash
make migrate-up                    # Apply all pending migrations
make migrate-down                  # Rollback last migration
make migrate-create name=add_foo   # Create new migration
```

# Atoma - Project Context

## Overview

Atoma is a minimalist iOS app for mindful living: habit tracking, task management, and budget control.

**Tagline**: habits, tasks, money — in one flow

**Production API**: https://api.azamatbigali.online

**Repository**: https://github.com/Azamatfg/atoma

## Tech Stack

| Component | Technology | Version |
|-----------|------------|---------|
| Backend | Go | 1.23+ |
| Framework | Gin | 1.9+ |
| Database | PostgreSQL | 16 |
| iOS | SwiftUI | iOS 17+ |
| Auth | JWT + Google/Apple Sign-In + Phone OTP |
| AI | OpenAI GPT-4 |
| Hosting | Hetzner VPS | Docker |

## Project Structure

```
/habitflow
├── CLAUDE.md              # Project context (this file)
├── TODO.md                # Current task list
├── REFACTORING_PLAN.md    # Refactoring progress tracker
│
├── backend/               # Go API server
│   ├── cmd/api/           # Application entrypoint
│   ├── internal/
│   │   ├── domain/        # Business entities
│   │   ├── handler/       # HTTP handlers
│   │   ├── repository/    # Database operations
│   │   └── middleware/    # Auth, rate limiting, ownership
│   ├── pkg/
│   │   ├── auth/          # JWT, Google, Apple verification
│   │   ├── ai/            # OpenAI client
│   │   ├── config/        # Configuration
│   │   └── database/      # PostgreSQL connection
│   ├── migrations/        # SQL migrations (1-16)
│   └── docs/
│       ├── openapi.yaml   # OpenAPI 3.1 specification
│       └── API_VERSIONING.md
│
├── ios/
│   └── HabitFlow/         # Xcode project (app name: Atoma)
│       ├── HabitFlow/
│       │   ├── Core/
│       │   │   ├── Auth/          # AuthManager
│       │   │   ├── Network/       # APIClient, Services
│       │   │   ├── Storage/       # DataManager, KeychainHelper
│       │   │   ├── Models/        # Habit, Task, Transaction
│       │   │   ├── Utilities/     # DateFormatters, Logger, LRUCache
│       │   │   └── DesignSystem/  # Theme, SyncErrorBanner
│       │   └── Features/
│       │       ├── Habits/        # HabitsView, EditHabitSheet
│       │       ├── Tasks/         # TasksView, EditTaskSheet
│       │       ├── Budget/        # BudgetView, TransactionSheet
│       │       └── Profile/       # ProfileView, AIChatView
│       └── HabitFlowTests/        # Unit tests
│
├── agents/                # AI agent prompts
├── docs/                  # Documentation
├── specs/                 # Feature specifications
└── deploy/                # Docker, nginx configs
```

## API Documentation

Full API specification: `backend/docs/openapi.yaml`

### Base URL
```
Production: https://api.azamatbigali.online/api/v1
Local:      http://localhost:8080/api/v1
```

### Authentication
All protected endpoints require JWT Bearer token:
```
Authorization: Bearer <access_token>
```

### Endpoints Overview

| Group | Endpoints |
|-------|-----------|
| **Auth** | `/auth/google`, `/auth/apple`, `/auth/otp/send`, `/auth/otp/verify`, `/auth/refresh`, `/auth/me`, `/auth/logout` |
| **Habits** | `/habits`, `/habits/:id`, `/habits/:id/toggle` |
| **Tasks** | `/tasks`, `/tasks/:id`, `/tasks/:id/toggle` |
| **Transactions** | `/transactions`, `/transactions/:id`, `/transactions/summary` |
| **Recurring** | `/recurring-transactions`, `/recurring-transactions/:id`, `/recurring-transactions/projection` |
| **Goals** | `/goals`, `/goals/:id` |
| **Savings** | `/savings-goal` |
| **AI** | `/ai/chat`, `/ai/insights`, `/ai/expense-analysis` |
| **Push** | `/push/register`, `/push/unregister` |

### Response Format
```json
// Success
{ "data": { ... } }

// Success with pagination
{ "data": [...], "meta": { "limit": 50, "offset": 0, "total": 100 } }

// Error
{ "error": { "code": "ERROR_CODE", "message": "..." } }
```

### Error Codes
| Code | HTTP | Description |
|------|------|-------------|
| VALIDATION_ERROR | 400 | Invalid request |
| UNAUTHORIZED | 401 | Auth required |
| NOT_FOUND | 404 | Resource not found |
| RATE_LIMITED | 429 | Too many requests |
| INTERNAL_ERROR | 500 | Server error |

## Database Schema

```sql
-- Core tables
users (id, email, phone, name, avatar_url, auth_provider, provider_id)
habits (id, user_id, title, icon, color, period, target_value, unit, goal_id, archived_at)
habit_completions (id, habit_id, completed_date)
habit_progress (id, habit_id, date, value)
tasks (id, user_id, title, is_completed, priority, due_date)
transactions (id, user_id, title, amount, type, category, date)
recurring_transactions (id, user_id, title, amount, type, category, frequency, next_date, is_active)
goals (id, user_id, title, icon, target_value, unit, deadline, archived_at)
savings_goals (id, user_id, monthly_target)

-- Auth & Security
otp_codes (id, phone, code, verified, expires_at)  -- code is bcrypt hashed
invalidated_tokens (id, token_hash, user_id, expires_at)
device_tokens (id, user_id, token, platform)

-- Indexes (migration 016)
idx_habits_user_id, idx_transactions_user_date, idx_tasks_user_due_date, etc.
```

## Deployment

### Server Info
- **Host**: 46.62.141.47 (Hetzner VPS)
- **Domain**: api.azamatbigali.online
- **SSL**: Let's Encrypt via nginx
- **Network**: deploy_habitflow-network (Docker)

### Quick Deploy
```bash
# 1. Sync code
rsync -avz --exclude '.git' backend/ root@46.62.141.47:/root/habitflow/backend/

# 2. Build & restart
ssh root@46.62.141.47 "cd /root/habitflow/backend && \
  docker build -t habitflow-api . && \
  docker stop habitflow-api && docker rm habitflow-api && \
  docker run -d --name habitflow-api \
    --network deploy_habitflow-network \
    -p 8080:8080 \
    -e DATABASE_URL='postgres://habitflow:PASSWORD@habitflow-db:5432/habitflow?sslmode=disable' \
    -e JWT_SECRET='your-secret' \
    -e DEBUG=false \
    habitflow-api"

# 3. Run migration (if needed)
ssh root@46.62.141.47 "docker exec habitflow-db psql -U habitflow -d habitflow -c 'SQL_HERE'"
```

### iOS Build
```bash
cd ios/HabitFlow
agvtool next-version -all  # Increment build number
xcodebuild -scheme HabitFlow -destination 'generic/platform=iOS' archive
```

## Key Files

### Backend
| File | Purpose |
|------|---------|
| `cmd/api/main.go` | Routes, middleware setup |
| `internal/handler/*.go` | HTTP handlers |
| `internal/middleware/auth.go` | JWT auth |
| `internal/middleware/ratelimit.go` | Rate limiting |
| `internal/handler/response.go` | Standardized responses |
| `pkg/auth/jwt.go` | Token generation/validation |
| `pkg/ai/openai.go` | OpenAI client |

### iOS
| File | Purpose |
|------|---------|
| `Core/Auth/AuthManager.swift` | Auth state |
| `Core/Network/APIClient.swift` | HTTP client with retry |
| `Core/Storage/DataManager.swift` | State + sync |
| `Core/Utilities/DateFormatters.swift` | Cached formatters |
| `Core/Utilities/LRUCache.swift` | Generic cache |
| `Core/Utilities/Logger.swift` | AppLogger |

## Conventions

### Go
```go
package handler

// Error handling
if err != nil {
    return fmt.Errorf("failed to create: %w", err)
}

// Context first
func (h *Handler) Create(ctx context.Context, req Request) error

// Standardized responses
respondOK(c, data)
respondCreated(c, data)
respondValidationError(c, "message")
respondNotFound(c, "Resource not found")
```

### Swift
```swift
// MVVM with @EnvironmentObject
@EnvironmentObject var dataManager: DataManager

// Async/await
let habits = try await api.request(endpoint: "habits")

// Optimistic updates
items.append(newItem)  // Update UI first
Task { try await sync() }  // Then sync

// Use cached DateFormatters
DateFormatters.apiDate.string(from: date)
```

## Environment Variables

```bash
# Backend (required)
DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=disable
JWT_SECRET=your-secret-key  # Required, fails on startup if missing

# Backend (optional)
PORT=8080
DEBUG=false
OPENAI_API_KEY=sk-...
```

## Bundle ID

iOS: `com.azamatbigali.habitflow`

## Quick Reference

### Test API
```bash
# Health check
curl https://api.azamatbigali.online/api/v1/health

# Get habits (with auth)
curl https://api.azamatbigali.online/api/v1/habits \
  -H "Authorization: Bearer TOKEN"
```

### Run Backend Locally
```bash
cd backend
go run ./cmd/api
```

### Run Tests
```bash
# Backend
cd backend && go test ./...

# iOS (in Xcode)
Cmd+U
```

## Related Documentation

- [OpenAPI Spec](backend/docs/openapi.yaml)
- [API Versioning](backend/docs/API_VERSIONING.md)
- [Refactoring Plan](REFACTORING_PLAN.md)
- [TODO List](TODO.md)
- [Agents Guide](AGENTS.md)

---
*Last updated: January 4, 2026*

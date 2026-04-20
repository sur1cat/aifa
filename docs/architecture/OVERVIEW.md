# Architecture Overview

## System Diagram

```mermaid
graph TB
    subgraph "Client Layer"
        iOS[iOS App<br/>SwiftUI]
    end

    subgraph "API Layer"
        LB[Load Balancer<br/>nginx/cloudflare]
        API[Go API<br/>Gin Framework]
    end

    subgraph "Data Layer"
        PG[(PostgreSQL<br/>Primary)]
        Redis[(Redis<br/>Cache/Sessions)]
    end

    iOS -->|HTTPS| LB
    LB --> API
    API --> PG
    API --> Redis
```

## Components

### iOS App (SwiftUI)

**Responsibilities**:
- User interface
- Local data storage (SwiftData)
- Offline support
- Sync with backend

**Architecture**: MVVM + Repository Pattern
- **Views**: SwiftUI views
- **ViewModels**: ObservableObject, business logic
- **Repositories**: Data access abstraction
- **Services**: Network, storage, sync

### Go API (Gin Framework)

**Responsibilities**:
- REST API endpoints
- Authentication (JWT)
- Business logic validation
- Data persistence

**Architecture**: Clean Architecture
- **Handler**: HTTP request/response
- **UseCase**: Business logic
- **Repository**: Data access
- **Domain**: Entities, interfaces

### PostgreSQL

**Responsibilities**:
- Primary data storage
- ACID transactions
- Data integrity

### Redis (Future)

**Responsibilities**:
- Session storage
- Rate limiting
- Caching hot data

## Data Flow

### Create Habit Flow

```mermaid
sequenceDiagram
    participant User
    participant iOS
    participant API
    participant DB

    User->>iOS: Tap "Add Habit"
    iOS->>iOS: Show create form
    User->>iOS: Fill form, tap Save
    iOS->>iOS: Validate locally
    iOS->>iOS: Save to SwiftData (optimistic)
    iOS->>API: POST /api/v1/habits
    API->>API: Validate request
    API->>API: Check user limits
    API->>DB: INSERT habit
    DB-->>API: Return habit
    API-->>iOS: 201 Created + habit
    iOS->>iOS: Update local with server ID
    iOS-->>User: Show success
```

### Complete Habit Flow

```mermaid
sequenceDiagram
    participant User
    participant iOS
    participant API
    participant DB

    User->>iOS: Tap habit to complete
    iOS->>iOS: Update UI immediately
    iOS->>iOS: Save completion locally
    iOS->>iOS: Trigger haptic + animation
    iOS->>API: POST /api/v1/habits/{id}/complete
    API->>DB: INSERT habit_completion
    API->>API: Calculate streak
    API-->>iOS: 200 OK + updated habit
    iOS->>iOS: Sync streak from server
```

## Authentication Flow

```mermaid
sequenceDiagram
    participant iOS
    participant API
    participant DB

    iOS->>API: POST /auth/login {email, password}
    API->>DB: Find user by email
    DB-->>API: User record
    API->>API: Verify password (bcrypt)
    API->>API: Generate JWT tokens
    API-->>iOS: {access_token, refresh_token}
    iOS->>iOS: Store in Keychain

    Note over iOS,API: Later requests...

    iOS->>API: GET /habits (Authorization: Bearer token)
    API->>API: Validate JWT
    API->>DB: Fetch habits
    API-->>iOS: habits[]
```

## Offline-First Strategy

### Principles

1. **Local first**: All data stored locally in SwiftData
2. **Optimistic updates**: UI updates before server confirms
3. **Background sync**: Sync happens automatically when online
4. **Conflict resolution**: Last-write-wins with server timestamp

### Sync Flow

```mermaid
graph LR
    A[User Action] --> B[Update Local DB]
    B --> C[Update UI]
    C --> D{Online?}
    D -->|Yes| E[Sync to Server]
    D -->|No| F[Queue for Later]
    F --> G[Network Available]
    G --> E
    E --> H{Success?}
    H -->|Yes| I[Mark Synced]
    H -->|No| J[Retry with Backoff]
```

### Conflict Resolution

| Scenario | Resolution |
|----------|------------|
| Same field modified | Server timestamp wins |
| Habit deleted on server | Remove from local |
| Habit created offline | Assign server ID on sync |
| Completion conflict | Both kept (union) |

## Security

### Authentication
- JWT tokens with RS256 signing
- Access token: 15 minutes
- Refresh token: 7 days
- Tokens stored in iOS Keychain

### Data Protection
- HTTPS only (TLS 1.3)
- Passwords hashed with bcrypt (cost 12)
- SQL injection prevention (parameterized queries)
- Input validation on all endpoints

### Rate Limiting
- 100 requests/minute per user
- 10 login attempts per hour
- 429 response when exceeded

## Scalability

### Current (MVP)
- Single API instance
- Single PostgreSQL instance
- ~1000 users target

### Future
- Horizontal API scaling (stateless)
- PostgreSQL read replicas
- Redis for caching
- CDN for static assets

## Monitoring

### Metrics (Future)
- Request latency (p50, p95, p99)
- Error rate
- Database connection pool
- Active users

### Logging
- Structured JSON logs
- Request ID for tracing
- No sensitive data in logs

## Deployment

See [deploy/README.md](/deploy/README.md) for deployment instructions.

```mermaid
graph LR
    subgraph "Development"
        DEV[Local Docker]
    end

    subgraph "CI/CD"
        GH[GitHub Actions]
    end

    subgraph "Production"
        DO[DigitalOcean<br/>or Railway]
    end

    DEV --> GH
    GH -->|main branch| DO
```

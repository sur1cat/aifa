# HabitFlow

Minimalist iOS app for habit tracking, task management, and budget control.

## Features

- **Habits**: Track daily, weekly, and monthly habits with streaks
- **Tasks**: Simple todo list focused on "today" (coming soon)
- **Budget**: Income/expense tracking with savings goals (coming soon)

## Tech Stack

- **Backend**: Go 1.22, Gin, PostgreSQL 16
- **iOS**: SwiftUI, iOS 17+, SwiftData
- **Infrastructure**: Docker, GitHub Actions

## Quick Start

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- Xcode 15+ (for iOS)

### Backend

```bash
# Start services
make dev

# Run tests
make test

# API is available at http://localhost:8080
curl http://localhost:8080/health
```

### iOS

```bash
cd ios/HabitFlow
open HabitFlow.xcodeproj
# Build and run in Xcode
```

## Project Structure

```
/habitflow
├── docs/           # Documentation
├── specs/          # Feature specifications
├── tasks/          # Implementation tasks
├── backend/        # Go API
├── ios/            # SwiftUI app
└── deploy/         # Docker, K8s
```

## Documentation

- [CLAUDE.md](./CLAUDE.md) - Project context and conventions
- [PRD](./docs/product/PRD.md) - Product requirements
- [Architecture](./docs/architecture/OVERVIEW.md) - System design
- [API Spec](./docs/api/openapi.yaml) - OpenAPI specification

## Development

```bash
# Apply migrations
make migrate-up

# Run linters
make lint

# Build for production
make build
```

## Contributing

1. Read [CLAUDE.md](./CLAUDE.md) for conventions
2. Pick a task from `/tasks/`
3. Create feature branch: `feature/your-feature`
4. Follow code review checklist
5. Submit PR

## License

MIT

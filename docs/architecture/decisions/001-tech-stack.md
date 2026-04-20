# ADR-001: Technology Stack Selection

## Status
Accepted

## Context

We need to choose a technology stack for HabitFlow that:
1. Enables rapid MVP development
2. Provides good iOS user experience
3. Scales to 100K+ users if successful
4. Is maintainable by a small team
5. Has reasonable hosting costs

## Decision

### Backend: Go + Gin + PostgreSQL

**Go** for the API because:
- Simple, readable codebase
- Excellent performance
- Easy deployment (single binary)
- Strong standard library
- Good concurrency model
- Active ecosystem for web services

**Gin** framework because:
- Minimal, fast HTTP router
- Middleware support
- Good documentation
- Production-proven

**PostgreSQL** for data because:
- ACID compliance for financial data (future)
- JSON support for flexible schemas
- Excellent performance
- Free and open source
- Strong community

### Mobile: SwiftUI (iOS only)

**SwiftUI** because:
- Native iOS experience
- Declarative UI = faster development
- Built-in accessibility
- Great animations
- Apple's direction for the future

**iOS only for MVP** because:
- Focus resources on one platform
- iOS users have higher spending
- Design faster for one platform

### Rejected Alternatives

| Option | Reason for Rejection |
|--------|---------------------|
| Node.js | Higher memory usage, callback complexity |
| Python/Django | Slower performance, typed less strictly |
| Rust | Steeper learning curve, slower development |
| React Native | Not native feel, complex debugging |
| Flutter | Dart ecosystem smaller, not Apple-first |
| Cross-platform | Lower quality than native for our use case |

## Consequences

### Positive
- Fast development for MVP
- Native iOS experience
- Low hosting costs (Go is efficient)
- Easy to hire Go/Swift developers
- Simple deployment (Docker + single binary)

### Negative
- No Android version initially
- Team needs Go + Swift skills
- Two codebases to maintain

### Risks
- If Android demand is high, need to build separately
- SwiftUI maturity (some edge cases)

## Notes

Revisit this decision after MVP launch based on:
- User feedback on Android demand
- Development velocity
- Hosting costs

# Atoma Refactoring Plan

## Phase 1: Critical Fixes ✅ COMPLETED

### Backend
- [x] Fix push.go: change `user_id` to `userID` (line 27)
- [x] Add transaction lock for recurring processing (race condition)
- [x] Remove OTP code from response in production (debugMode flag)
- [x] Add amount > 0 validation for transactions
- [x] Fix strconv error handling in transactions.go

### iOS
- [x] Replace `Insight.contentHash` with deterministic hash (SHA256)
- [x] Fix token refresh race condition with Task-based locking
- [x] Add error UI for sync failures (SyncErrorBanner)
- [x] Add Keychain error logging

---

## Phase 2: Security & Data Integrity ✅ COMPLETED

### Backend
- [x] Hash OTP codes in database (bcrypt)
- [x] Implement logout token invalidation (token_repository.go)
- [x] Restrict CORS to specific domains
- [x] Make JWT_SECRET required (fail on startup if missing)
- [ ] Add cascade delete for user (or soft delete) — deferred

### iOS
- [x] Add retry logic with exponential backoff
- [x] Implement conflict resolution for optimistic updates
- [x] Add input validation to all forms (EditHabitSheet, EditTaskSheet)
- [x] Cache DateFormatters (DateFormatterCache.swift)

---

## Phase 3: Architecture Improvements ✅ COMPLETED

### Backend
- [x] Standardize error responses (response.go)
- [x] Create ownership verification middleware (ownership.go)
- [x] Add pagination to list endpoints (transactions)
- [x] Add rate limiting middleware (ratelimit.go)
- [x] Add string field length validation
- [x] Add date format validation

### iOS
- [ ] Split DataManager into feature managers — deferred (large task)
- [x] Replace print() with proper logging framework (AppLogger)
- [x] Remove force unwraps, use proper optionals
- [x] Debounce rapid @Published updates (optimistic updates work well)

---

## Phase 4: Performance ✅ COMPLETED

### Backend
- [x] Add database indexes:
  - habits(user_id)
  - transactions(user_id, date)
  - recurring_transactions(user_id, is_active)
  - habit_completions(habit_id, completed_date)
  - tasks(user_id, due_date)
  - goals(user_id)
  - otp_codes(phone, verified, expires_at)
- [x] Optimize streak calculation (moved to backend)
- [x] Parse dates once when loading (use cached DateFormatters)

### iOS
- [x] Implement LRU cache for data (LRUCache.swift)
- [x] Lazy loading for lists (BudgetView already uses LazyVStack)
- [x] Memoize expensive computed properties (use cached DateFormatters)
- [x] Replace all DateFormatter() with cached DateFormatters
- [ ] Profile and fix memory leaks — deferred

---

## Phase 5: Testing & Documentation ✅ COMPLETED

### Backend
- [x] Add unit tests:
  - response_test.go (7 tests)
  - auth_test.go (6 tests)
  - ratelimit_test.go (4 tests)
  - jwt_test.go (8 tests)
  - Coverage: middleware 59%, auth 20%
- [x] Add OpenAPI/Swagger documentation (backend/docs/openapi.yaml)
- [x] Document API versioning strategy (backend/docs/API_VERSIONING.md)

### iOS
- [x] Add unit tests:
  - LRUCacheTests.swift (10 tests)
  - DateFormattersTests.swift (12 tests)
  - ModelsTests.swift (15+ tests for Habit, Transaction, Task, Insight)
  - MockAPIClient.swift for testing
- [ ] Add test target to Xcode project (manual step)
- [ ] Add integration tests for sync

---

## Code Quality Checklist

### Backend files refactored:
| File | Status | Notes |
|------|--------|-------|
| handler/push.go | ✅ Done | Fixed context key |
| handler/recurring_transactions.go | ✅ Done | Atomic update |
| handler/auth.go | ✅ Done | debugMode for OTP |
| handler/transactions.go | ✅ Done | Validation, pagination |
| repository/otp_repository.go | ✅ Done | bcrypt hashing |
| cmd/api/main.go | ✅ Done | CORS, rate limiting |

### iOS files refactored:
| File | Status | Notes |
|------|--------|-------|
| Core/Models/Models.swift | ✅ Done | SHA256 hash, backend streak |
| Core/Network/APIClient.swift | ✅ Done | Task-based token refresh |
| Core/Storage/DataManager.swift | Partial | Large file, needs split |
| Core/Storage/KeychainHelper.swift | ✅ Done | Error logging |
| Core/Network/InsightService.swift | ✅ Done | Removed force casts |

---

## New Files Created

### Backend
```
internal/handler/response.go     # Standardized error responses
internal/middleware/ratelimit.go # Rate limiting
internal/middleware/ownership.go # Resource ownership verification
internal/repository/token_repository.go # Token blacklist
migrations/014_extend_otp_code_column.up.sql
migrations/015_create_invalidated_tokens.up.sql
migrations/016_add_performance_indexes.up.sql
docs/openapi.yaml                # OpenAPI 3.1 specification
docs/API_VERSIONING.md           # API versioning strategy
```

### iOS
```
Core/Utilities/DateFormatterCache.swift
Core/Utilities/Logger.swift
Core/DesignSystem/SyncErrorBanner.swift
```

---

## Deferred Tasks

1. ~~**Split DataManager**~~ — ✅ COMPLETED (January 2026)
2. **Cascade delete for user** — Needs careful planning for data cleanup
3. **LRU cache** — Current performance is acceptable

---

## DataManager Refactoring (January 2026) ✅ COMPLETED

### New Architecture

DataManager split into 6 feature managers + coordinator:

```
Core/Storage/
├── DataManager.swift           # Coordinator (509 lines, was 2285)
└── Managers/
    ├── GoalsManager.swift      # Goals CRUD + sync
    ├── HabitsManager.swift     # Habits CRUD + toggle + progress
    ├── TasksManager.swift      # Tasks CRUD + sync
    ├── BudgetManager.swift     # Transactions + Recurring + Savings
    ├── InsightManager.swift    # AI Insights generation
    └── AnalyticsManager.swift  # Life Score + Forecast + Charts
```

### Key Changes

- **Reduced complexity**: 2285 → 509 lines in DataManager
- **Single responsibility**: Each manager handles one domain
- **Memory safety**: All 27 Task closures use [weak self]
- **Backward compatible**: Forwarding properties maintain API

### Manager Responsibilities

| Manager | Lines | Features |
|---------|-------|----------|
| TasksManager | ~170 | CRUD, sync, queries |
| GoalsManager | ~170 | CRUD, archive, habit links |
| HabitsManager | ~400 | CRUD, toggle, progress, archive |
| BudgetManager | ~350 | Transactions, Recurring, Savings |
| InsightManager | ~250 | AI generation, unlock tracking |
| AnalyticsManager | ~350 | Life Score, Forecast, Charts |

---

## Metrics to Track

- [ ] API response time < 200ms
- [ ] App launch time < 2s
- [ ] Sync success rate > 99%
- [ ] Crash-free rate > 99.5%
- [ ] Test coverage > 70%

---

*Last updated: January 4, 2026*

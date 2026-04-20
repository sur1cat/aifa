# Atoma - Task List

## Current Status: TestFlight Beta (Build 6)

---

## Completed

### Core Features
- [x] Habits tracking with server sync
- [x] Tasks management with server sync
- [x] Budget/Transactions with server sync
- [x] Recurring transactions (subscriptions)
- [x] Savings goals with progress tracking
- [x] Long-term goals
- [x] Google Sign-In
- [x] Apple Sign-In
- [x] Phone OTP authentication
- [x] Habit reminders with notifications
- [x] Edit/Delete for all entities
- [x] Swipe-to-delete gestures

### AI Features
- [x] OpenAI integration (backend)
- [x] AI Chat with 4 agents (Habit Coach, Task Assistant, Finance Advisor, Life Coach)
- [x] AI Insights generation (/ai/insights)
- [x] AI Expense Analysis (/ai/expense-analysis)
- [x] Voice input for budget (Speech Recognition)
- [x] Receipt scanner (Vision OCR)

### AI Insights System
- [x] Phase 1: Unlock System (14d habits, 7d tasks, 30d budget)
- [x] Phase 2: Local Insights (patterns, achievements, warnings, suggestions)
- [x] InsightCard UI component
- [x] Dismiss/action handling

### Infrastructure
- [x] Git + GitHub repository
- [x] Backend deployed on Hetzner VPS
- [x] PostgreSQL database with indexes
- [x] SSL/HTTPS via nginx
- [x] TestFlight distribution
- [x] Rate limiting middleware
- [x] JWT token invalidation (logout)
- [x] bcrypt hashed OTP codes

### Documentation
- [x] OpenAPI 3.1 specification
- [x] API versioning strategy
- [x] CLAUDE.md project context
- [x] AGENTS.md AI agents guide
- [x] SKILLS.md quick commands

### Refactoring (All Phases Complete)
- [x] Phase 1: Critical Fixes
- [x] Phase 2: Security & Data Integrity
- [x] Phase 3: Architecture Improvements
- [x] Phase 4: Performance (indexes, caching)
- [x] Phase 5: Testing & Documentation

### Testing
- [x] Backend unit tests (25 tests)
- [x] iOS unit tests

### Localization
- [x] English
- [x] Russian
- [x] Spanish
- [x] Portuguese

---

## Backlog

### Technical Improvements
- [x] Split DataManager into feature managers (6 managers created)
- [x] Add [weak self] to all Task closures (27 instances)
- [ ] Add test target to Xcode project (manual setup in Xcode)
- [ ] Integration tests for sync (iOS ↔ API ↔ DB full cycle)
- [ ] Profile with Instruments (optional verification)

### Future Features
- [ ] Widgets (iOS 18)
- [ ] Apple Watch app
- [ ] Data export (CSV/PDF)
- [ ] Custom themes
- [ ] Habit streaks visualization
- [ ] Budget categories management
- [ ] Charts & analytics improvements

### Later (Low Priority)
- [ ] Monetization (StoreKit 2, paywall, premium features)
- [ ] App Store Release (icons, screenshots, description)

---

## Bug Fixes
- [ ] (Add bugs here as discovered)

---

## Quick Commands

```bash
# Deploy backend
rsync -avz backend/ root@46.62.141.47:/root/habitflow/backend/
ssh root@46.62.141.47 "cd /root/habitflow/backend && docker build -t habitflow-api . && docker stop habitflow-api && docker rm habitflow-api && docker run -d --name habitflow-api --network deploy_habitflow-network -p 8080:8080 -e DATABASE_URL='...' -e JWT_SECRET='...' habitflow-api"

# Check API
curl https://api.azamatbigali.online/api/v1/health

# iOS build
cd ios/HabitFlow && agvtool next-version -all

# Run tests
cd backend && go test ./...
```

---

## Documentation Links

- [CLAUDE.md](CLAUDE.md) - Project context
- [AGENTS.md](AGENTS.md) - AI agents guide
- [SKILLS.md](SKILLS.md) - Quick commands
- [OpenAPI Spec](backend/docs/openapi.yaml) - API documentation

---

*Last updated: January 4, 2026*

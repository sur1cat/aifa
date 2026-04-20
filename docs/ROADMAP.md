# Atoma Roadmap

**Last updated:** 2026-01-04

---

## Completed ✅

### Core Features
- [x] Habits с server sync
- [x] Tasks с server sync
- [x] Budget (transactions) с server sync
- [x] Recurring Transactions
- [x] Google/Apple Sign-In
- [x] JWT Authentication (backend)

### AI Features
- [x] OpenAI Integration (backend)
- [x] 4 AI Agents (Habit Coach, Task Assistant, Finance Advisor, Life Coach)
- [x] Voice Input для транзакций (Speech Recognition)
- [x] Receipt Scanner (Vision OCR)

### Premium & Monetization
- [x] Stripe Integration (PaymentSheet)
- [x] Premium Gates для free users
- [x] Free limits: 5 habits, 5 tasks
- [x] Paywall UI

### Analytics & Insights
- [x] Life Score (единый показатель 0-100)
- [x] Life Score History Chart
- [x] Spending by Category Chart
- [x] Habit Streaks View
- [x] Income vs Expenses Chart
- [x] AI Insights система
- [x] Weekly Review
- [x] Budget Forecasting (next month projection)
- [x] AI Expense Analysis (spending patterns, questionable transactions)
- [x] Savings Goals with progress tracking
- [x] Collapsible sections in Budget

### iOS Features
- [x] iOS Widgets (Habits, Tasks, Budget)
- [x] Apple Watch App
- [x] Push Notifications (habit/task reminders, digest)
- [x] Siri Shortcuts (App Intents)
- [x] Alternative App Icons (5 variants)
- [x] Onboarding Flow
- [x] Data Export (CSV)
- [x] Localization (EN, RU, ES, PT)

### Infrastructure
- [x] Git + GitHub repo
- [x] Backend deployed (Hetzner VPS)
- [x] PostgreSQL database
- [x] Docker deployment
- [x] SSL (Let's Encrypt)
- [x] App Store release
- [x] TestFlight builds

---

## In Progress 🔧

### Google OAuth
- [ ] Change app name in Google Cloud Console ("HabitFlow" → "Atoma")

---

## Remaining Features 📋

### Low Priority
- [ ] Server-side Push Notifications (APNs from backend)
- [ ] Offline Mode (iCloud sync)
- [ ] Share Life Score card (image export)

### Monetization (Next Priority)
- [ ] Subscription Plans (Monthly/Yearly)
- [ ] In-App Purchases (StoreKit 2)
- [ ] Paywall optimization (A/B testing)
- [ ] Trial period (7 days free)
- [ ] Family Sharing support
- [ ] Promo codes
- [ ] Revenue analytics (RevenueCat?)

### Future Ideas
- [ ] Calendar integration
- [ ] Habit templates library
- [ ] Social features (challenges with friends)
- [ ] Apple Health integration
- [ ] Expense predictions (ML)

---

## Tech Stack

| Component | Technology |
|-----------|------------|
| iOS | SwiftUI, iOS 17+ |
| Backend | Go 1.23+, Gin |
| Database | PostgreSQL 16 |
| Auth | JWT, Google/Apple Sign-In |
| Payments | Stripe |
| AI | OpenAI GPT-4 |
| Hosting | Hetzner VPS, Docker |

---

## Release History

| Version | Build | Date | Notes |
|---------|-------|------|-------|
| 1.2 | 15 | 2026-01-04 | Budget Forecasting, AI Expense Analysis, Collapsible Sections, Goals |
| 1.2 | 6 | 2025-12-29 | Siri Shortcuts, Analytics Charts, Alt Icons |
| 1.2 | 5 | 2025-12-29 | Premium Gates |
| 1.2 | 4 | 2025-12-29 | Stripe PaymentSheet fix |
| 1.0 | 1 | - | Initial release |

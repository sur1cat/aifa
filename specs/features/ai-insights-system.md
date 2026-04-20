# AI Insights System

## Overview

Proactive AI insights that unlock after collecting sufficient user data. Inspired by Whoop's approach — no metrics until meaningful patterns emerge.

---

## Problem Statement

**Current state:** AI is a passive chat (sparkles button). 95% of users never tap it.

**Desired state:** AI proactively surfaces actionable insights when it has enough data to be meaningful.

---

## User Stories

### US-1: Data Collection Progress
```
As a new user
I want to see how much data I need to collect
So that I understand why AI insights aren't available yet
```

### US-2: Unlock Celebration
```
As a user who collected enough data
I want to be celebrated when AI unlocks
So that I feel rewarded for my consistency
```

### US-3: Proactive Insights
```
As a user with sufficient data
I want AI to show me patterns I didn't notice
So that I can make better decisions
```

---

## Minimum Data Requirements

| Section | Minimum Threshold | Rationale |
|---------|------------------|-----------|
| Habits | 14 days of tracking | Need 2 weeks for weekly pattern detection |
| Tasks | 7 days of tracking | Enough for completion rate patterns |
| Budget | 30 days OR 20 transactions | Full month for spending patterns |

---

## Insight Types by Section

### Habits Insights
| Insight | Required Data | Example |
|---------|--------------|---------|
| Best time of day | 14 days + reminder times | "Meditation works better before 10am (85% vs 40%)" |
| Weekly patterns | 14 days | "You skip Gym on Fridays. Move to Saturday?" |
| Streak analysis | 14 days | "Your longest streaks start on Mondays" |
| Habit correlation | 21 days + 2 habits | "When you meditate, you complete 80% more habits" |

### Tasks Insights
| Insight | Required Data | Example |
|---------|--------------|---------|
| Completion rate | 7 days | "You complete 73% of tasks. Top performers: High priority" |
| Best day | 14 days | "Monday is your most productive day (8/10 tasks)" |
| Overcommitment | 7 days | "You add 5+ tasks but complete 3. Try limiting to 3?" |

### Budget Insights
| Insight | Required Data | Example |
|---------|--------------|---------|
| Top category | 20 transactions | "Coffee is 15% of expenses (₽12,400/month)" |
| Spending trend | 30 days | "Spending up 20% vs last month" |
| Recurring detection | 30 days | "Netflix ₽799 detected monthly. Add to subscriptions?" |
| Income vs Expense | 30 days | "You save 25% of income. Great!" |

---

## UI Components

### 1. Locked State Card
```
┌─────────────────────────────────────────┐
│ 🔒 AI Insights                          │
│                                         │
│ ████████░░░░░░░░░░░░ 8/14 days          │
│                                         │
│ Keep tracking habits for 6 more days    │
│ to unlock personalized insights         │
└─────────────────────────────────────────┘
```

### 2. Unlock Celebration
```
┌─────────────────────────────────────────┐
│ 🎉 AI Insights Unlocked!                │
│                                         │
│ 14 days of data collected.              │
│ I can now see your patterns.            │
│                                         │
│ [See First Insight]                     │
└─────────────────────────────────────────┘
```

### 3. Insight Card
```
┌─────────────────────────────────────────┐
│ 💡 Insight · Habits                     │
│                                         │
│ "You complete Meditation 85% of the     │
│  time before 10am, but only 40% after." │
│                                         │
│ [Set morning reminder]  [Dismiss]       │
└─────────────────────────────────────────┘
```

### 4. Weekly Review Card
```
┌─────────────────────────────────────────┐
│ 📊 Your Week · Dec 23-29                │
│                                         │
│ Habits:  ████████░░ 78% (+5%)           │
│ Tasks:   12/15 completed                │
│ Budget:  ₽28,500 of ₽30,000             │
│                                         │
│ 🏆 Win: Meditation — 7 day streak!      │
│ ⚠️ Watch: Gym skipped 3 times           │
│                                         │
│ [View Details]                          │
└─────────────────────────────────────────┘
```

---

## Data Model

### InsightStatus
```swift
struct InsightStatus: Codable {
    var habitsUnlocked: Bool
    var habitsProgress: Int  // days tracked
    var tasksUnlocked: Bool
    var tasksProgress: Int
    var budgetUnlocked: Bool
    var budgetProgress: Int  // days or transaction count
    var lastInsightDate: Date?
}
```

### Insight
```swift
struct Insight: Identifiable, Codable {
    let id: UUID
    let section: AppSection  // habits, tasks, budget
    let type: InsightType
    let title: String
    let message: String
    let action: InsightAction?
    let createdAt: Date
    var isDismissed: Bool
}

enum InsightType: String, Codable {
    case pattern      // "You skip gym on Fridays"
    case achievement  // "7 day streak!"
    case warning      // "Spending up 20%"
    case suggestion   // "Try morning meditation"
    case weeklyReview // Weekly summary
}

struct InsightAction: Codable {
    let label: String       // "Set reminder"
    let actionType: String  // "setReminder", "adjustHabit", etc.
    let payload: [String: String]
}
```

---

## Implementation Phases

### Phase 1: Foundation (MVP)
**Goal:** Show unlock progress, no actual AI yet

- [ ] Add `InsightStatus` to DataManager
- [ ] Track days of activity per section
- [ ] Create `InsightProgressCard` UI component
- [ ] Show locked state on each section's main view
- [ ] Celebrate unlock moment

### Phase 2: Basic Insights
**Goal:** Simple pattern detection (no OpenAI needed)

- [ ] Implement local insight generators:
  - Habit: best time, streak analysis
  - Tasks: completion rate, best day
  - Budget: top categories, monthly comparison
- [ ] Create `InsightCard` UI component
- [ ] Store insights locally
- [ ] Add dismiss/action handling

### Phase 3: AI-Powered Insights
**Goal:** Use OpenAI for personalized insights

- [ ] Create insight generation prompts
- [ ] Backend endpoint: `POST /api/v1/insights/generate`
- [ ] Weekly scheduled insight generation
- [ ] Push notification for new insights

### Phase 4: Weekly Review
**Goal:** Automated weekly summary

- [ ] Generate weekly review every Sunday
- [ ] Combine all sections into one view
- [ ] Trend comparison (this week vs last)
- [ ] Achievements and warnings

---

## Success Metrics

| Metric | Target | How to Measure |
|--------|--------|----------------|
| Unlock rate | 40% reach 14 days | Users with habitsUnlocked = true |
| Insight engagement | 30% tap action | Action taps / Insights shown |
| Retention D14 | +20% vs current | Compare cohorts |
| Weekly review opens | 50% of unlocked users | Opens / Eligible users |

---

## Open Questions

1. **Notification strategy:** Push when new insight? Weekly digest only?
2. **Insight frequency:** Max 1 per day? 3 per week?
3. **Cross-section insights:** "When you meditate, you spend less on impulse buys" — Phase 4?
4. **Premium gating:** All insights free, or some behind paywall?

---

## References

- Whoop: Recovery Score unlock after 4 days
- Oura: Readiness Score requires baseline
- Apple Health: Trends appear after sufficient data

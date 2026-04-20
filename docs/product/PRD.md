# Product Requirements Document: HabitFlow

## Problem Statement

Modern productivity apps are overwhelming. Users download habit trackers, todo apps, and budget tools, but most abandon them within a week because:

1. **Too complex** - Dozens of features create decision fatigue
2. **Too separate** - Habits, tasks, and finances live in different apps
3. **No context** - Apps don't understand daily/weekly/monthly rhythms
4. **Poor UX** - Cluttered interfaces, slow interactions

## Solution

HabitFlow is a **minimalist all-in-one app** for mindful living:
- One app for habits, tasks, and budget
- Unified view by time period (day/week/month)
- Beautiful, distraction-free interface
- 3-second interactions

## Target Audience

**Primary**: Mindful professionals, 25-40 years old
- iOS users (iPhone primary, iPad secondary)
- Value simplicity and design
- Want to improve life without complexity
- Willing to pay for premium ($4.99/month)

See [personas.md](./personas.md) for detailed user personas.

## MVP Scope (Phase 1)

### In Scope: Habits Module

| Feature | Priority | Description |
|---------|----------|-------------|
| User Auth | P0 | Email/password registration and login |
| Create Habit | P0 | Title, icon, color, period (daily/weekly/monthly) |
| Habit List | P0 | View habits for today/this week/this month |
| Complete Habit | P0 | One-tap completion with animation |
| Streak Tracking | P1 | Current streak, longest streak |
| Basic Stats | P1 | Completion rate, calendar heatmap |
| Edit/Delete Habit | P1 | Modify or remove habits |

### Out of Scope (Future Phases)

**Phase 2**: Tasks
- Simple todo list
- Due dates
- Today focus view

**Phase 3**: Budget
- Income/expense tracking
- Categories
- Savings goals

**Phase 4**: Premium Features
- Sync across devices
- Advanced statistics
- Widgets
- Push notifications

## Features Breakdown

### 1. Authentication

**Registration**:
- Email + password
- Minimum 8 characters password
- Email verification (deferred to Phase 2)

**Login**:
- Email + password
- Remember me option
- Forgot password (deferred to Phase 2)

### 2. Habits

**Create Habit**:
- Title (required, max 50 chars)
- Icon (from predefined set, ~50 icons)
- Color (from predefined palette, 12 colors)
- Period: daily (default), weekly, monthly
- Target per period (default: 1)

**Habit List**:
- Segmented control: Today / Week / Month
- Shows habits for selected period
- Completion status visible
- Swipe to complete

**Complete Habit**:
- Tap or swipe to complete
- Haptic feedback
- Celebratory animation
- Undo within 3 seconds

**Streaks**:
- Current streak (consecutive periods with completion)
- Longest streak (all time)
- Streak shown on habit card

### 3. Statistics

**Completion Rate**:
- Overall: % of habits completed
- Per habit: % completion for each habit

**Calendar Heatmap**:
- Month view
- Color intensity = completion count
- Tap day to see details

## Success Metrics

### North Star Metric
**Weekly Active Completions**: Number of habit completions per week per user

### Leading Indicators
| Metric | Target | Description |
|--------|--------|-------------|
| DAU/MAU | > 40% | Daily engagement ratio |
| Day 1 Retention | > 60% | Users returning next day |
| Day 7 Retention | > 30% | Users returning after a week |
| Habits Created | > 3 | Average habits per user in first week |
| Completion Rate | > 50% | % of habits completed on time |

### Lagging Indicators
| Metric | Target | Description |
|--------|--------|-------------|
| Day 30 Retention | > 20% | Users still active after a month |
| Premium Conversion | > 5% | Free to paid conversion |
| App Store Rating | > 4.5 | Average rating |
| NPS | > 50 | Net Promoter Score |

## Technical Requirements

### Performance
- App launch: < 1 second
- Screen transition: < 300ms
- API response: < 200ms (p95)
- Offline support: full functionality without network

### Security
- HTTPS only
- JWT authentication
- Password hashing (bcrypt)
- No sensitive data in logs

### Compatibility
- iOS 17.0+
- iPhone SE (2nd gen) and newer
- Light and dark mode

## Timeline

| Phase | Features | Duration |
|-------|----------|----------|
| Phase 1 | Habits + Auth | 6 weeks |
| Phase 2 | Tasks | 4 weeks |
| Phase 3 | Budget | 4 weeks |
| Phase 4 | Premium | 4 weeks |

## Open Questions

1. Should we support Apple Sign-In in MVP?
2. What's the maximum number of habits allowed (free tier)?
3. Should streaks reset on miss or allow "grace days"?

## Appendix

- [User Personas](./personas.md)
- [User Stories](./user-stories.md)
- [Architecture Overview](/docs/architecture/OVERVIEW.md)

# Feature Specification: Habits

## Overview

Habits are the core feature of HabitFlow. Users create recurring habits (daily, weekly, monthly) and track their completion over time.

## User Stories

See [/docs/product/user-stories.md](/docs/product/user-stories.md) for full user stories.

Key stories:
- US-010: Create Habit
- US-011: View Habits List
- US-012: Complete Habit
- US-020: View Current Streak

## Business Rules

### Habit Creation

| Rule | Description |
|------|-------------|
| BR-001 | Title is required, max 50 characters |
| BR-002 | Icon must be from predefined set (SF Symbols) |
| BR-003 | Color must be from predefined palette (12 colors) |
| BR-004 | Period: daily, weekly, or monthly |
| BR-005 | Target per period: 1-10, default 1 |
| BR-006 | Maximum 20 habits per user (free tier) |
| BR-007 | Habit titles don't need to be unique |

### Habit Completion

| Rule | Description |
|------|-------------|
| BR-010 | Can complete for today and past dates |
| BR-011 | Cannot complete for future dates |
| BR-012 | Can complete multiple times per day (up to target) |
| BR-013 | Completion value must be positive integer |
| BR-014 | Uncomplete removes the completion record |

### Streak Calculation

| Rule | Description |
|------|-------------|
| BR-020 | Streak = consecutive periods with >= target completions |
| BR-021 | Daily: consecutive calendar days |
| BR-022 | Weekly: consecutive ISO weeks (Mon-Sun) |
| BR-023 | Monthly: consecutive calendar months |
| BR-024 | Missing a period resets current streak to 0 |
| BR-025 | Longest streak is preserved even after reset |
| BR-026 | Today doesn't break streak until end of day |

### Period Mapping

| Period | Displayed In | Group By |
|--------|--------------|----------|
| daily | Today, Week, Month | Day |
| weekly | Week, Month | ISO Week |
| monthly | Month | Month |

## Data Model

### Habit Entity

```go
type Habit struct {
    ID              uuid.UUID
    UserID          uuid.UUID
    Title           string    // max 50 chars
    Icon            string    // SF Symbol name
    Color           string    // Color name
    Period          string    // daily, weekly, monthly
    TargetPerPeriod int       // 1-10
    SortOrder       int
    CreatedAt       time.Time
    UpdatedAt       time.Time
    DeletedAt       *time.Time
}
```

### HabitCompletion Entity

```go
type HabitCompletion struct {
    ID             uuid.UUID
    UserID         uuid.UUID
    HabitID        uuid.UUID
    CompletionDate time.Time // Date only (no time)
    Value          int       // Usually 1
    CreatedAt      time.Time
    UpdatedAt      time.Time
    DeletedAt      *time.Time
}
```

### Computed Fields

```go
type HabitWithStats struct {
    Habit
    CurrentStreak   int
    LongestStreak   int
    CompletedToday  bool
    CompletionCount int  // For current period
}
```

## API Mapping

| Action | Endpoint | Method |
|--------|----------|--------|
| List habits | `/api/v1/habits` | GET |
| Create habit | `/api/v1/habits` | POST |
| Get habit | `/api/v1/habits/{id}` | GET |
| Update habit | `/api/v1/habits/{id}` | PUT |
| Delete habit | `/api/v1/habits/{id}` | DELETE |
| Complete | `/api/v1/habits/{id}/complete` | POST |
| Uncomplete | `/api/v1/habits/{id}/uncomplete` | POST |

See [OpenAPI spec](/docs/api/openapi.yaml) for request/response formats.

## UI Screens

### 1. Today Screen (Main)

Shows all habits due today:
- Daily habits (always shown)
- Weekly habits (if not completed this week)
- Monthly habits (if not completed this month)

**Elements**:
- Segmented control: Today / Week / Month
- Habit card: icon, title, streak, completion status
- FAB: Add new habit
- Empty state if no habits

### 2. Habit List Screen

Full list of all habits grouped by period.

**Elements**:
- Section headers: Daily / Weekly / Monthly
- Drag to reorder
- Swipe to delete
- Tap to view details

### 3. Create/Edit Habit Screen

Form for habit creation/editing.

**Elements**:
- Title input
- Icon picker (grid of SF Symbols)
- Color picker (12 color circles)
- Period selector (segmented control)
- Target stepper (1-10)
- Save button
- Delete button (edit only)

### 4. Habit Detail Screen

Detailed view of a single habit.

**Elements**:
- Header: icon, title, color
- Stats: current streak, longest streak, completion rate
- Calendar heatmap (last 3 months)
- History list (last 30 completions)
- Edit button

## Edge Cases

### Offline Behavior

| Scenario | Behavior |
|----------|----------|
| Create habit offline | Saved locally, synced when online |
| Complete habit offline | Saved locally, synced when online |
| Conflict on sync | Server wins (last-write-wins) |
| Deleted on server | Remove from local |

### Timezone Handling

| Scenario | Behavior |
|----------|----------|
| User changes timezone | Completions stay on original dates |
| Completion at midnight | Use user's timezone to determine day |
| API stores dates | UTC, converted on display |

### Streak Edge Cases

| Scenario | Behavior |
|----------|----------|
| Complete for past date | Recalculate streak from that date |
| Uncomplete past date | Recalculate, may break streak |
| First day of habit | Streak = 1 if completed |
| Skip days then resume | Streak resets to 1 |
| Weekly on week boundary | Streak intact if completed in each week |

### Validation Edge Cases

| Input | Behavior |
|-------|----------|
| Empty title | Show error, don't save |
| Title > 50 chars | Truncate or show error |
| Whitespace-only title | Trim, treat as empty |
| Emoji in title | Allowed |
| Target = 0 | Default to 1 |
| Target > 10 | Cap at 10 |

## Test Scenarios

### Create Habit

```
GIVEN I am logged in
WHEN I tap "Add Habit"
AND I enter title "Morning Run"
AND I select icon "figure.run"
AND I select color "green"
AND I select period "daily"
AND I tap Save
THEN I see the habit in my Today list
AND the habit has 0 streak
AND the habit is not completed
```

### Complete Habit

```
GIVEN I have habit "Morning Run" uncompleted today
WHEN I tap on the habit
THEN I see checkmark animation
AND I feel haptic feedback
AND streak shows "1 day"
AND habit appears completed (dimmed)
```

### Streak Calculation

```
GIVEN I have daily habit with 7-day streak
AND I completed yesterday
WHEN I complete today before midnight
THEN streak shows "8 days"

GIVEN I have daily habit with 7-day streak
AND I missed yesterday
WHEN I check today
THEN streak shows "0 days"
AND longest streak still shows "7 days"
```

### Offline Sync

```
GIVEN I am offline
WHEN I create habit "Read"
THEN habit appears in local list
WHEN I go online
THEN habit syncs to server
AND habit gets server UUID
```

## Icons (Predefined Set)

```swift
let habitIcons = [
    // Fitness
    "figure.run", "figure.walk", "dumbbell.fill", "heart.fill",
    // Wellness
    "bed.double.fill", "drop.fill", "brain.head.profile", "lungs.fill",
    // Learning
    "book.fill", "pencil", "graduationcap.fill", "lightbulb.fill",
    // Productivity
    "checkmark.circle.fill", "target", "clock.fill", "calendar",
    // Health
    "pill.fill", "cross.case.fill", "stethoscope", "waveform.path.ecg",
    // Social
    "person.2.fill", "phone.fill", "message.fill", "hand.wave.fill",
    // Finance
    "dollarsign.circle.fill", "creditcard.fill", "chart.line.uptrend.xyaxis",
    // Creativity
    "paintbrush.fill", "music.note", "camera.fill", "video.fill",
    // Mindfulness
    "leaf.fill", "sun.max.fill", "moon.fill", "sparkles"
]
```

## Colors (Predefined Palette)

```swift
let habitColors = [
    "red": "#FF6B6B",
    "orange": "#FFA94D",
    "yellow": "#FFD43B",
    "green": "#51CF66",
    "teal": "#20C997",
    "cyan": "#22B8CF",
    "blue": "#339AF0",
    "indigo": "#5C7CFA",
    "violet": "#845EF7",
    "pink": "#F06595",
    "gray": "#868E96",
    "dark": "#495057"
]
```

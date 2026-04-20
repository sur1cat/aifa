# Feature Specification: Tasks

> **Status**: Phase 2 (Not in MVP)

## Overview

Tasks is a simple todo list focused on quick capture and "today" view. Unlike habits, tasks are one-time items.

## User Stories

### US-T01: Create Task
**As a** user
**I want to** quickly add a task
**So that** I don't forget things I need to do

### US-T02: View Today Tasks
**As a** user
**I want to** see tasks due today
**So that** I know what to focus on

### US-T03: Complete Task
**As a** user
**I want to** mark tasks as done
**So that** I can track progress

### US-T04: Set Due Date
**As a** user
**I want to** set when a task is due
**So that** it appears on the right day

## Business Rules

| Rule | Description |
|------|-------------|
| BT-001 | Title required, max 200 characters |
| BT-002 | Due date is optional |
| BT-003 | No due date = appears in "Inbox" |
| BT-004 | Maximum 100 active tasks |
| BT-005 | Completed tasks hidden after 24 hours |
| BT-006 | Tasks can be reordered manually |

## Data Model

```go
type Task struct {
    ID          uuid.UUID
    UserID      uuid.UUID
    Title       string
    Status      string    // pending, completed
    DueDate     *time.Time
    SortOrder   int
    CompletedAt *time.Time
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   *time.Time
}
```

## API Endpoints

| Action | Endpoint | Method |
|--------|----------|--------|
| List tasks | `/api/v1/tasks` | GET |
| Create task | `/api/v1/tasks` | POST |
| Update task | `/api/v1/tasks/{id}` | PUT |
| Delete task | `/api/v1/tasks/{id}` | DELETE |
| Complete | `/api/v1/tasks/{id}/complete` | POST |
| Reopen | `/api/v1/tasks/{id}/reopen` | POST |

## UI Screens

### 1. Tasks Screen

**Sections**:
- Today (due today)
- Upcoming (future due dates)
- Inbox (no due date)
- Completed (collapsible)

**Interactions**:
- Swipe right to complete
- Swipe left to delete
- Tap to edit
- Long press to set due date

### 2. Quick Add

**Behavior**:
- Text field at top of screen
- Enter to add
- Natural language parsing: "Call mom tomorrow"

## Dependencies

- Requires Phase 1 (Auth + Habits) complete
- Shares sync infrastructure
- Same offline-first approach

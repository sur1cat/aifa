# Feature Specification: Budget

> **Status**: Phase 3 (Not in MVP)

## Overview

Budget helps users track income and expenses with simple categorization and savings goals.

## User Stories

### US-B01: Log Transaction
**As a** user
**I want to** quickly log income or expense
**So that** I know where my money goes

### US-B02: View Monthly Summary
**As a** user
**I want to** see my monthly spending
**So that** I understand my financial health

### US-B03: Set Savings Goal
**As a** user
**I want to** save toward specific goals
**So that** I stay motivated to save

## Business Rules

| Rule | Description |
|------|-------------|
| BB-001 | Amount required, positive number |
| BB-002 | Type: income or expense |
| BB-003 | Category optional (defaults to "Uncategorized") |
| BB-004 | Maximum 12 custom categories |
| BB-005 | Currency is per-user setting |
| BB-006 | Savings goals have target amount and optional deadline |

## Data Model

```go
type Transaction struct {
    ID         uuid.UUID
    UserID     uuid.UUID
    CategoryID *uuid.UUID
    Amount     decimal.Decimal
    Type       string    // income, expense
    Note       string
    Date       time.Time
    CreatedAt  time.Time
    UpdatedAt  time.Time
    DeletedAt  *time.Time
}

type Category struct {
    ID          uuid.UUID
    UserID      uuid.UUID
    Name        string
    Icon        string
    Color       string
    BudgetLimit *decimal.Decimal
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   *time.Time
}

type SavingsGoal struct {
    ID            uuid.UUID
    UserID        uuid.UUID
    Name          string
    TargetAmount  decimal.Decimal
    CurrentAmount decimal.Decimal
    Deadline      *time.Time
    CreatedAt     time.Time
    UpdatedAt     time.Time
    DeletedAt     *time.Time
}
```

## UI Screens

### 1. Budget Overview

**Elements**:
- Month selector
- Income vs Expense chart
- Top spending categories
- Savings goals progress

### 2. Add Transaction

**Elements**:
- Amount keypad
- Type toggle (income/expense)
- Category picker
- Date picker
- Note field

### 3. Savings Goals

**Elements**:
- Goal cards with progress bar
- Add to goal button
- Create new goal

## Dependencies

- Requires Phase 2 (Tasks) complete
- Needs decimal precision handling
- Consider financial data security

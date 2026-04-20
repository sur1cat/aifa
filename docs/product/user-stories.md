# User Stories

## Epic: Authentication

### US-001: User Registration
**As a** new user
**I want to** create an account with email and password
**So that** I can save my habits and access them later

**Acceptance Criteria**:
- [ ] User can enter email and password
- [ ] Email must be valid format
- [ ] Password must be at least 8 characters
- [ ] Password must contain at least one number
- [ ] Error messages are clear and specific
- [ ] Success redirects to empty habits list
- [ ] Email is stored lowercase

**Notes**: Email verification deferred to Phase 2

---

### US-002: User Login
**As a** registered user
**I want to** log in with my credentials
**So that** I can access my habits

**Acceptance Criteria**:
- [ ] User can enter email and password
- [ ] Invalid credentials show generic error (security)
- [ ] Success redirects to habits list
- [ ] JWT token stored securely in Keychain
- [ ] Token refreshes automatically before expiry

---

### US-003: User Logout
**As a** logged-in user
**I want to** log out of the app
**So that** I can secure my account on shared devices

**Acceptance Criteria**:
- [ ] Logout option in settings
- [ ] Confirmation dialog before logout
- [ ] All local tokens are cleared
- [ ] User redirected to login screen

---

## Epic: Habit Management

### US-010: Create Habit
**As a** user
**I want to** create a new habit
**So that** I can track something I want to do regularly

**Acceptance Criteria**:
- [ ] User can enter habit title (required, max 50 chars)
- [ ] User can select icon from predefined set
- [ ] User can select color from predefined palette
- [ ] User can select period: daily (default), weekly, monthly
- [ ] User can set target completions per period (default: 1)
- [ ] Habit is saved and appears in list immediately
- [ ] Empty title shows validation error
- [ ] Maximum 20 habits (free tier)

---

### US-011: View Habits List
**As a** user
**I want to** see my habits for today/week/month
**So that** I can know what I need to do

**Acceptance Criteria**:
- [ ] Segmented control: Today / Week / Month
- [ ] Today shows daily habits + weekly habits if day matches + monthly habits if day matches
- [ ] Week shows weekly habits with progress
- [ ] Month shows monthly habits with progress
- [ ] Each habit shows: icon, title, completion status, streak
- [ ] Completed habits are visually distinct (checkmark, muted)
- [ ] Empty state if no habits created

---

### US-012: Complete Habit
**As a** user
**I want to** mark a habit as complete
**So that** I can track my progress

**Acceptance Criteria**:
- [ ] Tap habit card to complete
- [ ] Swipe right to complete (alternative)
- [ ] Haptic feedback on completion
- [ ] Celebratory animation (confetti for streak milestones)
- [ ] Undo toast appears for 3 seconds
- [ ] Streak count updates immediately
- [ ] Completion syncs to server

---

### US-013: Uncomplete Habit
**As a** user
**I want to** undo a habit completion
**So that** I can fix mistakes

**Acceptance Criteria**:
- [ ] Tap completed habit to uncomplete
- [ ] Swipe left to uncomplete (alternative)
- [ ] Undo via toast notification
- [ ] Confirmation if past 3 seconds
- [ ] Streak recalculates correctly
- [ ] Syncs to server

---

### US-014: Edit Habit
**As a** user
**I want to** edit an existing habit
**So that** I can update its details

**Acceptance Criteria**:
- [ ] Long press habit to show edit option
- [ ] Can edit: title, icon, color, period, target
- [ ] Cannot change period if completions exist (warning)
- [ ] Changes save immediately
- [ ] Cancel returns to previous state

---

### US-015: Delete Habit
**As a** user
**I want to** delete a habit I no longer need
**So that** my list stays clean

**Acceptance Criteria**:
- [ ] Delete option in edit mode or swipe
- [ ] Confirmation dialog: "Delete [habit name]?"
- [ ] Shows warning about losing history
- [ ] Soft delete (can be restored within 30 days)
- [ ] Removed from list immediately

---

### US-016: Reorder Habits
**As a** user
**I want to** reorder my habits
**So that** important ones appear first

**Acceptance Criteria**:
- [ ] Drag and drop to reorder
- [ ] Order persists across sessions
- [ ] Order is per-period (daily order, weekly order)

---

## Epic: Streaks

### US-020: View Current Streak
**As a** user
**I want to** see my current streak for each habit
**So that** I stay motivated to maintain it

**Acceptance Criteria**:
- [ ] Streak count shown on habit card
- [ ] Streak = consecutive periods with at least target completions
- [ ] Daily habit: consecutive days
- [ ] Weekly habit: consecutive weeks
- [ ] Monthly habit: consecutive months
- [ ] Updates in real-time on completion

---

### US-021: Streak Milestones
**As a** user
**I want to** celebrate streak milestones
**So that** I feel accomplished

**Acceptance Criteria**:
- [ ] Special animation at: 7, 14, 30, 60, 90, 180, 365 days
- [ ] Milestone badge on habit card
- [ ] Historical milestones visible in habit detail

---

## Epic: Statistics

### US-030: View Completion Rate
**As a** user
**I want to** see my completion rate
**So that** I understand my consistency

**Acceptance Criteria**:
- [ ] Overall completion rate on stats screen
- [ ] Per-habit completion rate
- [ ] Timeframe selector: week, month, year, all time
- [ ] Visual representation (percentage, bar)

---

### US-031: View Calendar Heatmap
**As a** user
**I want to** see a calendar heatmap
**So that** I can visualize my habits over time

**Acceptance Criteria**:
- [ ] Month calendar view
- [ ] Each day colored by completion count
- [ ] Color intensity: 0 = gray, 1-2 = light, 3+ = dark
- [ ] Tap day to see details
- [ ] Navigate between months

---

## Priority Matrix

| Story | Priority | Effort | MVP |
|-------|----------|--------|-----|
| US-001 | P0 | M | Yes |
| US-002 | P0 | S | Yes |
| US-003 | P1 | S | Yes |
| US-010 | P0 | M | Yes |
| US-011 | P0 | M | Yes |
| US-012 | P0 | M | Yes |
| US-013 | P1 | S | Yes |
| US-014 | P1 | M | Yes |
| US-015 | P1 | S | Yes |
| US-016 | P2 | M | No |
| US-020 | P1 | M | Yes |
| US-021 | P2 | S | No |
| US-030 | P1 | M | Yes |
| US-031 | P2 | L | No |

**Legend**: P0 = Must have, P1 = Should have, P2 = Nice to have
**Effort**: S = Small (1-2 days), M = Medium (3-5 days), L = Large (1+ week)

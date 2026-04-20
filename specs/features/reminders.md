# Feature: Напоминания для привычек

**Status:** Ready for Development
**Priority:** Must Have
**Created:** 2025-12-26

---

## Product Manager

### Проблема
Пользователи создают привычки, но забывают их выполнять. Без напоминаний сложно выработать рутину — человек вспоминает о привычке только вечером, когда уже поздно.

### User Story
> As a user, I want to receive reminders for my habits at a specific time, so that I don't forget to complete them and can build consistent routines.

### Acceptance Criteria
- [ ] Пользователь может включить/выключить напоминание для привычки
- [ ] Пользователь может выбрать время напоминания
- [ ] Push-уведомление приходит в указанное время
- [ ] Уведомление показывает название привычки и иконку
- [ ] Тап по уведомлению открывает раздел Habits
- [ ] Напоминания работают без интернета (локально)

### Метрики успеха
- Completion rate привычек +20%
- Retention Day 7 +15%
- % пользователей с включенными напоминаниями > 50%

---

## Designer

### User Flow
```
1. Habits List → Tap habit row → Habit Detail Sheet
2. Habit Detail → Toggle "Reminder" → Show Time Picker
3. Select time → Save → Return to list (badge shows reminder set)
```

### UI Components

#### HabitRow — индикатор напоминания
```
[🔔] 08:00    ← маленький badge если напоминание включено
```

#### Habit Detail Sheet (новый)
```
┌─────────────────────────────────────┐
│  ✕                                  │
│                                     │
│         🎯                          │
│     Утренняя зарядка                │
│     Daily • 5 day streak            │
│                                     │
├─────────────────────────────────────┤
│  🔔 Reminder                    ○───│  ← Toggle
│                                     │
│  ⏰ Time                      08:00 │  ← только если toggle ON
│                                     │
├─────────────────────────────────────┤
│  🗑️ Delete Habit                    │
└─────────────────────────────────────┘
```

#### Time Picker
Нативный iOS wheel picker в inline стиле.

#### Push Notification
```
┌─────────────────────────────────────┐
│ Atoma                          now  │
│ 🎯 Утренняя зарядка                 │
│ Time to build your routine!         │
└─────────────────────────────────────┘
```

### States
- **Reminder OFF**: toggle выключен, time picker скрыт
- **Reminder ON**: toggle включён, показан time picker
- **Permission denied**: alert с кнопкой "Open Settings"

---

## Architect

### Решение
Полностью локальное (iOS Local Notifications). Без backend.

**Почему:**
- Напоминания не требуют синхронизации
- Работает offline
- Проще реализация
- Нет нагрузки на сервер

### Data Model

```swift
// Models.swift - добавить в Habit
struct Habit: Identifiable, Codable, Sendable {
    // ... existing fields
    var reminderEnabled: Bool = false
    var reminderTime: Date? = nil  // Только час:минуты
}
```

### NotificationManager

```swift
// Core/Notifications/NotificationManager.swift
@MainActor
class NotificationManager: ObservableObject {
    static let shared = NotificationManager()

    func requestPermission() async -> Bool
    func scheduleHabitReminder(habit: Habit) async
    func cancelHabitReminder(habitID: UUID)
    func rescheduleAllReminders()
}
```

### Data Flow
```
User toggles reminder ON
    ↓
NotificationManager.requestPermission()
    ↓
User selects time
    ↓
DataManager.updateHabit() — сохраняет reminderTime
    ↓
NotificationManager.scheduleHabitReminder()
    ↓
iOS schedules UNNotificationRequest (repeating)
```

### Files to Create/Modify

| File | Action |
|------|--------|
| `Core/Models/Models.swift` | Add reminder fields to Habit |
| `Core/Notifications/NotificationManager.swift` | NEW |
| `Features/Habits/HabitDetailSheet.swift` | NEW |
| `Features/Habits/HabitsView.swift` | Modify — add tap, badge |
| `HabitFlowApp.swift` | Request permission |

### Edge Cases
- Permission denied → alert с кнопкой Settings
- App killed → iOS сохраняет notifications
- Habit deleted → cancel notification
- Period changed → reschedule

---

## QA

**Checklist:** [/checklists/reminders.md](/checklists/reminders.md)

### Key Test Cases
1. TC-001: Открытие detail sheet
2. TC-002: Включение напоминания (permission flow)
3. TC-003: Выбор времени
4. TC-004: Получение push-уведомления
5. TC-007: Отключение напоминания
6. TC-009: Сохранение после перезапуска
7. TC-011: Permission denied flow

### Edge Cases
- Длинное название привычки
- Несколько привычек с одинаковым временем
- Weekly/monthly привычки

---

## Status

| Stage | Status | Date |
|-------|--------|------|
| PM | ✅ Done | 2025-12-26 |
| Designer | ✅ Done | 2025-12-26 |
| Architect | ✅ Done | 2025-12-26 |
| Developer | ✅ Done | 2025-12-26 |
| QA | 🔄 Testing | 2025-12-26 |
| DevOps | ⏭️ N/A | - |

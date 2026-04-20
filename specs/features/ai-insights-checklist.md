# AI Insights Implementation Checklist

## Phase 1: Unlock System (1-2 дня)

### iOS
- [ ] `InsightStatus` модель в Models.swift
- [ ] Трекинг дней активности в DataManager
- [ ] `InsightProgressCard` компонент (locked state)
- [ ] `InsightUnlockView` (celebration момент)
- [ ] Добавить карточку на главные экраны (Habits, Tasks, Budget)
- [ ] Сохранение прогресса в UserDefaults

### Логика unlock
```
Habits:  firstHabitDate + 14 days
Tasks:   firstTaskDate + 7 days
Budget:  firstTransactionDate + 30 days OR transactionCount >= 20
```

---

## Phase 2: Локальные инсайты (2-3 дня)

### Habits Insights (без AI)
- [ ] `HabitInsightGenerator` класс
- [ ] Алгоритм: best time of day
- [ ] Алгоритм: weekly skip patterns
- [ ] Алгоритм: streak analysis

### Tasks Insights
- [ ] `TaskInsightGenerator` класс
- [ ] Алгоритм: completion rate
- [ ] Алгоритм: best productive day
- [ ] Алгоритм: overcommitment detection

### Budget Insights
- [ ] `BudgetInsightGenerator` класс
- [ ] Алгоритм: top spending category
- [ ] Алгоритм: month-over-month trend
- [ ] Алгоритм: recurring transaction detection

### UI
- [ ] `InsightCard` компонент
- [ ] Действия: dismiss, tap action
- [ ] `InsightsListView` (все инсайты)

---

## Phase 3: AI инсайты (3-4 дня)

### Backend
- [ ] `POST /api/v1/insights/generate` endpoint
- [ ] OpenAI prompt для анализа паттернов
- [ ] Сохранение инсайтов в БД

### iOS
- [ ] InsightsService для API
- [ ] Sync инсайтов с сервера
- [ ] Background refresh

---

## Phase 4: Weekly Review (2 дня)

- [ ] `WeeklyReviewView` UI
- [ ] Генерация каждое воскресенье
- [ ] Push notification "Твоя неделя готова"
- [ ] Сравнение с прошлой неделей

---

## Приоритет реализации

| # | Что | Зачем | Срок |
|---|-----|-------|------|
| 1 | Phase 1 | Создаёт anticipation, retention | 1-2 дня |
| 2 | Phase 2 | Первая ценность без сервера | 2-3 дня |
| 3 | Phase 4 | Weekly review = habit открывать app | 2 дня |
| 4 | Phase 3 | AI polish, не критично для MVP | 3-4 дня |

**Total MVP (Phase 1+2):** ~4-5 дней

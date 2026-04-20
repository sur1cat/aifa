# habit-service

Owns habit tracking: periodic habits, target-value progress, and streak
calculation.

## Endpoints

Mount point: `/api/v1/habits/*`.

| Method | Path | Description |
|---|---|---|
| GET | `/habits` | List caller's habits |
| POST | `/habits` | Create habit |
| GET | `/habits/:id` | Get habit |
| PUT | `/habits/:id` | Update habit (partial — empty fields ignored) |
| DELETE | `/habits/:id` | Delete habit |
| POST | `/habits/:id/toggle` | Add/remove completion for a date; or set progress value |
| GET | `/health` | Liveness |

Toggle body:
```json
{ "date": "2026-04-21", "value": 5 }   // progress habit: sets value=5, marks complete if value >= target
{ "date": "2026-04-21" }               // boolean habit: flips completion state
```

## Events

Subscribes to NATS:
- `user.deleted` — deletes all habits for the user (cascades to completions + progress).
- `goal.deleted` — nulls out `goal_id` on any habit pointing at the deleted goal.

Publishing: none (for now).

## Storage

Postgres schema `habits`:
- `habits (id, user_id, goal_id?, title, icon, color, period, target_value?, unit?, archived_at?, timestamps)`
- `habit_completions (habit_id, completed_date)` — daily boolean marks
- `habit_progress (habit_id, progress_date, progress_value)` — for habits with a target

No cross-schema foreign keys. `user_id` and `goal_id` are plain UUIDs;
referential integrity is maintained via NATS events.

## Run

```bash
cp .env.example .env
make tidy && make run
```

Via root compose:
```bash
cd ../.. && docker compose up habit-service
```

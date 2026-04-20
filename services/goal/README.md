# goal-service

CRUD for long-term goals. Each goal is owned by a user, has an icon,
optional target value + unit (e.g. "10 books"), and an optional deadline.

## Endpoints

Mount point: `/api/v1/goals/*`.

| Method | Path | Description |
|---|---|---|
| GET | `/goals` | List caller's goals |
| POST | `/goals` | Create |
| GET | `/goals/:id` | Read |
| PUT | `/goals/:id` | Update (partial — empty strings ignored) |
| DELETE | `/goals/:id` | Delete + publish `goal.deleted` |
| GET | `/health` | Liveness |

## Events

Publishes:
- `goal.deleted` — `{ goal_id, user_id }`. habit-service subscribes to null out `goal_id` on habits pointing to the deleted goal.

Subscribes:
- `user.deleted` — purges all goals for the user.

## Storage

Postgres schema `goals`: single table `goals` (id, user_id, title, icon,
target_value?, unit?, deadline?, archived_at?, timestamps). No FK to users.

## Run

```bash
cp .env.example .env
make tidy && make run
```

Via root compose: `cd ../.. && docker compose up goal-service`.

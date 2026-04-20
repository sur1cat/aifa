# task-service

CRUD for dated tasks with priority (low/medium/high/urgent) and a
due-date field. Ordering:
- default list: priority (urgent→low), then due_date ASC, then newest
- date-filtered list: incomplete first, priority, then newest

## Endpoints

Mount point: `/api/v1/tasks/*`.

| Method | Path | Description |
|---|---|---|
| GET | `/tasks` | List; optional `?date=YYYY-MM-DD&limit=&offset=` |
| POST | `/tasks` | Create (priority defaults to `medium`, due_date to today) |
| GET | `/tasks/:id` | Read |
| PUT | `/tasks/:id` | Partial update |
| DELETE | `/tasks/:id` | Delete |
| POST | `/tasks/:id/toggle` | Flip `is_completed` |
| GET | `/health` | Liveness |

## Events

Subscribes to NATS:
- `user.deleted` — deletes all tasks for that user.

Publishing: none.

## Storage

Postgres schema `tasks`, single table with `(user_id, due_date, priority)`
indexes.

## Run

```bash
cp .env.example .env
make tidy && make run
```

Via root compose: `cd ../.. && docker compose up task-service`.

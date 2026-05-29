# goal-service

Owns **financial goals**: long-term money targets the user is working
toward. Four kinds:

- `savings`    — accumulate money toward a number (emergency fund,
                 vacation fund, anything you save up for).
- `debt`       — pay down a balance. `current_amount` tracks how much
                 has been paid; reaching `target_amount` clears it.
- `purchase`   — discrete one-time purchase to be made when
                 `current_amount` reaches `target_amount`.
- `investment` — track contributions into investments (NOT market
                 value — just what was deposited).

Generic non-monetary goals from the legacy schema have been migrated
into the savings type with their old `target_value` copied into
`target_amount` and the default currency.

## Endpoints

Mount point: `/api/v1/goals/*`.

| Method | Path | Description |
|---|---|---|
| GET | `/goals` | List caller's goals (with progress) |
| POST | `/goals` | Create — `goal_type` defaults to `savings`, `currency` to `DEFAULT_CURRENCY` env |
| GET | `/goals/:id` | Read |
| PUT | `/goals/:id` | Partial update — including incrementing `current_amount` |
| DELETE | `/goals/:id` | Delete + publish `goal.deleted` |
| GET | `/health` | Liveness |

Create body example (savings goal):
```json
{
  "title": "Emergency fund",
  "goal_type": "savings",
  "target_amount": 10000,
  "current_amount": 1500,
  "currency": "USD",
  "deadline": "2026-12-31T00:00:00Z",
  "icon": "🛟"
}
```

Response includes a derived `progress` (0..1) so clients don't have
to do the math:
```json
{
  "data": {
    "id": "...",
    "goal_type": "savings",
    "target_amount": 10000,
    "current_amount": 1500,
    "currency": "USD",
    "progress": 0.15,
    ...
  }
}
```

## Events

Publishes:
- `goal.deleted` — `{ goal_id, user_id }`. habit-service nulls `goal_id`
  on any habits linked to the deleted goal.

Subscribes:
- `user.deleted` — purges all goals for the user.

## Storage

Postgres schema `goals`:
- adds `goal_type`, `target_amount` (DECIMAL), `current_amount`,
  `currency` to the existing `goals` table
- drops legacy `target_value` (INTEGER) and `unit` — superseded by
  `target_amount` + `currency`. The migration migrates any existing
  values into the new columns first.

## Run

```bash
cp .env.example .env
make tidy && make run
```

Via root compose: `cd ../.. && docker compose up goal-service`.

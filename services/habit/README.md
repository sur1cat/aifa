# habit-service

Owns **financial habits** — repeating behaviors with a budgetary impact.
Three kinds:

- `tracking` — improves financial awareness ("Log every expense"). No
  direct money flow attached.
- `saving`  — completion is a deposit toward savings ("Skip Starbucks
  today: +$5"). `expected_amount` is the per-completion contribution.
- `spending` — the user keeps within a category cap ("Stay under
  $30/day on food"). `expected_amount` is the cap; staying under for
  the period counts as a completion.

Boolean toggle, target-value progress, and streak calculation work as
before. New columns are additive: existing rows still load and behave
identically.

## Endpoints

Mount point: `/api/v1/habits/*`.

| Method | Path | Description |
|---|---|---|
| GET | `/habits` | List caller's habits |
| POST | `/habits` | Create — `kind` defaults to `tracking`, `currency` to `DEFAULT_CURRENCY` env |
| GET | `/habits/:id` | Get habit |
| PUT | `/habits/:id` | Partial update |
| DELETE | `/habits/:id` | Delete |
| POST | `/habits/:id/toggle` | Add/remove completion or set progress value |
| GET | `/health` | Liveness |

Create body example (saving habit):
```json
{
  "title": "Pack lunch",
  "kind": "saving",
  "currency": "USD",
  "financial_category": "Food",
  "expected_amount": 12.50,
  "period": "daily",
  "icon": "🥪"
}
```

## Events

Subscribes:
- `user.deleted` — purge all habits for the user.
- `goal.deleted` — null out `goal_id` on any linked habits.

## Storage

Postgres schema `habits`:
- `habits` — adds `kind`, `currency`, `financial_category`, `expected_amount`
- `habit_completions`, `habit_progress` — unchanged

`financial_category` aligns with `transactions.category` from
finance-service so the AI agent can correlate habit completions with
actual spending.

## Run

```bash
cp .env.example .env
make tidy && make run
```

Via root compose: `cd ../.. && docker compose up habit-service`.

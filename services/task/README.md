# task-service

Owns **financial action items**: bill reminders, payment due dates,
invoice send-outs, subscription cancellations. Generic to-dos still
work — the financial fields are optional.

Each task has a `kind`:
- `todo`   — generic action item; `amount`/`category` may still be set as
             context but the iOS client doesn't try to create a transaction.
- `bill`   — money the user OWES (rent, electricity, subscription).
             Marking complete typically corresponds to making a payment;
             the client may emit a matching expense transaction.
- `income` — money the user is OWED (invoice, payday). Completion may
             yield an income transaction.

## Endpoints

Mount point: `/api/v1/tasks/*`.

| Method | Path | Description |
|---|---|---|
| GET | `/tasks` | List; optional `?date=YYYY-MM-DD&limit=&offset=` |
| POST | `/tasks` | Create — `kind` defaults to `todo`, `currency` to `DEFAULT_CURRENCY` env |
| GET | `/tasks/:id` | Read |
| PUT | `/tasks/:id` | Partial update |
| DELETE | `/tasks/:id` | Delete |
| POST | `/tasks/:id/toggle` | Atomic flip of `is_completed` |
| GET | `/health` | Liveness |

Create body example (bill reminder):
```json
{
  "title": "Pay electricity",
  "kind": "bill",
  "amount": 87.40,
  "currency": "USD",
  "category": "Utilities",
  "priority": "high",
  "due_date": "2026-05-01"
}
```

## Events

Subscribes to NATS:
- `user.deleted` — purges all tasks for the user.

Publishing: none (today). The client owns the bill→transaction handoff.

## Storage

Postgres schema `tasks`:
- adds `kind`, `amount`, `currency`, `category` to the existing `tasks` table

`category` aligns with `transactions.category` from finance-service so
spending follow-ups can be cross-referenced.

## Run

```bash
cp .env.example .env
make tidy && make run
```

Via root compose: `cd ../.. && docker compose up task-service`.

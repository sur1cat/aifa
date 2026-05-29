# finance-service

Owns the money side of AIFA: transactions, recurring schedules, and a
monthly savings-goal progress view. All three domains live together
because they share one thing — the `transactions` table (savings progress
and recurring projection both derive from it).

## Endpoints

Mounted at `/api/v1/*`.

**Transactions**
| Method | Path | Description |
|---|---|---|
| GET | `/transactions` | List; optional `?year=&month=&limit=&offset=` |
| POST | `/transactions` | Create (`type` income/expense, amount > 0) |
| GET | `/transactions/summary` | Month income + expense + balance |
| GET | `/transactions/:id` | Read |
| PUT | `/transactions/:id` | Partial update |
| DELETE | `/transactions/:id` | Delete |

**Recurring transactions** (subscriptions, regular bills)
| Method | Path | Description |
|---|---|---|
| GET | `/recurring-transactions` | List |
| POST | `/recurring-transactions` | Create (frequency: weekly/biweekly/monthly/quarterly/yearly) |
| GET | `/recurring-transactions/projection` | Monthly-normalized income/expense projection |
| POST | `/recurring-transactions/process` | Advance all due schedules forward; emit one concrete transaction per occurrence |
| GET `/:id` / PUT `/:id` / DELETE `/:id` | | standard CRUD |

**Savings goal**
| Method | Path | Description |
|---|---|---|
| GET | `/savings-goal` | Current goal with progress (null if unset) |
| POST | `/savings-goal` | Set monthly target |
| DELETE | `/savings-goal` | Remove |

## Events

Subscribes to NATS:
- `user.deleted` — purges transactions, recurring, and savings-goal rows.

Publishing: none yet. (Future: `transaction.created` consumed by
ai-service for background insight caching.)

## Concurrency guarantees

`POST /recurring-transactions/process` uses an atomic `UPDATE ... WHERE
next_date = $expected` (see `RecurringRepository.AdvanceIfDue`). That
means the scheduler-worker and an iOS client hitting the same endpoint
concurrently can never double-charge — whoever loses the CAS just stops
iterating.

## Run

```bash
cp .env.example .env
make tidy && make run
```

Via root compose: `cd ../.. && docker compose up finance-service`.

# scheduler-worker

A small cron container that publishes NATS heartbeat events. Service has
no HTTP API beyond `/health` and no database — it's a producer of timed
events that downstream services can react to.

## Schedules

| Schedule env | Default | Subject published |
|---|---|---|
| `CRON_RECURRING` | `*/5 * * * *` | `cron.recurring.tick` |
| `CRON_REMINDERS` | `0 9 * * *` | `cron.reminder.tick` |

Cron expressions are standard 5-field (no seconds), evaluated in UTC.
Empty value disables that schedule.

## Wiring

`cron.recurring.tick` is the canonical replacement for the
iOS-client-triggered `POST /api/v1/recurring-transactions/process`
endpoint. finance-service is expected to subscribe to it and advance
every user's due recurring rows on each tick. (The HTTP endpoint stays
available as a manual escape hatch.)

`cron.reminder.tick` has no subscribers today. Publishing it now means
notification-service or others can opt in later without scheduler
changes — the schedule contract is one-way.

## Run

```bash
cp .env.example .env
make tidy && make run
```

Via root compose: `cd ../.. && docker compose up scheduler-worker`.

## Why a separate service

Putting cron inside finance- or notification-service would tie
horizontally-scaled replicas together — multiple replicas would each
fire the same tick. A dedicated single-replica worker keeps the
schedule canonical and lets the consumer services scale freely.

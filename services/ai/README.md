# ai-service

Thin wrapper around OpenAI chat-completions for the AIFA app's coach and
insight endpoints. Stateless — no Postgres, no NATS; just Redis for JWT
blacklist lookups on incoming requests.

## Endpoints

Mount point: `/api/v1/ai/*`.

| Method | Path | Description |
|---|---|---|
| POST | `/ai/chat` | Agent-style chat (`habit_coach` / `task_assistant` / `finance_advisor` / `life_coach`) |
| POST | `/ai/insights` | Per-domain insights (`habits` / `tasks` / `budget` / `weekly`) |
| POST | `/ai/expense-analysis` | Spending patterns + questionable-transaction flags |
| POST | `/ai/goal-to-habits` | Convert an outcome goal into 2–4 process habits |
| POST | `/ai/goal-clarify` | Generate clarifying questions for a goal |
| GET | `/health` | Liveness |

Every endpoint accepts the user's data as a payload field (`context` /
`data`) — the client is responsible for gathering it from habit-,
task-, and finance-service before calling AI. The service never reaches
into those stores directly.

## Behavior

- When `OPENAI_API_KEY` is empty, all AI endpoints return `503 AI_ERROR`
  with body `OpenAI API key is not configured`. The service still starts
  and serves `/health` so the container stays healthy during partial
  outages.
- Prompts expect the model to emit JSON. If parsing fails, the handler
  falls back to `{ "data": { "raw": "<model text>" } }` so the iOS
  client can render the raw response instead of crashing.

## Run

```bash
cp .env.example .env   # paste OPENAI_API_KEY
make tidy && make run
```

Via root compose: `cd ../.. && docker compose up ai-service`.

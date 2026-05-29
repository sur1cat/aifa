# AIFA — Microservices Backend

Привычки, задачи, финансы — одним потоком. Мобильный клиент iOS,
бэкенд — набор независимых Go-микросервисов за Traefik-шлюзом.

## Архитектура

```
iOS / Web → Traefik (:8080) → one of:
  /api/v1/auth/*          → auth-service
  /api/v1/users/*         → user-service
  /api/v1/habits/*        → habit-service
  /api/v1/goals/*         → goal-service
  /api/v1/tasks/*         → task-service
  /api/v1/transactions/*  → finance-service
  /api/v1/recurring-transactions/* → finance-service
  /api/v1/savings-goal    → finance-service
  /api/v1/ai/*            → ai-service
  /api/v1/push/*          → notification-service

Инфра:  Postgres 16 (schema-per-service) · Redis · NATS JetStream
Крон:   scheduler-worker (recurring transactions, reminders)
```

Каждый сервис — отдельный Go-модуль со своим `go.mod`, `Dockerfile`,
CI-workflow и Helm-чартом. Подробнее: [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md).

## Запуск локально

```bash
cp .env.example .env
docker compose up -d
```

Traefik дашборд: http://localhost:8090. Публичный API: http://localhost:8080.

Остановить всё: `docker compose down`. Полностью сбросить БД: `docker compose down -v`.

## Разработка одного сервиса

Каждый сервис независим — заходишь в директорию и работаешь как с отдельным проектом:

```bash
cd services/auth
cp .env.example .env
make run          # go run ./cmd/server
make test         # go test ./...
make docker       # build image
make migrate-up   # apply schema migrations
```

Для запуска сервиса в изоляции (с инфрой из корня):

```bash
docker compose up -d postgres redis nats
cd services/auth
docker compose -f docker-compose.service.yml up --build
```

## Сервисы

| Сервис | Порт | Назначение |
|---|---|---|
| auth-service | 8001 | OAuth Google/Apple, JWT, logout blacklist |
| user-service | 8002 | Профили, удаление аккаунта |
| habit-service | 8003 | Привычки + completions |
| goal-service | 8004 | Долгосрочные цели |
| task-service | 8005 | Задачи |
| finance-service | 8006 | Транзакции, recurring, savings goal |
| ai-service | 8007 | OpenAI assistants |
| notification-service | 8008 | APNS push |
| scheduler-worker | — | Cron для recurring + reminders |

## Межсервисное взаимодействие

- **JWT**: все сервисы валидируют пользовательский токен локально через общий `JWT_SECRET` — без ходок в auth-service.
- **Logout**: auth-service пишет хэш токена в Redis (TTL = остаток жизни), остальные сервисы проверяют blacklist там же.
- **События**: `user.deleted`, `transaction.created` и др. — через NATS JetStream. Подписчики чистят/реагируют асинхронно.
- **Service-to-service REST**: ai-service дёргает habit/finance с отдельным `SERVICE_JWT_SECRET`, claim `service:<name>`.

## Миграции

Каждый сервис хранит миграции в `services/<name>/migrations/` и запускает их на старте `cmd/server`
против своей схемы в общей БД. Схемы создаются один раз в [`deploy/postgres/init.sql`](deploy/postgres/init.sql).

## CI/CD

Per-service GitHub Actions в `services/<name>/.github/workflows/ci.yml`:
- `go test ./...`
- `go build ./...`
- docker build + push (main only)
- опциональный Helm lint

## Helm

Чарты: `services/<name>/helm/<name>/`. Универсальный шаблон Deployment + Service + Ingress + ConfigMap + Secret. Values подставляются при деплое в конкретное окружение.

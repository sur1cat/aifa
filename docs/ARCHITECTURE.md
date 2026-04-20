# Architecture

## Decomposition rationale

Монолит `habitflow` имел четыре естественных границы: аутентификация,
управление сущностями пользователя (habits / goals / tasks), финансы,
AI/уведомления. Каждая граница вынесена в отдельный сервис, что даёт:

- независимый деплой и масштабирование (AI-сервис горизонтально под нагрузку OpenAI, finance — вертикально под Postgres);
- строгую локальность данных (каждая схема Postgres — одна команда);
- изоляцию инцидентов (падение ai-service не ломает логирование привычек).

## Принципы

1. **Один сервис — один go-модуль.** Не разделяем код через shared-пакеты. Мелкие утилиты (JWT verify, middleware, event bus wrapper) дублируются внутри `services/<name>/internal/<package>/`. Цена — копипаст; выгода — нулевая связность, возможность извлечь сервис в отдельный репо за `git subtree split`.
2. **Schema-per-service, общий Postgres инстанс.** Каждый сервис пишет только в свою схему, читает только свою. Кросс-доменные связи — по UUID без внешних ключей. Если понадобится — база легко разбивается на физические инстансы без изменения кода.
3. **JWT-валидация локальная.** Все сервисы знают `JWT_SECRET` и сами проверяют подпись. Auth-service занимается только выдачей и отзывом — без HTTP-вызовов под каждый запрос. Отозванные токены хранятся в Redis с TTL равным остатку жизни токена.
4. **События — через NATS JetStream.** Для side-effects, не лежащих на критическом пути запроса: удаление пользователя, постобработка транзакции, отложенные напоминания.
5. **Синхронный inter-service REST** — только когда нужен ответ в рамках одного запроса пользователя (AI за контекстом привычек/финансов). Подписан отдельным `SERVICE_JWT_SECRET` с claim `service:<name>`.

## Пользовательский JWT

```
header.payload.signature
payload:
  sub:  <user_uuid>
  typ:  access | refresh
  iat:  <unix>
  exp:  <unix>
  jti:  <uuid>   // используется для blacklist
```

Подпись: HS256, общий секрет `JWT_SECRET`. Access TTL — 30 дней, refresh — 365 (унаследовано из монолита, мобильный клиент редко переаутентифицируется).

## Logout

1. Клиент шлёт `POST /auth/logout` с access и refresh токенами.
2. auth-service кладёт `blacklist:<jti>` в Redis с TTL равным `exp - now`.
3. Остальные сервисы в JWT middleware проверяют `blacklist:<jti>` в Redis до обработки запроса. ≈1 мс overhead на запрос.

## События NATS

| Субъект | Издатель | Подписчики | Назначение |
|---|---|---|---|
| `user.deleted` | auth / user | habit, goal, task, finance, notification | Каскадное удаление данных пользователя |
| `transaction.created` | finance | ai (для инсайтов) | Прогрев кэша, триггер пересчётов |
| `reminder.due` | scheduler | notification | Отправка push |
| `recurring.processed` | finance | — | Аудит |

Стримы конфигурируются при старте сервиса (idempotent JetStream ensure-stream).

## Сетевой путь запроса

```
iOS
  └── POST /api/v1/habits/{id}/toggle
      Authorization: Bearer <access>
          │
          ▼
      Traefik (routes by path prefix via labels)
          │  + cors middleware
          │  + ratelimit-general (100/min/IP)
          │  + strip-api-prefix
          ▼
      habit-service:8003
          │  1. parse JWT (HS256 + JWT_SECRET)
          │  2. check Redis blacklist:<jti>
          │  3. inject user_id into ctx
          │  4. handler → repository → postgres.habits schema
          ▼
      200 OK { completed: true, ... }
```

## Схема каталогов одного сервиса

```
services/auth/
├── cmd/server/main.go            # composition root
├── internal/
│   ├── config/                   # env-based config
│   ├── domain/                   # структуры сущностей этого сервиса
│   ├── handler/                  # Gin-handlers
│   ├── middleware/               # JWT, ratelimit, recover
│   ├── oauth/                    # google / apple verifiers
│   ├── repository/               # SQL (pgxpool)
│   └── service/                  # business logic
├── migrations/                   # *.up.sql / *.down.sql — запускаются на старте
├── helm/auth-service/            # Deployment, Service, Ingress, Values
├── .github/workflows/ci.yml      # test + build + push
├── Dockerfile
├── docker-compose.service.yml    # standalone-запуск с внешней инфрой
├── .env.example
├── Makefile
├── README.md
├── go.mod
└── go.sum
```

## Что сознательно НЕ делаем

- **Общий Go-модуль.** Даже ради JWT-утилиты. Копипаст дешевле связности.
- **Database-per-service физически.** На схеме границ хватает. Физическое разделение — когда нагрузка потребует.
- **gRPC между сервисами.** REST+JSON понятнее и дешевле. Если P99 между сервисами станет узким местом — мигрируем точечно.
- **API-композиция в gateway.** Traefik только маршрутизирует. Композиция (habit + goal для UI) — забота клиента или отдельного BFF, если понадобится.

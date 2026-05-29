# user-service

Profile management for AIFA users. Reads a pre-existing identity (created
by `auth-service` on first sign-in) and exposes profile CRUD.

## Endpoints

Mount point at the gateway: `/api/v1/users/*`.

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/users/me` | Bearer | Return the caller's profile |
| PUT | `/users/me` | Bearer | Update name, avatar, locale, timezone |
| GET | `/health` | — | Liveness |

## Events

Subscribes to NATS:
- `user.provisioned` — creates/updates a profile from OAuth-provided fields (email, name, avatar).
- `user.deleted` — deletes the profile row (cascade to this service's data only — each service handles its own cleanup).

Publishing: none.

## Storage

Postgres schema `users`:
```
profiles(
  id         UUID PK (= auth.users.id),
  email      TEXT,
  name       TEXT,
  avatar_url TEXT,
  locale     TEXT NOT NULL DEFAULT 'en',
  timezone   TEXT NOT NULL DEFAULT 'UTC',
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
)
```

JWT revocation is read from Redis (`auth:blacklist:<jti>`), owned by `auth-service`.

## Run

```bash
cp .env.example .env
make tidy       # first time only
make run
make test
```

Via root compose:

```bash
cd ../.. && docker compose up user-service
```

## Notes on race with auth

`auth-service` emits `user.provisioned` over plain NATS (fire-and-forget)
immediately after inserting the identity row and **before** returning
tokens to the client. In practice, by the time the iOS client follows up
with `GET /users/me`, the profile row already exists. If it doesn't (NATS
was briefly down, user-service was restarting), the handler returns 404;
the client should retry.

Durable JetStream delivery is a planned hardening step — tracked for the
final consolidation phase, not this service.

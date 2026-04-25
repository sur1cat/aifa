# auth-service

Identity + JWT issuance for the AIFA backend. Owns nothing except the
provider → user_id binding; profile data lives in `user-service`.

## Endpoints

Mount point at the gateway: `/api/v1/auth/*` (Traefik strips the prefix).

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/auth/google` | — | Exchange Google ID token for AIFA token pair |
| POST | `/auth/apple` | — | Exchange Apple ID token for AIFA token pair |
| POST | `/auth/otp/send` | — | Send OTP code to phone number |
| POST | `/auth/otp/verify` | — | Verify OTP code and authenticate |
| POST | `/auth/refresh` | — | Rotate an access token using a refresh token |
| POST | `/auth/logout` | Bearer | Revoke current access + passed refresh token |
| GET | `/auth/me` | Bearer | Identity record (id, provider, created_at) |
| DELETE | `/auth/account` | Bearer | Delete identity + publish `user.deleted` |
| GET | `/health` | — | Liveness |

Response shape:
```json
{
  "data": {
    "user": { "id": "<uuid>", "auth_provider": "google", "created_at": "..." },
    "tokens": { "access_token": "...", "refresh_token": "...", "expires_at": 1735689600 },
    "is_new_user": true
  }
}
```

## Events

Published to NATS:
- `user.provisioned` — emitted on first sign-in, carries `{user_id, provider, email, name, avatar_url}`. `user-service` subscribes to create a profile.
- `user.deleted` — emitted on account deletion. All user-data services subscribe and purge.

## Storage

- **Postgres** (schema `auth`): one table `users(id, auth_provider, provider_id, created_at)`.
- **Redis** (`auth:blacklist:<jti>`): revoked JWT IDs with TTL = remaining token lifetime.
- **Redis** (`auth:otp:<phone>`): OTP codes with 5-minute TTL, max 5 verification attempts.

## Run

```bash
cp .env.example .env
make tidy       # first time only
make run        # local
make docker     # build image
make test       # unit tests
```

Or via root compose (preferred, wires everything up):

```bash
cd ../.. && docker compose up auth-service
```

## Deploy

Helm chart in [`helm/auth-service`](helm/auth-service). Values required:
- `image.repository`, `image.tag`
- `env.JWT_SECRET`, `env.SERVICE_JWT_SECRET` (from Secret)
- `env.DATABASE_URL`, `env.REDIS_ADDR`, `env.NATS_URL`
- `ingress.host`

```bash
helm upgrade --install auth ./helm/auth-service -f values.prod.yaml
```

## Migrations

SQL files in `migrations/*.up.sql` are applied automatically on server start — no separate step needed.

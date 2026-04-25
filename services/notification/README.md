# notification-service

Manages iOS device tokens and ships APNs push notifications. The HTTP
surface is small (register / unregister); most of the work happens via
NATS event handlers.

## Endpoints

Mount point: `/api/v1/push/*`.

| Method | Path | Description |
|---|---|---|
| POST | `/push/register` | Register/refresh device token (`platform` defaults to `ios`) |
| POST | `/push/unregister` | Drop a device token |
| GET | `/health` | Liveness |

## Events

Subscribes to NATS:
- `user.deleted` — drop all device tokens for the user.
- `reminder.due` — payload `{ user_id, title, body, data? }`. Service looks
  up the user's tokens and sends an APNs push to each.

Publishing: none.

## Storage

Postgres schema `notifications`, single table `device_tokens`. Token column
is unique — re-registering the same token from a different user (rare:
after a phone reset) rebinds it to the new owner via `ON CONFLICT (token)
DO UPDATE`.

## APNs configuration

The APNs client is constructed from `APNS_KEY_PATH`, `APNS_KEY_ID`,
`APNS_TEAM_ID`, `APNS_BUNDLE_ID`, and `APNS_PRODUCTION`. When any of those
is empty, the service starts in "register-only" mode: token endpoints
keep working, but `reminder.due` events are silently dropped. This lets
local dev run without provisioning Apple credentials.

The provider JWT is signed with ES256 and cached for 50 minutes (Apple
allows up to 60). Mount the `.p8` key into the container at the path
referenced by `APNS_KEY_PATH`.

## Run

```bash
cp .env.example .env
make tidy && make run
```

Via root compose: `cd ../.. && docker compose up notification-service`.

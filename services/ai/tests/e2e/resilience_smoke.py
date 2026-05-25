import json
import os
from urllib import request, error


BASE_URL = os.getenv("AIFA_E2E_BASE_URL", "http://127.0.0.1:8080")
TOKEN = os.getenv("AIFA_E2E_TOKEN", "")


def post_json(path: str, body: dict):
    req = request.Request(
        BASE_URL + path,
        data=json.dumps(body).encode(),
        method="POST",
        headers={
            "Authorization": f"Bearer {TOKEN}",
            "Content-Type": "application/json",
            "Accept": "application/json",
        },
    )
    with request.urlopen(req, timeout=60) as resp:
        return resp.status, json.loads(resp.read().decode())


def main() -> None:
    if not TOKEN:
        raise SystemExit("set AIFA_E2E_TOKEN")

    checks = []
    try:
        status, body = post_json("/api/v1/ai/categorize", {"text": ""})
        checks.append({"name": "categorize_empty", "status": status, "body": body})
    except error.HTTPError as exc:
        checks.append({"name": "categorize_empty", "status": exc.code, "body": exc.read().decode()})

    try:
        status, body = post_json("/api/v1/ai/forecast", {"transactions": [], "horizon_days": 3})
        checks.append({"name": "forecast_empty", "status": status, "body": body})
    except error.HTTPError as exc:
        checks.append({"name": "forecast_empty", "status": exc.code, "body": exc.read().decode()})

    print(json.dumps(checks, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()

import json
import os
import time
from concurrent.futures import ThreadPoolExecutor, as_completed
from urllib import request


BASE_URL = os.getenv("AIFA_E2E_BASE_URL", "http://127.0.0.1:8080")
TOKEN = os.getenv("AIFA_E2E_TOKEN", "")
REQUESTS = int(os.getenv("AIFA_LOAD_REQUESTS", "20"))
WORKERS = int(os.getenv("AIFA_LOAD_WORKERS", "5"))


def hit() -> float:
    started = time.perf_counter()
    req = request.Request(
        BASE_URL + "/api/v1/ai/categorize",
        data=json.dumps({"text": "кофе"}).encode(),
        method="POST",
        headers={
            "Authorization": f"Bearer {TOKEN}",
            "Content-Type": "application/json",
            "Accept": "application/json",
        },
    )
    with request.urlopen(req, timeout=60) as resp:
        resp.read()
    return time.perf_counter() - started


def main() -> None:
    if not TOKEN:
        raise SystemExit("set AIFA_E2E_TOKEN")
    latencies = []
    with ThreadPoolExecutor(max_workers=WORKERS) as pool:
        futures = [pool.submit(hit) for _ in range(REQUESTS)]
        for future in as_completed(futures):
            latencies.append(future.result())
    latencies.sort()
    p95 = latencies[int(len(latencies) * 0.95) - 1]
    avg = sum(latencies) / len(latencies)
    print(json.dumps({"requests": REQUESTS, "workers": WORKERS, "avg_s": avg, "p95_s": p95}, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()

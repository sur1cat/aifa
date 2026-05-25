# AI Testing

## Local ai-local unit suite

```bash
cd /Users/kara/Downloads/atoma-main
docker compose --env-file .env exec -T ai-local-service python -m unittest discover -s tests -v
```

## Real OCR golden tests

These use real receipt images in:

- `services/ai-local/tests/fixtures/receipts/ticket.jpg`
- `services/ai-local/tests/fixtures/receipts/unknown-1.png`

Run:

```bash
cd /Users/kara/Downloads/atoma-main
docker compose --env-file .env exec -T ai-local-service env RUN_GOLDEN_OCR=1 python -m unittest tests.test_receipt_golden -v
```

## Real STT golden tests

Place a real voice file at:

- `services/ai-local/tests/fixtures/audio/sample.m4a`

Run:

```bash
cd /Users/kara/Downloads/atoma-main
docker compose --env-file .env exec -T ai-local-service env RUN_GOLDEN_STT=1 python -m unittest tests.test_stt_golden -v
```

## ai-service handler/orchestration tests

```bash
cd /Users/kara/Downloads/atoma-main/services/ai
go test ./...
```

## Gateway end-to-end tests

Requires:

- full stack running
- valid bearer token in `AIFA_E2E_TOKEN`

Run:

```bash
cd /Users/kara/Downloads/atoma-main
PYTHONPATH=services/ai-local python3 -m unittest discover -s services/ai/tests/e2e -p 'test_*.py' -v
```

Optional variables:

- `AIFA_E2E_BASE_URL=http://127.0.0.1:8080`
- `AIFA_E2E_RECEIPT=/abs/path/to/receipt.jpg`
- `AIFA_E2E_AUDIO=/abs/path/to/sample.m4a`

## Performance smoke

```bash
cd /Users/kara/Downloads/atoma-main
python3 services/ai/tests/e2e/perf_smoke.py
```

Environment:

- `AIFA_E2E_TOKEN`
- `AIFA_LOAD_REQUESTS`
- `AIFA_LOAD_WORKERS`

## Resilience smoke

```bash
cd /Users/kara/Downloads/atoma-main
python3 services/ai/tests/e2e/resilience_smoke.py
```

This validates bad-input behavior. It does not deliberately kill containers.

# Local Whisper/OCR Testing

This branch supports local-first voice transcription and receipt OCR through `ai-local-service`.

## What changed

- `POST /api/v1/ai/voice/transcribe`
  - first tries local Whisper in `ai-local-service`
  - falls back to OpenAI only if local processing fails
- `POST /api/v1/ai/receipt/scan`
  - first tries local OCR in `ai-local-service`
  - falls back to OpenAI Vision only if local processing fails

## Prerequisites

- Docker Desktop running
- enough free disk space for the Whisper model
- `.env` based on `.env.example`

Recommended local AI settings:

```env
WHISPER_MODEL_SIZE=small
WHISPER_DEVICE=cpu
WHISPER_COMPUTE_TYPE=int8
OCR_LANGS=eng+rus+kaz
```

## Build and run

```bash
cd /Users/kara/Downloads/atoma-main
docker compose --env-file .env.example up --build -d ai-local-service ai-service
```

Check health:

```bash
curl http://localhost:8010/health
curl http://localhost:8007/health
```

## Test local Whisper

Send an audio file:

```bash
curl -X POST http://localhost:8007/ai/voice/transcribe \
  -F "audio=@/absolute/path/to/sample.m4a" \
  -F "language=ru"
```

Expected response shape:

```json
{
  "data": {
    "transcript": "такси 2500",
    "amount": 2500,
    "currency": "KZT",
    "description": "Такси",
    "category": "transport",
    "label_ru": "Транспорт",
    "label_kz": "Көлік",
    "confidence": 0.8
  }
}
```

## Test local OCR

Send a receipt image:

```bash
curl -X POST http://localhost:8007/ai/receipt/scan \
  -F "image=@/absolute/path/to/receipt.jpg"
```

Expected response shape:

```json
{
  "data": {
    "amount": 1850,
    "currency": "KZT",
    "date": "2026-05-22",
    "merchant": "MAGNUM",
    "category": "food",
    "label_ru": "Продукты",
    "label_kz": "Азық-түлік",
    "items": ["Молоко 650", "Хлеб 220"],
    "confidence": 0.74,
    "raw_total": "1850",
    "raw_text": "..."
  }
}
```

## Demo notes

- Voice and receipt endpoints are now local-first.
- OpenAI is no longer the primary path for speech/OCR.
- For a defense demo, prepare:
  - one short audio with an expense command
  - one clear receipt photo
  - one blurred/noisy receipt to show graceful degradation

## Unit tests

Fast offline regression tests for the OCR parser:

```bash
cd /Users/kara/Downloads/atoma-main/services/ai-local
PYTHONPATH=. python3 -m unittest tests.test_receipt_ocr
```

Current coverage:

- date extraction from split `date + time` lines
- total extraction from `Итого / Барлығы / Total`
- merchant cleanup from receipt header text
- end-to-end normalization of mocked OCR output

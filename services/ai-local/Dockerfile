FROM python:3.11-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    tesseract-ocr \
    tesseract-ocr-rus \
    tesseract-ocr-kaz \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

# Обучаем модель при сборке образа
RUN python training/train.py

# Предзагружаем локальную Whisper-модель до live-демо.
RUN python - <<'PY'
from app.stt import _load_model
_load_model()
print("whisper model preloaded")
PY

EXPOSE 8000

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]

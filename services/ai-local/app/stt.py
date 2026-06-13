import os
import tempfile
from dataclasses import dataclass
from pathlib import Path
from threading import Lock
from typing import Optional


_model = None
_model_lock = Lock()


@dataclass
class STTResult:
    transcript: str
    language: str
    language_probability: float


def _load_model():
    global _model
    if _model is not None:
        return _model

    with _model_lock:
        if _model is not None:
            return _model

        try:
            from faster_whisper import WhisperModel
        except ImportError as exc:
            raise RuntimeError("faster-whisper is not installed") from exc

        model_size = os.getenv("WHISPER_MODEL_SIZE") or "small"
        device = os.getenv("WHISPER_DEVICE") or "cpu"
        compute_type = os.getenv("WHISPER_COMPUTE_TYPE") or "int8"
        _model = WhisperModel(model_size, device=device, compute_type=compute_type)
        return _model




import re

def _fix_financial_transcript(text):
    # тысячи грамм -> тысячи
    text = re.sub(r'(тысяч[аи]?) грамм\w*', lambda m: m.group(1), text, flags=re.IGNORECASE)
    # тенен/теней -> тысяч
    text = re.sub(r'\bтенен\b|\bтеней\b|\bтенений\b', 'тысяч', text, flags=re.IGNORECASE)
    # 5тенге -> 5 тенге
    text = re.sub(r'(\d)(тенг[еи])', lambda m: m.group(1) + ' тенге', text, flags=re.IGNORECASE)
    # трятенге/трятенги -> 3 тенге
    merges = [('тря', '3'), ('три', '3'), ('пять', '5'), ('семь', '7'),
              ('два', '2'), ('десять', '10'), ('сто', '100'), ('один', '1')]
    for word, num in merges:
        text = re.sub(r'\b' + word + r'тенг\w*', num + ' тенге', text, flags=re.IGNORECASE)
    # тенги -> тенге
    text = re.sub(r'\bтенги\b', 'тенге', text, flags=re.IGNORECASE)
    return text


def transcribe_audio(audio_bytes: bytes, filename: str, language: Optional[str] = None) -> STTResult:
    if not audio_bytes:
        raise ValueError("audio file is empty")

    suffix = Path(filename or "audio.wav").suffix or ".wav"
    tmp_path = None
    try:
        with tempfile.NamedTemporaryFile(delete=False, suffix=suffix) as tmp:
            tmp.write(audio_bytes)
            tmp.flush()
            tmp_path = tmp.name

        model = _load_model()
        initial_prompt = (
            "Финансы, расходы, доходы, тенге, тысяч, купил, потратил, заработал, переводы, оплата. "
            "Қаржы, шығын, табыс, теңге, мың, сатып алдым, жұмсадым, аударма. "
            "Finance, expenses, income, tenge, spent, earned, transfer."
        )
        _ALLOWED_LANGS = {"ru", "kk", "en"}
        effective_lang = language if language in _ALLOWED_LANGS else None

        segments, info = model.transcribe(
            tmp_path,
            language=effective_lang,
            vad_filter=True,
            beam_size=5,
            initial_prompt=initial_prompt,
            condition_on_previous_text=False,
        )
        raw = " ".join(segment.text.strip() for segment in segments if segment.text.strip()).strip()

        detected = getattr(info, "language", "") or ""
        # If Whisper auto-detected a language outside our supported set, re-transcribe as Russian
        if effective_lang is None and detected not in _ALLOWED_LANGS:
            segments2, info = model.transcribe(
                tmp_path,
                language="ru",
                vad_filter=True,
                beam_size=5,
                initial_prompt=initial_prompt,
                condition_on_previous_text=False,
            )
            raw = " ".join(s.text.strip() for s in segments2 if s.text.strip()).strip()
            detected = "ru"

        transcript = _fix_financial_transcript(raw)
        return STTResult(
            transcript=transcript,
            language=detected,
            language_probability=float(getattr(info, "language_probability", 0.0) or 0.0),
        )
    finally:
        if tmp_path:
            try:
                os.remove(tmp_path)
            except OSError:
                pass

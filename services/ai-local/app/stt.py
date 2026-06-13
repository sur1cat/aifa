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
        segments, info = model.transcribe(
            tmp_path,
            language=language or None,
            vad_filter=True,
            beam_size=5,
        )
        transcript = " ".join(segment.text.strip() for segment in segments if segment.text.strip()).strip()
        return STTResult(
            transcript=transcript,
            language=getattr(info, "language", "") or "",
            language_probability=float(getattr(info, "language_probability", 0.0) or 0.0),
        )
    finally:
        if tmp_path:
            try:
                os.remove(tmp_path)
            except OSError:
                pass

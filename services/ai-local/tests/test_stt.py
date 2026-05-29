import os
import sys
import unittest
from types import SimpleNamespace
from unittest.mock import patch

sys.modules.setdefault(
    "faster_whisper",
    SimpleNamespace(WhisperModel=object),
)

from app import stt


class _FakeSegment:
    def __init__(self, text: str) -> None:
        self.text = text


class _FakeInfo:
    def __init__(self, language: str = "ru", language_probability: float = 0.97) -> None:
        self.language = language
        self.language_probability = language_probability


class _FakeModel:
    def __init__(self) -> None:
        self.calls = []

    def transcribe(self, path: str, language=None, vad_filter=True, beam_size=5):
        self.calls.append(
            {
                "path": path,
                "language": language,
                "vad_filter": vad_filter,
                "beam_size": beam_size,
            }
        )
        return [
            _FakeSegment(" hello "),
            _FakeSegment("world"),
            _FakeSegment(""),
        ], _FakeInfo()


class STTUnitTests(unittest.TestCase):
    def setUp(self) -> None:
        stt._model = None

    def tearDown(self) -> None:
        stt._model = None
        os.environ.pop("WHISPER_MODEL_SIZE", None)
        os.environ.pop("WHISPER_DEVICE", None)
        os.environ.pop("WHISPER_COMPUTE_TYPE", None)

    def test_transcribe_audio_rejects_empty_file(self) -> None:
        with self.assertRaisesRegex(ValueError, "audio file is empty"):
            stt.transcribe_audio(b"", "voice.m4a")

    @patch("app.stt._load_model")
    def test_transcribe_audio_returns_joined_transcript_and_metadata(self, mock_load_model) -> None:
        fake_model = _FakeModel()
        mock_load_model.return_value = fake_model

        result = stt.transcribe_audio(b"audio-bytes", "voice.m4a", language="ru")

        self.assertEqual(result.transcript, "hello world")
        self.assertEqual(result.language, "ru")
        self.assertEqual(result.language_probability, 0.97)
        self.assertEqual(len(fake_model.calls), 1)
        self.assertEqual(fake_model.calls[0]["language"], "ru")
        self.assertTrue(fake_model.calls[0]["path"].endswith(".m4a"))

    def test_load_model_uses_env_configuration_and_caches_instance(self) -> None:
        sentinel = object()
        os.environ["WHISPER_MODEL_SIZE"] = "medium"
        os.environ["WHISPER_DEVICE"] = "cpu"
        os.environ["WHISPER_COMPUTE_TYPE"] = "int8"

        with patch.object(sys.modules["faster_whisper"], "WhisperModel", return_value=sentinel) as mock_whisper_model:
            first = stt._load_model()
            second = stt._load_model()

        self.assertIs(first, sentinel)
        self.assertIs(second, sentinel)
        mock_whisper_model.assert_called_once_with("medium", device="cpu", compute_type="int8")


if __name__ == "__main__":
    unittest.main()

import os
from pathlib import Path
import unittest

from app.stt import transcribe_audio


FIXTURES_DIR = Path(__file__).parent / "fixtures" / "audio"


@unittest.skipUnless(os.getenv("RUN_GOLDEN_STT") == "1", "set RUN_GOLDEN_STT=1 to run real STT golden tests")
class STTGoldenTests(unittest.TestCase):
    def test_real_voice_fixture(self) -> None:
        audio = FIXTURES_DIR / "sample.m4a"
        if not audio.exists():
            self.skipTest("missing audio fixture tests/fixtures/audio/sample.m4a")

        result = transcribe_audio(audio.read_bytes(), audio.name, language="ru")

        self.assertTrue(result.transcript.strip())
        self.assertEqual(result.language, "ru")
        self.assertGreaterEqual(result.language_probability, 0.0)


if __name__ == "__main__":
    unittest.main()

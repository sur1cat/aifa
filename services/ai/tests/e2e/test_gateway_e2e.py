import json
import mimetypes
import os
from pathlib import Path
import unittest
from urllib import request, error


BASE_URL = os.getenv("AIFA_E2E_BASE_URL", "http://127.0.0.1:8080")
TOKEN = os.getenv("AIFA_E2E_TOKEN", "")
RECEIPT_FIXTURE = Path(os.getenv(
    "AIFA_E2E_RECEIPT",
    str(Path(__file__).resolve().parents[3] / "ai-local" / "tests" / "fixtures" / "receipts" / "ticket.jpg"),
))
AUDIO_FIXTURE = Path(os.getenv(
    "AIFA_E2E_AUDIO",
    str(Path(__file__).resolve().parents[3] / "ai-local" / "tests" / "fixtures" / "audio" / "sample.m4a"),
))


def _multipart_body(field_name: str, file_path: Path, extra_fields: dict[str, str] | None = None):
    boundary = "----AIFAIntegrationBoundary"
    parts: list[bytes] = []
    extra_fields = extra_fields or {}
    for key, value in extra_fields.items():
        parts.append(
            (
                f"--{boundary}\r\n"
                f'Content-Disposition: form-data; name="{key}"\r\n\r\n'
                f"{value}\r\n"
            ).encode()
        )
    mime = mimetypes.guess_type(file_path.name)[0] or "application/octet-stream"
    parts.append(
        (
            f"--{boundary}\r\n"
            f'Content-Disposition: form-data; name="{field_name}"; filename="{file_path.name}"\r\n'
            f"Content-Type: {mime}\r\n\r\n"
        ).encode()
    )
    parts.append(file_path.read_bytes())
    parts.append(f"\r\n--{boundary}--\r\n".encode())
    return boundary, b"".join(parts)


@unittest.skipUnless(TOKEN, "set AIFA_E2E_TOKEN to run gateway e2e tests")
class GatewayE2ETests(unittest.TestCase):
    def _post_multipart(self, path: str, field_name: str, file_path: Path, extra_fields: dict[str, str] | None = None):
        boundary, body = _multipart_body(field_name, file_path, extra_fields)
        req = request.Request(
            BASE_URL + path,
            data=body,
            method="POST",
            headers={
                "Authorization": f"Bearer {TOKEN}",
                "Content-Type": f"multipart/form-data; boundary={boundary}",
                "Accept": "application/json",
            },
        )
        with request.urlopen(req, timeout=120) as resp:
            payload = json.loads(resp.read().decode())
        return payload["data"]

    def test_receipt_scan_through_gateway(self):
        self.assertTrue(RECEIPT_FIXTURE.exists(), f"missing fixture: {RECEIPT_FIXTURE}")
        data = self._post_multipart("/api/v1/ai/receipt/scan", "image", RECEIPT_FIXTURE)
        self.assertEqual(data["currency"], "RUB")
        self.assertEqual(round(float(data["amount"]), 2), 785.0)
        self.assertIn("merchant", data)

    def test_voice_transcribe_through_gateway(self):
        if not AUDIO_FIXTURE.exists():
            self.skipTest(f"missing audio fixture: {AUDIO_FIXTURE}")
        data = self._post_multipart("/api/v1/ai/voice/transcribe", "audio", AUDIO_FIXTURE, {"language": "ru"})
        self.assertIn("transcript", data)
        self.assertTrue(data["transcript"].strip())


if __name__ == "__main__":
    unittest.main()

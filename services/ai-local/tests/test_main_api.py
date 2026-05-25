import sys
import unittest
from types import SimpleNamespace
from unittest.mock import patch

from fastapi.testclient import TestClient


sys.modules.setdefault(
    "app.model",
    SimpleNamespace(
        get_classifier=lambda: SimpleNamespace(
            meta={"categories": ["shopping"], "labels_ru": {}, "labels_kz": {}},
            predict=lambda _text: SimpleNamespace(
                category="shopping",
                label_ru="Покупки",
                label_kz="Сатып алулар",
                confidence=0.9,
                confident=True,
            ),
        ),
        CategoryResult=object,
        CATEGORY_THRESHOLDS={},
        CONFIDENCE_THRESHOLD=0.5,
    ),
)

from app.main import app
from app.receipt_ocr import ReceiptScanResult
from app.stt import STTResult


class MainAPIUnitTests(unittest.TestCase):
    def setUp(self) -> None:
        self.client = TestClient(app)

    def test_health_endpoint(self) -> None:
        response = self.client.get("/health")

        self.assertEqual(response.status_code, 200)
        self.assertEqual(response.json()["service"], "ai-local")

    @patch("app.main.get_classifier")
    def test_categorize_endpoint(self, mock_get_classifier) -> None:
        mock_get_classifier.return_value = SimpleNamespace(
            predict=lambda _text: SimpleNamespace(
                category="shopping",
                label_ru="Покупки",
                label_kz="Сатып алулар",
                confidence=0.87,
                confident=True,
            ),
            meta={},
        )

        response = self.client.post("/categorize", json={"text": "купил продукты"})

        self.assertEqual(response.status_code, 200)
        self.assertEqual(response.json()["category"], "shopping")

    def test_parse_message_endpoint(self) -> None:
        response = self.client.post("/parse-message", json={"message": "купил кофе за 1500 тг"})

        self.assertEqual(response.status_code, 200)
        self.assertEqual(response.json()["intent"], "create_transaction")

    @patch("app.main.scan_receipt")
    def test_receipt_scan_endpoint(self, mock_scan_receipt) -> None:
        mock_scan_receipt.return_value = ReceiptScanResult(
            amount=14.25,
            currency="USD",
            date="2026-05-24T09:41:05",
            merchant="ALPHA STORE",
            category="shopping",
            label_ru="Покупки",
            label_kz="Сатып алулар",
            items=["Fresh juice", "Snack mix"],
            confidence=0.9,
            raw_total="14.25",
            raw_text="raw",
        )

        response = self.client.post(
            "/receipt/scan",
            files={"image": ("receipt.jpg", b"fake-image", "image/jpeg")},
        )

        self.assertEqual(response.status_code, 200)
        self.assertEqual(response.json()["merchant"], "ALPHA STORE")

    @patch("app.main.transcribe_audio")
    def test_voice_transcribe_endpoint(self, mock_transcribe_audio) -> None:
        mock_transcribe_audio.return_value = STTResult(
            transcript="купил кофе за 1500 тг",
            language="ru",
            language_probability=0.98,
        )

        response = self.client.post(
            "/voice/transcribe",
            files={"audio": ("voice.m4a", b"fake-audio", "audio/mp4")},
        )

        self.assertEqual(response.status_code, 200)
        self.assertEqual(response.json()["transcript"], "купил кофе за 1500 тг")


if __name__ == "__main__":
    unittest.main()

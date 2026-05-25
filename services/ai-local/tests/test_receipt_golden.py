import os
from pathlib import Path
import unittest

from app.receipt_ocr import scan_receipt


FIXTURES_DIR = Path(__file__).parent / "fixtures" / "receipts"


@unittest.skipUnless(os.getenv("RUN_GOLDEN_OCR") == "1", "set RUN_GOLDEN_OCR=1 to run real OCR golden tests")
class ReceiptOCRGoldenTests(unittest.TestCase):
    def test_ticket_receipt_real_fixture(self) -> None:
        image = FIXTURES_DIR / "ticket.jpg"
        self.assertTrue(image.exists(), f"missing fixture: {image}")

        result = scan_receipt(image.read_bytes(), languages="eng+rus+kaz")

        self.assertEqual(result.currency, "RUB")
        self.assertEqual(result.amount, 785.0)
        self.assertEqual(result.date, "2025-07-29T16:36:00")
        self.assertIn("АВТОКАФЕ", result.merchant.upper())
        self.assertTrue(any("морс" in item.lower() for item in result.items))

    def test_unknown_fixture_real_receipt(self) -> None:
        image = FIXTURES_DIR / "unknown-1.png"
        if not image.exists():
            self.skipTest("unknown-1.png fixture not provided")

        result = scan_receipt(image.read_bytes(), languages="eng+rus+kaz")

        self.assertIsNotNone(result.amount)
        self.assertEqual(result.amount, 200.0)
        self.assertIsNotNone(result.date)
        self.assertTrue(result.date.startswith("2026-03-05"))
        self.assertTrue(result.merchant.strip())


if __name__ == "__main__":
    unittest.main()

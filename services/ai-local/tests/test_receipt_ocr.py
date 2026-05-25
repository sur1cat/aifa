import os
import sys
import unittest
from types import SimpleNamespace
from unittest.mock import patch

sys.modules.setdefault(
    "app.model",
    SimpleNamespace(get_classifier=lambda: None),
)

from app import receipt_ocr


GENERIC_RECEIPT_TEXT = """\
"ALPHA STORE" limited liability partnership
Main street 10

Sales
Fresh juice
2 x 5.50 = 11.00
Snack mix
1 x 3.25 = 3.25

Total 14.25 USD
Cash 14.25 USD
Document 42 24.05.2026,
09:41:05
"""


class ReceiptOCRUnitTests(unittest.TestCase):
    def tearDown(self) -> None:
        os.environ.pop("OCR_MERCHANT_ALIASES", None)
        os.environ.pop("OCR_MERCHANT_HINTS", None)
        receipt_ocr._merchant_aliases.cache_clear()
        receipt_ocr._merchant_hints.cache_clear()

    def test_parse_money_handles_valid_and_invalid_inputs(self) -> None:
        self.assertEqual(receipt_ocr._parse_money("1 234,56"), 1234.56)
        self.assertEqual(receipt_ocr._parse_money("77.5"), 77.5)
        self.assertIsNone(receipt_ocr._parse_money("0"))
        self.assertIsNone(receipt_ocr._parse_money("not-a-number"))

    def test_detect_currency_supports_known_markers_and_default(self) -> None:
        self.assertEqual(receipt_ocr._detect_currency("Total 50 ₸"), "KZT")
        self.assertEqual(receipt_ocr._detect_currency("Итого 100 руб"), "RUB")
        self.assertEqual(receipt_ocr._detect_currency("Amount due $20"), "USD")
        self.assertEqual(receipt_ocr._detect_currency("Total 30 EUR"), "EUR")
        self.assertEqual(receipt_ocr._detect_currency("No markers here"), "KZT")

    def test_extract_date_supports_datetime_date_only_and_missing(self) -> None:
        self.assertEqual(
            receipt_ocr._extract_date(["Document 42 24.05.2026,", "09:41:05"]),
            "2026-05-24T09:41:05",
        )
        self.assertEqual(
            receipt_ocr._extract_date(["Date 2026/05/24"]),
            "2026-05-24",
        )
        self.assertIsNone(receipt_ocr._extract_date(["No date here"]))

    def test_extract_total_prefers_total_lines_then_falls_back(self) -> None:
        amount, raw_total = receipt_ocr._extract_total(
            [
                "Fresh juice 11.00",
                "Total 14.25 USD",
                "09:41:05",
            ]
        )
        self.assertEqual(amount, 14.25)
        self.assertEqual(raw_total, "14.25")

        fallback_amount, fallback_raw = receipt_ocr._extract_total(
            [
                "Fresh juice 11.00",
                "Snack mix 3.25",
            ]
        )
        self.assertEqual(fallback_amount, 11.0)
        self.assertEqual(fallback_raw, "11.00")

    def test_extract_total_from_zone_uses_total_hints_only(self) -> None:
        amount, raw_total = receipt_ocr._extract_total_from_zone(
            [
                "Subtotal 13.00",
                "Total 14.25 USD",
            ]
        )
        self.assertEqual(amount, 14.25)
        self.assertEqual(raw_total, "14.25")

    def test_extract_merchant_handles_quotes_suffixes_and_aggregation(self) -> None:
        lines = receipt_ocr._normalize_lines(GENERIC_RECEIPT_TEXT)

        merchant = receipt_ocr._extract_merchant(
            lines,
            [
                '"ALPHA STORE" limited liability partnership',
                '"ALPHA STORE"',
                '"ALPHA STORE"',
            ],
        )

        self.assertEqual(merchant, "ALPHA STORE")

    def test_extract_merchant_returns_empty_for_noise_only(self) -> None:
        merchant = receipt_ocr._extract_merchant(
            [],
            [
                "Cash receipt",
                "Main street 10",
                "Document 42",
            ],
        )

        self.assertEqual(merchant, "")

    def test_extract_merchant_applies_alias_and_hint_corrections(self) -> None:
        os.environ["OCR_MERCHANT_ALIASES"] = "ALPNA STORE=ALPHA STORE"
        os.environ["OCR_MERCHANT_HINTS"] = "ALPHA STORE,BETA MARKET"
        receipt_ocr._merchant_aliases.cache_clear()
        receipt_ocr._merchant_hints.cache_clear()

        alias_merchant = receipt_ocr._extract_merchant([], ['"ALPNA STORE" llp'])
        hinted_merchant = receipt_ocr._correct_merchant_with_hints("BETA MARK3T")

        self.assertEqual(alias_merchant, "ALPHA STORE")
        self.assertEqual(hinted_merchant, "BETA MARKET")

    def test_extract_items_supports_explicit_and_implicit_layouts(self) -> None:
        explicit_items = receipt_ocr._extract_items(
            [
                "Sales",
                "Fresh juice",
                "2 x 5.50 = 11.00",
                "Snack mix",
                "1 x 3.25 = 3.25",
                "Total 14.25 USD",
            ],
            merchant="ALPHA STORE",
            raw_total="14.25",
            receipt_date="2026-05-24T09:41:05",
        )
        implicit_items = receipt_ocr._extract_items(
            [
                "Bakery box",
                "1 x 7.00 = 7.00",
                "Tea",
                "1 x 2.50 = 2.50",
                "Total 9.50",
            ],
            merchant="",
            raw_total="9.50",
            receipt_date=None,
        )

        self.assertEqual(explicit_items, ["Fresh juice", "Snack mix"])
        self.assertEqual(implicit_items, ["Bakery box", "Tea"])

    @patch("app.receipt_ocr.get_classifier")
    @patch("app.receipt_ocr._extract_total_candidates")
    @patch("app.receipt_ocr._extract_footer_candidates")
    @patch("app.receipt_ocr._extract_header_candidates")
    @patch("app.receipt_ocr._extract_text")
    def test_scan_receipt_returns_normalized_result(
        self,
        mock_extract_text,
        mock_header_candidates,
        mock_footer_candidates,
        mock_total_candidates,
        mock_get_classifier,
    ) -> None:
        mock_extract_text.return_value = GENERIC_RECEIPT_TEXT
        mock_header_candidates.return_value = [
            '"ALPHA STORE" limited liability partnership',
            '"ALPHA STORE"',
        ]
        mock_footer_candidates.return_value = [
            "Document 42 24.05.2026,",
            "09:41:05",
        ]
        mock_total_candidates.return_value = [
            "Total 14.25 USD",
        ]
        mock_get_classifier.return_value = SimpleNamespace(
            predict=lambda _text: SimpleNamespace(
                category="shopping",
                label_ru="Покупки",
                label_kz="Сатып алулар",
                confidence=0.91,
            )
        )

        result = receipt_ocr.scan_receipt(b"fake-image", languages="eng+rus+kaz")

        self.assertEqual(result.amount, 14.25)
        self.assertEqual(result.currency, "USD")
        self.assertEqual(result.date, "2026-05-24T09:41:05")
        self.assertEqual(result.merchant, "ALPHA STORE")
        self.assertEqual(result.items, ["Fresh juice", "Snack mix"])
        self.assertEqual(result.raw_total, "14.25")
        self.assertEqual(result.category, "shopping")
        self.assertGreaterEqual(result.confidence, 0.8)


if __name__ == "__main__":
    unittest.main()

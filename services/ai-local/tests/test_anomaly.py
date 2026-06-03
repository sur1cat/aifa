import unittest
from unittest.mock import patch

import pandas as pd

from app import anomaly


class AnomalyUnitTests(unittest.TestCase):
    def test_detect_anomalies_returns_empty_for_no_valid_transactions(self) -> None:
        result = anomaly.detect_anomalies([])
        self.assertEqual(result.total_anomalies, 0)
        self.assertEqual(result.anomalies, [])

    def test_modified_zscore_handles_constant_series(self) -> None:
        values = anomaly._modified_zscore(pd.Series([10.0, 10.0, 10.0]).values)
        self.assertEqual(list(values), [0.0, 0.0, 0.0])

    def test_detect_anomalies_finds_zscore_outlier(self) -> None:
        txs = [
            {"date": "2026-05-01", "amount": 1000.0, "category": "food"},
            {"date": "2026-05-02", "amount": 1100.0, "category": "food"},
            {"date": "2026-05-03", "amount": 950.0, "category": "food"},
            {"date": "2026-05-04", "amount": 5000.0, "category": "food"},
        ]

        result = anomaly.detect_anomalies(txs, sensitivity="high")

        self.assertGreaterEqual(result.total_anomalies, 1)
        self.assertEqual(result.anomalies[0].category, "food")

    def test_detect_anomalies_flags_pair_spike_with_short_history(self) -> None:
        txs = [
            {"date": "2026-05-01", "amount": 900.0, "category": "shopping"},
            {"date": "2026-05-02", "amount": 60000.0, "category": "shopping"},
        ]

        result = anomaly.detect_anomalies(txs, sensitivity="medium")

        self.assertEqual(result.total_anomalies, 1)
        self.assertEqual(result.anomalies[0].category, "shopping")
        self.assertEqual(result.anomalies[0].amount, 60000.0)

    def test_detect_anomalies_flags_dominant_single_spend_with_short_history(self) -> None:
        txs = [
            {"date": "2026-05-01", "amount": 60000.0, "category": "shopping"},
            {"date": "2026-05-01", "amount": 500.0, "category": "food"},
            {"date": "2026-05-02", "amount": 900.0, "category": "gift"},
        ]

        result = anomaly.detect_anomalies(txs, sensitivity="medium")

        self.assertEqual(result.total_anomalies, 1)
        self.assertEqual(result.anomalies[0].category, "shopping")
        self.assertEqual(result.anomalies[0].source, "single_spend")
        self.assertFalse(result.enough_history)
        self.assertEqual(result.observation_days, 2)

    @patch("app.anomaly._isolation_anomalies")
    def test_detect_anomalies_includes_isolation_hits_for_long_series(self, mock_iso) -> None:
        mock_iso.return_value = [
            anomaly.AnomalyPoint(
                date="2026-05-11",
                category="total",
                label_ru="Суммарные расходы за день",
                label_kz="Күндік жалпы шығындар",
                amount=9000.0,
                mean=2000.0,
                std=500.0,
                z_score=4.0,
                severity="high",
                source="isolation_forest",
                expected_lower=1200.0,
                expected_upper=2800.0,
            )
        ]
        txs = [
            {"date": f"2026-05-{idx:02d}", "amount": 1000.0 + idx, "category": "food"}
            for idx in range(1, 12)
        ]

        result = anomaly.detect_anomalies(txs)

        self.assertEqual(result.method, "zscore+isolation_forest")
        self.assertTrue(any(item.source == "isolation_forest" for item in result.anomalies))


if __name__ == "__main__":
    unittest.main()

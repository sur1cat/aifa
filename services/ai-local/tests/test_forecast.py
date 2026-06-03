from datetime import date
import unittest
from unittest.mock import patch

import numpy as np

from app import forecast


class ForecastUnitTests(unittest.TestCase):
    def test_forecast_category_returns_none_without_matching_transactions(self) -> None:
        result = forecast.forecast_category([], "food")
        self.assertIsNone(result)

    def test_forecast_category_uses_mean_for_short_series(self) -> None:
        txs = [
            {"date": "2026-05-01", "amount": 1000.0, "category": "food"},
            {"date": "2026-05-02", "amount": 2000.0, "category": "food"},
        ]

        result = forecast.forecast_category(txs, "food", horizon_days=2, ref_date=date(2026, 5, 2))

        self.assertEqual(result.method, "mean")
        self.assertEqual(len(result.daily), 2)
        self.assertEqual(result.daily[0].date, "2026-05-03")

    def test_forecast_category_uses_moving_average_for_medium_series(self) -> None:
        txs = [
            {"date": f"2026-05-{idx:02d}", "amount": 1000.0 + idx, "category": "food"}
            for idx in range(1, 6)
        ]

        result = forecast.forecast_category(txs, "food", horizon_days=3, ref_date=date(2026, 5, 5))

        self.assertEqual(result.method, "moving_avg")
        self.assertEqual(len(result.daily), 3)

    @patch("app.forecast._holt_winters")
    def test_forecast_category_uses_holt_winters_for_long_series(self, mock_hw) -> None:
        mock_hw.return_value = (
            np.array([10.0, 11.0]),
            np.array([8.0, 9.0]),
            np.array([12.0, 13.0]),
        )
        txs = [
            {"date": f"2026-05-{idx:02d}", "amount": 1000.0, "category": "food"}
            for idx in range(1, 15)
        ]

        result = forecast.forecast_category(txs, "food", horizon_days=2, ref_date=date(2026, 5, 14))

        self.assertEqual(result.method, "holt_winters")
        self.assertEqual(result.total_predicted, 21.0)

    def test_forecast_all_returns_sorted_categories(self) -> None:
        txs = [
            {"date": "2026-05-01", "amount": 1000.0, "category": "food"},
            {"date": "2026-05-02", "amount": 1200.0, "category": "food"},
            {"date": "2026-05-03", "amount": 100.0, "category": "transport"},
            {"date": "2026-05-04", "amount": 150.0, "category": "transport"},
        ]

        results = forecast.forecast_all(txs, horizon_days=1, ref_date=date(2026, 5, 4))

        self.assertGreaterEqual(results[0].total_predicted, results[-1].total_predicted)

    def test_forecast_all_excludes_non_spending_like_categories(self) -> None:
        txs = [
            {"date": "2026-05-01", "amount": 1000.0, "category": "food"},
            {"date": "2026-05-02", "amount": 1200.0, "category": "food"},
            {"date": "2026-05-03", "amount": 50000.0, "category": "salary"},
            {"date": "2026-05-04", "amount": 7000.0, "category": "savings"},
            {"date": "2026-05-05", "amount": 900.0, "category": "gift"},
            {"date": "2026-05-06", "amount": 3000.0, "category": "transfer"},
        ]

        results = forecast.forecast_all(txs, horizon_days=1, ref_date=date(2026, 5, 6))

        self.assertEqual([r.category for r in results], ["food"])


if __name__ == "__main__":
    unittest.main()

import unittest
from unittest.mock import patch

from app import insights


class InsightsUnitTests(unittest.TestCase):
    def test_spending_summary_handles_empty_input(self) -> None:
        result = insights.spending_summary([])

        self.assertEqual(result.total_income, 0)
        self.assertEqual(result.total_expense, 0)
        self.assertEqual(result.expense_trend, "unknown")

    def test_spending_summary_computes_totals_and_trend(self) -> None:
        txs = [
            {"date": "2026-05-01", "amount": 100000.0, "type": "income", "category": "salary"},
            {"date": "2026-05-01", "amount": 1000.0, "type": "expense", "category": "food"},
            {"date": "2026-05-02", "amount": 2000.0, "type": "expense", "category": "food"},
            {"date": "2026-05-03", "amount": 6000.0, "type": "expense", "category": "shopping"},
            {"date": "2026-05-04", "amount": 7000.0, "type": "expense", "category": "shopping"},
        ]

        result = insights.spending_summary(txs)

        self.assertEqual(result.total_income, 100000.0)
        self.assertEqual(result.total_expense, 16000.0)
        self.assertEqual(result.top_categories[0].category, "shopping")
        self.assertIn(result.expense_trend, {"up", "down", "stable", "unknown"})

    @patch("app.insights.date")
    def test_budget_suggestions_returns_prioritized_limits(self, mock_date) -> None:
        from datetime import date as real_date

        mock_date.today.return_value = real_date(2026, 5, 31)
        mock_date.side_effect = lambda *args, **kwargs: real_date(*args, **kwargs)

        txs = [
            {"date": "2026-03-05", "amount": 10000.0, "type": "expense", "category": "food"},
            {"date": "2026-03-12", "amount": 12000.0, "type": "expense", "category": "food"},
            {"date": "2026-04-07", "amount": 15000.0, "type": "expense", "category": "food"},
            {"date": "2026-04-19", "amount": 16000.0, "type": "expense", "category": "food"},
            {"date": "2026-05-01", "amount": 7000.0, "type": "expense", "category": "transport"},
            {"date": "2026-05-18", "amount": 8000.0, "type": "expense", "category": "transport"},
        ]

        result = insights.budget_suggestions(txs, lookback_days=120)

        self.assertTrue(result)
        self.assertTrue(all(item.suggested_limit >= 0 for item in result))


if __name__ == "__main__":
    unittest.main()

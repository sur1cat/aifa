import json
import os
from pathlib import Path
import unittest
from urllib import request


BASE_URL = os.getenv("AIFA_E2E_BASE_URL", "http://127.0.0.1:8080")
TOKEN = os.getenv("AIFA_E2E_TOKEN", "")


@unittest.skipUnless(TOKEN, "set AIFA_E2E_TOKEN to run contract tests")
class ContractShapeTests(unittest.TestCase):
    def _post_json(self, path: str, body: dict):
        req = request.Request(
            BASE_URL + path,
            data=json.dumps(body).encode(),
            method="POST",
            headers={
                "Authorization": f"Bearer {TOKEN}",
                "Content-Type": "application/json",
                "Accept": "application/json",
            },
        )
        with request.urlopen(req, timeout=60) as resp:
            payload = json.loads(resp.read().decode())
        return payload["data"]

    def test_parse_message_shape_matches_frontend_expectations(self):
        data = self._post_json("/api/v1/ai/parse-message", {"message": "купил кофе за 1500 тг"})
        for key in ("intent", "response"):
            self.assertIn(key, data)

    def test_categorize_shape_matches_frontend_expectations(self):
        data = self._post_json("/api/v1/ai/categorize", {"text": "кофе"})
        for key in ("category", "label_ru", "label_kz", "confidence"):
            self.assertIn(key, data)

    def test_forecast_shape_matches_frontend_expectations(self):
        data = self._post_json(
            "/api/v1/ai/forecast",
            {
                "transactions": [
                    {"date": "2026-05-01", "amount": 1000, "category": "food"},
                    {"date": "2026-05-02", "amount": 2000, "category": "food"},
                    {"date": "2026-05-03", "amount": 3000, "category": "food"},
                ],
                "horizon_days": 3,
            },
        )
        self.assertIn("forecasts", data)
        self.assertIn("horizon_days", data)


if __name__ == "__main__":
    unittest.main()

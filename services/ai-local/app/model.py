"""
Загрузка и использование обученной модели классификации расходов.
"""

import json
import os
from dataclasses import dataclass
from typing import Optional

import joblib
import numpy as np

from app.encoder import MultilingualEncoder  # noqa: F401 — needed for joblib deserialization

_BASE = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
_MODEL_PATH = os.path.join(_BASE, "models", "expense_classifier.joblib")
_META_PATH  = os.path.join(_BASE, "models", "expense_classifier_meta.json")

# Дефолтный порог. Используется если категория не найдена в CATEGORY_THRESHOLDS.
CONFIDENCE_THRESHOLD = 0.55

# Пороги по категориям — выставлены на основе F1 из cross-val:
#   высокий F1 (transport 0.84, transfer 0.82) → доверяем при низкой уверенности
#   низкий F1 (food 0.66, cafe 0.71)           → требуем более высокой уверенности
CATEGORY_THRESHOLDS: dict[str, float] = {
    "transport":     0.45,
    "transfer":      0.45,
    "health":        0.50,
    "entertainment": 0.50,
    "education":     0.50,
    "travel":        0.50,
    "utilities":     0.55,
    "shopping":      0.55,
    "cafe":          0.65,
    "food":          0.65,
}


@dataclass
class CategoryResult:
    category: str
    label_ru: str
    label_kz: str
    confidence: float
    confident: bool  # True если можно доверять предсказанию


class ExpenseClassifier:
    def __init__(self):
        self._pipeline = None
        self._meta: dict = {}
        self._load()

    def _load(self):
        if not os.path.exists(_MODEL_PATH):
            raise FileNotFoundError(
                f"Model not found at {_MODEL_PATH}. Run: python training/train.py"
            )
        self._pipeline = joblib.load(_MODEL_PATH)
        if os.path.exists(_META_PATH):
            with open(_META_PATH, encoding="utf-8") as f:
                self._meta = json.load(f)

    def predict(self, text: str) -> CategoryResult:
        proba = self._pipeline.predict_proba([text])[0]
        top_idx = int(np.argmax(proba))
        category = self._pipeline.classes_[top_idx]
        confidence = float(proba[top_idx])
        label_ru = self._meta.get("labels_ru", {}).get(category, category)
        label_kz = self._meta.get("labels_kz", {}).get(category, category)
        threshold = CATEGORY_THRESHOLDS.get(category, CONFIDENCE_THRESHOLD)
        return CategoryResult(
            category=category,
            label_ru=label_ru,
            label_kz=label_kz,
            confidence=confidence,
            confident=confidence >= threshold,
        )

    def predict_batch(self, texts: list[str]) -> list[CategoryResult]:
        return [self.predict(t) for t in texts]

    @property
    def meta(self) -> dict:
        return self._meta


_classifier: Optional[ExpenseClassifier] = None


def get_classifier() -> ExpenseClassifier:
    global _classifier
    if _classifier is None:
        _classifier = ExpenseClassifier()
    return _classifier

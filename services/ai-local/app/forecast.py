"""
Прогнозирование расходов по категориям.
Алгоритм: Holt-Winters (ExponentialSmoothing) из statsmodels.
При нехватке данных — взвешенное скользящее среднее как fallback.
"""

from __future__ import annotations

import warnings
from dataclasses import dataclass
from datetime import date, timedelta
from typing import Optional

import numpy as np
import pandas as pd

warnings.filterwarnings("ignore")


@dataclass
class ForecastPoint:
    date: str
    predicted: float
    lower: float
    upper: float


@dataclass
class CategoryForecast:
    category: str
    label_ru: str
    label_kz: str
    horizon_days: int
    total_predicted: float
    daily: list[ForecastPoint]
    method: str          # "holt_winters" | "moving_avg" | "mean"
    confidence: float    # 0–1


LABELS_RU = {
    "food": "Продукты", "cafe": "Кафе и рестораны", "transport": "Транспорт",
    "health": "Здоровье", "entertainment": "Развлечения", "utilities": "Коммунальные услуги",
    "shopping": "Покупки", "education": "Образование", "travel": "Путешествия",
    "transfer": "Переводы",
}
LABELS_KZ = {
    "food": "Азық-түлік", "cafe": "Мейрамханалар", "transport": "Көлік",
    "health": "Денсаулық", "entertainment": "Ойын-сауық", "utilities": "Коммуналдық қызметтер",
    "shopping": "Сатып алу", "education": "Білім", "travel": "Саяхат",
    "transfer": "Аударым",
}


def _holt_winters(series: pd.Series, horizon: int) -> tuple[np.ndarray, np.ndarray, np.ndarray]:
    from statsmodels.tsa.holtwinters import ExponentialSmoothing

    n = len(series)
    # Сезонность только если достаточно данных (≥2 периода по 7 дней)
    seasonal = "add" if n >= 14 else None
    seasonal_periods = 7 if seasonal else None

    model = ExponentialSmoothing(
        series,
        trend="add",
        seasonal=seasonal,
        seasonal_periods=seasonal_periods,
        initialization_method="estimated",
    ).fit(optimized=True)

    forecast = model.forecast(horizon)
    forecast = np.maximum(forecast, 0)

    # Доверительный интервал ±1 sigma остатков
    sigma = np.std(model.resid) if len(model.resid) > 1 else series.mean() * 0.2
    lower = np.maximum(forecast - 1.64 * sigma, 0)
    upper = forecast + 1.64 * sigma

    return forecast, lower, upper


def _moving_avg(series: pd.Series, horizon: int) -> tuple[np.ndarray, np.ndarray, np.ndarray]:
    window = min(7, len(series))
    avg = series.iloc[-window:].mean()
    std = series.iloc[-window:].std() if window > 1 else avg * 0.2
    std = std if not np.isnan(std) else avg * 0.2

    forecast = np.full(horizon, avg)
    lower = np.maximum(forecast - 1.64 * std, 0)
    upper = forecast + 1.64 * std
    return forecast, lower, upper


def forecast_category(
    transactions: list[dict],
    category: str,
    horizon_days: int = 30,
    ref_date: Optional[date] = None,
) -> Optional[CategoryForecast]:
    """
    transactions: [{"date": "YYYY-MM-DD", "amount": float, "category": str}]
    """
    ref_date = ref_date or date.today()

    cat_txs = [t for t in transactions if t.get("category") == category and t.get("amount", 0) > 0]
    if not cat_txs:
        return None

    # Агрегируем по дням
    df = pd.DataFrame(cat_txs)
    df["date"] = pd.to_datetime(df["date"])
    daily = df.groupby("date")["amount"].sum().asfreq("D", fill_value=0)

    # Не расширяем за последнюю транзакцию нулями — это искажает скользящее среднее.
    # ref_date используется только для меток дат в прогнозе.

    n = len(daily)
    method: str
    confidence: float

    if n >= 14:
        try:
            pred, low, high = _holt_winters(daily, horizon_days)
            method = "holt_winters"
            confidence = 0.85
        except Exception:
            pred, low, high = _moving_avg(daily, horizon_days)
            method = "moving_avg"
            confidence = 0.65
    elif n >= 3:
        pred, low, high = _moving_avg(daily, horizon_days)
        method = "moving_avg"
        confidence = 0.65
    else:
        avg = daily.mean()
        pred = np.full(horizon_days, avg)
        low = np.zeros(horizon_days)
        high = pred * 1.5
        method = "mean"
        confidence = 0.40

    daily_points = [
        ForecastPoint(
            date=(ref_date + timedelta(days=i + 1)).isoformat(),
            predicted=round(float(pred[i]), 2),
            lower=round(float(low[i]), 2),
            upper=round(float(high[i]), 2),
        )
        for i in range(horizon_days)
    ]

    return CategoryForecast(
        category=category,
        label_ru=LABELS_RU.get(category, category),
        label_kz=LABELS_KZ.get(category, category),
        horizon_days=horizon_days,
        total_predicted=round(float(pred.sum()), 2),
        daily=daily_points,
        method=method,
        confidence=confidence,
    )


def forecast_all(
    transactions: list[dict],
    horizon_days: int = 30,
    ref_date: Optional[date] = None,
) -> list[CategoryForecast]:
    categories = list({t["category"] for t in transactions if t.get("category")})
    results = []
    for cat in categories:
        fc = forecast_category(transactions, cat, horizon_days, ref_date)
        if fc:
            results.append(fc)
    results.sort(key=lambda x: x.total_predicted, reverse=True)
    return results

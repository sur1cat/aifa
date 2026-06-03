"""
Обнаружение аномальных расходов.

Алгоритм:
  1. Z-score per category — флагируем дни, где трата сильно выбивается из нормы.
  2. IsolationForest по суммарным дневным расходам — ловим дни с необычным
     общим паттерном (много категорий одновременно или очень крупный день).

Чувствительность (sensitivity):
  high   → z-порог 2.0  (много срабатываний)
  medium → z-порог 2.5  (баланс)
  low    → z-порог 3.0  (только явные выбросы)
"""

from __future__ import annotations

from dataclasses import dataclass

import numpy as np
import pandas as pd

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

SENSITIVITY_THRESHOLDS = {"high": 2.0, "medium": 2.5, "low": 3.0}
PAIR_RATIO_THRESHOLDS = {"high": 2.0, "medium": 3.0, "low": 4.0}


@dataclass
class AnomalyPoint:
    date: str
    category: str
    label_ru: str
    label_kz: str
    amount: float
    mean: float
    std: float
    z_score: float
    severity: str           # "low" | "medium" | "high"
    source: str             # "zscore" | "isolation_forest" | "both"
    expected_lower: float
    expected_upper: float


@dataclass
class AnomalyResult:
    anomalies: list[AnomalyPoint]
    total_anomalies: int
    sensitivity: str
    z_threshold: float
    method: str             # "zscore" | "zscore+isolation_forest"
    observation_days: int
    enough_history: bool
    stats: dict             # summary per category


def _severity(z: float, threshold: float) -> str:
    if z >= threshold + 1.0:
        return "high"
    if z >= threshold + 0.5:
        return "medium"
    return "low"


def _modified_zscore(values: np.ndarray) -> np.ndarray:
    """Modified Z-score через MAD — устойчив к выбросам (Iglewicz & Hoaglin, 1993)."""
    median = np.median(values)
    mad = np.median(np.abs(values - median))
    if mad == 0:
        # Все значения одинаковы или MAD=0 — fallback на обычный z-score
        std = values.std()
        if std == 0:
            return np.zeros(len(values))
        return np.abs(values - values.mean()) / std
    return 0.6745 * (values - median) / mad


def _zscore_anomalies(
    daily_cat: pd.DataFrame,      # columns: date, category, amount
    threshold: float,
    sensitivity: str,
) -> list[AnomalyPoint]:
    results: list[AnomalyPoint] = []

    for cat, grp in daily_cat.groupby("category"):
        if len(grp) < 2:
            continue

        amounts = grp["amount"].values.astype(float)
        if len(grp) == 2:
            ratio_threshold = PAIR_RATIO_THRESHOLDS.get(sensitivity, 3.0)
            ordered = grp.sort_values("amount").reset_index(drop=True)
            low = float(ordered.loc[0, "amount"])
            high = float(ordered.loc[1, "amount"])
            ratio = high / max(low, 1.0)
            if ratio >= ratio_threshold:
                row = ordered.loc[1]
                scale = max(high - low, low * 0.25, 1.0)
                z = max(threshold, ratio)
                results.append(AnomalyPoint(
                    date=str(pd.Timestamp(row["date"]).date()),
                    category=str(cat),
                    label_ru=LABELS_RU.get(str(cat), str(cat)),
                    label_kz=LABELS_KZ.get(str(cat), str(cat)),
                    amount=round(high, 2),
                    mean=round(low, 2),
                    std=round(scale, 2),
                    z_score=round(float(z), 3),
                    severity=_severity(float(z), threshold),
                    source="zscore",
                    expected_lower=round(max(0.0, low * 0.8), 2),
                    expected_upper=round(low * 1.2, 2),
                ))
            continue

        median = float(np.median(amounts))
        mad = float(np.median(np.abs(amounts - median)))
        # Для отображения expected range используем median±1.64*MAD/0.6745
        scale = mad / 0.6745 if mad > 0 else float(np.std(amounts))

        mz_scores = _modified_zscore(amounts)

        for (_, row), mz in zip(grp.iterrows(), mz_scores):
            if mz >= threshold:
                results.append(AnomalyPoint(
                    date=str(row["date"].date()),
                    category=str(cat),
                    label_ru=LABELS_RU.get(str(cat), str(cat)),
                    label_kz=LABELS_KZ.get(str(cat), str(cat)),
                    amount=round(float(row["amount"]), 2),
                    mean=round(median, 2),
                    std=round(scale, 2),
                    z_score=round(float(mz), 3),
                    severity=_severity(float(mz), threshold),
                    source="zscore",
                    expected_lower=round(max(0.0, median - 1.64 * scale), 2),
                    expected_upper=round(median + 1.64 * scale, 2),
                ))

    return results


def _isolation_anomalies(
    daily_total: pd.Series,       # index=date, values=total amount
    threshold: float,
    known_dates: set[str],
) -> list[AnomalyPoint]:
    """Ловит аномальные дни по суммарным тратам — только если не пойман z-score."""
    from sklearn.ensemble import IsolationForest

    if len(daily_total) < 5:
        return []

    X = daily_total.values.reshape(-1, 1)
    contamination = min(0.1, max(0.01, 1.0 / len(X)))
    clf = IsolationForest(contamination=contamination, random_state=42)
    labels = clf.fit_predict(X)       # -1 = аномалия

    mean = float(daily_total.mean())
    std = float(daily_total.std()) if daily_total.std() > 0 else mean * 0.2

    results: list[AnomalyPoint] = []
    for date, amount, lbl in zip(daily_total.index, daily_total.values, labels):
        date_str = str(date.date()) if hasattr(date, "date") else str(date)
        if lbl == -1 and date_str not in known_dates:
            z = abs(float(amount) - mean) / std if std > 0 else 0.0
            results.append(AnomalyPoint(
                date=date_str,
                category="total",
                label_ru="Суммарные расходы за день",
                label_kz="Күндік жалпы шығындар",
                amount=round(float(amount), 2),
                mean=round(mean, 2),
                std=round(std, 2),
                z_score=round(z, 3),
                severity=_severity(z, threshold),
                source="isolation_forest",
                expected_lower=round(max(0.0, mean - 1.64 * std), 2),
                expected_upper=round(mean + 1.64 * std, 2),
            ))

    return results


def _single_spend_anomalies(
    df: pd.DataFrame,
    threshold: float,
) -> list[AnomalyPoint]:
    """Fallback for short histories: one spend dominates the whole window."""
    if df.empty:
        return []

    total = float(df["amount"].sum())
    if total <= 0:
        return []

    max_idx = df["amount"].idxmax()
    row = df.loc[max_idx]
    amount = float(row["amount"])
    share = amount / total

    threshold_map = {3.0: 0.70, 2.5: 0.55, 2.0: 0.45}
    min_share = threshold_map.get(threshold, 0.55)
    if share < min_share:
        return []

    cat = str(row["category"])
    baseline = max(total - amount, 0.0)
    z = max(threshold, share * 5.0)
    return [
        AnomalyPoint(
            date=str(pd.Timestamp(row["date"]).date()),
            category=cat,
            label_ru=LABELS_RU.get(cat, cat),
            label_kz=LABELS_KZ.get(cat, cat),
            amount=round(amount, 2),
            mean=round(baseline, 2),
            std=round(max(baseline * 0.25, 1.0), 2),
            z_score=round(float(z), 3),
            severity=_severity(float(z), threshold),
            source="single_spend",
            expected_lower=0.0,
            expected_upper=round(max(total * 0.30, baseline), 2),
        )
    ]


def detect_anomalies(
    transactions: list[dict],
    sensitivity: str = "medium",
) -> AnomalyResult:
    """
    transactions: [{"date": "YYYY-MM-DD", "amount": float, "category": str}]
    sensitivity:  "low" | "medium" | "high"
    """
    threshold = SENSITIVITY_THRESHOLDS.get(sensitivity, 2.5)

    valid = [t for t in transactions if t.get("amount", 0) > 0 and t.get("category") and t.get("date")]
    if not valid:
        return AnomalyResult([], 0, sensitivity, threshold, "zscore", {})

    df = pd.DataFrame(valid)
    df["date"] = pd.to_datetime(df["date"])
    df["amount"] = df["amount"].astype(float)

    daily_cat = df.groupby(["date", "category"], as_index=False)["amount"].sum()
    daily_total = df.groupby("date")["amount"].sum()
    observation_days = int(df["date"].dt.date.nunique())
    enough_history = observation_days >= 5 or any(
        len(grp) >= 2 for _, grp in daily_cat.groupby("category")
    )

    zscore_hits = _zscore_anomalies(daily_cat, threshold, sensitivity)
    known = {a.date for a in zscore_hits}

    iso_hits = _isolation_anomalies(daily_total, threshold, known)
    single_hits = _single_spend_anomalies(df, threshold)

    seen = {(a.date, a.category, a.source) for a in zscore_hits + iso_hits}
    for item in single_hits:
        key = (item.date, item.category, item.source)
        if key not in seen:
            zscore_hits.append(item)
            seen.add(key)

    all_anomalies = sorted(zscore_hits + iso_hits, key=lambda a: (-a.z_score, a.date))

    # stats per category
    stats: dict = {}
    for cat, grp in daily_cat.groupby("category"):
        if len(grp) >= 2:
            stats[str(cat)] = {
                "mean_daily": round(float(grp["amount"].mean()), 2),
                "std_daily": round(float(grp["amount"].std()), 2),
                "max_daily": round(float(grp["amount"].max()), 2),
                "n_days": int(len(grp)),
            }

    method = "zscore+isolation_forest" if len(daily_total) >= 5 else "zscore"

    return AnomalyResult(
        anomalies=all_anomalies,
        total_anomalies=len(all_anomalies),
        sensitivity=sensitivity,
        z_threshold=threshold,
        method=method,
        observation_days=observation_days,
        enough_history=enough_history,
        stats=stats,
    )

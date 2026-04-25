"""
Статистические инсайты по расходам.

Endpoints (через main.py):
  POST /insights/summary          — сводка за период
  POST /insights/budget-suggest   — рекомендации бюджета по категориям
"""

from __future__ import annotations

from dataclasses import dataclass, field
from datetime import date, timedelta
from typing import Optional

import numpy as np
import pandas as pd


LABELS_RU = {
    "food": "Продукты", "cafe": "Кафе и рестораны", "transport": "Транспорт",
    "health": "Здоровье", "entertainment": "Развлечения", "utilities": "Коммунальные услуги",
    "shopping": "Покупки", "education": "Образование", "travel": "Путешествия",
    "transfer": "Переводы", "salary": "Зарплата", "savings": "Накопления",
}
LABELS_KZ = {
    "food": "Азық-түлік", "cafe": "Мейрамханалар", "transport": "Көлік",
    "health": "Денсаулық", "entertainment": "Ойын-сауық", "utilities": "Коммуналдық",
    "shopping": "Сатып алу", "education": "Білім", "travel": "Саяхат",
    "transfer": "Аударым", "salary": "Жалақы", "savings": "Жинақ",
}


# ── Summary ───────────────────────────────────────────────────────────────────

@dataclass
class CategoryStat:
    category: str
    label_ru: str
    label_kz: str
    amount: float
    pct: float          # доля от total_expense
    tx_count: int
    avg_per_tx: float


@dataclass
class SpendingSummary:
    period_start: str
    period_end: str
    total_income: float
    total_expense: float
    savings_rate: float         # (income - expense) / income, если income > 0
    net: float                  # income - expense
    avg_daily_expense: float
    top_categories: list[CategoryStat]
    expense_trend: str          # "up" | "down" | "stable" | "unknown"
    expense_trend_pct: float    # изменение в % относительно предыдущего периода


def spending_summary(
    transactions: list[dict],
    period_start: Optional[str] = None,
    period_end: Optional[str] = None,
    compare_prev: bool = True,
) -> SpendingSummary:
    """
    transactions: [{"date": "YYYY-MM-DD", "amount": float, "type": "income"|"expense", "category": str}]
    period_start / period_end: фильтр дат (включительно), по умолчанию — весь набор.
    """
    if not transactions:
        today = date.today().isoformat()
        return SpendingSummary(today, today, 0, 0, 0, 0, 0, [], "unknown", 0)

    df = pd.DataFrame(transactions)
    df["date"] = pd.to_datetime(df["date"], errors="coerce")
    df = df.dropna(subset=["date"])
    df["amount"] = pd.to_numeric(df["amount"], errors="coerce").fillna(0)

    if period_start:
        df = df[df["date"] >= pd.Timestamp(period_start)]
    if period_end:
        df = df[df["date"] <= pd.Timestamp(period_end)]

    if df.empty:
        today = date.today().isoformat()
        return SpendingSummary(today, today, 0, 0, 0, 0, 0, [], "unknown", 0)

    p_start = df["date"].min().date().isoformat()
    p_end   = df["date"].max().date().isoformat()

    expenses = df[df["type"] == "expense"]
    incomes  = df[df["type"] == "income"]

    total_expense = float(expenses["amount"].sum())
    total_income  = float(incomes["amount"].sum())
    net = total_income - total_expense
    savings_rate = round((net / total_income), 4) if total_income > 0 else 0.0

    # Дней в периоде
    days = max(1, (df["date"].max() - df["date"].min()).days + 1)
    avg_daily = round(total_expense / days, 2)

    # Топ категории расходов
    top_cats: list[CategoryStat] = []
    if not expenses.empty:
        cat_group = expenses.groupby("category").agg(
            amount=("amount", "sum"),
            tx_count=("amount", "count"),
        ).reset_index()
        cat_group = cat_group.sort_values("amount", ascending=False)
        for _, row in cat_group.iterrows():
            cat = str(row["category"])
            amt = float(row["amount"])
            cnt = int(row["tx_count"])
            top_cats.append(CategoryStat(
                category=cat,
                label_ru=LABELS_RU.get(cat, cat),
                label_kz=LABELS_KZ.get(cat, cat),
                amount=round(amt, 2),
                pct=round(amt / total_expense, 4) if total_expense > 0 else 0,
                tx_count=cnt,
                avg_per_tx=round(amt / cnt, 2) if cnt > 0 else 0,
            ))

    # Тренд: сравниваем первую и вторую половины периода
    trend = "unknown"
    trend_pct = 0.0
    if compare_prev and days >= 4:
        mid = df["date"].min() + pd.Timedelta(days=days // 2)
        first_half = expenses[expenses["date"] < mid]["amount"].sum()
        second_half = expenses[expenses["date"] >= mid]["amount"].sum()
        if first_half > 0:
            change = (second_half - first_half) / first_half
            trend_pct = round(change * 100, 1)
            if change > 0.05:
                trend = "up"
            elif change < -0.05:
                trend = "down"
            else:
                trend = "stable"

    return SpendingSummary(
        period_start=p_start,
        period_end=p_end,
        total_income=round(total_income, 2),
        total_expense=round(total_expense, 2),
        savings_rate=savings_rate,
        net=round(net, 2),
        avg_daily_expense=avg_daily,
        top_categories=top_cats,
        expense_trend=trend,
        expense_trend_pct=trend_pct,
    )


# ── Budget suggestions ────────────────────────────────────────────────────────

@dataclass
class BudgetSuggestion:
    category: str
    label_ru: str
    label_kz: str
    current_monthly_avg: float   # среднее за последние N месяцев
    suggested_limit: float       # 75-й перцентиль месячных трат
    overspend_months: int        # сколько месяцев превышали лимит
    reason: str
    priority: str                # "high" | "medium" | "low"


def budget_suggestions(
    transactions: list[dict],
    lookback_days: int = 90,
    percentile: float = 75,
) -> list[BudgetSuggestion]:
    """
    Анализирует последние `lookback_days` дней расходов и предлагает бюджет.
    Лимит = `percentile`-й перцентиль месячных трат по каждой категории.
    Категории с нестабильными тратами (CV > 0.5) помечаются как high priority.
    """
    if not transactions:
        return []

    df = pd.DataFrame(transactions)
    df["date"] = pd.to_datetime(df["date"], errors="coerce")
    df = df.dropna(subset=["date"])
    df["amount"] = pd.to_numeric(df["amount"], errors="coerce").fillna(0)

    cutoff = pd.Timestamp(date.today() - timedelta(days=lookback_days))
    df = df[(df["date"] >= cutoff) & (df["type"] == "expense") & (df["amount"] > 0)]

    if df.empty:
        return []

    df["month"] = df["date"].dt.to_period("M")

    suggestions: list[BudgetSuggestion] = []

    for cat, grp in df.groupby("category"):
        cat = str(cat)
        monthly = grp.groupby("month")["amount"].sum()
        if len(monthly) < 2:
            continue

        monthly_vals = monthly.values.astype(float)
        avg = float(np.mean(monthly_vals))
        suggested = float(np.percentile(monthly_vals, percentile))
        suggested = round(suggested / 500) * 500  # округляем до 500

        cv = float(np.std(monthly_vals) / avg) if avg > 0 else 0
        overspend = int(np.sum(monthly_vals > suggested))

        if cv > 0.5:
            priority = "high"
            reason = f"Траты нестабильны (разброс {round(cv*100)}%), стоит контролировать"
        elif overspend >= len(monthly_vals) // 2:
            priority = "high"
            reason = f"Превышение лимита в {overspend} из {len(monthly_vals)} месяцев"
        elif suggested < avg * 0.9:
            priority = "medium"
            reason = f"Средние траты {round(avg)} ₸, лимит поможет сэкономить"
        else:
            priority = "low"
            reason = f"Траты стабильны, лимит для контроля"

        suggestions.append(BudgetSuggestion(
            category=cat,
            label_ru=LABELS_RU.get(cat, cat),
            label_kz=LABELS_KZ.get(cat, cat),
            current_monthly_avg=round(avg, 2),
            suggested_limit=suggested,
            overspend_months=overspend,
            reason=reason,
            priority=priority,
        ))

    # Сортировка: high → medium → low, внутри — по avg убыванию
    priority_order = {"high": 0, "medium": 1, "low": 2}
    suggestions.sort(key=lambda s: (priority_order[s.priority], -s.current_monthly_avg))
    return suggestions

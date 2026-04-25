"""
AI-Local Service — FastAPI сервис с локальной ML моделью.

Endpoints:
  POST /categorize        — классифицировать одну транзакцию
  POST /categorize/batch  — классифицировать список транзакций
  POST /forecast          — прогноз расходов по категориям (Holt-Winters)
  POST /parse-message     — парсинг свободного текста в транзакцию (без OpenAI)
  GET  /model/info        — метаинформация и сравнение моделей
  GET  /model/report      — precision/recall/F1 по каждой категории + confusion matrix
  GET  /health            — health check
"""

from contextlib import asynccontextmanager
from datetime import date
from typing import Optional

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field

from app.model import get_classifier, CategoryResult, CATEGORY_THRESHOLDS, CONFIDENCE_THRESHOLD
from app.forecast import forecast_all, ForecastPoint, CategoryForecast
from app.anomaly import detect_anomalies, AnomalyResult
from app.message_parser import parse_message
from app.insights import spending_summary, budget_suggestions


@asynccontextmanager
async def lifespan(app: FastAPI):
    get_classifier()  # прогреваем модель при старте
    yield


app = FastAPI(
    title="AIFA AI-Local Service",
    description="Локальная ML модель классификации расходов",
    version="1.0.0",
    lifespan=lifespan,
)


# ── Schemas ──────────────────────────────────────────────────────────────────

class CategorizeRequest(BaseModel):
    text: str = Field(..., min_length=1, max_length=500, description="Название транзакции")

class CategorizeResponse(BaseModel):
    category: str
    label_ru: str
    label_kz: str
    confidence: float
    confident: bool
    text: str

class BatchCategorizeRequest(BaseModel):
    texts: list[str] = Field(..., min_length=1, max_length=100)

class BatchCategorizeResponse(BaseModel):
    results: list[CategorizeResponse]


class TransactionItem(BaseModel):
    date: str = Field(..., description="YYYY-MM-DD")
    amount: float = Field(..., gt=0)
    category: str

class ForecastRequest(BaseModel):
    transactions: list[TransactionItem] = Field(..., min_length=1)
    horizon_days: int = Field(30, ge=1, le=365)
    ref_date: Optional[str] = Field(None, description="YYYY-MM-DD, default=today")

class ForecastPointSchema(BaseModel):
    date: str
    predicted: float
    lower: float
    upper: float

class CategoryForecastSchema(BaseModel):
    category: str
    label_ru: str
    label_kz: str
    horizon_days: int
    total_predicted: float
    daily: list[ForecastPointSchema]
    method: str
    confidence: float

class ForecastResponse(BaseModel):
    forecasts: list[CategoryForecastSchema]
    horizon_days: int
    ref_date: str


class AnomalyRequest(BaseModel):
    transactions: list[TransactionItem] = Field(..., min_length=1)
    sensitivity: str = Field("medium", pattern="^(low|medium|high)$")

class AnomalyPointSchema(BaseModel):
    date: str
    category: str
    label_ru: str
    label_kz: str
    amount: float
    mean: float
    std: float
    z_score: float
    severity: str
    source: str
    expected_lower: float
    expected_upper: float

class AnomalyResponse(BaseModel):
    anomalies: list[AnomalyPointSchema]
    total_anomalies: int
    sensitivity: str
    z_threshold: float
    method: str
    stats: dict


# ── Routes ────────────────────────────────────────────────────────────────────

@app.get("/health")
def health():
    return {"status": "ok", "service": "ai-local"}


@app.get("/model/info")
def model_info():
    clf = get_classifier()
    meta = clf.meta
    return {
        "winner": meta.get("winner", "LogisticRegression"),
        "architecture": meta.get("architecture", "TF-IDF + LogisticRegression"),
        "cv_accuracy": meta.get("cv_accuracy"),
        "cv_std": meta.get("cv_std"),
        "n_training_samples": meta.get("n_training_samples"),
        "model_comparison": meta.get("model_comparison", {}),
        "categories": meta.get("categories", []),
        "labels_ru": meta.get("labels_ru", {}),
        "labels_kz": meta.get("labels_kz", {}),
        "confidence_thresholds": {
            cat: CATEGORY_THRESHOLDS.get(cat, CONFIDENCE_THRESHOLD)
            for cat in meta.get("categories", [])
        },
    }


@app.get("/model/report")
def model_report():
    clf = get_classifier()
    meta = clf.meta
    return {
        "winner": meta.get("winner"),
        "cv_accuracy": meta.get("cv_accuracy"),
        "macro_f1": meta.get("macro_f1"),
        "per_category": meta.get("per_category", {}),
        "confusion_matrix": meta.get("confusion_matrix", {}),
    }


@app.post("/categorize", response_model=CategorizeResponse)
def categorize(req: CategorizeRequest):
    clf = get_classifier()
    result: CategoryResult = clf.predict(req.text)
    return CategorizeResponse(
        text=req.text,
        category=result.category,
        label_ru=result.label_ru,
        label_kz=result.label_kz,
        confidence=round(result.confidence, 4),
        confident=result.confident,
    )


@app.post("/forecast", response_model=ForecastResponse)
def forecast(req: ForecastRequest):
    ref = None
    if req.ref_date:
        try:
            ref = date.fromisoformat(req.ref_date)
        except ValueError:
            raise HTTPException(status_code=422, detail="ref_date must be YYYY-MM-DD")

    txs = [t.model_dump() for t in req.transactions]
    results = forecast_all(txs, horizon_days=req.horizon_days, ref_date=ref)

    if not results:
        raise HTTPException(status_code=422, detail="No valid transactions to forecast")

    effective_ref = (ref or date.today()).isoformat()
    return ForecastResponse(
        forecasts=[
            CategoryForecastSchema(
                category=fc.category,
                label_ru=fc.label_ru,
                label_kz=fc.label_kz,
                horizon_days=fc.horizon_days,
                total_predicted=fc.total_predicted,
                daily=[ForecastPointSchema(**vars(p)) for p in fc.daily],
                method=fc.method,
                confidence=fc.confidence,
            )
            for fc in results
        ],
        horizon_days=req.horizon_days,
        ref_date=effective_ref,
    )


@app.post("/anomalies", response_model=AnomalyResponse)
def anomalies(req: AnomalyRequest):
    txs = [t.model_dump() for t in req.transactions]
    result = detect_anomalies(txs, sensitivity=req.sensitivity)
    return AnomalyResponse(
        anomalies=[AnomalyPointSchema(**vars(a)) for a in result.anomalies],
        total_anomalies=result.total_anomalies,
        sensitivity=result.sensitivity,
        z_threshold=result.z_threshold,
        method=result.method,
        stats=result.stats,
    )


class ParseMessageRequest(BaseModel):
    message: str = Field(..., min_length=1)
    debts_context: Optional[list] = Field(None, description="Список долгов пользователя для поиска при update_debt")


class ParsedTransactionSchema(BaseModel):
    type: str
    amount: float
    title: str
    category: str
    category_label: str
    date: str


class ParseMessageResponse(BaseModel):
    intent: str
    response: str
    transaction: Optional[ParsedTransactionSchema] = None
    counterparty: Optional[str] = None
    debt_direction: Optional[str] = None
    debt_update: Optional[dict] = None
    frequency: Optional[str] = None
    clarify_questions: list = Field(default_factory=list)
    # Task
    task_title: Optional[str] = None
    task_time: Optional[str] = None
    task_day: Optional[str] = None
    task_keywords: list = Field(default_factory=list)
    # Habit
    habit_title: Optional[str] = None
    habit_duration_days: Optional[int] = None
    habit_time: Optional[str] = None
    # Finance ecosystem
    goal_title: Optional[str] = None
    savings_amount: Optional[float] = None
    savings_period: Optional[str] = None
    alert_limit: Optional[float] = None
    alert_period: Optional[str] = None


@app.post("/parse-message", response_model=ParseMessageResponse)
def parse_message_endpoint(req: ParseMessageRequest):
    result = parse_message(req.message, debts_context=req.debts_context)
    tx = None
    if result.intent == "create_transaction":
        tx = ParsedTransactionSchema(
            type=result.tx_type,
            amount=result.amount,
            title=result.title,
            category=result.category,
            category_label=result.category_label,
            date=result.tx_date,
        )
    elif result.intent == "create_recurring" and result.amount:
        tx = ParsedTransactionSchema(
            type=result.tx_type or "expense",
            amount=result.amount,
            title=result.title or "Регулярный платёж",
            category=result.category or "utilities",
            category_label=result.category_label or "Коммунальные услуги",
            date=result.tx_date or "",
        )
    return ParseMessageResponse(
        intent=result.intent,
        response=result.response,
        transaction=tx,
        counterparty=result.counterparty,
        debt_direction=result.debt_direction,
        debt_update=result.debt_update,
        frequency=result.frequency,
        clarify_questions=result.clarify_questions,
        task_title=result.task_title,
        task_time=result.task_time,
        task_day=result.task_day,
        task_keywords=result.task_keywords,
        habit_title=result.habit_title,
        habit_duration_days=result.habit_duration_days,
        habit_time=result.habit_time,
        goal_title=result.goal_title,
        savings_amount=result.savings_amount,
        savings_period=result.savings_period,
        alert_limit=result.alert_limit,
        alert_period=result.alert_period,
    )


# ── Insights ──────────────────────────────────────────────────────────────────

class InsightTransactionItem(BaseModel):
    date: str
    amount: float = Field(..., gt=0)
    type: str = Field(..., pattern="^(income|expense)$")
    category: str = Field(default="")

class SummaryRequest(BaseModel):
    transactions: list[InsightTransactionItem] = Field(..., min_length=1)
    period_start: Optional[str] = Field(None, description="YYYY-MM-DD")
    period_end: Optional[str] = Field(None, description="YYYY-MM-DD")

class CategoryStatSchema(BaseModel):
    category: str
    label_ru: str
    label_kz: str
    amount: float
    pct: float
    tx_count: int
    avg_per_tx: float

class SummaryResponse(BaseModel):
    period_start: str
    period_end: str
    total_income: float
    total_expense: float
    savings_rate: float
    net: float
    avg_daily_expense: float
    top_categories: list[CategoryStatSchema]
    expense_trend: str
    expense_trend_pct: float

class BudgetSuggestRequest(BaseModel):
    transactions: list[InsightTransactionItem] = Field(..., min_length=1)
    lookback_days: int = Field(90, ge=30, le=365)
    percentile: float = Field(75, ge=50, le=95)

class BudgetSuggestionSchema(BaseModel):
    category: str
    label_ru: str
    label_kz: str
    current_monthly_avg: float
    suggested_limit: float
    overspend_months: int
    reason: str
    priority: str

class BudgetSuggestResponse(BaseModel):
    suggestions: list[BudgetSuggestionSchema]
    lookback_days: int
    percentile: float


@app.post("/insights/summary", response_model=SummaryResponse)
def insights_summary(req: SummaryRequest):
    txs = [t.model_dump() for t in req.transactions]
    result = spending_summary(txs, period_start=req.period_start, period_end=req.period_end)
    return SummaryResponse(
        period_start=result.period_start,
        period_end=result.period_end,
        total_income=result.total_income,
        total_expense=result.total_expense,
        savings_rate=result.savings_rate,
        net=result.net,
        avg_daily_expense=result.avg_daily_expense,
        top_categories=[CategoryStatSchema(**vars(c)) for c in result.top_categories],
        expense_trend=result.expense_trend,
        expense_trend_pct=result.expense_trend_pct,
    )


@app.post("/insights/budget-suggest", response_model=BudgetSuggestResponse)
def insights_budget_suggest(req: BudgetSuggestRequest):
    txs = [t.model_dump() for t in req.transactions]
    suggestions = budget_suggestions(txs, lookback_days=req.lookback_days, percentile=req.percentile)
    return BudgetSuggestResponse(
        suggestions=[BudgetSuggestionSchema(**vars(s)) for s in suggestions],
        lookback_days=req.lookback_days,
        percentile=req.percentile,
    )


@app.post("/categorize/batch", response_model=BatchCategorizeResponse)
def categorize_batch(req: BatchCategorizeRequest):
    clf = get_classifier()
    results = clf.predict_batch(req.texts)
    return BatchCategorizeResponse(
        results=[
            CategorizeResponse(
                text=text,
                category=r.category,
                label_ru=r.label_ru,
                label_kz=r.label_kz,
                confidence=round(r.confidence, 4),
                confident=r.confident,
            )
            for text, r in zip(req.texts, results)
        ]
    )

"""
Парсинг свободного текста пользователя в структурированный intent.

Поддерживаемые intents:
  create_transaction — обычный расход/доход
  create_debt        — создание долга (я должен / мне должны)
  update_debt        — частичный возврат долга
  create_recurring   — регулярный платёж / зарплата
  ask_clarify        — нужны уточнения (кредит)
  chat               — не распознано
"""

import re
from dataclasses import dataclass, field
from datetime import date
from typing import Optional

from app.model import get_classifier

# ── Нормализация суммы ────────────────────────────────────────────────────────
_AMOUNT_RE = re.compile(
    r'(\d[\d\s]*(?:[.,]\d{1,2})?)\s*([кkКK](?!\w))?'   # число + опциональный "к"
    r'\s*(?:тг|тенге|₸|руб|р(?:\b|\.)|kzt)?',
    re.IGNORECASE | re.UNICODE,
)

_TIME_PATTERN = re.compile(r'\b\d{1,2}:\d{2}\b')
_NON_FINANCIAL_UNITS = re.compile(
    r'\b(\d+)\s*(минут|мин|дней|день|дня|недел|месяц|лет|год|раз|км|кг|шт)\b',
    re.IGNORECASE | re.UNICODE,
)

def _extract_amount(text: str) -> Optional[float]:
    # Убираем временны́е паттерны (19:00) и нефинансовые числа (30 дней, 20 минут)
    clean = _TIME_PATTERN.sub('', text)
    clean = _NON_FINANCIAL_UNITS.sub('', clean)
    for m in _AMOUNT_RE.finditer(clean):
        raw = m.group(1).replace(' ', '').replace(',', '.')
        k_suffix = m.group(2)
        try:
            val = float(raw)
            if k_suffix:
                val *= 1000
            if val > 0:
                return val
        except ValueError:
            continue
    return None

# ── Тип транзакции ────────────────────────────────────────────────────────────
_INCOME_KW = [
    'закинул', 'закинула', 'скинул', 'скинула', 'перевёл', 'перевела',
    'получил', 'получила', 'заработал', 'заработала', 'зарплата', 'зп',
    'пришло', 'пришли', 'начислили', 'вернул', 'вернула',
    'earned', 'received', 'salary', 'income', 'got',
]
_EXPENSE_KW = [
    'потратил', 'потратила', 'заплатил', 'заплатила', 'купил', 'купила',
    'оплатил', 'оплатила', 'отправить', 'отправил', 'отправила',
    'перевести', 'скинуть', 'снял', 'сняла', 'взял', 'взяла',
    'paid', 'spent', 'bought', 'cost',
]

def _detect_type(text: str) -> Optional[str]:
    lower = text.lower()
    for kw in _INCOME_KW:
        if kw in lower:
            return 'income'
    for kw in _EXPENSE_KW:
        if kw in lower:
            return 'expense'
    # Если нет явного глагола но есть категориальное слово + сумма — expense по умолчанию
    for keywords, _ in _KEYWORD_CATEGORY:
        if any(kw in lower for kw in keywords):
            return 'expense'
    return None

# ── Контрагент (имя человека или организации) ─────────────────────────────────
_PREPOSITIONS = {'на', 'за', 'в', 'по', 'из', 'с', 'и', 'от', 'мне', 'у',
                 'ко', 'до', 'при', 'через', 'для', 'о', 'об', 'к', 'со'}

_COMMON_WORDS = {
    'я', 'он', 'она', 'они', 'мы', 'вы', 'ты', 'это', 'тот', 'та', 'те',
    'вот', 'нет', 'да', 'ну', 'ок', 'хочу', 'хочет', 'есть', 'нужно',
    'сегодня', 'вчера', 'завтра', 'купил', 'купила', 'взял', 'взяла',
}

def _extract_counterparty(text: str) -> Optional[str]:
    """Ищет имя собственное: слово с заглавной буквы (включая первое слово)."""
    words = text.split()
    candidates = []
    for i, w in enumerate(words):
        clean = re.sub(r'[^\w]', '', w)
        if not clean or clean.isdigit():
            continue
        low = clean.lower()
        if clean[0].isupper() and low not in _PREPOSITIONS and low not in _COMMON_WORDS:
            # Первое слово берём только если оно явно имя (не глагол/местоимение)
            if i == 0 and len(clean) < 3:
                continue
            candidates.append(clean)
    for c in candidates:
        if not c.isdigit():
            return c
    return None

# ── Категории расходов ────────────────────────────────────────────────────────
_KEYWORD_CATEGORY: list[tuple[list[str], str]] = [
    (['обед', 'ужин', 'завтрак', 'продукт', 'еда', 'супермаркет', 'grocery', 'lunch', 'dinner'], 'food'),
    (['кофе', 'кафе', 'ресторан', 'бар', 'coffee', 'cafe', 'restaurant'], 'cafe'),
    (['такси', 'метро', 'автобус', 'маршрутка', 'бензин', 'проезд', 'taxi', 'uber', 'transport'], 'transport'),
    (['аренд', 'коммунальн', 'электр', 'газ', 'интернет', 'услуг', 'rent', 'utility'], 'utilities'),
    (['одежда', 'обувь', 'покупк', 'shopping', 'clothes'], 'shopping'),
    (['аптек', 'врач', 'больниц', 'лекарств', 'health', 'pharmacy', 'doctor'], 'health'),
    (['кино', 'концерт', 'игр', 'развлечен', 'cinema', 'game', 'entertainment'], 'entertainment'),
    (['курс', 'книг', 'учёб', 'образован', 'school', 'course', 'education'], 'education'),
    (['отел', 'отпуск', 'авиа', 'поезд', 'hotel', 'flight', 'travel'], 'travel'),
    (['перевод', 'transfer', 'скинул', 'скинула', 'закинул', 'закинула', 'отправил', 'должен', 'долг'], 'transfer'),
    (['зарплат', 'доход', 'salary', 'income', 'заработ', 'зп'], 'income'),
]

_CATEGORY_LABELS_RU = {
    'food': 'Продукты', 'cafe': 'Кафе и рестораны', 'transport': 'Транспорт',
    'health': 'Здоровье', 'entertainment': 'Развлечения', 'utilities': 'Коммунальные услуги',
    'shopping': 'Покупки', 'education': 'Образование', 'travel': 'Путешествия',
    'transfer': 'Переводы', 'income': 'Доход',
}

def _categorize(text: str, tx_type: str) -> str:
    lower = text.lower()
    for keywords, cat in _KEYWORD_CATEGORY:
        if any(kw in lower for kw in keywords):
            return cat
    clf = get_classifier()
    if clf is not None and tx_type == 'expense':
        return clf.predict(text).category
    return 'income' if tx_type == 'income' else 'shopping'

# ── Построение заголовка транзакции ───────────────────────────────────────────
_STOP_WORDS = _PREPOSITIONS | {
    'я', 'мне', 'меня', 'мой', 'моя', 'моё', 'мои',
    'потратил', 'потратила', 'заплатил', 'купил', 'купила',
    'получил', 'получила', 'закинул', 'закинула', 'скинул', 'скинула',
    'оплатил', 'отправить', 'отправил', 'перевести',
    'тг', 'тенге', 'руб', 'рублей', 'рубль',
    'это', 'тот', 'эта', 'эти',
}

def _make_title(text: str, counterparty: Optional[str], tx_type: str) -> str:
    if counterparty:
        direction = 'от' if tx_type == 'income' else 'для'
        return f"{direction.capitalize()} {counterparty}"
    nums = re.sub(r'\d[\d\s.,]*(?:тг|тенге|₸|руб|р\b|kzt|[кkКK]\b)?', '', text, flags=re.IGNORECASE)
    words = [w for w in nums.split() if re.sub(r'[^\w]', '', w).lower() not in _STOP_WORDS and len(w) > 1]
    title = ' '.join(words[:3]).strip().capitalize()
    return title if title else ('Доход' if tx_type == 'income' else 'Расход')

# ── Детекция intent ───────────────────────────────────────────────────────────

_DEBT_I_OWE_KW = [
    'взял в долг', 'я должен', 'я должна', 'занял у', 'заняла у',
    'взял кредит', 'заняла кредит',
]
_DEBT_THEY_OWE_KW = [
    'должен мне', 'должна мне', 'должны мне', 'одолжил', 'одолжила',
    'нурс должен', 'он должен', 'она должна',
]
_UPDATE_DEBT_KW = [
    'вернул', 'вернула', 'за долг', 'отдал', 'отдала', 'погасил', 'погасила',
    'половину долга', 'часть долга',
]
_RECURRING_KW = [
    'зп', 'зарплата', 'каждый месяц', 'ежемесячно', 'плачу кредит',
    'регулярно', 'monthly', 'every month',
]
_CREDIT_KW = [
    'кредит', 'рассрочка', 'ипотека', 'кредитка',
]

# ── Task keywords ─────────────────────────────────────────────────────────────
_CREATE_TASK_KW = [
    'напомни', 'напоминай', 'добавь задачу', 'поставь задачу', 'создай задачу',
    'задача', 'нужно сделать', 'не забудь', 'remind me', 'добавь таск',
]
_COMPLETE_TASK_KW = [
    'я сделал', 'я сделала', 'выполнил', 'выполнила', 'закончил', 'закончила',
    'сделано', 'готово', 'я закончил', 'выполнено', 'done',
]

# ── Habit keywords ────────────────────────────────────────────────────────────
_CREATE_HABIT_KW = [
    'хочу читать', 'хочу бегать', 'хочу отжиматься', 'хочу заниматься',
    'хочу просыпаться', 'хочу медитировать', 'привычка', 'каждый день',
    'ежедневно', 'каждое утро', 'хочу делать', 'буду делать каждый',
    'дней подряд', 'челлендж',
]
_ARCHIVE_HABIT_KW = [
    'больше не хочу', 'перестать', 'прекратить', 'остановить привычку',
    'убери привычку', 'удали привычку', 'больше не буду',
]

# ── Finance ecosystem keywords ────────────────────────────────────────────────
_SAVINGS_PLAN_KW = [
    'хочу копить', 'хочу накопить', 'откладывать каждый месяц',
    'накопить на', 'копить на', 'откладывать ежемесячно',
    'каждый месяц откладывать', 'ежемесячно откладывать',
]
_SAVINGS_RULE_KW = [
    'с каждого дохода', 'с любого дохода', 'каждый раз когда получаю',
    'каждый раз, когда получаю', 'при каждом доходе',
    'откладывать с дохода', 'откладывать с каждого',
]
_SPENDING_ALERT_KW = [
    'предупреждай', 'предупреди', 'если трачу больше', 'трачу больше',
    'дневной лимит', 'суточный лимит', 'не трать больше',
    'alert', 'предупреждение о расходах',
]

# ── Time normalization ────────────────────────────────────────────────────────
_TIME_OF_DAY = {
    'утром': '09:00', 'утр': '09:00',
    'днём': '13:00', 'днем': '13:00',
    'вечером': '19:00', 'вечер': '19:00',
    'ночью': '22:00', 'ночью': '22:00',
}

_TIME_RE = re.compile(r'\b(\d{1,2})[:\.](\d{2})\b')
_TIME_HOUR_RE = re.compile(r'\bв\s+(\d{1,2})\s*(утра|вечера|ночи|дня|часов|час)?\b', re.IGNORECASE | re.UNICODE)
_RELATIVE_DAY_RE = re.compile(r'\b(сегодня|завтра|послезавтра)\b', re.IGNORECASE)

_HOUR_OFFSET = {'утра': 0, 'дня': 0, 'вечера': 12, 'ночи': 0, 'часов': 0, 'час': 0}


def _extract_task_time(text: str) -> Optional[str]:
    """Возвращает 'HH:MM' или None."""
    lower = text.lower()
    # Точное время HH:MM
    m = _TIME_RE.search(text)
    if m:
        h, mi = int(m.group(1)), int(m.group(2))
        if 0 <= h <= 23 and 0 <= mi <= 59:
            return f'{h:02d}:{mi:02d}'
    # "в 5 утра", "в 19 часов"
    m2 = _TIME_HOUR_RE.search(lower)
    if m2:
        h = int(m2.group(1))
        part = (m2.group(2) or '').lower()
        if part == 'вечера' and h < 12:
            h += 12
        if 0 <= h <= 23:
            return f'{h:02d}:00'
    # Время суток
    for word, t in _TIME_OF_DAY.items():
        if word in lower:
            return t
    return None


def _extract_task_day(text: str) -> str:
    """Возвращает 'today' | 'tomorrow' | 'day_after_tomorrow'."""
    lower = text.lower()
    if 'послезавтра' in lower:
        return 'day_after_tomorrow'
    if 'завтра' in lower:
        return 'tomorrow'
    return 'today'


def _extract_task_title(text: str) -> str:
    """Извлекает название задачи из текста."""
    lower = text.lower()
    # Убираем триггерные слова
    strip_kw = [
        'напомни мне', 'напомни', 'добавь задачу', 'поставь задачу',
        'создай задачу', 'задача', 'нужно сделать', 'не забудь',
        'сегодня', 'завтра', 'послезавтра',
        'утром', 'днём', 'днем', 'вечером', 'ночью',
    ]
    clean = lower
    for kw in strip_kw:
        clean = clean.replace(kw, ' ')
    # Убираем время
    clean = _TIME_RE.sub('', clean)
    # Убираем предлоги в конце ("в", "на", "по")
    clean = re.sub(r'\b(в|на|по|к)\s*$', '', clean.strip())
    words = [w for w in clean.split() if len(w) > 1]
    title = ' '.join(words[:5]).strip().capitalize()
    return title if title else 'Задача'


def _extract_habit_title(text: str) -> str:
    """Извлекает название привычки."""
    lower = text.lower()
    strip_kw = [
        'больше не хочу', 'больше не буду', 'не хочу', 'хочу',
        'буду делать', 'буду', 'хочу делать',
        'каждый день', 'ежедневно', 'каждое утро', 'каждую неделю',
        'дней подряд', 'подряд', 'челлендж', 'привычка',
        'перестать', 'прекратить', 'остановить', 'убери', 'удали',
        'по утрам', 'утром', 'утра', 'вечером',
        'минут', 'мин', 'по',
    ]
    clean = lower
    for kw in sorted(strip_kw, key=len, reverse=True):  # длинные сначала
        clean = clean.replace(kw, ' ')
    clean = re.sub(r'\b\d+\b', '', clean)
    words = [w for w in clean.split() if len(w) > 2]
    title = ' '.join(words[:3]).strip().capitalize()
    return title if title else 'Привычка'


def _extract_goal_title(text: str) -> str:
    """Извлекает название цели из 'накопить на X' / 'копить на X'."""
    m = re.search(r'(?:накопить|копить|откладывать|копить)\s+на\s+([\w\s]+?)(?:\s*$|,|\.|!)', text.lower())
    if m:
        return m.group(1).strip().capitalize()
    return 'Накопления'


def _extract_duration_days(text: str) -> Optional[int]:
    """Извлекает '30 дней' → 30."""
    m = re.search(r'(\d+)\s*(?:дн|день|дней|days)', text.lower())
    return int(m.group(1)) if m else None


def _detect_intent(text: str, amount: Optional[float], tx_type: Optional[str]) -> str:
    lower = text.lower()

    # Финансовая экосистема — раньше общих финансов
    if any(kw in lower for kw in _SPENDING_ALERT_KW) and amount:
        return 'create_spending_alert'
    if any(kw in lower for kw in _SAVINGS_RULE_KW) and amount:
        return 'create_savings_rule'
    if any(kw in lower for kw in _SAVINGS_PLAN_KW) and amount:
        return 'create_savings_plan'

    # Признаки финансового сообщения (деньги/суммы с валютой)
    has_finance = bool(amount) and any(
        kw in lower for kw in ['тг', 'тенге', '₸', 'руб', 'kzt', 'потратил', 'заплатил',
                                'купил', 'получил', 'доход', 'расход', 'зарплат', 'зп',
                                'долг', 'должен', 'кредит']
    )

    # Архивировать привычку
    if any(kw in lower for kw in _ARCHIVE_HABIT_KW):
        return 'archive_habit'

    # Создать привычку — без финансового контекста
    if any(kw in lower for kw in _CREATE_HABIT_KW) and not has_finance:
        return 'create_habit'

    # Задача выполнена
    if any(kw in lower for kw in _COMPLETE_TASK_KW) and not has_finance:
        return 'complete_task'

    # Создать задачу
    if any(kw in lower for kw in _CREATE_TASK_KW) and not has_finance:
        return 'create_task'

    # Регулярный платёж / зарплата — проверяем раньше кредита
    if any(kw in lower for kw in _RECURRING_KW) and amount:
        return 'create_recurring'

    # Кредит без признаков регулярности → нужны уточнения
    if any(kw in lower for kw in _CREDIT_KW) and amount and amount > 1000:
        return 'ask_clarify'

    # Возврат долга
    if any(kw in lower for kw in _UPDATE_DEBT_KW):
        return 'update_debt'

    # Долг (я должен)
    if any(kw in lower for kw in _DEBT_I_OWE_KW):
        return 'create_debt'

    # Долг (мне должны)
    if any(kw in lower for kw in _DEBT_THEY_OWE_KW):
        return 'create_debt'

    # Обычная транзакция
    if amount and tx_type:
        return 'create_transaction'

    return 'chat'

# ── Направление долга ─────────────────────────────────────────────────────────
def _debt_direction(text: str) -> str:
    """i_owe — я должен; they_owe — мне должны."""
    lower = text.lower()
    if any(kw in lower for kw in _DEBT_THEY_OWE_KW):
        return 'they_owe'
    return 'i_owe'

# ── Частичный возврат ─────────────────────────────────────────────────────────
def _partial_return(text: str, amount: Optional[float]) -> dict:
    """Возвращает {type: 'half'|'amount', value: float|None}."""
    lower = text.lower()
    if 'половин' in lower:
        return {'type': 'half', 'value': None}
    if amount:
        return {'type': 'amount', 'value': amount}
    return {'type': 'unknown', 'value': None}

# ── Периодичность для recurring ───────────────────────────────────────────────
def _detect_frequency(text: str) -> str:
    lower = text.lower()
    if any(w in lower for w in ['каждую неделю', 'weekly', 'еженедельно']):
        return 'weekly'
    if any(w in lower for w in ['каждый день', 'ежедневно', 'каждое утро', 'daily', 'подряд']):
        return 'daily'
    return 'monthly'

# ─────────────────────────────────────────────────────────────────────────────
# Главная функция
# ─────────────────────────────────────────────────────────────────────────────

@dataclass
class ParsedMessage:
    intent: str
    response: str
    tx_type: Optional[str] = None
    amount: Optional[float] = None
    title: Optional[str] = None
    category: Optional[str] = None
    category_label: Optional[str] = None
    tx_date: Optional[str] = None
    counterparty: Optional[str] = None
    debt_direction: Optional[str] = None
    debt_update: Optional[dict] = None
    frequency: Optional[str] = None
    clarify_questions: list = field(default_factory=list)
    # Task fields
    task_title: Optional[str] = None
    task_time: Optional[str] = None      # "HH:MM"
    task_day: Optional[str] = None       # today | tomorrow | day_after_tomorrow
    task_keywords: list = field(default_factory=list)
    # Habit fields
    habit_title: Optional[str] = None
    habit_duration_days: Optional[int] = None
    habit_time: Optional[str] = None
    # Finance ecosystem fields
    goal_title: Optional[str] = None
    savings_amount: Optional[float] = None
    savings_period: Optional[str] = None   # monthly | on_income
    alert_limit: Optional[float] = None
    alert_period: Optional[str] = None    # daily | monthly


def parse_message(text: str, debts_context: Optional[list] = None) -> ParsedMessage:
    today = date.today().isoformat()
    amount = _extract_amount(text)
    tx_type = _detect_type(text)
    counterparty = _extract_counterparty(text)
    intent = _detect_intent(text, amount, tx_type)

    # ── create_transaction ────────────────────────────────────────────────────
    if intent == 'create_transaction':
        category = _categorize(text, tx_type)
        label = _CATEGORY_LABELS_RU.get(category, category)
        title = _make_title(text, counterparty, tx_type)

        if counterparty and tx_type == 'income':
            response = f"Деньги от {counterparty} получены, запись добавлена. {amount:,.0f}₸ → {label}."
        elif counterparty and tx_type == 'expense':
            response = f"Перевод для {counterparty} записан. {amount:,.0f}₸ → {label}."
        elif tx_type == 'income':
            response = f"Записал доход {amount:,.0f}₸ ({label})."
        else:
            response = f"Записал расход {amount:,.0f}₸ ({label})."

        return ParsedMessage(
            intent='create_transaction', response=response,
            tx_type=tx_type, amount=amount, title=title,
            category=category, category_label=label,
            tx_date=today, counterparty=counterparty,
        )

    # ── create_debt ───────────────────────────────────────────────────────────
    if intent == 'create_debt':
        direction = _debt_direction(text)
        if not amount:
            return ParsedMessage(intent='chat', response='Укажи сумму долга.')
        party = counterparty or 'Неизвестно'
        if direction == 'i_owe':
            response = f"Записал: ты должен {party} — {amount:,.0f}₸."
        else:
            response = f"Записал: {party} теперь должен тебе {amount:,.0f}₸."
        return ParsedMessage(
            intent='create_debt', response=response,
            amount=amount, counterparty=party,
            debt_direction=direction, tx_date=today,
        )

    # ── update_debt ───────────────────────────────────────────────────────────
    if intent == 'update_debt':
        party = counterparty or ''
        partial = _partial_return(text, amount)

        # Ищем долг в контексте
        debt_found = None
        if debts_context and party:
            for d in debts_context:
                if party.lower() in (d.get('counterparty') or '').lower():
                    debt_found = d
                    break

        if not debt_found and debts_context is not None:
            return ParsedMessage(
                intent='update_debt',
                response=f'Не нашёл долг от {party}. Проверь список долгов.',
                counterparty=party, debt_update={'type': 'not_found'},
            )

        if partial['type'] == 'half' and debt_found:
            reduce_by = debt_found.get('amount', 0) / 2
            response = f"{party} вернул половину долга ({reduce_by:,.0f}₸), всё пересчитано."
        elif partial['type'] == 'amount' and partial['value']:
            reduce_by = partial['value']
            if debt_found and reduce_by > debt_found.get('amount', 0):
                response = f"Сумма возврата ({reduce_by:,.0f}₸) больше остатка долга. Уточни сумму."
            else:
                response = f"Возврат долга от {party} на {reduce_by:,.0f}₸ записан."
        else:
            reduce_by = amount
            response = f"Возврат долга от {party} записан."

        return ParsedMessage(
            intent='update_debt', response=response,
            amount=reduce_by, counterparty=party,
            tx_date=today,
            debt_update={'type': partial['type'], 'reduce_by': reduce_by, 'debt_id': debt_found.get('id') if debt_found else None},
        )

    # ── create_recurring ─────────────────────────────────────────────────────
    if intent == 'create_recurring':
        if not amount:
            return ParsedMessage(intent='chat', response='Укажи сумму регулярного платежа.')
        freq = _detect_frequency(text)
        # Зарплата → income, иначе expense
        lower = text.lower()
        rec_type = 'income' if any(w in lower for w in ['зп', 'зарплат', 'salary']) else 'expense'
        title = 'Зарплата' if rec_type == 'income' else _make_title(text, counterparty, rec_type)
        if rec_type == 'income':
            response = f"Зарплата сохранена ({amount:,.0f}₸/мес), буду учитывать автоматически каждый месяц."
        else:
            response = f"Регулярный платёж {amount:,.0f}₸ ({freq}) добавлен."
        return ParsedMessage(
            intent='create_recurring', response=response,
            tx_type=rec_type, amount=amount, title=title,
            category='income' if rec_type == 'income' else 'utilities',
            tx_date=today, frequency=freq,
        )

    # ── create_task ───────────────────────────────────────────────────────────
    if intent == 'create_task':
        task_title = _extract_task_title(text)
        task_time = _extract_task_time(text)
        task_day = _extract_task_day(text)
        day_ru = {'today': 'сегодня', 'tomorrow': 'завтра', 'day_after_tomorrow': 'послезавтра'}.get(task_day, 'сегодня')
        time_str = f' в {task_time}' if task_time else ''
        response = f"Напомню «{task_title}» {day_ru}{time_str}."
        return ParsedMessage(
            intent='create_task', response=response,
            task_title=task_title, task_time=task_time, task_day=task_day,
        )

    # ── complete_task ─────────────────────────────────────────────────────────
    if intent == 'complete_task':
        # Извлекаем ключевые слова для поиска задачи
        lower = text.lower()
        stop = {'я', 'сделал', 'сделала', 'выполнил', 'выполнила', 'закончил', 'закончила', 'готово', 'сделано'}
        keywords = [w for w in re.findall(r'\w+', lower) if w not in stop and len(w) > 2]
        if keywords:
            response = f"Отлично! Отметил «{keywords[0]}» как выполненное."
        else:
            response = "Что именно ты сделал? Уточни название задачи."
        return ParsedMessage(
            intent='complete_task', response=response,
            task_keywords=keywords,
        )

    # ── create_habit ──────────────────────────────────────────────────────────
    if intent == 'create_habit':
        habit_title = _extract_habit_title(text)
        duration = _extract_duration_days(text)
        habit_time = _extract_task_time(text)
        freq = _detect_frequency(text)
        if duration:
            response = f"Запустили челлендж «{habit_title}» на {duration} дней, погнали!"
        else:
            response = f"Привычка «{habit_title}» добавлена, буду напоминать {('каждый день' if freq == 'daily' else 'каждую неделю')}."
        return ParsedMessage(
            intent='create_habit', response=response,
            habit_title=habit_title,
            habit_duration_days=duration,
            habit_time=habit_time,
            frequency='daily' if freq == 'monthly' else freq,
        )

    # ── archive_habit ─────────────────────────────────────────────────────────
    if intent == 'archive_habit':
        habit_title = _extract_habit_title(text)
        response = f"Окей, привычка «{habit_title}» остановлена."
        return ParsedMessage(
            intent='archive_habit', response=response,
            habit_title=habit_title,
        )

    # ── create_savings_plan ───────────────────────────────────────────────────
    if intent == 'create_savings_plan':
        goal_title = _extract_goal_title(text)
        freq = _detect_frequency(text)
        period = 'monthly' if freq in ('monthly', 'daily') else freq
        response = (f"Добавил цель «{goal_title}» и создал правило откладывать "
                    f"{amount:,.0f}₸ каждый месяц. Буду напоминать!")
        return ParsedMessage(
            intent='create_savings_plan', response=response,
            goal_title=goal_title,
            savings_amount=amount,
            savings_period=period,
            frequency=period,
        )

    # ── create_savings_rule ───────────────────────────────────────────────────
    if intent == 'create_savings_rule':
        response = (f"Теперь с каждого дохода буду автоматически откладывать "
                    f"{amount:,.0f}₸ в накопления.")
        return ParsedMessage(
            intent='create_savings_rule', response=response,
            savings_amount=amount,
            savings_period='on_income',
        )

    # ── create_spending_alert ─────────────────────────────────────────────────
    if intent == 'create_spending_alert':
        response = (f"Понял! Если суточные расходы превысят {amount:,.0f}₸ — "
                    f"сразу предупрежу.")
        return ParsedMessage(
            intent='create_spending_alert', response=response,
            alert_limit=amount,
            alert_period='daily',
        )

    # ── ask_clarify (кредит) ──────────────────────────────────────────────────
    if intent == 'ask_clarify':
        questions = [
            'Какова общая сумма кредита?',
            'На какой срок (в месяцах)?',
            'Процентная ставка (% годовых)? Если рассрочка — напиши 0.',
            'Дата первого платежа?',
            'Название кредита (например: Kaspi, Отбасы)?',
        ]
        response = (
            f"Добавил кредит с платежом {amount:,.0f}₸/мес. "
            "Давай уточним детали, чтобы всё корректно считать:"
        )
        return ParsedMessage(
            intent='ask_clarify', response=response,
            amount=amount, tx_date=today,
            clarify_questions=questions,
        )

    # ── chat (не распознано) ──────────────────────────────────────────────────
    return ParsedMessage(
        intent='chat',
        response='Не могу распознать операцию. Уточни сумму и тип (потратил/получил/должен).',
    )

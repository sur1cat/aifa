"""
AIFA e2e test suite.
Запуск: python3 e2e_test.py
Требует: все контейнеры запущены через docker compose up -d
"""

import json
import sys
import time
import urllib.error
import urllib.request
import subprocess

# ── Конфигурация ──────────────────────────────────────────────────────────────

SERVICES = {
    "auth":    "http://localhost:8001",
    "user":    "http://localhost:8002",
    "habit":   "http://localhost:8003",
    "goal":    "http://localhost:8004",
    "task":    "http://localhost:8005",
    "finance": "http://localhost:8006",
    "ai":      "http://localhost:8007",
    "budget":  "http://localhost:8008",
    "ai-local":"http://localhost:8010",
}

PHONE = "+77009876543"

# ── Helpers ───────────────────────────────────────────────────────────────────

PASS = 0
FAIL = 0

def req(method, url, body=None, token=None, expect=None, multipart=False):
    data = json.dumps(body, ensure_ascii=False).encode() if body else None
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    r = urllib.request.Request(url, data=data, method=method, headers=headers)
    try:
        with urllib.request.urlopen(r) as resp:
            raw = json.loads(resp.read())
            status = resp.status
    except urllib.error.HTTPError as e:
        body = e.read()
        try:
            raw = json.loads(body)
        except Exception:
            raw = {"_raw": body.decode(errors="replace")}
        status = e.code
    if expect and status != expect:
        raise AssertionError(f"Expected HTTP {expect}, got {status}: {raw}")
    return raw, status


def check(name, expr, got=None):
    global PASS, FAIL
    if expr:
        PASS += 1
        print(f"  ✓  {name}")
    else:
        FAIL += 1
        print(f"  ✗  {name}" + (f" | got: {got}" if got is not None else ""))


def section(title):
    print(f"\n{'─'*55}")
    print(f"  {title}")
    print(f"{'─'*55}")


def get_token():
    """OTP-аутентификация; возвращает access_token."""
    req("POST", SERVICES["auth"] + "/auth/otp/send", {"phone": PHONE})
    time.sleep(0.5)
    logs = subprocess.run(
        ["docker", "compose", "logs", "auth-service"],
        capture_output=True, text=True
    ).stdout
    code = None
    for line in reversed(logs.splitlines()):
        if "OTP generated" in line and PHONE in line:
            import re
            m = re.search(r'"code":"(\d+)"', line)
            if m:
                code = m.group(1)
                break
    assert code, "Could not find OTP in logs"
    resp, _ = req("POST", SERVICES["auth"] + "/auth/otp/verify", {"phone": PHONE, "code": code})
    return resp["data"]["tokens"]["access_token"]

# ── Tests ─────────────────────────────────────────────────────────────────────

def test_health():
    section("Health checks")
    for name, base in SERVICES.items():
        resp, status = req("GET", base + "/health")
        check(f"{name} /health → 200", status == 200)


def test_auth(token):
    section("Auth")
    # Нельзя получить профиль без токена
    _, status = req("GET", SERVICES["user"] + "/users/me")
    check("GET /users/me без токена → 401", status == 401)

    # С токеном — OK
    resp, status = req("GET", SERVICES["user"] + "/users/me", token=token)
    check("GET /users/me с токеном → 200", status == 200)
    check("users/me содержит id", "id" in resp.get("data", {}))


def test_transactions(token):
    section("Transactions")

    # Создать расход
    resp, status = req("POST", SERVICES["finance"] + "/transactions", {
        "title": "Обед", "amount": 1500, "type": "expense",
        "category": "food", "date": "2026-04-24"
    }, token=token)
    check("POST /transactions → 201", status == 201)
    tx_id = resp.get("data", {}).get("id")
    check("transaction has id", bool(tx_id))

    # Создать доход
    resp2, status2 = req("POST", SERVICES["finance"] + "/transactions", {
        "title": "Зарплата", "amount": 300000, "type": "income",
        "category": "income", "date": "2026-04-24"
    }, token=token)
    check("POST /transactions income → 201", status2 == 201)

    # Список
    resp3, status3 = req("GET", SERVICES["finance"] + "/transactions", token=token)
    check("GET /transactions → 200", status3 == 200)
    check("transactions list not empty", len(resp3.get("data", [])) > 0)

    # Summary
    resp4, status4 = req("GET", SERVICES["finance"] + "/transactions/summary", token=token)
    check("GET /transactions/summary → 200", status4 == 200)
    check("summary has income/expense", "income" in resp4.get("data", {}) or "total_income" in resp4.get("data", {}))

    # Delete
    _, status5 = req("DELETE", SERVICES["finance"] + f"/transactions/{tx_id}", token=token)
    check("DELETE /transactions/:id → 200", status5 == 200)

    return resp2["data"]["id"]  # income tx id for cleanup later


def test_budgets(token):
    section("Budgets")

    resp, status = req("POST", SERVICES["finance"] + "/budgets", {
        "category": "food", "monthly_limit": 50000
    }, token=token)
    check("POST /budgets → 201", status == 201)
    budget_id = resp.get("data", {}).get("id")

    resp2, status2 = req("GET", SERVICES["finance"] + "/budgets", token=token)
    check("GET /budgets → 200", status2 == 200)
    check("budgets list has food", any(b["category"] == "food" for b in resp2.get("data", [])))

    _, status3 = req("PUT", SERVICES["finance"] + f"/budgets/{budget_id}", {
        "monthly_limit": 60000
    }, token=token)
    check("PUT /budgets/:id → 200", status3 == 200)

    _, status4 = req("DELETE", SERVICES["finance"] + f"/budgets/{budget_id}", token=token)
    check("DELETE /budgets/:id → 200", status4 == 200)


def test_debts(token):
    section("Debts")

    # Создать долг (мне должны)
    resp, status = req("POST", SERVICES["finance"] + "/debts", {
        "counterparty": "Ким", "direction": "they_owe", "amount": 10000
    }, token=token)
    check("POST /debts (they_owe) → 201", status == 201)
    debt_id = resp.get("data", {}).get("id")
    check("debt has id", bool(debt_id))
    check("debt direction=they_owe", resp["data"].get("direction") == "they_owe")
    check("debt amount=10000", resp["data"].get("amount") == 10000)

    # Создать долг (я должен)
    resp2, status2 = req("POST", SERVICES["finance"] + "/debts", {
        "counterparty": "Айгерим", "direction": "i_owe", "amount": 5000,
        "note": "за кофе"
    }, token=token)
    check("POST /debts (i_owe) → 201", status2 == 201)
    debt2_id = resp2["data"]["id"]

    # Список активных
    resp3, status3 = req("GET", SERVICES["finance"] + "/debts?settled=false", token=token)
    check("GET /debts?settled=false → 200", status3 == 200)
    check("active debts count >= 2", len(resp3.get("data", [])) >= 2)

    # Частичный возврат reduce_by
    resp4, status4 = req("PATCH", SERVICES["finance"] + f"/debts/{debt_id}", {
        "reduce_by": 3000
    }, token=token)
    check("PATCH /debts/:id reduce_by=3000 → 200", status4 == 200)
    check("amount reduced to 7000", resp4["data"].get("amount") == 7000)
    check("not settled after partial", resp4["data"].get("settled") == False)

    # Полное погашение
    resp5, status5 = req("PATCH", SERVICES["finance"] + f"/debts/{debt_id}", {
        "settle": True
    }, token=token)
    check("PATCH /debts/:id settle=true → 200", status5 == 200)
    check("amount=0 after settle", resp5["data"].get("amount") == 0)
    check("settled=true", resp5["data"].get("settled") == True)

    # Список закрытых
    resp6, status6 = req("GET", SERVICES["finance"] + "/debts?settled=true", token=token)
    check("GET /debts?settled=true → 200", status6 == 200)
    check("settled list has Ким", any(d["counterparty"] == "Ким" for d in resp6.get("data", [])))

    # Удалить долг Айгерим
    _, status7 = req("DELETE", SERVICES["finance"] + f"/debts/{debt2_id}", token=token)
    check("DELETE /debts/:id → 200", status7 == 200)

    # Невалидный direction
    _, status8 = req("POST", SERVICES["finance"] + "/debts", {
        "counterparty": "Х", "direction": "wrong", "amount": 100
    }, token=token)
    check("POST /debts invalid direction → 400", status8 == 400)

    return debt_id  # закрытый долг Кима


def test_parse_message(token):
    section("AI: parse-message")

    AI = SERVICES["ai"]

    def pm(msg, extra=None):
        body = {"message": msg}
        if extra:
            body.update(extra)
        resp, _ = req("POST", AI + "/ai/parse-message", body, token=token)
        return resp.get("data", {})

    # 1. Простой расход без глагола
    d = pm("обед 7000")
    check("1: 'обед 7000' → create_transaction", d.get("intent") == "create_transaction")
    check("1: type=expense", d.get("transaction", {}).get("type") == "expense")
    check("1: amount=7000", d.get("transaction", {}).get("amount") == 7000)
    check("1: category=food", d.get("transaction", {}).get("category") == "food")

    # 2. Доход с суффиксом "к" и именем
    d = pm("Сестра закинула 3к")
    check("2: 'Сестра закинула 3к' → create_transaction", d.get("intent") == "create_transaction")
    check("2: amount=3000 (3к)", d.get("transaction", {}).get("amount") == 3000)
    check("2: type=income", d.get("transaction", {}).get("type") == "income")
    check("2: counterparty=Сестра", d.get("counterparty") == "Сестра")

    # 3. Долг — мне должны
    d = pm("Нурс должен мне 3к")
    check("3: 'Нурс должен мне 3к' → create_debt", d.get("intent") == "create_debt")
    check("3: direction=they_owe", d.get("debt_direction") == "they_owe")
    check("3: counterparty=Нурс", d.get("counterparty") == "Нурс")
    check("3: amount=3000", d.get("transaction") is None and "3,000" in d.get("response", ""))

    # 4. Долг — я должен
    d = pm("Я взял в долг у Кима 5000")
    check("4: 'взял в долг у Кима' → create_debt", d.get("intent") == "create_debt")
    check("4: direction=i_owe", d.get("debt_direction") == "i_owe")
    check("4: counterparty=Кима", d.get("counterparty") == "Кима")

    # 5. Частичный возврат долга с контекстом
    d = pm("Ким вернул половину долга", {
        "debts_context": [{"id": "debt-abc", "counterparty": "Ким", "amount": 10000}]
    })
    check("5: 'Ким вернул половину' → update_debt", d.get("intent") == "update_debt")
    check("5: debt_update.type=half", (d.get("debt_update") or {}).get("type") == "half")
    check("5: reduce_by=5000", (d.get("debt_update") or {}).get("reduce_by") == 5000)
    check("5: debt_id передан", (d.get("debt_update") or {}).get("debt_id") == "debt-abc")

    # 6. Кредит → ask_clarify
    d = pm("Взял кредит в Kaspi 50000 в месяц")
    check("6: кредит → ask_clarify", d.get("intent") == "ask_clarify")
    check("6: clarify_questions не пустой", len(d.get("clarify_questions", [])) >= 3)

    # 7. Регулярный платёж — кредит каждый месяц
    d = pm("Плачу кредит каждый месяц 50000")
    check("7: 'плачу каждый месяц' → create_recurring", d.get("intent") == "create_recurring")
    check("7: frequency=monthly", d.get("frequency") == "monthly")

    # 8. Зарплата 900к
    d = pm("У меня зп 900к")
    check("8: 'зп 900к' → create_recurring", d.get("intent") == "create_recurring")
    check("8: amount=900000", d.get("transaction", {}).get("amount") == 900000)
    check("8: type=income", d.get("transaction", {}).get("type") == "income")
    check("8: зарплата сохранена", "зарплата" in d.get("response", "").lower())


def test_categorize(token):
    section("AI: categorize")

    # Одиночная категоризация
    resp, status = req("POST", SERVICES["ai"] + "/ai/categorize",
                       {"text": "McDonald's"}, token=token)
    check("POST /ai/categorize → 200", status == 200)
    check("category не пустой", bool(resp.get("data", {}).get("category")))

    # Batch
    resp2, status2 = req("POST", SERVICES["ai"] + "/ai/categorize/batch",
                         {"texts": ["такси", "продукты", "кофе"]}, token=token)
    check("POST /ai/categorize/batch → 200", status2 == 200)
    check("batch results=3", len(resp2.get("data", {}).get("results", [])) == 3)


def test_habits(token):
    section("Habits")

    resp, status = req("POST", SERVICES["habit"] + "/habits", {
        "title": "Утренняя зарядка", "period": "daily",
        "icon": "🏃", "color": "#FF5733"
    }, token=token)
    check("POST /habits → 201", status == 201)
    habit_id = resp.get("data", {}).get("id")
    check("habit has id", bool(habit_id))

    resp2, status2 = req("GET", SERVICES["habit"] + "/habits", token=token)
    check("GET /habits → 200", status2 == 200)

    # Выполнить привычку (toggle)
    resp3, status3 = req("POST", SERVICES["habit"] + f"/habits/{habit_id}/toggle",
                         {"date": "2026-04-25"}, token=token)
    check("POST /habits/:id/toggle → 200 or 201", status3 in (200, 201))

    _, status4 = req("DELETE", SERVICES["habit"] + f"/habits/{habit_id}", token=token)
    check("DELETE /habits/:id → 200", status4 == 200)


def test_goals(token):
    section("Goals")

    resp, status = req("POST", SERVICES["goal"] + "/goals", {
        "title": "Накопить на машину", "target_amount": 2000000,
        "deadline": "2027-01-01T00:00:00Z"
    }, token=token)
    check("POST /goals → 201", status == 201)
    goal_id = resp.get("data", {}).get("id")
    check("goal has id", bool(goal_id))

    resp2, status2 = req("GET", SERVICES["goal"] + "/goals", token=token)
    check("GET /goals → 200", status2 == 200)

    _, status3 = req("DELETE", SERVICES["goal"] + f"/goals/{goal_id}", token=token)
    check("DELETE /goals/:id → 200", status3 == 200)


def test_tasks(token):
    section("Tasks")

    resp, status = req("POST", SERVICES["task"] + "/tasks", {
        "title": "Составить бюджет", "priority": "high",
        "kind": "todo", "due_date": "2026-05-01"
    }, token=token)
    check("POST /tasks → 201", status == 201)
    task_id = resp.get("data", {}).get("id")
    check("task has id", bool(task_id))

    resp2, status2 = req("GET", SERVICES["task"] + "/tasks", token=token)
    check("GET /tasks → 200", status2 == 200)

    _, status3 = req("POST", SERVICES["task"] + f"/tasks/{task_id}/toggle",
                     token=token)
    check("POST /tasks/:id/toggle → 200 or 201", status3 in (200, 201))

    _, status4 = req("DELETE", SERVICES["task"] + f"/tasks/{task_id}", token=token)
    check("DELETE /tasks/:id → 200", status4 == 200)


def test_recurring(token):
    section("Recurring transactions")

    resp, status = req("POST", SERVICES["finance"] + "/recurring-transactions", {
        "title": "Аренда", "amount": 150000, "type": "expense",
        "category": "utilities", "frequency": "monthly", "start_date": "2026-04-01"
    }, token=token)
    check("POST /recurring-transactions → 201", status == 201)
    rec_id = resp.get("data", {}).get("id")
    check("recurring has id", bool(rec_id))

    resp2, status2 = req("GET", SERVICES["finance"] + "/recurring-transactions", token=token)
    check("GET /recurring-transactions → 200", status2 == 200)

    _, status3 = req("DELETE", SERVICES["finance"] + f"/recurring-transactions/{rec_id}", token=token)
    check("DELETE /recurring-transactions/:id → 200", status3 == 200)


def test_parse_message_tasks_habits(token):
    section("AI: parse-message — tasks & habits")

    AI = SERVICES["ai"]

    def pm(msg):
        resp, _ = req("POST", AI + "/ai/parse-message", {"message": msg}, token=token)
        return resp.get("data", {})

    # Task 1: с конкретным временем
    d = pm("Напомни мне позвонить маме в 19:00")
    check("task1: create_task", d.get("intent") == "create_task")
    check("task1: task_title=Позвонить маме", d.get("task_title") == "Позвонить маме")
    check("task1: task_time=19:00", d.get("task_time") == "19:00")
    check("task1: task_day=today", d.get("task_day") == "today")

    # Task 2: завтра утром
    d = pm("Добавь задачу сделать домашку завтра утром")
    check("task2: create_task", d.get("intent") == "create_task")
    check("task2: task_day=tomorrow", d.get("task_day") == "tomorrow")
    check("task2: task_time=09:00", d.get("task_time") == "09:00")

    # Task 3: выполнение
    d = pm("Я сделал тренировку")
    check("task3: complete_task", d.get("intent") == "complete_task")
    check("task3: keywords содержат тренировк", any("тренировк" in kw for kw in d.get("task_keywords", [])))

    # Habit 4: ежедневная
    d = pm("Хочу читать каждый день по 20 минут")
    check("habit4: create_habit", d.get("intent") == "create_habit")
    check("habit4: habit_title=Читать", d.get("habit_title") == "Читать")

    # Habit 5: 30-дневный челлендж
    d = pm("Хочу отжиматься 30 дней подряд")
    check("habit5: create_habit", d.get("intent") == "create_habit")
    check("habit5: duration=30", d.get("habit_duration_days") == 30)

    # Habit 6: архивировать
    d = pm("Я больше не хочу бегать по утрам")
    check("habit6: archive_habit", d.get("intent") == "archive_habit")
    check("habit6: habit_title=Бегать", d.get("habit_title") == "Бегать")

    # Habit 7: комплексный (бег 30 дней с временем)
    d = pm("Хочу бегать в 5 утра 30 дней подряд")
    check("habit7: create_habit", d.get("intent") == "create_habit")
    check("habit7: duration=30", d.get("habit_duration_days") == 30)
    check("habit7: habit_time=05:00", d.get("habit_time") == "05:00")


def test_savings_rules(token):
    section("Savings Rules & Auto-Savings")

    # Create spending_alert rule
    r, s = req("POST", SERVICES["finance"] + "/savings-rules",
               {"kind": "spending_alert", "amount": 9999}, token)
    check("POST /savings-rules spending_alert → 201", s == 201)
    alert_id = r.get("data", {}).get("id")
    check("spending_alert has id", bool(alert_id))

    # Create on_income_savings rule
    r2, s2 = req("POST", SERVICES["finance"] + "/savings-rules",
                 {"kind": "on_income_savings", "amount": 500}, token)
    check("POST /savings-rules on_income_savings → 201", s2 == 201)
    rule_id = r2.get("data", {}).get("id")
    check("on_income_savings has id", bool(rule_id))

    # Create monthly_savings rule
    r3, s3 = req("POST", SERVICES["finance"] + "/savings-rules",
                 {"kind": "monthly_savings", "amount": 2000}, token)
    check("POST /savings-rules monthly_savings → 201", s3 == 201)

    # List rules
    r4, s4 = req("GET", SERVICES["finance"] + "/savings-rules", token=token)
    rules = r4.get("data", [])
    check("GET /savings-rules → 200", s4 == 200)
    check("list contains ≥3 rules", len(rules) >= 3)

    # Daily spent
    r5, s5 = req("GET", SERVICES["finance"] + "/spending/daily", token=token)
    check("GET /spending/daily → 200", s5 == 200)
    check("daily_spent is number", isinstance(r5.get("data", {}).get("daily_spent"), (int, float)))

    # Auto-savings: create income tx → auto savings tx should appear
    import time as _time
    r6, _ = req("GET", SERVICES["finance"] + "/transactions", token=token)
    before_count = len(r6.get("data", []))

    req("POST", SERVICES["finance"] + "/transactions", {
        "title": "Зарплата e2e", "amount": 80000, "type": "income",
        "category": "salary", "date": "2026-04-25"
    }, token)
    _time.sleep(2)

    r7, _ = req("GET", SERVICES["finance"] + "/transactions", token=token)
    all_txs = r7.get("data", [])
    savings_txs = [t for t in all_txs if t.get("category") == "savings"]
    check("auto savings tx created on income", len(savings_txs) >= 1)

    # Deactivate rule
    r8, s8 = req("PATCH", SERVICES["finance"] + f"/savings-rules/{alert_id}/deactivate", token=token)
    check("PATCH /savings-rules/:id/deactivate → 200", s8 == 200)

    # Delete rule
    r9, s9 = req("DELETE", SERVICES["finance"] + f"/savings-rules/{rule_id}", token=token)
    check("DELETE /savings-rules/:id → 200", s9 == 200)


def test_insights(token):
    section("AI: Insights (summary + budget-suggest)")

    AI = SERVICES["ai"]

    txs = [
        # Апрель
        {"date": "2026-04-01", "amount": 8000,   "type": "expense", "category": "food"},
        {"date": "2026-04-05", "amount": 3500,   "type": "expense", "category": "cafe"},
        {"date": "2026-04-08", "amount": 1200,   "type": "expense", "category": "transport"},
        {"date": "2026-04-15", "amount": 150000, "type": "income",  "category": "salary"},
        {"date": "2026-04-22", "amount": 5000,   "type": "expense", "category": "cafe"},
        # Март
        {"date": "2026-03-02", "amount": 9000,   "type": "expense", "category": "food"},
        {"date": "2026-03-05", "amount": 4000,   "type": "expense", "category": "cafe"},
        {"date": "2026-03-15", "amount": 150000, "type": "income",  "category": "salary"},
        {"date": "2026-03-20", "amount": 12000,  "type": "expense", "category": "shopping"},
        # Февраль
        {"date": "2026-02-03", "amount": 7500,   "type": "expense", "category": "food"},
        {"date": "2026-02-14", "amount": 150000, "type": "income",  "category": "salary"},
        {"date": "2026-02-22", "amount": 20000,  "type": "expense", "category": "shopping"},
        {"date": "2026-02-28", "amount": 3000,   "type": "expense", "category": "entertainment"},
    ]

    # Summary
    r, s = req("POST", AI + "/ai/insights/summary", {
        "transactions": txs,
        "period_start": "2026-04-01",
        "period_end": "2026-04-30",
    }, token)
    d = r.get("data", r)
    check("POST /ai/insights/summary → 200", s == 200)
    check("summary: total_income=150000", d.get("total_income") == 150000)
    check("summary: total_expense > 0", d.get("total_expense", 0) > 0)
    check("summary: savings_rate > 0", d.get("savings_rate", 0) > 0)
    check("summary: top_categories не пустой", len(d.get("top_categories", [])) > 0)
    check("summary: expense_trend задан", d.get("expense_trend") in ("up", "down", "stable", "unknown"))

    # Budget suggestions
    r2, s2 = req("POST", AI + "/ai/insights/budget-suggest", {
        "transactions": txs,
        "lookback_days": 90,
        "percentile": 75,
    }, token)
    d2 = r2.get("data", r2)
    check("POST /ai/insights/budget-suggest → 200", s2 == 200)
    check("budget-suggest: suggestions не пустой", len(d2.get("suggestions", [])) > 0)
    first = d2.get("suggestions", [{}])[0]
    check("budget-suggest: suggested_limit > 0", first.get("suggested_limit", 0) > 0)
    check("budget-suggest: priority задан", first.get("priority") in ("high", "medium", "low"))
    check("budget-suggest: reason не пустой", bool(first.get("reason")))


def test_ai_local_direct():
    section("AI-local: прямые вызовы")

    # Health
    resp, status = req("GET", SERVICES["ai-local"] + "/health")
    check("ai-local /health → 200", status == 200)

    # Categorize
    resp2, status2 = req("POST", SERVICES["ai-local"] + "/categorize", {"text": "такси"})
    check("ai-local /categorize → 200", status2 == 200)
    check("ai-local category=transport", resp2.get("category") == "transport")

    # Parse-message прямо
    resp3, status3 = req("POST", SERVICES["ai-local"] + "/parse-message", {"message": "кофе 500"})
    check("ai-local /parse-message → 200", status3 == 200)
    check("ai-local parse intent=create_transaction", resp3.get("intent") == "create_transaction")


# ── Main ──────────────────────────────────────────────────────────────────────

if __name__ == "__main__":
    print("\n" + "═"*55)
    print("  AIFA e2e Test Suite")
    print("═"*55)

    print("\n⏳ Получаем токен...")
    try:
        token = get_token()
        print(f"  ✓  Token получен ({token[:20]}...)\n")
    except Exception as e:
        print(f"  ✗  Не удалось получить токен: {e}")
        sys.exit(1)

    test_health()
    test_auth(token)
    test_transactions(token)
    test_budgets(token)
    test_debts(token)
    test_parse_message(token)
    test_categorize(token)
    test_habits(token)
    test_goals(token)
    test_tasks(token)
    test_recurring(token)
    test_parse_message_tasks_habits(token)
    test_savings_rules(token)
    test_insights(token)
    test_ai_local_direct()

    total = PASS + FAIL
    print(f"\n{'═'*55}")
    print(f"  Результат: {PASS}/{total} passed", end="")
    if FAIL:
        print(f"  |  {FAIL} failed ✗")
    else:
        print("  ✓  all passed")
    print("═"*55 + "\n")

    sys.exit(0 if FAIL == 0 else 1)

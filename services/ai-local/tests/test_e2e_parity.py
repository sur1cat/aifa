"""
Зеркало e2e_test.py → test_parse_message + test_parse_message_tasks_habits.
Запуск: python3 -m pytest tests/test_e2e_parity.py -v
Не требует Docker — тестирует message_parser напрямую.
"""
import sys
import unittest
from datetime import date, timedelta
from types import SimpleNamespace

sys.modules.setdefault(
    "app.model",
    SimpleNamespace(
        get_classifier=lambda: SimpleNamespace(
            predict=lambda _text: SimpleNamespace(category="shopping")
        )
    ),
)

from app import message_parser


def _pm(msg, debts_context=None):
    return message_parser.parse_message(msg, debts_context=debts_context)


class TestParseMessageE2EParity(unittest.TestCase):
    """Повторяет test_parse_message из e2e_test.py."""

    # ── Транзакции ────────────────────────────────────────────────────────────

    def test_1_obid_7000_expense_food(self):
        r = _pm("обед 7000")
        self.assertEqual(r.intent, "create_transaction")
        self.assertEqual(r.tx_type, "expense")
        self.assertEqual(r.amount, 7000.0)
        self.assertEqual(r.category, "food")

    def test_2_sister_3k_income(self):
        r = _pm("Сестра закинула 3к")
        self.assertEqual(r.intent, "create_transaction")
        self.assertEqual(r.amount, 3000.0)
        self.assertEqual(r.tx_type, "income")
        self.assertEqual(r.counterparty, "Сестра")

    def test_3_nurs_they_owe(self):
        r = _pm("Нурс должен мне 3к")
        self.assertEqual(r.intent, "create_debt")
        self.assertEqual(r.debt_direction, "they_owe")
        self.assertEqual(r.counterparty, "Нурс")
        self.assertIsNone(r.tx_type)          # нет транзакции
        self.assertIn("3,000", r.response)

    def test_4_kim_i_owe(self):
        r = _pm("Я взял в долг у Кима 5000")
        self.assertEqual(r.intent, "create_debt")
        self.assertEqual(r.debt_direction, "i_owe")
        self.assertEqual(r.counterparty, "Кима")

    def test_5_kim_returned_half(self):
        ctx = [{"id": "debt-abc", "counterparty": "Ким", "amount": 10000}]
        r = _pm("Ким вернул половину долга", debts_context=ctx)
        self.assertEqual(r.intent, "update_debt")
        self.assertEqual(r.debt_update["type"], "half")
        self.assertEqual(r.debt_update["reduce_by"], 5000.0)
        self.assertEqual(r.debt_update["debt_id"], "debt-abc")

    def test_6_credit_clarify(self):
        r = _pm("Взял кредит в Kaspi 50000 в месяц")
        self.assertEqual(r.intent, "ask_clarify")
        self.assertGreaterEqual(len(r.clarify_questions), 3)

    def test_7_recurring_monthly(self):
        r = _pm("Плачу кредит каждый месяц 50000")
        self.assertEqual(r.intent, "create_recurring")
        self.assertEqual(r.frequency, "monthly")

    def test_8_salary_900k(self):
        r = _pm("У меня зп 900к")
        self.assertEqual(r.intent, "create_recurring")
        self.assertEqual(r.amount, 900000.0)
        self.assertEqual(r.tx_type, "income")
        self.assertIn("зарплата", r.response.lower())

    # ── Относительные даты ────────────────────────────────────────────────────

    def test_9_date_yesterday(self):
        yesterday = (date.today() - timedelta(days=1)).isoformat()
        r = _pm("купил кофе вчера за 800 тг")
        self.assertEqual(r.tx_date, yesterday)

    def test_10_date_day_before(self):
        day_before = (date.today() - timedelta(days=2)).isoformat()
        r = _pm("потратил 500 тг позавчера на такси")
        self.assertEqual(r.tx_date, day_before)

    def test_11_date_today_default(self):
        today = date.today().isoformat()
        r = _pm("получил зарплату 200000 тг")
        self.assertEqual(r.tx_date, today)


class TestParseMessageTasksHabitsE2EParity(unittest.TestCase):
    """Повторяет test_parse_message_tasks_habits из e2e_test.py."""

    # ── Задачи ────────────────────────────────────────────────────────────────

    def test_task1_remind_call_mom_at_19(self):
        r = _pm("Напомни мне позвонить маме в 19:00")
        self.assertEqual(r.intent, "create_task")
        self.assertEqual(r.task_title, "Позвонить маме")
        self.assertEqual(r.task_time, "19:00")
        self.assertEqual(r.task_day, "today")

    def test_task2_homework_tomorrow_morning(self):
        r = _pm("Добавь задачу сделать домашку завтра утром")
        self.assertEqual(r.intent, "create_task")
        self.assertEqual(r.task_day, "tomorrow")
        self.assertEqual(r.task_time, "09:00")

    def test_task3_complete_workout(self):
        r = _pm("Я сделал тренировку")
        self.assertEqual(r.intent, "complete_task")
        self.assertTrue(any("тренировк" in kw for kw in r.task_keywords))

    # ── Привычки ─────────────────────────────────────────────────────────────

    def test_habit4_read_daily(self):
        r = _pm("Хочу читать каждый день по 20 минут")
        self.assertEqual(r.intent, "create_habit")
        self.assertEqual(r.habit_title, "Читать")

    def test_habit5_pushups_30_days(self):
        r = _pm("Хочу отжиматься 30 дней подряд")
        self.assertEqual(r.intent, "create_habit")
        self.assertEqual(r.habit_duration_days, 30)

    def test_habit6_archive_running(self):
        r = _pm("Я больше не хочу бегать по утрам")
        self.assertEqual(r.intent, "archive_habit")
        self.assertEqual(r.habit_title, "Бегать")

    def test_habit7_run_5am_30days(self):
        r = _pm("Хочу бегать в 5 утра 30 дней подряд")
        self.assertEqual(r.intent, "create_habit")
        self.assertEqual(r.habit_duration_days, 30)
        self.assertEqual(r.habit_time, "05:00")


if __name__ == "__main__":
    unittest.main()

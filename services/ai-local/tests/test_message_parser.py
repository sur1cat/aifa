import sys
import unittest
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


class MessageParserUnitTests(unittest.TestCase):
    def test_extract_amount_handles_suffix_and_ignores_time_units(self) -> None:
        self.assertEqual(message_parser._extract_amount("потратил 12.5к тг"), 12500.0)
        self.assertIsNone(message_parser._extract_amount("напомни через 20 минут в 19:00"))

    def test_detect_transaction_intent_and_fields(self) -> None:
        result = message_parser.parse_message("купил кофе за 1500 тг")

        self.assertEqual(result.intent, "create_transaction")
        self.assertEqual(result.tx_type, "expense")
        self.assertEqual(result.amount, 1500.0)
        self.assertEqual(result.category, "cafe")

    def test_detect_income_transaction(self) -> None:
        result = message_parser.parse_message("получил зарплату 350000 тг")

        self.assertEqual(result.intent, "create_transaction")
        self.assertEqual(result.tx_type, "income")
        self.assertEqual(result.amount, 350000.0)

    def test_detect_create_debt(self) -> None:
        result = message_parser.parse_message("я должен Алишеру 20000 тг")

        self.assertEqual(result.intent, "create_debt")
        self.assertEqual(result.debt_direction, "i_owe")
        self.assertEqual(result.amount, 20000.0)
        self.assertEqual(result.counterparty, "Алишеру")

    def test_detect_update_debt_with_context(self) -> None:
        debts_context = [{"id": "d1", "counterparty": "Алишер", "amount": 20000}]

        result = message_parser.parse_message("Алишер вернул 5000 тг", debts_context=debts_context)

        self.assertEqual(result.intent, "update_debt")
        self.assertEqual(result.debt_update["reduce_by"], 5000.0)
        self.assertEqual(result.debt_update["debt_id"], "d1")

    def test_detect_create_recurring(self) -> None:
        result = message_parser.parse_message("зарплата 300000 тг каждый месяц")

        self.assertEqual(result.intent, "create_recurring")
        self.assertEqual(result.tx_type, "income")
        self.assertEqual(result.frequency, "monthly")

    def test_detect_create_task(self) -> None:
        result = message_parser.parse_message("напомни завтра в 19:30 купить подарок")

        self.assertEqual(result.intent, "create_task")
        self.assertEqual(result.task_day, "tomorrow")
        self.assertEqual(result.task_time, "19:30")
        self.assertIn("Купить", result.task_title)

    def test_detect_complete_task(self) -> None:
        result = message_parser.parse_message("я сделал тренировку")

        self.assertEqual(result.intent, "complete_task")
        self.assertIn("тренировк", "".join(result.task_keywords))

    def test_detect_create_habit(self) -> None:
        result = message_parser.parse_message("хочу читать каждый день 30 дней")

        self.assertEqual(result.intent, "create_habit")
        self.assertEqual(result.habit_duration_days, 30)
        self.assertEqual(result.frequency, "daily")

    def test_detect_archive_habit(self) -> None:
        result = message_parser.parse_message("больше не хочу бегать")

        self.assertEqual(result.intent, "archive_habit")
        self.assertIn("Бегать", result.habit_title)

    def test_detect_savings_plan_rule_and_alert(self) -> None:
        plan = message_parser.parse_message("хочу копить на отпуск 50000 тг каждый месяц")
        rule = message_parser.parse_message("с каждого дохода откладывать 10000 тг")
        alert = message_parser.parse_message("предупреждай если трачу больше 15000 тг")

        self.assertEqual(plan.intent, "create_savings_plan")
        self.assertEqual(plan.savings_amount, 50000.0)
        self.assertEqual(rule.intent, "create_savings_rule")
        self.assertEqual(rule.savings_period, "on_income")
        self.assertEqual(alert.intent, "create_spending_alert")
        self.assertEqual(alert.alert_limit, 15000.0)

    def test_detect_credit_clarify_and_chat(self) -> None:
        clarify = message_parser.parse_message("кредит 85000 тг")
        chat = message_parser.parse_message("что нового")

        self.assertEqual(clarify.intent, "ask_clarify")
        self.assertTrue(clarify.clarify_questions)
        self.assertEqual(chat.intent, "chat")


if __name__ == "__main__":
    unittest.main()

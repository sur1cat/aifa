package com.atoma.app.domain.model

import com.atoma.app.R
import java.time.LocalDate
import java.time.LocalDateTime
import java.util.UUID

// MARK: - User
data class User(
    val id: String,
    val email: String,
    val name: String,
    val avatarUrl: String?,
    val isPremium: Boolean = false
)

// MARK: - Habit
enum class HabitPeriod {
    DAILY, WEEKLY, MONTHLY;

    val title: String
        get() = when (this) {
            DAILY -> "Daily"
            WEEKLY -> "Weekly"
            MONTHLY -> "Monthly"
        }

    val shortTitle: String
        get() = when (this) {
            DAILY -> "day"
            WEEKLY -> "week"
            MONTHLY -> "month"
        }
}

data class Habit(
    val id: String = UUID.randomUUID().toString(),
    val title: String,
    val icon: String = "circle.fill",
    val color: String = "green",
    val period: HabitPeriod = HabitPeriod.DAILY,
    val completedDates: List<String> = emptyList(),
    val createdAt: LocalDateTime = LocalDateTime.now(),
    val reminderEnabled: Boolean = false,
    val reminderTime: LocalDateTime? = null,
    val archivedAt: LocalDateTime? = null,
    val goalId: String? = null
) {
    val isCompletedToday: Boolean
        get() = completedDates.contains(LocalDate.now().toString())

    val isActive: Boolean
        get() = archivedAt == null

    val streak: Int
        get() {
            if (completedDates.isEmpty()) return 0
            val sorted = completedDates.sorted().reversed()
            var count = 0
            var checkDate = LocalDate.now()

            for (dateStr in sorted) {
                val date = LocalDate.parse(dateStr)
                if (date == checkDate || date == checkDate.minusDays(1)) {
                    count++
                    checkDate = date
                } else {
                    break
                }
            }
            return count
        }
}

// MARK: - Task
enum class TaskPriority {
    LOW, MEDIUM, HIGH, URGENT;

    val title: String
        get() = when (this) {
            LOW -> "Low"
            MEDIUM -> "Medium"
            HIGH -> "High"
            URGENT -> "Urgent"
        }
}

data class DailyTask(
    val id: String = UUID.randomUUID().toString(),
    val title: String,
    val isCompleted: Boolean = false,
    val priority: TaskPriority = TaskPriority.MEDIUM,
    val dueDate: LocalDate = LocalDate.now(),
    val createdAt: LocalDateTime = LocalDateTime.now()
)

// MARK: - Transaction
enum class TransactionType {
    INCOME, EXPENSE
}

enum class TransactionCategory(
    val key: String,
    val titleRes: Int,
    val icon: String
) {
    FOOD("food", R.string.category_food, "restaurant"),
    TRANSPORT("transport", R.string.category_transport, "directions_car"),
    SHOPPING("shopping", R.string.category_shopping, "shopping_bag"),
    ENTERTAINMENT("entertainment", R.string.category_entertainment, "movie"),
    HEALTH("health", R.string.category_health, "favorite"),
    EDUCATION("education", R.string.category_education, "school"),
    BILLS("bills", R.string.category_bills, "receipt_long"),
    SALARY("salary", R.string.category_salary, "payments"),
    FREELANCE("freelance", R.string.category_freelance, "laptop"),
    INVESTMENT("investment", R.string.category_investment, "trending_up"),
    GIFT("gift", R.string.category_gift, "redeem"),
    OTHER("other", R.string.category_other, "more_horiz");

    companion object {
        fun fromKey(key: String): TransactionCategory =
            entries.find { it.key == key } ?: OTHER

        val expenseCategories: List<TransactionCategory>
            get() = listOf(FOOD, TRANSPORT, SHOPPING, ENTERTAINMENT, HEALTH, EDUCATION, BILLS, GIFT, OTHER)

        val incomeCategories: List<TransactionCategory>
            get() = listOf(SALARY, FREELANCE, INVESTMENT, GIFT, OTHER)
    }
}

data class Transaction(
    val id: String = UUID.randomUUID().toString(),
    val title: String,
    val amount: Double,
    val type: TransactionType,
    val category: String = TransactionCategory.OTHER.key,
    val date: LocalDate = LocalDate.now()
) {
    val categoryEnum: TransactionCategory
        get() = TransactionCategory.fromKey(category)
}

// MARK: - Budget Summary
data class BudgetSummary(
    val income: Double,
    val expenses: Double
) {
    val balance: Double get() = income - expenses
}

// MARK: - Recurring Transaction
enum class RecurrenceFrequency {
    WEEKLY, BIWEEKLY, MONTHLY, QUARTERLY, YEARLY;

    val title: String
        get() = when (this) {
            WEEKLY -> "Weekly"
            BIWEEKLY -> "Biweekly"
            MONTHLY -> "Monthly"
            QUARTERLY -> "Quarterly"
            YEARLY -> "Yearly"
        }

    val apiValue: String
        get() = name.lowercase()

    companion object {
        fun fromApi(value: String): RecurrenceFrequency =
            entries.find { it.name.equals(value, ignoreCase = true) } ?: MONTHLY
    }
}

enum class RecurringCategory(val key: String, val icon: String) {
    SUBSCRIPTIONS("subscriptions", "sync"),
    RENT("rent", "home"),
    UTILITIES("utilities", "power"),
    INSURANCE("insurance", "shield"),
    SALARY("salary", "payments"),
    LOAN("loan", "account_balance"),
    OTHER("other", "more_horiz");

    companion object {
        fun fromKey(key: String): RecurringCategory =
            entries.find { it.key == key } ?: OTHER
    }
}

data class RecurringTransaction(
    val id: String = UUID.randomUUID().toString(),
    val title: String,
    val amount: Double,
    val type: TransactionType,
    val category: RecurringCategory = RecurringCategory.SUBSCRIPTIONS,
    val frequency: RecurrenceFrequency = RecurrenceFrequency.MONTHLY,
    val startDate: LocalDate = LocalDate.now(),
    val nextDate: LocalDate = LocalDate.now(),
    val endDate: LocalDate? = null,
    val isActive: Boolean = true
)

data class RecurringProjection(
    val date: LocalDate,
    val amount: Double,
    val type: TransactionType,
    val recurringId: String,
    val title: String
)

// MARK: - Goal
data class Goal(
    val id: String = UUID.randomUUID().toString(),
    val title: String,
    val icon: String = "🎯",
    val targetValue: Int? = null,
    val unit: String? = null,
    val deadline: LocalDate? = null,
    val createdAt: LocalDateTime = LocalDateTime.now(),
    val archivedAt: LocalDateTime? = null
) {
    val isActive: Boolean get() = archivedAt == null
}

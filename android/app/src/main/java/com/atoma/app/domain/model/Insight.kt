package com.atoma.app.domain.model

import java.util.UUID

enum class InsightSection {
    HABITS, TASKS, BUDGET;

    val title: String
        get() = when (this) {
            HABITS -> "Habits"
            TASKS -> "Tasks"
            BUDGET -> "Budget"
        }

    val icon: String
        get() = when (this) {
            HABITS -> "repeat"
            TASKS -> "check_circle"
            BUDGET -> "credit_card"
        }

    val requiredDays: Int
        get() = when (this) {
            HABITS -> 14
            TASKS -> 7
            BUDGET -> 30
        }
}

enum class InsightType {
    POSITIVE, WARNING, INFO, ACHIEVEMENT
}

data class Insight(
    val id: String = UUID.randomUUID().toString(),
    val section: InsightSection,
    val type: InsightType,
    val title: String,
    val message: String,
    val emoji: String = "",
    val isDismissed: Boolean = false
)

// Achievement/Badge system
data class Achievement(
    val id: String,
    val title: String,
    val description: String,
    val icon: String,
    val isUnlocked: Boolean,
    val unlockedAt: Long? = null,
    val progress: Float = 0f, // 0.0 to 1.0
    val requirement: Int = 0,
    val currentValue: Int = 0
)

object AchievementDefinitions {
    val habitAchievements = listOf(
        Achievement(
            id = "first_habit",
            title = "First Step",
            description = "Create your first habit",
            icon = "🌱",
            isUnlocked = false,
            requirement = 1
        ),
        Achievement(
            id = "habit_streak_7",
            title = "Week Warrior",
            description = "Maintain a 7-day streak",
            icon = "🔥",
            isUnlocked = false,
            requirement = 7
        ),
        Achievement(
            id = "habit_streak_30",
            title = "Monthly Master",
            description = "Maintain a 30-day streak",
            icon = "⭐",
            isUnlocked = false,
            requirement = 30
        ),
        Achievement(
            id = "habits_5",
            title = "Habit Builder",
            description = "Create 5 habits",
            icon = "🏗️",
            isUnlocked = false,
            requirement = 5
        ),
        Achievement(
            id = "perfect_day",
            title = "Perfect Day",
            description = "Complete all habits in a day",
            icon = "✨",
            isUnlocked = false,
            requirement = 1
        )
    )

    val taskAchievements = listOf(
        Achievement(
            id = "first_task",
            title = "Getting Started",
            description = "Complete your first task",
            icon = "✅",
            isUnlocked = false,
            requirement = 1
        ),
        Achievement(
            id = "tasks_10",
            title = "Task Crusher",
            description = "Complete 10 tasks",
            icon = "💪",
            isUnlocked = false,
            requirement = 10
        ),
        Achievement(
            id = "tasks_50",
            title = "Productivity Pro",
            description = "Complete 50 tasks",
            icon = "🚀",
            isUnlocked = false,
            requirement = 50
        ),
        Achievement(
            id = "high_priority",
            title = "Priority Master",
            description = "Complete 5 high-priority tasks",
            icon = "🎯",
            isUnlocked = false,
            requirement = 5
        )
    )

    val budgetAchievements = listOf(
        Achievement(
            id = "first_transaction",
            title = "Money Tracker",
            description = "Log your first transaction",
            icon = "💰",
            isUnlocked = false,
            requirement = 1
        ),
        Achievement(
            id = "positive_balance",
            title = "In the Green",
            description = "End the month with positive balance",
            icon = "📈",
            isUnlocked = false,
            requirement = 1
        ),
        Achievement(
            id = "transactions_30",
            title = "Budget Keeper",
            description = "Log 30 transactions",
            icon = "📊",
            isUnlocked = false,
            requirement = 30
        ),
        Achievement(
            id = "saver",
            title = "Super Saver",
            description = "Save more than 20% of income",
            icon = "🏦",
            isUnlocked = false,
            requirement = 20
        )
    )
}

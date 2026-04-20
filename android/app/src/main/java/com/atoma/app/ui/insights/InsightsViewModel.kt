package com.atoma.app.ui.insights

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.atoma.app.data.repository.HabitsRepository
import com.atoma.app.data.repository.TasksRepository
import com.atoma.app.data.repository.BudgetRepository
import com.atoma.app.domain.model.*
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import java.time.LocalDate
import javax.inject.Inject

data class InsightsUiState(
    val habits: List<Habit> = emptyList(),
    val tasks: List<DailyTask> = emptyList(),
    val transactions: List<Transaction> = emptyList(),
    val achievements: List<Achievement> = emptyList(),
    val insights: List<Insight> = emptyList(),
    val weeklyHabitsCompleted: Int = 0,
    val weeklyTasksCompleted: Int = 0,
    val weeklySavingsRate: Double = 0.0,
    val isLoading: Boolean = false,
    val error: String? = null
) {
    val unlockedAchievements: List<Achievement>
        get() = achievements.filter { it.isUnlocked }

    val lockedAchievements: List<Achievement>
        get() = achievements.filter { !it.isUnlocked }

    val habitAchievements: List<Achievement>
        get() = achievements.filter { it.id.startsWith("habit") || it.id == "first_habit" || it.id == "perfect_day" }

    val taskAchievements: List<Achievement>
        get() = achievements.filter { it.id.startsWith("task") || it.id == "first_task" || it.id == "high_priority" }

    val budgetAchievements: List<Achievement>
        get() = achievements.filter { it.id.startsWith("transaction") || it.id == "first_transaction" || it.id == "positive_balance" || it.id == "saver" }
}

@HiltViewModel
class InsightsViewModel @Inject constructor(
    private val habitsRepository: HabitsRepository,
    private val tasksRepository: TasksRepository,
    private val budgetRepository: BudgetRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(InsightsUiState())
    val uiState: StateFlow<InsightsUiState> = _uiState

    init {
        loadData()
    }

    fun loadData() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }

            // Load habits
            habitsRepository.getHabits()
                .onSuccess { habits ->
                    _uiState.update { it.copy(habits = habits) }
                }

            // Load tasks
            tasksRepository.getTasks()
                .onSuccess { tasks ->
                    _uiState.update { it.copy(tasks = tasks) }
                }

            // Load transactions
            budgetRepository.getTransactions()
                .onSuccess { transactions ->
                    _uiState.update { it.copy(transactions = transactions) }
                }

            // Calculate achievements, insights, and weekly stats
            val state = _uiState.value
            val achievements = calculateAchievements(state.habits, state.tasks, state.transactions)
            val insights = generateInsights(state.habits, state.tasks, state.transactions)
            val weeklyStats = calculateWeeklyStats(state.habits, state.tasks, state.transactions)

            _uiState.update {
                it.copy(
                    achievements = achievements,
                    insights = insights,
                    weeklyHabitsCompleted = weeklyStats.habitsCompleted,
                    weeklyTasksCompleted = weeklyStats.tasksCompleted,
                    weeklySavingsRate = weeklyStats.savingsRate,
                    isLoading = false
                )
            }
        }
    }

    private fun calculateAchievements(
        habits: List<Habit>,
        tasks: List<DailyTask>,
        transactions: List<Transaction>
    ): List<Achievement> {
        val today = LocalDate.now()
        val todayString = today.toString()
        val completedTasks = tasks.filter { it.isCompleted }
        val monthTransactions = transactions.filter {
            it.date.year == today.year && it.date.monthValue == today.monthValue
        }
        val monthIncome = monthTransactions.filter { it.type == TransactionType.INCOME }.sumOf { it.amount }
        val monthExpenses = monthTransactions.filter { it.type == TransactionType.EXPENSE }.sumOf { it.amount }

        // Calculate habit achievements
        val habitAchievements = AchievementDefinitions.habitAchievements.map { achievement ->
            when (achievement.id) {
                "first_habit" -> {
                    val count = habits.size
                    achievement.copy(
                        isUnlocked = count >= 1,
                        currentValue = count,
                        progress = minOf(1f, count.toFloat() / 1)
                    )
                }
                "habit_streak_7" -> {
                    val maxStreak = habits.maxOfOrNull { it.streak } ?: 0
                    achievement.copy(
                        isUnlocked = maxStreak >= 7,
                        currentValue = maxStreak,
                        progress = minOf(1f, maxStreak.toFloat() / 7)
                    )
                }
                "habit_streak_30" -> {
                    val maxStreak = habits.maxOfOrNull { it.streak } ?: 0
                    achievement.copy(
                        isUnlocked = maxStreak >= 30,
                        currentValue = maxStreak,
                        progress = minOf(1f, maxStreak.toFloat() / 30)
                    )
                }
                "habits_5" -> {
                    val count = habits.size
                    achievement.copy(
                        isUnlocked = count >= 5,
                        currentValue = count,
                        progress = minOf(1f, count.toFloat() / 5)
                    )
                }
                "perfect_day" -> {
                    val habitsToday = habits.filter { it.createdAt.toLocalDate() <= today }
                    val completedToday = habitsToday.count { it.completedDates.contains(todayString) }
                    val isPerfect = habitsToday.isNotEmpty() && completedToday == habitsToday.size
                    achievement.copy(
                        isUnlocked = isPerfect,
                        currentValue = if (isPerfect) 1 else 0,
                        progress = if (habitsToday.isEmpty()) 0f else completedToday.toFloat() / habitsToday.size
                    )
                }
                else -> achievement
            }
        }

        // Calculate task achievements
        val taskAchievements = AchievementDefinitions.taskAchievements.map { achievement ->
            when (achievement.id) {
                "first_task" -> {
                    val count = completedTasks.size
                    achievement.copy(
                        isUnlocked = count >= 1,
                        currentValue = count,
                        progress = minOf(1f, count.toFloat() / 1)
                    )
                }
                "tasks_10" -> {
                    val count = completedTasks.size
                    achievement.copy(
                        isUnlocked = count >= 10,
                        currentValue = count,
                        progress = minOf(1f, count.toFloat() / 10)
                    )
                }
                "tasks_50" -> {
                    val count = completedTasks.size
                    achievement.copy(
                        isUnlocked = count >= 50,
                        currentValue = count,
                        progress = minOf(1f, count.toFloat() / 50)
                    )
                }
                "high_priority" -> {
                    val count = completedTasks.count { it.priority == TaskPriority.HIGH }
                    achievement.copy(
                        isUnlocked = count >= 5,
                        currentValue = count,
                        progress = minOf(1f, count.toFloat() / 5)
                    )
                }
                else -> achievement
            }
        }

        // Calculate budget achievements
        val budgetAchievements = AchievementDefinitions.budgetAchievements.map { achievement ->
            when (achievement.id) {
                "first_transaction" -> {
                    val count = transactions.size
                    achievement.copy(
                        isUnlocked = count >= 1,
                        currentValue = count,
                        progress = minOf(1f, count.toFloat() / 1)
                    )
                }
                "positive_balance" -> {
                    val balance = monthIncome - monthExpenses
                    achievement.copy(
                        isUnlocked = balance > 0 && monthIncome > 0,
                        currentValue = if (balance > 0 && monthIncome > 0) 1 else 0,
                        progress = if (balance > 0 && monthIncome > 0) 1f else 0f
                    )
                }
                "transactions_30" -> {
                    val count = transactions.size
                    achievement.copy(
                        isUnlocked = count >= 30,
                        currentValue = count,
                        progress = minOf(1f, count.toFloat() / 30)
                    )
                }
                "saver" -> {
                    val savingsRate = if (monthIncome > 0) ((monthIncome - monthExpenses) / monthIncome * 100).toInt() else 0
                    achievement.copy(
                        isUnlocked = savingsRate >= 20,
                        currentValue = maxOf(0, savingsRate),
                        progress = minOf(1f, maxOf(0f, savingsRate.toFloat() / 20))
                    )
                }
                else -> achievement
            }
        }

        return habitAchievements + taskAchievements + budgetAchievements
    }

    private fun generateInsights(
        habits: List<Habit>,
        tasks: List<DailyTask>,
        transactions: List<Transaction>
    ): List<Insight> {
        val insights = mutableListOf<Insight>()
        val today = LocalDate.now()
        val todayString = today.toString()

        // Habit insights
        val maxStreak = habits.maxOfOrNull { it.streak } ?: 0
        if (maxStreak >= 7) {
            val topHabit = habits.maxByOrNull { it.streak }
            insights.add(
                Insight(
                    section = InsightSection.HABITS,
                    type = InsightType.POSITIVE,
                    title = "Great Streak!",
                    message = "${topHabit?.icon} ${topHabit?.title} is on a ${maxStreak}-day streak. Keep it up!",
                    emoji = "🔥"
                )
            )
        }

        val habitsToday = habits.filter { it.createdAt.toLocalDate() <= today }
        val completedToday = habitsToday.count { it.completedDates.contains(todayString) }
        if (habitsToday.isNotEmpty() && completedToday < habitsToday.size) {
            val remaining = habitsToday.size - completedToday
            insights.add(
                Insight(
                    section = InsightSection.HABITS,
                    type = InsightType.INFO,
                    title = "Daily Reminder",
                    message = "You have $remaining habit${if (remaining > 1) "s" else ""} left to complete today.",
                    emoji = "⏰"
                )
            )
        }

        // Task insights
        val pendingHighPriority = tasks.filter { !it.isCompleted && it.priority == TaskPriority.HIGH }
        if (pendingHighPriority.isNotEmpty()) {
            insights.add(
                Insight(
                    section = InsightSection.TASKS,
                    type = InsightType.WARNING,
                    title = "High Priority",
                    message = "${pendingHighPriority.size} high-priority task${if (pendingHighPriority.size > 1) "s" else ""} need${if (pendingHighPriority.size == 1) "s" else ""} your attention.",
                    emoji = "⚠️"
                )
            )
        }

        val completedTasks = tasks.filter { it.isCompleted }
        if (completedTasks.size >= 10) {
            insights.add(
                Insight(
                    section = InsightSection.TASKS,
                    type = InsightType.POSITIVE,
                    title = "Productive!",
                    message = "You've completed ${completedTasks.size} tasks. Great progress!",
                    emoji = "🎉"
                )
            )
        }

        // Budget insights
        val monthTransactions = transactions.filter {
            it.date.year == today.year && it.date.monthValue == today.monthValue
        }
        val monthIncome = monthTransactions.filter { it.type == TransactionType.INCOME }.sumOf { it.amount }
        val monthExpenses = monthTransactions.filter { it.type == TransactionType.EXPENSE }.sumOf { it.amount }

        if (monthIncome > 0) {
            val savingsRate = ((monthIncome - monthExpenses) / monthIncome * 100).toInt()
            if (savingsRate >= 20) {
                insights.add(
                    Insight(
                        section = InsightSection.BUDGET,
                        type = InsightType.POSITIVE,
                        title = "Savings Star",
                        message = "You're saving $savingsRate% of your income this month!",
                        emoji = "💰"
                    )
                )
            } else if (savingsRate < 0) {
                insights.add(
                    Insight(
                        section = InsightSection.BUDGET,
                        type = InsightType.WARNING,
                        title = "Budget Alert",
                        message = "You've spent more than you earned this month. Consider reviewing expenses.",
                        emoji = "📉"
                    )
                )
            }
        }

        // Find top spending category
        val expensesByCategory = monthTransactions
            .filter { it.type == TransactionType.EXPENSE }
            .groupBy { it.category }
            .mapValues { it.value.sumOf { tx -> tx.amount } }

        if (expensesByCategory.isNotEmpty()) {
            val topCategory = expensesByCategory.maxByOrNull { it.value }
            if (topCategory != null && topCategory.value > 0) {
                insights.add(
                    Insight(
                        section = InsightSection.BUDGET,
                        type = InsightType.INFO,
                        title = "Top Spending",
                        message = "Your biggest expense category is ${topCategory.key.replaceFirstChar { it.uppercase() }} ($${topCategory.value.toInt()}).",
                        emoji = "📊"
                    )
                )
            }
        }

        return insights
    }

    private data class WeeklyStats(
        val habitsCompleted: Int,
        val tasksCompleted: Int,
        val savingsRate: Double
    )

    private fun calculateWeeklyStats(
        habits: List<Habit>,
        tasks: List<DailyTask>,
        transactions: List<Transaction>
    ): WeeklyStats {
        val today = LocalDate.now()
        val weekStart = today.minusDays(6)

        // Count habit completions this week
        val weeklyHabitCompletions = habits.sumOf { habit ->
            habit.completedDates.count { dateStr ->
                try {
                    val date = LocalDate.parse(dateStr)
                    !date.isBefore(weekStart) && !date.isAfter(today)
                } catch (e: Exception) {
                    false
                }
            }
        }

        // Count completed tasks this week
        val weeklyTasksCompleted = tasks.count { task ->
            task.isCompleted && !task.dueDate.isBefore(weekStart) && !task.dueDate.isAfter(today)
        }

        // Calculate savings rate for this month
        val monthTransactions = transactions.filter {
            it.date.year == today.year && it.date.monthValue == today.monthValue
        }
        val monthIncome = monthTransactions.filter { it.type == TransactionType.INCOME }.sumOf { it.amount }
        val monthExpenses = monthTransactions.filter { it.type == TransactionType.EXPENSE }.sumOf { it.amount }
        val savingsRate = if (monthIncome > 0) {
            ((monthIncome - monthExpenses) / monthIncome * 100)
        } else {
            0.0
        }

        return WeeklyStats(
            habitsCompleted = weeklyHabitCompletions,
            tasksCompleted = weeklyTasksCompleted,
            savingsRate = maxOf(0.0, savingsRate)
        )
    }

    fun dismissInsight(insight: Insight) {
        _uiState.update { state ->
            state.copy(insights = state.insights.filter { it.id != insight.id })
        }
    }

    fun clearError() {
        _uiState.update { it.copy(error = null) }
    }
}

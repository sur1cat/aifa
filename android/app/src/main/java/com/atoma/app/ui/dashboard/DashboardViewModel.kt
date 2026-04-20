package com.atoma.app.ui.dashboard

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.atoma.app.data.repository.HabitsRepository
import com.atoma.app.data.repository.TasksRepository
import com.atoma.app.data.repository.BudgetRepository
import com.atoma.app.domain.model.Habit
import com.atoma.app.domain.model.DailyTask
import com.atoma.app.domain.model.Transaction
import com.atoma.app.domain.model.TransactionType
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import java.time.LocalDate
import java.time.LocalDateTime
import javax.inject.Inject

data class DashboardUiState(
    val habits: List<Habit> = emptyList(),
    val tasks: List<DailyTask> = emptyList(),
    val transactions: List<Transaction> = emptyList(),
    val isLoading: Boolean = false,
    val error: String? = null
) {
    private val today = LocalDate.now()
    private val todayString = today.toString()

    val activeHabits: List<Habit>
        get() = habits.filter { habit ->
            habit.createdAt.toLocalDate() <= today
        }

    val habitsCompletedToday: Int
        get() = activeHabits.count { it.completedDates.contains(todayString) }

    val habitsTotalToday: Int
        get() = activeHabits.size

    val habitsProgress: Double
        get() = if (habitsTotalToday > 0) habitsCompletedToday.toDouble() / habitsTotalToday else 0.0

    val pendingHabits: List<Habit>
        get() = activeHabits.filter { !it.completedDates.contains(todayString) }

    val tasksForToday: List<DailyTask>
        get() = tasks.filter { it.dueDate == today }

    val tasksCompletedToday: Int
        get() = tasksForToday.count { it.isCompleted }

    val tasksTotalToday: Int
        get() = tasksForToday.size

    val tasksProgress: Double
        get() = if (tasksTotalToday > 0) tasksCompletedToday.toDouble() / tasksTotalToday else 0.0

    val pendingTasks: List<DailyTask>
        get() = tasksForToday.filter { !it.isCompleted }.sortedBy { it.priority.ordinal }

    val monthTransactions: List<Transaction>
        get() = transactions.filter {
            it.date.year == today.year && it.date.monthValue == today.monthValue
        }

    val monthIncome: Double
        get() = monthTransactions.filter { it.type == TransactionType.INCOME }.sumOf { it.amount }

    val monthExpenses: Double
        get() = monthTransactions.filter { it.type == TransactionType.EXPENSE }.sumOf { it.amount }

    val monthBalance: Double
        get() = monthIncome - monthExpenses

    val topStreakHabits: List<Habit>
        get() = activeHabits.filter { it.streak > 0 }.sortedByDescending { it.streak }

    // Life Score calculation
    val habitsScore: Double
        get() = habitsProgress * 100

    val tasksScore: Double
        get() = tasksProgress * 100

    val budgetScore: Double
        get() {
            if (monthIncome == 0.0 && monthExpenses == 0.0) return 0.0
            if (monthIncome == 0.0) return maxOf(0.0, 100 - monthExpenses / 10)
            val savingsRate = (monthIncome - monthExpenses) / monthIncome
            val normalized = (savingsRate + 0.5) / 0.8
            return maxOf(0.0, minOf(100.0, normalized * 100))
        }

    val lifeScore: Double
        get() = (habitsScore * 0.40) + (tasksScore * 0.30) + (budgetScore * 0.30)

    // Greeting based on time of day
    val greeting: String
        get() {
            val hour = LocalDateTime.now().hour
            return when {
                hour < 12 -> "Good morning"
                hour < 17 -> "Good afternoon"
                else -> "Good evening"
            }
        }

    val statusMessage: String
        get() {
            val totalItems = habitsTotalToday + tasksTotalToday
            val completedItems = habitsCompletedToday + tasksCompletedToday
            return when {
                totalItems == 0 -> "No habits or tasks for today"
                completedItems == totalItems -> "All done! Great job today"
                else -> {
                    val remaining = totalItems - completedItems
                    "$remaining item${if (remaining == 1) "" else "s"} remaining"
                }
            }
        }

    val totalProgress: Double
        get() {
            val totalItems = habitsTotalToday + tasksTotalToday
            val completedItems = habitsCompletedToday + tasksCompletedToday
            return if (totalItems > 0) completedItems.toDouble() / totalItems else 0.0
        }
}

@HiltViewModel
class DashboardViewModel @Inject constructor(
    private val habitsRepository: HabitsRepository,
    private val tasksRepository: TasksRepository,
    private val budgetRepository: BudgetRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(DashboardUiState())
    val uiState: StateFlow<DashboardUiState> = _uiState

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
                .onFailure { e ->
                    _uiState.update { it.copy(error = e.message) }
                }

            // Load tasks
            tasksRepository.getTasks()
                .onSuccess { tasks ->
                    _uiState.update { it.copy(tasks = tasks) }
                }
                .onFailure { e ->
                    _uiState.update { it.copy(error = e.message) }
                }

            // Load transactions
            budgetRepository.getTransactions()
                .onSuccess { transactions ->
                    _uiState.update { it.copy(transactions = transactions) }
                }
                .onFailure { e ->
                    _uiState.update { it.copy(error = e.message) }
                }

            _uiState.update { it.copy(isLoading = false) }
        }
    }

    fun clearError() {
        _uiState.update { it.copy(error = null) }
    }
}

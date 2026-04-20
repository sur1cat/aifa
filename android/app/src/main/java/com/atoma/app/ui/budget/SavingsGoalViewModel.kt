package com.atoma.app.ui.budget

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.atoma.app.data.repository.BudgetRepository
import com.atoma.app.domain.model.SavingsGoal
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import java.time.YearMonth
import javax.inject.Inject

data class SavingsGoalUiState(
    val savingsGoal: SavingsGoal? = null,
    val monthlyIncome: Double = 0.0,
    val monthlyExpenses: Double = 0.0,
    val showEditSheet: Boolean = false,
    val isLoading: Boolean = false,
    val error: String? = null
) {
    val currentSavings: Double
        get() = (monthlyIncome - monthlyExpenses).coerceAtLeast(0.0)
}

@HiltViewModel
class SavingsGoalViewModel @Inject constructor(
    private val budgetRepository: BudgetRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(SavingsGoalUiState())
    val uiState: StateFlow<SavingsGoalUiState> = _uiState

    init {
        loadData()
    }

    fun loadData() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }

            val currentMonth = YearMonth.now()

            // Get summary for income/expenses
            budgetRepository.getSummary(currentMonth)
                .onSuccess { summary ->
                    _uiState.update {
                        it.copy(
                            monthlyIncome = summary.income,
                            monthlyExpenses = summary.expenses
                        )
                    }
                }

            // Get savings goal
            budgetRepository.getSavingsGoal()
                .onSuccess { goal ->
                    // Update goal with current savings
                    val updatedGoal = goal?.copy(
                        currentSavings = _uiState.value.currentSavings
                    )
                    _uiState.update { it.copy(savingsGoal = updatedGoal, isLoading = false) }
                }
                .onFailure { e ->
                    _uiState.update { it.copy(error = e.message, isLoading = false) }
                }
        }
    }

    fun saveGoal(monthlyTarget: Double) {
        viewModelScope.launch {
            val currentState = _uiState.value
            val newGoal = SavingsGoal(
                id = currentState.savingsGoal?.id ?: java.util.UUID.randomUUID().toString(),
                monthlyTarget = monthlyTarget,
                currentSavings = currentState.currentSavings
            )

            // Optimistic update
            _uiState.update { it.copy(savingsGoal = newGoal, showEditSheet = false) }

            budgetRepository.saveSavingsGoal(monthlyTarget)
                .onFailure { e ->
                    // Rollback
                    _uiState.update { it.copy(savingsGoal = currentState.savingsGoal, error = e.message) }
                }
        }
    }

    fun deleteGoal() {
        viewModelScope.launch {
            val previousGoal = _uiState.value.savingsGoal

            // Optimistic update
            _uiState.update { it.copy(savingsGoal = null, showEditSheet = false) }

            budgetRepository.deleteSavingsGoal()
                .onFailure { e ->
                    // Rollback
                    _uiState.update { it.copy(savingsGoal = previousGoal, error = e.message) }
                }
        }
    }

    fun showEditSheet() {
        _uiState.update { it.copy(showEditSheet = true) }
    }

    fun hideEditSheet() {
        _uiState.update { it.copy(showEditSheet = false) }
    }

    fun clearError() {
        _uiState.update { it.copy(error = null) }
    }
}

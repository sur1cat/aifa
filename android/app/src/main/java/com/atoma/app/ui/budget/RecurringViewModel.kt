package com.atoma.app.ui.budget

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.atoma.app.data.repository.RecurringRepository
import com.atoma.app.domain.model.RecurrenceFrequency
import com.atoma.app.domain.model.RecurringCategory
import com.atoma.app.domain.model.RecurringProjection
import com.atoma.app.domain.model.RecurringTransaction
import com.atoma.app.domain.model.TransactionType
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import java.time.LocalDate
import javax.inject.Inject

data class RecurringUiState(
    val transactions: List<RecurringTransaction> = emptyList(),
    val projections: List<RecurringProjection> = emptyList(),
    val isLoading: Boolean = false,
    val error: String? = null,
    val showAddSheet: Boolean = false
) {
    val activeTransactions: List<RecurringTransaction>
        get() = transactions.filter { it.isActive }

    val inactiveTransactions: List<RecurringTransaction>
        get() = transactions.filter { !it.isActive }

    val monthlyExpenses: Double
        get() = activeTransactions
            .filter { it.type == TransactionType.EXPENSE }
            .sumOf { it.monthlyAmount }

    val monthlyIncome: Double
        get() = activeTransactions
            .filter { it.type == TransactionType.INCOME }
            .sumOf { it.monthlyAmount }
}

val RecurringTransaction.monthlyAmount: Double
    get() = when (frequency) {
        RecurrenceFrequency.WEEKLY -> amount * 4.33
        RecurrenceFrequency.BIWEEKLY -> amount * 2.17
        RecurrenceFrequency.MONTHLY -> amount
        RecurrenceFrequency.QUARTERLY -> amount / 3
        RecurrenceFrequency.YEARLY -> amount / 12
    }

@HiltViewModel
class RecurringViewModel @Inject constructor(
    private val repository: RecurringRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(RecurringUiState())
    val uiState: StateFlow<RecurringUiState> = _uiState

    init {
        loadData()
    }

    fun loadData() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }

            repository.getRecurringTransactions()
                .onSuccess { transactions ->
                    _uiState.update { it.copy(transactions = transactions, isLoading = false) }
                }
                .onFailure { e ->
                    _uiState.update { it.copy(error = e.message, isLoading = false) }
                }

            // Also load projections
            repository.getProjection(3)
                .onSuccess { projections ->
                    _uiState.update { it.copy(projections = projections) }
                }
        }
    }

    fun createRecurring(
        title: String,
        amount: Double,
        type: TransactionType,
        category: RecurringCategory,
        frequency: RecurrenceFrequency
    ) {
        viewModelScope.launch {
            repository.createRecurringTransaction(title, amount, type, category, frequency)
                .onSuccess {
                    _uiState.update { it.copy(showAddSheet = false) }
                    loadData()
                }
                .onFailure { e ->
                    _uiState.update { it.copy(error = e.message) }
                }
        }
    }

    fun toggleActive(transaction: RecurringTransaction) {
        viewModelScope.launch {
            // Optimistic update
            _uiState.update { state ->
                state.copy(
                    transactions = state.transactions.map {
                        if (it.id == transaction.id) it.copy(isActive = !it.isActive) else it
                    }
                )
            }

            repository.toggleActive(transaction)
                .onFailure { loadData() }
        }
    }

    fun deleteRecurring(transaction: RecurringTransaction) {
        viewModelScope.launch {
            // Optimistic update
            _uiState.update { state ->
                state.copy(transactions = state.transactions.filter { it.id != transaction.id })
            }

            repository.deleteRecurringTransaction(transaction.id)
                .onFailure { loadData() }
        }
    }

    fun showAddSheet() {
        _uiState.update { it.copy(showAddSheet = true) }
    }

    fun hideAddSheet() {
        _uiState.update { it.copy(showAddSheet = false) }
    }

    fun clearError() {
        _uiState.update { it.copy(error = null) }
    }
}

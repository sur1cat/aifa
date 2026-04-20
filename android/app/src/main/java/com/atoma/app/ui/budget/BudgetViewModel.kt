package com.atoma.app.ui.budget

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.atoma.app.data.repository.BudgetRepository
import com.atoma.app.domain.model.BudgetSummary
import com.atoma.app.domain.model.Transaction
import com.atoma.app.domain.model.TransactionType
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import java.time.LocalDate
import java.time.YearMonth
import javax.inject.Inject

data class BudgetUiState(
    val transactions: List<Transaction> = emptyList(),
    val summary: BudgetSummary = BudgetSummary(0.0, 0.0),
    val isLoading: Boolean = false,
    val error: String? = null,
    val showAddDialog: Boolean = false,
    val selectedMonth: YearMonth = YearMonth.now()
)

@HiltViewModel
class BudgetViewModel @Inject constructor(
    private val repository: BudgetRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(BudgetUiState())
    val uiState: StateFlow<BudgetUiState> = _uiState

    init {
        loadData()
    }

    fun loadData() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }

            val month = _uiState.value.selectedMonth

            repository.getTransactions(month)
                .onSuccess { transactions ->
                    _uiState.update { it.copy(transactions = transactions.sortedByDescending { t -> t.date }) }
                }

            repository.getSummary(month)
                .onSuccess { summary ->
                    _uiState.update { it.copy(summary = summary, isLoading = false) }
                }
                .onFailure { e ->
                    _uiState.update { it.copy(error = e.message, isLoading = false) }
                }
        }
    }

    fun createTransaction(
        title: String,
        amount: Double,
        type: TransactionType,
        category: String
    ) {
        viewModelScope.launch {
            repository.createTransaction(title, amount, type, category, LocalDate.now())
                .onSuccess {
                    _uiState.update { it.copy(showAddDialog = false) }
                    loadData()
                }
                .onFailure { e ->
                    _uiState.update { it.copy(error = e.message) }
                }
        }
    }

    fun deleteTransaction(transaction: Transaction) {
        viewModelScope.launch {
            _uiState.update { state ->
                state.copy(transactions = state.transactions.filter { it.id != transaction.id })
            }
            repository.deleteTransaction(transaction.id)
                .onFailure { loadData() }
        }
    }

    fun setMonth(yearMonth: YearMonth) {
        _uiState.update { it.copy(selectedMonth = yearMonth) }
        loadData()
    }

    fun showAddDialog() {
        _uiState.update { it.copy(showAddDialog = true) }
    }

    fun hideAddDialog() {
        _uiState.update { it.copy(showAddDialog = false) }
    }
}

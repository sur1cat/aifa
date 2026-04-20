package com.atoma.app.ui.budget

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.atoma.app.data.repository.BudgetRepository
import com.atoma.app.data.repository.RecurringRepository
import com.atoma.app.domain.model.*
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import java.time.LocalDate
import javax.inject.Inject

data class ForecastUiState(
    val transactions: List<Transaction> = emptyList(),
    val recurringTransactions: List<RecurringTransaction> = emptyList(),
    val forecast: BudgetForecast? = null,
    val isLoading: Boolean = false,
    val error: String? = null
)

@HiltViewModel
class ForecastViewModel @Inject constructor(
    private val budgetRepository: BudgetRepository,
    private val recurringRepository: RecurringRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(ForecastUiState())
    val uiState: StateFlow<ForecastUiState> = _uiState

    init {
        loadData()
    }

    fun loadData() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }

            budgetRepository.getTransactions()
                .onSuccess { transactions ->
                    _uiState.update { it.copy(transactions = transactions) }
                }
                .onFailure { e ->
                    _uiState.update { it.copy(error = e.message) }
                }

            recurringRepository.getRecurringTransactions()
                .onSuccess { recurring ->
                    _uiState.update { it.copy(recurringTransactions = recurring) }
                }
                .onFailure { e ->
                    _uiState.update { it.copy(error = e.message) }
                }

            // Generate forecast
            val state = _uiState.value
            val forecast = generateForecast(state.transactions, state.recurringTransactions)
            _uiState.update { it.copy(forecast = forecast, isLoading = false) }
        }
    }

    private fun generateForecast(
        transactions: List<Transaction>,
        recurring: List<RecurringTransaction>
    ): BudgetForecast {
        val today = LocalDate.now()
        val nextMonth = today.plusMonths(1)

        // Get historical monthly data (last 6 months)
        val historicalMonths = getHistoricalMonthlyData(transactions, 6)

        // Calculate category forecasts
        val categoryForecasts = calculateCategoryForecasts(historicalMonths, recurring)

        // Calculate projected expenses from category forecasts
        val projectedExpenses = categoryForecasts.sumOf { it.projectedAmount }

        // Calculate projected income
        val projectedIncome = calculateProjectedIncome(historicalMonths) +
                getProjectedMonthlyIncomeFromRecurring(recurring)

        // Calculate trend
        val trend = calculateExpenseTrend(historicalMonths)

        // Calculate confidence based on data availability
        val confidence = minOf(1.0, historicalMonths.size * 0.15 + 0.2)

        // Detect seasonal factors
        val seasonalFactors = detectSeasonality(nextMonth)

        return BudgetForecast(
            forecastMonth = nextMonth,
            categoryForecasts = categoryForecasts,
            projectedExpenses = projectedExpenses,
            projectedIncome = projectedIncome,
            projectedSavings = projectedIncome - projectedExpenses,
            expenseTrend = trend,
            confidenceScore = confidence,
            seasonalFactors = seasonalFactors
        )
    }

    private fun getHistoricalMonthlyData(transactions: List<Transaction>, months: Int): List<List<Transaction>> {
        val today = LocalDate.now()
        val result = mutableListOf<List<Transaction>>()

        for (monthOffset in 1..months) {
            val targetDate = today.minusMonths(monthOffset.toLong())
            val monthTransactions = transactions.filter {
                it.date.year == targetDate.year && it.date.monthValue == targetDate.monthValue
            }
            if (monthTransactions.isNotEmpty()) {
                result.add(monthTransactions)
            }
        }

        return result
    }

    private fun calculateCategoryForecasts(
        historicalMonths: List<List<Transaction>>,
        recurring: List<RecurringTransaction>
    ): List<CategoryForecast> {
        val categoryTotals = mutableMapOf<String, MutableList<Double>>()

        // Collect expenses by category for each month
        for (monthData in historicalMonths) {
            val monthCategoryTotals = mutableMapOf<String, Double>()
            for (tx in monthData.filter { it.type == TransactionType.EXPENSE }) {
                val category = tx.category.ifEmpty { "other" }
                monthCategoryTotals[category] = (monthCategoryTotals[category] ?: 0.0) + tx.amount
            }
            for ((category, total) in monthCategoryTotals) {
                categoryTotals.getOrPut(category) { mutableListOf() }.add(total)
            }
        }

        val forecasts = mutableListOf<CategoryForecast>()

        for ((category, amounts) in categoryTotals) {
            if (amounts.isEmpty()) continue

            val average = amounts.average()
            val weightedAvg = calculateWeightedAverage(amounts)
            val recurringAmount = getRecurringAmountForCategory(category, recurring)
            val variableAmount = maxOf(0.0, weightedAvg - recurringAmount)
            val trend = determineTrend(amounts)
            val changePercent = if (amounts.size > 1 && amounts.last() > 0) {
                ((amounts.first() - amounts.last()) / amounts.last()) * 100
            } else 0.0

            forecasts.add(
                CategoryForecast(
                    category = category,
                    projectedAmount = weightedAvg,
                    historicalAverage = average,
                    changePercent = changePercent,
                    trend = trend,
                    recurringAmount = recurringAmount,
                    variableAmount = variableAmount
                )
            )
        }

        return forecasts.sortedByDescending { it.projectedAmount }
    }

    private fun calculateWeightedAverage(values: List<Double>): Double {
        if (values.isEmpty()) return 0.0

        var weightedSum = 0.0
        var totalWeight = 0.0

        for ((index, value) in values.withIndex()) {
            val weight = (values.size - index).toDouble()
            weightedSum += value * weight
            totalWeight += weight
        }

        return weightedSum / totalWeight
    }

    private fun getRecurringAmountForCategory(category: String, recurring: List<RecurringTransaction>): Double {
        val categoryLower = category.lowercase()
        return recurring
            .filter { it.isActive && it.type == TransactionType.EXPENSE }
            .filter {
                it.category.key.lowercase().contains(categoryLower) ||
                        categoryLower.contains(it.category.key.lowercase()) ||
                        it.title.lowercase().contains(categoryLower)
            }
            .sumOf { it.monthlyAmount }
    }

    private val RecurringTransaction.monthlyAmount: Double
        get() = when (frequency) {
            RecurrenceFrequency.WEEKLY -> amount * 4.33
            RecurrenceFrequency.BIWEEKLY -> amount * 2.17
            RecurrenceFrequency.MONTHLY -> amount
            RecurrenceFrequency.QUARTERLY -> amount / 3
            RecurrenceFrequency.YEARLY -> amount / 12
        }

    private fun getProjectedMonthlyIncomeFromRecurring(recurring: List<RecurringTransaction>): Double {
        return recurring
            .filter { it.isActive && it.type == TransactionType.INCOME }
            .sumOf { it.monthlyAmount }
    }

    private fun determineTrend(values: List<Double>): TrendDirection {
        if (values.size < 2) return TrendDirection.STABLE

        val recent = values.take(2).average()
        val older = values.takeLast(minOf(2, values.size)).average()

        if (older <= 0) return TrendDirection.STABLE
        val change = (recent - older) / older

        return when {
            change > 0.1 -> TrendDirection.UP
            change < -0.1 -> TrendDirection.DOWN
            else -> TrendDirection.STABLE
        }
    }

    private fun calculateProjectedIncome(historicalMonths: List<List<Transaction>>): Double {
        val incomes = historicalMonths.map { month ->
            month.filter { it.type == TransactionType.INCOME }.sumOf { it.amount }
        }

        return if (incomes.isEmpty()) 0.0 else calculateWeightedAverage(incomes)
    }

    private fun calculateExpenseTrend(historicalMonths: List<List<Transaction>>): TrendDirection {
        val monthlyExpenses = historicalMonths.map { month ->
            month.filter { it.type == TransactionType.EXPENSE }.sumOf { it.amount }
        }
        return determineTrend(monthlyExpenses)
    }

    private fun detectSeasonality(month: LocalDate): List<SeasonalFactor>? {
        val factors = mutableListOf<SeasonalFactor>()

        when (month.monthValue) {
            12 -> {
                factors.add(SeasonalFactor("Shopping", 1.5, "Holiday shopping"))
                factors.add(SeasonalFactor("Entertainment", 1.3, "Holiday gatherings"))
            }
            8, 9 -> {
                factors.add(SeasonalFactor("Education", 2.0, "Back to school"))
            }
            1 -> {
                factors.add(SeasonalFactor("Health", 1.4, "New Year resolutions"))
            }
            2 -> {
                factors.add(SeasonalFactor("Gift", 1.5, "Valentine's Day"))
            }
        }

        return factors.ifEmpty { null }
    }

    fun clearError() {
        _uiState.update { it.copy(error = null) }
    }
}

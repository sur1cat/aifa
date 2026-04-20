package com.atoma.app.domain.model

import java.time.LocalDate
import java.util.UUID

enum class TrendDirection {
    UP, DOWN, STABLE;

    val icon: String
        get() = when (this) {
            UP -> "trending_up"
            DOWN -> "trending_down"
            STABLE -> "trending_flat"
        }
}

data class CategoryForecast(
    val id: String = UUID.randomUUID().toString(),
    val category: String,
    val projectedAmount: Double,
    val historicalAverage: Double,
    val changePercent: Double = 0.0,
    val trend: TrendDirection = TrendDirection.STABLE,
    val recurringAmount: Double = 0.0,
    val variableAmount: Double = 0.0
)

data class SeasonalFactor(
    val category: String,
    val monthlyMultiplier: Double,
    val reason: String
)

data class BudgetForecast(
    val id: String = UUID.randomUUID().toString(),
    val forecastMonth: LocalDate,
    val generatedAt: LocalDate = LocalDate.now(),
    val categoryForecasts: List<CategoryForecast> = emptyList(),
    val projectedExpenses: Double = 0.0,
    val projectedIncome: Double = 0.0,
    val projectedSavings: Double = 0.0,
    val expenseTrend: TrendDirection = TrendDirection.STABLE,
    val confidenceScore: Double = 0.0,
    val seasonalFactors: List<SeasonalFactor>? = null
)

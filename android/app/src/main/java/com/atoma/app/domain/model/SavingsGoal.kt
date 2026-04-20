package com.atoma.app.domain.model

import java.util.UUID

data class SavingsGoal(
    val id: String = UUID.randomUUID().toString(),
    val monthlyTarget: Double,
    val currentSavings: Double = 0.0
) {
    val progress: Float
        get() = if (monthlyTarget > 0) {
            (currentSavings / monthlyTarget).toFloat().coerceIn(0f, 1f)
        } else 0f

    val remainingAmount: Double
        get() = (monthlyTarget - currentSavings).coerceAtLeast(0.0)

    val isCompleted: Boolean
        get() = currentSavings >= monthlyTarget
}

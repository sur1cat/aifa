package com.atoma.app.data.repository

import com.atoma.app.data.api.AtomaApi
import com.atoma.app.data.model.AIChatRequest
import com.atoma.app.data.model.AIContext
import com.atoma.app.data.model.AIHabitContext
import com.atoma.app.data.model.AITaskContext
import com.atoma.app.data.model.AITransactionContext
import com.atoma.app.domain.model.Habit
import com.atoma.app.domain.model.DailyTask
import com.atoma.app.domain.model.Transaction
import javax.inject.Inject
import javax.inject.Singleton

enum class AIAgent(val id: String, val displayName: String, val icon: String) {
    HABIT_COACH("habit_coach", "Habit Coach", "🎯"),
    TASK_ASSISTANT("task_assistant", "Task Assistant", "✅"),
    FINANCE_ADVISOR("finance_advisor", "Finance Advisor", "💰"),
    LIFE_COACH("life_coach", "Life Coach", "🌟")
}

@Singleton
class AIRepository @Inject constructor(
    private val api: AtomaApi
) {
    suspend fun chat(
        message: String,
        agent: AIAgent,
        habits: List<Habit> = emptyList(),
        tasks: List<DailyTask> = emptyList(),
        transactions: List<Transaction> = emptyList()
    ): Result<String> {
        return try {
            val context = AIContext(
                habits = habits.takeIf { it.isNotEmpty() }?.map { habit ->
                    AIHabitContext(
                        title = habit.title,
                        completedToday = habit.isCompletedToday,
                        streak = habit.streak
                    )
                },
                tasks = tasks.takeIf { it.isNotEmpty() }?.map { task ->
                    AITaskContext(
                        title = task.title,
                        isCompleted = task.isCompleted,
                        priority = task.priority.name.lowercase()
                    )
                },
                transactions = transactions.takeIf { it.isNotEmpty() }?.map { tx ->
                    AITransactionContext(
                        title = tx.title,
                        amount = tx.amount,
                        type = tx.type.name.lowercase(),
                        category = tx.category
                    )
                }
            )

            val request = AIChatRequest(
                message = message,
                agent = agent.id,
                context = context.takeIf {
                    it.habits != null || it.tasks != null || it.transactions != null
                }
            )

            val response = api.aiChat(request)
            if (response.data != null) {
                Result.success(response.data.response)
            } else {
                Result.failure(Exception(response.error?.message ?: "AI request failed"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}

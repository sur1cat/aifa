package com.atoma.app.data.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

// MARK: - Auth
@Serializable
data class GoogleAuthRequest(
    @SerialName("id_token") val idToken: String
)

@Serializable
data class AuthResponse(
    val tokens: TokensDto,
    val user: UserDto,
    @SerialName("is_new_user") val isNewUser: Boolean = false
)

@Serializable
data class TokensDto(
    @SerialName("access_token") val accessToken: String,
    @SerialName("refresh_token") val refreshToken: String
)

@Serializable
data class UserDto(
    val id: String,
    val email: String,
    val name: String,
    @SerialName("avatar_url") val avatarUrl: String? = null,
    @SerialName("is_premium") val isPremium: Boolean = false
)

// MARK: - Habits
@Serializable
data class HabitDto(
    val id: String,
    val title: String,
    val icon: String,
    val color: String,
    val period: String,
    @SerialName("completed_dates") val completedDates: List<String> = emptyList(),
    @SerialName("created_at") val createdAt: String,
    @SerialName("reminder_enabled") val reminderEnabled: Boolean = false,
    @SerialName("reminder_time") val reminderTime: String? = null,
    @SerialName("archived_at") val archivedAt: String? = null,
    @SerialName("goal_id") val goalId: String? = null
)

@Serializable
data class CreateHabitRequest(
    val title: String,
    val icon: String,
    val color: String,
    val period: String
)

@Serializable
data class UpdateHabitRequest(
    val title: String? = null,
    val icon: String? = null,
    val color: String? = null,
    val period: String? = null,
    @SerialName("archived_at") val archivedAt: String? = null,
    @SerialName("goal_id") val goalId: String? = null
)

@Serializable
data class ToggleHabitRequest(
    val date: String
)

// MARK: - Tasks
@Serializable
data class TaskDto(
    val id: String,
    val title: String,
    @SerialName("is_completed") val isCompleted: Boolean,
    val priority: String,
    @SerialName("due_date") val dueDate: String,
    @SerialName("created_at") val createdAt: String
)

@Serializable
data class CreateTaskRequest(
    val title: String,
    val priority: String,
    @SerialName("due_date") val dueDate: String
)

// MARK: - Transactions
@Serializable
data class TransactionDto(
    val id: String,
    val title: String,
    val amount: Double,
    val type: String,
    val category: String,
    val date: String
)

@Serializable
data class CreateTransactionRequest(
    val title: String,
    val amount: Double,
    val type: String,
    val category: String,
    val date: String
)

@Serializable
data class SummaryDto(
    val income: Double,
    val expenses: Double
)

// MARK: - Recurring Transactions
@Serializable
data class RecurringTransactionDto(
    val id: String,
    val title: String,
    val amount: Double,
    val type: String,
    val category: String,
    val frequency: String,
    @SerialName("start_date") val startDate: String,
    @SerialName("next_date") val nextDate: String,
    @SerialName("end_date") val endDate: String? = null,
    @SerialName("is_active") val isActive: Boolean = true
)

@Serializable
data class CreateRecurringTransactionRequest(
    val title: String,
    val amount: Double,
    val type: String,
    val category: String,
    val frequency: String,
    @SerialName("start_date") val startDate: String,
    @SerialName("next_date") val nextDate: String? = null,
    @SerialName("end_date") val endDate: String? = null,
    @SerialName("is_active") val isActive: Boolean = true
)

@Serializable
data class RecurringProjectionDto(
    val date: String,
    val amount: Double,
    val type: String,
    @SerialName("recurring_id") val recurringId: String,
    val title: String
)

// MARK: - Goals
@Serializable
data class GoalDto(
    val id: String,
    val title: String,
    val icon: String,
    @SerialName("target_value") val targetValue: Int? = null,
    val unit: String? = null,
    val deadline: String? = null,
    @SerialName("created_at") val createdAt: String,
    @SerialName("archived_at") val archivedAt: String? = null
)

@Serializable
data class CreateGoalRequest(
    val title: String,
    val icon: String,
    @SerialName("target_value") val targetValue: Int? = null,
    val unit: String? = null,
    val deadline: String? = null
)

@Serializable
data class UpdateGoalRequest(
    val title: String,
    val icon: String,
    @SerialName("target_value") val targetValue: Int? = null,
    val unit: String? = null,
    val deadline: String? = null,
    @SerialName("archived_at") val archivedAt: String? = null
)

// MARK: - Savings Goal
@Serializable
data class SavingsGoalDto(
    val id: String,
    @SerialName("monthly_target") val monthlyTarget: Double
)

@Serializable
data class SaveSavingsGoalRequest(
    @SerialName("monthly_target") val monthlyTarget: Double
)

// MARK: - AI
@Serializable
data class AIChatRequest(
    val message: String,
    val agent: String,
    val context: AIContext? = null
)

@Serializable
data class AIContext(
    val habits: List<AIHabitContext>? = null,
    val tasks: List<AITaskContext>? = null,
    val transactions: List<AITransactionContext>? = null
)

@Serializable
data class AIHabitContext(
    val title: String,
    @SerialName("completed_today") val completedToday: Boolean,
    val streak: Int
)

@Serializable
data class AITaskContext(
    val title: String,
    @SerialName("is_completed") val isCompleted: Boolean,
    val priority: String
)

@Serializable
data class AITransactionContext(
    val title: String,
    val amount: Double,
    val type: String,
    val category: String
)

@Serializable
data class AIChatResponse(
    val response: String,
    val agent: String
)

// MARK: - Generic API Response
@Serializable
data class ApiResponse<T>(
    val data: T? = null,
    val error: ApiError? = null
)

@Serializable
data class ApiError(
    val code: String,
    val message: String
)

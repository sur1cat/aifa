package com.atoma.app.data.api

import com.atoma.app.data.model.*
import retrofit2.http.*

interface AtomaApi {

    // Auth
    @POST("api/v1/auth/google")
    suspend fun googleAuth(@Body request: GoogleAuthRequest): ApiResponse<AuthResponse>

    @POST("api/v1/auth/refresh")
    suspend fun refreshToken(): ApiResponse<AuthResponse>

    @GET("api/v1/auth/me")
    suspend fun getCurrentUser(): ApiResponse<UserDto>

    @DELETE("api/v1/auth/me")
    suspend fun deleteAccount(): ApiResponse<Unit>

    // Habits
    @GET("api/v1/habits")
    suspend fun getHabits(): ApiResponse<List<HabitDto>>

    @POST("api/v1/habits")
    suspend fun createHabit(@Body request: CreateHabitRequest): ApiResponse<HabitDto>

    @PUT("api/v1/habits/{id}")
    suspend fun updateHabit(
        @Path("id") id: String,
        @Body request: UpdateHabitRequest
    ): ApiResponse<HabitDto>

    @DELETE("api/v1/habits/{id}")
    suspend fun deleteHabit(@Path("id") id: String): ApiResponse<Unit>

    @POST("api/v1/habits/{id}/toggle")
    suspend fun toggleHabit(
        @Path("id") id: String,
        @Body request: ToggleHabitRequest
    ): ApiResponse<HabitDto>

    // Tasks
    @GET("api/v1/tasks")
    suspend fun getTasks(@Query("date") date: String? = null): ApiResponse<List<TaskDto>>

    @POST("api/v1/tasks")
    suspend fun createTask(@Body request: CreateTaskRequest): ApiResponse<TaskDto>

    @PUT("api/v1/tasks/{id}")
    suspend fun updateTask(
        @Path("id") id: String,
        @Body request: CreateTaskRequest
    ): ApiResponse<TaskDto>

    @DELETE("api/v1/tasks/{id}")
    suspend fun deleteTask(@Path("id") id: String): ApiResponse<Unit>

    @POST("api/v1/tasks/{id}/toggle")
    suspend fun toggleTask(@Path("id") id: String): ApiResponse<TaskDto>

    // Transactions
    @GET("api/v1/transactions")
    suspend fun getTransactions(
        @Query("year") year: Int? = null,
        @Query("month") month: Int? = null
    ): ApiResponse<List<TransactionDto>>

    @POST("api/v1/transactions")
    suspend fun createTransaction(@Body request: CreateTransactionRequest): ApiResponse<TransactionDto>

    @PUT("api/v1/transactions/{id}")
    suspend fun updateTransaction(
        @Path("id") id: String,
        @Body request: CreateTransactionRequest
    ): ApiResponse<TransactionDto>

    @DELETE("api/v1/transactions/{id}")
    suspend fun deleteTransaction(@Path("id") id: String): ApiResponse<Unit>

    @GET("api/v1/transactions/summary")
    suspend fun getSummary(
        @Query("year") year: Int,
        @Query("month") month: Int
    ): ApiResponse<SummaryDto>

    // Recurring Transactions
    @GET("api/v1/recurring-transactions")
    suspend fun getRecurringTransactions(): ApiResponse<List<RecurringTransactionDto>>

    @POST("api/v1/recurring-transactions")
    suspend fun createRecurringTransaction(
        @Body request: CreateRecurringTransactionRequest
    ): ApiResponse<RecurringTransactionDto>

    @PUT("api/v1/recurring-transactions/{id}")
    suspend fun updateRecurringTransaction(
        @Path("id") id: String,
        @Body request: CreateRecurringTransactionRequest
    ): ApiResponse<RecurringTransactionDto>

    @DELETE("api/v1/recurring-transactions/{id}")
    suspend fun deleteRecurringTransaction(@Path("id") id: String): ApiResponse<Unit>

    @GET("api/v1/recurring-transactions/projection")
    suspend fun getRecurringProjection(
        @Query("months") months: Int = 3
    ): ApiResponse<List<RecurringProjectionDto>>

    // Goals
    @GET("api/v1/goals")
    suspend fun getGoals(): ApiResponse<List<GoalDto>>

    @POST("api/v1/goals")
    suspend fun createGoal(@Body request: CreateGoalRequest): ApiResponse<GoalDto>

    @PUT("api/v1/goals/{id}")
    suspend fun updateGoal(
        @Path("id") id: String,
        @Body request: UpdateGoalRequest
    ): ApiResponse<GoalDto>

    @DELETE("api/v1/goals/{id}")
    suspend fun deleteGoal(@Path("id") id: String): ApiResponse<Unit>

    // AI
    @POST("api/v1/ai/chat")
    suspend fun aiChat(@Body request: AIChatRequest): ApiResponse<AIChatResponse>

    // Savings Goal
    @GET("api/v1/savings-goal")
    suspend fun getSavingsGoal(): ApiResponse<SavingsGoalDto>

    @PUT("api/v1/savings-goal")
    suspend fun saveSavingsGoal(@Body request: SaveSavingsGoalRequest): ApiResponse<SavingsGoalDto>

    @DELETE("api/v1/savings-goal")
    suspend fun deleteSavingsGoal(): ApiResponse<Unit>
}

package com.atoma.app.data.repository

import com.atoma.app.data.api.AtomaApi
import com.atoma.app.data.model.CreateGoalRequest
import com.atoma.app.data.model.GoalDto
import com.atoma.app.data.model.UpdateGoalRequest
import com.atoma.app.domain.model.Goal
import java.time.LocalDate
import java.time.LocalDateTime
import java.time.format.DateTimeFormatter
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class GoalsRepository @Inject constructor(
    private val api: AtomaApi
) {
    suspend fun getGoals(): Result<List<Goal>> {
        return try {
            val response = api.getGoals()
            if (response.data != null) {
                Result.success(response.data.map { it.toDomain() })
            } else {
                Result.failure(Exception(response.error?.message ?: "Failed to get goals"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun createGoal(
        title: String,
        icon: String,
        targetValue: Int? = null,
        unit: String? = null,
        deadline: LocalDate? = null
    ): Result<Goal> {
        return try {
            val request = CreateGoalRequest(
                title = title,
                icon = icon,
                targetValue = targetValue,
                unit = unit,
                deadline = deadline?.toString()
            )
            val response = api.createGoal(request)
            if (response.data != null) {
                Result.success(response.data.toDomain())
            } else {
                Result.failure(Exception(response.error?.message ?: "Failed to create goal"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun updateGoal(goal: Goal): Result<Goal> {
        return try {
            val request = UpdateGoalRequest(
                title = goal.title,
                icon = goal.icon,
                targetValue = goal.targetValue,
                unit = goal.unit,
                deadline = goal.deadline?.toString(),
                archivedAt = goal.archivedAt?.format(DateTimeFormatter.ISO_DATE_TIME) ?: ""
            )
            val response = api.updateGoal(goal.id, request)
            if (response.data != null) {
                Result.success(response.data.toDomain())
            } else {
                Result.failure(Exception(response.error?.message ?: "Failed to update goal"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun archiveGoal(goalId: String): Result<Goal> {
        return try {
            val request = UpdateGoalRequest(
                title = "",  // Will be ignored by server for archive
                icon = "",
                archivedAt = LocalDateTime.now().format(DateTimeFormatter.ISO_DATE_TIME)
            )
            val response = api.updateGoal(goalId, request)
            if (response.data != null) {
                Result.success(response.data.toDomain())
            } else {
                Result.failure(Exception(response.error?.message ?: "Failed to archive goal"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun unarchiveGoal(goalId: String): Result<Goal> {
        return try {
            val request = UpdateGoalRequest(
                title = "",
                icon = "",
                archivedAt = ""  // Empty string to clear archived_at
            )
            val response = api.updateGoal(goalId, request)
            if (response.data != null) {
                Result.success(response.data.toDomain())
            } else {
                Result.failure(Exception(response.error?.message ?: "Failed to unarchive goal"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun deleteGoal(goalId: String): Result<Unit> {
        return try {
            val response = api.deleteGoal(goalId)
            if (response.error == null) {
                Result.success(Unit)
            } else {
                Result.failure(Exception(response.error.message))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    private fun GoalDto.toDomain(): Goal {
        return Goal(
            id = id,
            title = title,
            icon = icon,
            targetValue = targetValue,
            unit = unit,
            deadline = deadline?.let { LocalDate.parse(it.substring(0, 10)) },
            createdAt = LocalDateTime.parse(createdAt, DateTimeFormatter.ISO_DATE_TIME),
            archivedAt = archivedAt?.takeIf { it.isNotBlank() }?.let {
                LocalDateTime.parse(it, DateTimeFormatter.ISO_DATE_TIME)
            }
        )
    }
}

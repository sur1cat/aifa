package com.atoma.app.data.repository

import com.atoma.app.data.api.AtomaApi
import com.atoma.app.data.model.CreateHabitRequest
import com.atoma.app.data.model.HabitDto
import com.atoma.app.data.model.ToggleHabitRequest
import com.atoma.app.data.model.UpdateHabitRequest
import com.atoma.app.domain.model.Habit
import com.atoma.app.domain.model.HabitPeriod
import java.time.LocalDate
import java.time.LocalDateTime
import java.time.format.DateTimeFormatter
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class HabitsRepository @Inject constructor(
    private val api: AtomaApi
) {
    suspend fun getHabits(): Result<List<Habit>> {
        return try {
            val response = api.getHabits()
            if (response.data != null) {
                Result.success(response.data.map { it.toDomain() })
            } else {
                Result.failure(Exception(response.error?.message ?: "Failed to get habits"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun createHabit(title: String, icon: String, color: String, period: HabitPeriod): Result<Habit> {
        return try {
            val request = CreateHabitRequest(
                title = title,
                icon = icon,
                color = color,
                period = period.name.lowercase()
            )
            val response = api.createHabit(request)
            if (response.data != null) {
                Result.success(response.data.toDomain())
            } else {
                Result.failure(Exception(response.error?.message ?: "Failed to create habit"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun toggleHabit(habitId: String, date: LocalDate = LocalDate.now()): Result<Habit> {
        return try {
            val request = ToggleHabitRequest(date.toString())
            val response = api.toggleHabit(habitId, request)
            if (response.data != null) {
                Result.success(response.data.toDomain())
            } else {
                Result.failure(Exception(response.error?.message ?: "Failed to toggle habit"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun deleteHabit(habitId: String): Result<Unit> {
        return try {
            val response = api.deleteHabit(habitId)
            if (response.error == null) {
                Result.success(Unit)
            } else {
                Result.failure(Exception(response.error.message))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun archiveHabit(habitId: String): Result<Habit> {
        return try {
            val request = UpdateHabitRequest(
                archivedAt = LocalDateTime.now().format(DateTimeFormatter.ISO_DATE_TIME)
            )
            val response = api.updateHabit(habitId, request)
            if (response.data != null) {
                Result.success(response.data.toDomain())
            } else {
                Result.failure(Exception(response.error?.message ?: "Failed to archive habit"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun unarchiveHabit(habitId: String): Result<Habit> {
        return try {
            val request = UpdateHabitRequest(archivedAt = null)
            val response = api.updateHabit(habitId, request)
            if (response.data != null) {
                Result.success(response.data.toDomain())
            } else {
                Result.failure(Exception(response.error?.message ?: "Failed to unarchive habit"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    private fun HabitDto.toDomain(): Habit {
        return Habit(
            id = id,
            title = title,
            icon = icon,
            color = color,
            period = HabitPeriod.valueOf(period.uppercase()),
            completedDates = completedDates,
            createdAt = LocalDateTime.parse(createdAt, DateTimeFormatter.ISO_DATE_TIME),
            reminderEnabled = reminderEnabled,
            reminderTime = reminderTime?.let {
                LocalDateTime.parse(it, DateTimeFormatter.ISO_DATE_TIME)
            },
            archivedAt = archivedAt?.let {
                LocalDateTime.parse(it, DateTimeFormatter.ISO_DATE_TIME)
            },
            goalId = goalId
        )
    }
}

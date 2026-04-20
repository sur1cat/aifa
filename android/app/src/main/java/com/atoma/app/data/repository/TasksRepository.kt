package com.atoma.app.data.repository

import com.atoma.app.data.api.AtomaApi
import com.atoma.app.data.model.CreateTaskRequest
import com.atoma.app.data.model.TaskDto
import com.atoma.app.domain.model.DailyTask
import com.atoma.app.domain.model.TaskPriority
import java.time.LocalDate
import java.time.LocalDateTime
import java.time.format.DateTimeFormatter
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class TasksRepository @Inject constructor(
    private val api: AtomaApi
) {
    suspend fun getTasks(date: LocalDate? = null): Result<List<DailyTask>> {
        return try {
            val response = api.getTasks(date?.toString())
            if (response.data != null) {
                Result.success(response.data.map { it.toDomain() })
            } else {
                Result.failure(Exception(response.error?.message ?: "Failed to get tasks"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun createTask(title: String, priority: TaskPriority, dueDate: LocalDate): Result<DailyTask> {
        return try {
            val request = CreateTaskRequest(
                title = title,
                priority = priority.name.lowercase(),
                dueDate = dueDate.toString()
            )
            val response = api.createTask(request)
            if (response.data != null) {
                Result.success(response.data.toDomain())
            } else {
                Result.failure(Exception(response.error?.message ?: "Failed to create task"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun toggleTask(taskId: String): Result<DailyTask> {
        return try {
            val response = api.toggleTask(taskId)
            if (response.data != null) {
                Result.success(response.data.toDomain())
            } else {
                Result.failure(Exception(response.error?.message ?: "Failed to toggle task"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun deleteTask(taskId: String): Result<Unit> {
        return try {
            val response = api.deleteTask(taskId)
            if (response.error == null) {
                Result.success(Unit)
            } else {
                Result.failure(Exception(response.error.message))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    private fun TaskDto.toDomain(): DailyTask {
        return DailyTask(
            id = id,
            title = title,
            isCompleted = isCompleted,
            priority = TaskPriority.valueOf(priority.uppercase()),
            dueDate = LocalDate.parse(dueDate),
            createdAt = LocalDateTime.parse(createdAt, DateTimeFormatter.ISO_DATE_TIME)
        )
    }
}

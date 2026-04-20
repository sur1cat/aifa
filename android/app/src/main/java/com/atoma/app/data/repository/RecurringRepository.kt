package com.atoma.app.data.repository

import com.atoma.app.data.api.AtomaApi
import com.atoma.app.data.model.CreateRecurringTransactionRequest
import com.atoma.app.data.model.RecurringProjectionDto
import com.atoma.app.data.model.RecurringTransactionDto
import com.atoma.app.domain.model.RecurrenceFrequency
import com.atoma.app.domain.model.RecurringCategory
import com.atoma.app.domain.model.RecurringProjection
import com.atoma.app.domain.model.RecurringTransaction
import com.atoma.app.domain.model.TransactionType
import java.time.LocalDate
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class RecurringRepository @Inject constructor(
    private val api: AtomaApi
) {
    suspend fun getRecurringTransactions(): Result<List<RecurringTransaction>> {
        return try {
            val response = api.getRecurringTransactions()
            if (response.data != null) {
                Result.success(response.data.map { it.toDomain() })
            } else {
                Result.failure(Exception(response.error?.message ?: "Failed to get recurring transactions"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun createRecurringTransaction(
        title: String,
        amount: Double,
        type: TransactionType,
        category: RecurringCategory,
        frequency: RecurrenceFrequency,
        startDate: LocalDate = LocalDate.now()
    ): Result<RecurringTransaction> {
        return try {
            val request = CreateRecurringTransactionRequest(
                title = title,
                amount = amount,
                type = type.name.lowercase(),
                category = category.key,
                frequency = frequency.apiValue,
                startDate = startDate.toString(),
                nextDate = startDate.toString(),
                isActive = true
            )
            val response = api.createRecurringTransaction(request)
            if (response.data != null) {
                Result.success(response.data.toDomain())
            } else {
                Result.failure(Exception(response.error?.message ?: "Failed to create recurring transaction"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun updateRecurringTransaction(transaction: RecurringTransaction): Result<RecurringTransaction> {
        return try {
            val request = CreateRecurringTransactionRequest(
                title = transaction.title,
                amount = transaction.amount,
                type = transaction.type.name.lowercase(),
                category = transaction.category.key,
                frequency = transaction.frequency.apiValue,
                startDate = transaction.startDate.toString(),
                nextDate = transaction.nextDate.toString(),
                endDate = transaction.endDate?.toString(),
                isActive = transaction.isActive
            )
            val response = api.updateRecurringTransaction(transaction.id, request)
            if (response.data != null) {
                Result.success(response.data.toDomain())
            } else {
                Result.failure(Exception(response.error?.message ?: "Failed to update recurring transaction"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun deleteRecurringTransaction(id: String): Result<Unit> {
        return try {
            val response = api.deleteRecurringTransaction(id)
            if (response.error == null) {
                Result.success(Unit)
            } else {
                Result.failure(Exception(response.error.message))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun toggleActive(transaction: RecurringTransaction): Result<RecurringTransaction> {
        return updateRecurringTransaction(transaction.copy(isActive = !transaction.isActive))
    }

    suspend fun getProjection(months: Int = 3): Result<List<RecurringProjection>> {
        return try {
            val response = api.getRecurringProjection(months)
            if (response.data != null) {
                Result.success(response.data.map { it.toDomain() })
            } else {
                Result.failure(Exception(response.error?.message ?: "Failed to get projection"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    private fun RecurringTransactionDto.toDomain(): RecurringTransaction {
        return RecurringTransaction(
            id = id,
            title = title,
            amount = amount,
            type = if (type.equals("income", ignoreCase = true)) TransactionType.INCOME else TransactionType.EXPENSE,
            category = RecurringCategory.fromKey(category),
            frequency = RecurrenceFrequency.fromApi(frequency),
            startDate = LocalDate.parse(startDate.substring(0, 10)),
            nextDate = LocalDate.parse(nextDate.substring(0, 10)),
            endDate = endDate?.takeIf { it.isNotBlank() }?.let { LocalDate.parse(it.substring(0, 10)) },
            isActive = isActive
        )
    }

    private fun RecurringProjectionDto.toDomain(): RecurringProjection {
        return RecurringProjection(
            date = LocalDate.parse(date.substring(0, 10)),
            amount = amount,
            type = if (type.equals("income", ignoreCase = true)) TransactionType.INCOME else TransactionType.EXPENSE,
            recurringId = recurringId,
            title = title
        )
    }
}

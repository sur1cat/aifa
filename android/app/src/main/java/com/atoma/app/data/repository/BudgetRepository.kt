package com.atoma.app.data.repository

import com.atoma.app.data.api.AtomaApi
import com.atoma.app.data.model.CreateTransactionRequest
import com.atoma.app.data.model.SaveSavingsGoalRequest
import com.atoma.app.data.model.TransactionDto
import com.atoma.app.domain.model.BudgetSummary
import com.atoma.app.domain.model.SavingsGoal
import com.atoma.app.domain.model.Transaction
import com.atoma.app.domain.model.TransactionType
import java.time.LocalDate
import java.time.YearMonth
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class BudgetRepository @Inject constructor(
    private val api: AtomaApi
) {
    suspend fun getTransactions(yearMonth: YearMonth? = null): Result<List<Transaction>> {
        return try {
            val response = api.getTransactions(
                year = yearMonth?.year,
                month = yearMonth?.monthValue
            )
            if (response.data != null) {
                Result.success(response.data.map { it.toDomain() })
            } else {
                Result.failure(Exception(response.error?.message ?: "Failed to get transactions"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun createTransaction(
        title: String,
        amount: Double,
        type: TransactionType,
        category: String,
        date: LocalDate
    ): Result<Transaction> {
        return try {
            val request = CreateTransactionRequest(
                title = title,
                amount = amount,
                type = type.name.lowercase(),
                category = category,
                date = date.toString()
            )
            val response = api.createTransaction(request)
            if (response.data != null) {
                Result.success(response.data.toDomain())
            } else {
                Result.failure(Exception(response.error?.message ?: "Failed to create transaction"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun deleteTransaction(transactionId: String): Result<Unit> {
        return try {
            val response = api.deleteTransaction(transactionId)
            if (response.error == null) {
                Result.success(Unit)
            } else {
                Result.failure(Exception(response.error.message))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun getSummary(yearMonth: YearMonth): Result<BudgetSummary> {
        return try {
            val response = api.getSummary(yearMonth.year, yearMonth.monthValue)
            if (response.data != null) {
                Result.success(
                    BudgetSummary(
                        income = response.data.income,
                        expenses = response.data.expenses
                    )
                )
            } else {
                Result.failure(Exception(response.error?.message ?: "Failed to get summary"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    // Savings Goal
    suspend fun getSavingsGoal(): Result<SavingsGoal?> {
        return try {
            val response = api.getSavingsGoal()
            if (response.data != null) {
                Result.success(
                    SavingsGoal(
                        id = response.data.id,
                        monthlyTarget = response.data.monthlyTarget
                    )
                )
            } else if (response.error?.code == "NOT_FOUND") {
                Result.success(null)
            } else {
                Result.failure(Exception(response.error?.message ?: "Failed to get savings goal"))
            }
        } catch (e: Exception) {
            // No goal set yet
            Result.success(null)
        }
    }

    suspend fun saveSavingsGoal(monthlyTarget: Double): Result<SavingsGoal> {
        return try {
            val request = SaveSavingsGoalRequest(monthlyTarget = monthlyTarget)
            val response = api.saveSavingsGoal(request)
            if (response.data != null) {
                Result.success(
                    SavingsGoal(
                        id = response.data.id,
                        monthlyTarget = response.data.monthlyTarget
                    )
                )
            } else {
                Result.failure(Exception(response.error?.message ?: "Failed to save savings goal"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun deleteSavingsGoal(): Result<Unit> {
        return try {
            val response = api.deleteSavingsGoal()
            if (response.error == null) {
                Result.success(Unit)
            } else {
                Result.failure(Exception(response.error.message))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    private fun TransactionDto.toDomain(): Transaction {
        return Transaction(
            id = id,
            title = title,
            amount = amount,
            type = TransactionType.valueOf(type.uppercase()),
            category = category,
            date = LocalDate.parse(date)
        )
    }
}

package com.atoma.app.data.sync

import com.atoma.app.data.network.NetworkConnectivityObserver
import com.atoma.app.data.network.NetworkStatus
import com.atoma.app.data.repository.HabitsRepository
import com.atoma.app.data.repository.TasksRepository
import com.atoma.app.data.repository.BudgetRepository
import com.atoma.app.data.repository.RecurringRepository
import com.atoma.app.data.repository.GoalsRepository
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject
import javax.inject.Singleton

data class SyncState(
    val isSyncing: Boolean = false,
    val lastSyncTime: Long? = null,
    val error: String? = null,
    val isOnline: Boolean = true
)

@Singleton
class SyncManager @Inject constructor(
    private val networkObserver: NetworkConnectivityObserver,
    private val habitsRepository: HabitsRepository,
    private val tasksRepository: TasksRepository,
    private val budgetRepository: BudgetRepository,
    private val recurringRepository: RecurringRepository,
    private val goalsRepository: GoalsRepository
) {
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.Main)

    private val _syncState = MutableStateFlow(SyncState())
    val syncState: StateFlow<SyncState> = _syncState

    private var hasInitialSync = false

    init {
        observeNetworkChanges()
    }

    private fun observeNetworkChanges() {
        scope.launch {
            networkObserver.observe().collect { status ->
                _syncState.update { it.copy(isOnline = status == NetworkStatus.Available) }

                when (status) {
                    NetworkStatus.Available -> {
                        // Sync when network becomes available
                        if (hasInitialSync) {
                            syncAll()
                        }
                    }
                    else -> { /* No action needed */ }
                }
            }
        }
    }

    suspend fun syncAll(): Result<Unit> {
        if (_syncState.value.isSyncing) {
            return Result.success(Unit)
        }

        _syncState.update { it.copy(isSyncing = true, error = null) }

        return try {
            // Sync all data in parallel
            val results = listOf(
                habitsRepository.getHabits(),
                tasksRepository.getTasks(),
                budgetRepository.getTransactions(),
                recurringRepository.getRecurringTransactions(),
                goalsRepository.getGoals()
            )

            val firstError = results.firstOrNull { it.isFailure }?.exceptionOrNull()

            if (firstError != null) {
                _syncState.update {
                    it.copy(
                        isSyncing = false,
                        error = firstError.message
                    )
                }
                Result.failure(firstError)
            } else {
                hasInitialSync = true
                _syncState.update {
                    it.copy(
                        isSyncing = false,
                        lastSyncTime = System.currentTimeMillis(),
                        error = null
                    )
                }
                Result.success(Unit)
            }
        } catch (e: Exception) {
            _syncState.update {
                it.copy(
                    isSyncing = false,
                    error = e.message
                )
            }
            Result.failure(e)
        }
    }

    suspend fun syncHabits(): Result<Unit> {
        return habitsRepository.getHabits().map { }
    }

    suspend fun syncTasks(): Result<Unit> {
        return tasksRepository.getTasks().map { }
    }

    suspend fun syncTransactions(): Result<Unit> {
        return budgetRepository.getTransactions().map { }
    }

    suspend fun syncRecurring(): Result<Unit> {
        return recurringRepository.getRecurringTransactions().map { }
    }

    suspend fun syncGoals(): Result<Unit> {
        return goalsRepository.getGoals().map { }
    }

    fun clearError() {
        _syncState.update { it.copy(error = null) }
    }
}

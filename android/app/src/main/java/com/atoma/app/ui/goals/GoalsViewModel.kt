package com.atoma.app.ui.goals

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.atoma.app.data.repository.GoalsRepository
import com.atoma.app.domain.model.Goal
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import java.time.LocalDate
import javax.inject.Inject

data class GoalsUiState(
    val goals: List<Goal> = emptyList(),
    val archivedGoals: List<Goal> = emptyList(),
    val isLoading: Boolean = false,
    val error: String? = null,
    val showAddDialog: Boolean = false,
    val showArchivedSheet: Boolean = false
)

@HiltViewModel
class GoalsViewModel @Inject constructor(
    private val repository: GoalsRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(GoalsUiState())
    val uiState: StateFlow<GoalsUiState> = _uiState

    init {
        loadGoals()
    }

    fun loadGoals() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            repository.getGoals()
                .onSuccess { allGoals ->
                    _uiState.update {
                        it.copy(
                            goals = allGoals.filter { goal -> goal.isActive },
                            archivedGoals = allGoals.filter { goal -> !goal.isActive },
                            isLoading = false
                        )
                    }
                }
                .onFailure { e ->
                    _uiState.update { it.copy(error = e.message, isLoading = false) }
                }
        }
    }

    fun createGoal(
        title: String,
        icon: String,
        targetValue: Int? = null,
        unit: String? = null,
        deadline: LocalDate? = null
    ) {
        viewModelScope.launch {
            repository.createGoal(title, icon, targetValue, unit, deadline)
                .onSuccess {
                    _uiState.update { it.copy(showAddDialog = false) }
                    loadGoals()
                }
                .onFailure { e ->
                    _uiState.update { it.copy(error = e.message) }
                }
        }
    }

    fun archiveGoal(goal: Goal) {
        viewModelScope.launch {
            // Optimistic update
            _uiState.update { state ->
                state.copy(
                    goals = state.goals.filter { it.id != goal.id },
                    archivedGoals = state.archivedGoals + goal.copy(archivedAt = java.time.LocalDateTime.now())
                )
            }

            repository.archiveGoal(goal.id)
                .onFailure {
                    loadGoals()
                }
        }
    }

    fun unarchiveGoal(goal: Goal) {
        viewModelScope.launch {
            // Optimistic update
            _uiState.update { state ->
                state.copy(
                    archivedGoals = state.archivedGoals.filter { it.id != goal.id },
                    goals = state.goals + goal.copy(archivedAt = null)
                )
            }

            repository.unarchiveGoal(goal.id)
                .onFailure {
                    loadGoals()
                }
        }
    }

    fun deleteGoal(goal: Goal) {
        viewModelScope.launch {
            // Optimistic update
            _uiState.update { state ->
                state.copy(
                    goals = state.goals.filter { it.id != goal.id },
                    archivedGoals = state.archivedGoals.filter { it.id != goal.id }
                )
            }

            repository.deleteGoal(goal.id)
                .onFailure {
                    loadGoals()
                }
        }
    }

    fun showAddDialog() {
        _uiState.update { it.copy(showAddDialog = true) }
    }

    fun hideAddDialog() {
        _uiState.update { it.copy(showAddDialog = false) }
    }

    fun showArchivedSheet() {
        _uiState.update { it.copy(showArchivedSheet = true) }
    }

    fun hideArchivedSheet() {
        _uiState.update { it.copy(showArchivedSheet = false) }
    }

    fun clearError() {
        _uiState.update { it.copy(error = null) }
    }
}

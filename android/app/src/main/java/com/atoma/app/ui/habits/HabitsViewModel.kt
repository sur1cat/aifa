package com.atoma.app.ui.habits

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.atoma.app.data.repository.HabitsRepository
import com.atoma.app.domain.model.Habit
import com.atoma.app.domain.model.HabitPeriod
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import java.time.LocalDate
import javax.inject.Inject

data class HabitsUiState(
    val allHabits: List<Habit> = emptyList(),
    val isLoading: Boolean = false,
    val error: String? = null,
    val showAddDialog: Boolean = false,
    val showArchivedSheet: Boolean = false
) {
    val habits: List<Habit>
        get() = allHabits.filter { it.isActive }

    val archivedHabits: List<Habit>
        get() = allHabits.filter { !it.isActive }
}

@HiltViewModel
class HabitsViewModel @Inject constructor(
    private val repository: HabitsRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(HabitsUiState())
    val uiState: StateFlow<HabitsUiState> = _uiState

    init {
        loadHabits()
    }

    fun loadHabits() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            repository.getHabits()
                .onSuccess { habits ->
                    _uiState.update { it.copy(allHabits = habits, isLoading = false) }
                }
                .onFailure { e ->
                    _uiState.update { it.copy(error = e.message, isLoading = false) }
                }
        }
    }

    fun toggleHabit(habit: Habit, date: LocalDate = LocalDate.now()) {
        viewModelScope.launch {
            val dateStr = date.toString()
            val isCompletedOnDate = habit.completedDates.contains(dateStr)

            // Optimistic update
            val updatedHabits = _uiState.value.allHabits.map {
                if (it.id == habit.id) {
                    val newCompletedDates = if (isCompletedOnDate) {
                        it.completedDates - dateStr
                    } else {
                        it.completedDates + dateStr
                    }
                    it.copy(completedDates = newCompletedDates)
                } else it
            }
            _uiState.update { it.copy(allHabits = updatedHabits) }

            repository.toggleHabit(habit.id, date)
                .onFailure {
                    // Revert on failure
                    loadHabits()
                }
        }
    }

    fun createHabit(title: String, icon: String, color: String, period: HabitPeriod) {
        viewModelScope.launch {
            repository.createHabit(title, icon, color, period)
                .onSuccess {
                    _uiState.update { it.copy(showAddDialog = false) }
                    loadHabits()
                }
                .onFailure { e ->
                    _uiState.update { it.copy(error = e.message) }
                }
        }
    }

    fun deleteHabit(habit: Habit) {
        viewModelScope.launch {
            // Optimistic update
            _uiState.update { state ->
                state.copy(allHabits = state.allHabits.filter { it.id != habit.id })
            }

            repository.deleteHabit(habit.id)
                .onFailure {
                    loadHabits()
                }
        }
    }

    fun archiveHabit(habit: Habit) {
        viewModelScope.launch {
            // Optimistic update - move to archived
            val updatedHabits = _uiState.value.allHabits.map {
                if (it.id == habit.id) {
                    it.copy(archivedAt = java.time.LocalDateTime.now())
                } else it
            }
            _uiState.update { it.copy(allHabits = updatedHabits) }

            repository.archiveHabit(habit.id)
                .onFailure {
                    loadHabits()
                }
        }
    }

    fun unarchiveHabit(habit: Habit) {
        viewModelScope.launch {
            // Optimistic update - move back to active
            val updatedHabits = _uiState.value.allHabits.map {
                if (it.id == habit.id) {
                    it.copy(archivedAt = null)
                } else it
            }
            _uiState.update { it.copy(allHabits = updatedHabits) }

            repository.unarchiveHabit(habit.id)
                .onFailure {
                    loadHabits()
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

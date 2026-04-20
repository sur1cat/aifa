package com.atoma.app.ui.tasks

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.atoma.app.data.repository.TasksRepository
import com.atoma.app.domain.model.DailyTask
import com.atoma.app.domain.model.TaskPriority
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import java.time.LocalDate
import javax.inject.Inject

data class TasksUiState(
    val tasks: List<DailyTask> = emptyList(),
    val isLoading: Boolean = false,
    val error: String? = null,
    val showAddDialog: Boolean = false,
    val selectedDate: LocalDate = LocalDate.now()
)

@HiltViewModel
class TasksViewModel @Inject constructor(
    private val repository: TasksRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(TasksUiState())
    val uiState: StateFlow<TasksUiState> = _uiState

    init {
        loadTasks()
    }

    fun loadTasks() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            repository.getTasks(_uiState.value.selectedDate)
                .onSuccess { tasks ->
                    _uiState.update { it.copy(tasks = tasks.sortedBy { t -> t.isCompleted }, isLoading = false) }
                }
                .onFailure { e ->
                    _uiState.update { it.copy(error = e.message, isLoading = false) }
                }
        }
    }

    fun toggleTask(task: DailyTask) {
        viewModelScope.launch {
            // Optimistic update
            val updatedTasks = _uiState.value.tasks.map {
                if (it.id == task.id) it.copy(isCompleted = !it.isCompleted) else it
            }.sortedBy { it.isCompleted }
            _uiState.update { it.copy(tasks = updatedTasks) }

            repository.toggleTask(task.id)
                .onFailure { loadTasks() }
        }
    }

    fun createTask(title: String, priority: TaskPriority, dueDate: LocalDate = _uiState.value.selectedDate) {
        viewModelScope.launch {
            repository.createTask(title, priority, dueDate)
                .onSuccess {
                    _uiState.update { it.copy(showAddDialog = false) }
                    loadTasks()
                }
                .onFailure { e ->
                    _uiState.update { it.copy(error = e.message) }
                }
        }
    }

    fun deleteTask(task: DailyTask) {
        viewModelScope.launch {
            _uiState.update { state ->
                state.copy(tasks = state.tasks.filter { it.id != task.id })
            }
            repository.deleteTask(task.id)
                .onFailure { loadTasks() }
        }
    }

    fun setDate(date: LocalDate) {
        _uiState.update { it.copy(selectedDate = date) }
        loadTasks()
    }

    fun showAddDialog() {
        _uiState.update { it.copy(showAddDialog = true) }
    }

    fun hideAddDialog() {
        _uiState.update { it.copy(showAddDialog = false) }
    }
}

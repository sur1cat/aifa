package com.atoma.app.ui.ai

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.atoma.app.data.repository.AIAgent
import com.atoma.app.data.repository.AIRepository
import com.atoma.app.data.repository.HabitsRepository
import com.atoma.app.data.repository.TasksRepository
import com.atoma.app.data.repository.BudgetRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class ChatMessage(
    val content: String,
    val isUser: Boolean,
    val timestamp: Long = System.currentTimeMillis()
)

data class AIChatUiState(
    val messages: List<ChatMessage> = emptyList(),
    val isLoading: Boolean = false,
    val error: String? = null,
    val currentAgent: AIAgent = AIAgent.LIFE_COACH,
    val inputText: String = ""
)

@HiltViewModel
class AIChatViewModel @Inject constructor(
    private val aiRepository: AIRepository,
    private val habitsRepository: HabitsRepository,
    private val tasksRepository: TasksRepository,
    private val budgetRepository: BudgetRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(AIChatUiState())
    val uiState: StateFlow<AIChatUiState> = _uiState

    fun setAgent(agent: AIAgent) {
        _uiState.update { it.copy(currentAgent = agent, messages = emptyList()) }
    }

    fun updateInputText(text: String) {
        _uiState.update { it.copy(inputText = text) }
    }

    fun sendMessage() {
        val message = _uiState.value.inputText.trim()
        if (message.isBlank()) return

        // Add user message
        val userMessage = ChatMessage(content = message, isUser = true)
        _uiState.update {
            it.copy(
                messages = it.messages + userMessage,
                inputText = "",
                isLoading = true,
                error = null
            )
        }

        viewModelScope.launch {
            // Gather context based on agent
            val habits = when (_uiState.value.currentAgent) {
                AIAgent.HABIT_COACH, AIAgent.LIFE_COACH -> {
                    habitsRepository.getHabits().getOrNull() ?: emptyList()
                }
                else -> emptyList()
            }

            val tasks = when (_uiState.value.currentAgent) {
                AIAgent.TASK_ASSISTANT, AIAgent.LIFE_COACH -> {
                    tasksRepository.getTasks().getOrNull() ?: emptyList()
                }
                else -> emptyList()
            }

            val transactions = when (_uiState.value.currentAgent) {
                AIAgent.FINANCE_ADVISOR, AIAgent.LIFE_COACH -> {
                    budgetRepository.getTransactions().getOrNull() ?: emptyList()
                }
                else -> emptyList()
            }

            aiRepository.chat(
                message = message,
                agent = _uiState.value.currentAgent,
                habits = habits,
                tasks = tasks,
                transactions = transactions
            ).onSuccess { response ->
                val aiMessage = ChatMessage(content = response, isUser = false)
                _uiState.update {
                    it.copy(
                        messages = it.messages + aiMessage,
                        isLoading = false
                    )
                }
            }.onFailure { e ->
                _uiState.update {
                    it.copy(
                        isLoading = false,
                        error = e.message ?: "Failed to get AI response"
                    )
                }
            }
        }
    }

    fun clearError() {
        _uiState.update { it.copy(error = null) }
    }

    fun clearChat() {
        _uiState.update { it.copy(messages = emptyList()) }
    }
}

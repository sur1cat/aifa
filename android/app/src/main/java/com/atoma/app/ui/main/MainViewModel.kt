package com.atoma.app.ui.main

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.atoma.app.data.sync.SyncManager
import com.atoma.app.data.sync.SyncState
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class MainViewModel @Inject constructor(
    private val syncManager: SyncManager
) : ViewModel() {

    val syncState: StateFlow<SyncState> = syncManager.syncState
        .stateIn(
            scope = viewModelScope,
            started = SharingStarted.WhileSubscribed(5000),
            initialValue = SyncState()
        )

    init {
        // Initial sync on app start
        syncAll()
    }

    fun syncAll() {
        viewModelScope.launch {
            syncManager.syncAll()
        }
    }

    fun clearError() {
        syncManager.clearError()
    }
}

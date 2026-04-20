package com.atoma.app.data.local

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.booleanPreferencesKey
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map
import javax.inject.Inject
import javax.inject.Singleton

private val Context.dataStore: DataStore<Preferences> by preferencesDataStore(name = "atoma_prefs")

@Singleton
class TokenManager @Inject constructor(
    @ApplicationContext private val context: Context
) {
    companion object {
        private val ACCESS_TOKEN_KEY = stringPreferencesKey("access_token")
        private val REFRESH_TOKEN_KEY = stringPreferencesKey("refresh_token")
        private val USER_ID_KEY = stringPreferencesKey("user_id")
        private val USER_NAME_KEY = stringPreferencesKey("user_name")
        private val USER_EMAIL_KEY = stringPreferencesKey("user_email")
        private val USER_AVATAR_KEY = stringPreferencesKey("user_avatar")
        private val HAS_COMPLETED_ONBOARDING_KEY = booleanPreferencesKey("has_completed_onboarding")
    }

    // Cached token for synchronous access in interceptor
    @Volatile
    private var cachedAccessToken: String? = null

    val accessToken: Flow<String?> = context.dataStore.data.map { prefs ->
        prefs[ACCESS_TOKEN_KEY].also { cachedAccessToken = it }
    }

    // Synchronous access for interceptor - returns cached value
    fun getAccessTokenSync(): String? = cachedAccessToken

    val refreshToken: Flow<String?> = context.dataStore.data.map { prefs ->
        prefs[REFRESH_TOKEN_KEY]
    }

    val isLoggedIn: Flow<Boolean> = accessToken.map { !it.isNullOrEmpty() }

    val userName: Flow<String?> = context.dataStore.data.map { prefs ->
        prefs[USER_NAME_KEY]
    }

    val userEmail: Flow<String?> = context.dataStore.data.map { prefs ->
        prefs[USER_EMAIL_KEY]
    }

    val userAvatar: Flow<String?> = context.dataStore.data.map { prefs ->
        prefs[USER_AVATAR_KEY]
    }

    val hasCompletedOnboarding: Flow<Boolean> = context.dataStore.data.map { prefs ->
        prefs[HAS_COMPLETED_ONBOARDING_KEY] ?: false
    }

    suspend fun setOnboardingCompleted() {
        context.dataStore.edit { prefs ->
            prefs[HAS_COMPLETED_ONBOARDING_KEY] = true
        }
    }

    suspend fun saveTokens(accessToken: String, refreshToken: String) {
        cachedAccessToken = accessToken // Update cache immediately
        context.dataStore.edit { prefs ->
            prefs[ACCESS_TOKEN_KEY] = accessToken
            prefs[REFRESH_TOKEN_KEY] = refreshToken
        }
    }

    suspend fun saveUser(id: String, name: String, email: String, avatarUrl: String?) {
        context.dataStore.edit { prefs ->
            prefs[USER_ID_KEY] = id
            prefs[USER_NAME_KEY] = name
            prefs[USER_EMAIL_KEY] = email
            avatarUrl?.let { prefs[USER_AVATAR_KEY] = it }
        }
    }

    suspend fun clearAll() {
        cachedAccessToken = null // Clear cache
        context.dataStore.edit { it.clear() }
    }
}

package com.atoma.app.data.repository

import com.atoma.app.data.api.AtomaApi
import com.atoma.app.data.local.TokenManager
import com.atoma.app.data.model.GoogleAuthRequest
import com.atoma.app.domain.model.User
import kotlinx.coroutines.flow.Flow
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class AuthRepository @Inject constructor(
    private val api: AtomaApi,
    private val tokenManager: TokenManager
) {
    val isLoggedIn: Flow<Boolean> = tokenManager.isLoggedIn

    suspend fun signInWithGoogle(idToken: String): Result<User> {
        return try {
            val response = api.googleAuth(GoogleAuthRequest(idToken))
            if (response.data != null) {
                val authData = response.data
                tokenManager.saveTokens(authData.tokens.accessToken, authData.tokens.refreshToken)
                tokenManager.saveUser(
                    id = authData.user.id,
                    name = authData.user.name,
                    email = authData.user.email,
                    avatarUrl = authData.user.avatarUrl
                )
                Result.success(
                    User(
                        id = authData.user.id,
                        email = authData.user.email,
                        name = authData.user.name,
                        avatarUrl = authData.user.avatarUrl,
                        isPremium = authData.user.isPremium
                    )
                )
            } else {
                Result.failure(Exception(response.error?.message ?: "Auth failed"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun signOut() {
        tokenManager.clearAll()
    }

    suspend fun deleteAccount(): Result<Unit> {
        return try {
            val response = api.deleteAccount()
            if (response.error == null) {
                tokenManager.clearAll()
                Result.success(Unit)
            } else {
                Result.failure(Exception(response.error.message ?: "Delete failed"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun getCurrentUser(): Result<User> {
        return try {
            val response = api.getCurrentUser()
            if (response.data != null) {
                val user = response.data
                Result.success(
                    User(
                        id = user.id,
                        email = user.email,
                        name = user.name,
                        avatarUrl = user.avatarUrl,
                        isPremium = user.isPremium
                    )
                )
            } else {
                Result.failure(Exception(response.error?.message ?: "Failed to get user"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}

import Foundation
import SwiftUI
import Combine
import GoogleSignIn
import AuthenticationServices
import os

@MainActor
class AuthManager: ObservableObject {
    static let shared = AuthManager()

    @Published var isAuthenticated = false
    @Published var currentUser: User?
    @Published var isLoading = false
    @Published var error: String?

    private let api = APIClient.shared
    private let keychain = KeychainHelper.shared

    private init() {
        Task { await checkAuthStatus() }
    }

    func checkAuthStatus() async {
        guard let tokens = keychain.getTokens() else {
            isAuthenticated = false
            return
        }
        await api.setAccessToken(tokens.accessToken)
        do {
            let user: User = try await api.request(endpoint: "auth/me", requiresAuth: true)
            self.currentUser = user
            self.isAuthenticated = true
        } catch {
            isAuthenticated = false
        }
    }

    func signInWithGoogle() async {
        isLoading = true
        error = nil

        AppLogger.auth.info("Starting Google Sign-In...")

        guard let windowScene = UIApplication.shared.connectedScenes.first as? UIWindowScene,
              let rootVC = windowScene.windows.first?.rootViewController else {
            error = "Cannot get root view controller"
            AppLogger.auth.error("Cannot get root view controller")
            isLoading = false
            return
        }

        do {
            AppLogger.auth.info("Presenting Google Sign-In...")
            let result = try await GIDSignIn.sharedInstance.signIn(withPresenting: rootVC)
            AppLogger.auth.info("Google Sign-In completed, user: \(result.user.profile?.email ?? "unknown")")

            guard let idToken = result.user.idToken?.tokenString else {
                error = "Failed to get Google ID token"
                AppLogger.auth.error("No ID token received")
                isLoading = false
                return
            }

            AppLogger.auth.info("Got ID token, sending to backend...")
            let request = GoogleSignInRequest(idToken: idToken)
            let response: AuthResponse = try await api.request(
                endpoint: "auth/google",
                method: "POST",
                body: request
            )
            AppLogger.auth.info("Backend auth successful, user: \(response.user.email ?? "unknown")")
            await handleAuthSuccess(response)
        } catch let gidError as GIDSignInError {
            AppLogger.auth.error("Google Sign-In error: \(gidError.localizedDescription), code: \(gidError.code.rawValue)")
            if gidError.code == .canceled {
                self.error = nil // User cancelled, don't show error
            } else {
                self.error = gidError.localizedDescription
            }
        } catch {
            AppLogger.auth.error("Auth error: \(error.localizedDescription)")
            self.error = error.localizedDescription
        }
        isLoading = false
    }

    func handleAppleSignIn(_ authorization: ASAuthorization) async {
        isLoading = true
        error = nil

        guard let credential = authorization.credential as? ASAuthorizationAppleIDCredential,
              let identityToken = credential.identityToken,
              let idToken = String(data: identityToken, encoding: .utf8) else {
            error = "Failed to get Apple ID token"
            isLoading = false
            return
        }

        do {
            let request = AppleSignInRequest(idToken: idToken, user: nil)
            let response: AuthResponse = try await api.request(
                endpoint: "auth/apple",
                method: "POST",
                body: request
            )
            await handleAuthSuccess(response)
        } catch {
            self.error = error.localizedDescription
        }
        isLoading = false
    }

    func signOut() async {
        // Invalidate refresh token on backend (best effort)
        if let tokens = keychain.getTokens() {
            let request = LogoutRequest(refreshToken: tokens.refreshToken)
            do {
                let _: EmptyResponse = try await api.request(
                    endpoint: "auth/logout",
                    method: "POST",
                    body: request,
                    requiresAuth: true
                )
            } catch {
                AppLogger.auth.warning("Logout request failed (continuing with local logout): \(error.localizedDescription)")
            }
        }

        // Clear all local data
        DataManager.shared.clearAllData()
        keychain.deleteTokens()
        await api.setAccessToken(nil)
        GIDSignIn.sharedInstance.signOut()
        currentUser = nil
        isAuthenticated = false
    }

    func deleteAccount() async throws {
        // Call backend to delete account
        let _: EmptyResponse = try await api.request(
            endpoint: "auth/account",
            method: "DELETE",
            requiresAuth: true
        )
        // Sign out after successful deletion
        await signOut()
    }

    private func handleAuthSuccess(_ response: AuthResponse) async {
        keychain.saveTokens(response.tokens)
        await api.setAccessToken(response.tokens.accessToken)
        currentUser = response.user
        isAuthenticated = true
    }
}

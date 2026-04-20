import Foundation
import os

enum APIError: LocalizedError {
    case invalidURL
    case invalidResponse
    case serverError(String)
    case unauthorized
    case networkError(Error)

    var errorDescription: String? {
        switch self {
        case .invalidURL: return "Invalid URL"
        case .invalidResponse: return "Invalid response"
        case .serverError(let msg): return msg
        case .unauthorized: return "Unauthorized"
        case .networkError(let error): return "Network error: \(error.localizedDescription)"
        }
    }

    var isRetryable: Bool {
        switch self {
        case .networkError: return true
        case .serverError: return false // 5xx errors could be retried, but keeping it simple
        default: return false
        }
    }
}

struct APIResponse<T: Decodable>: Decodable {
    let data: T?
    let error: APIErrorResponse?
}

struct APIErrorResponse: Decodable, Sendable {
    let code: String
    let message: String
}

actor APIClient {
    static let shared = APIClient()

    private let baseURL = URL(string: "http://localhost:8080/api/v1/")!
    private var accessToken: String?
    private var refreshTask: Task<Bool, Never>?

    // Retry configuration
    private let maxRetries = 3
    private let baseRetryDelay: UInt64 = 1_000_000_000 // 1 second in nanoseconds

    func setAccessToken(_ token: String?) {
        self.accessToken = token
        AppLogger.network.debug("Token set: \(token != nil ? "YES" : "nil")")
    }

    func hasToken() -> Bool {
        return accessToken != nil
    }

    func request<T: Decodable>(
        endpoint: String,
        method: String = "GET",
        body: Encodable? = nil,
        requiresAuth: Bool = false
    ) async throws -> T {
        var lastError: Error?

        for attempt in 0..<maxRetries {
            do {
                return try await performRequest(endpoint: endpoint, method: method, body: body, requiresAuth: requiresAuth, isRetry: false)
            } catch let error as APIError where error.isRetryable {
                lastError = error
                let delay = baseRetryDelay * UInt64(1 << attempt) // Exponential: 1s, 2s, 4s
                AppLogger.network.warning("Retry \(attempt + 1)/\(self.maxRetries) for \(endpoint) after \(delay / 1_000_000_000)s")
                try? await Task.sleep(nanoseconds: delay)
            } catch {
                throw error // Non-retryable errors are thrown immediately
            }
        }

        throw lastError ?? APIError.networkError(NSError(domain: "APIClient", code: -1, userInfo: [NSLocalizedDescriptionKey: "Max retries exceeded"]))
    }

    func requestOptional<T: Decodable>(
        endpoint: String,
        method: String = "GET",
        body: Encodable? = nil,
        requiresAuth: Bool = false
    ) async throws -> T? {
        guard let url = URL(string: endpoint, relativeTo: baseURL) else {
            throw APIError.invalidURL
        }

        var request = URLRequest(url: url)
        request.httpMethod = method
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        if requiresAuth {
            var token = accessToken

            if token == nil, let tokens = KeychainHelper.shared.getTokens() {
                token = tokens.accessToken
                self.accessToken = token
                AppLogger.network.debug("Token loaded from Keychain")
            }

            if let token = token {
                request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
            }
        }

        if let body = body {
            request.httpBody = try JSONEncoder().encode(body)
        }

        let data: Data
        let response: URLResponse
        do {
            (data, response) = try await URLSession.shared.data(for: request)
        } catch {
            throw APIError.networkError(error)
        }

        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError.invalidResponse
        }

        if httpResponse.statusCode == 401 {
            if requiresAuth {
                if await refreshTokenIfNeeded() {
                    return try await requestOptional(
                        endpoint: endpoint,
                        method: method,
                        body: body,
                        requiresAuth: requiresAuth
                    )
                }
            }
            throw APIError.unauthorized
        }

        let apiResponse = try JSONDecoder().decode(APIResponse<T>.self, from: data)

        if let error = apiResponse.error {
            throw APIError.serverError(error.message)
        }

        return apiResponse.data
    }

    private func performRequest<T: Decodable>(
        endpoint: String,
        method: String,
        body: Encodable?,
        requiresAuth: Bool,
        isRetry: Bool
    ) async throws -> T {
        guard let url = URL(string: endpoint, relativeTo: baseURL) else {
            throw APIError.invalidURL
        }

        var request = URLRequest(url: url)
        request.httpMethod = method
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        if requiresAuth {
            var token = accessToken

            // Fallback: try to get token from Keychain if not in memory
            if token == nil {
                if let tokens = KeychainHelper.shared.getTokens() {
                    token = tokens.accessToken
                    self.accessToken = token // Cache it
                    AppLogger.network.debug("Token loaded from Keychain")
                }
            }

            if let token = token {
                request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
                AppLogger.network.debug("\(method) \(endpoint) - with token")
            } else {
                AppLogger.network.warning("\(method) \(endpoint) - NO TOKEN! Request will fail")
            }
        }

        if let body = body {
            request.httpBody = try JSONEncoder().encode(body)
        }

        let data: Data
        let response: URLResponse
        do {
            (data, response) = try await URLSession.shared.data(for: request)
        } catch {
            // Wrap URLSession errors as retryable network errors
            throw APIError.networkError(error)
        }

        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError.invalidResponse
        }

        if httpResponse.statusCode == 401 {
            // Try to refresh token if this is not already a retry
            if requiresAuth && !isRetry {
                AppLogger.network.info("Got 401, attempting token refresh...")
                if await refreshTokenIfNeeded() {
                    // Retry the original request with new token
                    return try await performRequest(endpoint: endpoint, method: method, body: body, requiresAuth: requiresAuth, isRetry: true)
                }
            }
            throw APIError.unauthorized
        }

        let apiResponse = try JSONDecoder().decode(APIResponse<T>.self, from: data)

        if let error = apiResponse.error {
            throw APIError.serverError(error.message)
        }

        guard let responseData = apiResponse.data else {
            throw APIError.invalidResponse
        }

        return responseData
    }

    /// Refresh token with protection against concurrent calls
    /// All concurrent callers will wait on the same refresh task
    private func refreshTokenIfNeeded() async -> Bool {
        // If a refresh is already in progress, wait for it
        if let existingTask = refreshTask {
            return await existingTask.value
        }

        // Create a new refresh task
        let task = Task<Bool, Never> { [weak self] in
            guard let self = self else { return false }
            return await self.performTokenRefresh()
        }

        refreshTask = task
        let result = await task.value
        refreshTask = nil

        return result
    }

    private func performTokenRefresh() async -> Bool {
        guard let tokens = KeychainHelper.shared.getTokens() else {
            AppLogger.network.warning("No refresh token available")
            return false
        }

        guard let url = URL(string: "auth/refresh", relativeTo: baseURL) else {
            return false
        }

        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        struct RefreshRequest: Encodable {
            let refreshToken: String
            enum CodingKeys: String, CodingKey {
                case refreshToken = "refresh_token"
            }
        }

        do {
            request.httpBody = try JSONEncoder().encode(RefreshRequest(refreshToken: tokens.refreshToken))
            let (data, response) = try await URLSession.shared.data(for: request)

            guard let httpResponse = response as? HTTPURLResponse,
                  httpResponse.statusCode == 200 else {
                AppLogger.network.error("Token refresh failed with status: \((response as? HTTPURLResponse)?.statusCode ?? 0)")
                return false
            }

            let apiResponse = try JSONDecoder().decode(APIResponse<TokenPair>.self, from: data)

            guard let newTokens = apiResponse.data else {
                AppLogger.network.error("Token refresh returned no data")
                return false
            }

            // Save new tokens
            KeychainHelper.shared.saveTokens(newTokens)
            self.accessToken = newTokens.accessToken
            AppLogger.network.info("Token refreshed successfully")
            return true
        } catch {
            AppLogger.network.error("Token refresh error: \(error.localizedDescription)")
            return false
        }
    }
}

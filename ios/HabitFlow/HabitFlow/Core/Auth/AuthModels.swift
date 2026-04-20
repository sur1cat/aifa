import Foundation

struct User: Codable, Identifiable, Sendable {
    let id: String
    let email: String?
    let phone: String?
    let name: String?
    let avatarURL: String?
    let authProvider: String
    let createdAt: String

    enum CodingKeys: String, CodingKey {
        case id, email, phone, name
        case avatarURL = "avatar_url"
        case authProvider = "auth_provider"
        case createdAt = "created_at"
    }

    var displayName: String {
        name ?? email ?? phone ?? "User"
    }

    var providerDisplayName: String {
        switch authProvider {
        case "google": return "Google Account"
        case "apple": return "Apple ID"
        case "phone": return "Phone"
        default: return authProvider.capitalized
        }
    }

    var providerIcon: String {
        switch authProvider {
        case "google": return "g.circle.fill"
        case "apple": return "apple.logo"
        case "phone": return "phone.fill"
        default: return "person.fill"
        }
    }
}

struct TokenPair: Codable, Sendable {
    let accessToken: String
    let refreshToken: String
    let expiresAt: Int

    enum CodingKeys: String, CodingKey {
        case accessToken = "access_token"
        case refreshToken = "refresh_token"
        case expiresAt = "expires_at"
    }
}

struct AuthResponse: Codable, Sendable {
    let user: User
    let tokens: TokenPair
    let isNewUser: Bool

    enum CodingKeys: String, CodingKey {
        case user, tokens
        case isNewUser = "is_new_user"
    }
}

struct GoogleSignInRequest: Encodable, Sendable {
    let idToken: String
    enum CodingKeys: String, CodingKey {
        case idToken = "id_token"
    }
}

struct AppleSignInRequest: Encodable, Sendable {
    let idToken: String
    let user: String?
    enum CodingKeys: String, CodingKey {
        case idToken = "id_token"
        case user
    }
}

struct LogoutRequest: Encodable, Sendable {
    let refreshToken: String
    enum CodingKeys: String, CodingKey {
        case refreshToken = "refresh_token"
    }
}

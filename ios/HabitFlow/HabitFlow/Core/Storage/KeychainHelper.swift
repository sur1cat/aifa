import Foundation
import Security
import os

class KeychainHelper {
    static let shared = KeychainHelper()
    private let service = "com.atoma.app"

    func saveTokens(_ tokens: TokenPair) {
        // Use safe encoding (UTF-8 encoding should never fail for strings, but avoid force unwrap)
        guard let accessData = tokens.accessToken.data(using: .utf8),
              let refreshData = tokens.refreshToken.data(using: .utf8),
              let expiresData = String(tokens.expiresAt).data(using: .utf8) else {
            AppLogger.storage.error("Failed to encode token data")
            return
        }
        save(key: "accessToken", data: accessData)
        save(key: "refreshToken", data: refreshData)
        save(key: "expiresAt", data: expiresData)
    }

    func getTokens() -> TokenPair? {
        guard let accessData = load(key: "accessToken"),
              let refreshData = load(key: "refreshToken"),
              let access = String(data: accessData, encoding: .utf8),
              let refresh = String(data: refreshData, encoding: .utf8) else {
            return nil
        }
        var expiresAt = 0
        if let expiresData = load(key: "expiresAt"),
           let expiresStr = String(data: expiresData, encoding: .utf8),
           let expires = Int(expiresStr) {
            expiresAt = expires
        }
        return TokenPair(accessToken: access, refreshToken: refresh, expiresAt: expiresAt)
    }

    func deleteTokens() {
        delete(key: "accessToken")
        delete(key: "refreshToken")
        delete(key: "expiresAt")
    }

    private func save(key: String, data: Data) {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key,
            kSecValueData as String: data
        ]

        // Delete existing item first
        let deleteStatus = SecItemDelete(query as CFDictionary)
        if deleteStatus != errSecSuccess && deleteStatus != errSecItemNotFound {
            AppLogger.storage.warning("Failed to delete existing key '\(key)': \(self.keychainErrorMessage(deleteStatus))")
        }

        // Add new item
        let addStatus = SecItemAdd(query as CFDictionary, nil)
        if addStatus != errSecSuccess {
            AppLogger.storage.error("Failed to save key '\(key)': \(self.keychainErrorMessage(addStatus))")
        }
    }

    private func load(key: String) -> Data? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key,
            kSecReturnData as String: true
        ]

        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)

        switch status {
        case errSecSuccess:
            return result as? Data
        case errSecItemNotFound:
            // Normal case - key doesn't exist
            return nil
        default:
            AppLogger.storage.error("Failed to load key '\(key)': \(self.keychainErrorMessage(status))")
            return nil
        }
    }

    private func delete(key: String) {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key
        ]

        let status = SecItemDelete(query as CFDictionary)
        if status != errSecSuccess && status != errSecItemNotFound {
            AppLogger.storage.error("Failed to delete key '\(key)': \(self.keychainErrorMessage(status))")
        }
    }

    /// Convert OSStatus to human-readable error message
    private func keychainErrorMessage(_ status: OSStatus) -> String {
        switch status {
        case errSecSuccess: return "Success"
        case errSecItemNotFound: return "Item not found"
        case errSecDuplicateItem: return "Duplicate item"
        case errSecAuthFailed: return "Authentication failed"
        case errSecInteractionNotAllowed: return "Interaction not allowed (device locked?)"
        case errSecDecode: return "Decode error"
        case errSecParam: return "Invalid parameter"
        case errSecAllocate: return "Allocation error"
        case errSecNotAvailable: return "Keychain not available"
        case errSecDataNotAvailable: return "Data not available"
        case errSecDataNotModifiable: return "Data not modifiable"
        case errSecNoAccessForItem: return "No access for item"
        case errSecMissingEntitlement: return "Missing entitlement"
        default: return "Unknown error (code: \(status))"
        }
    }
}

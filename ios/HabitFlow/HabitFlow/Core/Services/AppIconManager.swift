import SwiftUI
import UIKit
import Combine
import os

@MainActor
class AppIconManager: ObservableObject {
    static let shared = AppIconManager()

    @Published var currentIcon: AppIcon = .default

    // Available app icons
    enum AppIcon: String, CaseIterable, Identifiable {
        case `default` = "AppIcon"
        case dark = "AppIcon-Dark"
        case gradient = "AppIcon-Gradient"
        case minimal = "AppIcon-Minimal"

        var id: String { rawValue }

        var displayName: String {
            switch self {
            case .default: return "Default"
            case .dark: return "Dark"
            case .gradient: return "Gradient"
            case .minimal: return "Minimal"
            }
        }

        var iconName: String? {
            self == .default ? nil : rawValue
        }

        var previewColor: Color {
            switch self {
            case .default: return Color.hf.accent
            case .dark: return .black
            case .gradient: return .purple
            case .minimal: return .gray
            }
        }

        var previewGradient: LinearGradient {
            switch self {
            case .default:
                return LinearGradient(colors: [Color.hf.accent, Color.hf.accent.opacity(0.7)], startPoint: .topLeading, endPoint: .bottomTrailing)
            case .dark:
                return LinearGradient(colors: [.black, Color(white: 0.2)], startPoint: .topLeading, endPoint: .bottomTrailing)
            case .gradient:
                return LinearGradient(colors: [.purple, .pink, .orange], startPoint: .topLeading, endPoint: .bottomTrailing)
            case .minimal:
                return LinearGradient(colors: [Color(white: 0.95), Color(white: 0.85)], startPoint: .topLeading, endPoint: .bottomTrailing)
            }
        }
    }

    private init() {
        loadCurrentIcon()
    }

    private func loadCurrentIcon() {
        if let iconName = UIApplication.shared.alternateIconName {
            currentIcon = AppIcon(rawValue: iconName) ?? .default
        } else {
            currentIcon = .default
        }
    }

    func setIcon(_ icon: AppIcon) async {
        guard UIApplication.shared.supportsAlternateIcons else { return }

        do {
            try await UIApplication.shared.setAlternateIconName(icon.iconName)
            currentIcon = icon
        } catch {
            AppLogger.app.error("Failed to set app icon: \(error.localizedDescription)")
        }
    }
}


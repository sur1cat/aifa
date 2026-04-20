import SwiftUI

extension Color {
    static let hf = HFColors()
}

struct HFColors {
    // MARK: - Brand
    let accent = Color(hex: "4CAF50")
    let accentLight = Color(hex: "81C784")
    let accentDark = Color(hex: "388E3C")

    // MARK: - Backgrounds (system adaptive)
    let background = Color(uiColor: .systemBackground)
    let surface = Color(uiColor: .secondarySystemBackground)
    let groupedBackground = Color(uiColor: .systemGroupedBackground)

    // MARK: - Text (system adaptive)
    let textPrimary = Color(uiColor: .label)
    let textSecondary = Color(uiColor: .secondaryLabel)
    let textTertiary = Color(uiColor: .tertiaryLabel)

    // MARK: - Semantic Colors (adaptive for dark mode)
    let income = Color.adaptive(light: "2E7D32", dark: "66BB6A")
    let expense = Color.adaptive(light: "C62828", dark: "EF5350")
    let warning = Color.adaptive(light: "E65100", dark: "FF9800")
    let info = Color.adaptive(light: "1565C0", dark: "42A5F5")

    // MARK: - Priority Colors (adaptive)
    let priorityLow = Color.adaptive(light: "757575", dark: "9E9E9E")
    let priorityMedium = Color.adaptive(light: "1565C0", dark: "42A5F5")
    let priorityHigh = Color.adaptive(light: "E65100", dark: "FF9800")
    let priorityUrgent = Color.adaptive(light: "C62828", dark: "EF5350")

    // MARK: - Recording States
    let recordingActive = Color.adaptive(light: "C62828", dark: "EF5350")
    let recordingInactive = Color.adaptive(light: "2E7D32", dark: "66BB6A")

    // MARK: - UI Elements
    let cardBackground = Color(uiColor: .secondarySystemBackground)
    let border = Color(uiColor: .separator)

    // MARK: - Special
    let premium = Color.adaptive(light: "F9A825", dark: "FFD54F")
    let checkmarkComplete = Color(hex: "4CAF50")
    let checkmarkIncomplete = Color.adaptive(light: "BDBDBD", dark: "616161")

    // MARK: - Habit Colors (for user selection)
    static let habitColors: [(name: String, color: Color)] = [
        ("green", Color(hex: "4CAF50")),
        ("blue", Color(hex: "2196F3")),
        ("purple", Color(hex: "9C27B0")),
        ("orange", Color(hex: "FF9800")),
        ("pink", Color(hex: "E91E63")),
        ("red", Color(hex: "F44336"))
    ]
}

// MARK: - Color Extensions

extension Color {
    init(hex: String) {
        let hex = hex.trimmingCharacters(in: CharacterSet.alphanumerics.inverted)
        var int: UInt64 = 0
        Scanner(string: hex).scanHexInt64(&int)
        let r, g, b: UInt64
        switch hex.count {
        case 6:
            (r, g, b) = (int >> 16, int >> 8 & 0xFF, int & 0xFF)
        default:
            (r, g, b) = (0, 0, 0)
        }
        self.init(.sRGB, red: Double(r) / 255, green: Double(g) / 255, blue: Double(b) / 255, opacity: 1)
    }

    /// Creates an adaptive color that changes based on light/dark mode
    static func adaptive(light: String, dark: String) -> Color {
        Color(uiColor: UIColor { traitCollection in
            traitCollection.userInterfaceStyle == .dark
                ? UIColor(Color(hex: dark))
                : UIColor(Color(hex: light))
        })
    }
}

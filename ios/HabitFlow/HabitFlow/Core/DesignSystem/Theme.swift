import SwiftUI

// MARK: - Corner Radius Tokens
enum CornerRadius {
    /// 8pt - Small elements like tags, badges
    static let small: CGFloat = 8
    /// 12pt - Cards, list items, buttons
    static let medium: CGFloat = 12
    /// 16pt - Large cards, modals
    static let large: CGFloat = 16
    /// 20pt - Extra large elements, sheets
    static let extraLarge: CGFloat = 20
    /// 24pt - Full rounded for pills
    static let pill: CGFloat = 24
}

// MARK: - Spacing Tokens
enum Spacing {
    /// 4pt
    static let xxs: CGFloat = 4
    /// 8pt
    static let xs: CGFloat = 8
    /// 12pt
    static let sm: CGFloat = 12
    /// 16pt
    static let md: CGFloat = 16
    /// 20pt
    static let lg: CGFloat = 20
    /// 24pt
    static let xl: CGFloat = 24
    /// 32pt
    static let xxl: CGFloat = 32
}

// MARK: - Animation Tokens
enum AnimationDuration {
    static let quick: Double = 0.15
    static let normal: Double = 0.25
    static let slow: Double = 0.4
}

struct AppTheme {
    // MARK: - Login Background (adaptive gradient)
    static func loginBackground(for colorScheme: ColorScheme) -> LinearGradient {
        LinearGradient(
            colors: colorScheme == .dark
                ? [Color(hex: "1A2A1A"), Color(hex: "0D1F0D")]
                : [Color(hex: "E8F5E9"), Color(hex: "C8E6C9")],
            startPoint: .top,
            endPoint: .bottom
        )
    }

    // MARK: - App Background (adaptive)
    static func appBackground(for colorScheme: ColorScheme) -> Color {
        colorScheme == .dark
            ? Color(hex: "121212")
            : Color(hex: "F5F7F5")
    }

    // MARK: - Card Background (adaptive)
    static func cardBackground(for colorScheme: ColorScheme) -> Color {
        colorScheme == .dark
            ? Color(hex: "1E1E1E")
            : Color.white
    }

    // MARK: - Profile Streak Gradient (adaptive)
    static func streakGradient(for colorScheme: ColorScheme) -> LinearGradient {
        LinearGradient(
            colors: colorScheme == .dark
                ? [Color(hex: "1B5E20"), Color(hex: "2E7D32")]
                : [Color(hex: "66BB6A"), Color(hex: "81C784")],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
        )
    }

    // MARK: - Premium Gradient
    static func premiumGradient(for colorScheme: ColorScheme) -> LinearGradient {
        LinearGradient(
            colors: colorScheme == .dark
                ? [Color(hex: "F9A825"), Color(hex: "FF8F00")]
                : [Color(hex: "FFD54F"), Color(hex: "FFCA28")],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
        )
    }
}

// MARK: - View Modifiers

struct AdaptiveBackground: ViewModifier {
    @Environment(\.colorScheme) var colorScheme

    func body(content: Content) -> some View {
        content
            .background(AppTheme.appBackground(for: colorScheme))
    }
}

struct AdaptiveCardBackground: ViewModifier {
    @Environment(\.colorScheme) var colorScheme

    func body(content: Content) -> some View {
        content
            .background(AppTheme.cardBackground(for: colorScheme))
    }
}

extension View {
    func appBackground() -> some View {
        modifier(AdaptiveBackground())
    }

    func cardBackground() -> some View {
        modifier(AdaptiveCardBackground())
    }
}

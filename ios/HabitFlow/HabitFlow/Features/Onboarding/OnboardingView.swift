import SwiftUI

struct OnboardingView: View {
    @AppStorage("hasCompletedOnboarding") private var hasCompletedOnboarding = false
    @State private var currentPage = 0
    @Environment(\.colorScheme) var colorScheme

    let pages: [OnboardingPage] = [
        OnboardingPage(
            icon: "sparkles",
            title: "Welcome to Atoma",
            subtitle: "habits, tasks, money — in one flow",
            description: "Your minimalist companion for building a better life, one day at a time."
        ),
        OnboardingPage(
            icon: "repeat",
            title: "Build Habits",
            subtitle: "Small steps, big changes",
            description: "Track daily, weekly, or monthly habits. Watch your streaks grow and stay motivated."
        ),
        OnboardingPage(
            icon: "checkmark.circle",
            title: "Manage Tasks",
            subtitle: "Focus on what matters",
            description: "Organize your daily tasks with priorities. Complete what's important, defer what's not."
        ),
        OnboardingPage(
            icon: "creditcard",
            title: "Track Money",
            subtitle: "Know where it goes",
            description: "Log income and expenses. Understand your spending patterns and save more."
        ),
        OnboardingPage(
            icon: "brain.head.profile",
            title: "AI Insights",
            subtitle: "Your personal coach",
            description: "Get personalized advice from AI agents for habits, tasks, and finances."
        )
    ]

    var body: some View {
        VStack(spacing: 0) {
            // Skip button
            HStack {
                Spacer()
                if currentPage < pages.count - 1 {
                    Button("Skip") {
                        completeOnboarding()
                    }
                    .font(.system(size: 16, weight: .medium))
                    .foregroundStyle(.secondary)
                    .padding()
                }
            }

            // Page content
            TabView(selection: $currentPage) {
                ForEach(0..<pages.count, id: \.self) { index in
                    OnboardingPageView(page: pages[index])
                        .tag(index)
                }
            }
            .tabViewStyle(.page(indexDisplayMode: .never))
            .animation(.easeInOut, value: currentPage)

            // Page indicators
            HStack(spacing: 8) {
                ForEach(0..<pages.count, id: \.self) { index in
                    Circle()
                        .fill(index == currentPage ? Color.hf.accent : Color.gray.opacity(0.3))
                        .frame(width: 8, height: 8)
                        .animation(.easeInOut, value: currentPage)
                }
            }
            .padding(.bottom, 32)

            // Action button
            Button {
                if currentPage < pages.count - 1 {
                    withAnimation {
                        currentPage += 1
                    }
                } else {
                    completeOnboarding()
                }
            } label: {
                Text(currentPage < pages.count - 1 ? "Continue" : "Get Started")
                    .font(.system(size: 17, weight: .semibold))
                    .foregroundStyle(.white)
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 16)
                    .background(Color.hf.accent)
                    .clipShape(RoundedRectangle(cornerRadius: 14))
            }
            .padding(.horizontal, 24)
            .padding(.bottom, 32)
        }
        .background(AppTheme.appBackground(for: colorScheme))
    }

    private func completeOnboarding() {
        withAnimation {
            hasCompletedOnboarding = true
        }
    }
}

struct OnboardingPage: Identifiable {
    let id = UUID()
    let icon: String
    let title: String
    let subtitle: String
    let description: String
}

struct OnboardingPageView: View {
    let page: OnboardingPage
    @Environment(\.colorScheme) var colorScheme

    var body: some View {
        VStack(spacing: 24) {
            Spacer()

            // Icon
            ZStack {
                Circle()
                    .fill(Color.hf.accent.opacity(0.15))
                    .frame(width: 120, height: 120)

                Image(systemName: page.icon)
                    .font(.system(size: 50, weight: .medium))
                    .foregroundStyle(Color.hf.accent)
            }

            // Title
            VStack(spacing: 8) {
                Text(page.title)
                    .font(.system(size: 28, weight: .bold))
                    .multilineTextAlignment(.center)

                Text(page.subtitle)
                    .font(.system(size: 17, weight: .medium))
                    .foregroundStyle(Color.hf.accent)
                    .multilineTextAlignment(.center)
            }

            // Description
            Text(page.description)
                .font(.system(size: 16))
                .foregroundStyle(.secondary)
                .multilineTextAlignment(.center)
                .padding(.horizontal, 32)

            Spacer()
            Spacer()
        }
        .padding()
    }
}

#Preview {
    OnboardingView()
}

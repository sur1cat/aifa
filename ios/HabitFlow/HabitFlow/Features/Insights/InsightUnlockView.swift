import SwiftUI

struct InsightUnlockView: View {
    let section: InsightSection
    var onDismiss: () -> Void

    @State private var showContent = false
    @State private var showSparkles = false
    @State private var ringScale: CGFloat = 0.5
    @State private var iconScale: CGFloat = 0

    var body: some View {
        ZStack {
            // Background
            Color.black.opacity(0.9)
                .ignoresSafeArea()

            VStack(spacing: 32) {
                Spacer()

                // Animated unlock ring
                ZStack {
                    // Outer glow
                    Circle()
                        .fill(sectionColor.opacity(0.3))
                        .frame(width: 180, height: 180)
                        .blur(radius: 30)
                        .scaleEffect(showSparkles ? 1.2 : 0.8)

                    // Ring
                    Circle()
                        .stroke(
                            LinearGradient(
                                colors: [sectionColor, sectionColor.opacity(0.6)],
                                startPoint: .topLeading,
                                endPoint: .bottomTrailing
                            ),
                            lineWidth: 6
                        )
                        .frame(width: 140, height: 140)
                        .scaleEffect(ringScale)

                    // Icon
                    Image(systemName: section.icon)
                        .font(.system(size: 48, weight: .medium))
                        .foregroundStyle(sectionColor)
                        .scaleEffect(iconScale)

                    // Sparkles
                    if showSparkles {
                        ForEach(0..<8) { i in
                            Image(systemName: "sparkle")
                                .font(.system(size: 16, weight: .bold))
                                .foregroundStyle(sectionColor)
                                .offset(sparkleOffset(for: i))
                                .opacity(showSparkles ? 1 : 0)
                                .animation(
                                    .easeInOut(duration: 0.8)
                                    .repeatForever(autoreverses: true)
                                    .delay(Double(i) * 0.1),
                                    value: showSparkles
                                )
                        }
                    }
                }

                // Text content
                if showContent {
                    VStack(spacing: 12) {
                        Text("unlocked".uppercased())
                            .font(.system(size: 12, weight: .bold, design: .rounded))
                            .tracking(3)
                            .foregroundStyle(sectionColor)

                        Text(sectionTitle)
                            .font(.system(size: 28, weight: .bold, design: .rounded))
                            .foregroundStyle(.white)

                        Text(sectionDescription)
                            .font(.system(size: 16))
                            .foregroundStyle(.white.opacity(0.7))
                            .multilineTextAlignment(.center)
                            .padding(.horizontal, 40)
                    }
                    .transition(.opacity.combined(with: .move(edge: .bottom)))
                }

                Spacer()

                // Continue button
                if showContent {
                    Button {
                        onDismiss()
                    } label: {
                        Text("View Insights")
                            .font(.system(size: 17, weight: .semibold))
                            .foregroundStyle(.white)
                            .frame(maxWidth: .infinity)
                            .padding(.vertical, 16)
                            .background(sectionColor)
                            .clipShape(RoundedRectangle(cornerRadius: 14))
                    }
                    .padding(.horizontal, 24)
                    .padding(.bottom, 40)
                    .transition(.opacity.combined(with: .move(edge: .bottom)))
                }
            }
        }
        .onAppear {
            animateUnlock()
        }
    }

    private var sectionColor: Color {
        switch section {
        case .habits: return Color.hf.accent
        case .tasks: return Color.hf.info
        case .budget: return Color.hf.premium
        }
    }

    private var sectionTitle: LocalizedStringKey {
        switch section {
        case .habits: return "Habits Insights"
        case .tasks: return "Tasks Insights"
        case .budget: return "Budget Insights"
        }
    }

    private var sectionDescription: LocalizedStringKey {
        switch section {
        case .habits: return "AI-powered analysis of your habit patterns and personalized recommendations"
        case .tasks: return "Smart insights about your productivity and task completion trends"
        case .budget: return "Detailed spending analysis and financial recommendations from AI"
        }
    }

    private func sparkleOffset(for index: Int) -> CGSize {
        let angle = Double(index) * (360.0 / 8.0) * .pi / 180
        let radius: Double = 100
        return CGSize(
            width: Foundation.cos(angle) * radius,
            height: Foundation.sin(angle) * radius
        )
    }

    private func animateUnlock() {
        // Ring scale
        withAnimation(.spring(response: 0.6, dampingFraction: 0.7)) {
            ringScale = 1.0
        }

        // Icon pop
        withAnimation(.spring(response: 0.5, dampingFraction: 0.6).delay(0.2)) {
            iconScale = 1.0
        }

        // Sparkles
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.4) {
            withAnimation(.easeInOut(duration: 0.3)) {
                showSparkles = true
            }
        }

        // Content
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.6) {
            withAnimation(.spring(response: 0.5, dampingFraction: 0.8)) {
                showContent = true
            }
        }
    }
}

// MARK: - Preview

#Preview {
    InsightUnlockView(section: .habits) {
        print("Dismissed")
    }
}

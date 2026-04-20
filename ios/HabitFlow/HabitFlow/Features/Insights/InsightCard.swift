import SwiftUI

// MARK: - Insight Carousel (WHOOP-style)

struct InsightCarousel: View {
    let insights: [Insight]
    var onDismiss: ((Insight) -> Void)? = nil
    var onAction: ((InsightAction) -> Void)? = nil

    @State private var currentIndex = 0
    @State private var timer: Timer?

    private var iconColor: Color {
        guard currentIndex < insights.count else { return Color.hf.accent }
        let insight = insights[currentIndex]
        switch insight.type.color {
        case "info": return Color.hf.info
        case "accent": return Color.hf.accent
        case "warning": return Color.hf.warning
        default: return Color.hf.accent
        }
    }

    var body: some View {
        if insights.isEmpty {
            EmptyView()
        } else {
            VStack(spacing: 12) {
                // Insight Card
                TabView(selection: $currentIndex) {
                    ForEach(Array(insights.enumerated()), id: \.element.id) { index, insight in
                        InsightCarouselCard(
                            insight: insight,
                            onDismiss: { onDismiss?(insight) },
                            onAction: onAction
                        )
                        .tag(index)
                    }
                }
                .tabViewStyle(.page(indexDisplayMode: .never))
                .frame(height: 120)

                // Page Indicators
                if insights.count > 1 {
                    HStack(spacing: 6) {
                        ForEach(0..<insights.count, id: \.self) { index in
                            Circle()
                                .fill(index == currentIndex ? iconColor : Color.gray.opacity(0.3))
                                .frame(width: 6, height: 6)
                                .animation(.easeInOut(duration: 0.2), value: currentIndex)
                        }
                    }
                }
            }
            .padding(.horizontal, 16)
            .onAppear { startAutoScroll() }
            .onDisappear { stopAutoScroll() }
            .onChange(of: currentIndex) { _ in restartAutoScroll() }
        }
    }

    private func startAutoScroll() {
        guard insights.count > 1 else { return }
        timer = Timer.scheduledTimer(withTimeInterval: 5.0, repeats: true) { _ in
            withAnimation(.easeInOut(duration: 0.3)) {
                currentIndex = (currentIndex + 1) % insights.count
            }
        }
    }

    private func stopAutoScroll() {
        timer?.invalidate()
        timer = nil
    }

    private func restartAutoScroll() {
        stopAutoScroll()
        startAutoScroll()
    }
}

struct InsightCarouselCard: View {
    let insight: Insight
    var onDismiss: (() -> Void)? = nil
    var onAction: ((InsightAction) -> Void)? = nil

    private var iconColor: Color {
        switch insight.type.color {
        case "info": return Color.hf.info
        case "accent": return Color.hf.accent
        case "warning": return Color.hf.warning
        default: return Color.hf.accent
        }
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            // Header
            HStack(spacing: 8) {
                Image(systemName: insight.type.icon)
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundStyle(iconColor)

                Text(insight.title)
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundStyle(.primary)

                Spacer()

                Button {
                    onDismiss?()
                } label: {
                    Image(systemName: "xmark")
                        .font(.system(size: 11, weight: .medium))
                        .foregroundStyle(.tertiary)
                        .padding(6)
                        .background(Color.primary.opacity(0.05))
                        .clipShape(Circle())
                }
            }

            // Message
            Text(insight.message)
                .font(.system(size: 14))
                .foregroundStyle(.secondary)
                .lineSpacing(2)
                .lineLimit(3)

            // Action button (if present)
            if let action = insight.action {
                Button {
                    onAction?(action)
                } label: {
                    Text(action.label)
                        .font(.system(size: 13, weight: .medium))
                        .foregroundStyle(iconColor)
                }
            }
        }
        .padding(14)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(iconColor.opacity(0.08))
        .clipShape(RoundedRectangle(cornerRadius: 14))
    }
}

// MARK: - Compact Insight Chip (new organic design)

struct CompactInsightChip: View {
    let insight: Insight
    var onDismiss: (() -> Void)? = nil

    private var iconColor: Color {
        switch insight.type.color {
        case "info": return Color.hf.info
        case "accent": return Color.hf.accent
        case "warning": return Color.hf.warning
        default: return Color.hf.accent
        }
    }

    var body: some View {
        HStack(spacing: 10) {
            Image(systemName: insight.type.icon)
                .font(.system(size: 13, weight: .medium))
                .foregroundStyle(iconColor)

            Text(insight.message)
                .font(.system(size: 13))
                .foregroundStyle(.secondary)
                .lineLimit(2)
                .frame(maxWidth: .infinity, alignment: .leading)

            Button {
                onDismiss?()
            } label: {
                Image(systemName: "xmark")
                    .font(.system(size: 10, weight: .medium))
                    .foregroundStyle(.tertiary)
                    .padding(4)
            }
        }
        .padding(.horizontal, 14)
        .padding(.vertical, 12)
        .frame(maxWidth: .infinity)
        .background(iconColor.opacity(0.08))
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

// MARK: - Inline Insights View (vertical stack, full width)

struct InlineInsightsRow: View {
    let insights: [Insight]
    var onDismiss: ((Insight) -> Void)? = nil

    var body: some View {
        if insights.isEmpty {
            EmptyView()
        } else {
            VStack(spacing: 8) {
                ForEach(insights) { insight in
                    CompactInsightChip(
                        insight: insight,
                        onDismiss: { onDismiss?(insight) }
                    )
                }
            }
            .padding(.horizontal, 16)
        }
    }
}

// MARK: - Original InsightCard (kept for reference/legacy)

struct InsightCard: View {
    let insight: Insight
    var onAction: ((InsightAction) -> Void)? = nil
    var onDismiss: (() -> Void)? = nil

    private var iconColor: Color {
        switch insight.type.color {
        case "info": return Color.hf.info
        case "accent": return Color.hf.accent
        case "warning": return Color.hf.warning
        default: return Color.hf.accent
        }
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            // Header
            HStack(spacing: 10) {
                Image(systemName: insight.type.icon)
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundStyle(iconColor)

                Text(insight.title)
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundStyle(.primary)

                Spacer()

                Button {
                    onDismiss?()
                } label: {
                    Image(systemName: "xmark")
                        .font(.system(size: 12, weight: .medium))
                        .foregroundStyle(.tertiary)
                }
            }

            // Message
            Text(insight.message)
                .font(.system(size: 15))
                .foregroundStyle(.secondary)
                .lineSpacing(2)
                .fixedSize(horizontal: false, vertical: true)

            // Action button (if present)
            if let action = insight.action {
                Button {
                    onAction?(action)
                } label: {
                    Text(action.label)
                        .font(.system(size: 14, weight: .medium))
                        .foregroundStyle(iconColor)
                }
                .padding(.top, 4)
            }
        }
        .padding(16)
        .background(Color.hf.cardBackground)
        .clipShape(RoundedRectangle(cornerRadius: 16))
        .overlay(
            RoundedRectangle(cornerRadius: 16)
                .stroke(iconColor.opacity(0.2), lineWidth: 1)
        )
    }
}

// MARK: - Insights List View (vertical, for backwards compatibility)

struct InsightsListView: View {
    let insights: [Insight]
    var onAction: ((InsightAction) -> Void)? = nil
    var onDismiss: ((Insight) -> Void)? = nil

    var body: some View {
        if insights.isEmpty {
            EmptyView()
        } else {
            VStack(alignment: .leading, spacing: 12) {
                HStack {
                    Image(systemName: "lightbulb.fill")
                        .font(.system(size: 12))
                        .foregroundStyle(Color.hf.accent)

                    Text("Insights")
                        .font(.system(size: 14, weight: .medium))
                        .foregroundStyle(.secondary)
                }
                .padding(.horizontal, 4)

                ForEach(insights) { insight in
                    InsightCard(
                        insight: insight,
                        onAction: onAction,
                        onDismiss: {
                            onDismiss?(insight)
                        }
                    )
                }
            }
        }
    }
}

// MARK: - Preview

#Preview {
    ScrollView {
        VStack(spacing: 16) {
            InsightCard(
                insight: Insight(
                    section: .habits,
                    type: .achievement,
                    title: "7 Day Streak!",
                    message: "Meditation — you've been consistent for 7 days straight!"
                )
            )

            InsightCard(
                insight: Insight(
                    section: .habits,
                    type: .warning,
                    title: "Needs Attention",
                    message: "Gym has been challenging lately. Would a different time work better?",
                    action: InsightAction(label: "Ask AI", actionType: "openAIChat")
                )
            )

            InsightCard(
                insight: Insight(
                    section: .budget,
                    type: .pattern,
                    title: "Top Spending",
                    message: "Food is 35% of your expenses ($450 this month)"
                )
            )
        }
        .padding()
    }
    .background(Color.hf.surface)
}

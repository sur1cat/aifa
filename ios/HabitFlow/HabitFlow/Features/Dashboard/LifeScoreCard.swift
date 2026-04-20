import SwiftUI

struct LifeScoreCard: View {
    @EnvironmentObject var dataManager: DataManager
    let period: AnalyticsPeriod

    @State private var animatedScore: Double = 0

    private var score: Double {
        dataManager.lifeScore(for: period)
    }

    private var scoreColor: Color {
        switch animatedScore {
        case 80...100: return Color.hf.accent
        case 60..<80: return Color.hf.info
        case 40..<60: return Color.hf.warning
        default: return Color.hf.expense
        }
    }

    var body: some View {
        VStack(spacing: 16) {
            // Ring with score
            LifeScoreRing(score: animatedScore, color: scoreColor)
                .frame(width: 120, height: 120)

            // Label
            VStack(spacing: 4) {
                Text("Life Score")
                    .font(.system(size: 15, weight: .medium))
                    .foregroundStyle(.primary)

                Text("Your weekly balance")
                    .font(.system(size: 13))
                    .foregroundStyle(.secondary)
            }

            // Breakdown
            LifeScoreBreakdown(period: period)
        }
        .padding(.vertical, 20)
        .frame(maxWidth: .infinity)
        .background(Color.hf.cardBackground)
        .clipShape(RoundedRectangle(cornerRadius: 16))
        .onAppear {
            withAnimation(.spring(response: 1.0, dampingFraction: 0.8)) {
                animatedScore = score
            }
        }
        .onChange(of: score) { _, newValue in
            withAnimation(.spring(response: 0.6, dampingFraction: 0.8)) {
                animatedScore = newValue
            }
        }
    }
}

struct LifeScoreRing: View {
    let score: Double
    let color: Color

    var body: some View {
        ZStack {
            // Background ring
            Circle()
                .stroke(color.opacity(0.2), lineWidth: 12)

            // Progress ring
            Circle()
                .trim(from: 0, to: score / 100)
                .stroke(color, style: StrokeStyle(lineWidth: 12, lineCap: .round))
                .rotationEffect(.degrees(-90))
                .animation(.spring(response: 1.0, dampingFraction: 0.8), value: score)

            // Score text
            Text("\(Int(score))")
                .font(.system(size: 36, weight: .bold, design: .rounded))
                .foregroundStyle(color)
                .contentTransition(.numericText())
        }
    }
}

struct LifeScoreBreakdown: View {
    @EnvironmentObject var dataManager: DataManager
    let period: AnalyticsPeriod

    var body: some View {
        let components = dataManager.lifeScoreComponents(for: period)

        HStack(spacing: 24) {
            BreakdownItem(
                icon: "flame.fill",
                value: Int(components.habits),
                color: Color.hf.accent
            )
            BreakdownItem(
                icon: "checkmark.circle.fill",
                value: Int(components.tasks),
                color: Color.hf.info
            )
            BreakdownItem(
                icon: "dollarsign.circle.fill",
                value: Int(components.budget),
                color: Color.hf.income
            )
        }
    }
}

struct BreakdownItem: View {
    let icon: String
    let value: Int
    let color: Color

    var body: some View {
        HStack(spacing: 4) {
            Image(systemName: icon)
                .font(.system(size: 14))
                .foregroundStyle(color)
            Text("\(value)%")
                .font(.system(size: 13, weight: .medium))
                .foregroundStyle(.secondary)
        }
    }
}

#Preview {
    LifeScoreCard(period: .week)
        .environmentObject(DataManager.shared)
        .padding()
}

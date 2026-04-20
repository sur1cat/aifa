import SwiftUI

struct InsightProgressCard: View {
    @EnvironmentObject var dataManager: DataManager
    let section: InsightSection
    var onUnlockTap: (() -> Void)? = nil

    @State private var showCelebration = false

    private var isUnlocked: Bool {
        dataManager.hasEnoughData(for: section)
    }

    private var needsCelebration: Bool {
        // Celebrate when first unlocked (data threshold reached)
        isUnlocked && !UserDefaults.standard.bool(forKey: "insight_celebrated_\(section.rawValue)")
    }

    private var dataProgress: (current: Int, required: Int, label: String) {
        switch section {
        case .habits:
            // 14 days of tracking for behavioral patterns
            let uniqueDays = Set(dataManager.habits.flatMap { $0.completedDates }).count
            return (min(uniqueDays, 14), 14, "days tracked")
        case .tasks:
            // 14 days / 10 tasks for task patterns
            let calendar = Calendar.current
            let fourteenDaysAgo = calendar.date(byAdding: .day, value: -14, to: Date()) ?? Date()
            let recentTasks = dataManager.tasks.filter { $0.dueDate >= fourteenDaysAgo }.count
            return (min(recentTasks, 10), 10, "tasks in 2 weeks")
        case .budget:
            // 30 transactions for spending patterns
            return (min(dataManager.transactions.count, 30), 30, "transactions")
        }
    }

    private var progress: Double {
        let (current, required, _) = dataProgress
        return min(Double(current) / Double(required), 1.0)
    }

    var body: some View {
        Group {
            if isUnlocked {
                unlockedCard
            } else {
                lockedCard
            }
        }
        .sheet(isPresented: $showCelebration) {
            InsightUnlockView(section: section) {
                UserDefaults.standard.set(true, forKey: "insight_celebrated_\(section.rawValue)")
                showCelebration = false
            }
        }
        .onAppear {
            if needsCelebration {
                DispatchQueue.main.asyncAfter(deadline: .now() + 0.5) {
                    showCelebration = true
                }
            }
        }
    }

    // MARK: - Locked State

    private var lockedCard: some View {
        VStack(alignment: .leading, spacing: 12) {
            // Header
            HStack(spacing: 8) {
                Image(systemName: "brain.head.profile")
                    .font(.system(size: 14, weight: .medium))
                    .foregroundStyle(Color.hf.accent)

                Text("Daily Insights")
                    .font(.system(size: 15, weight: .semibold))
                    .foregroundStyle(.primary)

                Spacer()

                // Data count
                let (current, required, _) = dataProgress
                Text("\(current)/\(required)")
                    .font(.system(size: 13, weight: .medium, design: .rounded))
                    .foregroundStyle(Color.hf.accent)
            }

            // Progress bar
            GeometryReader { geometry in
                ZStack(alignment: .leading) {
                    RoundedRectangle(cornerRadius: 4)
                        .fill(Color.hf.accent.opacity(0.2))
                        .frame(height: 8)

                    RoundedRectangle(cornerRadius: 4)
                        .fill(Color.hf.accent)
                        .frame(width: geometry.size.width * progress, height: 8)
                        .animation(.easeInOut(duration: 0.3), value: progress)
                }
            }
            .frame(height: 8)

            // Description
            Text(progressDescription)
                .font(.system(size: 13))
                .foregroundStyle(.secondary)
        }
        .padding(16)
        .background(Color.hf.cardBackground)
        .clipShape(RoundedRectangle(cornerRadius: 16))
    }

    private var progressDescription: String {
        let (current, required, label) = dataProgress
        let remaining = required - current

        if remaining <= 0 {
            return "Insights will be available soon!"
        } else if remaining == 1 {
            return "1 more \(label.replacingOccurrences(of: "s ", with: " ").replacingOccurrences(of: "s$", with: "", options: .regularExpression)) to unlock personalized insights"
        } else {
            return "\(remaining) more \(label) to unlock personalized insights"
        }
    }

    // MARK: - Unlocked State

    private var unlockedCard: some View {
        HStack(spacing: 16) {
            ZStack {
                Circle()
                    .fill(Color.hf.accent.opacity(0.15))

                Image(systemName: "brain.head.profile")
                    .font(.system(size: 18, weight: .medium))
                    .foregroundStyle(Color.hf.accent)
            }
            .frame(width: 44, height: 44)

            VStack(alignment: .leading, spacing: 4) {
                Text("Daily Insights")
                    .font(.system(size: 15, weight: .semibold))
                    .foregroundStyle(.primary)

                Text("Powered by AI")
                    .font(.system(size: 13))
                    .foregroundStyle(.secondary)
            }

            Spacer()

            // Last updated indicator
            Image(systemName: "checkmark.circle.fill")
                .font(.system(size: 20))
                .foregroundStyle(Color.hf.accent)
        }
        .padding(16)
        .background(Color.hf.cardBackground)
        .clipShape(RoundedRectangle(cornerRadius: 16))
    }
}

// MARK: - Preview

#Preview {
    VStack(spacing: 16) {
        InsightProgressCard(section: .habits)
        InsightProgressCard(section: .tasks)
        InsightProgressCard(section: .budget)
    }
    .padding()
    .background(Color.hf.surface)
    .environmentObject(DataManager.shared)
}

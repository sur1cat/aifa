import SwiftUI

struct WeeklyReviewView: View {
    @EnvironmentObject var dataManager: DataManager
    @Environment(\.dismiss) var dismiss

    let review: WeeklyReview

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(spacing: 20) {
                    // Header
                    headerSection

                    // Stats Grid
                    statsSection

                    // Wins
                    if !review.wins.isEmpty {
                        highlightsSection(title: "Wins", items: review.wins, color: Color.hf.accent, icon: "trophy.fill")
                    }

                    // Warnings
                    if !review.warnings.isEmpty {
                        highlightsSection(title: "Watch", items: review.warnings, color: Color.hf.warning, icon: "exclamationmark.triangle.fill")
                    }

                    // Budget Summary
                    if review.totalIncome > 0 || review.totalExpenses > 0 {
                        budgetSection
                    }
                }
                .padding()
            }
            .background(Color.hf.surface)
            .navigationTitle("Weekly Review")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button("Done") {
                        dismiss()
                    }
                }
            }
        }
    }

    // MARK: - Header

    private var headerSection: some View {
        VStack(spacing: 8) {
            Image(systemName: "calendar")
                .font(.system(size: 32))
                .foregroundStyle(Color.hf.accent)

            Text(review.weekDateRange)
                .font(.system(size: 20, weight: .semibold))

            Text("Your week in review")
                .font(.system(size: 14))
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity)
        .padding(.vertical, 20)
        .background(Color.hf.cardBackground)
        .clipShape(RoundedRectangle(cornerRadius: 16))
    }

    // MARK: - Stats

    private var statsSection: some View {
        HStack(spacing: 12) {
            // Habits
            StatBox(
                title: "Habits",
                value: "\(Int(review.habitsCompletionRate * 100))%",
                change: review.habitsCompletionRateChange,
                icon: "repeat",
                color: Color.hf.accent
            )

            // Budget
            StatBox(
                title: "Spent",
                value: dataManager.formatCurrency(review.totalExpenses),
                change: review.expensesChange,
                icon: "creditcard",
                color: Color.hf.expense,
                invertChange: true
            )
        }
    }

    // MARK: - Highlights

    private func highlightsSection(title: String, items: [String], color: Color, icon: String) -> some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack(spacing: 8) {
                Image(systemName: icon)
                    .font(.system(size: 14))
                    .foregroundStyle(color)

                Text(title)
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundStyle(.secondary)
            }

            VStack(spacing: 8) {
                ForEach(items, id: \.self) { item in
                    HStack(spacing: 12) {
                        Circle()
                            .fill(color)
                            .frame(width: 6, height: 6)

                        Text(item)
                            .font(.system(size: 15))
                            .foregroundStyle(.primary)

                        Spacer()
                    }
                    .padding(12)
                    .background(Color.hf.cardBackground)
                    .clipShape(RoundedRectangle(cornerRadius: 12))
                }
            }
        }
    }

    // MARK: - Budget

    private var budgetSection: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack(spacing: 8) {
                Image(systemName: "chart.pie.fill")
                    .font(.system(size: 14))
                    .foregroundStyle(Color.hf.info)

                Text("Budget")
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundStyle(.secondary)
            }

            VStack(spacing: 12) {
                HStack {
                    Text("Income")
                        .foregroundStyle(.secondary)
                    Spacer()
                    Text(dataManager.formatCurrency(review.totalIncome))
                        .font(.system(size: 16, weight: .medium))
                        .foregroundStyle(Color.hf.income)
                }

                HStack {
                    Text("Expenses")
                        .foregroundStyle(.secondary)
                    Spacer()
                    Text(dataManager.formatCurrency(review.totalExpenses))
                        .font(.system(size: 16, weight: .medium))
                        .foregroundStyle(Color.hf.expense)
                }

                Divider()

                HStack {
                    Text("Net")
                        .fontWeight(.medium)
                    Spacer()
                    let net = review.totalIncome - review.totalExpenses
                    Text(dataManager.formatCurrency(abs(net)))
                        .font(.system(size: 16, weight: .semibold))
                        .foregroundStyle(net >= 0 ? Color.hf.income : Color.hf.expense)
                }

                if let category = review.topCategory, let amount = review.topCategoryAmount {
                    Divider()

                    HStack {
                        Text("Top: \(category)")
                            .foregroundStyle(.secondary)
                        Spacer()
                        Text(dataManager.formatCurrency(amount))
                            .font(.system(size: 14))
                            .foregroundStyle(.secondary)
                    }
                }
            }
            .padding(16)
            .background(Color.hf.cardBackground)
            .clipShape(RoundedRectangle(cornerRadius: 12))
        }
    }
}

// MARK: - Stat Box

struct StatBox: View {
    let title: String
    let value: String
    let change: Double
    let icon: String
    let color: Color
    var invertChange: Bool = false

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Image(systemName: icon)
                    .font(.system(size: 14))
                    .foregroundStyle(color)

                Spacer()

                if change != 0 {
                    changeIndicator
                }
            }

            Text(value)
                .font(.system(size: 24, weight: .bold, design: .rounded))

            Text(title)
                .font(.system(size: 13))
                .foregroundStyle(.secondary)
        }
        .padding(16)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(Color.hf.cardBackground)
        .clipShape(RoundedRectangle(cornerRadius: 16))
    }

    private var changeIndicator: some View {
        let isPositive = invertChange ? change < 0 : change > 0
        let displayChange = Int(abs(change) * 100)

        return HStack(spacing: 2) {
            Image(systemName: change > 0 ? "arrow.up" : "arrow.down")
                .font(.system(size: 10, weight: .bold))

            Text("\(displayChange)%")
                .font(.system(size: 11, weight: .medium))
        }
        .foregroundStyle(isPositive ? Color.hf.income : Color.hf.expense)
    }
}

// MARK: - Weekly Review Card (for Dashboard/Profile)

struct WeeklyReviewCard: View {
    @EnvironmentObject var dataManager: DataManager
    @State private var showReview = false

    var body: some View {
        Button {
            showReview = true
        } label: {
            HStack(spacing: 16) {
                ZStack {
                    Circle()
                        .fill(Color.hf.accent.opacity(0.15))

                    Image(systemName: "calendar")
                        .font(.system(size: 18, weight: .medium))
                        .foregroundStyle(Color.hf.accent)
                }
                .frame(width: 44, height: 44)

                VStack(alignment: .leading, spacing: 4) {
                    Text("Weekly Review")
                        .font(.system(size: 15, weight: .semibold))
                        .foregroundStyle(.primary)

                    Text("See your week's highlights")
                        .font(.system(size: 13))
                        .foregroundStyle(.secondary)
                }

                Spacer()

                Image(systemName: "chevron.right")
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundStyle(.tertiary)
            }
            .padding(16)
            .background(Color.hf.cardBackground)
            .clipShape(RoundedRectangle(cornerRadius: 16))
        }
        .buttonStyle(.plain)
        .sheet(isPresented: $showReview) {
            WeeklyReviewView(
                review: WeeklyReviewGenerator.generateReview(
                    habits: dataManager.habits,
                    transactions: dataManager.transactions,
                    currency: dataManager.profile.currency
                )
            )
        }
    }
}

// MARK: - Preview

#Preview {
    WeeklyReviewView(
        review: WeeklyReview(
            weekStart: Date().addingTimeInterval(-7 * 24 * 60 * 60),
            weekEnd: Date().addingTimeInterval(-1 * 24 * 60 * 60),
            habitsCompletionRate: 0.78,
            habitsCompletionRateChange: 0.05,
            totalHabitsCompleted: 32,
            bestHabit: "Meditation",
            habitStreak: 14,
            totalIncome: 50000,
            totalExpenses: 28500,
            expensesChange: -0.12,
            topCategory: "Food",
            topCategoryAmount: 8500,
            wins: ["Habits: 78% completion!", "Meditation — 7 day streak!", "Spending down 12%"],
            warnings: ["Gym skipped 3 times"]
        )
    )
    .environmentObject(DataManager.shared)
}

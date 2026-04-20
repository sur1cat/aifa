import SwiftUI
import Charts

// MARK: - Life Score History Chart

struct LifeScoreHistoryChart: View {
    @EnvironmentObject var dataManager: DataManager
    let period: AnalyticsPeriod

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Life Score Trend")
                .font(.system(size: 14, weight: .medium))
                .foregroundStyle(.primary)
                .padding(.horizontal, 4)

            let data = dataManager.lifeScoreHistory(for: period)

            if data.isEmpty {
                Text("Not enough data yet")
                    .font(.system(size: 14))
                    .foregroundStyle(.secondary)
                    .frame(maxWidth: .infinity, minHeight: 150)
                    .background(Color.hf.cardBackground)
                    .clipShape(RoundedRectangle(cornerRadius: 16))
            } else {
                Chart(data) { point in
                    LineMark(
                        x: .value("Date", point.date),
                        y: .value("Score", point.score)
                    )
                    .foregroundStyle(Color.hf.accent.gradient)
                    .lineStyle(StrokeStyle(lineWidth: 2.5, lineCap: .round))

                    AreaMark(
                        x: .value("Date", point.date),
                        y: .value("Score", point.score)
                    )
                    .foregroundStyle(
                        LinearGradient(
                            colors: [Color.hf.accent.opacity(0.3), Color.hf.accent.opacity(0.0)],
                            startPoint: .top,
                            endPoint: .bottom
                        )
                    )

                    PointMark(
                        x: .value("Date", point.date),
                        y: .value("Score", point.score)
                    )
                    .foregroundStyle(Color.hf.accent)
                    .symbolSize(30)
                }
                .chartYScale(domain: 0...100)
                .chartYAxis {
                    AxisMarks(position: .leading, values: [0, 25, 50, 75, 100]) { value in
                        AxisGridLine(stroke: StrokeStyle(lineWidth: 0.5, dash: [4]))
                            .foregroundStyle(Color.secondary.opacity(0.3))
                        AxisValueLabel {
                            if let intValue = value.as(Int.self) {
                                Text("\(intValue)")
                                    .font(.system(size: 10))
                                    .foregroundStyle(.secondary)
                            }
                        }
                    }
                }
                .chartXAxis {
                    AxisMarks(values: .automatic(desiredCount: 5)) { value in
                        AxisValueLabel {
                            if let date = value.as(Date.self) {
                                Text(formatDate(date, for: period))
                                    .font(.system(size: 10))
                                    .foregroundStyle(.secondary)
                            }
                        }
                    }
                }
                .frame(height: 150)
                .padding()
                .background(Color.hf.cardBackground)
                .clipShape(RoundedRectangle(cornerRadius: 16))
            }
        }
    }

    private func formatDate(_ date: Date, for period: AnalyticsPeriod) -> String {
        let formatter = DateFormatter()
        switch period {
        case .week:
            formatter.dateFormat = "EEE"
        case .month:
            formatter.dateFormat = "d"
        case .year:
            formatter.dateFormat = "MMM"
        }
        return formatter.string(from: date)
    }
}

// MARK: - Life Score Data Point

struct LifeScoreDataPoint: Identifiable {
    let id = UUID()
    let date: Date
    let score: Double
}

// MARK: - Spending by Category Chart

struct SpendingByCategoryChart: View {
    @EnvironmentObject var dataManager: DataManager
    let period: AnalyticsPeriod

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Spending by Category")
                .font(.system(size: 14, weight: .medium))
                .foregroundStyle(.primary)
                .padding(.horizontal, 4)

            let data = dataManager.spendingByCategory(for: period)

            if data.isEmpty {
                Text("No expenses yet")
                    .font(.system(size: 14))
                    .foregroundStyle(.secondary)
                    .frame(maxWidth: .infinity, minHeight: 150)
                    .background(Color.hf.cardBackground)
                    .clipShape(RoundedRectangle(cornerRadius: 16))
            } else {
                HStack(spacing: 16) {
                    // Pie Chart
                    Chart(data) { item in
                        SectorMark(
                            angle: .value("Amount", item.amount),
                            innerRadius: .ratio(0.5),
                            angularInset: 1.5
                        )
                        .foregroundStyle(item.color)
                        .cornerRadius(4)
                    }
                    .frame(width: 120, height: 120)

                    // Legend
                    VStack(alignment: .leading, spacing: 8) {
                        ForEach(data.prefix(5)) { item in
                            HStack(spacing: 8) {
                                Circle()
                                    .fill(item.color)
                                    .frame(width: 10, height: 10)

                                Text(item.category)
                                    .font(.system(size: 12))
                                    .foregroundStyle(.primary)
                                    .lineLimit(1)

                                Spacer()

                                Text(dataManager.formatCurrency(item.amount))
                                    .font(.system(size: 12, weight: .medium))
                                    .foregroundStyle(.secondary)
                            }
                        }
                    }
                    .frame(maxWidth: .infinity)
                }
                .padding()
                .background(Color.hf.cardBackground)
                .clipShape(RoundedRectangle(cornerRadius: 16))
            }
        }
    }
}

// MARK: - Category Spending Data

struct CategorySpending: Identifiable {
    let id = UUID()
    let category: String
    let amount: Double
    let color: Color
}

// MARK: - Habits Streak Calendar

struct HabitsStreakView: View {
    @EnvironmentObject var dataManager: DataManager

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Habit Streaks")
                .font(.system(size: 14, weight: .medium))
                .foregroundStyle(.primary)
                .padding(.horizontal, 4)

            if dataManager.habits.isEmpty {
                Text("No habits yet")
                    .font(.system(size: 14))
                    .foregroundStyle(.secondary)
                    .frame(maxWidth: .infinity, minHeight: 80)
                    .background(Color.hf.cardBackground)
                    .clipShape(RoundedRectangle(cornerRadius: 16))
            } else {
                VStack(spacing: 12) {
                    ForEach(dataManager.habits.sorted(by: { $0.streak > $1.streak }).prefix(3)) { habit in
                        HStack(spacing: 12) {
                            Text(habit.icon)
                                .font(.system(size: 20))

                            VStack(alignment: .leading, spacing: 2) {
                                Text(habit.title)
                                    .font(.system(size: 14, weight: .medium))
                                    .lineLimit(1)

                                Text(habit.streak > 0 ? "\(habit.streak) day streak" : "No streak")
                                    .font(.system(size: 12))
                                    .foregroundStyle(.secondary)
                            }

                            Spacer()

                            // Mini streak bar
                            HStack(spacing: 2) {
                                ForEach(0..<7, id: \.self) { dayOffset in
                                    let date = Calendar.current.date(byAdding: .day, value: -6 + dayOffset, to: Date())!
                                    let dateStr = Habit.dateString(from: date)
                                    let isCompleted: Bool = {
                                        // For habits with goals, check progressValues
                                        if let target = habit.targetValue, target > 0 {
                                            let progress = habit.progressValues[dateStr] ?? 0
                                            return progress >= target || habit.completedDates.contains(dateStr)
                                        }
                                        return habit.completedDates.contains(dateStr)
                                    }()

                                    RoundedRectangle(cornerRadius: 2)
                                        .fill(isCompleted ? habit.swiftUIColor : Color.secondary.opacity(0.2))
                                        .frame(width: 8, height: 20)
                                }
                            }
                        }
                        .padding(.horizontal, 16)
                        .padding(.vertical, 12)
                        .background(Color.hf.cardBackground)
                        .clipShape(RoundedRectangle(cornerRadius: 12))
                    }
                }
            }
        }
    }
}

// MARK: - Income vs Expenses Chart

struct IncomeExpensesChart: View {
    @EnvironmentObject var dataManager: DataManager
    let period: AnalyticsPeriod

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Income vs Expenses")
                .font(.system(size: 14, weight: .medium))
                .foregroundStyle(.primary)
                .padding(.horizontal, 4)

            let income = dataManager.incomeForPeriod(period)
            let expenses = dataManager.expensesForPeriod(period)
            let total = income + expenses

            if total == 0 {
                Text("No transactions yet")
                    .font(.system(size: 14))
                    .foregroundStyle(.secondary)
                    .frame(maxWidth: .infinity, minHeight: 60)
                    .background(Color.hf.cardBackground)
                    .clipShape(RoundedRectangle(cornerRadius: 16))
            } else {
                VStack(spacing: 16) {
                    // Progress bar
                    GeometryReader { geometry in
                        HStack(spacing: 0) {
                            Rectangle()
                                .fill(Color.hf.income)
                                .frame(width: geometry.size.width * (income / total))

                            Rectangle()
                                .fill(Color.hf.expense)
                                .frame(width: geometry.size.width * (expenses / total))
                        }
                        .clipShape(RoundedRectangle(cornerRadius: 6))
                    }
                    .frame(height: 12)

                    // Labels
                    HStack {
                        HStack(spacing: 6) {
                            Circle()
                                .fill(Color.hf.income)
                                .frame(width: 10, height: 10)
                            Text("Income")
                                .font(.system(size: 12))
                            Text(dataManager.formatCurrency(income))
                                .font(.system(size: 12, weight: .semibold))
                                .foregroundStyle(Color.hf.income)
                        }

                        Spacer()

                        HStack(spacing: 6) {
                            Circle()
                                .fill(Color.hf.expense)
                                .frame(width: 10, height: 10)
                            Text("Expenses")
                                .font(.system(size: 12))
                            Text(dataManager.formatCurrency(expenses))
                                .font(.system(size: 12, weight: .semibold))
                                .foregroundStyle(Color.hf.expense)
                        }
                    }
                }
                .padding()
                .background(Color.hf.cardBackground)
                .clipShape(RoundedRectangle(cornerRadius: 16))
            }
        }
    }
}

import SwiftUI

struct DashboardView: View {
    @EnvironmentObject var dataManager: DataManager
    @Environment(\.colorScheme) var colorScheme
    @State private var showAIChat = false

    private let calendar = Calendar.current
    private var today: Date { Date() }
    private var todayString: String { Habit.dateString(from: today) }

    var body: some View {
        NavigationStack {
            List {
                // Tagline
                Section {
                    Text("your daily overview")
                        .font(.system(size: 15))
                        .foregroundStyle(.secondary)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .listRowBackground(Color.clear)
                        .listRowInsets(EdgeInsets(top: 0, leading: 16, bottom: 8, trailing: 16))
                }
                .listRowSeparator(.hidden)

                // Today's Progress Summary
                Section {
                    TodayProgressCard(
                        habitsCompleted: habitsCompletedToday,
                        habitsTotal: habitsTotalToday,
                        tasksCompleted: tasksCompletedToday,
                        tasksTotal: tasksTotalToday,
                        monthBalance: monthBalance
                    )
                    .listRowBackground(Color.clear)
                    .listRowInsets(EdgeInsets(top: 0, leading: 16, bottom: 16, trailing: 16))
                }
                .listRowSeparator(.hidden)

                // Quick Stats Row
                Section {
                    HStack(spacing: 12) {
                        QuickStatCard(
                            title: "Habits",
                            value: "\(habitsCompletedToday)/\(habitsTotalToday)",
                            subtitle: habitsTotalToday > 0 ? "\(Int(habitsProgress * 100))%" : "No habits",
                            icon: "repeat",
                            color: Color.hf.accent
                        )

                        QuickStatCard(
                            title: "Tasks",
                            value: "\(tasksCompletedToday)/\(tasksTotalToday)",
                            subtitle: tasksTotalToday > 0 ? "\(Int(tasksProgress * 100))%" : "No tasks",
                            icon: "checkmark.circle",
                            color: Color.hf.info
                        )
                    }
                    .listRowBackground(Color.clear)
                    .listRowInsets(EdgeInsets(top: 0, leading: 16, bottom: 12, trailing: 16))
                }
                .listRowSeparator(.hidden)

                // Pending Items Section
                if !pendingHabits.isEmpty || !pendingTasks.isEmpty {
                    Section {
                        VStack(alignment: .leading, spacing: 12) {
                            Text("Up Next")
                                .font(.system(size: 15, weight: .semibold))
                                .foregroundStyle(.primary)
                                .padding(.horizontal, 4)

                            VStack(spacing: 8) {
                                ForEach(pendingHabits.prefix(3)) { habit in
                                    PendingItemRow(
                                        icon: habit.icon,
                                        title: habit.title,
                                        type: .habit
                                    )
                                }

                                ForEach(pendingTasks.prefix(3)) { task in
                                    PendingItemRow(
                                        icon: "circle",
                                        title: task.title,
                                        type: .task(priority: task.priority)
                                    )
                                }
                            }
                        }
                        .listRowBackground(Color.clear)
                        .listRowInsets(EdgeInsets(top: 0, leading: 16, bottom: 12, trailing: 16))
                    }
                    .listRowSeparator(.hidden)
                }

                // Monthly Budget Card
                Section {
                    MonthBudgetCard(
                        income: monthIncome,
                        expenses: monthExpenses,
                        currency: dataManager.profile.currency
                    )
                    .listRowBackground(Color.clear)
                    .listRowInsets(EdgeInsets(top: 0, leading: 16, bottom: 12, trailing: 16))
                }
                .listRowSeparator(.hidden)

                // Streaks Section
                if !topStreakHabits.isEmpty {
                    Section {
                        VStack(alignment: .leading, spacing: 12) {
                            Text("Active Streaks")
                                .font(.system(size: 15, weight: .semibold))
                                .foregroundStyle(.primary)
                                .padding(.horizontal, 4)

                            VStack(spacing: 8) {
                                ForEach(topStreakHabits.prefix(3)) { habit in
                                    StreakRow(habit: habit)
                                }
                            }
                        }
                        .listRowBackground(Color.clear)
                        .listRowInsets(EdgeInsets(top: 0, leading: 16, bottom: 20, trailing: 16))
                    }
                    .listRowSeparator(.hidden)
                }
            }
            .listStyle(.plain)
            .scrollContentBackground(.hidden)
            .background(AppTheme.appBackground(for: colorScheme))
            .navigationTitle("Atoma")
            .toolbar {
                ToolbarItem(placement: .topBarLeading) {
                    Button {
                        showAIChat = true
                    } label: {
                        Image(systemName: "atom")
                            .fontWeight(.semibold)
                            .foregroundStyle(Color.hf.accent)
                    }
                }
            }
            .fullScreenCover(isPresented: $showAIChat) {
                AIChatView(agent: .lifeCoach) {
                    buildDashboardContext()
                }
            }
        }
    }

    // MARK: - AI Context

    private func buildDashboardContext() -> String {
        let dateFormatter = DateFormatter()
        dateFormatter.dateFormat = "yyyy-MM-dd"

        var context = "=== USER'S LIFE OVERVIEW ===\n"
        context += "Date: \(dateFormatter.string(from: today))\n\n"

        // Habits summary
        context += "HABITS TODAY:\n"
        context += "- Completed: \(habitsCompletedToday)/\(habitsTotalToday)\n"
        if habitsTotalToday > 0 {
            context += "- Completion rate: \(Int(habitsProgress * 100))%\n"
        }
        if !pendingHabits.isEmpty {
            context += "- Pending: "
            context += pendingHabits.prefix(3).map { $0.icon + " " + $0.title }.joined(separator: ", ")
            context += "\n"
        }

        // Top streaks
        if !topStreakHabits.isEmpty {
            context += "- Active streaks: "
            context += topStreakHabits.prefix(3).map { "\($0.icon) \($0.streak) days" }.joined(separator: ", ")
            context += "\n"
        }

        // Tasks summary
        context += "\nTASKS TODAY:\n"
        context += "- Completed: \(tasksCompletedToday)/\(tasksTotalToday)\n"
        if tasksTotalToday > 0 {
            context += "- Completion rate: \(Int(tasksProgress * 100))%\n"
        }
        if !pendingTasks.isEmpty {
            context += "- Pending by priority:\n"
            for task in pendingTasks.prefix(5) {
                context += "  [\(task.priority.rawValue)] \(task.title)\n"
            }
        }

        // Budget summary
        context += "\nBUDGET THIS MONTH:\n"
        let currency = dataManager.profile.currency.symbol
        context += "- Income: \(currency)\(Int(monthIncome))\n"
        context += "- Expenses: \(currency)\(Int(monthExpenses))\n"
        context += "- Balance: \(currency)\(Int(monthBalance))\n"

        return context
    }

    // MARK: - Computed Properties

    private var habitsForToday: [Habit] {
        dataManager.habits.filter { habit in
            let createdBefore = calendar.compare(habit.createdAt, to: today, toGranularity: .day) != .orderedDescending
            let isActive = habit.archivedAt == nil
            return createdBefore && isActive
        }
    }

    private var habitsCompletedToday: Int {
        habitsForToday.filter { habit in
            if let target = habit.targetValue, target > 0 {
                let progress = habit.progressValues[todayString] ?? 0
                return progress >= target || habit.completedDates.contains(todayString)
            }
            return habit.completedDates.contains(todayString)
        }.count
    }

    private var habitsTotalToday: Int {
        habitsForToday.count
    }

    private var habitsProgress: Double {
        guard habitsTotalToday > 0 else { return 0 }
        return Double(habitsCompletedToday) / Double(habitsTotalToday)
    }

    private var pendingHabits: [Habit] {
        habitsForToday.filter { habit in
            if let target = habit.targetValue, target > 0 {
                let progress = habit.progressValues[todayString] ?? 0
                return progress < target && !habit.completedDates.contains(todayString)
            }
            return !habit.completedDates.contains(todayString)
        }
    }

    private var tasksForToday: [DailyTask] {
        dataManager.tasksForDate(today)
    }

    private var tasksCompletedToday: Int {
        tasksForToday.filter { $0.isCompleted }.count
    }

    private var tasksTotalToday: Int {
        tasksForToday.count
    }

    private var tasksProgress: Double {
        guard tasksTotalToday > 0 else { return 0 }
        return Double(tasksCompletedToday) / Double(tasksTotalToday)
    }

    private var pendingTasks: [DailyTask] {
        tasksForToday.filter { !$0.isCompleted }
            .sorted { $0.priority.sortOrder < $1.priority.sortOrder }
    }

    private var monthTransactions: [Transaction] {
        dataManager.transactions.filter {
            calendar.isDate($0.date, equalTo: today, toGranularity: .month)
        }
    }

    private var monthIncome: Double {
        monthTransactions.filter { $0.type == .income }.reduce(0) { $0 + $1.amount }
    }

    private var monthExpenses: Double {
        monthTransactions.filter { $0.type == .expense }.reduce(0) { $0 + $1.amount }
    }

    private var monthBalance: Double {
        monthIncome - monthExpenses
    }

    private var topStreakHabits: [Habit] {
        habitsForToday
            .filter { $0.streak > 0 }
            .sorted { $0.streak > $1.streak }
    }
}

// MARK: - Today Progress Card

struct TodayProgressCard: View {
    let habitsCompleted: Int
    let habitsTotal: Int
    let tasksCompleted: Int
    let tasksTotal: Int
    let monthBalance: Double

    private var totalItems: Int {
        habitsTotal + tasksTotal
    }

    private var completedItems: Int {
        habitsCompleted + tasksCompleted
    }

    private var progress: Double {
        guard totalItems > 0 else { return 0 }
        return Double(completedItems) / Double(totalItems)
    }

    private var greeting: String {
        let hour = Calendar.current.component(.hour, from: Date())
        if hour < 12 {
            return "Good morning"
        } else if hour < 17 {
            return "Good afternoon"
        } else {
            return "Good evening"
        }
    }

    private var statusMessage: String {
        if totalItems == 0 {
            return "No habits or tasks for today"
        } else if completedItems == totalItems {
            return "All done! Great job today"
        } else {
            let remaining = totalItems - completedItems
            return "\(remaining) item\(remaining == 1 ? "" : "s") remaining"
        }
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            // Greeting and status
            VStack(alignment: .leading, spacing: 4) {
                Text(greeting)
                    .font(.system(size: 22, weight: .bold))
                    .foregroundStyle(.primary)

                Text(statusMessage)
                    .font(.system(size: 15))
                    .foregroundStyle(.secondary)
            }

            // Progress bar
            if totalItems > 0 {
                VStack(alignment: .leading, spacing: 8) {
                    GeometryReader { geometry in
                        ZStack(alignment: .leading) {
                            RoundedRectangle(cornerRadius: 4)
                                .fill(Color.hf.accent.opacity(0.15))
                                .frame(height: 8)

                            RoundedRectangle(cornerRadius: 4)
                                .fill(Color.hf.accent)
                                .frame(width: geometry.size.width * progress, height: 8)
                                .animation(.easeInOut(duration: 0.3), value: progress)
                        }
                    }
                    .frame(height: 8)

                    HStack {
                        Text("\(completedItems) of \(totalItems) completed")
                            .font(.system(size: 13))
                            .foregroundStyle(.secondary)

                        Spacer()

                        Text("\(Int(progress * 100))%")
                            .font(.system(size: 13, weight: .semibold))
                            .foregroundStyle(Color.hf.accent)
                    }
                }
            }
        }
        .padding(16)
        .background(Color.hf.cardBackground)
        .clipShape(RoundedRectangle(cornerRadius: 16))
    }
}

// MARK: - Quick Stat Card

struct QuickStatCard: View {
    let title: String
    let value: String
    let subtitle: String
    let icon: String
    let color: Color

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Image(systemName: icon)
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundStyle(color)

                Spacer()

                Text(subtitle)
                    .font(.system(size: 12, weight: .medium))
                    .foregroundStyle(color)
            }

            Text(value)
                .font(.system(size: 24, weight: .bold, design: .rounded))
                .foregroundStyle(.primary)

            Text(title)
                .font(.system(size: 13))
                .foregroundStyle(.secondary)
        }
        .padding(14)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(Color.hf.cardBackground)
        .clipShape(RoundedRectangle(cornerRadius: 14))
    }
}

// MARK: - Pending Item Row

enum PendingItemType {
    case habit
    case task(priority: TaskPriority)
}

struct PendingItemRow: View {
    let icon: String
    let title: String
    let type: PendingItemType

    var body: some View {
        HStack(spacing: 12) {
            switch type {
            case .habit:
                Text(icon)
                    .font(.system(size: 18))
            case .task(let priority):
                Image(systemName: icon)
                    .font(.system(size: 16))
                    .foregroundStyle(priority.color.opacity(0.7))
            }

            Text(title)
                .font(.system(size: 15))
                .foregroundStyle(.primary)
                .lineLimit(1)

            Spacer()

            switch type {
            case .habit:
                Text("Habit")
                    .font(.system(size: 11, weight: .medium))
                    .foregroundStyle(Color.hf.accent)
                    .padding(.horizontal, 8)
                    .padding(.vertical, 4)
                    .background(Color.hf.accent.opacity(0.1))
                    .clipShape(Capsule())
            case .task(let priority):
                Text(priority.rawValue)
                    .font(.system(size: 11, weight: .medium))
                    .foregroundStyle(priority.color)
                    .padding(.horizontal, 8)
                    .padding(.vertical, 4)
                    .background(priority.color.opacity(0.1))
                    .clipShape(Capsule())
            }
        }
        .padding(.horizontal, 14)
        .padding(.vertical, 12)
        .background(Color.hf.cardBackground)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

// MARK: - Month Budget Card

struct MonthBudgetCard: View {
    let income: Double
    let expenses: Double
    let currency: Currency

    private var balance: Double {
        income - expenses
    }

    private var monthName: String {
        let formatter = DateFormatter()
        formatter.dateFormat = "MMMM"
        return formatter.string(from: Date())
    }

    private func formatAmount(_ amount: Double) -> String {
        let absAmount = abs(amount)
        if absAmount >= 1_000_000 {
            return "\(currency.symbol)\(String(format: "%.1fM", absAmount / 1_000_000))"
        } else if absAmount >= 1000 {
            return "\(currency.symbol)\(String(format: "%.1fK", absAmount / 1000))"
        } else {
            return "\(currency.symbol)\(Int(absAmount))"
        }
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Text("\(monthName) Budget")
                    .font(.system(size: 15, weight: .semibold))
                    .foregroundStyle(.primary)

                Spacer()

                Text(balance >= 0 ? "+\(formatAmount(balance))" : "-\(formatAmount(balance))")
                    .font(.system(size: 15, weight: .bold, design: .rounded))
                    .foregroundStyle(balance >= 0 ? Color.hf.income : Color.hf.expense)
            }

            HStack(spacing: 16) {
                HStack(spacing: 6) {
                    Circle()
                        .fill(Color.hf.income)
                        .frame(width: 8, height: 8)
                    Text("Income: \(formatAmount(income))")
                        .font(.system(size: 13))
                        .foregroundStyle(.secondary)
                }

                HStack(spacing: 6) {
                    Circle()
                        .fill(Color.hf.expense)
                        .frame(width: 8, height: 8)
                    Text("Expenses: \(formatAmount(expenses))")
                        .font(.system(size: 13))
                        .foregroundStyle(.secondary)
                }

                Spacer()
            }
        }
        .padding(16)
        .background(Color.hf.cardBackground)
        .clipShape(RoundedRectangle(cornerRadius: 14))
    }
}

// MARK: - Streak Row

struct StreakRow: View {
    let habit: Habit

    var body: some View {
        HStack(spacing: 12) {
            Text(habit.icon)
                .font(.system(size: 18))

            Text(habit.title)
                .font(.system(size: 15))
                .foregroundStyle(.primary)
                .lineLimit(1)

            Spacer()

            HStack(spacing: 4) {
                Image(systemName: "flame.fill")
                    .font(.system(size: 12))
                    .foregroundStyle(.orange)

                Text("\(habit.streak)")
                    .font(.system(size: 14, weight: .semibold, design: .rounded))
                    .foregroundStyle(.primary)
            }
            .padding(.horizontal, 10)
            .padding(.vertical, 6)
            .background(Color.orange.opacity(0.1))
            .clipShape(Capsule())
        }
        .padding(.horizontal, 14)
        .padding(.vertical, 12)
        .background(Color.hf.cardBackground)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

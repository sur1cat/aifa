import Foundation

// MARK: - Habit Insight Generator

struct HabitInsightGenerator {

    static func generateInsights(from habits: [Habit]) -> [Insight] {
        var insights: [Insight] = []

        // Need at least one habit with data
        guard !habits.isEmpty else { return insights }

        // Check for streak achievements
        for habit in habits {
            if let streakInsight = checkStreakAchievement(habit) {
                insights.append(streakInsight)
            }
        }

        // Check for best performing habit
        if let bestHabit = findBestPerformingHabit(habits) {
            insights.append(bestHabit)
        }

        // Check for struggling habit
        if let strugglingInsight = findStrugglingHabit(habits) {
            insights.append(strugglingInsight)
        }

        // Check for weekly patterns
        if let patternInsight = detectWeeklyPattern(habits) {
            insights.append(patternInsight)
        }

        return insights
    }

    private static func checkStreakAchievement(_ habit: Habit) -> Insight? {
        let streak = habit.streak

        // Milestone streaks
        let milestones = [7, 14, 21, 30, 60, 90, 100, 365]

        for milestone in milestones {
            if streak == milestone {
                return Insight(
                    section: .habits,
                    type: .achievement,
                    title: "\(milestone) Day Streak!",
                    message: "\(habit.icon) \(habit.title) — you've been consistent for \(milestone) days straight!",
                    action: nil
                )
            }
        }

        return nil
    }

    private static func findBestPerformingHabit(_ habits: [Habit]) -> Insight? {
        // Calculate completion rate for last 14 days
        let calendar = Calendar.current
        var habitRates: [(habit: Habit, rate: Double)] = []

        for habit in habits {
            var completedDays = 0
            for dayOffset in 0..<14 {
                guard let date = calendar.date(byAdding: .day, value: -dayOffset, to: Date()) else { continue }
                let dateStr = Habit.dateString(from: date)
                if habit.completedDates.contains(dateStr) {
                    completedDays += 1
                }
            }
            let rate = Double(completedDays) / 14.0
            habitRates.append((habit, rate))
        }

        // Find best habit with at least 70% completion
        guard let best = habitRates.max(by: { $0.rate < $1.rate }),
              best.rate >= 0.7 else { return nil }

        let percentage = Int(best.rate * 100)

        return Insight(
            section: .habits,
            type: .achievement,
            title: "Top Performer",
            message: "\(best.habit.icon) \(best.habit.title) has \(percentage)% completion rate in the last 2 weeks!",
            action: nil
        )
    }

    private static func findStrugglingHabit(_ habits: [Habit]) -> Insight? {
        let calendar = Calendar.current
        var habitRates: [(habit: Habit, rate: Double)] = []

        for habit in habits {
            var completedDays = 0
            for dayOffset in 0..<14 {
                guard let date = calendar.date(byAdding: .day, value: -dayOffset, to: Date()) else { continue }
                let dateStr = Habit.dateString(from: date)
                if habit.completedDates.contains(dateStr) {
                    completedDays += 1
                }
            }
            let rate = Double(completedDays) / 14.0
            habitRates.append((habit, rate))
        }

        // Find struggling habit with less than 30% completion
        guard let worst = habitRates.min(by: { $0.rate < $1.rate }),
              worst.rate < 0.3,
              worst.rate > 0 else { return nil }

        return Insight(
            section: .habits,
            type: .warning,
            title: "Needs Attention",
            message: "\(worst.habit.icon) \(worst.habit.title) has been challenging lately. Would a different time work better?",
            action: nil
        )
    }

    private static func detectWeeklyPattern(_ habits: [Habit]) -> Insight? {
        // Find which day of week has lowest completion
        let calendar = Calendar.current
        var dayCompletions: [Int: (completed: Int, total: Int)] = [:]

        // Initialize all days
        for day in 1...7 {
            dayCompletions[day] = (0, 0)
        }

        for habit in habits {
            for dayOffset in 0..<28 { // Look at 4 weeks
                guard let date = calendar.date(byAdding: .day, value: -dayOffset, to: Date()) else { continue }
                let weekday = calendar.component(.weekday, from: date)
                let dateStr = Habit.dateString(from: date)

                var current = dayCompletions[weekday] ?? (0, 0)
                current.total += 1
                if habit.completedDates.contains(dateStr) {
                    current.completed += 1
                }
                dayCompletions[weekday] = current
            }
        }

        // Find worst day
        var worstDay = 1
        var worstRate = 1.0

        for (day, counts) in dayCompletions {
            guard counts.total > 0 else { continue }
            let rate = Double(counts.completed) / Double(counts.total)
            if rate < worstRate {
                worstRate = rate
                worstDay = day
            }
        }

        // Only show if significantly worse (< 40%)
        guard worstRate < 0.4 else { return nil }

        let dayName = dayOfWeekName(worstDay)
        let percentage = Int(worstRate * 100)

        return Insight(
            section: .habits,
            type: .pattern,
            title: "Weekly Pattern",
            message: "\(dayName) is your hardest day — only \(percentage)% completion. Plan ahead!",
            action: nil
        )
    }

    private static func dayOfWeekName(_ weekday: Int) -> String {
        let formatter = DateFormatter()
        formatter.locale = Locale.current
        return formatter.weekdaySymbols[weekday - 1]
    }
}

// MARK: - Task Insight Generator

struct TaskInsightGenerator {

    static func generateInsights(from tasks: [DailyTask], allTasks: [DailyTask] = []) -> [Insight] {
        var insights: [Insight] = []

        // Filter to last 14 days for relevant insights
        let calendar = Calendar.current
        let twoWeeksAgo = calendar.date(byAdding: .day, value: -14, to: Date()) ?? Date()
        let recentTasks = tasks.filter { $0.dueDate >= twoWeeksAgo }

        // Check completion rate for last 2 weeks
        if let rateInsight = checkCompletionRate(recentTasks) {
            insights.append(rateInsight)
        }

        // Check for best day pattern
        if let patternInsight = findBestDayPattern(tasks) {
            insights.append(patternInsight)
        }

        // Check priority effectiveness
        if let priorityInsight = checkPriorityEffectiveness(recentTasks) {
            insights.append(priorityInsight)
        }

        return insights
    }

    private static func checkCompletionRate(_ tasks: [DailyTask]) -> Insight? {
        guard tasks.count >= 5 else { return nil }

        let completed = tasks.filter { $0.isCompleted }.count
        let total = tasks.count
        let rate = Double(completed) / Double(total)
        let percentage = Int(rate * 100)

        if rate >= 0.9 {
            return Insight(
                section: .tasks,
                type: .achievement,
                title: "Great 2 Weeks!",
                message: "\(completed)/\(total) tasks done (\(percentage)%) in the last 2 weeks",
                action: nil
            )
        } else if rate >= 0.7 {
            return Insight(
                section: .tasks,
                type: .pattern,
                title: "Good Progress",
                message: "\(percentage)% completion rate in the last 2 weeks (\(completed)/\(total))",
                action: nil
            )
        } else if rate < 0.5 {
            return Insight(
                section: .tasks,
                type: .suggestion,
                title: "Focus Tip",
                message: "Only \(percentage)% tasks done in 2 weeks. Try fewer daily tasks.",
                action: nil
            )
        }

        return nil
    }

    private static func findBestDayPattern(_ tasks: [DailyTask]) -> Insight? {
        let calendar = Calendar.current
        var dayCompletions: [Int: (completed: Int, total: Int)] = [:]

        // Initialize all days
        for day in 1...7 {
            dayCompletions[day] = (0, 0)
        }

        for task in tasks {
            let weekday = calendar.component(.weekday, from: task.dueDate)
            var current = dayCompletions[weekday] ?? (0, 0)
            current.total += 1
            if task.isCompleted {
                current.completed += 1
            }
            dayCompletions[weekday] = current
        }

        // Find best day
        var bestDay = 1
        var bestRate = 0.0

        for (day, counts) in dayCompletions {
            guard counts.total >= 2 else { continue }
            let rate = Double(counts.completed) / Double(counts.total)
            if rate > bestRate {
                bestRate = rate
                bestDay = day
            }
        }

        guard bestRate >= 0.8 else { return nil }

        let dayName = dayOfWeekName(bestDay)
        let percentage = Int(bestRate * 100)

        return Insight(
            section: .tasks,
            type: .pattern,
            title: "Best Day",
            message: "\(dayName) is your most productive day — \(percentage)% completion",
            action: nil
        )
    }

    private static func checkPriorityEffectiveness(_ tasks: [DailyTask]) -> Insight? {
        let highPriorityTasks = tasks.filter { $0.priority == .high || $0.priority == .urgent }
        guard highPriorityTasks.count >= 3 else { return nil }

        let completedHigh = highPriorityTasks.filter { $0.isCompleted }.count
        let rate = Double(completedHigh) / Double(highPriorityTasks.count)

        if rate >= 0.9 {
            return Insight(
                section: .tasks,
                type: .achievement,
                title: "Priorities Done",
                message: "\(completedHigh)/\(highPriorityTasks.count) high priority tasks completed",
                action: nil
            )
        } else if rate < 0.5 {
            let pending = highPriorityTasks.count - completedHigh
            return Insight(
                section: .tasks,
                type: .warning,
                title: "High Priority Pending",
                message: "\(pending) high priority tasks still open. Focus on these first!",
                action: nil
            )
        }

        return nil
    }

    private static func dayOfWeekName(_ weekday: Int) -> String {
        let formatter = DateFormatter()
        formatter.locale = Locale.current
        return formatter.weekdaySymbols[weekday - 1]
    }
}

// MARK: - Budget Insight Generator

struct BudgetInsightGenerator {

    static func generateInsights(from transactions: [Transaction], currency: Currency) -> [Insight] {
        var insights: [Insight] = []

        let calendar = Calendar.current
        let thisMonth = transactions.filter { calendar.isDate($0.date, equalTo: Date(), toGranularity: .month) }

        // Check top spending category
        if let categoryInsight = findTopCategory(thisMonth, currency: currency) {
            insights.append(categoryInsight)
        }

        // Check income vs expenses
        if let balanceInsight = checkIncomeVsExpenses(thisMonth, currency: currency) {
            insights.append(balanceInsight)
        }

        // Check for large expense
        if let largeExpenseInsight = findLargeExpense(thisMonth, currency: currency) {
            insights.append(largeExpenseInsight)
        }

        // Compare to last month
        if let comparisonInsight = compareToLastMonth(transactions, currency: currency) {
            insights.append(comparisonInsight)
        }

        return insights
    }

    private static func findTopCategory(_ transactions: [Transaction], currency: Currency) -> Insight? {
        let expenses = transactions.filter { $0.type == .expense }
        guard expenses.count >= 5 else { return nil }

        // Group by category (using title words as pseudo-category)
        var categoryTotals: [String: Double] = [:]

        for expense in expenses {
            let category = expense.category.isEmpty ? "Other" : expense.category
            categoryTotals[category, default: 0] += expense.amount
        }

        guard let topCategory = categoryTotals.max(by: { $0.value < $1.value }) else { return nil }

        let totalExpenses = expenses.reduce(0) { $0 + $1.amount }
        let percentage = Int((topCategory.value / totalExpenses) * 100)

        guard percentage >= 20 else { return nil }

        return Insight(
            section: .budget,
            type: .pattern,
            title: "Top Spending",
            message: "\(topCategory.key) is \(percentage)% of your expenses (\(formatCurrency(topCategory.value, currency: currency)) this month)",
            action: nil
        )
    }

    private static func checkIncomeVsExpenses(_ transactions: [Transaction], currency: Currency) -> Insight? {
        let income = transactions.filter { $0.type == .income }.reduce(0) { $0 + $1.amount }
        let expenses = transactions.filter { $0.type == .expense }.reduce(0) { $0 + $1.amount }

        guard income > 0 else { return nil }

        let savingsRate = (income - expenses) / income
        let savingsPercentage = Int(savingsRate * 100)

        if savingsRate >= 0.2 {
            return Insight(
                section: .budget,
                type: .achievement,
                title: "Great Savings!",
                message: "You're saving \(savingsPercentage)% of your income this month. Keep it up!",
                action: nil
            )
        } else if savingsRate < 0 {
            let overspend = formatCurrency(expenses - income, currency: currency)
            return Insight(
                section: .budget,
                type: .warning,
                title: "Overspending",
                message: "You've spent \(overspend) more than earned this month.",
                action: nil
            )
        }

        return nil
    }

    private static func findLargeExpense(_ transactions: [Transaction], currency: Currency) -> Insight? {
        let expenses = transactions.filter { $0.type == .expense }
        guard !expenses.isEmpty else { return nil }

        let totalExpenses = expenses.reduce(0) { $0 + $1.amount }
        let averageExpense = totalExpenses / Double(expenses.count)

        // Find expense significantly above average (3x or more)
        guard let largeExpense = expenses.first(where: { $0.amount >= averageExpense * 3 && $0.amount >= 100 }) else {
            return nil
        }

        return Insight(
            section: .budget,
            type: .pattern,
            title: "Large Expense",
            message: "\(largeExpense.title): \(formatCurrency(largeExpense.amount, currency: currency)) — that's 3x your average expense",
            action: nil
        )
    }

    private static func compareToLastMonth(_ transactions: [Transaction], currency: Currency) -> Insight? {
        let calendar = Calendar.current

        guard let lastMonthDate = calendar.date(byAdding: .month, value: -1, to: Date()) else { return nil }

        let thisMonthExpenses = transactions
            .filter { $0.type == .expense && calendar.isDate($0.date, equalTo: Date(), toGranularity: .month) }
            .reduce(0) { $0 + $1.amount }

        let lastMonthExpenses = transactions
            .filter { $0.type == .expense && calendar.isDate($0.date, equalTo: lastMonthDate, toGranularity: .month) }
            .reduce(0) { $0 + $1.amount }

        guard lastMonthExpenses > 0 else { return nil }

        let change = (thisMonthExpenses - lastMonthExpenses) / lastMonthExpenses
        let changePercentage = Int(abs(change) * 100)

        if change >= 0.2 {
            return Insight(
                section: .budget,
                type: .warning,
                title: "Spending Up",
                message: "Expenses are up \(changePercentage)% compared to last month. Review your spending!",
                action: nil
            )
        } else if change <= -0.2 {
            return Insight(
                section: .budget,
                type: .achievement,
                title: "Spending Down",
                message: "Great! Expenses are down \(changePercentage)% compared to last month",
                action: nil
            )
        }

        return nil
    }

    private static func formatCurrency(_ amount: Double, currency: Currency) -> String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        formatter.currencySymbol = currency.symbol
        formatter.maximumFractionDigits = currency == .jpy || currency == .krw ? 0 : 2
        return formatter.string(from: NSNumber(value: amount)) ?? "\(currency.symbol)\(Int(amount))"
    }
}

// MARK: - Weekly Review Generator

struct WeeklyReviewGenerator {

    static func generateReview(
        habits: [Habit],
        transactions: [Transaction],
        currency: Currency
    ) -> WeeklyReview {
        let calendar = Calendar.current

        // Calculate current week boundaries (Monday to Sunday)
        let today = Date()
        let weekday = calendar.component(.weekday, from: today)
        let daysFromMonday = (weekday == 1) ? 6 : weekday - 2  // Sunday = 1 in Calendar

        guard let thisMonday = calendar.date(byAdding: .day, value: -daysFromMonday, to: today),
              let thisSunday = calendar.date(byAdding: .day, value: 6, to: thisMonday) else {
            return WeeklyReview(weekStart: today, weekEnd: today)
        }

        // Previous week for comparison
        guard let prevSunday = calendar.date(byAdding: .day, value: -1, to: thisMonday),
              let prevMonday = calendar.date(byAdding: .day, value: -7, to: thisMonday) else {
            return WeeklyReview(weekStart: thisMonday, weekEnd: thisSunday)
        }

        // --- Habits Analysis ---
        let (habitsRate, habitsCompleted, bestHabit, streak) = analyzeHabits(
            habits: habits,
            weekStart: thisMonday,
            weekEnd: thisSunday
        )

        let (prevHabitsRate, _, _, _) = analyzeHabits(
            habits: habits,
            weekStart: prevMonday,
            weekEnd: prevSunday
        )

        let habitsChange = prevHabitsRate > 0 ? (habitsRate - prevHabitsRate) / prevHabitsRate : 0

        // --- Tasks Analysis (we don't have historical tasks, so use estimates) ---
        // For now, tasks completion is based on current day only
        let tasksRate = 0.0
        let tasksChange = 0.0
        let tasksCompleted = 0
        let tasksCreated = 0

        // --- Budget Analysis ---
        let weekTransactions = transactions.filter { tx in
            tx.date >= thisMonday && tx.date <= thisSunday
        }

        let prevWeekTransactions = transactions.filter { tx in
            tx.date >= prevMonday && tx.date <= prevSunday
        }

        let weekIncome = weekTransactions.filter { $0.type == .income }.reduce(0) { $0 + $1.amount }
        let weekExpenses = weekTransactions.filter { $0.type == .expense }.reduce(0) { $0 + $1.amount }

        let prevExpenses = prevWeekTransactions.filter { $0.type == .expense }.reduce(0) { $0 + $1.amount }
        let expensesChange = prevExpenses > 0 ? (weekExpenses - prevExpenses) / prevExpenses : 0

        // Top category
        let (topCategory, topAmount) = findTopCategory(weekTransactions)

        // --- Generate Wins and Warnings ---
        var wins: [String] = []
        var warnings: [String] = []

        // Habits wins/warnings
        if habitsRate >= 0.8 {
            wins.append("Habits: \(Int(habitsRate * 100))% completion rate!")
        } else if habitsRate < 0.5 && habitsRate > 0 {
            warnings.append("Habits dropped to \(Int(habitsRate * 100))%")
        }

        if let best = bestHabit {
            wins.append("\(best) — perfect week!")
        }

        if let s = streak, s >= 7 {
            wins.append("\(s) day streak achieved!")
        }

        // Budget wins/warnings
        if weekIncome > 0 {
            let savingsRate = (weekIncome - weekExpenses) / weekIncome
            if savingsRate >= 0.3 {
                wins.append("Saved \(Int(savingsRate * 100))% of income")
            } else if savingsRate < 0 {
                warnings.append("Spent more than earned")
            }
        }

        if expensesChange >= 0.3 {
            warnings.append("Spending up \(Int(expensesChange * 100))% vs last week")
        } else if expensesChange <= -0.2 {
            wins.append("Spending down \(Int(abs(expensesChange) * 100))%")
        }

        return WeeklyReview(
            weekStart: thisMonday,
            weekEnd: thisSunday,
            habitsCompletionRate: habitsRate,
            habitsCompletionRateChange: habitsChange,
            totalHabitsCompleted: habitsCompleted,
            bestHabit: bestHabit,
            habitStreak: streak,
            tasksCompletionRate: tasksRate,
            tasksCompletionRateChange: tasksChange,
            totalTasksCompleted: tasksCompleted,
            totalTasksCreated: tasksCreated,
            totalIncome: weekIncome,
            totalExpenses: weekExpenses,
            expensesChange: expensesChange,
            topCategory: topCategory,
            topCategoryAmount: topAmount,
            wins: wins,
            warnings: warnings
        )
    }

    private static func analyzeHabits(
        habits: [Habit],
        weekStart: Date,
        weekEnd: Date
    ) -> (rate: Double, completed: Int, bestHabit: String?, streak: Int?) {
        guard !habits.isEmpty else { return (0, 0, nil, nil) }

        let calendar = Calendar.current
        var totalPossible = 0
        var totalCompleted = 0
        var habitCompletions: [(habit: Habit, count: Int)] = []

        for habit in habits {
            var possibleDays = 0
            var completedDays = 0

            var currentDate = weekStart
            while currentDate <= weekEnd {
                let dateStr = Habit.dateString(from: currentDate)

                // Check if habit was applicable this day
                switch habit.period {
                case .daily:
                    possibleDays += 1
                    if habit.completedDates.contains(dateStr) {
                        completedDays += 1
                    }
                case .weekly:
                    // For weekly habits, count once per week
                    if calendar.component(.weekday, from: currentDate) == 2 { // Monday
                        possibleDays += 1
                        if habit.completedDates.contains(where: { Habit.areDatesInSameWeek($0, dateStr) }) {
                            completedDays += 1
                        }
                    }
                case .monthly:
                    break // Skip monthly for weekly review
                }

                currentDate = calendar.date(byAdding: .day, value: 1, to: currentDate) ?? currentDate
            }

            totalPossible += possibleDays
            totalCompleted += completedDays
            habitCompletions.append((habit, completedDays))
        }

        let rate = totalPossible > 0 ? Double(totalCompleted) / Double(totalPossible) : 0

        // Find best habit (100% completion for the week)
        let bestHabit = habitCompletions
            .filter { $0.count == 7 }
            .first
            .map { "\($0.habit.icon) \($0.habit.title)" }

        // Find max streak
        let maxStreak = habits.map { $0.streak }.max()

        return (rate, totalCompleted, bestHabit, maxStreak)
    }

    private static func findTopCategory(_ transactions: [Transaction]) -> (String?, Double?) {
        let expenses = transactions.filter { $0.type == .expense }
        guard !expenses.isEmpty else { return (nil, nil) }

        var categoryTotals: [String: Double] = [:]
        for expense in expenses {
            let category = expense.category.isEmpty ? "Other" : expense.category
            categoryTotals[category, default: 0] += expense.amount
        }

        guard let top = categoryTotals.max(by: { $0.value < $1.value }) else {
            return (nil, nil)
        }

        return (top.key, top.value)
    }
}

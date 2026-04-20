import XCTest
@testable import HabitFlow

/// Tests for feature managers (local operations only, no network)
final class ManagerTests: XCTestCase {

    // MARK: - TasksManager Tests

    func testTasksManagerAddAndRemove() async {
        let manager = await TasksManager()

        // Add task
        let task = MockData.createTask(title: "Test Task")
        await MainActor.run {
            manager.items.append(task)
        }

        // Verify added
        let count = await MainActor.run { manager.items.count }
        XCTAssertEqual(count, 1)

        // Remove task
        await MainActor.run {
            manager.items.removeAll { $0.id == task.id }
        }

        // Verify removed
        let countAfter = await MainActor.run { manager.items.count }
        XCTAssertEqual(countAfter, 0)
    }

    func testTasksManagerSorted() async {
        let manager = await TasksManager()

        await MainActor.run {
            manager.items = [
                MockData.createTask(title: "Low", priority: .low),
                MockData.createTask(title: "High", priority: .high),
                MockData.createTask(title: "Medium", priority: .medium)
            ]
        }

        let sorted = await MainActor.run { manager.sorted }

        // High priority should come first
        XCTAssertEqual(sorted[0].priority, .high)
        XCTAssertEqual(sorted[1].priority, .medium)
        XCTAssertEqual(sorted[2].priority, .low)
    }

    func testTasksManagerCompletionRate() async {
        let manager = await TasksManager()

        await MainActor.run {
            manager.items = [
                MockData.createTask(title: "Done", isCompleted: true),
                MockData.createTask(title: "Done2", isCompleted: true),
                MockData.createTask(title: "Not Done", isCompleted: false)
            ]
        }

        let rate = await MainActor.run { manager.completionRate() }

        // 2/3 = 66.67%
        XCTAssertEqual(rate, 200.0 / 3.0, accuracy: 0.01)
    }

    func testTasksManagerTasksForDate() async {
        let manager = await TasksManager()
        let today = Date()
        let yesterday = Calendar.current.date(byAdding: .day, value: -1, to: today)!

        await MainActor.run {
            manager.items = [
                DailyTask(title: "Today", isCompleted: false, priority: .high, dueDate: today),
                DailyTask(title: "Yesterday", isCompleted: false, priority: .high, dueDate: yesterday)
            ]
        }

        let todayTasks = await MainActor.run { manager.tasksForDate(today) }
        XCTAssertEqual(todayTasks.count, 1)
        XCTAssertEqual(todayTasks[0].title, "Today")
    }

    // MARK: - HabitsManager Tests

    func testHabitsManagerActiveAndArchived() async {
        let manager = await HabitsManager()

        let activeHabit = MockData.createHabit(title: "Active")
        var archivedHabit = MockData.createHabit(title: "Archived")
        archivedHabit.archivedAt = Date()

        await MainActor.run {
            manager.items = [activeHabit, archivedHabit]
        }

        let active = await MainActor.run { manager.active }
        let archived = await MainActor.run { manager.archived }

        XCTAssertEqual(active.count, 1)
        XCTAssertEqual(active[0].title, "Active")
        XCTAssertEqual(archived.count, 1)
        XCTAssertEqual(archived[0].title, "Archived")
    }

    func testHabitsManagerCompletionRate() async {
        let manager = await HabitsManager()
        let today = Habit.todayString

        await MainActor.run {
            manager.items = [
                Habit(title: "Done", icon: "✅", color: "green", period: .daily, completedDates: [today]),
                Habit(title: "Not Done", icon: "❌", color: "red", period: .daily, completedDates: [])
            ]
        }

        let rate = await MainActor.run { manager.completionRate(for: .week) }

        // 1/2 = 50%
        XCTAssertEqual(rate, 50.0, accuracy: 0.01)
    }

    func testHabitsManagerClearGoalLinks() async {
        let manager = await HabitsManager()
        let goalId = UUID()

        await MainActor.run {
            var habit1 = MockData.createHabit(title: "With Goal")
            habit1.goalId = goalId
            var habit2 = MockData.createHabit(title: "Different Goal")
            habit2.goalId = UUID()

            manager.items = [habit1, habit2]
        }

        await MainActor.run {
            manager.clearGoalLinks(goalId: goalId)
        }

        let items = await MainActor.run { manager.items }
        XCTAssertNil(items[0].goalId, "Goal link should be cleared")
        XCTAssertNotNil(items[1].goalId, "Other goal link should remain")
    }

    // MARK: - GoalsManager Tests

    func testGoalsManagerActiveAndArchived() async {
        let manager = await GoalsManager()

        let activeGoal = Goal(title: "Active Goal", icon: "🎯", targetValue: 100, unit: "steps")
        var archivedGoal = Goal(title: "Archived Goal", icon: "📦", targetValue: 50, unit: "items")
        archivedGoal.archivedAt = Date()

        await MainActor.run {
            manager.items = [activeGoal, archivedGoal]
        }

        let active = await MainActor.run { manager.active }
        let archived = await MainActor.run { manager.archived }

        XCTAssertEqual(active.count, 1)
        XCTAssertEqual(archived.count, 1)
    }

    // MARK: - BudgetManager Tests

    func testBudgetManagerBalance() async {
        let manager = await BudgetManager()

        await MainActor.run {
            manager.transactions = [
                MockData.createTransaction(title: "Income", amount: 1000, type: .income),
                MockData.createTransaction(title: "Expense", amount: 300, type: .expense),
                MockData.createTransaction(title: "Expense2", amount: 200, type: .expense)
            ]
        }

        let balance = await MainActor.run { manager.balance }

        // 1000 - 300 - 200 = 500
        XCTAssertEqual(balance, 500.0, accuracy: 0.01)
    }

    func testBudgetManagerMonthlyTotals() async {
        let manager = await BudgetManager()

        // Create transactions for current month
        await MainActor.run {
            manager.transactions = [
                Transaction(title: "Salary", amount: 5000, type: .income, category: "Salary", date: Date()),
                Transaction(title: "Rent", amount: 1500, type: .expense, category: "Housing", date: Date()),
                Transaction(title: "Food", amount: 500, type: .expense, category: "Food", date: Date())
            ]
        }

        let income = await MainActor.run { manager.monthlyIncome }
        let expenses = await MainActor.run { manager.monthlyExpenses }

        XCTAssertEqual(income, 5000.0)
        XCTAssertEqual(expenses, 2000.0)
    }

    func testBudgetManagerTransactionsForPeriod() async {
        let manager = await BudgetManager()
        let today = Date()
        let lastMonth = Calendar.current.date(byAdding: .month, value: -1, to: today)!

        await MainActor.run {
            manager.transactions = [
                Transaction(title: "This Month", amount: 100, type: .expense, category: "Test", date: today),
                Transaction(title: "Last Month", amount: 200, type: .expense, category: "Test", date: lastMonth)
            ]
        }

        let thisMonth = await MainActor.run { manager.transactionsForPeriod(.month) }
        XCTAssertEqual(thisMonth.count, 1)
        XCTAssertEqual(thisMonth[0].title, "This Month")
    }

    func testBudgetManagerProjectedMonthly() async {
        let manager = await BudgetManager()

        await MainActor.run {
            manager.recurringTransactions = [
                RecurringTransaction(
                    title: "Salary",
                    amount: 5000,
                    type: .income,
                    category: .other,
                    frequency: .monthly,
                    startDate: Date(),
                    nextDate: Date(),
                    isActive: true
                ),
                RecurringTransaction(
                    title: "Rent",
                    amount: 1200,
                    type: .expense,
                    category: .utilities,
                    frequency: .monthly,
                    startDate: Date(),
                    nextDate: Date(),
                    isActive: true
                ),
                RecurringTransaction(
                    title: "Inactive",
                    amount: 999,
                    type: .expense,
                    category: .other,
                    frequency: .monthly,
                    startDate: Date(),
                    nextDate: Date(),
                    isActive: false  // Should be ignored
                )
            ]
        }

        let projectedIncome = await MainActor.run { manager.projectedMonthlyIncome }
        let projectedExpenses = await MainActor.run { manager.projectedMonthlyExpenses }

        XCTAssertEqual(projectedIncome, 5000.0)
        XCTAssertEqual(projectedExpenses, 1200.0)
    }

    // MARK: - InsightManager Tests

    func testInsightManagerInsightsForSection() async {
        let manager = await InsightManager()

        await MainActor.run {
            manager.insights = [
                MockData.createInsight(section: .habits, title: "Habit Insight"),
                MockData.createInsight(section: .tasks, title: "Task Insight"),
                MockData.createInsight(section: .budget, title: "Budget Insight")
            ]
        }

        let habitInsights = await MainActor.run { manager.insights(for: .habits) }
        XCTAssertEqual(habitInsights.count, 1)
        XCTAssertEqual(habitInsights[0].title, "Habit Insight")
    }

    func testInsightManagerDismiss() async {
        let manager = await InsightManager()
        let insight = MockData.createInsight(section: .habits, title: "To Dismiss")

        await MainActor.run {
            manager.insights = [insight]
            manager.dismiss(insight)
        }

        let remaining = await MainActor.run { manager.insights.count }
        XCTAssertEqual(remaining, 0)
    }

    // MARK: - AnalyticsManager Tests

    func testAnalyticsManagerLifeScoreEmpty() async {
        let manager = await AnalyticsManager()

        // Without coordinator, should return 0
        let score = await MainActor.run { manager.lifeScore(for: .week) }
        XCTAssertEqual(score, 0.0)
    }
}

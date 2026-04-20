import XCTest
@testable import HabitFlow

final class HabitTests: XCTestCase {

    func testHabitCreation() {
        let habit = Habit(
            title: "Meditation",
            icon: "🧘",
            color: "purple",
            period: .daily
        )

        XCTAssertEqual(habit.title, "Meditation")
        XCTAssertEqual(habit.icon, "🧘")
        XCTAssertEqual(habit.color, "purple")
        XCTAssertEqual(habit.period, .daily)
        XCTAssertTrue(habit.completedDates.isEmpty)
    }

    func testHabitIsActive() {
        var habit = Habit(title: "Test")
        XCTAssertTrue(habit.isActive)

        habit.archivedAt = Date()
        XCTAssertFalse(habit.isActive)
    }

    func testHabitDateString() {
        let date = DateComponents(
            calendar: Calendar.current,
            year: 2024,
            month: 5,
            day: 10
        ).date!

        let dateString = Habit.dateString(from: date)
        XCTAssertEqual(dateString, "2024-05-10")
    }

    func testHabitTodayString() {
        let todayString = Habit.todayString
        let expectedFormat = DateFormatters.apiDate.string(from: Date())
        XCTAssertEqual(todayString, expectedFormat)
    }

    func testHabitIsCompletedInCurrentPeriod_Daily() {
        var habit = Habit(title: "Daily", period: .daily)

        XCTAssertFalse(habit.isCompletedInCurrentPeriod)

        habit.completedDates.append(Habit.todayString)
        XCTAssertTrue(habit.isCompletedInCurrentPeriod)
    }

    func testHabitWithGoal() {
        var habit = Habit(
            title: "Read",
            targetValue: 30,
            unit: "pages"
        )

        XCTAssertTrue(habit.hasGoal)
        XCTAssertEqual(habit.currentProgress, 0)
        XCTAssertEqual(habit.progressPercentage, 0.0)

        habit.progressValues[Habit.todayString] = 15
        XCTAssertEqual(habit.currentProgress, 15)
        XCTAssertEqual(habit.progressPercentage, 0.5)

        habit.progressValues[Habit.todayString] = 30
        XCTAssertEqual(habit.progressPercentage, 1.0)
    }

    func testHabitColor() {
        let habit = Habit(title: "Test", color: "green")
        XCTAssertNotNil(habit.swiftUIColor)
    }
}

final class TransactionTests: XCTestCase {

    func testTransactionCreation() {
        let transaction = Transaction(
            title: "Coffee",
            amount: 5.50,
            type: .expense,
            category: "Food"
        )

        XCTAssertEqual(transaction.title, "Coffee")
        XCTAssertEqual(transaction.amount, 5.50)
        XCTAssertEqual(transaction.type, .expense)
        XCTAssertEqual(transaction.category, "Food")
    }

    func testTransactionTypes() {
        XCTAssertEqual(TransactionType.income.rawValue, "income")
        XCTAssertEqual(TransactionType.expense.rawValue, "expense")
    }
}

final class DailyTaskTests: XCTestCase {

    func testTaskCreation() {
        let task = DailyTask(
            title: "Buy groceries",
            priority: .high
        )

        XCTAssertEqual(task.title, "Buy groceries")
        XCTAssertEqual(task.priority, .high)
        XCTAssertFalse(task.isCompleted)
    }

    func testTaskPriorities() {
        XCTAssertEqual(TaskPriority.low.rawValue, "low")
        XCTAssertEqual(TaskPriority.medium.rawValue, "medium")
        XCTAssertEqual(TaskPriority.high.rawValue, "high")
    }

    func testTaskPriorityColor() {
        XCTAssertNotNil(TaskPriority.low.color)
        XCTAssertNotNil(TaskPriority.medium.color)
        XCTAssertNotNil(TaskPriority.high.color)
    }
}

final class InsightTests: XCTestCase {

    func testInsightCreation() {
        let insight = Insight(
            section: .habits,
            type: .achievement,
            title: "Great streak!",
            message: "You've completed 7 days"
        )

        XCTAssertEqual(insight.section, .habits)
        XCTAssertEqual(insight.type, .achievement)
        XCTAssertEqual(insight.title, "Great streak!")
        XCTAssertFalse(insight.isDismissed)
    }

    func testInsightContentHash() {
        let insight1 = Insight(
            section: .habits,
            type: .achievement,
            title: "Test",
            message: "Message"
        )

        let insight2 = Insight(
            section: .habits,
            type: .achievement,
            title: "Test",
            message: "Message"
        )

        // Same content should produce same hash
        XCTAssertEqual(insight1.contentHash, insight2.contentHash)

        let insight3 = Insight(
            section: .habits,
            type: .achievement,
            title: "Different",
            message: "Message"
        )

        // Different content should produce different hash
        XCTAssertNotEqual(insight1.contentHash, insight3.contentHash)
    }

    func testInsightHashIsDeterministic() {
        let insight = Insight(
            section: .budget,
            type: .warning,
            title: "Budget alert",
            message: "You're over budget"
        )

        let hash1 = insight.contentHash
        let hash2 = insight.contentHash

        XCTAssertEqual(hash1, hash2, "Hash should be deterministic")
    }
}

final class WeeklyReviewTests: XCTestCase {

    func testWeeklyReviewCreation() {
        let now = Date()
        let weekEnd = Calendar.current.date(byAdding: .day, value: 6, to: now)!

        let review = WeeklyReview(
            weekStart: now,
            weekEnd: weekEnd
        )

        XCTAssertEqual(review.weekStart, now)
        XCTAssertEqual(review.weekEnd, weekEnd)
        XCTAssertEqual(review.habitsCompletionRate, 0)
        XCTAssertEqual(review.totalTasksCompleted, 0)
        XCTAssertEqual(review.totalExpenses, 0)
    }
}

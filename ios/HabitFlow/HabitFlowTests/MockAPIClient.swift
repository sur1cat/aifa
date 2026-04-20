import Foundation
@testable import HabitFlow

/// Mock API Client for testing
actor MockAPIClient {
    var shouldFail = false
    var failureError: APIError = .serverError("Mock error")
    var responses: [String: Any] = [:]
    var requestLog: [(endpoint: String, method: String)] = []

    func setResponse<T: Encodable>(for endpoint: String, response: T) {
        responses[endpoint] = response
    }

    func setShouldFail(_ fail: Bool, error: APIError = .serverError("Mock error")) {
        shouldFail = fail
        failureError = error
    }

    func request<T: Decodable>(
        endpoint: String,
        method: String = "GET",
        body: Encodable? = nil,
        requiresAuth: Bool = false
    ) async throws -> T {
        requestLog.append((endpoint, method))

        if shouldFail {
            throw failureError
        }

        guard let response = responses[endpoint] else {
            throw APIError.serverError("No mock response for endpoint: \(endpoint)")
        }

        // Convert response to expected type
        if let typedResponse = response as? T {
            return typedResponse
        }

        throw APIError.invalidResponse
    }

    func getRequestLog() -> [(endpoint: String, method: String)] {
        return requestLog
    }

    func clearLog() {
        requestLog.removeAll()
    }
}

// MARK: - Mock Data Generators

enum MockData {
    static func createHabit(
        title: String = "Test Habit",
        icon: String = "🎯",
        color: String = "green",
        period: HabitPeriod = .daily,
        completedDates: [String] = []
    ) -> Habit {
        Habit(
            title: title,
            icon: icon,
            color: color,
            period: period,
            completedDates: completedDates
        )
    }

    static func createTask(
        title: String = "Test Task",
        priority: TaskPriority = .medium,
        isCompleted: Bool = false
    ) -> DailyTask {
        DailyTask(
            title: title,
            isCompleted: isCompleted,
            priority: priority
        )
    }

    static func createTransaction(
        title: String = "Test Transaction",
        amount: Double = 100.0,
        type: TransactionType = .expense,
        category: String = "Other"
    ) -> Transaction {
        Transaction(
            title: title,
            amount: amount,
            type: type,
            category: category
        )
    }

    static func createInsight(
        section: InsightSection = .habits,
        type: InsightType = .achievement,
        title: String = "Test Insight",
        message: String = "Test message"
    ) -> Insight {
        Insight(
            section: section,
            type: type,
            title: title,
            message: message
        )
    }

    static var sampleHabits: [Habit] {
        [
            createHabit(title: "Meditation", icon: "🧘", color: "purple"),
            createHabit(title: "Exercise", icon: "💪", color: "orange"),
            createHabit(title: "Reading", icon: "📚", color: "blue")
        ]
    }

    static var sampleTasks: [DailyTask] {
        [
            createTask(title: "Buy groceries", priority: .high),
            createTask(title: "Call mom", priority: .medium),
            createTask(title: "Clean room", priority: .low, isCompleted: true)
        ]
    }

    static var sampleTransactions: [Transaction] {
        [
            createTransaction(title: "Salary", amount: 5000, type: .income, category: "Salary"),
            createTransaction(title: "Groceries", amount: 150, type: .expense, category: "Food"),
            createTransaction(title: "Netflix", amount: 15, type: .expense, category: "Entertainment")
        ]
    }
}

import Foundation

// MARK: - API Models

struct AIInsightRequest: Encodable {
    let type: String  // habits, tasks, budget, weekly
    let data: String  // JSON string with user data
}

struct AIInsightItem: Decodable {
    let type: String
    let title: String
    let message: String
}

struct AIInsightResponse: Decodable {
    let insights: [AIInsightItem]
}

struct AIWeeklyResponse: Decodable {
    let summary: String
    let wins: [String]
    let improvements: [String]
    let tip: String
}

// MARK: - Expense Analysis API Models

struct AIExpenseAnalysisRequest: Encodable {
    let data: String
}

struct AIExpenseInsightItem: Decodable {
    let type: String
    let title: String
    let message: String
    let amount: Double?
    let category: String?
    let priority: Int?
}

struct AIQuestionableItem: Decodable {
    let transactionId: String
    let reason: String
    let category: String
    let potentialSavings: Double?
}

struct AISuggestionItem: Decodable {
    let category: String
    let currentSpending: Double
    let suggestedBudget: Double
    let potentialSavings: Double
    let reason: String
    let difficulty: String
}

struct AIExpenseAnalysisResponse: Decodable {
    let insights: [AIExpenseInsightItem]?
    let questionableTransactions: [AIQuestionableItem]?
    let savingsSuggestions: [AISuggestionItem]?
}

// MARK: - Insight Service

@MainActor
class InsightService {
    static let shared = InsightService()
    private let apiClient = APIClient.shared

    private init() {}

    /// Generate AI insights for habits
    func generateHabitsInsights(habits: [Habit]) async throws -> [Insight] {
        let habitsData = prepareHabitsData(habits)
        return try await generateInsights(type: "habits", data: habitsData)
    }

    /// Generate AI insights for tasks
    func generateTasksInsights(tasks: [DailyTask]) async throws -> [Insight] {
        let tasksData = prepareTasksData(tasks)
        return try await generateInsights(type: "tasks", data: tasksData)
    }

    /// Generate AI insights for budget
    func generateBudgetInsights(transactions: [Transaction], currency: Currency) async throws -> [Insight] {
        let budgetData = prepareBudgetData(transactions, currency: currency)
        return try await generateInsights(type: "budget", data: budgetData)
    }

    /// Generate AI weekly review
    func generateWeeklyReview(
        habits: [Habit],
        tasks: [DailyTask],
        transactions: [Transaction],
        currency: Currency
    ) async throws -> (summary: String, wins: [String], improvements: [String], tip: String) {
        let data = prepareWeeklyData(habits: habits, tasks: tasks, transactions: transactions, currency: currency)

        let request = AIInsightRequest(type: "weekly", data: data)

        struct Response: Decodable {
            let data: AIWeeklyResponse
        }

        let response: Response = try await apiClient.request(
            endpoint: "/ai/insights",
            method: "POST",
            body: request
        )

        return (
            summary: response.data.summary,
            wins: response.data.wins,
            improvements: response.data.improvements,
            tip: response.data.tip
        )
    }

    /// Generate detailed AI expense analysis
    func generateExpenseAnalysis(
        transactions: [Transaction],
        recurringTransactions: [RecurringTransaction],
        savingsGoal: SavingsGoal?,
        currency: Currency
    ) async throws -> AIExpenseAnalysis {
        let data = prepareExpenseAnalysisData(
            transactions: transactions,
            recurringTransactions: recurringTransactions,
            savingsGoal: savingsGoal,
            currency: currency
        )

        let request = AIExpenseAnalysisRequest(data: data)

        struct Response: Decodable {
            let data: AIExpenseAnalysisResponse
        }

        let response: Response = try await apiClient.request(
            endpoint: "/ai/expense-analysis",
            method: "POST",
            body: request
        )

        return convertToAnalysis(response.data, transactions: transactions)
    }

    // MARK: - Private

    private func generateInsights(type: String, data: String) async throws -> [Insight] {
        let request = AIInsightRequest(type: type, data: data)

        struct Response: Decodable {
            let data: AIInsightResponse
        }

        let response: Response = try await apiClient.request(
            endpoint: "/ai/insights",
            method: "POST",
            body: request
        )

        // Convert API response to Insight models
        return response.data.insights.map { item in
            let section: InsightSection = {
                switch type {
                case "habits": return .habits
                case "tasks": return .tasks
                case "budget": return .budget
                default: return .habits
                }
            }()

            let insightType: InsightType = {
                switch item.type {
                case "pattern": return .pattern
                case "achievement": return .achievement
                case "warning": return .warning
                case "suggestion": return .suggestion
                default: return .suggestion
                }
            }()

            return Insight(
                section: section,
                type: insightType,
                title: item.title,
                message: item.message,
                action: InsightAction(label: "Ask AI", actionType: "openAIChat")
            )
        }
    }

    // MARK: - Data Preparation

    private func prepareHabitsData(_ habits: [Habit]) -> String {
        var data: [[String: Any]] = []

        for habit in habits {
            let habitData: [String: Any] = [
                "title": habit.title,
                "icon": habit.icon,
                "period": habit.period.rawValue,
                "streak": habit.streak,
                "completedDates": Array(habit.completedDates.suffix(30)), // Already strings
                "reminderEnabled": habit.reminderEnabled,
                "isCompletedToday": habit.isCompletedInCurrentPeriod
            ]
            data.append(habitData)
        }

        return jsonString(from: ["habits": data, "count": habits.count])
    }

    private func prepareTasksData(_ tasks: [DailyTask]) -> String {
        let completed = tasks.filter { $0.isCompleted }.count
        let total = tasks.count

        var taskList: [[String: Any]] = []
        for task in tasks {
            taskList.append([
                "title": task.title,
                "priority": task.priority.rawValue,
                "isCompleted": task.isCompleted
            ])
        }

        return jsonString(from: [
            "tasks": taskList,
            "completedCount": completed,
            "totalCount": total,
            "completionRate": total > 0 ? Double(completed) / Double(total) : 0
        ])
    }

    private func prepareBudgetData(_ transactions: [Transaction], currency: Currency) -> String {
        let calendar = Calendar.current
        let thisMonth = transactions.filter { calendar.isDate($0.date, equalTo: Date(), toGranularity: .month) }

        let income = thisMonth.filter { $0.type == .income }.reduce(0) { $0 + $1.amount }
        let expenses = thisMonth.filter { $0.type == .expense }.reduce(0) { $0 + $1.amount }

        // Group by category
        var categories: [String: Double] = [:]
        for tx in thisMonth where tx.type == .expense {
            let category = tx.category.isEmpty ? "Other" : tx.category
            categories[category, default: 0] += tx.amount
        }

        var recentTransactions: [[String: Any]] = []
        for tx in thisMonth.sorted(by: { $0.date > $1.date }).prefix(20) {
            recentTransactions.append([
                "title": tx.title,
                "amount": tx.amount,
                "type": tx.type.rawValue,
                "category": tx.category
            ])
        }

        return jsonString(from: [
            "currency": currency.symbol,
            "totalIncome": income,
            "totalExpenses": expenses,
            "balance": income - expenses,
            "categorySummary": categories,
            "recentTransactions": recentTransactions
        ])
    }

    private func prepareWeeklyData(
        habits: [Habit],
        tasks: [DailyTask],
        transactions: [Transaction],
        currency: Currency
    ) -> String {
        return jsonString(from: [
            "habits": prepareHabitsData(habits),
            "tasks": prepareTasksData(tasks),
            "budget": prepareBudgetData(transactions, currency: currency)
        ])
    }

    private func jsonString(from dict: [String: Any]) -> String {
        guard let data = try? JSONSerialization.data(withJSONObject: dict),
              let string = String(data: data, encoding: .utf8) else {
            return "{}"
        }
        return string
    }

    // MARK: - Expense Analysis Helpers

    private func prepareExpenseAnalysisData(
        transactions: [Transaction],
        recurringTransactions: [RecurringTransaction],
        savingsGoal: SavingsGoal?,
        currency: Currency
    ) -> String {
        let calendar = Calendar.current

        // Current month data
        let thisMonth = transactions.filter {
            calendar.isDate($0.date, equalTo: Date(), toGranularity: .month)
        }

        // Last month for comparison
        let lastMonthDate = calendar.date(byAdding: .month, value: -1, to: Date())!
        let lastMonth = transactions.filter {
            calendar.isDate($0.date, equalTo: lastMonthDate, toGranularity: .month)
        }

        // Group by category with transaction details
        struct CategoryData {
            var total: Double = 0
            var count: Int = 0
            var transactions: [[String: Any]] = []
        }
        var categoryData: [String: CategoryData] = [:]

        for tx in thisMonth where tx.type == .expense {
            let category = tx.category.isEmpty ? "Other" : tx.category
            var data = categoryData[category] ?? CategoryData()
            data.total += tx.amount
            data.count += 1
            data.transactions.append([
                "id": tx.id.uuidString,
                "title": tx.title,
                "amount": tx.amount,
                "date": DateFormatters.iso8601.string(from: tx.date)
            ])
            categoryData[category] = data
        }

        // Convert to dictionary format for JSON
        let categoryBreakdown: [String: [String: Any]] = categoryData.mapValues { data in
            ["total": data.total, "count": data.count, "transactions": data.transactions]
        }

        // Analyze patterns
        let weekdaySpending = analyzeWeekdaySpending(thisMonth)
        let timeOfDaySpending = analyzeTimeOfDaySpending(thisMonth)

        // Calculate totals
        let thisMonthIncome = thisMonth.filter { $0.type == .income }.reduce(0) { $0 + $1.amount }
        let thisMonthExpenses = thisMonth.filter { $0.type == .expense }.reduce(0) { $0 + $1.amount }
        let lastMonthExpenses = lastMonth.filter { $0.type == .expense }.reduce(0) { $0 + $1.amount }

        // Recurring summary
        let recurringData: [[String: Any]] = recurringTransactions.filter { $0.isActive }.map { rt in
            [
                "title": rt.title,
                "amount": rt.amount,
                "frequency": rt.frequency.rawValue,
                "category": rt.category.rawValue
            ]
        }

        var analysisData: [String: Any] = [
            "currency": currency.symbol,
            "thisMonth": [
                "income": thisMonthIncome,
                "expenses": thisMonthExpenses,
                "savings": thisMonthIncome - thisMonthExpenses,
                "transactionCount": thisMonth.count
            ],
            "lastMonth": [
                "expenses": lastMonthExpenses
            ],
            "categoryBreakdown": categoryBreakdown,
            "patterns": [
                "weekdaySpending": weekdaySpending,
                "timeOfDaySpending": timeOfDaySpending
            ],
            "recurringTransactions": recurringData
        ]

        if let goal = savingsGoal {
            analysisData["savingsGoal"] = [
                "target": goal.monthlyTarget,
                "current": goal.currentSavings,
                "remaining": max(0, goal.monthlyTarget - goal.currentSavings),
                "progress": goal.progress
            ]
        }

        return jsonString(from: analysisData)
    }

    private func analyzeWeekdaySpending(_ transactions: [Transaction]) -> [String: Double] {
        var weekdayTotals: [String: Double] = [:]
        let formatter = DateFormatter()
        formatter.dateFormat = "EEEE"

        for tx in transactions where tx.type == .expense {
            let weekday = formatter.string(from: tx.date)
            weekdayTotals[weekday, default: 0] += tx.amount
        }

        return weekdayTotals
    }

    private func analyzeTimeOfDaySpending(_ transactions: [Transaction]) -> [String: Double] {
        let calendar = Calendar.current
        var timeSlots: [String: Double] = [
            "morning": 0,
            "afternoon": 0,
            "evening": 0,
            "night": 0
        ]

        for tx in transactions where tx.type == .expense {
            let hour = calendar.component(.hour, from: tx.date)
            switch hour {
            case 6..<12: timeSlots["morning"]! += tx.amount
            case 12..<18: timeSlots["afternoon"]! += tx.amount
            case 18..<22: timeSlots["evening"]! += tx.amount
            default: timeSlots["night"]! += tx.amount
            }
        }

        return timeSlots
    }

    private func convertToAnalysis(_ response: AIExpenseAnalysisResponse, transactions: [Transaction]) -> AIExpenseAnalysis {
        // Convert insights
        let insights = (response.insights ?? []).map { item in
            ExpenseInsight(
                type: ExpenseInsightType(rawValue: item.type) ?? .pattern,
                title: item.title,
                message: item.message,
                amount: item.amount,
                category: item.category,
                priority: item.priority ?? 2
            )
        }

        // Convert questionable transactions
        let questionable = (response.questionableTransactions ?? []).compactMap { item -> QuestionableTransaction? in
            guard let txId = UUID(uuidString: item.transactionId) else { return nil }
            return QuestionableTransaction(
                transactionId: txId,
                reason: item.reason,
                category: QuestionableCategory(rawValue: item.category) ?? .unnecessary,
                potentialSavings: item.potentialSavings
            )
        }

        // Convert suggestions
        let suggestions = (response.savingsSuggestions ?? []).map { item in
            SavingsSuggestion(
                category: item.category,
                currentSpending: item.currentSpending,
                suggestedBudget: item.suggestedBudget,
                potentialSavings: item.potentialSavings,
                reason: item.reason,
                difficulty: SuggestionDifficulty(rawValue: item.difficulty) ?? .medium
            )
        }

        return AIExpenseAnalysis(
            insights: insights,
            questionableTransactions: questionable,
            savingsSuggestions: suggestions
        )
    }
}

import Foundation

// MARK: - API Models
struct TransactionAPIResponse: Codable, Sendable {
    let id: String
    let title: String
    let amount: Double
    let type: String
    let category: String
    let date: String
    let createdAt: String

    enum CodingKeys: String, CodingKey {
        case id, title, amount, type, category, date
        case createdAt = "created_at"
    }
}

struct CreateTransactionRequest: Encodable, Sendable {
    let title: String
    let amount: Double
    let type: String
    let category: String
    let date: String
}

struct UpdateTransactionRequest: Encodable, Sendable {
    let title: String?
    let amount: Double?
    let type: String?
    let category: String?
    let date: String?
}

struct SummaryAPIResponse: Codable, Sendable {
    let income: Double
    let expense: Double
    let balance: Double
}

// MARK: - Savings Goal API Models
struct SavingsGoalAPIResponse: Codable, Sendable {
    let id: String
    let monthlyTarget: Double
    let currentSavings: Double
    let monthlyIncome: Double
    let monthlyExpenses: Double
    let progress: Double
    let createdAt: String
    let updatedAt: String
}

struct SetSavingsGoalRequest: Encodable, Sendable {
    let monthlyTarget: Double
}

// MARK: - BudgetService
actor BudgetService {
    static let shared = BudgetService()
    private let api = APIClient.shared

    // MARK: - List Transactions
    func getTransactions(year: Int? = nil, month: Int? = nil) async throws -> [Transaction] {
        var endpoint = "transactions"
        if let year = year, let month = month {
            endpoint += "?year=\(year)&month=\(month)"
        }
        let response: [TransactionAPIResponse] = try await api.request(
            endpoint: endpoint,
            requiresAuth: true
        )
        return response.map { mapToTransaction($0) }
    }

    // MARK: - Create Transaction
    func createTransaction(_ transaction: Transaction) async throws -> Transaction {
        let request = CreateTransactionRequest(
            title: transaction.title,
            amount: transaction.amount,
            type: transaction.type.rawValue,
            category: transaction.category,
            date: DateFormatters.apiDate.string(from: transaction.date)
        )
        let response: TransactionAPIResponse = try await api.request(
            endpoint: "transactions",
            method: "POST",
            body: request,
            requiresAuth: true
        )
        return mapToTransaction(response)
    }

    // MARK: - Update Transaction
    func updateTransaction(_ transaction: Transaction) async throws -> Transaction {
        let request = UpdateTransactionRequest(
            title: transaction.title,
            amount: transaction.amount,
            type: transaction.type.rawValue,
            category: transaction.category,
            date: DateFormatters.apiDate.string(from: transaction.date)
        )
        let response: TransactionAPIResponse = try await api.request(
            endpoint: "transactions/\(transaction.id.uuidString)",
            method: "PUT",
            body: request,
            requiresAuth: true
        )
        return mapToTransaction(response)
    }

    // MARK: - Delete Transaction
    func deleteTransaction(_ id: UUID) async throws {
        let _: EmptyResponse = try await api.request(
            endpoint: "transactions/\(id.uuidString)",
            method: "DELETE",
            requiresAuth: true
        )
    }

    // MARK: - Get Summary
    func getSummary(year: Int, month: Int) async throws -> (income: Double, expense: Double, balance: Double) {
        let response: SummaryAPIResponse = try await api.request(
            endpoint: "transactions/summary?year=\(year)&month=\(month)",
            requiresAuth: true
        )
        return (response.income, response.expense, response.balance)
    }

    // MARK: - Savings Goal

    func getSavingsGoal() async throws -> SavingsGoal? {
        let response: SavingsGoalAPIResponse? = try await api.requestOptional(
            endpoint: "savings-goal",
            requiresAuth: true
        )

        guard let response = response else {
            return nil
        }

        return SavingsGoal(
            id: UUID(uuidString: response.id) ?? UUID(),
            monthlyTarget: response.monthlyTarget,
            currentSavings: response.currentSavings,
            monthlyIncome: response.monthlyIncome,
            monthlyExpenses: response.monthlyExpenses,
            progress: response.progress
        )
    }

    func setSavingsGoal(target: Double) async throws -> SavingsGoal {
        let request = SetSavingsGoalRequest(monthlyTarget: target)
        let response: SavingsGoalAPIResponse = try await api.request(
            endpoint: "savings-goal",
            method: "POST",
            body: request,
            requiresAuth: true
        )

        return SavingsGoal(
            id: UUID(uuidString: response.id) ?? UUID(),
            monthlyTarget: response.monthlyTarget,
            currentSavings: response.currentSavings,
            monthlyIncome: response.monthlyIncome,
            monthlyExpenses: response.monthlyExpenses,
            progress: response.progress
        )
    }

    func deleteSavingsGoal() async throws {
        let _: EmptyResponse = try await api.request(
            endpoint: "savings-goal",
            method: "DELETE",
            requiresAuth: true
        )
    }

    // MARK: - Helper
    private func mapToTransaction(_ response: TransactionAPIResponse) -> Transaction {
        let date = DateFormatters.apiDate.date(from: response.date) ?? Date()

        return Transaction(
            id: UUID(uuidString: response.id) ?? UUID(),
            title: response.title,
            amount: response.amount,
            type: TransactionType(rawValue: response.type) ?? .expense,
            category: response.category,
            date: date
        )
    }
}

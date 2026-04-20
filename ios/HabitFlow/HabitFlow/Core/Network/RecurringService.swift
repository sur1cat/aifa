import Foundation

// MARK: - API Models
struct RecurringAPIResponse: Codable, Sendable {
    let id: String
    let title: String
    let amount: Double
    let type: String
    let category: String
    let frequency: String
    let startDate: String
    let nextDate: String
    let endDate: String?
    let remainingPayments: Int?
    let isActive: Bool
    let createdAt: String

    enum CodingKeys: String, CodingKey {
        case id, title, amount, type, category, frequency
        case startDate = "start_date"
        case nextDate = "next_date"
        case endDate = "end_date"
        case remainingPayments = "remaining_payments"
        case isActive = "is_active"
        case createdAt = "created_at"
    }
}

struct CreateRecurringRequest: Encodable, Sendable {
    let title: String
    let amount: Double
    let type: String
    let category: String
    let frequency: String
    let startDate: String
    let endDate: String?
    let remainingPayments: Int?

    enum CodingKeys: String, CodingKey {
        case title, amount, type, category, frequency
        case startDate = "start_date"
        case endDate = "end_date"
        case remainingPayments = "remaining_payments"
    }
}

struct UpdateRecurringRequest: Encodable, Sendable {
    let title: String?
    let amount: Double?
    let type: String?
    let category: String?
    let frequency: String?
    let startDate: String?
    let nextDate: String?
    let endDate: String?
    let remainingPayments: Int?
    let isActive: Bool?

    enum CodingKeys: String, CodingKey {
        case title, amount, type, category, frequency
        case startDate = "start_date"
        case nextDate = "next_date"
        case endDate = "end_date"
        case remainingPayments = "remaining_payments"
        case isActive = "is_active"
    }
}

struct ProcessRecurringResponse: Codable, Sendable {
    let processed: Int
    let transactionsCreated: Int

    enum CodingKeys: String, CodingKey {
        case processed
        case transactionsCreated = "transactions_created"
    }
}

struct ProjectionResponse: Codable, Sendable {
    let monthlyIncome: Double
    let monthlyExpense: Double
    let monthlyNet: Double

    enum CodingKeys: String, CodingKey {
        case monthlyIncome = "monthly_income"
        case monthlyExpense = "monthly_expense"
        case monthlyNet = "monthly_net"
    }
}

// MARK: - RecurringService
actor RecurringService {
    static let shared = RecurringService()
    private let api = APIClient.shared

    // MARK: - List Recurring Transactions
    func getRecurringTransactions() async throws -> [RecurringTransaction] {
        let response: [RecurringAPIResponse] = try await api.request(
            endpoint: "recurring-transactions",
            requiresAuth: true
        )
        return response.map { mapToRecurring($0) }
    }

    // MARK: - Create Recurring Transaction
    func createRecurring(_ recurring: RecurringTransaction) async throws -> RecurringTransaction {
        let dateFormatter = DateFormatter()
        dateFormatter.dateFormat = "yyyy-MM-dd"

        let request = CreateRecurringRequest(
            title: recurring.title,
            amount: recurring.amount,
            type: recurring.type.rawValue,
            category: recurring.category.rawValue,
            frequency: recurring.frequency.rawValue,
            startDate: dateFormatter.string(from: recurring.startDate),
            endDate: recurring.endDate.map { dateFormatter.string(from: $0) },
            remainingPayments: recurring.remainingPayments
        )
        let response: RecurringAPIResponse = try await api.request(
            endpoint: "recurring-transactions",
            method: "POST",
            body: request,
            requiresAuth: true
        )
        return mapToRecurring(response)
    }

    // MARK: - Update Recurring Transaction
    func updateRecurring(_ recurring: RecurringTransaction) async throws -> RecurringTransaction {
        let dateFormatter = DateFormatter()
        dateFormatter.dateFormat = "yyyy-MM-dd"

        let request = UpdateRecurringRequest(
            title: recurring.title,
            amount: recurring.amount,
            type: recurring.type.rawValue,
            category: recurring.category.rawValue,
            frequency: recurring.frequency.rawValue,
            startDate: dateFormatter.string(from: recurring.startDate),
            nextDate: dateFormatter.string(from: recurring.nextDate),
            endDate: recurring.endDate.map { dateFormatter.string(from: $0) },
            remainingPayments: recurring.remainingPayments,
            isActive: recurring.isActive
        )
        let response: RecurringAPIResponse = try await api.request(
            endpoint: "recurring-transactions/\(recurring.id.uuidString)",
            method: "PUT",
            body: request,
            requiresAuth: true
        )
        return mapToRecurring(response)
    }

    // MARK: - Process Recurring Transactions
    func processRecurring() async throws -> (processed: Int, created: Int) {
        let response: ProcessRecurringResponse = try await api.request(
            endpoint: "recurring-transactions/process",
            method: "POST",
            requiresAuth: true
        )
        return (response.processed, response.transactionsCreated)
    }

    // MARK: - Delete Recurring Transaction
    func deleteRecurring(_ id: UUID) async throws {
        let _: EmptyResponse = try await api.request(
            endpoint: "recurring-transactions/\(id.uuidString)",
            method: "DELETE",
            requiresAuth: true
        )
    }

    // MARK: - Get Projection
    func getProjection() async throws -> (income: Double, expense: Double, net: Double) {
        let response: ProjectionResponse = try await api.request(
            endpoint: "recurring-transactions/projection",
            requiresAuth: true
        )
        return (response.monthlyIncome, response.monthlyExpense, response.monthlyNet)
    }

    // MARK: - Helper
    private func mapToRecurring(_ response: RecurringAPIResponse) -> RecurringTransaction {
        let dateFormatter = DateFormatter()
        dateFormatter.dateFormat = "yyyy-MM-dd"

        let startDate = dateFormatter.date(from: response.startDate) ?? Date()
        let nextDate = dateFormatter.date(from: response.nextDate) ?? Date()
        let endDate = response.endDate.flatMap { dateFormatter.date(from: $0) }

        return RecurringTransaction(
            id: UUID(uuidString: response.id) ?? UUID(),
            title: response.title,
            amount: response.amount,
            type: TransactionType(rawValue: response.type) ?? .expense,
            category: RecurringCategory(rawValue: response.category) ?? .subscriptions,
            frequency: RecurrenceFrequency(rawValue: response.frequency) ?? .monthly,
            startDate: startDate,
            nextDate: nextDate,
            endDate: endDate,
            remainingPayments: response.remainingPayments,
            isActive: response.isActive
        )
    }
}

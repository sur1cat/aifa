import Foundation

enum AIAgent: String, Codable {
    case habitCoach = "habit_coach"
    case taskAssistant = "task_assistant"
    case financeAdvisor = "finance_advisor"
    case lifeCoach = "life_coach"

    var displayName: String {
        switch self {
        case .habitCoach: return "Habit Coach"
        case .taskAssistant: return "Task Assistant"
        case .financeAdvisor: return "Finance Advisor"
        case .lifeCoach: return "Life Coach"
        }
    }

    var icon: String {
        switch self {
        case .habitCoach: return "leaf.fill"
        case .taskAssistant: return "checkmark.circle.fill"
        case .financeAdvisor: return "dollarsign.circle.fill"
        case .lifeCoach: return "person.fill"
        }
    }
}

struct AIChatRequest: Codable, Sendable {
    let agent: String
    let message: String
    let context: String?
}

struct AIChatResponse: Codable, Sendable {
    let response: String
}

struct AICommandRequest: Codable, Sendable {
    let message: String
    let context: String?
}

struct AICommandTransaction: Codable, Sendable {
    let id: String
    let title: String
    let amount: Double
    let type: String
    let category: String
    let date: String
    let createdAt: String

    enum CodingKeys: String, CodingKey {
        case id
        case title
        case amount
        case type
        case category
        case date
        case createdAt = "created_at"
    }
}

struct AICommandGoal: Codable, Sendable {
    let id: String
    let title: String
    let icon: String
    let targetValue: Int?
    let unit: String?
    let deadline: String?

    enum CodingKeys: String, CodingKey {
        case id
        case title
        case icon
        case targetValue = "target_value"
        case unit
        case deadline
    }
}

struct AICommandRecurring: Codable, Sendable {
    let id: String
    let title: String
    let amount: Double
    let type: String
    let category: String
    let frequency: String
    let startDate: String

    enum CodingKeys: String, CodingKey {
        case id
        case title
        case amount
        case type
        case category
        case frequency
        case startDate = "start_date"
    }
}

struct AICommandResponse: Codable, Sendable {
    let status: String
    let intent: String
    let message: String
    let missingFields: [String]?
    let transaction: AICommandTransaction?
    let transactions: [AICommandTransaction]?
    let goal: AICommandGoal?
    let recurring: AICommandRecurring?

    enum CodingKeys: String, CodingKey {
        case status
        case intent
        case message
        case missingFields = "missing_fields"
        case transaction
        case transactions
        case goal
        case recurring
    }
}

// MARK: - Goal to Habits Conversion

struct GoalToHabitsRequest: Codable, Sendable {
    let goalTitle: String
    let goalDeadline: String?
    let targetValue: String?
    let context: String?
}

struct SuggestedHabit: Codable, Sendable, Identifiable {
    let title: String
    let icon: String
    let color: String
    let period: String
    let reason: String

    var id: String { title }
}

struct GoalToHabitsResponse: Codable, Sendable {
    let habits: [SuggestedHabit]
    let explanation: String
}

// MARK: - Goal Clarify Questions

struct GoalClarifyRequest: Codable, Sendable {
    let goalTitle: String
}

struct ClarifyQuestion: Codable, Sendable, Identifiable {
    let id: String
    let question: String
    let placeholder: String
    let type: String
}

struct GoalClarifyResponse: Codable, Sendable {
    let questions: [ClarifyQuestion]
    let contextHint: String

    enum CodingKeys: String, CodingKey {
        case questions
        case contextHint = "context_hint"
    }
}

actor AIService {
    private let client: APIClient

    init(client: APIClient = .shared) {
        self.client = client
    }

    func chat(agent: AIAgent, message: String, context: String? = nil) async throws -> String {
        if agent == .financeAdvisor {
            let commandResponse = try await command(message: message, context: context)
            if commandResponse.status == "unsupported" {
                return try await chatMessage(agent: agent, message: message, context: context)
            }
            return formatCommandResponse(commandResponse)
        }

        return try await chatMessage(agent: agent, message: message, context: context)
    }

    private func chatMessage(agent: AIAgent, message: String, context: String? = nil) async throws -> String {

        let request = AIChatRequest(
            agent: agent.rawValue,
            message: message,
            context: context
        )

        let response: AIChatResponse = try await client.request(
            endpoint: "ai/chat",
            method: "POST",
            body: request,
            requiresAuth: true
        )

        return response.response
    }

    func command(message: String, context: String? = nil) async throws -> AICommandResponse {
        let request = AICommandRequest(message: message, context: context)

        let response: AICommandResponse = try await client.request(
            endpoint: "ai/command",
            method: "POST",
            body: request,
            requiresAuth: true
        )

        return response
    }

    private func formatCommandResponse(_ response: AICommandResponse) -> String {
        switch response.status {
        case "completed":
            if let transactions = response.transactions, !transactions.isEmpty {
                let lines = transactions.map { "\($0.title) • \($0.amount) • \($0.date)" }
                return "\(response.message)\n\n" + lines.joined(separator: "\n")
            }
            if let transaction = response.transaction {
                return "\(response.message)\n\n\(transaction.title) • \(transaction.amount) • \(transaction.date)"
            }
            if let goal = response.goal {
                return "\(response.message)\n\n\(goal.icon) \(goal.title)"
            }
            if let recurring = response.recurring {
                return "\(response.message)\n\n\(recurring.title) • \(recurring.amount) • \(recurring.frequency)"
            }
            return response.message
        case "needs_clarification":
            return response.message
        case "unsupported":
            return "This request is not supported yet."
        default:
            return response.message
        }
    }

    /// Converts an outcome goal into process habits using AI
    func generateHabitsFromGoal(
        title: String,
        deadline: String? = nil,
        targetValue: String? = nil,
        context: String? = nil
    ) async throws -> GoalToHabitsResponse {
        let request = GoalToHabitsRequest(
            goalTitle: title,
            goalDeadline: deadline,
            targetValue: targetValue,
            context: context
        )

        let response: GoalToHabitsResponse = try await client.request(
            endpoint: "ai/goal-to-habits",
            method: "POST",
            body: request,
            requiresAuth: true
        )

        return response
    }

    /// Generates clarifying questions for a goal to get better context
    func generateGoalQuestions(title: String) async throws -> GoalClarifyResponse {
        let request = GoalClarifyRequest(goalTitle: title)

        let response: GoalClarifyResponse = try await client.request(
            endpoint: "ai/goal-clarify",
            method: "POST",
            body: request,
            requiresAuth: true
        )

        return response
    }
}

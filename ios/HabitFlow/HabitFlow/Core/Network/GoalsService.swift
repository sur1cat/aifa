import Foundation

// MARK: - API Models
struct GoalAPIResponse: Codable, Sendable {
    let id: String
    let title: String
    let icon: String
    let targetValue: Int?
    let unit: String?
    let deadline: String?
    let createdAt: String
    let archivedAt: String?

    enum CodingKeys: String, CodingKey {
        case id, title, icon, unit, deadline
        case targetValue = "target_value"
        case createdAt = "created_at"
        case archivedAt = "archived_at"
    }
}

struct CreateGoalRequest: Encodable, Sendable {
    let title: String
    let icon: String
    let targetValue: Int?
    let unit: String?
    let deadline: String?

    enum CodingKeys: String, CodingKey {
        case title, icon, unit, deadline
        case targetValue = "target_value"
    }
}

struct UpdateGoalRequest: Encodable, Sendable {
    let title: String?
    let icon: String?
    let targetValue: Int?
    let unit: String?
    let deadline: String?
    let archivedAt: String?

    enum CodingKeys: String, CodingKey {
        case title, icon, unit, deadline
        case targetValue = "target_value"
        case archivedAt = "archived_at"
    }
}

// MARK: - GoalsService
actor GoalsService {
    static let shared = GoalsService()
    private let api = APIClient.shared

    // MARK: - List Goals
    func getGoals() async throws -> [Goal] {
        let response: [GoalAPIResponse] = try await api.request(
            endpoint: "goals",
            requiresAuth: true
        )
        return response.map { mapToGoal($0) }
    }

    // MARK: - Create Goal
    func createGoal(_ goal: Goal) async throws -> Goal {
        let dateFormatter = ISO8601DateFormatter()
        let deadlineString: String? = goal.deadline.map { dateFormatter.string(from: $0) }

        let request = CreateGoalRequest(
            title: goal.title,
            icon: goal.icon,
            targetValue: goal.targetValue,
            unit: goal.unit,
            deadline: deadlineString
        )
        let response: GoalAPIResponse = try await api.request(
            endpoint: "goals",
            method: "POST",
            body: request,
            requiresAuth: true
        )
        return mapToGoal(response)
    }

    // MARK: - Update Goal
    func updateGoal(_ goal: Goal) async throws -> Goal {
        let dateFormatter = ISO8601DateFormatter()
        let deadlineString: String? = goal.deadline.map { dateFormatter.string(from: $0) }
        // Send "" to clear archived_at, or the date string to set it
        let archivedAtString: String? = goal.archivedAt.map { dateFormatter.string(from: $0) } ?? ""

        let request = UpdateGoalRequest(
            title: goal.title,
            icon: goal.icon,
            targetValue: goal.targetValue,
            unit: goal.unit,
            deadline: deadlineString,
            archivedAt: archivedAtString
        )
        let response: GoalAPIResponse = try await api.request(
            endpoint: "goals/\(goal.id.uuidString)",
            method: "PUT",
            body: request,
            requiresAuth: true
        )
        return mapToGoal(response)
    }

    // MARK: - Delete Goal
    func deleteGoal(_ id: UUID) async throws {
        let _: EmptyResponse = try await api.request(
            endpoint: "goals/\(id.uuidString)",
            method: "DELETE",
            requiresAuth: true
        )
    }

    // MARK: - Helper
    private func mapToGoal(_ response: GoalAPIResponse) -> Goal {
        let dateFormatter = DateFormatter()
        dateFormatter.dateFormat = "yyyy-MM-dd'T'HH:mm:ssZ"
        let createdAt = dateFormatter.date(from: response.createdAt) ?? Date()

        var deadline: Date? = nil
        if let deadlineString = response.deadline {
            deadline = dateFormatter.date(from: deadlineString)
        }

        var archivedAt: Date? = nil
        if let archivedAtString = response.archivedAt {
            archivedAt = dateFormatter.date(from: archivedAtString)
        }

        return Goal(
            id: UUID(uuidString: response.id) ?? UUID(),
            title: response.title,
            icon: response.icon,
            targetValue: response.targetValue,
            unit: response.unit,
            deadline: deadline,
            createdAt: createdAt,
            archivedAt: archivedAt
        )
    }
}

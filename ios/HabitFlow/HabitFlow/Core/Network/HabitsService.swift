import Foundation

// MARK: - API Models
struct HabitAPIResponse: Codable, Sendable {
    let id: String
    let goalId: String?
    let title: String
    let icon: String
    let color: String
    let period: String
    let completedDates: [String]
    let targetValue: Int?
    let unit: String?
    let progressValues: [String: Int]?
    let streak: Int
    let createdAt: String
    let archivedAt: String?

    enum CodingKeys: String, CodingKey {
        case id, title, icon, color, period, unit, streak
        case goalId = "goal_id"
        case completedDates = "completed_dates"
        case targetValue = "target_value"
        case progressValues = "progress_values"
        case createdAt = "created_at"
        case archivedAt = "archived_at"
    }
}

struct CreateHabitRequest: Encodable, Sendable {
    let goalId: String?
    let title: String
    let icon: String
    let color: String
    let period: String
    let targetValue: Int?
    let unit: String?

    enum CodingKeys: String, CodingKey {
        case title, icon, color, period, unit
        case goalId = "goal_id"
        case targetValue = "target_value"
    }
}

struct UpdateHabitRequest: Encodable, Sendable {
    let goalId: String?
    let title: String?
    let icon: String?
    let color: String?
    let period: String?
    let targetValue: Int?
    let unit: String?
    let archivedAt: String?

    enum CodingKeys: String, CodingKey {
        case title, icon, color, period, unit
        case goalId = "goal_id"
        case targetValue = "target_value"
        case archivedAt = "archived_at"
    }
}

struct ToggleCompletionRequest: Encodable, Sendable {
    let date: String
    let value: Int?
}

// MARK: - HabitsService
actor HabitsService {
    static let shared = HabitsService()
    private let api = APIClient.shared

    // MARK: - List Habits
    func getHabits() async throws -> [Habit] {
        let response: [HabitAPIResponse] = try await api.request(
            endpoint: "habits",
            requiresAuth: true
        )
        return response.map { mapToHabit($0) }
    }

    // MARK: - Create Habit
    func createHabit(_ habit: Habit) async throws -> Habit {
        let request = CreateHabitRequest(
            goalId: habit.goalId?.uuidString,
            title: habit.title,
            icon: habit.icon,
            color: habit.color,
            period: habit.period.rawValue,
            targetValue: habit.targetValue,
            unit: habit.unit
        )
        let response: HabitAPIResponse = try await api.request(
            endpoint: "habits",
            method: "POST",
            body: request,
            requiresAuth: true
        )
        return mapToHabit(response)
    }

    // MARK: - Update Habit
    func updateHabit(_ habit: Habit) async throws -> Habit {
        let dateFormatter = ISO8601DateFormatter()
        // Send "" to clear archived_at, or the date string to set it
        let archivedAtString: String? = habit.archivedAt.map { dateFormatter.string(from: $0) } ?? ""

        let request = UpdateHabitRequest(
            goalId: habit.goalId?.uuidString,
            title: habit.title,
            icon: habit.icon,
            color: habit.color,
            period: habit.period.rawValue,
            targetValue: habit.targetValue,
            unit: habit.unit,
            archivedAt: archivedAtString
        )
        let response: HabitAPIResponse = try await api.request(
            endpoint: "habits/\(habit.id.uuidString)",
            method: "PUT",
            body: request,
            requiresAuth: true
        )
        return mapToHabit(response)
    }

    // MARK: - Delete Habit
    func deleteHabit(_ id: UUID) async throws {
        let _: EmptyResponse = try await api.request(
            endpoint: "habits/\(id.uuidString)",
            method: "DELETE",
            requiresAuth: true
        )
    }

    // MARK: - Toggle Completion
    func toggleCompletion(habitID: UUID, date: String, value: Int? = nil) async throws -> Habit {
        let request = ToggleCompletionRequest(date: date, value: value)
        let response: HabitAPIResponse = try await api.request(
            endpoint: "habits/\(habitID.uuidString)/toggle",
            method: "POST",
            body: request,
            requiresAuth: true
        )
        return mapToHabit(response)
    }

    // MARK: - Set Progress
    func setProgress(habitID: UUID, date: String, value: Int) async throws -> Habit {
        return try await toggleCompletion(habitID: habitID, date: date, value: value)
    }

    // MARK: - Helper
    private func mapToHabit(_ response: HabitAPIResponse) -> Habit {
        let dateFormatter = DateFormatter()
        dateFormatter.dateFormat = "yyyy-MM-dd'T'HH:mm:ssZ"
        let createdAt = dateFormatter.date(from: response.createdAt) ?? Date()

        var archivedAt: Date? = nil
        if let archivedAtString = response.archivedAt {
            archivedAt = dateFormatter.date(from: archivedAtString)
        }

        var goalId: UUID? = nil
        if let goalIdString = response.goalId {
            goalId = UUID(uuidString: goalIdString)
        }

        return Habit(
            id: UUID(uuidString: response.id) ?? UUID(),
            goalId: goalId,
            title: response.title,
            icon: response.icon,
            color: response.color,
            period: HabitPeriod(rawValue: response.period) ?? .daily,
            completedDates: response.completedDates,
            createdAt: createdAt,
            archivedAt: archivedAt,
            targetValue: response.targetValue,
            unit: response.unit,
            progressValues: response.progressValues ?? [:],
            streak: response.streak
        )
    }
}

// Empty response for delete
struct EmptyResponse: Codable, Sendable {
    let message: String?
}

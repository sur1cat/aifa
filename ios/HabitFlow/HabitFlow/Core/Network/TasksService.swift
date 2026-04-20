import Foundation

// MARK: - API Models
struct TaskAPIResponse: Codable, Sendable {
    let id: String
    let title: String
    let isCompleted: Bool
    let priority: String
    let dueDate: String
    let createdAt: String

    enum CodingKeys: String, CodingKey {
        case id, title, priority
        case isCompleted = "is_completed"
        case dueDate = "due_date"
        case createdAt = "created_at"
    }
}

struct CreateTaskRequest: Encodable, Sendable {
    let title: String
    let priority: String
    let dueDate: String

    enum CodingKeys: String, CodingKey {
        case title, priority
        case dueDate = "due_date"
    }
}

struct UpdateTaskRequest: Encodable, Sendable {
    let title: String?
    let isCompleted: Bool?
    let priority: String?
    let dueDate: String?

    enum CodingKeys: String, CodingKey {
        case title, priority
        case isCompleted = "is_completed"
        case dueDate = "due_date"
    }
}

// MARK: - TasksService
actor TasksService {
    static let shared = TasksService()
    private let api = APIClient.shared

    // MARK: - List Tasks
    func getTasks(date: String? = nil) async throws -> [DailyTask] {
        var endpoint = "tasks"
        if let date = date {
            endpoint += "?date=\(date)"
        }
        let response: [TaskAPIResponse] = try await api.request(
            endpoint: endpoint,
            requiresAuth: true
        )
        return response.map { mapToTask($0) }
    }

    // MARK: - Create Task
    func createTask(_ task: DailyTask) async throws -> DailyTask {
        let request = CreateTaskRequest(
            title: task.title,
            priority: task.priority.rawValue,
            dueDate: DateFormatters.apiDate.string(from: task.dueDate)
        )
        let response: TaskAPIResponse = try await api.request(
            endpoint: "tasks",
            method: "POST",
            body: request,
            requiresAuth: true
        )
        return mapToTask(response)
    }

    // MARK: - Update Task
    func updateTask(_ task: DailyTask) async throws -> DailyTask {
        let request = UpdateTaskRequest(
            title: task.title,
            isCompleted: task.isCompleted,
            priority: task.priority.rawValue,
            dueDate: DateFormatters.apiDate.string(from: task.dueDate)
        )
        let response: TaskAPIResponse = try await api.request(
            endpoint: "tasks/\(task.id.uuidString)",
            method: "PUT",
            body: request,
            requiresAuth: true
        )
        return mapToTask(response)
    }

    // MARK: - Delete Task
    func deleteTask(_ id: UUID) async throws {
        let _: EmptyResponse = try await api.request(
            endpoint: "tasks/\(id.uuidString)",
            method: "DELETE",
            requiresAuth: true
        )
    }

    // MARK: - Toggle Task
    func toggleTask(_ id: UUID) async throws -> DailyTask {
        let response: TaskAPIResponse = try await api.request(
            endpoint: "tasks/\(id.uuidString)/toggle",
            method: "POST",
            requiresAuth: true
        )
        return mapToTask(response)
    }

    // MARK: - Helper
    private func mapToTask(_ response: TaskAPIResponse) -> DailyTask {
        let dueDate = DateFormatters.apiDate.date(from: response.dueDate) ?? Date()
        let createdAt = DateFormatters.iso8601Basic.date(from: response.createdAt) ?? Date()

        return DailyTask(
            id: UUID(uuidString: response.id) ?? UUID(),
            title: response.title,
            isCompleted: response.isCompleted,
            priority: TaskPriority(rawValue: response.priority) ?? .medium,
            dueDate: dueDate,
            createdAt: createdAt
        )
    }
}

import AppIntents
import SwiftUI

// MARK: - Add Task Intent

struct AddTaskIntent: AppIntent {
    static var title: LocalizedStringResource = "Add Task"
    static var description = IntentDescription("Add a new task to your list")
    static var openAppWhenRun: Bool = false

    @Parameter(title: "Task Name")
    var taskName: String

    @Parameter(title: "Priority", default: .medium)
    var priority: TaskPriorityEnum

    @MainActor
    func perform() async throws -> some IntentResult & ProvidesDialog {
        guard !taskName.trimmingCharacters(in: .whitespaces).isEmpty else {
            return .result(dialog: "Please provide a task name")
        }

        let dataManager = DataManager.shared
        let taskPriority: TaskPriority = switch priority {
        case .low: .low
        case .medium: .medium
        case .high: .high
        }

        let task = DailyTask(title: taskName, priority: taskPriority, dueDate: Date())
        dataManager.addTask(task)

        return .result(dialog: "Added '\(taskName)' to your tasks!")
    }

    static var parameterSummary: some ParameterSummary {
        Summary("Add task \(\.$taskName) with \(\.$priority) priority")
    }
}

// MARK: - Complete Task Intent

struct CompleteTaskIntent: AppIntent {
    static var title: LocalizedStringResource = "Complete Task"
    static var description = IntentDescription("Mark a task as complete")
    static var openAppWhenRun: Bool = false

    @Parameter(title: "Task")
    var task: TaskEntity?

    @MainActor
    func perform() async throws -> some IntentResult & ProvidesDialog {
        guard let task = task else {
            return .result(dialog: "Please select a task")
        }

        let dataManager = DataManager.shared
        if let taskObject = dataManager.tasks.first(where: { $0.id.uuidString == task.id }) {
            if taskObject.isCompleted {
                return .result(dialog: "'\(task.title)' is already complete!")
            }
            dataManager.toggleTask(taskObject)
            return .result(dialog: "Completed '\(task.title)'! Great job!")
        }

        return .result(dialog: "Task not found")
    }

    static var parameterSummary: some ParameterSummary {
        Summary("Complete \(\.$task)")
    }
}

// MARK: - List Tasks Intent

struct ListTasksIntent: AppIntent {
    static var title: LocalizedStringResource = "List Today's Tasks"
    static var description = IntentDescription("Show your tasks for today")
    static var openAppWhenRun: Bool = false

    @MainActor
    func perform() async throws -> some IntentResult & ProvidesDialog {
        let dataManager = DataManager.shared
        let tasks = dataManager.tasks

        if tasks.isEmpty {
            return .result(dialog: "You don't have any tasks for today. Open Atoma to add one!")
        }

        let completed = tasks.filter { $0.isCompleted }.count
        let total = tasks.count

        var taskList = tasks.prefix(5).map { task in
            let status = task.isCompleted ? "✅" : "⭕"
            let priority = task.priority == .high ? "🔴" : (task.priority == .medium ? "🟡" : "🟢")
            return "\(status) \(priority) \(task.title)"
        }.joined(separator: "\n")

        if tasks.count > 5 {
            taskList += "\n... and \(tasks.count - 5) more"
        }

        return .result(dialog: "Tasks today: \(completed)/\(total) complete\n\n\(taskList)")
    }
}

// MARK: - Task Priority Enum

enum TaskPriorityEnum: String, AppEnum {
    case low
    case medium
    case high

    static var typeDisplayRepresentation: TypeDisplayRepresentation = "Priority"

    static var caseDisplayRepresentations: [TaskPriorityEnum: DisplayRepresentation] = [
        .low: DisplayRepresentation(title: "Low", subtitle: "Can wait"),
        .medium: DisplayRepresentation(title: "Medium", subtitle: "Normal priority"),
        .high: DisplayRepresentation(title: "High", subtitle: "Urgent")
    ]
}

// MARK: - Task Entity

struct TaskEntity: AppEntity {
    var id: String
    var title: String
    var isCompleted: Bool
    var priority: String

    static var typeDisplayRepresentation: TypeDisplayRepresentation = "Task"
    static var defaultQuery = TaskEntityQuery()

    var displayRepresentation: DisplayRepresentation {
        let priorityIcon = priority == "high" ? "🔴" : (priority == "medium" ? "🟡" : "🟢")
        return DisplayRepresentation(
            title: "\(title)",
            subtitle: isCompleted ? "✅ Complete" : "\(priorityIcon) \(priority.capitalized)"
        )
    }
}

struct TaskEntityQuery: EntityQuery {
    @MainActor
    func entities(for identifiers: [String]) async throws -> [TaskEntity] {
        let dataManager = DataManager.shared
        return dataManager.tasks
            .filter { identifiers.contains($0.id.uuidString) }
            .map { task in
                TaskEntity(
                    id: task.id.uuidString,
                    title: task.title,
                    isCompleted: task.isCompleted,
                    priority: task.priority.rawValue
                )
            }
    }

    @MainActor
    func suggestedEntities() async throws -> [TaskEntity] {
        let dataManager = DataManager.shared
        return dataManager.tasks
            .filter { !$0.isCompleted }
            .map { task in
                TaskEntity(
                    id: task.id.uuidString,
                    title: task.title,
                    isCompleted: task.isCompleted,
                    priority: task.priority.rawValue
                )
            }
    }
}

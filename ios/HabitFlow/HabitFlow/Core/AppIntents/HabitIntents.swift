import AppIntents
import SwiftUI

// MARK: - Toggle Habit Intent

struct ToggleHabitIntent: AppIntent {
    static var title: LocalizedStringResource = "Toggle Habit"
    static var description = IntentDescription("Mark a habit as complete or incomplete for today")
    static var openAppWhenRun: Bool = false

    @Parameter(title: "Habit")
    var habit: HabitEntity?

    @MainActor
    func perform() async throws -> some IntentResult & ProvidesDialog {
        guard let habit = habit else {
            return .result(dialog: "Please select a habit")
        }

        // Toggle the habit
        let dataManager = DataManager.shared
        if let habitObject = dataManager.habits.first(where: { $0.id.uuidString == habit.id }) {
            let wasCompleted = habitObject.isCompletedInCurrentPeriod
            dataManager.toggleHabit(habitObject)

            let message = !wasCompleted
                ? "Marked '\(habit.title)' as complete!"
                : "Marked '\(habit.title)' as incomplete"

            return .result(dialog: "\(message)")
        }

        return .result(dialog: "Habit not found")
    }

    static var parameterSummary: some ParameterSummary {
        Summary("Toggle \(\.$habit)")
    }
}

// MARK: - List Habits Intent

struct ListHabitsIntent: AppIntent {
    static var title: LocalizedStringResource = "List Today's Habits"
    static var description = IntentDescription("Show your habits and their completion status for today")
    static var openAppWhenRun: Bool = false

    @MainActor
    func perform() async throws -> some IntentResult & ProvidesDialog {
        let dataManager = DataManager.shared
        let habits = dataManager.habits

        if habits.isEmpty {
            return .result(dialog: "You don't have any habits yet. Open Atoma to create one!")
        }

        let completed = habits.filter { $0.isCompletedInCurrentPeriod }.count
        let total = habits.count

        var habitList = habits.prefix(5).map { habit in
            let status = habit.isCompletedInCurrentPeriod ? "✅" : "⭕"
            return "\(status) \(habit.title)"
        }.joined(separator: "\n")

        if habits.count > 5 {
            habitList += "\n... and \(habits.count - 5) more"
        }

        return .result(dialog: "Habits today: \(completed)/\(total) complete\n\n\(habitList)")
    }
}

// MARK: - Check Habit Streak Intent

struct CheckHabitStreakIntent: AppIntent {
    static var title: LocalizedStringResource = "Check Habit Streak"
    static var description = IntentDescription("Check the current streak for a habit")
    static var openAppWhenRun: Bool = false

    @Parameter(title: "Habit")
    var habit: HabitEntity?

    @MainActor
    func perform() async throws -> some IntentResult & ProvidesDialog {
        guard let habit = habit else {
            return .result(dialog: "Please select a habit")
        }

        let dataManager = DataManager.shared
        if let habitObject = dataManager.habits.first(where: { $0.id.uuidString == habit.id }) {
            let streak = habitObject.streak
            if streak == 0 {
                return .result(dialog: "'\(habit.title)' has no active streak. Complete it today to start one!")
            } else if streak == 1 {
                return .result(dialog: "'\(habit.title)' has a 1 day streak. Keep it going!")
            } else {
                return .result(dialog: "'\(habit.title)' has a \(streak) day streak! Great job!")
            }
        }

        return .result(dialog: "Habit not found")
    }

    static var parameterSummary: some ParameterSummary {
        Summary("Check streak for \(\.$habit)")
    }
}

// MARK: - Habit Entity

struct HabitEntity: AppEntity {
    var id: String
    var title: String
    var icon: String
    var isCompleted: Bool
    var streak: Int

    static var typeDisplayRepresentation: TypeDisplayRepresentation = "Habit"
    static var defaultQuery = HabitEntityQuery()

    var displayRepresentation: DisplayRepresentation {
        DisplayRepresentation(
            title: "\(title)",
            subtitle: isCompleted ? "✅ Complete" : "⭕ Incomplete",
            image: .init(systemName: icon)
        )
    }
}

struct HabitEntityQuery: EntityQuery {
    @MainActor
    func entities(for identifiers: [String]) async throws -> [HabitEntity] {
        let dataManager = DataManager.shared
        return dataManager.habits
            .filter { identifiers.contains($0.id.uuidString) }
            .map { habit in
                HabitEntity(
                    id: habit.id.uuidString,
                    title: habit.title,
                    icon: habit.icon,
                    isCompleted: habit.isCompletedInCurrentPeriod,
                    streak: habit.streak
                )
            }
    }

    @MainActor
    func suggestedEntities() async throws -> [HabitEntity] {
        let dataManager = DataManager.shared
        return dataManager.habits.map { habit in
            HabitEntity(
                id: habit.id.uuidString,
                title: habit.title,
                icon: habit.icon,
                isCompleted: habit.isCompletedInCurrentPeriod,
                streak: habit.streak
            )
        }
    }
}

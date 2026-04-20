import AppIntents

struct AtomaShortcuts: AppShortcutsProvider {
    static var appShortcuts: [AppShortcut] {
        // Life Score
        AppShortcut(
            intent: CheckLifeScoreIntent(),
            phrases: [
                "Check my \(.applicationName) score",
                "What's my \(.applicationName) life score",
                "How am I doing in \(.applicationName)"
            ],
            shortTitle: "Life Score",
            systemImageName: "chart.line.uptrend.xyaxis"
        )

        // Habits
        AppShortcut(
            intent: ListHabitsIntent(),
            phrases: [
                "Show my \(.applicationName) habits",
                "List my habits in \(.applicationName)",
                "Check \(.applicationName) habits"
            ],
            shortTitle: "Today's Habits",
            systemImageName: "repeat"
        )

        AppShortcut(
            intent: ToggleHabitIntent(),
            phrases: [
                "Complete habit in \(.applicationName)",
                "Mark habit done in \(.applicationName)",
                "Toggle \(.applicationName) habit"
            ],
            shortTitle: "Toggle Habit",
            systemImageName: "checkmark.circle"
        )

        // Tasks
        AppShortcut(
            intent: ListTasksIntent(),
            phrases: [
                "Show my \(.applicationName) tasks",
                "List my tasks in \(.applicationName)",
                "Check \(.applicationName) tasks"
            ],
            shortTitle: "Today's Tasks",
            systemImageName: "checklist"
        )

        AppShortcut(
            intent: AddTaskIntent(),
            phrases: [
                "Add task in \(.applicationName)",
                "Create task in \(.applicationName)",
                "New \(.applicationName) task"
            ],
            shortTitle: "Add Task",
            systemImageName: "plus.circle"
        )

        // Budget
        AppShortcut(
            intent: CheckBalanceIntent(),
            phrases: [
                "Check my \(.applicationName) balance",
                "What's my balance in \(.applicationName)",
                "Show \(.applicationName) budget"
            ],
            shortTitle: "Check Balance",
            systemImageName: "creditcard"
        )

        AppShortcut(
            intent: AddExpenseIntent(),
            phrases: [
                "Add expense in \(.applicationName)",
                "Log expense in \(.applicationName)",
                "Record \(.applicationName) expense"
            ],
            shortTitle: "Add Expense",
            systemImageName: "minus.circle"
        )

        AppShortcut(
            intent: AddIncomeIntent(),
            phrases: [
                "Add income in \(.applicationName)",
                "Log income in \(.applicationName)",
                "Record \(.applicationName) income"
            ],
            shortTitle: "Add Income",
            systemImageName: "plus.circle"
        )
    }
}

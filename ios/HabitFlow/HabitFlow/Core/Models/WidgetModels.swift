import Foundation

// Models shared between main app and widget extension
// These must match the models in AtomaWidgets/SharedData.swift

struct WidgetHabitData: Codable, Identifiable {
    let id: UUID
    let title: String
    let icon: String
    let color: String
    let isCompleted: Bool
    let streak: Int
}

struct WidgetTaskData: Codable, Identifiable {
    let id: UUID
    let title: String
    let isCompleted: Bool
    let priority: String
}

struct WidgetBudgetData: Codable {
    let balance: Double
    let income: Double
    let expenses: Double
    let currency: String
    let currencySymbol: String
}

struct WidgetDataPayload: Codable {
    let habits: [WidgetHabitData]
    let tasks: [WidgetTaskData]
    let budget: WidgetBudgetData
    let lastUpdated: Date
}

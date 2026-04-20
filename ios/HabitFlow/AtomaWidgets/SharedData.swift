import Foundation

// App Group identifier for sharing data between main app and widgets
let appGroupIdentifier = "group.com.azamatbigali.habitflow"

struct WidgetHabit: Codable, Identifiable {
    let id: UUID
    let title: String
    let icon: String
    let color: String
    let isCompleted: Bool
    let streak: Int
}

struct WidgetTask: Codable, Identifiable {
    let id: UUID
    let title: String
    let isCompleted: Bool
    let priority: String
}

struct WidgetBudget: Codable {
    let balance: Double
    let income: Double
    let expenses: Double
    let currency: String
    let currencySymbol: String
}

struct WidgetData: Codable {
    let habits: [WidgetHabit]
    let tasks: [WidgetTask]
    let budget: WidgetBudget
    let lastUpdated: Date

    static var empty: WidgetData {
        WidgetData(
            habits: [],
            tasks: [],
            budget: WidgetBudget(balance: 0, income: 0, expenses: 0, currency: "USD", currencySymbol: "$"),
            lastUpdated: Date()
        )
    }
}

class WidgetDataManager {
    static let shared = WidgetDataManager()

    private let userDefaults: UserDefaults?

    private init() {
        userDefaults = UserDefaults(suiteName: appGroupIdentifier)
    }

    func saveData(_ data: WidgetData) {
        guard let encoded = try? JSONEncoder().encode(data) else { return }
        userDefaults?.set(encoded, forKey: "widgetData")
    }

    func loadData() -> WidgetData {
        guard let data = userDefaults?.data(forKey: "widgetData"),
              let decoded = try? JSONDecoder().decode(WidgetData.self, from: data) else {
            return .empty
        }
        return decoded
    }
}

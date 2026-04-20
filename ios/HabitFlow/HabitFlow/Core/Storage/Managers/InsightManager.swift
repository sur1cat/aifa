import Foundation
import Combine
import CryptoKit
import os

@MainActor
class InsightManager: ObservableObject {
    @Published var insights: [Insight] = []
    @Published var status: InsightStatus = InsightStatus()

    private let insightsKey = "insights"
    private let statusKey = "insight_status"
    private let lastInsightDateKey = "last_insight_date"

    private let service = InsightService.shared

    weak var coordinator: DataManager?

    // MARK: - Load & Save
    func load() {
        if let data = UserDefaults.standard.data(forKey: statusKey),
           let decoded = try? JSONDecoder().decode(InsightStatus.self, from: data) {
            status = decoded
        }

        if let data = UserDefaults.standard.data(forKey: insightsKey),
           let decoded = try? JSONDecoder().decode([Insight].self, from: data) {
            insights = decoded.filter { !$0.isDismissed }
        }
    }

    func saveStatus() {
        if let data = try? JSONEncoder().encode(status) {
            UserDefaults.standard.set(data, forKey: statusKey)
        }
    }

    func saveInsights() {
        if let data = try? JSONEncoder().encode(insights) {
            UserDefaults.standard.set(data, forKey: insightsKey)
        }
    }

    // MARK: - Unlock Tracking
    func checkUnlocks() {
        var updated = false

        for section in InsightSection.allCases {
            if status.isUnlocked(for: section) { continue }

            if status.daysRemaining(for: section) == 0 {
                status.setUnlocked(for: section)
                updated = true
            }
        }

        if updated {
            saveStatus()
        }
    }

    func trackFirstDate(for section: InsightSection) {
        status.setFirstDate(for: section, date: Date())
        saveStatus()
        checkUnlocks()
    }

    func markCelebrated(for section: InsightSection) {
        status.setCelebrated(for: section)
        saveStatus()
    }

    // MARK: - Data Checks
    func hasEnoughData(for section: InsightSection) -> Bool {
        guard let coordinator = coordinator else { return false }

        switch section {
        case .habits:
            let uniqueDays = Set(coordinator.habits.flatMap { $0.completedDates }).count
            return uniqueDays >= 14

        case .tasks:
            let uniqueTaskDays = Set(coordinator.tasksManager.items.map { DateFormatters.apiDate.string(from: $0.dueDate) }).count
            let completedTasks = coordinator.tasksManager.items.filter { $0.isCompleted }.count
            return uniqueTaskDays >= 7 && completedTasks >= 10

        case .budget:
            let transactions = coordinator.budgetManager.transactions
            guard let firstTransaction = transactions.min(by: { $0.date < $1.date }) else {
                return false
            }

            let calendar = Calendar.current
            let daysSinceFirst = calendar.dateComponents([.day], from: firstTransaction.date, to: Date()).day ?? 0

            let hasIncome = transactions.contains { $0.type == .income }
            let hasExpenses = transactions.contains { $0.type == .expense }

            return daysSinceFirst >= 14 && transactions.count >= 15 && hasIncome && hasExpenses
        }
    }

    // MARK: - Generation
    private func wasGeneratedToday(for section: InsightSection) -> Bool {
        let key = "\(lastInsightDateKey)_\(section.rawValue)"
        guard let lastDate = UserDefaults.standard.object(forKey: key) as? Date else {
            return false
        }
        return Calendar.current.isDateInToday(lastDate)
    }

    private func markGenerated(for section: InsightSection) {
        let key = "\(lastInsightDateKey)_\(section.rawValue)"
        UserDefaults.standard.set(Date(), forKey: key)
    }

    private func dataSnapshotHash(for section: InsightSection) -> String {
        guard let coordinator = coordinator else { return "" }

        let input: String
        switch section {
        case .habits:
            let habits = coordinator.habits
            let streaks = habits.map { "\($0.id):\($0.streak)" }.sorted().joined()
            let completions = habits.flatMap { $0.completedDates }.sorted().suffix(30).joined()
            input = "\(habits.count)-\(streaks)-\(completions)"
        case .tasks:
            let tasks = coordinator.tasksManager.items
            let completed = tasks.filter { $0.isCompleted }.count
            let total = tasks.count
            let priorities = tasks.map { $0.priority.rawValue }.sorted().joined()
            input = "\(total)-\(completed)-\(priorities)"
        case .budget:
            let transactions = coordinator.budgetManager.transactions
            let totalIncome = transactions.filter { $0.type == .income }.reduce(0) { $0 + $1.amount }
            let totalExpense = transactions.filter { $0.type == .expense }.reduce(0) { $0 + $1.amount }
            let categories = Set(transactions.map { $0.category }).sorted().joined()
            input = "\(transactions.count)-\(Int(totalIncome))-\(Int(totalExpense))-\(categories)"
        }
        let data = Data(input.utf8)
        let hash = SHA256.hash(data: data)
        return hash.compactMap { String(format: "%02x", $0) }.joined().prefix(16).description
    }

    private func hasSignificantDataChange(for section: InsightSection) -> Bool {
        let key = "insight_data_hash_\(section.rawValue)"
        let currentHash = dataSnapshotHash(for: section)
        let lastHash = UserDefaults.standard.string(forKey: key)
        return lastHash != currentHash
    }

    private func saveDataHash(for section: InsightSection) {
        let key = "insight_data_hash_\(section.rawValue)"
        let currentHash = dataSnapshotHash(for: section)
        UserDefaults.standard.set(currentHash, forKey: key)
    }

    func generate(for section: InsightSection) {
        guard hasEnoughData(for: section) else { return }

        let needsRegeneration = !wasGeneratedToday(for: section) || hasSignificantDataChange(for: section)
        guard needsRegeneration else { return }

        Task { [weak self] in
            guard let self = self, let coordinator = self.coordinator else { return }

            do {
                var newInsights: [Insight] = []

                switch section {
                case .habits:
                    newInsights = try await service.generateHabitsInsights(habits: coordinator.habits)
                case .tasks:
                    newInsights = try await service.generateTasksInsights(tasks: coordinator.tasksManager.items)
                case .budget:
                    newInsights = try await service.generateBudgetInsights(
                        transactions: coordinator.budgetManager.transactions,
                        currency: coordinator.profile.currency
                    )
                }

                await MainActor.run {
                    self.insights.removeAll { $0.section == section }
                    let topInsights = Array(newInsights.prefix(5))
                    self.insights.append(contentsOf: topInsights)

                    self.saveInsights()
                    self.markGenerated(for: section)
                    self.saveDataHash(for: section)
                }
            } catch {
                AppLogger.ai.error("Failed to generate AI insights for \(section.rawValue): \(error.localizedDescription)")
                await MainActor.run {
                    self.generateLocal(for: section)
                    self.saveDataHash(for: section)
                }
            }
        }
    }

    private func generateLocal(for section: InsightSection) {
        guard let coordinator = coordinator else { return }

        insights.removeAll { $0.section == section }

        var newInsights: [Insight] = []
        switch section {
        case .habits:
            newInsights = HabitInsightGenerator.generateInsights(from: coordinator.habits)
        case .tasks:
            newInsights = TaskInsightGenerator.generateInsights(from: coordinator.tasksManager.items)
        case .budget:
            newInsights = BudgetInsightGenerator.generateInsights(
                from: coordinator.budgetManager.transactions,
                currency: coordinator.profile.currency
            )
        }

        let topInsights = Array(newInsights.prefix(5))
        insights.append(contentsOf: topInsights)
        saveInsights()
    }

    func refreshAll() {
        for section in InsightSection.allCases {
            generate(for: section)
        }
    }

    // MARK: - Queries
    func insights(for section: InsightSection) -> [Insight] {
        let dismissedHashes = getDismissedHashes()
        return insights.filter { $0.section == section && !dismissedHashes.contains($0.contentHash) }
    }

    // MARK: - Dismiss
    func dismiss(_ insight: Insight) {
        var dismissedHashes = getDismissedHashes()
        dismissedHashes.insert(insight.contentHash)
        saveDismissedHashes(dismissedHashes)

        insights.removeAll { $0.id == insight.id }
        saveInsights()
    }

    private func getDismissedHashes() -> Set<String> {
        let key = "dismissed_insight_hashes"
        let array = UserDefaults.standard.stringArray(forKey: key) ?? []
        return Set(array)
    }

    private func saveDismissedHashes(_ hashes: Set<String>) {
        let key = "dismissed_insight_hashes"
        UserDefaults.standard.set(Array(hashes), forKey: key)
    }

    // MARK: - Clear
    func clear() {
        insights = []
        status = InsightStatus()
        UserDefaults.standard.removeObject(forKey: insightsKey)
        UserDefaults.standard.removeObject(forKey: statusKey)
    }
}

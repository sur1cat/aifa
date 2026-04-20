import Foundation
import Combine

@MainActor
class GoalsManager: ObservableObject {
    @Published var items: [Goal] = []

    private let storageKey = "goals"
    private let service = GoalsService.shared

    weak var coordinator: DataManager?

    // MARK: - Computed Properties
    var active: [Goal] {
        items.filter { $0.isActive }
    }

    var archived: [Goal] {
        items.filter { !$0.isActive }
    }

    // MARK: - Load & Save
    func load() {
        if let data = UserDefaults.standard.data(forKey: storageKey),
           let decoded = try? JSONDecoder().decode([Goal].self, from: data) {
            items = decoded
        }
    }

    func save() {
        if let data = try? JSONEncoder().encode(items) {
            UserDefaults.standard.set(data, forKey: storageKey)
        }
    }

    // MARK: - Sync
    func sync() async {
        guard coordinator?.isDemoMode != true else { return }
        do {
            let serverGoals = try await service.getGoals()
            items = serverGoals
            save()
        } catch {
            coordinator?.syncError = error.localizedDescription
        }
    }

    // MARK: - CRUD Operations
    func add(_ goal: Goal) {
        coordinator?.recordActivity()

        // Optimistic update
        items.append(goal)
        save()

        // Sync to server
        Task { [weak self] in
            guard let self = self else { return }
            do {
                let serverGoal = try await service.createGoal(goal)
                if let index = items.firstIndex(where: { $0.id == goal.id }) {
                    items[index] = serverGoal
                    save()
                }
            } catch {
                items.removeAll { $0.id == goal.id }
                save()
                coordinator?.syncError = error.localizedDescription
            }
        }
    }

    func update(_ goal: Goal) {
        guard let index = items.firstIndex(where: { $0.id == goal.id }) else { return }
        let oldGoal = items[index]

        // Optimistic update
        items[index] = goal
        save()

        // Sync to server
        Task { [weak self] in
            guard let self = self else { return }
            do {
                let serverGoal = try await service.updateGoal(goal)
                if let idx = items.firstIndex(where: { $0.id == goal.id }) {
                    items[idx] = serverGoal
                    save()
                }
            } catch {
                if let idx = items.firstIndex(where: { $0.id == goal.id }) {
                    items[idx] = oldGoal
                    save()
                }
                coordinator?.syncError = error.localizedDescription
            }
        }
    }

    func delete(_ goal: Goal) {
        let removedGoal = goal
        items.removeAll { $0.id == goal.id }
        save()

        // Clear goalId from habits linked to this goal
        coordinator?.clearHabitGoalLinks(goalId: goal.id)

        // Sync to server
        Task { [weak self] in
            guard let self = self else { return }
            do {
                try await service.deleteGoal(goal.id)
            } catch {
                items.append(removedGoal)
                save()
                coordinator?.syncError = error.localizedDescription
            }
        }
    }

    func archive(_ goal: Goal) {
        guard let index = items.firstIndex(where: { $0.id == goal.id }) else { return }

        let oldGoal = items[index]
        var updatedGoal = goal
        updatedGoal.archivedAt = Date()

        items[index] = updatedGoal
        save()

        Task { [weak self] in
            guard let self = self else { return }
            do {
                let serverGoal = try await service.updateGoal(updatedGoal)
                if let idx = items.firstIndex(where: { $0.id == goal.id }) {
                    items[idx] = serverGoal
                    save()
                }
            } catch {
                if let idx = items.firstIndex(where: { $0.id == goal.id }) {
                    items[idx] = oldGoal
                    save()
                }
                coordinator?.syncError = error.localizedDescription
            }
        }
    }

    func unarchive(_ goal: Goal) {
        guard let index = items.firstIndex(where: { $0.id == goal.id }) else { return }

        let oldGoal = items[index]
        var updatedGoal = goal
        updatedGoal.archivedAt = nil

        items[index] = updatedGoal
        save()

        Task { [weak self] in
            guard let self = self else { return }
            do {
                let serverGoal = try await service.updateGoal(updatedGoal)
                if let idx = items.firstIndex(where: { $0.id == goal.id }) {
                    items[idx] = serverGoal
                    save()
                }
            } catch {
                if let idx = items.firstIndex(where: { $0.id == goal.id }) {
                    items[idx] = oldGoal
                    save()
                }
                coordinator?.syncError = error.localizedDescription
            }
        }
    }

    // MARK: - Queries
    func habitsForGoal(_ goal: Goal) -> [Habit] {
        coordinator?.habits.filter { $0.goalId == goal.id && $0.isActive } ?? []
    }

    func habitsWithoutGoal() -> [Habit] {
        coordinator?.habits.filter { $0.goalId == nil && $0.isActive } ?? []
    }

    // MARK: - Clear
    func clear() {
        items = []
        UserDefaults.standard.removeObject(forKey: storageKey)
    }
}

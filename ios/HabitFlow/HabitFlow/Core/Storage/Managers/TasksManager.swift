import Foundation
import Combine

@MainActor
class TasksManager: ObservableObject {
    @Published var items: [DailyTask] = []

    private let storageKey = "tasks_v2"
    private let service = TasksService.shared

    weak var coordinator: DataManager?

    // MARK: - Sorted Tasks
    var sorted: [DailyTask] {
        items.sorted { t1, t2 in
            if t1.isCompleted != t2.isCompleted {
                return !t1.isCompleted
            }
            return t1.priority.sortOrder < t2.priority.sortOrder
        }
    }

    // MARK: - Load & Save
    func load() {
        if let data = UserDefaults.standard.data(forKey: storageKey),
           let decoded = try? JSONDecoder().decode([DailyTask].self, from: data) {
            items = decoded
        }
    }

    func save() {
        if let data = try? JSONEncoder().encode(items) {
            UserDefaults.standard.set(data, forKey: storageKey)
        }
        coordinator?.updateWidgetData()
    }

    // MARK: - Sync
    func sync() async {
        guard coordinator?.isDemoMode != true else { return }
        do {
            let serverTasks = try await service.getTasks(date: nil)
            items = serverTasks
            save()
            coordinator?.generateInsights(for: .tasks)
        } catch {
            coordinator?.syncError = error.localizedDescription
        }
    }

    // MARK: - CRUD Operations
    func add(_ task: DailyTask) {
        coordinator?.trackInsightFirstDate(for: .tasks)
        coordinator?.recordActivity()

        // Optimistic update
        items.append(task)
        save()

        // Sync to server
        Task { [weak self] in
            guard let self = self else { return }
            do {
                let serverTask = try await service.createTask(task)
                if let index = items.firstIndex(where: { $0.id == task.id }) {
                    items[index] = serverTask
                    save()
                }
            } catch {
                items.removeAll { $0.id == task.id }
                save()
                coordinator?.syncError = error.localizedDescription
            }
        }
    }

    func toggle(_ task: DailyTask) {
        guard let index = items.firstIndex(where: { $0.id == task.id }) else { return }

        // Optimistic update
        items[index].isCompleted.toggle()
        save()

        // Sync to server
        Task { [weak self] in
            guard let self = self else { return }
            do {
                let updatedTask = try await service.toggleTask(task.id)
                if let idx = items.firstIndex(where: { $0.id == task.id }) {
                    items[idx] = updatedTask
                    save()
                }
            } catch {
                if let idx = items.firstIndex(where: { $0.id == task.id }) {
                    items[idx].isCompleted.toggle()
                    save()
                }
                coordinator?.syncError = error.localizedDescription
            }
        }
    }

    func delete(_ task: DailyTask) {
        let removedTask = task
        items.removeAll { $0.id == task.id }
        save()

        // Sync to server
        Task { [weak self] in
            guard let self = self else { return }
            do {
                try await service.deleteTask(task.id)
            } catch {
                items.append(removedTask)
                save()
                coordinator?.syncError = error.localizedDescription
            }
        }
    }

    func update(_ task: DailyTask) {
        guard let index = items.firstIndex(where: { $0.id == task.id }) else { return }
        let oldTask = items[index]

        // Optimistic update
        items[index] = task
        save()

        // Sync to server
        Task { [weak self] in
            guard let self = self else { return }
            do {
                let serverTask = try await service.updateTask(task)
                if let idx = items.firstIndex(where: { $0.id == task.id }) {
                    items[idx] = serverTask
                    save()
                }
            } catch {
                if let idx = items.firstIndex(where: { $0.id == task.id }) {
                    items[idx] = oldTask
                    save()
                }
                coordinator?.syncError = error.localizedDescription
            }
        }
    }

    // MARK: - Queries
    func tasksForDate(_ date: Date) -> [DailyTask] {
        let calendar = Calendar.current
        return items.filter { calendar.isDate($0.dueDate, inSameDayAs: date) }
    }

    func completionRate() -> Double {
        guard !items.isEmpty else { return 0 }
        let completed = items.filter { $0.isCompleted }.count
        return Double(completed) / Double(items.count) * 100
    }

    // MARK: - Clear
    func clear() {
        items = []
        UserDefaults.standard.removeObject(forKey: storageKey)
    }
}

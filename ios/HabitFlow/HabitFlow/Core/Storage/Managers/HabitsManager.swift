import Foundation
import Combine

@MainActor
class HabitsManager: ObservableObject {
    @Published var items: [Habit] = []

    private let storageKey = "habits_v2"
    private let service = HabitsService.shared

    weak var coordinator: DataManager?

    // MARK: - Computed Properties
    var active: [Habit] {
        items.filter { $0.isActive }
    }

    var archived: [Habit] {
        items.filter { !$0.isActive }
    }

    // MARK: - Load & Save
    func load() {
        if let data = UserDefaults.standard.data(forKey: storageKey),
           let decoded = try? JSONDecoder().decode([Habit].self, from: data) {
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
            let serverHabits = try await service.getHabits()
            items = serverHabits
            save()
            coordinator?.generateInsights(for: .habits)
        } catch {
            coordinator?.syncError = error.localizedDescription
        }
    }

    // MARK: - CRUD Operations
    func add(_ habit: Habit) {
        coordinator?.trackInsightFirstDate(for: .habits)
        coordinator?.recordActivity()

        // Optimistic update
        items.append(habit)
        save()

        // Sync to server
        Task { [weak self] in
            guard let self = self else { return }
            do {
                let serverHabit = try await service.createHabit(habit)
                if let index = items.firstIndex(where: { $0.id == habit.id }) {
                    items[index] = serverHabit
                    save()
                }
            } catch {
                items.removeAll { $0.id == habit.id }
                save()
                coordinator?.syncError = error.localizedDescription
            }
        }
    }

    func update(_ habit: Habit) {
        guard let index = items.firstIndex(where: { $0.id == habit.id }) else { return }
        let oldHabit = items[index]

        // Optimistic update
        items[index] = habit
        save()

        // Sync to server
        Task { [weak self] in
            guard let self = self else { return }
            do {
                let serverHabit = try await service.updateHabit(habit)
                if let idx = items.firstIndex(where: { $0.id == habit.id }) {
                    items[idx] = serverHabit
                    save()
                }
            } catch {
                if let idx = items.firstIndex(where: { $0.id == habit.id }) {
                    items[idx] = oldHabit
                    save()
                }
                coordinator?.syncError = error.localizedDescription
            }
        }
    }

    func delete(_ habit: Habit) {
        // Cancel reminder
        NotificationManager.shared.cancelHabitReminder(habitID: habit.id)

        let removedHabit = habit
        items.removeAll { $0.id == habit.id }
        save()

        // Sync to server
        Task { [weak self] in
            guard let self = self else { return }
            do {
                try await service.deleteHabit(habit.id)
            } catch {
                items.append(removedHabit)
                save()
                coordinator?.syncError = error.localizedDescription
            }
        }
    }

    // MARK: - Toggle Completion
    func toggle(_ habit: Habit) {
        guard let index = items.firstIndex(where: { $0.id == habit.id }) else { return }
        let today = Habit.todayString

        // Optimistic update
        if items[index].completedDates.contains(today) {
            items[index].completedDates.removeAll { $0 == today }
        } else {
            items[index].completedDates.append(today)
        }
        save()

        // Sync to server
        Task { [weak self] in
            guard let self = self else { return }
            do {
                let updatedHabit = try await service.toggleCompletion(habitID: habit.id, date: today)
                if let idx = items.firstIndex(where: { $0.id == habit.id }) {
                    items[idx] = updatedHabit
                    save()
                }
            } catch {
                // Rollback
                if let idx = items.firstIndex(where: { $0.id == habit.id }) {
                    if items[idx].completedDates.contains(today) {
                        items[idx].completedDates.removeAll { $0 == today }
                    } else {
                        items[idx].completedDates.append(today)
                    }
                    save()
                }
                coordinator?.syncError = error.localizedDescription
            }
        }
    }

    func toggleForDate(_ habit: Habit, date: Date) {
        guard let index = items.firstIndex(where: { $0.id == habit.id }) else { return }
        let dateString = Habit.dateString(from: date)

        // Optimistic update
        if items[index].completedDates.contains(dateString) {
            items[index].completedDates.removeAll { $0 == dateString }
        } else {
            items[index].completedDates.append(dateString)
        }
        save()

        // Sync to server
        Task { [weak self] in
            guard let self = self else { return }
            do {
                let updatedHabit = try await service.toggleCompletion(habitID: habit.id, date: dateString)
                if let idx = items.firstIndex(where: { $0.id == habit.id }) {
                    items[idx] = updatedHabit
                    save()
                }
            } catch {
                // Rollback
                if let idx = items.firstIndex(where: { $0.id == habit.id }) {
                    if items[idx].completedDates.contains(dateString) {
                        items[idx].completedDates.removeAll { $0 == dateString }
                    } else {
                        items[idx].completedDates.append(dateString)
                    }
                    save()
                }
                coordinator?.syncError = error.localizedDescription
            }
        }
    }

    // MARK: - Progress
    func incrementProgress(_ habit: Habit) {
        guard let index = items.firstIndex(where: { $0.id == habit.id }) else { return }
        let today = Habit.todayString

        let currentProgress = items[index].progressValues[today] ?? 0
        let newProgress = currentProgress + 1

        // Optimistic update
        items[index].progressValues[today] = newProgress
        if let target = items[index].targetValue, newProgress >= target {
            if !items[index].completedDates.contains(today) {
                items[index].completedDates.append(today)
            }
        }
        save()

        // Sync to server
        Task { [weak self] in
            guard let self = self else { return }
            do {
                let updatedHabit = try await service.setProgress(habitID: habit.id, date: today, value: newProgress)
                if let idx = items.firstIndex(where: { $0.id == habit.id }) {
                    items[idx] = updatedHabit
                    save()
                }
            } catch {
                // Rollback
                if let idx = items.firstIndex(where: { $0.id == habit.id }) {
                    items[idx].progressValues[today] = currentProgress
                    if let target = items[idx].targetValue, currentProgress < target {
                        items[idx].completedDates.removeAll { $0 == today }
                    }
                    save()
                }
                coordinator?.syncError = error.localizedDescription
            }
        }
    }

    func incrementProgressForDate(_ habit: Habit, date: Date) {
        guard let index = items.firstIndex(where: { $0.id == habit.id }) else { return }
        let dateString = Habit.dateString(from: date)

        let currentProgress = items[index].progressValues[dateString] ?? 0
        let newProgress = currentProgress + 1

        // Optimistic update
        items[index].progressValues[dateString] = newProgress
        if let target = items[index].targetValue, newProgress >= target {
            if !items[index].completedDates.contains(dateString) {
                items[index].completedDates.append(dateString)
            }
        }
        save()

        // Sync to server
        Task { [weak self] in
            guard let self = self else { return }
            do {
                let updatedHabit = try await service.setProgress(habitID: habit.id, date: dateString, value: newProgress)
                if let idx = items.firstIndex(where: { $0.id == habit.id }) {
                    items[idx] = updatedHabit
                    save()
                }
            } catch {
                // Rollback
                if let idx = items.firstIndex(where: { $0.id == habit.id }) {
                    items[idx].progressValues[dateString] = currentProgress
                    if let target = items[idx].targetValue, currentProgress < target {
                        items[idx].completedDates.removeAll { $0 == dateString }
                    }
                    save()
                }
                coordinator?.syncError = error.localizedDescription
            }
        }
    }

    func setProgress(_ habit: Habit, value: Int, date: Date = Date()) {
        guard let index = items.firstIndex(where: { $0.id == habit.id }) else { return }
        let dateString = Habit.dateString(from: date)

        let currentProgress = items[index].progressValues[dateString] ?? 0

        // Optimistic update
        items[index].progressValues[dateString] = value
        if let target = items[index].targetValue, value >= target {
            if !items[index].completedDates.contains(dateString) {
                items[index].completedDates.append(dateString)
            }
        } else {
            items[index].completedDates.removeAll { $0 == dateString }
        }
        save()

        // Sync to server
        Task { [weak self] in
            guard let self = self else { return }
            do {
                let updatedHabit = try await service.setProgress(habitID: habit.id, date: dateString, value: value)
                if let idx = items.firstIndex(where: { $0.id == habit.id }) {
                    items[idx] = updatedHabit
                    save()
                }
            } catch {
                // Rollback
                if let idx = items.firstIndex(where: { $0.id == habit.id }) {
                    items[idx].progressValues[dateString] = currentProgress
                    if let target = items[idx].targetValue, currentProgress >= target {
                        if !items[idx].completedDates.contains(dateString) {
                            items[idx].completedDates.append(dateString)
                        }
                    } else {
                        items[idx].completedDates.removeAll { $0 == dateString }
                    }
                    save()
                }
                coordinator?.syncError = error.localizedDescription
            }
        }
    }

    // MARK: - Archive
    func archive(_ habit: Habit) {
        guard let index = items.firstIndex(where: { $0.id == habit.id }) else { return }

        // Cancel reminder
        NotificationManager.shared.cancelHabitReminder(habitID: habit.id)

        let oldHabit = items[index]
        var updatedHabit = habit
        updatedHabit.archivedAt = Date()
        updatedHabit.reminderEnabled = false

        // Optimistic update
        items[index] = updatedHabit
        save()

        // Sync to server
        Task { [weak self] in
            guard let self = self else { return }
            do {
                let serverHabit = try await service.updateHabit(updatedHabit)
                if let idx = items.firstIndex(where: { $0.id == habit.id }) {
                    items[idx] = serverHabit
                    save()
                }
            } catch {
                if let idx = items.firstIndex(where: { $0.id == habit.id }) {
                    items[idx] = oldHabit
                    save()
                }
                coordinator?.syncError = error.localizedDescription
            }
        }
    }

    func unarchive(_ habit: Habit) {
        guard let index = items.firstIndex(where: { $0.id == habit.id }) else { return }

        let oldHabit = items[index]
        var updatedHabit = habit
        updatedHabit.archivedAt = nil

        // Optimistic update
        items[index] = updatedHabit
        save()

        // Sync to server
        Task { [weak self] in
            guard let self = self else { return }
            do {
                let serverHabit = try await service.updateHabit(updatedHabit)
                if let idx = items.firstIndex(where: { $0.id == habit.id }) {
                    items[idx] = serverHabit
                    save()
                }
            } catch {
                if let idx = items.firstIndex(where: { $0.id == habit.id }) {
                    items[idx] = oldHabit
                    save()
                }
                coordinator?.syncError = error.localizedDescription
            }
        }
    }

    // MARK: - Reminder
    func updateReminder(_ habit: Habit) {
        guard let index = items.firstIndex(where: { $0.id == habit.id }) else { return }
        items[index].reminderEnabled = habit.reminderEnabled
        items[index].reminderTime = habit.reminderTime
        save()
    }

    // MARK: - Queries
    func completionRate(for period: AnalyticsPeriod) -> Double {
        guard !items.isEmpty else { return 0 }
        let completed = items.filter { $0.isCompletedInCurrentPeriod }.count
        return Double(completed) / Double(items.count) * 100
    }

    func completionsForWeek(_ habit: Habit) -> [(day: String, completed: Bool)] {
        let calendar = Calendar.current
        var result: [(String, Bool)] = []

        for dayOffset in (0..<7).reversed() {
            guard let date = calendar.date(byAdding: .day, value: -dayOffset, to: Date()) else { continue }
            let dayName = DateFormatters.shortWeekday.string(from: date)
            let dateStr = Habit.dateString(from: date)
            let completed = habit.completedDates.contains(dateStr)
            result.append((dayName, completed))
        }
        return result
    }

    // MARK: - Goal Links
    func clearGoalLinks(goalId: UUID) {
        for i in items.indices {
            if items[i].goalId == goalId {
                items[i].goalId = nil
            }
        }
        save()
    }

    // MARK: - Clear
    func clear() {
        items = []
        UserDefaults.standard.removeObject(forKey: storageKey)
    }
}

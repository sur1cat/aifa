import Foundation
import WatchConnectivity
import Combine
import os

@MainActor
class WatchConnectivityManager: NSObject, ObservableObject {
    static let shared = WatchConnectivityManager()

    @Published var isWatchAppInstalled = false
    @Published var isReachable = false

    private var session: WCSession?

    override init() {
        super.init()
        if WCSession.isSupported() {
            session = WCSession.default
            session?.delegate = self
            session?.activate()
        }
    }

    func sendDataToWatch() {
        guard let session = session, session.isReachable else { return }

        let dataManager = DataManager.shared

        // Prepare habits data
        let watchHabits = dataManager.habits.map { habit in
            [
                "id": habit.id.uuidString,
                "title": habit.title,
                "icon": habit.icon,
                "color": habit.color,
                "isCompleted": habit.isCompletedInCurrentPeriod,
                "streak": habit.streak
            ] as [String: Any]
        }

        // Prepare tasks data
        let watchTasks = dataManager.tasks.map { task in
            [
                "id": task.id.uuidString,
                "title": task.title,
                "isCompleted": task.isCompleted,
                "priority": task.priority.rawValue
            ] as [String: Any]
        }

        // Prepare budget data
        let income = dataManager.transactions.filter { $0.type == .income }.reduce(0) { $0 + $1.amount }
        let expenses = dataManager.transactions.filter { $0.type == .expense }.reduce(0) { $0 + $1.amount }

        let watchBudget: [String: Any] = [
            "balance": income - expenses,
            "currencySymbol": dataManager.profile.currency.symbol
        ]

        // Send data
        do {
            let habitsData = try JSONSerialization.data(withJSONObject: watchHabits)
            let tasksData = try JSONSerialization.data(withJSONObject: watchTasks)
            let budgetData = try JSONSerialization.data(withJSONObject: watchBudget)

            let message: [String: Any] = [
                "habits": habitsData,
                "tasks": tasksData,
                "budget": budgetData
            ]

            session.sendMessage(message, replyHandler: nil, errorHandler: { error in
                AppLogger.watch.error("Error sending to watch: \(error.localizedDescription)")
            })
        } catch {
            AppLogger.watch.error("Error encoding watch data: \(error.localizedDescription)")
        }
    }

    func updateApplicationContext() {
        guard let session = session, session.activationState == .activated else { return }

        let dataManager = DataManager.shared

        do {
            let habitsData = try JSONEncoder().encode(
                dataManager.habits.map { habit in
                    WatchHabitDTO(
                        id: habit.id,
                        title: habit.title,
                        icon: habit.icon,
                        color: habit.color,
                        isCompleted: habit.isCompletedInCurrentPeriod,
                        streak: habit.streak
                    )
                }
            )

            let tasksData = try JSONEncoder().encode(
                dataManager.tasks.map { task in
                    WatchTaskDTO(
                        id: task.id,
                        title: task.title,
                        isCompleted: task.isCompleted,
                        priority: task.priority.rawValue
                    )
                }
            )

            let income = dataManager.transactions.filter { $0.type == .income }.reduce(0) { $0 + $1.amount }
            let expenses = dataManager.transactions.filter { $0.type == .expense }.reduce(0) { $0 + $1.amount }

            let budgetData = try JSONEncoder().encode(
                WatchBudgetDTO(
                    balance: income - expenses,
                    currencySymbol: dataManager.profile.currency.symbol
                )
            )

            try session.updateApplicationContext([
                "habits": habitsData,
                "tasks": tasksData,
                "budget": budgetData
            ])
        } catch {
            print("Error updating watch context: \(error)")
        }
    }
}

extension WatchConnectivityManager: WCSessionDelegate {
    nonisolated func session(_ session: WCSession, activationDidCompleteWith activationState: WCSessionActivationState, error: Error?) {
        Task { @MainActor in
            self.isWatchAppInstalled = session.isWatchAppInstalled
        }
    }

    nonisolated func sessionReachabilityDidChange(_ session: WCSession) {
        Task { @MainActor in
            self.isReachable = session.isReachable
        }
    }

    nonisolated func sessionDidBecomeInactive(_ session: WCSession) {}
    nonisolated func sessionDidDeactivate(_ session: WCSession) {
        session.activate()
    }

    nonisolated func session(_ session: WCSession, didReceiveMessage message: [String: Any], replyHandler: @escaping ([String: Any]) -> Void) {
        Task { @MainActor in
            if message["request"] as? String == "data" {
                // Watch is requesting data
                let dataManager = DataManager.shared

                do {
                    let habitsData = try JSONEncoder().encode(
                        dataManager.habits.map { habit in
                            WatchHabitDTO(
                                id: habit.id,
                                title: habit.title,
                                icon: habit.icon,
                                color: habit.color,
                                isCompleted: habit.isCompletedInCurrentPeriod,
                                streak: habit.streak
                            )
                        }
                    )

                    let tasksData = try JSONEncoder().encode(
                        dataManager.tasks.map { task in
                            WatchTaskDTO(
                                id: task.id,
                                title: task.title,
                                isCompleted: task.isCompleted,
                                priority: task.priority.rawValue
                            )
                        }
                    )

                    let income = dataManager.transactions.filter { $0.type == .income }.reduce(0) { $0 + $1.amount }
                    let expenses = dataManager.transactions.filter { $0.type == .expense }.reduce(0) { $0 + $1.amount }

                    let budgetData = try JSONEncoder().encode(
                        WatchBudgetDTO(
                            balance: income - expenses,
                            currencySymbol: dataManager.profile.currency.symbol
                        )
                    )

                    replyHandler([
                        "habits": habitsData,
                        "tasks": tasksData,
                        "budget": budgetData
                    ])
                } catch {
                    replyHandler([:])
                }
            } else if let action = message["action"] as? String {
                let dm = DataManager.shared
                if action == "toggleHabit", let habitIdString = message["habitId"] as? String,
                   let habitId = UUID(uuidString: habitIdString),
                   let habit = dm.habits.first(where: { $0.id == habitId }) {
                    dm.toggleHabit(habit)
                } else if action == "toggleTask", let taskIdString = message["taskId"] as? String,
                          let taskId = UUID(uuidString: taskIdString),
                          let task = dm.tasks.first(where: { $0.id == taskId }) {
                    dm.toggleTask(task)
                }
                replyHandler([:])
            }
        }
    }
}

// DTOs for Watch communication
struct WatchHabitDTO: Codable {
    let id: UUID
    let title: String
    let icon: String
    let color: String
    let isCompleted: Bool
    let streak: Int
}

struct WatchTaskDTO: Codable {
    let id: UUID
    let title: String
    let isCompleted: Bool
    let priority: String
}

struct WatchBudgetDTO: Codable {
    let balance: Double
    let currencySymbol: String
}

import Foundation
import WatchConnectivity

struct WatchHabit: Identifiable, Codable {
    let id: UUID
    let title: String
    let icon: String
    let color: String
    var isCompleted: Bool
    let streak: Int
}

struct WatchTask: Identifiable, Codable {
    let id: UUID
    let title: String
    var isCompleted: Bool
    let priority: String
}

struct WatchBudget: Codable {
    let balance: Double
    let currencySymbol: String
}

class WatchDataStore: NSObject, ObservableObject {
    static let shared = WatchDataStore()

    @Published var habits: [WatchHabit] = []
    @Published var tasks: [WatchTask] = []
    @Published var budget: WatchBudget = WatchBudget(balance: 0, currencySymbol: "$")
    @Published var isConnected = false

    private var session: WCSession?

    override init() {
        super.init()
        if WCSession.isSupported() {
            session = WCSession.default
            session?.delegate = self
            session?.activate()
        }
        loadCachedData()
    }

    func requestData() {
        guard let session = session, session.isReachable else { return }
        session.sendMessage(["request": "data"], replyHandler: { response in
            DispatchQueue.main.async {
                self.processResponse(response)
            }
        }, errorHandler: { error in
            print("Watch request error: \(error)")
        })
    }

    func toggleHabit(_ habit: WatchHabit) {
        guard let session = session, session.isReachable else { return }

        // Optimistic update
        if let index = habits.firstIndex(where: { $0.id == habit.id }) {
            habits[index].isCompleted.toggle()
        }

        session.sendMessage([
            "action": "toggleHabit",
            "habitId": habit.id.uuidString
        ], replyHandler: nil, errorHandler: { error in
            print("Toggle habit error: \(error)")
        })
    }

    func toggleTask(_ task: WatchTask) {
        guard let session = session, session.isReachable else { return }

        // Optimistic update
        if let index = tasks.firstIndex(where: { $0.id == task.id }) {
            tasks[index].isCompleted.toggle()
        }

        session.sendMessage([
            "action": "toggleTask",
            "taskId": task.id.uuidString
        ], replyHandler: nil, errorHandler: { error in
            print("Toggle task error: \(error)")
        })
    }

    private func processResponse(_ response: [String: Any]) {
        if let habitsData = response["habits"] as? Data,
           let decoded = try? JSONDecoder().decode([WatchHabit].self, from: habitsData) {
            habits = decoded
        }

        if let tasksData = response["tasks"] as? Data,
           let decoded = try? JSONDecoder().decode([WatchTask].self, from: tasksData) {
            tasks = decoded
        }

        if let budgetData = response["budget"] as? Data,
           let decoded = try? JSONDecoder().decode(WatchBudget.self, from: budgetData) {
            budget = decoded
        }

        saveCachedData()
    }

    private func loadCachedData() {
        let defaults = UserDefaults.standard

        if let data = defaults.data(forKey: "watchHabits"),
           let decoded = try? JSONDecoder().decode([WatchHabit].self, from: data) {
            habits = decoded
        }

        if let data = defaults.data(forKey: "watchTasks"),
           let decoded = try? JSONDecoder().decode([WatchTask].self, from: data) {
            tasks = decoded
        }

        if let data = defaults.data(forKey: "watchBudget"),
           let decoded = try? JSONDecoder().decode(WatchBudget.self, from: data) {
            budget = decoded
        }
    }

    private func saveCachedData() {
        let defaults = UserDefaults.standard

        if let data = try? JSONEncoder().encode(habits) {
            defaults.set(data, forKey: "watchHabits")
        }
        if let data = try? JSONEncoder().encode(tasks) {
            defaults.set(data, forKey: "watchTasks")
        }
        if let data = try? JSONEncoder().encode(budget) {
            defaults.set(data, forKey: "watchBudget")
        }
    }
}

extension WatchDataStore: WCSessionDelegate {
    func session(_ session: WCSession, activationDidCompleteWith activationState: WCSessionActivationState, error: Error?) {
        DispatchQueue.main.async {
            self.isConnected = activationState == .activated
            if self.isConnected {
                self.requestData()
            }
        }
    }

    func session(_ session: WCSession, didReceiveMessage message: [String: Any]) {
        DispatchQueue.main.async {
            self.processResponse(message)
        }
    }

    func session(_ session: WCSession, didReceiveApplicationContext applicationContext: [String: Any]) {
        DispatchQueue.main.async {
            self.processResponse(applicationContext)
        }
    }
}

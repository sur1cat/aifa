import Foundation
import UserNotifications
import Combine
import UIKit
import SwiftUI

@MainActor
class NotificationManager: ObservableObject {
    static let shared = NotificationManager()

    @Published var isAuthorized = false
    @Published var deviceToken: String?

    // User preferences
    @AppStorage("morningDigestEnabled") var morningDigestEnabled = false
    @AppStorage("morningDigestTime") var morningDigestHour = 8
    @AppStorage("eveningDigestEnabled") var eveningDigestEnabled = false
    @AppStorage("eveningDigestTime") var eveningDigestHour = 21

    private let center = UNUserNotificationCenter.current()

    private init() {
        Task {
            await checkAuthorizationStatus()
        }
    }

    // MARK: - Authorization

    func checkAuthorizationStatus() async {
        let settings = await center.notificationSettings()
        isAuthorized = settings.authorizationStatus == .authorized
    }

    func requestPermission() async -> Bool {
        do {
            let granted = try await center.requestAuthorization(options: [.alert, .sound, .badge])
            await checkAuthorizationStatus()
            return granted
        } catch {
            print("Notification permission error: \(error)")
            return false
        }
    }

    // MARK: - Schedule Reminders

    func scheduleHabitReminder(habit: Habit) async {
        guard habit.reminderEnabled, let reminderTime = habit.reminderTime else {
            cancelHabitReminder(habitID: habit.id)
            return
        }

        // Cancel existing notification first
        cancelHabitReminder(habitID: habit.id)

        // Create content
        let content = UNMutableNotificationContent()
        content.title = "Atoma"
        content.body = "\(habit.icon) \(habit.title)"
        content.subtitle = "Time to build your routine!"
        content.sound = .default
        content.categoryIdentifier = "HABIT_REMINDER"
        content.userInfo = ["habitID": habit.id.uuidString]

        // Create trigger based on period
        let trigger = createTrigger(for: habit.period, at: reminderTime)

        // Create request
        let request = UNNotificationRequest(
            identifier: notificationID(for: habit.id),
            content: content,
            trigger: trigger
        )

        do {
            try await center.add(request)
            print("Scheduled reminder for \(habit.title) at \(reminderTime)")
        } catch {
            print("Failed to schedule notification: \(error)")
        }
    }

    func cancelHabitReminder(habitID: UUID) {
        center.removePendingNotificationRequests(withIdentifiers: [notificationID(for: habitID)])
    }

    func cancelAllReminders() {
        center.removeAllPendingNotificationRequests()
    }

    func rescheduleAllReminders(habits: [Habit]) async {
        cancelAllReminders()

        for habit in habits where habit.reminderEnabled {
            await scheduleHabitReminder(habit: habit)
        }
    }

    // MARK: - Helpers

    private func notificationID(for habitID: UUID) -> String {
        "habit_reminder_\(habitID.uuidString)"
    }

    private func createTrigger(for period: HabitPeriod, at time: Date) -> UNNotificationTrigger {
        let calendar = Calendar.current
        let components = calendar.dateComponents([.hour, .minute], from: time)

        switch period {
        case .daily:
            // Repeat every day at the specified time
            var dateComponents = DateComponents()
            dateComponents.hour = components.hour
            dateComponents.minute = components.minute
            return UNCalendarNotificationTrigger(dateMatching: dateComponents, repeats: true)

        case .weekly:
            // Repeat every week on the same day
            let weekday = calendar.component(.weekday, from: Date())
            var dateComponents = DateComponents()
            dateComponents.weekday = weekday
            dateComponents.hour = components.hour
            dateComponents.minute = components.minute
            return UNCalendarNotificationTrigger(dateMatching: dateComponents, repeats: true)

        case .monthly:
            // Repeat every month on the same day
            let day = calendar.component(.day, from: Date())
            var dateComponents = DateComponents()
            dateComponents.day = min(day, 28) // Avoid issues with months having fewer days
            dateComponents.hour = components.hour
            dateComponents.minute = components.minute
            return UNCalendarNotificationTrigger(dateMatching: dateComponents, repeats: true)
        }
    }

    // MARK: - Task Reminders

    func scheduleTaskReminder(task: DailyTask, reminderTime: Date) async {
        // Cancel existing
        cancelTaskReminder(taskID: task.id)

        // Don't schedule if task is completed or in the past
        guard !task.isCompleted else { return }

        let content = UNMutableNotificationContent()
        content.title = "Atoma Tasks"
        content.body = "📋 \(task.title)"
        content.subtitle = task.priority == .high ? "High priority!" : "Don't forget!"
        content.sound = .default
        content.categoryIdentifier = "TASK_REMINDER"
        content.userInfo = ["taskID": task.id.uuidString]

        // Create trigger for specific date/time
        let triggerDate = Calendar.current.dateComponents([.year, .month, .day, .hour, .minute], from: reminderTime)
        let trigger = UNCalendarNotificationTrigger(dateMatching: triggerDate, repeats: false)

        let request = UNNotificationRequest(
            identifier: "task_reminder_\(task.id.uuidString)",
            content: content,
            trigger: trigger
        )

        do {
            try await center.add(request)
            print("Scheduled task reminder for \(task.title)")
        } catch {
            print("Failed to schedule task notification: \(error)")
        }
    }

    func cancelTaskReminder(taskID: UUID) {
        center.removePendingNotificationRequests(withIdentifiers: ["task_reminder_\(taskID.uuidString)"])
    }

    // MARK: - Morning/Evening Digest

    func scheduleMorningDigest() async {
        guard morningDigestEnabled else {
            center.removePendingNotificationRequests(withIdentifiers: ["morning_digest"])
            return
        }

        let content = UNMutableNotificationContent()
        content.title = "Good morning! ☀️"
        content.body = "Check your habits and tasks for today"
        content.sound = .default
        content.categoryIdentifier = "DAILY_DIGEST"

        var dateComponents = DateComponents()
        dateComponents.hour = morningDigestHour
        dateComponents.minute = 0

        let trigger = UNCalendarNotificationTrigger(dateMatching: dateComponents, repeats: true)

        let request = UNNotificationRequest(
            identifier: "morning_digest",
            content: content,
            trigger: trigger
        )

        do {
            try await center.add(request)
            print("Scheduled morning digest at \(morningDigestHour):00")
        } catch {
            print("Failed to schedule morning digest: \(error)")
        }
    }

    func scheduleEveningDigest() async {
        guard eveningDigestEnabled else {
            center.removePendingNotificationRequests(withIdentifiers: ["evening_digest"])
            return
        }

        let content = UNMutableNotificationContent()
        content.title = "Evening check-in 🌙"
        content.body = "Did you complete your habits today?"
        content.sound = .default
        content.categoryIdentifier = "DAILY_DIGEST"

        var dateComponents = DateComponents()
        dateComponents.hour = eveningDigestHour
        dateComponents.minute = 0

        let trigger = UNCalendarNotificationTrigger(dateMatching: dateComponents, repeats: true)

        let request = UNNotificationRequest(
            identifier: "evening_digest",
            content: content,
            trigger: trigger
        )

        do {
            try await center.add(request)
            print("Scheduled evening digest at \(eveningDigestHour):00")
        } catch {
            print("Failed to schedule evening digest: \(error)")
        }
    }

    func updateDigestSchedule() async {
        await scheduleMorningDigest()
        await scheduleEveningDigest()
    }

    // MARK: - Remote Push (APNs)

    func registerForRemotePush() {
        UIApplication.shared.registerForRemoteNotifications()
    }

    func setDeviceToken(_ token: Data) {
        let tokenString = token.map { String(format: "%02.2hhx", $0) }.joined()
        self.deviceToken = tokenString
        print("Device token: \(tokenString)")

        // Send token to backend for server-side push
        Task {
            await registerDeviceTokenOnServer(tokenString)
        }
    }

    private func registerDeviceTokenOnServer(_ token: String) async {
        do {
            let _: EmptyResponse = try await APIClient.shared.request(
                endpoint: "push/register",
                method: "POST",
                body: ["token": token, "platform": "ios"],
                requiresAuth: true
            )
            print("Device token registered on server")
        } catch {
            print("Failed to register device token: \(error)")
        }
    }

    struct EmptyResponse: Codable {}

    // MARK: - Debug

    func listPendingNotifications() async {
        let requests = await center.pendingNotificationRequests()
        print("Pending notifications: \(requests.count)")
        for request in requests {
            print("  - \(request.identifier): \(request.content.body)")
        }
    }
}

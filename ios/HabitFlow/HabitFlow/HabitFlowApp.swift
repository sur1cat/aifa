import SwiftUI
import GoogleSignIn
import UserNotifications
import os

@main
struct HabitFlowApp: App {
    @StateObject private var authManager = AuthManager.shared
    @StateObject private var dataManager = DataManager.shared
    @StateObject private var notificationManager = NotificationManager.shared
    @AppStorage("hasCompletedOnboarding") private var hasCompletedOnboarding = false
    @UIApplicationDelegateAdaptor(AppDelegate.self) var appDelegate

    var body: some Scene {
        WindowGroup {
            Group {
                if authManager.isAuthenticated {
                    if hasCompletedOnboarding {
                        MainTabView()
                    } else {
                        OnboardingView()
                    }
                } else {
                    LoginView()
                }
            }
            .environmentObject(authManager)
            .environmentObject(dataManager)
            .onOpenURL { url in
                // Handle Google Sign-In callback
                // URL scheme is reversed client ID: com.googleusercontent.apps.CLIENT_ID
                if let scheme = url.scheme, scheme.hasPrefix("com.googleusercontent.apps") {
                    GIDSignIn.sharedInstance.handle(url)
                    return
                }
            }
            .task {
                // Request notification permission on first launch
                await notificationManager.checkAuthorizationStatus()
            }
        }
    }
}

// MARK: - App Delegate for Notifications
class AppDelegate: NSObject, UIApplicationDelegate, UNUserNotificationCenterDelegate {
    func application(_ application: UIApplication, didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?) -> Bool {
        UNUserNotificationCenter.current().delegate = self

        // Restore previous Google Sign-In state
        GIDSignIn.sharedInstance.restorePreviousSignIn { user, error in
            if let error = error {
                AppLogger.auth.error("Google Sign-In restore error: \(error.localizedDescription)")
            } else if let user = user {
                AppLogger.auth.info("Google Sign-In restored user: \(user.profile?.email ?? "unknown")")
            }
        }

        return true
    }

    // Handle Google Sign-In URL callback (iOS < 13 fallback)
    func application(_ app: UIApplication, open url: URL, options: [UIApplication.OpenURLOptionsKey : Any] = [:]) -> Bool {
        if let scheme = url.scheme, scheme.hasPrefix("com.googleusercontent.apps") {
            return GIDSignIn.sharedInstance.handle(url)
        }
        return false
    }

    // Handle notification when app is in foreground
    func userNotificationCenter(_ center: UNUserNotificationCenter, willPresent notification: UNNotification, withCompletionHandler completionHandler: @escaping (UNNotificationPresentationOptions) -> Void) {
        completionHandler([.banner, .sound])
    }

    // Handle notification tap
    func userNotificationCenter(_ center: UNUserNotificationCenter, didReceive response: UNNotificationResponse, withCompletionHandler completionHandler: @escaping () -> Void) {
        let userInfo = response.notification.request.content.userInfo

        // Handle different notification types
        if let habitID = userInfo["habitID"] as? String {
            AppLogger.notifications.debug("Tapped habit notification: \(habitID)")
            // Navigate to habits tab
        } else if let taskID = userInfo["taskID"] as? String {
            AppLogger.notifications.debug("Tapped task notification: \(taskID)")
            // Navigate to tasks tab
        }

        completionHandler()
    }

    // MARK: - Remote Push Token
    func application(_ application: UIApplication, didRegisterForRemoteNotificationsWithDeviceToken deviceToken: Data) {
        Task { @MainActor in
            NotificationManager.shared.setDeviceToken(deviceToken)
        }
    }

    func application(_ application: UIApplication, didFailToRegisterForRemoteNotificationsWithError error: Error) {
        AppLogger.notifications.error("Failed to register for remote notifications: \(error.localizedDescription)")
    }
}

struct MainTabView: View {
    @State private var selectedTab = 1  // Default to Habits tab

    var body: some View {
        TabView(selection: $selectedTab) {
            DashboardView()
                .tabItem {
                    Image(systemName: "square.grid.2x2")
                    Text("Dashboard")
                }
                .tag(0)

            HabitsView()
                .tabItem {
                    Image(systemName: "repeat")
                    Text("Habits")
                }
                .tag(1)

            TasksView()
                .tabItem {
                    Image(systemName: "checkmark.circle")
                    Text("Tasks")
                }
                .tag(2)

            BudgetView()
                .tabItem {
                    Image(systemName: "creditcard")
                    Text("Budget")
                }
                .tag(3)

            ProfileView()
                .tabItem {
                    Image(systemName: "person")
                    Text("Profile")
                }
                .tag(4)
        }
        .tint(Color.hf.accent)
        .withSyncErrorBanner()
    }
}

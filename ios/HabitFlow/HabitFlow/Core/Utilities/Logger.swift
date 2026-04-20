import Foundation
import os

/// Centralized logging utility for Atoma app
enum AppLogger {
    /// Subsystem identifier for the app
    private static let subsystem = Bundle.main.bundleIdentifier ?? "com.azamatbigali.habitflow"

    // MARK: - Category Loggers

    /// Networking operations (API calls, responses)
    static let network = Logger(subsystem: subsystem, category: "Network")

    /// Authentication operations
    static let auth = Logger(subsystem: subsystem, category: "Auth")

    /// Data synchronization
    static let sync = Logger(subsystem: subsystem, category: "Sync")

    /// Local storage operations
    static let storage = Logger(subsystem: subsystem, category: "Storage")

    /// Push notifications
    static let notifications = Logger(subsystem: subsystem, category: "Notifications")

    /// AI/Insights operations
    static let ai = Logger(subsystem: subsystem, category: "AI")

    /// General app operations
    static let app = Logger(subsystem: subsystem, category: "App")

    /// Widgets
    static let widget = Logger(subsystem: subsystem, category: "Widget")

    /// Watch connectivity
    static let watch = Logger(subsystem: subsystem, category: "Watch")
}

// MARK: - Convenience Extensions

extension Logger {
    /// Log debug message with file/line info
    func debugWithContext(_ message: String, file: String = #file, function: String = #function, line: Int = #line) {
        #if DEBUG
        let filename = (file as NSString).lastPathComponent
        self.debug("\(filename):\(line) \(function) - \(message)")
        #endif
    }

    /// Log error with Error object
    func error(_ message: String, error: Error) {
        self.error("\(message): \(error.localizedDescription)")
    }
}

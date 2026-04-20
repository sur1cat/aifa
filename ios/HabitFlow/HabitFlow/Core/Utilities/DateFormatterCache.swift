import Foundation

/// Cached DateFormatters for common date/time formatting patterns
/// DateFormatter creation is expensive, so we cache instances for reuse
enum DateFormatters {
    /// "yyyy-MM-dd" - Standard date format for API
    static let apiDate: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd"
        formatter.locale = Locale(identifier: "en_US_POSIX")
        return formatter
    }()

    /// ISO8601 with full timestamp
    static let iso8601: ISO8601DateFormatter = {
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        return formatter
    }()

    /// ISO8601 basic (without fractional seconds)
    static let iso8601Basic: ISO8601DateFormatter = {
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime]
        return formatter
    }()

    /// "HH:mm" - 24-hour time
    static let time24h: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateFormat = "HH:mm"
        return formatter
    }()

    /// "h:mm a" - 12-hour time with AM/PM
    static let time12h: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateFormat = "h:mm a"
        return formatter
    }()

    /// "MMM d" - Short month and day (e.g., "Jan 5")
    static let shortMonthDay: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateFormat = "MMM d"
        return formatter
    }()

    /// "MMMM yyyy" - Full month and year (e.g., "January 2024")
    static let monthYear: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateFormat = "MMMM yyyy"
        return formatter
    }()

    /// "MMMM d, yyyy" - Full date (e.g., "January 5, 2024")
    static let fullDate: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateFormat = "MMMM d, yyyy"
        return formatter
    }()

    /// "MMM d, yyyy" - Medium date (e.g., "Jan 5, 2024")
    static let mediumDate: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateFormat = "MMM d, yyyy"
        return formatter
    }()

    /// "EEEE" - Full weekday name (e.g., "Monday")
    static let weekday: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateFormat = "EEEE"
        return formatter
    }()

    /// "EEE" - Short weekday name (e.g., "Mon")
    static let shortWeekday: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateFormat = "EEE"
        return formatter
    }()

    /// "d" - Day of month (e.g., "5")
    static let dayOfMonth: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateFormat = "d"
        return formatter
    }()

    /// "MMM" - Short month (e.g., "Jan")
    static let shortMonth: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateFormat = "MMM"
        return formatter
    }()

    /// "MMMM" - Full month name (e.g., "January")
    static let fullMonth: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateFormat = "MMMM"
        return formatter
    }()

    /// "yyyy" - Year only
    static let year: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy"
        return formatter
    }()

    /// "MMM d, h:mm a" - Date with time (e.g., "Jan 5, 2:30 PM")
    static let dateTime: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateFormat = "MMM d, h:mm a"
        return formatter
    }()
}

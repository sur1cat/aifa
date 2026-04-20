import Foundation
import SwiftUI
import CryptoKit

// MARK: - Habit
enum HabitPeriod: String, Codable, CaseIterable, Sendable {
    case daily = "daily"
    case weekly = "weekly"
    case monthly = "monthly"

    var title: LocalizedStringKey {
        switch self {
        case .daily: return "Daily"
        case .weekly: return "Weekly"
        case .monthly: return "Monthly"
        }
    }

    var shortTitle: LocalizedStringKey {
        switch self {
        case .daily: return "day"
        case .weekly: return "week"
        case .monthly: return "month"
        }
    }

    var shortTitleString: String {
        switch self {
        case .daily: return String(localized: "day")
        case .weekly: return String(localized: "week")
        case .monthly: return String(localized: "month")
        }
    }
}

// MARK: - Goal
struct Goal: Identifiable, Codable, Sendable {
    let id: UUID
    var title: String
    var icon: String
    var targetValue: Int?
    var unit: String?
    var deadline: Date?
    let createdAt: Date
    var archivedAt: Date?

    init(id: UUID = UUID(), title: String, icon: String = "🎯", targetValue: Int? = nil, unit: String? = nil, deadline: Date? = nil, createdAt: Date = Date(), archivedAt: Date? = nil) {
        self.id = id
        self.title = title
        self.icon = icon
        self.targetValue = targetValue
        self.unit = unit
        self.deadline = deadline
        self.createdAt = createdAt
        self.archivedAt = archivedAt
    }

    /// Whether goal is currently active (not archived)
    var isActive: Bool {
        archivedAt == nil
    }
}

struct Habit: Identifiable, Codable, Sendable {
    let id: UUID
    var goalId: UUID?  // Optional goal this habit belongs to
    var title: String
    var icon: String
    var color: String
    var period: HabitPeriod
    var completedDates: [String]
    let createdAt: Date
    var archivedAt: Date?  // When habit was archived (nil = active)
    var reminderEnabled: Bool
    var reminderTime: Date?
    var targetValue: Int?
    var unit: String?
    var progressValues: [String: Int]
    var streak: Int  // Calculated on backend

    init(id: UUID = UUID(), goalId: UUID? = nil, title: String, icon: String = "circle.fill", color: String = "green", period: HabitPeriod = .daily, completedDates: [String] = [], createdAt: Date = Date(), archivedAt: Date? = nil, reminderEnabled: Bool = false, reminderTime: Date? = nil, targetValue: Int? = nil, unit: String? = nil, progressValues: [String: Int] = [:], streak: Int = 0) {
        self.id = id
        self.goalId = goalId
        self.title = title
        self.icon = icon
        self.color = color
        self.period = period
        self.completedDates = completedDates
        self.createdAt = createdAt
        self.archivedAt = archivedAt
        self.reminderEnabled = reminderEnabled
        self.reminderTime = reminderTime
        self.targetValue = targetValue
        self.unit = unit
        self.progressValues = progressValues
        self.streak = streak
    }

    /// Whether habit is currently active (not archived)
    var isActive: Bool {
        archivedAt == nil
    }

    /// Whether habit has a measurable goal
    var hasGoal: Bool {
        targetValue != nil && targetValue! > 0
    }

    /// Current progress for today (or current period)
    var currentProgress: Int {
        if hasGoal {
            return progressValues[Self.todayString] ?? 0
        }
        return isCompletedInCurrentPeriod ? 1 : 0
    }

    /// Progress percentage (0.0 to 1.0)
    var progressPercentage: Double {
        guard let target = targetValue, target > 0 else {
            return isCompletedInCurrentPeriod ? 1.0 : 0.0
        }
        return min(Double(currentProgress) / Double(target), 1.0)
    }

    var isCompletedInCurrentPeriod: Bool {
        // For habits with goals, check if progress >= target
        if let target = targetValue, target > 0 {
            let progress = progressValues[Self.todayString] ?? 0
            return progress >= target
        }
        // For simple habits, check completedDates
        switch period {
        case .daily:
            return completedDates.contains(Self.todayString)
        case .weekly:
            return completedDates.contains(where: { Self.isDateInCurrentWeek($0) })
        case .monthly:
            return completedDates.contains(where: { Self.isDateInCurrentMonth($0) })
        }
    }

    static var todayString: String {
        dateString(from: Date())
    }

    static func dateString(from date: Date) -> String {
        DateFormatters.apiDate.string(from: date)
    }

    static func isDateInCurrentWeek(_ dateStr: String) -> Bool {
        guard let date = DateFormatters.apiDate.date(from: dateStr) else { return false }
        return Calendar.current.isDate(date, equalTo: Date(), toGranularity: .weekOfYear)
    }

    static func isDateInCurrentMonth(_ dateStr: String) -> Bool {
        guard let date = DateFormatters.apiDate.date(from: dateStr) else { return false }
        return Calendar.current.isDate(date, equalTo: Date(), toGranularity: .month)
    }

    static func areDatesInSameWeek(_ date1: String, _ date2: String) -> Bool {
        guard let d1 = DateFormatters.apiDate.date(from: date1),
              let d2 = DateFormatters.apiDate.date(from: date2) else { return false }
        return Calendar.current.isDate(d1, equalTo: d2, toGranularity: .weekOfYear)
    }

    static func areDatesInSameMonth(_ date1: String, _ date2: String) -> Bool {
        guard let d1 = DateFormatters.apiDate.date(from: date1),
              let d2 = DateFormatters.apiDate.date(from: date2) else { return false }
        return Calendar.current.isDate(d1, equalTo: d2, toGranularity: .month)
    }

    var swiftUIColor: Color {
        switch color {
        case "green": return .green
        case "blue": return .blue
        case "purple": return .purple
        case "orange": return .orange
        case "pink": return .pink
        case "red": return .red
        default: return .green
        }
    }
}

// MARK: - Task Priority
enum TaskPriority: String, Codable, CaseIterable, Sendable {
    case low = "low"
    case medium = "medium"
    case high = "high"
    case urgent = "urgent"

    var title: LocalizedStringKey {
        switch self {
        case .low: return "Low"
        case .medium: return "Medium"
        case .high: return "High"
        case .urgent: return "Urgent"
        }
    }

    var color: Color {
        switch self {
        case .low: return Color.hf.priorityLow
        case .medium: return Color.hf.priorityMedium
        case .high: return Color.hf.priorityHigh
        case .urgent: return Color.hf.priorityUrgent
        }
    }

    var sortOrder: Int {
        switch self {
        case .urgent: return 0
        case .high: return 1
        case .medium: return 2
        case .low: return 3
        }
    }
}

// MARK: - DailyTask
struct DailyTask: Identifiable, Codable, Sendable {
    let id: UUID
    var title: String
    var isCompleted: Bool
    var priority: TaskPriority
    var dueDate: Date
    let createdAt: Date

    init(id: UUID = UUID(), title: String, isCompleted: Bool = false, priority: TaskPriority = .medium, dueDate: Date = Date(), createdAt: Date = Date()) {
        self.id = id
        self.title = title
        self.isCompleted = isCompleted
        self.priority = priority
        self.dueDate = dueDate
        self.createdAt = createdAt
    }
}

// MARK: - Transaction
struct Transaction: Identifiable, Codable, Sendable {
    let id: UUID
    var title: String
    var amount: Double
    var type: TransactionType
    var category: String
    var date: Date

    init(id: UUID = UUID(), title: String, amount: Double, type: TransactionType, category: String = "", date: Date = Date()) {
        self.id = id
        self.title = title
        self.amount = amount
        self.type = type
        self.category = category
        self.date = date
    }
}

enum TransactionType: String, Codable, CaseIterable, Sendable, Identifiable {
    case income = "income"
    case expense = "expense"

    var id: String { rawValue }
}

// MARK: - Transaction Category

enum TransactionCategory: String, Codable, CaseIterable, Sendable {
    case food = "food"
    case transport = "transport"
    case shopping = "shopping"
    case entertainment = "entertainment"
    case health = "health"
    case education = "education"
    case bills = "bills"
    case salary = "salary"
    case freelance = "freelance"
    case investment = "investment"
    case gift = "gift"
    case other = "other"

    var title: LocalizedStringKey {
        switch self {
        case .food: return "Food & Drinks"
        case .transport: return "Transport"
        case .shopping: return "Shopping"
        case .entertainment: return "Entertainment"
        case .health: return "Health"
        case .education: return "Education"
        case .bills: return "Bills"
        case .salary: return "Salary"
        case .freelance: return "Freelance"
        case .investment: return "Investment"
        case .gift: return "Gift"
        case .other: return "Other"
        }
    }

    var icon: String {
        switch self {
        case .food: return "fork.knife"
        case .transport: return "car.fill"
        case .shopping: return "bag.fill"
        case .entertainment: return "tv.fill"
        case .health: return "heart.fill"
        case .education: return "book.fill"
        case .bills: return "doc.text.fill"
        case .salary: return "banknote.fill"
        case .freelance: return "laptopcomputer"
        case .investment: return "chart.line.uptrend.xyaxis"
        case .gift: return "gift.fill"
        case .other: return "ellipsis.circle.fill"
        }
    }

    var color: Color {
        switch self {
        case .food: return .orange
        case .transport: return .blue
        case .shopping: return .pink
        case .entertainment: return .purple
        case .health: return .red
        case .education: return .cyan
        case .bills: return .gray
        case .salary: return .green
        case .freelance: return .teal
        case .investment: return .indigo
        case .gift: return .yellow
        case .other: return .secondary
        }
    }

    /// Categories for expense type
    static var expenseCategories: [TransactionCategory] {
        [.food, .transport, .shopping, .entertainment, .health, .education, .bills, .gift, .other]
    }

    /// Categories for income type
    static var incomeCategories: [TransactionCategory] {
        [.salary, .freelance, .investment, .gift, .other]
    }
}

// MARK: - Currency
enum Currency: String, Codable, CaseIterable, Sendable {
    case usd = "USD"
    case eur = "EUR"
    case gbp = "GBP"
    case rub = "RUB"
    case kzt = "KZT"
    case cny = "CNY"
    case jpy = "JPY"
    case krw = "KRW"
    case inr = "INR"
    case aed = "AED"

    var symbol: String {
        switch self {
        case .usd: return "$"
        case .eur: return "€"
        case .gbp: return "£"
        case .rub: return "₽"
        case .kzt: return "₸"
        case .cny: return "¥"
        case .jpy: return "¥"
        case .krw: return "₩"
        case .inr: return "₹"
        case .aed: return "د.إ"
        }
    }

    var name: LocalizedStringKey {
        switch self {
        case .usd: return "US Dollar"
        case .eur: return "Euro"
        case .gbp: return "British Pound"
        case .rub: return "Russian Ruble"
        case .kzt: return "Kazakh Tenge"
        case .cny: return "Chinese Yuan"
        case .jpy: return "Japanese Yen"
        case .krw: return "Korean Won"
        case .inr: return "Indian Rupee"
        case .aed: return "UAE Dirham"
        }
    }
}

// MARK: - UserProfile
struct UserProfile: Codable, Sendable {
    var name: String
    var currency: Currency
    var firstOpenDate: Date
    var lastOpenDate: Date

    init(name: String = "", currency: Currency = .usd, firstOpenDate: Date = Date(), lastOpenDate: Date = Date()) {
        self.name = name
        self.currency = currency
        self.firstOpenDate = firstOpenDate
        self.lastOpenDate = lastOpenDate
    }

    var consecutiveDays: Int {
        let calendar = Calendar.current
        let start = calendar.startOfDay(for: firstOpenDate)
        let end = calendar.startOfDay(for: lastOpenDate)
        let components = calendar.dateComponents([.day], from: start, to: end)
        return (components.day ?? 0) + 1
    }
}

// MARK: - Analytics Period
enum AnalyticsPeriod: String, CaseIterable, Sendable {
    case week = "week"
    case month = "month"
    case year = "year"

    var title: LocalizedStringKey {
        switch self {
        case .week: return "Week"
        case .month: return "Month"
        case .year: return "Year"
        }
    }
}

// MARK: - Recurring Transaction

enum RecurringCategory: String, Codable, CaseIterable, Sendable {
    case subscriptions = "subscriptions"
    case utilities = "utilities"
    case loans = "loans"
    case family = "family"
    case other = "other"

    var title: LocalizedStringKey {
        switch self {
        case .subscriptions: return "Subscriptions"
        case .utilities: return "Utilities"
        case .loans: return "Loans"
        case .family: return "Family"
        case .other: return "Other"
        }
    }

    var icon: String {
        switch self {
        case .subscriptions: return "play.rectangle.fill"
        case .utilities: return "bolt.fill"
        case .loans: return "creditcard.fill"
        case .family: return "heart.fill"
        case .other: return "ellipsis.circle.fill"
        }
    }
}

enum RecurrenceFrequency: String, Codable, CaseIterable, Sendable {
    case weekly = "weekly"
    case biweekly = "biweekly"
    case monthly = "monthly"
    case quarterly = "quarterly"
    case yearly = "yearly"

    var title: LocalizedStringKey {
        switch self {
        case .weekly: return "Weekly"
        case .biweekly: return "Biweekly"
        case .monthly: return "Monthly"
        case .quarterly: return "Quarterly"
        case .yearly: return "Yearly"
        }
    }
}

struct RecurringTransaction: Identifiable, Codable, Sendable {
    let id: UUID
    var title: String
    var amount: Double
    var type: TransactionType
    var category: RecurringCategory
    var frequency: RecurrenceFrequency
    var startDate: Date
    var nextDate: Date
    var endDate: Date?
    var remainingPayments: Int?
    var isActive: Bool

    init(id: UUID = UUID(), title: String, amount: Double, type: TransactionType, category: RecurringCategory = .subscriptions, frequency: RecurrenceFrequency = .monthly, startDate: Date = Date(), nextDate: Date = Date(), endDate: Date? = nil, remainingPayments: Int? = nil, isActive: Bool = true) {
        self.id = id
        self.title = title
        self.amount = amount
        self.type = type
        self.category = category
        self.frequency = frequency
        self.startDate = startDate
        self.nextDate = nextDate
        self.endDate = endDate
        self.remainingPayments = remainingPayments
        self.isActive = isActive
    }
}

// MARK: - AI Insights

enum InsightSection: String, Codable, CaseIterable, Sendable {
    case habits
    case tasks
    case budget

    var title: LocalizedStringKey {
        switch self {
        case .habits: return "Habits"
        case .tasks: return "Tasks"
        case .budget: return "Budget"
        }
    }

    var icon: String {
        switch self {
        case .habits: return "repeat"
        case .tasks: return "checkmark.circle"
        case .budget: return "creditcard"
        }
    }

    var requiredDays: Int {
        switch self {
        case .habits: return 14
        case .tasks: return 7
        case .budget: return 30
        }
    }
}

struct InsightStatus: Codable, Sendable {
    var habitsFirstDate: Date?
    var tasksFirstDate: Date?
    var budgetFirstDate: Date?

    var habitsUnlocked: Bool
    var tasksUnlocked: Bool
    var budgetUnlocked: Bool

    var habitsUnlockCelebrated: Bool
    var tasksUnlockCelebrated: Bool
    var budgetUnlockCelebrated: Bool

    init() {
        self.habitsFirstDate = nil
        self.tasksFirstDate = nil
        self.budgetFirstDate = nil
        self.habitsUnlocked = false
        self.tasksUnlocked = false
        self.budgetUnlocked = false
        self.habitsUnlockCelebrated = false
        self.tasksUnlockCelebrated = false
        self.budgetUnlockCelebrated = false
    }

    // Calculate progress (0.0 to 1.0)
    func progress(for section: InsightSection) -> Double {
        let firstDate: Date?
        let requiredDays = section.requiredDays

        switch section {
        case .habits: firstDate = habitsFirstDate
        case .tasks: firstDate = tasksFirstDate
        case .budget: firstDate = budgetFirstDate
        }

        guard let startDate = firstDate else { return 0 }

        let daysPassed = Calendar.current.dateComponents([.day], from: startDate, to: Date()).day ?? 0
        return min(1.0, Double(daysPassed) / Double(requiredDays))
    }

    // Days remaining until unlock
    func daysRemaining(for section: InsightSection) -> Int {
        let firstDate: Date?
        let requiredDays = section.requiredDays

        switch section {
        case .habits: firstDate = habitsFirstDate
        case .tasks: firstDate = tasksFirstDate
        case .budget: firstDate = budgetFirstDate
        }

        guard let startDate = firstDate else { return requiredDays }

        let daysPassed = Calendar.current.dateComponents([.day], from: startDate, to: Date()).day ?? 0
        return max(0, requiredDays - daysPassed)
    }

    // Check if section is unlocked
    func isUnlocked(for section: InsightSection) -> Bool {
        switch section {
        case .habits: return habitsUnlocked
        case .tasks: return tasksUnlocked
        case .budget: return budgetUnlocked
        }
    }

    // Check if unlock was celebrated
    func isCelebrated(for section: InsightSection) -> Bool {
        switch section {
        case .habits: return habitsUnlockCelebrated
        case .tasks: return tasksUnlockCelebrated
        case .budget: return budgetUnlockCelebrated
        }
    }

    // Mutating functions to update status
    mutating func setFirstDate(for section: InsightSection, date: Date) {
        switch section {
        case .habits:
            if habitsFirstDate == nil { habitsFirstDate = date }
        case .tasks:
            if tasksFirstDate == nil { tasksFirstDate = date }
        case .budget:
            if budgetFirstDate == nil { budgetFirstDate = date }
        }
    }

    mutating func setUnlocked(for section: InsightSection) {
        switch section {
        case .habits: habitsUnlocked = true
        case .tasks: tasksUnlocked = true
        case .budget: budgetUnlocked = true
        }
    }

    mutating func setCelebrated(for section: InsightSection) {
        switch section {
        case .habits: habitsUnlockCelebrated = true
        case .tasks: tasksUnlockCelebrated = true
        case .budget: budgetUnlockCelebrated = true
        }
    }
}

// MARK: - Insight

enum InsightType: String, Codable, Sendable {
    case pattern      // "You skip gym on Fridays"
    case achievement  // "7 day streak!"
    case warning      // "Spending up 20%"
    case suggestion   // "Try morning meditation"

    var icon: String {
        switch self {
        case .pattern: return "chart.line.uptrend.xyaxis"
        case .achievement: return "trophy.fill"
        case .warning: return "exclamationmark.triangle.fill"
        case .suggestion: return "lightbulb.fill"
        }
    }

    var color: String {
        switch self {
        case .pattern: return "info"
        case .achievement: return "accent"
        case .warning: return "warning"
        case .suggestion: return "accent"
        }
    }
}

struct InsightAction: Codable, Sendable {
    let label: String
    let actionType: String  // "setReminder", "openAIChat", "dismiss"
    let payload: [String: String]

    init(label: String, actionType: String, payload: [String: String] = [:]) {
        self.label = label
        self.actionType = actionType
        self.payload = payload
    }
}

struct Insight: Identifiable, Codable, Sendable {
    let id: UUID
    let section: InsightSection
    let type: InsightType
    let title: String
    let message: String
    let action: InsightAction?
    let createdAt: Date
    var isDismissed: Bool

    /// Unique hash based on content (title + message) for tracking dismissals
    /// Uses SHA256 for deterministic hashing across app sessions
    var contentHash: String {
        let input = "\(section.rawValue)_\(title)_\(message)"
        let data = Data(input.utf8)
        let hash = SHA256.hash(data: data)
        return hash.compactMap { String(format: "%02x", $0) }.joined().prefix(16).description
    }

    init(
        id: UUID = UUID(),
        section: InsightSection,
        type: InsightType,
        title: String,
        message: String,
        action: InsightAction? = nil,
        createdAt: Date = Date(),
        isDismissed: Bool = false
    ) {
        self.id = id
        self.section = section
        self.type = type
        self.title = title
        self.message = message
        self.action = action
        self.createdAt = createdAt
        self.isDismissed = isDismissed
    }
}

// MARK: - Weekly Review

struct WeeklyReview: Identifiable, Codable, Sendable {
    let id: UUID
    let weekStart: Date
    let weekEnd: Date
    let createdAt: Date

    // Habits
    let habitsCompletionRate: Double
    let habitsCompletionRateChange: Double  // vs previous week
    let totalHabitsCompleted: Int
    let bestHabit: String?
    let habitStreak: Int?

    // Tasks
    let tasksCompletionRate: Double
    let tasksCompletionRateChange: Double
    let totalTasksCompleted: Int
    let totalTasksCreated: Int

    // Budget
    let totalIncome: Double
    let totalExpenses: Double
    let expensesChange: Double  // vs previous week
    let topCategory: String?
    let topCategoryAmount: Double?

    // Highlights
    let wins: [String]
    let warnings: [String]

    init(
        id: UUID = UUID(),
        weekStart: Date,
        weekEnd: Date,
        createdAt: Date = Date(),
        habitsCompletionRate: Double = 0,
        habitsCompletionRateChange: Double = 0,
        totalHabitsCompleted: Int = 0,
        bestHabit: String? = nil,
        habitStreak: Int? = nil,
        tasksCompletionRate: Double = 0,
        tasksCompletionRateChange: Double = 0,
        totalTasksCompleted: Int = 0,
        totalTasksCreated: Int = 0,
        totalIncome: Double = 0,
        totalExpenses: Double = 0,
        expensesChange: Double = 0,
        topCategory: String? = nil,
        topCategoryAmount: Double? = nil,
        wins: [String] = [],
        warnings: [String] = []
    ) {
        self.id = id
        self.weekStart = weekStart
        self.weekEnd = weekEnd
        self.createdAt = createdAt
        self.habitsCompletionRate = habitsCompletionRate
        self.habitsCompletionRateChange = habitsCompletionRateChange
        self.totalHabitsCompleted = totalHabitsCompleted
        self.bestHabit = bestHabit
        self.habitStreak = habitStreak
        self.tasksCompletionRate = tasksCompletionRate
        self.tasksCompletionRateChange = tasksCompletionRateChange
        self.totalTasksCompleted = totalTasksCompleted
        self.totalTasksCreated = totalTasksCreated
        self.totalIncome = totalIncome
        self.totalExpenses = totalExpenses
        self.expensesChange = expensesChange
        self.topCategory = topCategory
        self.topCategoryAmount = topCategoryAmount
        self.wins = wins
        self.warnings = warnings
    }

    var weekDateRange: String {
        let formatter = DateFormatter()
        formatter.dateFormat = "MMM d"
        return "\(formatter.string(from: weekStart)) - \(formatter.string(from: weekEnd))"
    }

    var savingsRate: Double {
        guard totalIncome > 0 else { return 0 }
        return (totalIncome - totalExpenses) / totalIncome
    }
}

// MARK: - Budget Forecast

struct BudgetForecast: Identifiable, Codable, Sendable {
    let id: UUID
    let forecastMonth: Date
    let generatedAt: Date
    let categoryForecasts: [CategoryForecast]
    let projectedExpenses: Double
    let projectedIncome: Double
    let projectedSavings: Double
    let expenseTrend: TrendDirection
    let confidenceScore: Double
    let seasonalFactors: [SeasonalFactor]?

    init(
        id: UUID = UUID(),
        forecastMonth: Date,
        generatedAt: Date = Date(),
        categoryForecasts: [CategoryForecast] = [],
        projectedExpenses: Double = 0,
        projectedIncome: Double = 0,
        projectedSavings: Double = 0,
        expenseTrend: TrendDirection = .stable,
        confidenceScore: Double = 0,
        seasonalFactors: [SeasonalFactor]? = nil
    ) {
        self.id = id
        self.forecastMonth = forecastMonth
        self.generatedAt = generatedAt
        self.categoryForecasts = categoryForecasts
        self.projectedExpenses = projectedExpenses
        self.projectedIncome = projectedIncome
        self.projectedSavings = projectedSavings
        self.expenseTrend = expenseTrend
        self.confidenceScore = confidenceScore
        self.seasonalFactors = seasonalFactors
    }
}

struct CategoryForecast: Identifiable, Codable, Sendable {
    let id: UUID
    let category: String
    let projectedAmount: Double
    let historicalAverage: Double
    let changePercent: Double
    let trend: TrendDirection
    let recurringAmount: Double
    let variableAmount: Double

    init(
        id: UUID = UUID(),
        category: String,
        projectedAmount: Double,
        historicalAverage: Double,
        changePercent: Double = 0,
        trend: TrendDirection = .stable,
        recurringAmount: Double = 0,
        variableAmount: Double = 0
    ) {
        self.id = id
        self.category = category
        self.projectedAmount = projectedAmount
        self.historicalAverage = historicalAverage
        self.changePercent = changePercent
        self.trend = trend
        self.recurringAmount = recurringAmount
        self.variableAmount = variableAmount
    }
}

enum TrendDirection: String, Codable, Sendable {
    case up = "up"
    case down = "down"
    case stable = "stable"

    var icon: String {
        switch self {
        case .up: return "arrow.up.right"
        case .down: return "arrow.down.right"
        case .stable: return "arrow.right"
        }
    }

    var color: Color {
        switch self {
        case .up: return Color.hf.expense
        case .down: return Color.hf.income
        case .stable: return .secondary
        }
    }
}

struct SeasonalFactor: Codable, Sendable {
    let category: String
    let monthlyMultiplier: Double
    let reason: String
}

// MARK: - Savings Goal

struct SavingsGoal: Identifiable, Codable, Sendable {
    let id: UUID
    var monthlyTarget: Double
    var currentSavings: Double
    var monthlyIncome: Double
    var monthlyExpenses: Double
    var progress: Double  // Calculated on backend

    init(
        id: UUID = UUID(),
        monthlyTarget: Double,
        currentSavings: Double = 0,
        monthlyIncome: Double = 0,
        monthlyExpenses: Double = 0,
        progress: Double = 0
    ) {
        self.id = id
        self.monthlyTarget = monthlyTarget
        self.currentSavings = currentSavings
        self.monthlyIncome = monthlyIncome
        self.monthlyExpenses = monthlyExpenses
        self.progress = progress
    }
}

// MARK: - AI Expense Analysis

struct AIExpenseAnalysis: Identifiable, Codable, Sendable {
    let id: UUID
    let generatedAt: Date
    let insights: [ExpenseInsight]
    let questionableTransactions: [QuestionableTransaction]
    let savingsSuggestions: [SavingsSuggestion]

    init(
        id: UUID = UUID(),
        generatedAt: Date = Date(),
        insights: [ExpenseInsight] = [],
        questionableTransactions: [QuestionableTransaction] = [],
        savingsSuggestions: [SavingsSuggestion] = []
    ) {
        self.id = id
        self.generatedAt = generatedAt
        self.insights = insights
        self.questionableTransactions = questionableTransactions
        self.savingsSuggestions = savingsSuggestions
    }
}

struct ExpenseInsight: Identifiable, Codable, Sendable {
    let id: UUID
    let type: ExpenseInsightType
    let title: String
    let message: String
    let amount: Double?
    let category: String?
    let priority: Int

    init(
        id: UUID = UUID(),
        type: ExpenseInsightType,
        title: String,
        message: String,
        amount: Double? = nil,
        category: String? = nil,
        priority: Int = 2
    ) {
        self.id = id
        self.type = type
        self.title = title
        self.message = message
        self.amount = amount
        self.category = category
        self.priority = priority
    }
}

enum ExpenseInsightType: String, Codable, Sendable {
    case pattern = "pattern"
    case habit = "habit"
    case impulse = "impulse"
    case subscription = "subscription"
    case opportunity = "opportunity"

    var icon: String {
        switch self {
        case .pattern: return "chart.line.uptrend.xyaxis"
        case .habit: return "repeat"
        case .impulse: return "bolt.fill"
        case .subscription: return "creditcard.fill"
        case .opportunity: return "lightbulb.fill"
        }
    }

    var color: Color {
        switch self {
        case .pattern: return Color.hf.info
        case .habit: return Color.hf.warning
        case .impulse: return Color.hf.expense
        case .subscription: return .purple
        case .opportunity: return Color.hf.accent
        }
    }
}

struct QuestionableTransaction: Identifiable, Codable, Sendable {
    let id: UUID
    let transactionId: UUID
    let reason: String
    let category: QuestionableCategory
    let potentialSavings: Double?

    init(
        id: UUID = UUID(),
        transactionId: UUID,
        reason: String,
        category: QuestionableCategory,
        potentialSavings: Double? = nil
    ) {
        self.id = id
        self.transactionId = transactionId
        self.reason = reason
        self.category = category
        self.potentialSavings = potentialSavings
    }
}

enum QuestionableCategory: String, Codable, Sendable {
    case impulse = "impulse"
    case duplicate = "duplicate"
    case excessive = "excessive"
    case unnecessary = "unnecessary"

    var label: String {
        switch self {
        case .impulse: return "Impulse"
        case .duplicate: return "Duplicate"
        case .excessive: return "Excessive"
        case .unnecessary: return "Unnecessary"
        }
    }

    var color: Color {
        switch self {
        case .impulse: return Color.hf.warning
        case .duplicate: return Color.hf.expense
        case .excessive: return .orange
        case .unnecessary: return .purple
        }
    }
}

struct SavingsSuggestion: Identifiable, Codable, Sendable {
    let id: UUID
    let category: String
    let currentSpending: Double
    let suggestedBudget: Double
    let potentialSavings: Double
    let reason: String
    let difficulty: SuggestionDifficulty

    init(
        id: UUID = UUID(),
        category: String,
        currentSpending: Double,
        suggestedBudget: Double,
        potentialSavings: Double,
        reason: String,
        difficulty: SuggestionDifficulty
    ) {
        self.id = id
        self.category = category
        self.currentSpending = currentSpending
        self.suggestedBudget = suggestedBudget
        self.potentialSavings = potentialSavings
        self.reason = reason
        self.difficulty = difficulty
    }
}

enum SuggestionDifficulty: String, Codable, Sendable {
    case easy = "easy"
    case medium = "medium"
    case hard = "hard"

    var label: String {
        switch self {
        case .easy: return "Easy"
        case .medium: return "Moderate"
        case .hard: return "Hard"
        }
    }

    var color: Color {
        switch self {
        case .easy: return Color.hf.income
        case .medium: return Color.hf.warning
        case .hard: return Color.hf.expense
        }
    }
}

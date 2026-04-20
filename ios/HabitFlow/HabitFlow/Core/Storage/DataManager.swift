import Foundation
import SwiftUI
import Combine
import WidgetKit

@MainActor
class DataManager: ObservableObject {
    static let shared = DataManager()

    // MARK: - Feature Managers
    let goalsManager = GoalsManager()
    let habitsManager = HabitsManager()
    let tasksManager = TasksManager()
    let budgetManager = BudgetManager()
    let insightManager = InsightManager()
    let analyticsManager = AnalyticsManager()

    // MARK: - Shared State
    @Published var profile: UserProfile = UserProfile()
    @Published var isLoading = false
    @Published var syncError: String?
    @Published var isDemoMode = false
    @Published var daysWithAtoma: Int = 0

    // MARK: - Storage Keys
    private let profileKey = "profile_v2"
    private let activeDaysKey = "active_days"
    private var cancellables = Set<AnyCancellable>()

    // MARK: - Backward Compatibility (Forwarding Properties)
    var goals: [Goal] {
        get { goalsManager.items }
        set { goalsManager.items = newValue }
    }

    var habits: [Habit] {
        get { habitsManager.items }
        set { habitsManager.items = newValue }
    }

    var tasks: [DailyTask] {
        get { tasksManager.items }
        set { tasksManager.items = newValue }
    }

    var transactions: [Transaction] {
        get { budgetManager.transactions }
        set { budgetManager.transactions = newValue }
    }

    var recurringTransactions: [RecurringTransaction] {
        get { budgetManager.recurringTransactions }
        set { budgetManager.recurringTransactions = newValue }
    }

    var insights: [Insight] {
        get { insightManager.insights }
        set { insightManager.insights = newValue }
    }

    var insightStatus: InsightStatus {
        get { insightManager.status }
        set { insightManager.status = newValue }
    }

    var currentForecast: BudgetForecast? {
        get { analyticsManager.currentForecast }
        set { analyticsManager.currentForecast = newValue }
    }

    var savingsGoal: SavingsGoal? {
        get { budgetManager.savingsGoal }
        set { budgetManager.savingsGoal = newValue }
    }

    var aiExpenseAnalysis: AIExpenseAnalysis? {
        get { analyticsManager.aiExpenseAnalysis }
        set { analyticsManager.aiExpenseAnalysis = newValue }
    }

    // Computed forwarding
    var activeGoals: [Goal] { goalsManager.active }
    var archivedGoals: [Goal] { goalsManager.archived }
    var sortedTasks: [DailyTask] { tasksManager.sorted }
    var balance: Double { budgetManager.balance }
    var monthlyIncome: Double { budgetManager.monthlyIncome }
    var monthlyExpenses: Double { budgetManager.monthlyExpenses }
    var projectedMonthlyExpenses: Double { budgetManager.projectedMonthlyExpenses }
    var projectedMonthlyIncome: Double { budgetManager.projectedMonthlyIncome }

    // MARK: - Init
    private init() {
        setupManagerCoordinators()
        load()
        updateStreak()
        insightManager.checkUnlocks()
    }

    private func setupManagerCoordinators() {
        goalsManager.coordinator = self
        habitsManager.coordinator = self
        tasksManager.coordinator = self
        budgetManager.coordinator = self
        insightManager.coordinator = self
        analyticsManager.coordinator = self

        // Forward objectWillChange from managers to DataManager
        goalsManager.objectWillChange.sink { [weak self] _ in
            self?.objectWillChange.send()
        }.store(in: &cancellables)

        habitsManager.objectWillChange.sink { [weak self] _ in
            self?.objectWillChange.send()
        }.store(in: &cancellables)

        tasksManager.objectWillChange.sink { [weak self] _ in
            self?.objectWillChange.send()
        }.store(in: &cancellables)

        budgetManager.objectWillChange.sink { [weak self] _ in
            self?.objectWillChange.send()
        }.store(in: &cancellables)

        insightManager.objectWillChange.sink { [weak self] _ in
            self?.objectWillChange.send()
        }.store(in: &cancellables)

        analyticsManager.objectWillChange.sink { [weak self] _ in
            self?.objectWillChange.send()
        }.store(in: &cancellables)
    }

    // MARK: - Load
    private func load() {
        // Load profile
        if let data = UserDefaults.standard.data(forKey: profileKey),
           let decoded = try? JSONDecoder().decode(UserProfile.self, from: data) {
            profile = decoded
        }

        // Load days with Atoma
        loadDaysWithAtoma()

        // Load all managers
        goalsManager.load()
        habitsManager.load()
        tasksManager.load()
        budgetManager.load()
        insightManager.load()
        analyticsManager.load()
    }

    private func loadDaysWithAtoma() {
        if let activeDaysData = UserDefaults.standard.data(forKey: activeDaysKey),
           let activeDays = try? JSONDecoder().decode(Set<String>.self, from: activeDaysData) {
            daysWithAtoma = activeDays.count
        } else {
            daysWithAtoma = 0
        }
    }

    // MARK: - Save
    private func saveProfile() {
        if let data = try? JSONEncoder().encode(profile) {
            UserDefaults.standard.set(data, forKey: profileKey)
        }
    }

    // MARK: - Sync
    func syncAll() async {
        guard !isDemoMode else { return }
        isLoading = true
        syncError = nil

        // First: process recurring transactions
        await budgetManager.syncRecurring()

        // Then: run other syncs in parallel
        async let goalsSync: () = goalsManager.sync()
        async let habitsSync: () = habitsManager.sync()
        async let tasksSync: () = tasksManager.sync()
        async let transactionsSync: () = budgetManager.syncTransactions()

        await goalsSync
        await habitsSync
        await tasksSync
        await transactionsSync

        isLoading = false
    }

    func syncGoals() async {
        isLoading = true
        syncError = nil
        await goalsManager.sync()
        isLoading = false
    }

    func syncHabits() async {
        isLoading = true
        syncError = nil
        await habitsManager.sync()
        isLoading = false
    }

    func syncTasks() async {
        isLoading = true
        syncError = nil
        await tasksManager.sync()
        isLoading = false
    }

    func syncTransactions() async {
        isLoading = true
        syncError = nil
        await budgetManager.syncTransactions()
        isLoading = false
    }

    func syncRecurringTransactions() async {
        isLoading = true
        syncError = nil
        await budgetManager.syncRecurring()
        isLoading = false
    }

    // MARK: - Activity Tracking
    func recordActivity() {
        let today = DateFormatters.apiDate.string(from: Date())
        var activeDays: Set<String> = []

        if let data = UserDefaults.standard.data(forKey: activeDaysKey),
           let decoded = try? JSONDecoder().decode(Set<String>.self, from: data) {
            activeDays = decoded
        }

        if !activeDays.contains(today) {
            activeDays.insert(today)
            if let encoded = try? JSONEncoder().encode(activeDays) {
                UserDefaults.standard.set(encoded, forKey: activeDaysKey)
            }
            daysWithAtoma = activeDays.count
        }
    }

    private func updateStreak() {
        profile.lastOpenDate = Date()
        saveProfile()
    }

    // MARK: - Widget Data
    private let appGroupIdentifier = "group.com.azamatbigali.habitflow"

    func updateWidgetData() {
        guard let userDefaults = UserDefaults(suiteName: appGroupIdentifier) else { return }

        let widgetHabits = habits.map { habit in
            WidgetHabitData(
                id: habit.id,
                title: habit.title,
                icon: habit.icon,
                color: habit.color,
                isCompleted: habit.isCompletedInCurrentPeriod,
                streak: habit.streak
            )
        }

        let widgetTasks = tasks.map { task in
            WidgetTaskData(
                id: task.id,
                title: task.title,
                isCompleted: task.isCompleted,
                priority: task.priority.rawValue
            )
        }

        let income = transactions.filter { $0.type == .income }.reduce(0) { $0 + $1.amount }
        let expenses = transactions.filter { $0.type == .expense }.reduce(0) { $0 + $1.amount }

        let widgetBudget = WidgetBudgetData(
            balance: income - expenses,
            income: income,
            expenses: expenses,
            currency: profile.currency.rawValue,
            currencySymbol: profile.currency.symbol
        )

        let widgetData = WidgetDataPayload(
            habits: widgetHabits,
            tasks: widgetTasks,
            budget: widgetBudget,
            lastUpdated: Date()
        )

        if let encoded = try? JSONEncoder().encode(widgetData) {
            userDefaults.set(encoded, forKey: "widgetData")
        }

        WidgetCenter.shared.reloadAllTimelines()
    }

    // MARK: - Insight Forwarding
    func trackInsightFirstDate(for section: InsightSection) {
        insightManager.trackFirstDate(for: section)
    }

    func markInsightCelebrated(for section: InsightSection) {
        insightManager.markCelebrated(for: section)
    }

    func generateInsights(for section: InsightSection) {
        insightManager.generate(for: section)
    }

    func hasEnoughData(for section: InsightSection) -> Bool {
        insightManager.hasEnoughData(for: section)
    }

    func refreshAllInsights() {
        insightManager.refreshAll()
    }

    func insights(for section: InsightSection) -> [Insight] {
        insightManager.insights(for: section)
    }

    func dismissInsight(_ insight: Insight) {
        insightManager.dismiss(insight)
    }

    // MARK: - Goals Forwarding
    func addGoal(_ goal: Goal) { goalsManager.add(goal) }
    func updateGoal(_ goal: Goal) { goalsManager.update(goal) }
    func deleteGoal(_ goal: Goal) { goalsManager.delete(goal) }
    func archiveGoal(_ goal: Goal) { goalsManager.archive(goal) }
    func unarchiveGoal(_ goal: Goal) { goalsManager.unarchive(goal) }
    func habitsForGoal(_ goal: Goal) -> [Habit] { goalsManager.habitsForGoal(goal) }
    func habitsWithoutGoal() -> [Habit] { goalsManager.habitsWithoutGoal() }

    // Called by GoalsManager when deleting a goal
    func clearHabitGoalLinks(goalId: UUID) {
        habitsManager.clearGoalLinks(goalId: goalId)
    }

    // MARK: - Habits Forwarding
    func addHabit(_ habit: Habit) { habitsManager.add(habit) }
    func updateHabit(_ habit: Habit) { habitsManager.update(habit) }
    func deleteHabit(_ habit: Habit) { habitsManager.delete(habit) }
    func toggleHabit(_ habit: Habit) { habitsManager.toggle(habit) }
    func toggleHabitForDate(_ habit: Habit, date: Date) { habitsManager.toggleForDate(habit, date: date) }
    func incrementHabitProgress(_ habit: Habit) { habitsManager.incrementProgress(habit) }
    func incrementHabitProgressForDate(_ habit: Habit, date: Date) { habitsManager.incrementProgressForDate(habit, date: date) }
    func setHabitProgress(_ habit: Habit, value: Int, date: Date = Date()) { habitsManager.setProgress(habit, value: value, date: date) }
    func archiveHabit(_ habit: Habit) { habitsManager.archive(habit) }
    func unarchiveHabit(_ habit: Habit) { habitsManager.unarchive(habit) }
    func updateHabitReminder(_ habit: Habit) { habitsManager.updateReminder(habit) }

    // MARK: - Tasks Forwarding
    func addTask(_ task: DailyTask) { tasksManager.add(task) }
    func updateTask(_ task: DailyTask) { tasksManager.update(task) }
    func deleteTask(_ task: DailyTask) { tasksManager.delete(task) }
    func toggleTask(_ task: DailyTask) { tasksManager.toggle(task) }
    func tasksForDate(_ date: Date) -> [DailyTask] { tasksManager.tasksForDate(date) }

    // MARK: - Budget Forwarding
    func addTransaction(_ transaction: Transaction) { budgetManager.addTransaction(transaction) }
    func updateTransaction(_ transaction: Transaction) { budgetManager.updateTransaction(transaction) }
    func deleteTransaction(_ transaction: Transaction) { budgetManager.deleteTransaction(transaction) }
    func addRecurringTransaction(_ recurring: RecurringTransaction) { budgetManager.addRecurring(recurring) }
    func updateRecurringTransaction(_ recurring: RecurringTransaction) { budgetManager.updateRecurring(recurring) }
    func deleteRecurringTransaction(_ recurring: RecurringTransaction) { budgetManager.deleteRecurring(recurring) }
    func transactionsForDate(_ date: Date) -> [Transaction] { budgetManager.transactionsForDate(date) }
    func transactionsForPeriod(_ period: AnalyticsPeriod) -> [Transaction] { budgetManager.transactionsForPeriod(period) }
    func incomeForPeriod(_ period: AnalyticsPeriod) -> Double { budgetManager.incomeForPeriod(period) }
    func expensesForPeriod(_ period: AnalyticsPeriod) -> Double { budgetManager.expensesForPeriod(period) }
    func balanceForPeriod(_ period: AnalyticsPeriod) -> Double { budgetManager.balanceForPeriod(period) }
    func balanceForDate(_ date: Date) -> Double { budgetManager.balanceForDate(date) }

    // Savings Goal
    func setSavingsGoal(_ target: Double) { budgetManager.setSavingsGoal(target) }
    func syncSavingsGoal() async { await budgetManager.syncSavingsGoal() }
    func deleteSavingsGoal() { budgetManager.deleteSavingsGoal() }

    // MARK: - Analytics Forwarding
    func habitsCompletionRate(for period: AnalyticsPeriod) -> Double { habitsManager.completionRate(for: period) }
    func tasksCompletionRate() -> Double { tasksManager.completionRate() }
    func budgetHealthScore(for period: AnalyticsPeriod) -> Double { analyticsManager.budgetHealthScore(for: period) }
    func lifeScore(for period: AnalyticsPeriod = .week) -> Double { analyticsManager.lifeScore(for: period) }
    func lifeScoreComponents(for period: AnalyticsPeriod) -> (habits: Double, tasks: Double, budget: Double) { analyticsManager.lifeScoreComponents(for: period) }
    func lifeScoreHistory(for period: AnalyticsPeriod) -> [LifeScoreDataPoint] { analyticsManager.lifeScoreHistory(for: period) }
    func spendingByCategory(for period: AnalyticsPeriod) -> [CategorySpending] { analyticsManager.spendingByCategory(for: period) }
    func dailyExpensesForWeek() -> [(day: String, amount: Double)] { analyticsManager.dailyExpensesForWeek() }
    func habitCompletionsForWeek(_ habit: Habit) -> [(day: String, completed: Bool)] { habitsManager.completionsForWeek(habit) }
    func generateBudgetForecast() -> BudgetForecast { analyticsManager.generateBudgetForecast() }
    func formatCurrency(_ amount: Double) -> String { analyticsManager.formatCurrency(amount) }

    func saveAIAnalysis() {
        analyticsManager.saveAIAnalysis()
    }

    // MARK: - Profile
    func updateName(_ name: String) {
        profile.name = name
        saveProfile()
    }

    func updateCurrency(_ currency: Currency) {
        profile.currency = currency
        saveProfile()
    }

    // MARK: - Clear All Data
    func clearAllData() {
        let savedCurrency = profile.currency

        goalsManager.clear()
        habitsManager.clear()
        tasksManager.clear()
        budgetManager.clear()
        insightManager.clear()
        analyticsManager.clear()

        profile = UserProfile()
        UserDefaults.standard.removeObject(forKey: profileKey)

        // Restore currency preference
        profile.currency = savedCurrency
        saveProfile()
    }

    // MARK: - Demo Data
    func loadDemoData() {
        isDemoMode = true
        let calendar = Calendar.current
        let today = Date()

        func dateString(_ daysAgo: Int) -> String {
            let date = calendar.date(byAdding: .day, value: -daysAgo, to: today)!
            return DateFormatters.apiDate.string(from: date)
        }

        func date(_ daysAgo: Int) -> Date {
            calendar.date(byAdding: .day, value: -daysAgo, to: today)!
        }

        // Demo Habits
        let meditationDates = (0..<30).filter { $0 != 3 && $0 != 7 && $0 != 15 && $0 != 22 && $0 != 28 }.map { dateString($0) }
        let meditation = Habit(title: "Meditation", icon: "🧘", color: "purple", period: .daily, completedDates: meditationDates, createdAt: date(30))

        let exerciseDates = (0..<30).filter { $0 % 3 != 2 || $0 < 5 }.prefix(20).map { dateString($0) }
        let exercise = Habit(title: "Exercise", icon: "🏃", color: "green", period: .daily, completedDates: Array(exerciseDates), createdAt: date(30))

        let readingDates = (0..<30).filter { $0 % 5 != 4 }.prefix(18).map { dateString($0) }
        let reading = Habit(title: "Read 30 min", icon: "📚", color: "blue", period: .daily, completedDates: Array(readingDates), createdAt: date(30))

        let waterDates = [dateString(0), dateString(1), dateString(2), dateString(4), dateString(5)]
        let water = Habit(title: "Drink 8 glasses", icon: "💧", color: "blue", period: .daily, completedDates: waterDates, createdAt: date(14))

        let mealPrepDates = [dateString(0), dateString(7), dateString(21)]
        let mealPrep = Habit(title: "Meal prep", icon: "🍳", color: "orange", period: .weekly, completedDates: mealPrepDates, createdAt: date(28))

        habitsManager.items = [meditation, exercise, reading, water, mealPrep]
        habitsManager.save()

        // Demo Tasks
        var demoTasks: [DailyTask] = []
        demoTasks.append(DailyTask(title: "Review project proposal", isCompleted: true, priority: .high, dueDate: today))
        demoTasks.append(DailyTask(title: "Call with client", isCompleted: true, priority: .high, dueDate: today))
        demoTasks.append(DailyTask(title: "Update documentation", isCompleted: false, priority: .medium, dueDate: today))
        demoTasks.append(DailyTask(title: "Send weekly report", isCompleted: true, priority: .high, dueDate: date(1)))
        demoTasks.append(DailyTask(title: "Code review", isCompleted: true, priority: .medium, dueDate: date(1)))
        demoTasks.append(DailyTask(title: "Fix login bug", isCompleted: true, priority: .high, dueDate: date(2)))
        demoTasks.append(DailyTask(title: "Update tests", isCompleted: true, priority: .medium, dueDate: date(2)))
        demoTasks.append(DailyTask(title: "Deploy to staging", isCompleted: true, priority: .high, dueDate: date(3)))
        demoTasks.append(DailyTask(title: "Team meeting prep", isCompleted: true, priority: .medium, dueDate: date(4)))
        demoTasks.append(DailyTask(title: "Review PRs", isCompleted: true, priority: .medium, dueDate: date(4)))
        demoTasks.append(DailyTask(title: "Write API docs", isCompleted: true, priority: .low, dueDate: date(5)))
        demoTasks.append(DailyTask(title: "Refactor auth module", isCompleted: true, priority: .medium, dueDate: date(6)))
        demoTasks.append(DailyTask(title: "Sprint planning", isCompleted: true, priority: .high, dueDate: date(7)))
        demoTasks.append(DailyTask(title: "Design review", isCompleted: true, priority: .medium, dueDate: date(10)))
        demoTasks.append(DailyTask(title: "Quarterly goals", isCompleted: false, priority: .low, dueDate: date(14)))

        tasksManager.items = demoTasks
        tasksManager.save()

        // Demo Transactions
        var demoTransactions: [Transaction] = []
        let salaryDate = calendar.date(from: calendar.dateComponents([.year, .month], from: today))!
        demoTransactions.append(Transaction(title: "Salary", amount: 5000, type: .income, category: "Salary", date: salaryDate))
        demoTransactions.append(Transaction(title: "Freelance project", amount: 800, type: .income, category: "Freelance", date: date(10)))

        let expenseCategories: [(String, String, Double, Int)] = [
            ("Groceries", "Food", 45, 3), ("Coffee", "Food", 5, 1), ("Lunch", "Food", 15, 2),
            ("Gas", "Transport", 50, 7), ("Uber", "Transport", 25, 5),
            ("Netflix", "Entertainment", 15, 30), ("Spotify", "Entertainment", 10, 30),
            ("Gym", "Health", 50, 30), ("Pharmacy", "Health", 30, 14),
            ("Electric bill", "Utilities", 120, 30), ("Internet", "Utilities", 60, 30),
            ("Dinner out", "Food", 65, 7), ("Shopping", "Shopping", 150, 10), ("Books", "Education", 25, 14)
        ]

        for (title, category, amount, frequency) in expenseCategories {
            for day in stride(from: 0, to: 30, by: frequency) {
                if day == 0 || (day > 0 && Bool.random()) {
                    demoTransactions.append(Transaction(
                        title: title,
                        amount: amount * Double.random(in: 0.8...1.2),
                        type: .expense,
                        category: category,
                        date: date(day)
                    ))
                }
            }
        }

        demoTransactions.append(Transaction(title: "New headphones", amount: 199, type: .expense, category: "Shopping", date: date(5)))
        demoTransactions.append(Transaction(title: "Birthday gift", amount: 75, type: .expense, category: "Gifts", date: date(12)))
        demoTransactions.append(Transaction(title: "Doctor visit", amount: 150, type: .expense, category: "Health", date: date(20)))

        budgetManager.transactions = demoTransactions
        budgetManager.saveTransactions()

        // Demo Recurring
        budgetManager.recurringTransactions = [
            RecurringTransaction(title: "Netflix", amount: 15.99, type: .expense, category: .subscriptions, frequency: .monthly, startDate: date(60), nextDate: calendar.date(byAdding: .month, value: 1, to: salaryDate)!, isActive: true),
            RecurringTransaction(title: "Spotify", amount: 9.99, type: .expense, category: .subscriptions, frequency: .monthly, startDate: date(90), nextDate: calendar.date(byAdding: .month, value: 1, to: salaryDate)!, isActive: true),
            RecurringTransaction(title: "Gym membership", amount: 49.99, type: .expense, category: .subscriptions, frequency: .monthly, startDate: date(120), nextDate: calendar.date(byAdding: .month, value: 1, to: salaryDate)!, isActive: true)
        ]
        budgetManager.saveRecurring()

        // Update insight tracking
        insightManager.status.habitsFirstDate = date(30)
        insightManager.status.tasksFirstDate = date(30)
        insightManager.status.budgetFirstDate = date(30)
        insightManager.saveStatus()
    }
}

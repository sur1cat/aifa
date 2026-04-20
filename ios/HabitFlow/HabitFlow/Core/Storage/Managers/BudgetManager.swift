import Foundation
import Combine
import os

@MainActor
class BudgetManager: ObservableObject {
    @Published var transactions: [Transaction] = []
    @Published var recurringTransactions: [RecurringTransaction] = []
    @Published var savingsGoal: SavingsGoal?

    private let transactionsKey = "transactions"
    private let recurringKey = "recurring_transactions"
    private let savingsGoalKey = "savings_goal"

    private let budgetService = BudgetService.shared
    private let recurringService = RecurringService.shared

    private var syncTask: Task<Void, Never>?

    weak var coordinator: DataManager?

    // MARK: - Computed Properties
    var balance: Double {
        transactions.reduce(0) { result, t in
            t.type == .income ? result + t.amount : result - t.amount
        }
    }

    var monthlyIncome: Double {
        let calendar = Calendar.current
        return transactions
            .filter { t in
                t.type == .income && calendar.isDate(t.date, equalTo: Date(), toGranularity: .month)
            }
            .reduce(0) { $0 + $1.amount }
    }

    var monthlyExpenses: Double {
        let calendar = Calendar.current
        return transactions
            .filter { t in
                t.type == .expense && calendar.isDate(t.date, equalTo: Date(), toGranularity: .month)
            }
            .reduce(0) { $0 + $1.amount }
    }

    var projectedMonthlyExpenses: Double {
        recurringTransactions
            .filter { $0.isActive && $0.type == .expense }
            .reduce(0) { result, rt in
                switch rt.frequency {
                case .weekly: return result + rt.amount * 4.33
                case .biweekly: return result + rt.amount * 2.17
                case .monthly: return result + rt.amount
                case .quarterly: return result + rt.amount / 3
                case .yearly: return result + rt.amount / 12
                }
            }
    }

    var projectedMonthlyIncome: Double {
        recurringTransactions
            .filter { $0.isActive && $0.type == .income }
            .reduce(0) { result, rt in
                switch rt.frequency {
                case .weekly: return result + rt.amount * 4.33
                case .biweekly: return result + rt.amount * 2.17
                case .monthly: return result + rt.amount
                case .quarterly: return result + rt.amount / 3
                case .yearly: return result + rt.amount / 12
                }
            }
    }

    // MARK: - Load & Save
    func load() {
        if let data = UserDefaults.standard.data(forKey: transactionsKey),
           let decoded = try? JSONDecoder().decode([Transaction].self, from: data) {
            transactions = decoded
        }

        if let data = UserDefaults.standard.data(forKey: recurringKey),
           let decoded = try? JSONDecoder().decode([RecurringTransaction].self, from: data) {
            recurringTransactions = decoded
        }

        if let data = UserDefaults.standard.data(forKey: savingsGoalKey),
           let decoded = try? JSONDecoder().decode(SavingsGoal.self, from: data) {
            savingsGoal = decoded
        }
    }

    func saveTransactions() {
        if let data = try? JSONEncoder().encode(transactions) {
            UserDefaults.standard.set(data, forKey: transactionsKey)
        }
        coordinator?.updateWidgetData()
    }

    func saveRecurring() {
        if let data = try? JSONEncoder().encode(recurringTransactions) {
            UserDefaults.standard.set(data, forKey: recurringKey)
        }
    }

    func saveSavingsGoal() {
        if let data = try? JSONEncoder().encode(savingsGoal) {
            UserDefaults.standard.set(data, forKey: savingsGoalKey)
        }
    }

    // MARK: - Sync Transactions
    func syncTransactions() async {
        guard coordinator?.isDemoMode != true else { return }
        do {
            let serverTransactions = try await budgetService.getTransactions(year: nil, month: nil)
            transactions = serverTransactions
            saveTransactions()
            coordinator?.generateInsights(for: .budget)
        } catch {
            coordinator?.syncError = error.localizedDescription
        }
    }

    // MARK: - Sync Recurring
    func syncRecurring() async {
        guard coordinator?.isDemoMode != true else { return }
        do {
            // Process due recurring transactions first
            do {
                let result = try await recurringService.processRecurring()
                AppLogger.sync.info("syncRecurring: \(result.processed) checked, \(result.created) transactions created")
                if result.created > 0 {
                    let serverTransactions = try await budgetService.getTransactions(year: nil, month: nil)
                    transactions = serverTransactions
                    saveTransactions()
                }
            } catch {
                AppLogger.sync.error("Failed to process recurring: \(error.localizedDescription)")
            }
            // Fetch updated recurring transactions
            let serverRecurring = try await recurringService.getRecurringTransactions()
            recurringTransactions = serverRecurring
            saveRecurring()
        } catch {
            coordinator?.syncError = error.localizedDescription
        }
    }

    // MARK: - Transaction CRUD
    func addTransaction(_ transaction: Transaction) {
        coordinator?.trackInsightFirstDate(for: .budget)
        coordinator?.recordActivity()

        // Optimistic update
        transactions.append(transaction)
        saveTransactions()

        // Sync to server
        Task { [weak self] in
            guard let self = self else { return }
            do {
                let serverTransaction = try await budgetService.createTransaction(transaction)
                await MainActor.run {
                    if let index = self.transactions.firstIndex(where: { $0.id == transaction.id }) {
                        self.transactions[index] = serverTransaction
                        self.saveTransactions()
                    }
                }
                await syncSavingsGoal()
            } catch {
                await MainActor.run {
                    self.transactions.removeAll { $0.id == transaction.id }
                    self.saveTransactions()
                    self.coordinator?.syncError = error.localizedDescription
                }
            }
        }
    }

    func deleteTransaction(_ transaction: Transaction) {
        let removedTransaction = transaction
        transactions.removeAll { $0.id == transaction.id }
        saveTransactions()

        Task { [weak self] in
            guard let self = self else { return }
            do {
                try await budgetService.deleteTransaction(transaction.id)
                await syncSavingsGoal()
            } catch {
                await MainActor.run {
                    self.transactions.append(removedTransaction)
                    self.saveTransactions()
                    self.coordinator?.syncError = error.localizedDescription
                }
            }
        }
    }

    func updateTransaction(_ transaction: Transaction) {
        guard let index = transactions.firstIndex(where: { $0.id == transaction.id }) else { return }
        let oldTransaction = transactions[index]

        // Optimistic update
        transactions[index] = transaction
        saveTransactions()

        Task { [weak self] in
            guard let self = self else { return }
            do {
                let serverTransaction = try await budgetService.updateTransaction(transaction)
                await MainActor.run {
                    if let idx = self.transactions.firstIndex(where: { $0.id == transaction.id }) {
                        self.transactions[idx] = serverTransaction
                        self.saveTransactions()
                    }
                }
                await syncSavingsGoal()
            } catch {
                await MainActor.run {
                    if let idx = self.transactions.firstIndex(where: { $0.id == transaction.id }) {
                        self.transactions[idx] = oldTransaction
                        self.saveTransactions()
                    }
                    self.coordinator?.syncError = error.localizedDescription
                }
            }
        }
    }

    // MARK: - Recurring CRUD
    func addRecurring(_ recurring: RecurringTransaction) {
        coordinator?.recordActivity()

        // Optimistic update
        recurringTransactions.append(recurring)
        saveRecurring()

        Task { [weak self] in
            guard let self = self else { return }
            do {
                let serverRecurring = try await recurringService.createRecurring(recurring)
                if let index = recurringTransactions.firstIndex(where: { $0.id == recurring.id }) {
                    recurringTransactions[index] = serverRecurring
                    saveRecurring()
                }

                // Process recurring to create transaction if due today
                let result = try await recurringService.processRecurring()
                AppLogger.sync.info("After add recurring: \(result.created) transactions created")

                if result.created > 0 {
                    await syncTransactions()
                }
            } catch {
                recurringTransactions.removeAll { $0.id == recurring.id }
                saveRecurring()
                coordinator?.syncError = error.localizedDescription
            }
        }
    }

    func updateRecurring(_ recurring: RecurringTransaction) {
        guard let index = recurringTransactions.firstIndex(where: { $0.id == recurring.id }) else { return }
        let oldRecurring = recurringTransactions[index]

        // Optimistic update
        recurringTransactions[index] = recurring
        saveRecurring()

        Task { [weak self] in
            guard let self = self else { return }
            do {
                let serverRecurring = try await recurringService.updateRecurring(recurring)
                if let idx = recurringTransactions.firstIndex(where: { $0.id == recurring.id }) {
                    recurringTransactions[idx] = serverRecurring
                    saveRecurring()
                }
            } catch {
                if let idx = recurringTransactions.firstIndex(where: { $0.id == recurring.id }) {
                    recurringTransactions[idx] = oldRecurring
                    saveRecurring()
                }
                coordinator?.syncError = error.localizedDescription
            }
        }
    }

    func deleteRecurring(_ recurring: RecurringTransaction) {
        let removedRecurring = recurring
        recurringTransactions.removeAll { $0.id == recurring.id }
        saveRecurring()

        Task { [weak self] in
            guard let self = self else { return }
            do {
                try await recurringService.deleteRecurring(recurring.id)
            } catch {
                recurringTransactions.append(removedRecurring)
                saveRecurring()
                coordinator?.syncError = error.localizedDescription
            }
        }
    }

    // MARK: - Savings Goal
    func setSavingsGoal(_ target: Double) {
        Task { [weak self] in
            guard let self = self else { return }
            do {
                let goal = try await budgetService.setSavingsGoal(target: target)
                await MainActor.run {
                    self.savingsGoal = goal
                }
            } catch {
                AppLogger.sync.error("Failed to set savings goal: \(error)")
                await MainActor.run {
                    self.savingsGoal = SavingsGoal(
                        monthlyTarget: target,
                        currentSavings: self.monthlyIncome - self.monthlyExpenses,
                        monthlyIncome: self.monthlyIncome,
                        monthlyExpenses: self.monthlyExpenses
                    )
                    self.saveSavingsGoal()
                }
            }
        }
    }

    func syncSavingsGoal() async {
        do {
            let goal = try await budgetService.getSavingsGoal()
            await MainActor.run {
                self.savingsGoal = goal
                if goal != nil {
                    self.saveSavingsGoal()
                } else {
                    UserDefaults.standard.removeObject(forKey: self.savingsGoalKey)
                }
            }
        } catch {
            AppLogger.sync.error("Failed to sync savings goal: \(error)")
        }
    }

    func deleteSavingsGoal() {
        Task { [weak self] in
            do {
                try await self?.budgetService.deleteSavingsGoal()
            } catch {
                AppLogger.sync.error("Failed to delete savings goal on server: \(error)")
            }
        }
        savingsGoal = nil
        UserDefaults.standard.removeObject(forKey: savingsGoalKey)
    }

    // MARK: - Queries
    func transactionsForPeriod(_ period: AnalyticsPeriod) -> [Transaction] {
        let calendar = Calendar.current
        let now = Date()

        return transactions.filter { transaction in
            switch period {
            case .week:
                return calendar.isDate(transaction.date, equalTo: now, toGranularity: .weekOfYear)
            case .month:
                return calendar.isDate(transaction.date, equalTo: now, toGranularity: .month)
            case .year:
                return calendar.isDate(transaction.date, equalTo: now, toGranularity: .year)
            }
        }
    }

    func transactionsForDate(_ date: Date) -> [Transaction] {
        let calendar = Calendar.current
        return transactions.filter { calendar.isDate($0.date, inSameDayAs: date) }
    }

    func incomeForPeriod(_ period: AnalyticsPeriod) -> Double {
        transactionsForPeriod(period)
            .filter { $0.type == .income }
            .reduce(0) { $0 + $1.amount }
    }

    func expensesForPeriod(_ period: AnalyticsPeriod) -> Double {
        transactionsForPeriod(period)
            .filter { $0.type == .expense }
            .reduce(0) { $0 + $1.amount }
    }

    func balanceForPeriod(_ period: AnalyticsPeriod) -> Double {
        incomeForPeriod(period) - expensesForPeriod(period)
    }

    func balanceForDate(_ date: Date) -> Double {
        let dayTransactions = transactionsForDate(date)
        let income = dayTransactions.filter { $0.type == .income }.reduce(0) { $0 + $1.amount }
        let expenses = dayTransactions.filter { $0.type == .expense }.reduce(0) { $0 + $1.amount }
        return income - expenses
    }

    // MARK: - Clear
    func clear() {
        transactions = []
        recurringTransactions = []
        savingsGoal = nil
        UserDefaults.standard.removeObject(forKey: transactionsKey)
        UserDefaults.standard.removeObject(forKey: recurringKey)
        UserDefaults.standard.removeObject(forKey: savingsGoalKey)
    }
}

import AppIntents
import SwiftUI

// MARK: - Add Expense Intent

struct AddExpenseIntent: AppIntent {
    static var title: LocalizedStringResource = "Add Expense"
    static var description = IntentDescription("Add a new expense to your budget")
    static var openAppWhenRun: Bool = false

    @Parameter(title: "Description")
    var expenseDescription: String

    @Parameter(title: "Amount")
    var amount: Double

    @MainActor
    func perform() async throws -> some IntentResult & ProvidesDialog {
        guard !expenseDescription.trimmingCharacters(in: .whitespaces).isEmpty else {
            return .result(dialog: "Please provide a description")
        }

        guard amount > 0 else {
            return .result(dialog: "Amount must be greater than 0")
        }

        let dataManager = DataManager.shared
        let transaction = Transaction(
            title: expenseDescription.trimmingCharacters(in: .whitespaces),
            amount: amount,
            type: .expense
        )
        dataManager.addTransaction(transaction)

        let currency = dataManager.profile.currency.symbol
        return .result(dialog: "Added expense: \(currency)\(String(format: "%.2f", amount)) for '\(expenseDescription)'")
    }

    static var parameterSummary: some ParameterSummary {
        Summary("Add expense \(\.$amount) for \(\.$expenseDescription)")
    }
}

// MARK: - Add Income Intent

struct AddIncomeIntent: AppIntent {
    static var title: LocalizedStringResource = "Add Income"
    static var description = IntentDescription("Add income to your budget")
    static var openAppWhenRun: Bool = false

    @Parameter(title: "Description")
    var incomeDescription: String

    @Parameter(title: "Amount")
    var amount: Double

    @MainActor
    func perform() async throws -> some IntentResult & ProvidesDialog {
        guard !incomeDescription.trimmingCharacters(in: .whitespaces).isEmpty else {
            return .result(dialog: "Please provide a description")
        }

        guard amount > 0 else {
            return .result(dialog: "Amount must be greater than 0")
        }

        let dataManager = DataManager.shared
        let transaction = Transaction(
            title: incomeDescription.trimmingCharacters(in: .whitespaces),
            amount: amount,
            type: .income
        )
        dataManager.addTransaction(transaction)

        let currency = dataManager.profile.currency.symbol
        return .result(dialog: "Added income: \(currency)\(String(format: "%.2f", amount)) for '\(incomeDescription)'")
    }

    static var parameterSummary: some ParameterSummary {
        Summary("Add income \(\.$amount) for \(\.$incomeDescription)")
    }
}

// MARK: - Check Balance Intent

struct CheckBalanceIntent: AppIntent {
    static var title: LocalizedStringResource = "Check Balance"
    static var description = IntentDescription("Check your current budget balance")
    static var openAppWhenRun: Bool = false

    @MainActor
    func perform() async throws -> some IntentResult & ProvidesDialog {
        let dataManager = DataManager.shared
        let balance = dataManager.balance
        let income = dataManager.monthlyIncome
        let expenses = dataManager.monthlyExpenses
        let currency = dataManager.profile.currency.symbol

        let balanceStr = "\(currency)\(String(format: "%.2f", abs(balance)))"
        let incomeStr = "\(currency)\(String(format: "%.2f", income))"
        let expensesStr = "\(currency)\(String(format: "%.2f", expenses))"

        var message = "Your balance: \(balance >= 0 ? "" : "-")\(balanceStr)\n\n"
        message += "This month:\n"
        message += "📈 Income: \(incomeStr)\n"
        message += "📉 Expenses: \(expensesStr)"

        if balance < 0 {
            message += "\n\n⚠️ You're over budget this month"
        } else if income > 0 {
            let savingsRate = Int((balance / income) * 100)
            message += "\n\n💰 Savings rate: \(savingsRate)%"
        }

        return .result(dialog: "\(message)")
    }
}

// MARK: - Check Life Score Intent

struct CheckLifeScoreIntent: AppIntent {
    static var title: LocalizedStringResource = "Check Life Score"
    static var description = IntentDescription("Check your current Life Score")
    static var openAppWhenRun: Bool = false

    @MainActor
    func perform() async throws -> some IntentResult & ProvidesDialog {
        let dataManager = DataManager.shared
        let score = Int(dataManager.lifeScore(for: .week))
        let components = dataManager.lifeScoreComponents(for: .week)

        var message = "Your Life Score: \(score)/100\n\n"

        // Emoji based on score
        let emoji: String
        switch score {
        case 80...100: emoji = "🌟"
        case 60..<80: emoji = "👍"
        case 40..<60: emoji = "💪"
        default: emoji = "🎯"
        }

        message += "\(emoji) "
        switch score {
        case 80...100: message += "Excellent! You're crushing it!"
        case 60..<80: message += "Good job! Keep it up!"
        case 40..<60: message += "Room for improvement"
        default: message += "Let's get back on track!"
        }

        message += "\n\n"
        message += "🔥 Habits: \(Int(components.habits))%\n"
        message += "✅ Tasks: \(Int(components.tasks))%\n"
        message += "💰 Budget: \(Int(components.budget))%"

        return .result(dialog: "\(message)")
    }
}

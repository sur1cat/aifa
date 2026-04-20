import Foundation
import Combine
import SwiftUI

@MainActor
class AnalyticsManager: ObservableObject {
    @Published var currentForecast: BudgetForecast?
    @Published var aiExpenseAnalysis: AIExpenseAnalysis?

    private let forecastKey = "budget_forecast"
    private let aiAnalysisKey = "ai_expense_analysis"

    weak var coordinator: DataManager?

    // MARK: - Load & Save
    func load() {
        if let data = UserDefaults.standard.data(forKey: forecastKey),
           let decoded = try? JSONDecoder().decode(BudgetForecast.self, from: data) {
            currentForecast = decoded
        }

        if let data = UserDefaults.standard.data(forKey: aiAnalysisKey),
           let decoded = try? JSONDecoder().decode(AIExpenseAnalysis.self, from: data) {
            aiExpenseAnalysis = decoded
        }
    }

    func saveForecast() {
        if let data = try? JSONEncoder().encode(currentForecast) {
            UserDefaults.standard.set(data, forKey: forecastKey)
        }
    }

    func saveAIAnalysis() {
        if let data = try? JSONEncoder().encode(aiExpenseAnalysis) {
            UserDefaults.standard.set(data, forKey: aiAnalysisKey)
        }
    }

    // MARK: - Life Score
    func budgetHealthScore(for period: AnalyticsPeriod) -> Double {
        guard let budget = coordinator?.budgetManager else { return 0 }

        let income = budget.incomeForPeriod(period)
        let expenses = budget.expensesForPeriod(period)

        if income == 0 && expenses == 0 { return 0 }
        if income == 0 { return max(0, 100 - expenses / 10) }

        let savingsRate = (income - expenses) / income
        let normalized = (savingsRate + 0.5) / 0.8
        return max(0, min(100, normalized * 100))
    }

    func lifeScore(for period: AnalyticsPeriod = .week) -> Double {
        guard let coordinator = coordinator else { return 0 }

        let habitsScore = coordinator.habitsManager.completionRate(for: period)
        let tasksScore = coordinator.tasksManager.completionRate()
        let budgetScore = budgetHealthScore(for: period)

        return (habitsScore * 0.40) + (tasksScore * 0.30) + (budgetScore * 0.30)
    }

    func lifeScoreComponents(for period: AnalyticsPeriod) -> (habits: Double, tasks: Double, budget: Double) {
        guard let coordinator = coordinator else { return (0, 0, 0) }

        return (
            habits: coordinator.habitsManager.completionRate(for: period),
            tasks: coordinator.tasksManager.completionRate(),
            budget: budgetHealthScore(for: period)
        )
    }

    // MARK: - Life Score History
    func lifeScoreHistory(for period: AnalyticsPeriod) -> [LifeScoreDataPoint] {
        let calendar = Calendar.current
        var result: [LifeScoreDataPoint] = []

        let days: Int
        switch period {
        case .week: days = 7
        case .month: days = 30
        case .year: days = 12
        }

        if period == .year {
            for monthOffset in (0..<12).reversed() {
                guard let date = calendar.date(byAdding: .month, value: -monthOffset, to: Date()) else { continue }
                let score = calculateHistoricalLifeScore(for: date, granularity: .month)
                result.append(LifeScoreDataPoint(date: date, score: score))
            }
        } else {
            for dayOffset in (0..<days).reversed() {
                guard let date = calendar.date(byAdding: .day, value: -dayOffset, to: Date()) else { continue }
                let score = calculateHistoricalLifeScore(for: date, granularity: .day)
                result.append(LifeScoreDataPoint(date: date, score: score))
            }
        }

        return result
    }

    private func calculateHistoricalLifeScore(for date: Date, granularity: Calendar.Component) -> Double {
        guard let coordinator = coordinator else { return 0 }

        let calendar = Calendar.current
        let dateStr = Habit.dateString(from: date)
        let habits = coordinator.habits
        let transactions = coordinator.budgetManager.transactions

        // Habits completion
        var habitsCompleted = 0
        let habitsTotal = habits.count
        for habit in habits {
            if granularity == .day {
                if habit.completedDates.contains(dateStr) {
                    habitsCompleted += 1
                }
            } else {
                let hasCompletion = habit.completedDates.contains { dateString in
                    guard let completionDate = DateFormatters.apiDate.date(from: dateString) else { return false }
                    return calendar.isDate(completionDate, equalTo: date, toGranularity: .month)
                }
                if hasCompletion { habitsCompleted += 1 }
            }
        }
        let habitsRate = habitsTotal > 0 ? Double(habitsCompleted) / Double(habitsTotal) * 100 : 0

        // Tasks - use current rate
        let tasksRate = coordinator.tasksManager.completionRate()

        // Budget for period
        var income: Double = 0
        var expenses: Double = 0
        for transaction in transactions {
            let matches: Bool
            if granularity == .day {
                matches = calendar.isDate(transaction.date, inSameDayAs: date)
            } else {
                matches = calendar.isDate(transaction.date, equalTo: date, toGranularity: .month)
            }
            if matches {
                if transaction.type == .income {
                    income += transaction.amount
                } else {
                    expenses += transaction.amount
                }
            }
        }

        var budgetScore: Double = 50
        if income > 0 || expenses > 0 {
            if income == 0 {
                budgetScore = max(0, 100 - expenses / 10)
            } else {
                let savingsRate = (income - expenses) / income
                let normalized = (savingsRate + 0.5) / 0.8
                budgetScore = max(0, min(100, normalized * 100))
            }
        }

        return (habitsRate * 0.40) + (tasksRate * 0.30) + (budgetScore * 0.30)
    }

    // MARK: - Spending by Category
    func spendingByCategory(for period: AnalyticsPeriod) -> [CategorySpending] {
        guard let transactions = coordinator?.budgetManager.transactions else { return [] }

        let calendar = Calendar.current
        let now = Date()

        let startDate: Date
        switch period {
        case .week:
            startDate = calendar.date(byAdding: .day, value: -7, to: now) ?? now
        case .month:
            startDate = calendar.date(byAdding: .month, value: -1, to: now) ?? now
        case .year:
            startDate = calendar.date(byAdding: .year, value: -1, to: now) ?? now
        }

        var categoryTotals: [String: Double] = [:]
        for transaction in transactions {
            guard transaction.type == .expense,
                  transaction.date >= startDate else { continue }

            let category = transaction.category.isEmpty ? "Other" : transaction.category
            categoryTotals[category, default: 0] += transaction.amount
        }

        let colors: [Color] = [
            Color.hf.expense,
            Color.hf.warning,
            Color.hf.info,
            Color.hf.accent,
            .purple,
            .pink,
            .orange,
            .cyan
        ]

        let sorted = categoryTotals.sorted { $0.value > $1.value }
        return sorted.enumerated().map { index, item in
            CategorySpending(
                category: item.key,
                amount: item.value,
                color: colors[index % colors.count]
            )
        }
    }

    // MARK: - Daily Charts
    func dailyExpensesForWeek() -> [(day: String, amount: Double)] {
        guard let transactions = coordinator?.budgetManager.transactions else { return [] }

        let calendar = Calendar.current
        var result: [(String, Double)] = []

        for dayOffset in (0..<7).reversed() {
            guard let date = calendar.date(byAdding: .day, value: -dayOffset, to: Date()) else { continue }
            let dayName = DateFormatters.shortWeekday.string(from: date)
            let amount = transactions
                .filter { t in
                    t.type == .expense && calendar.isDate(t.date, inSameDayAs: date)
                }
                .reduce(0) { $0 + $1.amount }
            result.append((dayName, amount))
        }
        return result
    }

    // MARK: - Budget Forecast
    func generateBudgetForecast() -> BudgetForecast {
        guard let coordinator = coordinator else {
            return BudgetForecast(forecastMonth: Date())
        }

        let calendar = Calendar.current
        guard let nextMonth = calendar.date(byAdding: .month, value: 1, to: Date()) else {
            return BudgetForecast(forecastMonth: Date())
        }

        let historicalMonths = getHistoricalMonthlyData(months: 6)
        let categoryForecasts = calculateCategoryForecasts(from: historicalMonths)
        let projectedExpenses = categoryForecasts.reduce(0) { $0 + $1.projectedAmount }
        let projectedIncome = calculateProjectedIncome(from: historicalMonths) + coordinator.budgetManager.projectedMonthlyIncome

        let trend = calculateExpenseTrend(from: historicalMonths)
        let confidence = min(100, Double(historicalMonths.count) * 15 + 20)
        let seasonalFactors = detectSeasonality(for: nextMonth)

        let forecast = BudgetForecast(
            forecastMonth: nextMonth,
            categoryForecasts: categoryForecasts,
            projectedExpenses: projectedExpenses,
            projectedIncome: projectedIncome,
            projectedSavings: projectedIncome - projectedExpenses,
            expenseTrend: trend,
            confidenceScore: confidence,
            seasonalFactors: seasonalFactors
        )

        currentForecast = forecast
        saveForecast()
        return forecast
    }

    private func getHistoricalMonthlyData(months: Int) -> [[Transaction]] {
        guard let transactions = coordinator?.budgetManager.transactions else { return [] }

        let calendar = Calendar.current
        var result: [[Transaction]] = []

        for monthOffset in 1...months {
            guard let monthDate = calendar.date(byAdding: .month, value: -monthOffset, to: Date()) else { continue }
            let monthTransactions = transactions.filter {
                calendar.isDate($0.date, equalTo: monthDate, toGranularity: .month)
            }
            if !monthTransactions.isEmpty {
                result.append(monthTransactions)
            }
        }

        return result
    }

    private func calculateCategoryForecasts(from historicalMonths: [[Transaction]]) -> [CategoryForecast] {
        var categoryTotals: [String: [Double]] = [:]

        for monthData in historicalMonths {
            var monthCategoryTotals: [String: Double] = [:]
            for tx in monthData where tx.type == .expense {
                let category = tx.category.isEmpty ? "Other" : tx.category
                monthCategoryTotals[category, default: 0] += tx.amount
            }
            for (category, total) in monthCategoryTotals {
                categoryTotals[category, default: []].append(total)
            }
        }

        var forecasts: [CategoryForecast] = []

        for (category, amounts) in categoryTotals {
            guard !amounts.isEmpty else { continue }

            let average = amounts.reduce(0, +) / Double(amounts.count)
            let weightedAvg = calculateWeightedAverage(amounts)
            let recurringAmount = getRecurringAmountForCategory(category)
            let variableAmount = max(0, weightedAvg - recurringAmount)
            let trend = determineTrend(amounts)
            let changePercent = amounts.count > 1
                ? ((amounts.first! - amounts.last!) / max(amounts.last!, 1)) * 100
                : 0

            forecasts.append(CategoryForecast(
                category: category,
                projectedAmount: weightedAvg,
                historicalAverage: average,
                changePercent: changePercent,
                trend: trend,
                recurringAmount: recurringAmount,
                variableAmount: variableAmount
            ))
        }

        return forecasts.sorted { $0.projectedAmount > $1.projectedAmount }
    }

    private func calculateWeightedAverage(_ values: [Double]) -> Double {
        guard !values.isEmpty else { return 0 }

        var weightedSum: Double = 0
        var totalWeight: Double = 0

        for (index, value) in values.enumerated() {
            let weight = Double(values.count - index)
            weightedSum += value * weight
            totalWeight += weight
        }

        return weightedSum / totalWeight
    }

    private func getRecurringAmountForCategory(_ category: String) -> Double {
        guard let recurring = coordinator?.budgetManager.recurringTransactions else { return 0 }

        let categoryLower = category.lowercased()
        return recurring
            .filter { $0.isActive && $0.type == .expense }
            .filter {
                $0.category.rawValue.lowercased().contains(categoryLower) ||
                categoryLower.contains($0.category.rawValue.lowercased()) ||
                $0.title.lowercased().contains(categoryLower)
            }
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

    private func determineTrend(_ values: [Double]) -> TrendDirection {
        guard values.count >= 2 else { return .stable }

        let recent = values.prefix(2).reduce(0, +) / 2
        let older = values.suffix(from: max(0, values.count - 2)).reduce(0, +) / Double(min(2, values.count))

        guard older > 0 else { return .stable }
        let change = (recent - older) / older

        if change > 0.1 { return .up }
        if change < -0.1 { return .down }
        return .stable
    }

    private func calculateProjectedIncome(from historicalMonths: [[Transaction]]) -> Double {
        let incomes = historicalMonths.map { month in
            month.filter { $0.type == .income }.reduce(0) { $0 + $1.amount }
        }

        guard !incomes.isEmpty else { return 0 }
        return calculateWeightedAverage(incomes)
    }

    private func calculateExpenseTrend(from historicalMonths: [[Transaction]]) -> TrendDirection {
        let monthlyExpenses = historicalMonths.map { month in
            month.filter { $0.type == .expense }.reduce(0) { $0 + $1.amount }
        }
        return determineTrend(monthlyExpenses)
    }

    private func detectSeasonality(for month: Date) -> [SeasonalFactor]? {
        let calendar = Calendar.current
        let targetMonth = calendar.component(.month, from: month)

        var factors: [SeasonalFactor] = []

        switch targetMonth {
        case 12:
            factors.append(SeasonalFactor(category: "Shopping", monthlyMultiplier: 1.5, reason: "Holiday shopping"))
            factors.append(SeasonalFactor(category: "Entertainment", monthlyMultiplier: 1.3, reason: "Holiday gatherings"))
        case 8, 9:
            factors.append(SeasonalFactor(category: "Education", monthlyMultiplier: 2.0, reason: "Back to school"))
        case 1:
            factors.append(SeasonalFactor(category: "Health", monthlyMultiplier: 1.4, reason: "New Year resolutions"))
        case 2:
            factors.append(SeasonalFactor(category: "gift", monthlyMultiplier: 1.5, reason: "Valentine's Day"))
        default:
            break
        }

        return factors.isEmpty ? nil : factors
    }

    // MARK: - Currency Format
    func formatCurrency(_ amount: Double) -> String {
        guard let profile = coordinator?.profile else {
            return "$\(String(format: "%.2f", amount))"
        }

        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        formatter.currencySymbol = profile.currency.symbol
        formatter.maximumFractionDigits = profile.currency == .jpy || profile.currency == .krw ? 0 : 2
        return formatter.string(from: NSNumber(value: amount)) ?? "\(profile.currency.symbol)0"
    }

    // MARK: - Clear
    func clear() {
        currentForecast = nil
        aiExpenseAnalysis = nil
        UserDefaults.standard.removeObject(forKey: forecastKey)
        UserDefaults.standard.removeObject(forKey: aiAnalysisKey)
    }
}

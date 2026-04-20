import Foundation
import SwiftUI

class ExportService {
    static let shared = ExportService()

    private let dateFormatter: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd"
        return formatter
    }()

    private let dateTimeFormatter: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd HH:mm"
        return formatter
    }()

    // MARK: - Export Habits to CSV

    func exportHabitsToCSV(habits: [Habit]) -> URL? {
        var csv = "ID,Title,Icon,Color,Period,Created At,Streak,Total Completions,Completed Dates\n"

        for habit in habits {
            let completedDatesStr = habit.completedDates.joined(separator: "; ")
            let row = [
                habit.id.uuidString,
                escapeCSV(habit.title),
                habit.icon,
                habit.color,
                habit.period.rawValue,
                dateTimeFormatter.string(from: habit.createdAt),
                String(habit.streak),
                String(habit.completedDates.count),
                escapeCSV(completedDatesStr)
            ].joined(separator: ",")
            csv += row + "\n"
        }

        return saveToFile(content: csv, filename: "atoma_habits_\(dateFormatter.string(from: Date())).csv")
    }

    // MARK: - Export Tasks to CSV

    func exportTasksToCSV(tasks: [DailyTask]) -> URL? {
        var csv = "ID,Title,Priority,Status,Due Date,Created At\n"

        for task in tasks {
            let row = [
                task.id.uuidString,
                escapeCSV(task.title),
                task.priority.rawValue,
                task.isCompleted ? "Completed" : "Pending",
                dateFormatter.string(from: task.dueDate),
                dateTimeFormatter.string(from: task.createdAt)
            ].joined(separator: ",")
            csv += row + "\n"
        }

        return saveToFile(content: csv, filename: "atoma_tasks_\(dateFormatter.string(from: Date())).csv")
    }

    // MARK: - Export Transactions to CSV

    func exportTransactionsToCSV(transactions: [Transaction], currency: Currency) -> URL? {
        var csv = "ID,Title,Amount (\(currency.rawValue)),Type,Category,Date\n"

        for transaction in transactions {
            let row = [
                transaction.id.uuidString,
                escapeCSV(transaction.title),
                String(format: "%.2f", transaction.amount),
                transaction.type.rawValue,
                escapeCSV(transaction.category),
                dateFormatter.string(from: transaction.date)
            ].joined(separator: ",")
            csv += row + "\n"
        }

        return saveToFile(content: csv, filename: "atoma_transactions_\(dateFormatter.string(from: Date())).csv")
    }

    // MARK: - Export Recurring Transactions to CSV

    func exportRecurringToCSV(recurring: [RecurringTransaction], currency: Currency) -> URL? {
        var csv = "ID,Title,Amount (\(currency.rawValue)),Type,Category,Frequency,Start Date,Next Date,Active\n"

        for item in recurring {
            let row = [
                item.id.uuidString,
                escapeCSV(item.title),
                String(format: "%.2f", item.amount),
                item.type.rawValue,
                item.category.rawValue,
                item.frequency.rawValue,
                dateFormatter.string(from: item.startDate),
                dateFormatter.string(from: item.nextDate),
                item.isActive ? "Yes" : "No"
            ].joined(separator: ",")
            csv += row + "\n"
        }

        return saveToFile(content: csv, filename: "atoma_recurring_\(dateFormatter.string(from: Date())).csv")
    }

    // MARK: - Export All Data

    func exportAllData(habits: [Habit], tasks: [DailyTask], transactions: [Transaction], recurring: [RecurringTransaction], currency: Currency) -> URL? {
        let tempDir = FileManager.default.temporaryDirectory
        let exportDir = tempDir.appendingPathComponent("atoma_export_\(dateFormatter.string(from: Date()))")

        do {
            try FileManager.default.createDirectory(at: exportDir, withIntermediateDirectories: true)

            // Export each type
            if let habitsURL = exportHabitsToCSV(habits: habits) {
                let destURL = exportDir.appendingPathComponent(habitsURL.lastPathComponent)
                try? FileManager.default.copyItem(at: habitsURL, to: destURL)
            }

            if let tasksURL = exportTasksToCSV(tasks: tasks) {
                let destURL = exportDir.appendingPathComponent(tasksURL.lastPathComponent)
                try? FileManager.default.copyItem(at: tasksURL, to: destURL)
            }

            if let transactionsURL = exportTransactionsToCSV(transactions: transactions, currency: currency) {
                let destURL = exportDir.appendingPathComponent(transactionsURL.lastPathComponent)
                try? FileManager.default.copyItem(at: transactionsURL, to: destURL)
            }

            if !recurring.isEmpty, let recurringURL = exportRecurringToCSV(recurring: recurring, currency: currency) {
                let destURL = exportDir.appendingPathComponent(recurringURL.lastPathComponent)
                try? FileManager.default.copyItem(at: recurringURL, to: destURL)
            }

            // Create summary file
            let summary = createSummary(habits: habits, tasks: tasks, transactions: transactions, recurring: recurring, currency: currency)
            let summaryURL = exportDir.appendingPathComponent("summary.txt")
            try summary.write(to: summaryURL, atomically: true, encoding: .utf8)

            return exportDir
        } catch {
            print("Export error: \(error)")
            return nil
        }
    }

    // MARK: - Helpers

    private func escapeCSV(_ string: String) -> String {
        if string.contains(",") || string.contains("\"") || string.contains("\n") {
            return "\"\(string.replacingOccurrences(of: "\"", with: "\"\""))\""
        }
        return string
    }

    private func saveToFile(content: String, filename: String) -> URL? {
        let tempDir = FileManager.default.temporaryDirectory
        let fileURL = tempDir.appendingPathComponent(filename)

        do {
            try content.write(to: fileURL, atomically: true, encoding: .utf8)
            return fileURL
        } catch {
            print("Failed to save file: \(error)")
            return nil
        }
    }

    private func createSummary(habits: [Habit], tasks: [DailyTask], transactions: [Transaction], recurring: [RecurringTransaction], currency: Currency) -> String {
        let completedTasks = tasks.filter { $0.isCompleted }.count
        let totalIncome = transactions.filter { $0.type == .income }.reduce(0) { $0 + $1.amount }
        let totalExpenses = transactions.filter { $0.type == .expense }.reduce(0) { $0 + $1.amount }
        let monthlyRecurring = recurring.filter { $0.isActive && $0.type == .expense }.reduce(0) { $0 + $1.amount }

        return """
        ATOMA DATA EXPORT
        =================
        Export Date: \(dateTimeFormatter.string(from: Date()))

        HABITS
        ------
        Total habits: \(habits.count)
        Active streaks: \(habits.filter { $0.streak > 0 }.count)
        Longest streak: \(habits.map { $0.streak }.max() ?? 0) days

        TASKS
        -----
        Total tasks: \(tasks.count)
        Completed: \(completedTasks)
        Pending: \(tasks.count - completedTasks)
        Completion rate: \(tasks.isEmpty ? 0 : Int(Double(completedTasks) / Double(tasks.count) * 100))%

        BUDGET
        ------
        Currency: \(currency.rawValue) (\(currency.symbol))
        Total income: \(currency.symbol)\(String(format: "%.2f", totalIncome))
        Total expenses: \(currency.symbol)\(String(format: "%.2f", totalExpenses))
        Balance: \(currency.symbol)\(String(format: "%.2f", totalIncome - totalExpenses))
        Transactions: \(transactions.count)

        RECURRING
        ---------
        Active subscriptions: \(recurring.filter { $0.isActive }.count)
        Monthly recurring expenses: \(currency.symbol)\(String(format: "%.2f", monthlyRecurring))

        ---
        Exported from Atoma
        """
    }
}

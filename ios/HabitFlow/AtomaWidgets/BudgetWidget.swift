import WidgetKit
import SwiftUI

struct BudgetProvider: TimelineProvider {
    func placeholder(in context: Context) -> BudgetEntry {
        BudgetEntry(date: Date(), budget: WidgetBudget(
            balance: 1250.00,
            income: 3000.00,
            expenses: 1750.00,
            currency: "USD",
            currencySymbol: "$"
        ))
    }

    func getSnapshot(in context: Context, completion: @escaping (BudgetEntry) -> Void) {
        let data = WidgetDataManager.shared.loadData()
        let entry = BudgetEntry(date: Date(), budget: data.budget)
        completion(entry)
    }

    func getTimeline(in context: Context, completion: @escaping (Timeline<BudgetEntry>) -> Void) {
        let data = WidgetDataManager.shared.loadData()
        let entry = BudgetEntry(date: Date(), budget: data.budget)

        let nextUpdate = Calendar.current.date(byAdding: .minute, value: 30, to: Date())!
        let timeline = Timeline(entries: [entry], policy: .after(nextUpdate))
        completion(timeline)
    }
}

struct BudgetEntry: TimelineEntry {
    let date: Date
    let budget: WidgetBudget
}

struct BudgetWidgetEntryView: View {
    var entry: BudgetProvider.Entry
    @Environment(\.widgetFamily) var family

    var body: some View {
        switch family {
        case .systemSmall:
            SmallBudgetView(budget: entry.budget)
        case .systemMedium:
            MediumBudgetView(budget: entry.budget)
        default:
            SmallBudgetView(budget: entry.budget)
        }
    }
}

struct SmallBudgetView: View {
    let budget: WidgetBudget

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Image(systemName: "creditcard")
                    .font(.system(size: 14, weight: .semibold))
                Text("Budget")
                    .font(.system(size: 14, weight: .semibold))
                Spacer()
            }
            .foregroundStyle(.secondary)

            Spacer()

            VStack(alignment: .leading, spacing: 4) {
                Text("Balance")
                    .font(.system(size: 12))
                    .foregroundStyle(.secondary)

                Text("\(budget.currencySymbol)\(formatAmount(budget.balance))")
                    .font(.system(size: 24, weight: .bold, design: .rounded))
                    .foregroundStyle(budget.balance >= 0 ? .primary : .red)
            }

            Spacer()
        }
        .containerBackground(.fill.tertiary, for: .widget)
    }

    func formatAmount(_ amount: Double) -> String {
        let absAmount = abs(amount)
        if absAmount >= 1000000 {
            return String(format: "%.1fM", absAmount / 1000000)
        } else if absAmount >= 1000 {
            return String(format: "%.1fK", absAmount / 1000)
        } else {
            return String(format: "%.0f", absAmount)
        }
    }
}

struct MediumBudgetView: View {
    let budget: WidgetBudget

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Image(systemName: "creditcard")
                    .font(.system(size: 14, weight: .semibold))
                Text("Budget")
                    .font(.system(size: 14, weight: .semibold))
                Spacer()
                Text(budget.currency)
                    .font(.system(size: 12))
                    .foregroundStyle(.secondary)
            }
            .foregroundStyle(.secondary)

            Spacer()

            HStack(spacing: 16) {
                // Balance
                VStack(alignment: .leading, spacing: 4) {
                    Text("Balance")
                        .font(.system(size: 11))
                        .foregroundStyle(.secondary)

                    Text("\(budget.currencySymbol)\(formatAmount(budget.balance))")
                        .font(.system(size: 20, weight: .bold, design: .rounded))
                        .foregroundStyle(budget.balance >= 0 ? .primary : .red)
                }

                Spacer()

                // Income
                VStack(alignment: .trailing, spacing: 4) {
                    Text("Income")
                        .font(.system(size: 11))
                        .foregroundStyle(.secondary)

                    Text("+\(budget.currencySymbol)\(formatAmount(budget.income))")
                        .font(.system(size: 14, weight: .medium, design: .rounded))
                        .foregroundStyle(.green)
                }

                // Expenses
                VStack(alignment: .trailing, spacing: 4) {
                    Text("Expenses")
                        .font(.system(size: 11))
                        .foregroundStyle(.secondary)

                    Text("-\(budget.currencySymbol)\(formatAmount(budget.expenses))")
                        .font(.system(size: 14, weight: .medium, design: .rounded))
                        .foregroundStyle(.red)
                }
            }

            Spacer()
        }
        .containerBackground(.fill.tertiary, for: .widget)
    }

    func formatAmount(_ amount: Double) -> String {
        let absAmount = abs(amount)
        if absAmount >= 1000000 {
            return String(format: "%.1fM", absAmount / 1000000)
        } else if absAmount >= 1000 {
            return String(format: "%.1fK", absAmount / 1000)
        } else {
            return String(format: "%.0f", absAmount)
        }
    }
}

struct BudgetWidget: Widget {
    let kind: String = "BudgetWidget"

    var body: some WidgetConfiguration {
        StaticConfiguration(kind: kind, provider: BudgetProvider()) { entry in
            BudgetWidgetEntryView(entry: entry)
        }
        .configurationDisplayName("Budget")
        .description("View your current balance")
        .supportedFamilies([.systemSmall, .systemMedium])
    }
}

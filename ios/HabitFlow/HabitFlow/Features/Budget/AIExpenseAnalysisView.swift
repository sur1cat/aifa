import SwiftUI

// MARK: - AI Expense Insight Chip

struct AIExpenseInsightChip: View {
    let insight: ExpenseInsight
    let currency: Currency
    var onTap: (() -> Void)?

    private var iconName: String {
        switch insight.type {
        case .pattern: return "waveform.path.ecg"
        case .habit: return "repeat.circle"
        case .impulse: return "bolt.fill"
        case .subscription: return "calendar.badge.clock"
        case .opportunity: return "lightbulb.fill"
        }
    }

    private var iconColor: Color {
        switch insight.type {
        case .pattern: return Color.hf.info
        case .habit: return Color.hf.warning
        case .impulse: return Color.hf.expense
        case .subscription: return Color.hf.accent
        case .opportunity: return Color.hf.income
        }
    }

    var body: some View {
        Button {
            onTap?()
        } label: {
            HStack(spacing: 10) {
                Image(systemName: iconName)
                    .font(.system(size: 14))
                    .foregroundStyle(iconColor)

                VStack(alignment: .leading, spacing: 2) {
                    Text(insight.title)
                        .font(.system(size: 13, weight: .semibold))
                        .foregroundStyle(.primary)

                    Text(insight.message)
                        .font(.system(size: 12))
                        .foregroundStyle(.secondary)
                        .lineLimit(2)
                }

                Spacer()

                if let amount = insight.amount {
                    Text(formatCurrency(amount))
                        .font(.system(size: 14, weight: .semibold, design: .rounded))
                        .foregroundStyle(iconColor)
                }
            }
            .padding(12)
            .background(iconColor.opacity(0.1))
            .clipShape(RoundedRectangle(cornerRadius: 12))
        }
        .buttonStyle(.plain)
    }

    private func formatCurrency(_ amount: Double) -> String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        formatter.currencySymbol = currency.symbol
        formatter.maximumFractionDigits = 0
        return formatter.string(from: NSNumber(value: amount)) ?? "\(currency.symbol)0"
    }
}

// MARK: - AI Insights Section

struct AIExpenseInsightsSection: View {
    let insights: [ExpenseInsight]
    let currency: Currency
    @State private var showAllInsights = false

    private var displayedInsights: [ExpenseInsight] {
        showAllInsights ? insights : Array(insights.prefix(3))
    }

    var body: some View {
        VStack(spacing: 8) {
            HStack {
                HStack(spacing: 6) {
                    Image(systemName: "brain.head.profile")
                        .font(.system(size: 13))
                        .foregroundStyle(Color.hf.accent)

                    Text("AI Spending Insights")
                        .font(.system(size: 14, weight: .semibold))
                }

                Spacer()

                if insights.count > 3 {
                    Button {
                        withAnimation {
                            showAllInsights.toggle()
                        }
                    } label: {
                        Text(showAllInsights ? "Show Less" : "Show All")
                            .font(.system(size: 12, weight: .medium))
                            .foregroundStyle(Color.hf.accent)
                    }
                }
            }

            ForEach(displayedInsights) { insight in
                AIExpenseInsightChip(insight: insight, currency: currency)
            }
        }
    }
}

// MARK: - Savings Suggestion Card

struct SavingsSuggestionCard: View {
    let suggestion: SavingsSuggestion
    let currency: Currency

    private var difficultyColor: Color {
        switch suggestion.difficulty {
        case .easy: return Color.hf.income
        case .medium: return Color.hf.warning
        case .hard: return Color.hf.expense
        }
    }

    private var difficultyText: String {
        switch suggestion.difficulty {
        case .easy: return "Easy"
        case .medium: return "Medium"
        case .hard: return "Hard"
        }
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack {
                Text(suggestion.category)
                    .font(.system(size: 14, weight: .semibold))

                Spacer()

                // Difficulty badge
                Text(difficultyText)
                    .font(.system(size: 11, weight: .medium))
                    .foregroundStyle(difficultyColor)
                    .padding(.horizontal, 8)
                    .padding(.vertical, 3)
                    .background(difficultyColor.opacity(0.15))
                    .clipShape(Capsule())
            }

            // Spending comparison
            HStack(spacing: 16) {
                VStack(alignment: .leading, spacing: 2) {
                    Text("Current")
                        .font(.system(size: 11))
                        .foregroundStyle(.secondary)
                    Text(formatCurrency(suggestion.currentSpending))
                        .font(.system(size: 15, weight: .medium, design: .rounded))
                        .foregroundStyle(Color.hf.expense)
                }

                Image(systemName: "arrow.right")
                    .font(.system(size: 12))
                    .foregroundStyle(.secondary)

                VStack(alignment: .leading, spacing: 2) {
                    Text("Suggested")
                        .font(.system(size: 11))
                        .foregroundStyle(.secondary)
                    Text(formatCurrency(suggestion.suggestedBudget))
                        .font(.system(size: 15, weight: .medium, design: .rounded))
                        .foregroundStyle(Color.hf.income)
                }

                Spacer()

                VStack(alignment: .trailing, spacing: 2) {
                    Text("Save")
                        .font(.system(size: 11))
                        .foregroundStyle(.secondary)
                    Text(formatCurrency(suggestion.potentialSavings))
                        .font(.system(size: 15, weight: .semibold, design: .rounded))
                        .foregroundStyle(Color.hf.income)
                }
            }

            Text(suggestion.reason)
                .font(.system(size: 12))
                .foregroundStyle(.secondary)
                .lineLimit(2)
        }
        .padding(14)
        .background(Color.hf.cardBackground)
        .clipShape(RoundedRectangle(cornerRadius: 14))
    }

    private func formatCurrency(_ amount: Double) -> String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        formatter.currencySymbol = currency.symbol
        formatter.maximumFractionDigits = 0
        return formatter.string(from: NSNumber(value: amount)) ?? "\(currency.symbol)0"
    }
}

// MARK: - Savings Suggestions Section

struct SavingsSuggestionsSection: View {
    let suggestions: [SavingsSuggestion]
    let currency: Currency

    var body: some View {
        VStack(spacing: 8) {
            HStack {
                HStack(spacing: 6) {
                    Image(systemName: "lightbulb.fill")
                        .font(.system(size: 13))
                        .foregroundStyle(Color.hf.warning)

                    Text("Ways to Save")
                        .font(.system(size: 14, weight: .semibold))
                }

                Spacer()
            }

            ForEach(suggestions.prefix(3)) { suggestion in
                SavingsSuggestionCard(suggestion: suggestion, currency: currency)
            }
        }
    }
}

// MARK: - Questionable Transaction Badge

struct QuestionableTransactionBadge: View {
    let questionable: QuestionableTransaction
    let currency: Currency

    private var categoryColor: Color {
        switch questionable.category {
        case .impulse: return Color.hf.expense
        case .duplicate: return Color.hf.warning
        case .excessive: return Color.hf.expense
        case .unnecessary: return Color.hf.warning
        }
    }

    private var categoryIcon: String {
        switch questionable.category {
        case .impulse: return "bolt.fill"
        case .duplicate: return "doc.on.doc.fill"
        case .excessive: return "exclamationmark.triangle.fill"
        case .unnecessary: return "questionmark.circle.fill"
        }
    }

    var body: some View {
        HStack(spacing: 4) {
            Image(systemName: categoryIcon)
                .font(.system(size: 10))

            if let savings = questionable.potentialSavings {
                Text("-\(formatCurrency(savings))")
                    .font(.system(size: 10, weight: .semibold))
            } else {
                Text(questionable.category.rawValue.capitalized)
                    .font(.system(size: 10, weight: .medium))
            }
        }
        .foregroundStyle(.white)
        .padding(.horizontal, 6)
        .padding(.vertical, 3)
        .background(categoryColor)
        .clipShape(Capsule())
    }

    private func formatCurrency(_ amount: Double) -> String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        formatter.currencySymbol = currency.symbol
        formatter.maximumFractionDigits = 0
        return formatter.string(from: NSNumber(value: amount)) ?? "\(currency.symbol)0"
    }
}

// MARK: - Enhanced Transaction Row with Questionable Badge

struct TransactionRowWithAnalysis: View {
    let transaction: Transaction
    let currency: Currency
    let questionable: QuestionableTransaction?

    private var categoryInfo: TransactionCategory {
        TransactionCategory(rawValue: transaction.category) ?? .other
    }

    var body: some View {
        HStack(spacing: 12) {
            // Category icon
            ZStack {
                Circle()
                    .fill(categoryInfo.color.opacity(0.15))
                    .frame(width: 40, height: 40)

                Image(systemName: categoryInfo.icon)
                    .font(.system(size: 16))
                    .foregroundStyle(categoryInfo.color)
            }

            VStack(alignment: .leading, spacing: 4) {
                HStack(spacing: 6) {
                    Text(transaction.title)
                        .font(.system(size: 16, weight: .medium))

                    if let q = questionable {
                        QuestionableTransactionBadge(questionable: q, currency: currency)
                    }
                }

                if let q = questionable {
                    Text(q.reason)
                        .font(.system(size: 11))
                        .foregroundStyle(Color.hf.warning)
                        .lineLimit(1)
                } else {
                    Text(transaction.date.formatted(date: .abbreviated, time: .omitted))
                        .font(.system(size: 13))
                        .foregroundStyle(.secondary)
                }
            }

            Spacer()

            Text("\(transaction.type == .income ? "+" : "-")\(formatAmount(transaction.amount))")
                .font(.system(size: 16, weight: .semibold, design: .rounded))
                .foregroundStyle(transaction.type == .income ? Color.hf.income : Color.hf.expense)
        }
        .padding()
        .background(
            questionable != nil ?
            Color.hf.warning.opacity(0.05) :
            Color.hf.cardBackground
        )
        .clipShape(RoundedRectangle(cornerRadius: 14))
        .overlay(
            questionable != nil ?
            RoundedRectangle(cornerRadius: 14)
                .stroke(Color.hf.warning.opacity(0.3), lineWidth: 1) :
            nil
        )
    }

    func formatAmount(_ amount: Double) -> String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        formatter.currencySymbol = currency.symbol
        formatter.maximumFractionDigits = currency == .jpy || currency == .krw ? 0 : 2
        return formatter.string(from: NSNumber(value: amount)) ?? "\(currency.symbol)0"
    }
}

// MARK: - AI Analysis Loading View

struct AIAnalysisLoadingView: View {
    var body: some View {
        HStack(spacing: 12) {
            ProgressView()
                .scaleEffect(0.8)

            Text("Analyzing your spending...")
                .font(.system(size: 14))
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity)
        .padding(16)
        .background(Color.hf.cardBackground)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

// MARK: - Generate AI Analysis Button

struct GenerateAIAnalysisButton: View {
    let action: () -> Void
    let isLoading: Bool

    var body: some View {
        HStack(spacing: 8) {
            if isLoading {
                ProgressView()
                    .scaleEffect(0.8)
            } else {
                Image(systemName: "brain.head.profile")
                    .font(.system(size: 14))
            }

            Text(isLoading ? "Analyzing..." : "Analyze Spending with AI")
                .font(.system(size: 14, weight: .medium))
        }
        .foregroundStyle(Color.hf.accent)
        .frame(maxWidth: .infinity)
        .padding(.vertical, 14)
        .background(Color.hf.accent.opacity(0.1))
        .clipShape(RoundedRectangle(cornerRadius: 12))
        .contentShape(Rectangle())
        .onTapGesture {
            if !isLoading {
                action()
            }
        }
    }
}

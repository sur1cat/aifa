import SwiftUI

// MARK: - Budget Forecast Content (for collapsible section)

struct BudgetForecastContent: View {
    let forecast: BudgetForecast
    let currency: Currency

    var body: some View {
        VStack(spacing: 12) {
            // Summary row
            HStack(spacing: 0) {
                ForecastStatItem(
                    title: "Expenses",
                    amount: forecast.projectedExpenses,
                    trend: forecast.expenseTrend,
                    color: Color.hf.expense,
                    currency: currency
                )

                Divider()
                    .frame(height: 36)

                ForecastStatItem(
                    title: "Income",
                    amount: forecast.projectedIncome,
                    trend: .stable,
                    color: Color.hf.income,
                    currency: currency
                )

                Divider()
                    .frame(height: 36)

                ForecastStatItem(
                    title: "Savings",
                    amount: forecast.projectedSavings,
                    trend: forecast.projectedSavings >= 0 ? .up : .down,
                    color: forecast.projectedSavings >= 0 ? Color.hf.income : Color.hf.expense,
                    currency: currency
                )
            }

            // Confidence
            HStack {
                Text(forecastMonthString)
                    .font(.system(size: 12))
                    .foregroundStyle(.secondary)

                Spacer()

                Text("\(Int(forecast.confidenceScore * 100))% confidence")
                    .font(.system(size: 12))
                    .foregroundStyle(.secondary)
            }
        }
    }

    private var forecastMonthString: String {
        let formatter = DateFormatter()
        formatter.dateFormat = "MMMM yyyy"
        return formatter.string(from: forecast.forecastMonth)
    }
}

// MARK: - Budget Forecast Card (standalone)

struct BudgetForecastCard: View {
    let forecast: BudgetForecast
    let currency: Currency
    @State private var isExpanded = false

    private var projectedBalance: Double {
        forecast.projectedIncome - forecast.projectedExpenses
    }

    var body: some View {
        VStack(spacing: 0) {
            // Header
            Button {
                withAnimation(.easeInOut(duration: 0.2)) {
                    isExpanded.toggle()
                }
            } label: {
                HStack {
                    VStack(alignment: .leading, spacing: 4) {
                        HStack(spacing: 6) {
                            Image(systemName: "chart.line.uptrend.xyaxis")
                                .font(.system(size: 14))
                                .foregroundStyle(Color.hf.accent)

                            Text("Next Month Forecast")
                                .font(.system(size: 14, weight: .semibold))
                                .foregroundStyle(.primary)
                        }

                        Text(forecastMonthString)
                            .font(.system(size: 12))
                            .foregroundStyle(.secondary)
                    }

                    Spacer()

                    // Confidence indicator
                    HStack(spacing: 4) {
                        Text("\(Int(forecast.confidenceScore * 100))%")
                            .font(.system(size: 12, weight: .medium))
                            .foregroundStyle(.secondary)

                        Image(systemName: isExpanded ? "chevron.up" : "chevron.down")
                            .font(.system(size: 12, weight: .semibold))
                            .foregroundStyle(.secondary)
                    }
                }
            }
            .buttonStyle(.plain)
            .padding(16)

            Divider()
                .padding(.horizontal, 16)

            // Summary
            HStack(spacing: 0) {
                ForecastStatItem(
                    title: "Expenses",
                    amount: forecast.projectedExpenses,
                    trend: forecast.expenseTrend,
                    color: Color.hf.expense,
                    currency: currency
                )

                Divider()
                    .frame(height: 40)

                ForecastStatItem(
                    title: "Income",
                    amount: forecast.projectedIncome,
                    trend: .stable,
                    color: Color.hf.income,
                    currency: currency
                )

                Divider()
                    .frame(height: 40)

                ForecastStatItem(
                    title: "Savings",
                    amount: forecast.projectedSavings,
                    trend: forecast.projectedSavings >= 0 ? .up : .down,
                    color: forecast.projectedSavings >= 0 ? Color.hf.income : Color.hf.expense,
                    currency: currency
                )
            }
            .padding(.vertical, 12)

            // Seasonal warnings
            if let factors = forecast.seasonalFactors, !factors.isEmpty {
                Divider()
                    .padding(.horizontal, 16)

                VStack(alignment: .leading, spacing: 8) {
                    ForEach(factors, id: \.category) { factor in
                        SeasonalWarningRow(factor: factor)
                    }
                }
                .padding(12)
            }

            // Expandable category breakdown
            if isExpanded && !forecast.categoryForecasts.isEmpty {
                Divider()
                    .padding(.horizontal, 16)

                VStack(spacing: 0) {
                    ForEach(forecast.categoryForecasts.sorted { $0.projectedAmount > $1.projectedAmount }) { category in
                        CategoryForecastRow(forecast: category, currency: currency)

                        if category.id != forecast.categoryForecasts.sorted(by: { $0.projectedAmount > $1.projectedAmount }).last?.id {
                            Divider()
                                .padding(.leading, 44)
                        }
                    }
                }
                .padding(.vertical, 8)
            }
        }
        .background(Color.hf.cardBackground)
        .clipShape(RoundedRectangle(cornerRadius: 16))
    }

    private var forecastMonthString: String {
        let formatter = DateFormatter()
        formatter.dateFormat = "MMMM yyyy"
        return formatter.string(from: forecast.forecastMonth)
    }
}

// MARK: - Forecast Stat Item

struct ForecastStatItem: View {
    let title: String
    let amount: Double
    let trend: TrendDirection
    let color: Color
    let currency: Currency

    var body: some View {
        VStack(spacing: 4) {
            Text(title)
                .font(.system(size: 11))
                .foregroundStyle(.secondary)

            HStack(spacing: 2) {
                Text(formatCurrency(amount))
                    .font(.system(size: 15, weight: .semibold, design: .rounded))
                    .foregroundStyle(color)

                if trend != .stable {
                    Image(systemName: trend == .up ? "arrow.up.right" : "arrow.down.right")
                        .font(.system(size: 10, weight: .bold))
                        .foregroundStyle(trend == .up ? Color.hf.expense : Color.hf.income)
                }
            }
        }
        .frame(maxWidth: .infinity)
    }

    private func formatCurrency(_ amount: Double) -> String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        formatter.currencySymbol = currency.symbol
        formatter.maximumFractionDigits = currency == .jpy || currency == .krw ? 0 : 0
        return formatter.string(from: NSNumber(value: abs(amount))) ?? "\(currency.symbol)0"
    }
}

// MARK: - Seasonal Warning Row

private struct SeasonalWarningRow: View {
    let factor: SeasonalFactor

    private var increasePercent: Int {
        Int((factor.monthlyMultiplier - 1.0) * 100)
    }

    var body: some View {
        HStack(spacing: 10) {
            Image(systemName: "exclamationmark.triangle.fill")
                .font(.system(size: 12))
                .foregroundStyle(Color.hf.warning)

            Text(factor.category)
                .font(.system(size: 13, weight: .medium))
                .foregroundStyle(.primary)

            Text(factor.reason)
                .font(.system(size: 12))
                .foregroundStyle(.secondary)
                .lineLimit(1)

            Spacer()

            if increasePercent > 0 {
                Text("+\(increasePercent)%")
                    .font(.system(size: 13, weight: .medium))
                    .foregroundStyle(Color.hf.warning)
            }
        }
    }
}

// MARK: - Category Forecast Row

private struct CategoryForecastRow: View {
    let forecast: CategoryForecast
    let currency: Currency

    private var categoryInfo: TransactionCategory {
        TransactionCategory(rawValue: forecast.category) ?? .other
    }

    private var changePercent: Double {
        guard forecast.historicalAverage > 0 else { return 0 }
        return ((forecast.projectedAmount - forecast.historicalAverage) / forecast.historicalAverage) * 100
    }

    var body: some View {
        HStack(spacing: 12) {
            ZStack {
                Circle()
                    .fill(categoryInfo.color.opacity(0.15))
                    .frame(width: 32, height: 32)

                Image(systemName: categoryInfo.icon)
                    .font(.system(size: 14))
                    .foregroundStyle(categoryInfo.color)
            }

            VStack(alignment: .leading, spacing: 2) {
                Text(categoryInfo.title)
                    .font(.system(size: 14, weight: .medium))

                if forecast.recurringAmount > 0 {
                    Text("incl. \(formatCurrency(forecast.recurringAmount)) recurring")
                        .font(.system(size: 11))
                        .foregroundStyle(.secondary)
                }
            }

            Spacer()

            VStack(alignment: .trailing, spacing: 2) {
                Text(formatCurrency(forecast.projectedAmount))
                    .font(.system(size: 14, weight: .semibold, design: .rounded))

                if abs(changePercent) > 5 {
                    HStack(spacing: 2) {
                        Image(systemName: forecast.trend == .up ? "arrow.up" : "arrow.down")
                            .font(.system(size: 9, weight: .bold))
                        Text("\(Int(abs(changePercent)))%")
                            .font(.system(size: 11, weight: .medium))
                    }
                    .foregroundStyle(forecast.trend == .up ? Color.hf.expense : Color.hf.income)
                }
            }
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 10)
    }

    private func formatCurrency(_ amount: Double) -> String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        formatter.currencySymbol = currency.symbol
        formatter.maximumFractionDigits = 0
        return formatter.string(from: NSNumber(value: amount)) ?? "\(currency.symbol)0"
    }
}

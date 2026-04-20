import SwiftUI
import os

struct BudgetView: View {
    @EnvironmentObject var dataManager: DataManager
    @EnvironmentObject var authManager: AuthManager
    @Environment(\.colorScheme) var colorScheme
    @State private var selectedDate = Date()
    @State private var addTransactionType: TransactionType?
    @State private var showVoiceInput = false
    @State private var showReceiptScanner = false
    @State private var selectedTransaction: Transaction?
    @State private var showSavingsGoalSheet = false
    @State private var showAddRecurringSheet = false
    @State private var showAIChat = false
    @State private var showDeleteSavingsGoalAlert = false

    // Collapsible section states (persisted)
    @AppStorage("budget_recurring_expanded") private var recurringExpanded = true
    @AppStorage("budget_forecast_expanded") private var forecastExpanded = true
    @AppStorage("budget_savings_expanded") private var savingsExpanded = true
    @AppStorage("budget_transactions_expanded") private var transactionsExpanded = true

    private let calendar = Calendar.current

    private var isCurrentMonth: Bool {
        calendar.isDate(selectedDate, equalTo: Date(), toGranularity: .month)
    }

    private var transactionsForSelectedMonth: [Transaction] {
        dataManager.transactions.filter {
            calendar.isDate($0.date, equalTo: selectedDate, toGranularity: .month)
        }.sorted { $0.date > $1.date }
    }

    private var monthlyIncome: Double {
        transactionsForSelectedMonth
            .filter { $0.type == .income }
            .reduce(0) { $0 + $1.amount }
    }

    private var monthlyExpenses: Double {
        transactionsForSelectedMonth
            .filter { $0.type == .expense }
            .reduce(0) { $0 + $1.amount }
    }

    private var monthlyBalance: Double {
        monthlyIncome - monthlyExpenses
    }

    var body: some View {
        NavigationStack {
            List {
                // Tagline
                Section {
                    Text("control your money daily")
                        .font(.system(size: 15))
                        .foregroundStyle(.secondary)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .listRowBackground(Color.clear)
                        .listRowInsets(EdgeInsets(top: 0, leading: 16, bottom: 8, trailing: 16))
                }
                .listRowSeparator(.hidden)

                // Month Selector
                Section {
                    BudgetMonthSelector(selectedDate: $selectedDate)
                        .listRowBackground(Color.clear)
                        .listRowInsets(EdgeInsets(top: 0, leading: 8, bottom: 12, trailing: 8))
                }
                .listRowSeparator(.hidden)

                // Insights Carousel (WHOOP-style)
                if !dataManager.insights(for: .budget).isEmpty {
                    Section {
                        InsightCarousel(
                            insights: dataManager.insights(for: .budget),
                            onDismiss: { insight in
                                dataManager.dismissInsight(insight)
                            },
                            onAction: { action in
                                if action.actionType == "openAIChat" {
                                    showAIChat = true
                                }
                            }
                        )
                        .listRowBackground(Color.clear)
                        .listRowInsets(EdgeInsets(top: 0, leading: 0, bottom: 8, trailing: 0))
                    }
                    .listRowSeparator(.hidden)
                }

                if dataManager.isLoading && dataManager.transactions.isEmpty {
                    Section {
                        ProgressView()
                            .frame(maxWidth: .infinity)
                            .padding(.top, 60)
                            .listRowBackground(Color.clear)
                    }
                    .listRowSeparator(.hidden)
                } else {
                    // Balance Card
                    Section {
                        BudgetBalanceCard(
                            balance: monthlyBalance,
                            income: monthlyIncome,
                            expenses: monthlyExpenses,
                            currency: dataManager.profile.currency,
                            isCurrentMonth: isCurrentMonth
                        )
                        .listRowBackground(Color.clear)
                        .listRowInsets(EdgeInsets(top: 0, leading: 16, bottom: 12, trailing: 16))
                    }
                    .listRowSeparator(.hidden)

                    // Monthly Stats
                    Section {
                        HStack(spacing: 12) {
                            StatCard(
                                title: "Income",
                                amount: monthlyIncome,
                                color: Color.hf.income,
                                currency: dataManager.profile.currency
                            )
                            StatCard(
                                title: "Expenses",
                                amount: monthlyExpenses,
                                color: Color.hf.expense,
                                currency: dataManager.profile.currency
                            )
                        }
                        .listRowBackground(Color.clear)
                        .listRowInsets(EdgeInsets(top: 0, leading: 16, bottom: 12, trailing: 16))
                    }
                    .listRowSeparator(.hidden)

                    // Recurring Transactions (only show in current month)
                    if isCurrentMonth {
                        // Recurring Section (Collapsible)
                        Section {
                            CollapsibleSection(
                                title: "Regular Expenses",
                                icon: "repeat.circle.fill",
                                iconColor: Color.hf.accent,
                                isExpanded: $recurringExpanded,
                                actionIcon: "plus.circle.fill",
                                onAction: { showAddRecurringSheet = true }
                            ) {
                                RecurringSectionContent()
                                    .padding(.horizontal, 16)
                                    .padding(.bottom, 12)
                            }
                            .listRowBackground(Color.clear)
                            .listRowInsets(EdgeInsets(top: 0, leading: 16, bottom: 8, trailing: 16))
                        }
                        .listRowSeparator(.hidden)

                        // Budget Forecast (Collapsible)
                        if let forecast = dataManager.currentForecast {
                            Section {
                                CollapsibleSection(
                                    title: "Next Month Forecast",
                                    icon: "chart.line.uptrend.xyaxis",
                                    iconColor: Color.hf.info,
                                    isExpanded: $forecastExpanded
                                ) {
                                    BudgetForecastContent(
                                        forecast: forecast,
                                        currency: dataManager.profile.currency
                                    )
                                    .padding(.horizontal, 16)
                                    .padding(.bottom, 12)
                                }
                                .listRowBackground(Color.clear)
                                .listRowInsets(EdgeInsets(top: 0, leading: 16, bottom: 8, trailing: 16))
                            }
                            .listRowSeparator(.hidden)
                        }

                        // Savings Goal (Collapsible)
                        Section {
                            CollapsibleSection(
                                title: "Savings Goal",
                                icon: "target",
                                iconColor: Color.hf.income,
                                isExpanded: $savingsExpanded
                            ) {
                                if let goal = dataManager.savingsGoal {
                                    SavingsGoalContent(
                                        goal: goal,
                                        currency: dataManager.profile.currency,
                                        onEdit: { showSavingsGoalSheet = true },
                                        onDelete: { showDeleteSavingsGoalAlert = true }
                                    )
                                    .padding(.horizontal, 16)
                                    .padding(.bottom, 12)
                                } else {
                                    SetSavingsGoalButton { showSavingsGoalSheet = true }
                                        .padding(.horizontal, 16)
                                        .padding(.bottom, 12)
                                }
                            }
                            .listRowBackground(Color.clear)
                            .listRowInsets(EdgeInsets(top: 0, leading: 16, bottom: 8, trailing: 16))
                        }
                        .listRowSeparator(.hidden)

                    }

                    // Transactions (Collapsible)
                    if transactionsForSelectedMonth.isEmpty {
                        Section {
                            BudgetEmptyState(isCurrentMonth: isCurrentMonth)
                                .listRowBackground(Color.clear)
                                .listRowInsets(EdgeInsets(top: 20, leading: 16, bottom: 0, trailing: 16))
                        }
                        .listRowSeparator(.hidden)
                    } else {
                        Section {
                            VStack(spacing: 0) {
                                // Collapsible header
                                Button {
                                    withAnimation(.easeInOut(duration: 0.2)) {
                                        transactionsExpanded.toggle()
                                    }
                                } label: {
                                    HStack(spacing: 10) {
                                        Image(systemName: "creditcard.fill")
                                            .font(.system(size: 14))
                                            .foregroundStyle(Color.hf.expense)
                                            .frame(width: 20)

                                        Text("Transactions")
                                            .font(.system(size: 15, weight: .semibold))
                                            .foregroundStyle(.primary)

                                        Text("(\(transactionsForSelectedMonth.count))")
                                            .font(.system(size: 13))
                                            .foregroundStyle(.secondary)

                                        Spacer()

                                        Image(systemName: transactionsExpanded ? "chevron.up" : "chevron.down")
                                            .font(.system(size: 12, weight: .semibold))
                                            .foregroundStyle(.secondary)
                                    }
                                    .padding(.vertical, 14)
                                    .padding(.horizontal, 16)
                                    .background(Color.hf.cardBackground)
                                }
                                .buttonStyle(.plain)

                                if transactionsExpanded {
                                    LazyVStack(spacing: 8) {
                                        ForEach(transactionsForSelectedMonth) { transaction in
                                            transactionRowView(for: transaction)
                                                .onTapGesture {
                                                    selectedTransaction = transaction
                                                }
                                        }
                                    }
                                    .padding(.horizontal, 16)
                                    .padding(.vertical, 8)
                                    .background(Color.hf.cardBackground)
                                }
                            }
                            .clipShape(RoundedRectangle(cornerRadius: 14))
                            .listRowBackground(Color.clear)
                            .listRowInsets(EdgeInsets(top: 0, leading: 16, bottom: 8, trailing: 16))
                        }
                        .listRowSeparator(.hidden)
                    }
                }
            }
            .listStyle(.plain)
            .scrollContentBackground(.hidden)
            .background(AppTheme.appBackground(for: colorScheme))
            .navigationTitle("Atoma Budget")
            .onAppear {
                dataManager.generateInsights(for: .budget)
                if isCurrentMonth {
                    _ = dataManager.generateBudgetForecast()
                }
            }
            .toolbar {
                ToolbarItem(placement: .topBarLeading) {
                    HStack(spacing: 16) {
                        Button {
                            showAIChat = true
                        } label: {
                            Image(systemName: "atom")
                                .fontWeight(.semibold)
                        }
                        .tint(Color.hf.income)

                        Button {
                            showVoiceInput = true
                        } label: {
                            Image(systemName: "mic.fill")
                                .fontWeight(.semibold)
                        }
                        .tint(Color.hf.warning)

                        Button {
                            showReceiptScanner = true
                        } label: {
                            Image(systemName: "doc.text.viewfinder")
                                .fontWeight(.semibold)
                        }
                        .tint(Color.hf.info)
                    }
                }
                ToolbarItem(placement: .topBarTrailing) {
                    Menu {
                        Button {
                            addTransactionType = .income
                        } label: {
                            Label("Add Income", systemImage: "plus.circle")
                        }
                        Button {
                            addTransactionType = .expense
                        } label: {
                            Label("Add Expense", systemImage: "minus.circle")
                        }
                    } label: {
                        Image(systemName: "plus")
                            .fontWeight(.semibold)
                    }
                    .tint(.primary)
                }
            }
            .sheet(item: $addTransactionType) { type in
                AddTransactionSheet(type: type, initialDate: selectedDate)
            }
            .sheet(item: $selectedTransaction) { transaction in
                EditTransactionSheet(transaction: transaction)
            }
            .sheet(isPresented: $showVoiceInput) {
                VoiceInputView()
            }
            .sheet(isPresented: $showReceiptScanner) {
                ReceiptScannerView()
            }
            .sheet(isPresented: $showSavingsGoalSheet) {
                SavingsGoalSheet(existingGoal: dataManager.savingsGoal)
            }
            .sheet(isPresented: $showAddRecurringSheet) {
                AddRecurringSheet()
            }
            .fullScreenCover(isPresented: $showAIChat) {
                AIChatView(agent: .financeAdvisor) {
                    buildBudgetContext()
                }
            }
            .task {
                if authManager.isAuthenticated {
                    await dataManager.syncTransactions()
                    await dataManager.syncRecurringTransactions()
                    await dataManager.syncSavingsGoal()
                }
            }
            .refreshable {
                if authManager.isAuthenticated {
                    await dataManager.syncTransactions()
                    await dataManager.syncRecurringTransactions()
                    await dataManager.syncSavingsGoal()
                }
            }
            .alert("Delete Savings Goal?", isPresented: $showDeleteSavingsGoalAlert) {
                Button("Cancel", role: .cancel) {}
                Button("Delete", role: .destructive) {
                    dataManager.deleteSavingsGoal()
                }
            } message: {
                Text("Your savings goal will be removed.")
            }
        }
    }

    // MARK: - Helper Methods

    @ViewBuilder
    private func transactionRowView(for transaction: Transaction) -> some View {
        let questionable = dataManager.aiExpenseAnalysis?.questionableTransactions.first {
            $0.transactionId == transaction.id
        }

        if questionable != nil {
            TransactionRowWithAnalysis(
                transaction: transaction,
                currency: dataManager.profile.currency,
                questionable: questionable
            )
        } else {
            TransactionRow(transaction: transaction, currency: dataManager.profile.currency)
        }
    }

    // MARK: - AI Context

    private func buildBudgetContext() -> String {
        let currency = dataManager.profile.currency.symbol
        let dateFormatter = DateFormatter()
        dateFormatter.dateFormat = "MMMM yyyy"

        var context = "=== USER'S FINANCIAL DATA ===\n"
        context += "Month: \(dateFormatter.string(from: selectedDate))\n"
        context += "Currency: \(dataManager.profile.currency.rawValue)\n\n"

        // Monthly summary
        context += "MONTHLY SUMMARY:\n"
        context += "- Income: \(currency)\(String(format: "%.2f", monthlyIncome))\n"
        context += "- Expenses: \(currency)\(String(format: "%.2f", monthlyExpenses))\n"
        context += "- Balance: \(currency)\(String(format: "%.2f", monthlyBalance))\n\n"

        // Expense breakdown by category
        let expenses = transactionsForSelectedMonth.filter { $0.type == .expense }
        var categoryTotals: [String: Double] = [:]
        for expense in expenses {
            let cat = expense.category.isEmpty ? "Other" : expense.category
            categoryTotals[cat, default: 0] += expense.amount
        }

        if !categoryTotals.isEmpty {
            context += "EXPENSES BY CATEGORY:\n"
            let sorted = categoryTotals.sorted { $0.value > $1.value }
            for (category, total) in sorted {
                let percent = monthlyExpenses > 0 ? Int(total / monthlyExpenses * 100) : 0
                context += "- \(category): \(currency)\(String(format: "%.2f", total)) (\(percent)%)\n"
            }
            context += "\n"
        }

        // Recent transactions (last 10)
        context += "RECENT TRANSACTIONS:\n"
        for transaction in transactionsForSelectedMonth.prefix(10) {
            let sign = transaction.type == .income ? "+" : "-"
            let dateStr = DateFormatter.localizedString(from: transaction.date, dateStyle: .short, timeStyle: .none)
            let cat = transaction.category.isEmpty ? "" : " [\(transaction.category)]"
            context += "- \(dateStr): \(sign)\(currency)\(String(format: "%.2f", transaction.amount)) - \(transaction.title)\(cat)\n"
        }

        // Recurring expenses
        let recurring = dataManager.recurringTransactions
        if !recurring.isEmpty {
            context += "\nRECURRING EXPENSES:\n"
            var monthlyRecurring: Double = 0
            for r in recurring {
                let monthlyAmount: Double
                switch r.frequency {
                case .weekly: monthlyAmount = r.amount * 4
                case .biweekly: monthlyAmount = r.amount * 2
                case .monthly: monthlyAmount = r.amount
                case .quarterly: monthlyAmount = r.amount / 3
                case .yearly: monthlyAmount = r.amount / 12
                }
                monthlyRecurring += monthlyAmount
                context += "- \(r.title): \(currency)\(String(format: "%.2f", r.amount)) (\(r.frequency.rawValue))\n"
            }
            context += "Total monthly recurring: \(currency)\(String(format: "%.2f", monthlyRecurring))\n"
        }

        // Savings goal
        if let goal = dataManager.savingsGoal {
            context += "\nSAVINGS GOAL:\n"
            context += "- Target: \(currency)\(String(format: "%.2f", goal.monthlyTarget))\n"
            context += "- Current savings: \(currency)\(String(format: "%.2f", goal.currentSavings))\n"
            context += "- Progress: \(Int(goal.progress * 100))%\n"
        }

        return context
    }
}

// MARK: - Budget Month Selector

struct BudgetMonthSelector: View {
    @Binding var selectedDate: Date
    @State private var showMonthPicker = false

    private let calendar = Calendar.current

    private var months: [Date] {
        (-2...2).compactMap { offset in
            calendar.date(byAdding: .month, value: offset, to: selectedDate)
        }
    }

    var body: some View {
        VStack(spacing: 8) {
            // Year header - tap to open picker
            Button {
                showMonthPicker = true
            } label: {
                HStack(spacing: 4) {
                    Text(yearString)
                        .font(.system(size: 14, weight: .medium))
                    Image(systemName: "calendar")
                        .font(.system(size: 12))
                }
                .foregroundStyle(.secondary)
            }

            HStack(spacing: 6) {
                // Left arrow
                Button {
                    withAnimation {
                        if let newDate = calendar.date(byAdding: .month, value: -1, to: selectedDate) {
                            selectedDate = newDate
                        }
                    }
                } label: {
                    Image(systemName: "chevron.left")
                        .font(.system(size: 14, weight: .semibold))
                        .foregroundStyle(.secondary)
                        .frame(width: 24, height: 56)
                }

                ForEach(months, id: \.self) { date in
                    BudgetMonthCell(
                        date: date,
                        isSelected: calendar.isDate(date, equalTo: selectedDate, toGranularity: .month),
                        isCurrentMonth: calendar.isDate(date, equalTo: Date(), toGranularity: .month)
                    ) {
                        withAnimation(.easeInOut(duration: 0.2)) {
                            selectedDate = date
                        }
                    }
                }

                // Right arrow
                Button {
                    withAnimation {
                        if let newDate = calendar.date(byAdding: .month, value: 1, to: selectedDate) {
                            selectedDate = newDate
                        }
                    }
                } label: {
                    Image(systemName: "chevron.right")
                        .font(.system(size: 14, weight: .semibold))
                        .foregroundStyle(.secondary)
                        .frame(width: 24, height: 56)
                }
            }
            .padding(.horizontal, 4)

            // This month button if not viewing current month
            if !calendar.isDate(selectedDate, equalTo: Date(), toGranularity: .month) {
                Button {
                    withAnimation {
                        selectedDate = Date()
                    }
                } label: {
                    Text("This Month")
                        .font(.system(size: 12, weight: .medium))
                        .foregroundStyle(Color.hf.accent)
                }
            }
        }
        .sheet(isPresented: $showMonthPicker) {
            BudgetMonthPickerSheet(selectedDate: $selectedDate)
        }
    }

    private var yearString: String {
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy"
        return formatter.string(from: selectedDate)
    }
}

struct BudgetMonthPickerSheet: View {
    @Binding var selectedDate: Date
    @Environment(\.dismiss) var dismiss

    var body: some View {
        NavigationStack {
            DatePicker(
                "Select Month",
                selection: $selectedDate,
                displayedComponents: .date
            )
            .datePickerStyle(.graphical)
            .padding()
            .navigationTitle("Select Month")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .confirmationAction) {
                    Button("Done") {
                        dismiss()
                    }
                }
            }
        }
        .presentationDetents([.medium])
    }
}

struct BudgetMonthCell: View {
    let date: Date
    let isSelected: Bool
    let isCurrentMonth: Bool
    let action: () -> Void

    private let calendar = Calendar.current

    var body: some View {
        Button(action: action) {
            VStack(spacing: 4) {
                Text(monthAbbrev)
                    .font(.system(size: 13, weight: isSelected ? .bold : .semibold))
                    .foregroundStyle(isSelected ? .white : (isCurrentMonth ? Color.hf.accent : .primary))
            }
            .frame(width: 56, height: 44)
            .background(isSelected ? Color.hf.accent : Color.hf.cardBackground)
            .clipShape(RoundedRectangle(cornerRadius: 12))
            .overlay(
                RoundedRectangle(cornerRadius: 12)
                    .stroke(isCurrentMonth && !isSelected ? Color.hf.accent : Color.clear, lineWidth: 2)
            )
        }
        .buttonStyle(.plain)
    }

    private var monthAbbrev: String {
        let formatter = DateFormatter()
        formatter.dateFormat = "MMM"
        return formatter.string(from: date)
    }
}

// MARK: - Budget Balance Card

struct BudgetBalanceCard: View {
    let balance: Double
    let income: Double
    let expenses: Double
    let currency: Currency
    let isCurrentMonth: Bool

    private var savingsRate: Double {
        guard income > 0 else { return 0 }
        return max(0, (income - expenses) / income * 100)
    }

    var body: some View {
        VStack(spacing: 16) {
            // Balance
            VStack(spacing: 4) {
                Text(isCurrentMonth ? "This Month" : "Monthly Balance")
                    .font(.system(size: 13))
                    .foregroundStyle(.secondary)

                Text(formatCurrency(balance))
                    .font(.system(size: 32, weight: .bold, design: .rounded))
                    .foregroundStyle(balance >= 0 ? Color.hf.income : Color.hf.expense)
            }

            // Savings rate indicator
            if income > 0 {
                HStack(spacing: 8) {
                    // Progress bar
                    GeometryReader { geometry in
                        ZStack(alignment: .leading) {
                            RoundedRectangle(cornerRadius: 4)
                                .fill(Color.hf.accent.opacity(0.2))
                                .frame(height: 8)

                            RoundedRectangle(cornerRadius: 4)
                                .fill(savingsRate > 0 ? Color.hf.income : Color.hf.expense)
                                .frame(width: geometry.size.width * min(savingsRate / 100, 1), height: 8)
                                .animation(.easeInOut(duration: 0.3), value: savingsRate)
                        }
                    }
                    .frame(height: 8)

                    Text("\(Int(savingsRate))%")
                        .font(.system(size: 13, weight: .semibold, design: .rounded))
                        .foregroundStyle(savingsRate > 0 ? Color.hf.income : Color.hf.expense)
                        .frame(width: 44, alignment: .trailing)
                }

                Text(savingsRate > 0 ? "saved this month" : "over budget")
                    .font(.system(size: 12))
                    .foregroundStyle(.secondary)
            }
        }
        .padding(20)
        .frame(maxWidth: .infinity)
        .background(Color.hf.cardBackground)
        .clipShape(RoundedRectangle(cornerRadius: 20))
    }

    private func formatCurrency(_ amount: Double) -> String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        formatter.currencySymbol = currency.symbol
        formatter.maximumFractionDigits = currency == .jpy || currency == .krw ? 0 : 2
        return formatter.string(from: NSNumber(value: amount)) ?? "\(currency.symbol)0"
    }
}

// MARK: - Budget Empty State

struct BudgetEmptyState: View {
    let isCurrentMonth: Bool

    var body: some View {
        VStack(spacing: 12) {
            Image(systemName: isCurrentMonth ? "creditcard" : "calendar")
                .font(.system(size: 40))
                .foregroundStyle(.secondary)

            Text(isCurrentMonth ? "No transactions yet" : "No transactions this month")
                .font(.system(size: 17, weight: .semibold))
                .foregroundStyle(.primary)

            Text(isCurrentMonth ? "Tap + to add income or expense" : "Transactions will appear here")
                .font(.system(size: 14))
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity)
        .padding(.vertical, 40)
    }
}

// MARK: - Stat Card

struct StatCard: View {
    let title: LocalizedStringKey
    let amount: Double
    let color: Color
    let currency: Currency

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(title)
                .font(.system(size: 13))
                .foregroundStyle(.secondary)
            Text(formatCurrency(amount))
                .font(.system(size: 20, weight: .semibold, design: .rounded))
                .foregroundStyle(color)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding()
        .background(Color.hf.cardBackground)
        .clipShape(RoundedRectangle(cornerRadius: 16))
    }

    func formatCurrency(_ amount: Double) -> String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        formatter.currencySymbol = currency.symbol
        formatter.maximumFractionDigits = currency == .jpy || currency == .krw ? 0 : 2
        return formatter.string(from: NSNumber(value: amount)) ?? "\(currency.symbol)0"
    }
}

// MARK: - Transaction Row

struct TransactionRow: View {
    let transaction: Transaction
    let currency: Currency

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
                Text(transaction.title)
                    .font(.system(size: 16, weight: .medium))
                Text(transaction.date.formatted(date: .abbreviated, time: .omitted))
                    .font(.system(size: 13))
                    .foregroundStyle(.secondary)
            }

            Spacer()

            Text("\(transaction.type == .income ? "+" : "-")\(formatAmount(transaction.amount))")
                .font(.system(size: 16, weight: .semibold, design: .rounded))
                .foregroundStyle(transaction.type == .income ? Color.hf.income : Color.hf.expense)
        }
        .padding()
        .background(Color.hf.cardBackground)
        .clipShape(RoundedRectangle(cornerRadius: 14))
    }

    func formatAmount(_ amount: Double) -> String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        formatter.currencySymbol = currency.symbol
        formatter.maximumFractionDigits = currency == .jpy || currency == .krw ? 0 : 2
        return formatter.string(from: NSNumber(value: amount)) ?? "\(currency.symbol)0"
    }
}

// MARK: - Add Transaction Sheet

struct AddTransactionSheet: View {
    @EnvironmentObject var dataManager: DataManager
    @Environment(\.dismiss) var dismiss
    let type: TransactionType
    @State private var title = ""
    @State private var amount = ""
    @State private var selectedCategory: TransactionCategory
    @State private var date: Date

    init(type: TransactionType, initialDate: Date = Date()) {
        self.type = type
        _selectedCategory = State(initialValue: type == .income ? .salary : .food)
        _date = State(initialValue: initialDate)
    }

    private var availableCategories: [TransactionCategory] {
        type == .income ? TransactionCategory.incomeCategories : TransactionCategory.expenseCategories
    }

    private var titleError: String? {
        if title.isEmpty { return nil }
        if title.trimmingCharacters(in: .whitespaces).count < 2 {
            return "Title must be at least 2 characters"
        }
        if title.count > 100 {
            return "Title too long (max 100 characters)"
        }
        return nil
    }

    private var amountError: String? {
        if amount.isEmpty { return nil }
        guard let value = Double(amount.replacingOccurrences(of: ",", with: ".")) else {
            return "Invalid amount"
        }
        if value <= 0 {
            return "Amount must be positive"
        }
        if value > 999999999 {
            return "Amount too large"
        }
        return nil
    }

    private var isValid: Bool {
        let trimmedTitle = title.trimmingCharacters(in: .whitespaces)
        guard trimmedTitle.count >= 2 else { return false }
        guard let value = Double(amount.replacingOccurrences(of: ",", with: ".")) else { return false }
        return value > 0 && value <= 999999999
    }

    var body: some View {
        NavigationStack {
            Form {
                Section {
                    TextField("Description", text: $title)
                    if let error = titleError {
                        Text(error)
                            .font(.caption)
                            .foregroundStyle(.red)
                    }
                }

                Section {
                    HStack {
                        Text(dataManager.profile.currency.symbol)
                            .foregroundStyle(.secondary)
                        TextField("Amount", text: $amount)
                            .keyboardType(.decimalPad)
                    }
                    if let error = amountError {
                        Text(error)
                            .font(.caption)
                            .foregroundStyle(.red)
                    }
                }

                Section {
                    DatePicker("Date", selection: $date, displayedComponents: .date)
                }

                Section("Category") {
                    LazyVGrid(columns: [GridItem(.adaptive(minimum: 80))], spacing: 12) {
                        ForEach(availableCategories, id: \.self) { category in
                            CategoryButton(
                                category: category,
                                isSelected: selectedCategory == category
                            ) {
                                selectedCategory = category
                            }
                        }
                    }
                    .padding(.vertical, 8)
                }
            }
            .navigationTitle(type == .income ? "Add Income" : "Add Expense")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Add") {
                        if let amountValue = Double(amount.replacingOccurrences(of: ",", with: ".")) {
                            let transaction = Transaction(
                                title: title.trimmingCharacters(in: .whitespaces),
                                amount: amountValue,
                                type: type,
                                category: selectedCategory.rawValue,
                                date: date
                            )
                            dataManager.addTransaction(transaction)
                            HapticManager.completionSuccess()
                            dismiss()
                        }
                    }
                    .disabled(!isValid)
                }
            }
        }
        .presentationDetents([.large])
    }
}

// MARK: - Category Button

struct CategoryButton: View {
    let category: TransactionCategory
    let isSelected: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            VStack(spacing: 6) {
                ZStack {
                    Circle()
                        .fill(isSelected ? category.color : category.color.opacity(0.15))
                        .frame(width: 44, height: 44)

                    Image(systemName: category.icon)
                        .font(.system(size: 18))
                        .foregroundStyle(isSelected ? .white : category.color)
                }

                Text(category.title)
                    .font(.system(size: 11))
                    .foregroundStyle(isSelected ? .primary : .secondary)
                    .lineLimit(1)
            }
        }
        .buttonStyle(.plain)
    }
}

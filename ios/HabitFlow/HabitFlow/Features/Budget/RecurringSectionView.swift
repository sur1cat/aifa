import SwiftUI

// MARK: - Recurring Section Content (for CollapsibleSection, no header)

struct RecurringSectionContent: View {
    @EnvironmentObject var dataManager: DataManager
    @State private var selectedRecurring: RecurringTransaction?

    private var activeTransactions: [RecurringTransaction] {
        dataManager.recurringTransactions.filter { $0.isActive }
    }

    private var groupedTransactions: [(category: RecurringCategory, items: [RecurringTransaction])] {
        let grouped = Dictionary(grouping: activeTransactions) { $0.category }
        return RecurringCategory.allCases.compactMap { category in
            guard let items = grouped[category], !items.isEmpty else { return nil }
            return (category: category, items: items)
        }
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            if activeTransactions.isEmpty {
                VStack(spacing: 8) {
                    Image(systemName: "repeat.circle")
                        .font(.system(size: 32))
                        .foregroundStyle(.tertiary)
                    Text("No regular expenses")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                    Text("Tap + to add subscriptions, bills, etc.")
                        .font(.caption)
                        .foregroundStyle(.tertiary)
                }
                .frame(maxWidth: .infinity)
                .padding(.vertical, 20)
            } else {
                // Projected expenses
                HStack {
                    VStack(alignment: .leading, spacing: 4) {
                        Text("Monthly projection")
                            .font(.system(size: 12))
                            .foregroundStyle(.secondary)
                        Text(dataManager.formatCurrency(dataManager.projectedMonthlyExpenses))
                            .font(.system(size: 18, weight: .semibold, design: .rounded))
                            .foregroundStyle(Color.hf.expense)
                    }
                    Spacer()
                }
                .padding()
                .background(Color.hf.cardBackground)
                .clipShape(RoundedRectangle(cornerRadius: 12))

                // Grouped by category
                ForEach(groupedTransactions, id: \.category) { group in
                    VStack(alignment: .leading, spacing: 8) {
                        // Category header
                        HStack(spacing: 6) {
                            Image(systemName: group.category.icon)
                                .font(.system(size: 12))
                                .foregroundStyle(Color.hf.accent)
                            Text(group.category.title)
                                .font(.system(size: 12, weight: .medium))
                                .foregroundStyle(.secondary)
                        }
                        .padding(.leading, 4)

                        // Items in category
                        ForEach(group.items) { recurring in
                            RecurringRow(recurring: recurring, currency: dataManager.profile.currency)
                                .onTapGesture {
                                    selectedRecurring = recurring
                                }
                        }
                    }
                }
            }
        }
        .sheet(item: $selectedRecurring) { recurring in
            EditRecurringSheet(recurring: recurring)
        }
    }
}

// MARK: - Recurring Section View (legacy, with header)

struct RecurringSectionView: View {
    @EnvironmentObject var dataManager: DataManager
    @State private var showAddSheet = false
    @State private var selectedRecurring: RecurringTransaction?

    private var activeTransactions: [RecurringTransaction] {
        dataManager.recurringTransactions.filter { $0.isActive }
    }

    private var groupedTransactions: [(category: RecurringCategory, items: [RecurringTransaction])] {
        let grouped = Dictionary(grouping: activeTransactions) { $0.category }
        return RecurringCategory.allCases.compactMap { category in
            guard let items = grouped[category], !items.isEmpty else { return nil }
            return (category: category, items: items)
        }
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Text("Regular Expenses")
                    .font(.system(size: 14, weight: .medium))
                    .foregroundStyle(.secondary)

                Spacer()

                Button {
                    showAddSheet = true
                } label: {
                    Image(systemName: "plus.circle.fill")
                        .font(.system(size: 18))
                        .foregroundStyle(Color.hf.accent)
                }
            }
            .padding(.horizontal, 4)

            if dataManager.recurringTransactions.isEmpty {
                VStack(spacing: 8) {
                    Image(systemName: "repeat.circle")
                        .font(.system(size: 32))
                        .foregroundStyle(.tertiary)
                    Text("No regular expenses")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                }
                .frame(maxWidth: .infinity)
                .padding(.vertical, 24)
            } else {
                // Projected expenses
                HStack {
                    VStack(alignment: .leading, spacing: 4) {
                        Text("Monthly projection")
                            .font(.system(size: 12))
                            .foregroundStyle(.secondary)
                        Text(dataManager.formatCurrency(dataManager.projectedMonthlyExpenses))
                            .font(.system(size: 18, weight: .semibold, design: .rounded))
                            .foregroundStyle(Color.hf.expense)
                    }
                    Spacer()
                }
                .padding()
                .background(Color.hf.cardBackground)
                .clipShape(RoundedRectangle(cornerRadius: 12))

                // Grouped by category
                ForEach(groupedTransactions, id: \.category) { group in
                    VStack(alignment: .leading, spacing: 8) {
                        // Category header
                        HStack(spacing: 6) {
                            Image(systemName: group.category.icon)
                                .font(.system(size: 12))
                                .foregroundStyle(Color.hf.accent)
                            Text(group.category.title)
                                .font(.system(size: 12, weight: .medium))
                                .foregroundStyle(.secondary)
                        }
                        .padding(.leading, 4)

                        // Items in category
                        ForEach(group.items) { recurring in
                            RecurringRow(recurring: recurring, currency: dataManager.profile.currency)
                                .onTapGesture {
                                    selectedRecurring = recurring
                                }
                        }
                    }
                }
            }
        }
        .sheet(isPresented: $showAddSheet) {
            AddRecurringSheet()
        }
        .sheet(item: $selectedRecurring) { recurring in
            EditRecurringSheet(recurring: recurring)
        }
    }
}

struct RecurringRow: View {
    let recurring: RecurringTransaction
    let currency: Currency

    var body: some View {
        HStack {
            VStack(alignment: .leading, spacing: 4) {
                Text(recurring.title)
                    .font(.system(size: 16, weight: .medium))

                HStack(spacing: 8) {
                    Text(recurring.frequency.title)
                        .font(.system(size: 11))
                        .padding(.horizontal, 6)
                        .padding(.vertical, 2)
                        .background(Color.hf.accent.opacity(0.15))
                        .foregroundStyle(Color.hf.accent)
                        .clipShape(Capsule())

                    Text("Next: \(recurring.nextDate.formatted(date: .abbreviated, time: .omitted))")
                        .font(.system(size: 12))
                        .foregroundStyle(.secondary)

                    if let remaining = recurring.remainingPayments {
                        Text("\(remaining) left")
                            .font(.system(size: 11))
                            .padding(.horizontal, 6)
                            .padding(.vertical, 2)
                            .background(Color.hf.warning.opacity(0.15))
                            .foregroundStyle(Color.hf.warning)
                            .clipShape(Capsule())
                    }
                }
            }

            Spacer()

            Text(formatAmount(recurring.amount))
                .font(.system(size: 16, weight: .semibold, design: .rounded))
                .foregroundStyle(recurring.type == .income ? Color.hf.income : Color.hf.expense)
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

struct AddRecurringSheet: View {
    @EnvironmentObject var dataManager: DataManager
    @Environment(\.dismiss) var dismiss

    @State private var title = ""
    @State private var amount = ""
    @State private var type: TransactionType = .expense
    @State private var category: RecurringCategory = .subscriptions
    @State private var frequency: RecurrenceFrequency = .monthly
    @State private var startDate = Date()
    @State private var hasEndDate = false
    @State private var endDate = Date()
    @State private var hasRemainingPayments = false
    @State private var remainingPayments = ""

    var body: some View {
        NavigationStack {
            Form {
                Section {
                    TextField("Name (e.g., Netflix)", text: $title)
                    HStack {
                        Text(dataManager.profile.currency.symbol)
                            .foregroundStyle(.secondary)
                        TextField("Amount", text: $amount)
                            .keyboardType(.decimalPad)
                    }
                }

                Section("Type") {
                    Picker("Type", selection: $type) {
                        Text("Expense").tag(TransactionType.expense)
                        Text("Income").tag(TransactionType.income)
                    }
                    .pickerStyle(.segmented)
                }

                Section("Category") {
                    Picker("Category", selection: $category) {
                        ForEach(RecurringCategory.allCases, id: \.self) { cat in
                            Label {
                                Text(cat.title)
                            } icon: {
                                Image(systemName: cat.icon)
                            }
                            .tag(cat)
                        }
                    }
                }

                Section("Frequency") {
                    Picker("Frequency", selection: $frequency) {
                        ForEach(RecurrenceFrequency.allCases, id: \.self) { freq in
                            Text(freq.title).tag(freq)
                        }
                    }
                }

                Section("Start Date") {
                    DatePicker("Start", selection: $startDate, displayedComponents: .date)
                }

                Section("End Condition") {
                    Toggle("Has end date", isOn: $hasEndDate)
                    if hasEndDate {
                        DatePicker("End date", selection: $endDate, displayedComponents: .date)
                    }

                    Toggle("Fixed number of payments", isOn: $hasRemainingPayments)
                    if hasRemainingPayments {
                        HStack {
                            TextField("Payments", text: $remainingPayments)
                                .keyboardType(.numberPad)
                            Text("payments left")
                                .foregroundStyle(.secondary)
                        }
                    }
                }
            }
            .navigationTitle("Add Regular Expense")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Add") {
                        addRecurring()
                    }
                    .disabled(title.isEmpty || amount.isEmpty)
                }
            }
        }
        .presentationDetents([.large])
    }

    private func addRecurring() {
        guard let amountValue = Double(amount) else { return }

        let recurring = RecurringTransaction(
            title: title,
            amount: amountValue,
            type: type,
            category: category,
            frequency: frequency,
            startDate: startDate,
            nextDate: startDate,
            endDate: hasEndDate ? endDate : nil,
            remainingPayments: hasRemainingPayments ? Int(remainingPayments) : nil
        )
        dataManager.addRecurringTransaction(recurring)
        dismiss()
    }
}

struct EditRecurringSheet: View {
    @EnvironmentObject var dataManager: DataManager
    @Environment(\.dismiss) var dismiss

    let recurring: RecurringTransaction

    @State private var title: String
    @State private var amount: String
    @State private var type: TransactionType
    @State private var category: RecurringCategory
    @State private var frequency: RecurrenceFrequency
    @State private var startDate: Date
    @State private var hasEndDate: Bool
    @State private var endDate: Date
    @State private var hasRemainingPayments: Bool
    @State private var remainingPayments: String
    @State private var isActive: Bool
    @State private var showDeleteConfirmation = false

    init(recurring: RecurringTransaction) {
        self.recurring = recurring
        _title = State(initialValue: recurring.title)
        _amount = State(initialValue: String(format: "%.2f", recurring.amount))
        _type = State(initialValue: recurring.type)
        _category = State(initialValue: recurring.category)
        _frequency = State(initialValue: recurring.frequency)
        _startDate = State(initialValue: recurring.startDate)
        _hasEndDate = State(initialValue: recurring.endDate != nil)
        _endDate = State(initialValue: recurring.endDate ?? Date())
        _hasRemainingPayments = State(initialValue: recurring.remainingPayments != nil)
        _remainingPayments = State(initialValue: recurring.remainingPayments.map { String($0) } ?? "")
        _isActive = State(initialValue: recurring.isActive)
    }

    var body: some View {
        NavigationStack {
            Form {
                Section {
                    TextField("Name", text: $title)
                    HStack {
                        Text(dataManager.profile.currency.symbol)
                            .foregroundStyle(.secondary)
                        TextField("Amount", text: $amount)
                            .keyboardType(.decimalPad)
                    }
                }

                Section("Type") {
                    Picker("Type", selection: $type) {
                        Text("Expense").tag(TransactionType.expense)
                        Text("Income").tag(TransactionType.income)
                    }
                    .pickerStyle(.segmented)
                }

                Section("Category") {
                    Picker("Category", selection: $category) {
                        ForEach(RecurringCategory.allCases, id: \.self) { cat in
                            Label {
                                Text(cat.title)
                            } icon: {
                                Image(systemName: cat.icon)
                            }
                            .tag(cat)
                        }
                    }
                }

                Section("Frequency") {
                    Picker("Frequency", selection: $frequency) {
                        ForEach(RecurrenceFrequency.allCases, id: \.self) { freq in
                            Text(freq.title).tag(freq)
                        }
                    }
                }

                Section("End Condition") {
                    Toggle("Has end date", isOn: $hasEndDate)
                    if hasEndDate {
                        DatePicker("End date", selection: $endDate, displayedComponents: .date)
                    }

                    Toggle("Fixed number of payments", isOn: $hasRemainingPayments)
                    if hasRemainingPayments {
                        HStack {
                            TextField("Payments", text: $remainingPayments)
                                .keyboardType(.numberPad)
                            Text("payments left")
                                .foregroundStyle(.secondary)
                        }
                    }
                }

                Section {
                    Toggle("Active", isOn: $isActive)
                }

                Section {
                    Button(role: .destructive) {
                        showDeleteConfirmation = true
                    } label: {
                        Label("Delete", systemImage: "trash")
                            .frame(maxWidth: .infinity)
                    }
                }
            }
            .navigationTitle("Edit")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Save") {
                        saveRecurring()
                    }
                    .disabled(title.isEmpty || amount.isEmpty)
                }
            }
            .alert("Delete?", isPresented: $showDeleteConfirmation) {
                Button("Cancel", role: .cancel) {}
                Button("Delete", role: .destructive) {
                    let recurringToDelete = recurring
                    dismiss()
                    DispatchQueue.main.asyncAfter(deadline: .now() + 0.3) {
                        dataManager.deleteRecurringTransaction(recurringToDelete)
                    }
                }
            } message: {
                Text("This action cannot be undone.")
            }
        }
        .presentationDetents([.large])
    }

    private func saveRecurring() {
        guard let amountValue = Double(amount) else { return }

        var updated = recurring
        updated.title = title
        updated.amount = amountValue
        updated.type = type
        updated.category = category
        updated.frequency = frequency
        updated.startDate = startDate
        updated.endDate = hasEndDate ? endDate : nil
        updated.remainingPayments = hasRemainingPayments ? Int(remainingPayments) : nil
        updated.isActive = isActive

        dataManager.updateRecurringTransaction(updated)
        dismiss()
    }
}

import SwiftUI

struct EditTransactionSheet: View {
    @EnvironmentObject var dataManager: DataManager
    @Environment(\.dismiss) var dismiss
    @Environment(\.colorScheme) var colorScheme

    let transaction: Transaction

    @State private var title: String
    @State private var amount: String
    @State private var type: TransactionType
    @State private var selectedCategory: TransactionCategory
    @State private var date: Date
    @State private var showDeleteConfirmation = false

    init(transaction: Transaction) {
        self.transaction = transaction
        _title = State(initialValue: transaction.title)
        _amount = State(initialValue: String(format: "%.2f", transaction.amount))
        _type = State(initialValue: transaction.type)
        _selectedCategory = State(initialValue: TransactionCategory(rawValue: transaction.category) ?? .other)
        _date = State(initialValue: transaction.date)
    }

    private var availableCategories: [TransactionCategory] {
        type == .income ? TransactionCategory.incomeCategories : TransactionCategory.expenseCategories
    }

    // Validation
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

                Section("Type") {
                    Picker("Type", selection: $type) {
                        Text("Income").tag(TransactionType.income)
                        Text("Expense").tag(TransactionType.expense)
                    }
                    .pickerStyle(.segmented)
                    .onChange(of: type) { _, newType in
                        // Reset category when type changes
                        selectedCategory = newType == .income ? .salary : .food
                    }
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

                Section("Date") {
                    DatePicker(
                        "Date",
                        selection: $date,
                        displayedComponents: .date
                    )
                }

                Section {
                    Button(role: .destructive) {
                        showDeleteConfirmation = true
                    } label: {
                        Label("Delete Transaction", systemImage: "trash")
                            .frame(maxWidth: .infinity)
                    }
                }
            }
            .navigationTitle("Edit Transaction")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Save") {
                        saveTransaction()
                    }
                    .disabled(!isValid)
                }
            }
            .alert("Delete Transaction?", isPresented: $showDeleteConfirmation) {
                Button("Cancel", role: .cancel) {}
                Button("Delete", role: .destructive) {
                    let transactionToDelete = transaction
                    dismiss()
                    DispatchQueue.main.asyncAfter(deadline: .now() + 0.3) {
                        dataManager.deleteTransaction(transactionToDelete)
                    }
                }
            } message: {
                Text("This action cannot be undone.")
            }
        }
        .presentationDetents([.large])
    }

    private func saveTransaction() {
        guard let amountValue = Double(amount.replacingOccurrences(of: ",", with: ".")) else { return }

        var updatedTransaction = transaction
        updatedTransaction.title = title.trimmingCharacters(in: .whitespaces)
        updatedTransaction.amount = amountValue
        updatedTransaction.type = type
        updatedTransaction.category = selectedCategory.rawValue
        updatedTransaction.date = date

        dataManager.updateTransaction(updatedTransaction)
        HapticManager.completionSuccess()
        dismiss()
    }
}

import SwiftUI

// MARK: - Savings Goal Content (for collapsible section)

struct SavingsGoalContent: View {
    let goal: SavingsGoal
    let currency: Currency
    let onEdit: () -> Void
    let onDelete: () -> Void

    private var isGoalMet: Bool {
        goal.currentSavings >= goal.monthlyTarget
    }

    var body: some View {
        VStack(spacing: 12) {
            // Progress bar
            GeometryReader { geometry in
                ZStack(alignment: .leading) {
                    RoundedRectangle(cornerRadius: 6)
                        .fill(Color.hf.accent.opacity(0.2))
                        .frame(height: 10)

                    RoundedRectangle(cornerRadius: 6)
                        .fill(isGoalMet ? Color.hf.income : Color.hf.accent)
                        .frame(width: geometry.size.width * goal.progress, height: 10)
                        .animation(.easeInOut(duration: 0.3), value: goal.progress)
                }
            }
            .frame(height: 10)

            // Stats row
            HStack {
                VStack(alignment: .leading, spacing: 2) {
                    Text("Saved")
                        .font(.system(size: 11))
                        .foregroundStyle(.secondary)
                    Text(formatCurrency(goal.currentSavings))
                        .font(.system(size: 16, weight: .semibold, design: .rounded))
                        .foregroundStyle(goal.currentSavings >= 0 ? Color.hf.income : Color.hf.expense)
                }

                Spacer()

                VStack(alignment: .trailing, spacing: 2) {
                    Text("Target")
                        .font(.system(size: 11))
                        .foregroundStyle(.secondary)
                    Text(formatCurrency(goal.monthlyTarget))
                        .font(.system(size: 16, weight: .semibold, design: .rounded))
                }
            }

            // Edit/Delete buttons
            HStack(spacing: 16) {
                Button {
                    onEdit()
                } label: {
                    HStack(spacing: 4) {
                        Image(systemName: "pencil")
                            .font(.system(size: 11))
                        Text("Edit")
                            .font(.system(size: 13, weight: .medium))
                    }
                    .foregroundStyle(Color.hf.accent)
                    .padding(.horizontal, 12)
                    .padding(.vertical, 6)
                    .background(Color.hf.accent.opacity(0.1))
                    .clipShape(Capsule())
                }
                .buttonStyle(.plain)

                Button {
                    onDelete()
                } label: {
                    HStack(spacing: 4) {
                        Image(systemName: "trash")
                            .font(.system(size: 11))
                        Text("Delete")
                            .font(.system(size: 13, weight: .medium))
                    }
                    .foregroundStyle(Color.hf.expense)
                    .padding(.horizontal, 12)
                    .padding(.vertical, 6)
                    .background(Color.hf.expense.opacity(0.1))
                    .clipShape(Capsule())
                }
                .buttonStyle(.plain)

                Spacer()
            }
        }
    }

    private func formatCurrency(_ amount: Double) -> String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        formatter.currencySymbol = currency.symbol
        formatter.maximumFractionDigits = 0
        return formatter.string(from: NSNumber(value: abs(amount))) ?? "\(currency.symbol)0"
    }
}

// MARK: - Savings Goal Card (standalone)

struct SavingsGoalCard: View {
    let goal: SavingsGoal
    let currency: Currency
    let onEdit: () -> Void
    let onDelete: () -> Void

    private var remainingToSave: Double {
        max(0, goal.monthlyTarget - goal.currentSavings)
    }

    private var isGoalMet: Bool {
        goal.currentSavings >= goal.monthlyTarget
    }

    var body: some View {
        VStack(spacing: 12) {
            // Header
            HStack {
                HStack(spacing: 6) {
                    Image(systemName: isGoalMet ? "checkmark.seal.fill" : "target")
                        .font(.system(size: 14))
                        .foregroundStyle(isGoalMet ? Color.hf.income : Color.hf.accent)

                    Text("Savings Goal")
                        .font(.system(size: 14, weight: .semibold))
                }

                Spacer()

                Menu {
                    Button {
                        onEdit()
                    } label: {
                        Label("Edit Goal", systemImage: "pencil")
                    }

                    Button(role: .destructive) {
                        onDelete()
                    } label: {
                        Label("Delete Goal", systemImage: "trash")
                    }
                } label: {
                    Image(systemName: "ellipsis")
                        .font(.system(size: 14))
                        .foregroundStyle(.secondary)
                        .frame(width: 24, height: 24)
                }
            }

            // Progress
            VStack(spacing: 8) {
                // Progress bar
                GeometryReader { geometry in
                    ZStack(alignment: .leading) {
                        RoundedRectangle(cornerRadius: 6)
                            .fill(Color.hf.accent.opacity(0.2))
                            .frame(height: 12)

                        RoundedRectangle(cornerRadius: 6)
                            .fill(
                                isGoalMet ?
                                LinearGradient(
                                    colors: [Color.hf.income, Color.hf.income.opacity(0.8)],
                                    startPoint: .leading,
                                    endPoint: .trailing
                                ) :
                                LinearGradient(
                                    colors: [Color.hf.accent, Color.hf.accent.opacity(0.8)],
                                    startPoint: .leading,
                                    endPoint: .trailing
                                )
                            )
                            .frame(width: geometry.size.width * goal.progress, height: 12)
                            .animation(.easeInOut(duration: 0.3), value: goal.progress)
                    }
                }
                .frame(height: 12)

                // Stats
                HStack {
                    VStack(alignment: .leading, spacing: 2) {
                        Text("Saved")
                            .font(.system(size: 11))
                            .foregroundStyle(.secondary)
                        Text(formatCurrency(goal.currentSavings))
                            .font(.system(size: 15, weight: .semibold, design: .rounded))
                            .foregroundStyle(goal.currentSavings >= 0 ? Color.hf.income : Color.hf.expense)
                    }

                    Spacer()

                    VStack(alignment: .trailing, spacing: 2) {
                        Text("Target")
                            .font(.system(size: 11))
                            .foregroundStyle(.secondary)
                        Text(formatCurrency(goal.monthlyTarget))
                            .font(.system(size: 15, weight: .semibold, design: .rounded))
                    }
                }
            }

            // Status message
            if isGoalMet {
                HStack(spacing: 6) {
                    Image(systemName: "star.fill")
                        .font(.system(size: 12))
                    Text("Goal reached! Great job saving this month!")
                        .font(.system(size: 13))
                }
                .foregroundStyle(Color.hf.income)
                .frame(maxWidth: .infinity, alignment: .leading)
            } else if remainingToSave > 0 {
                HStack(spacing: 4) {
                    Text("\(formatCurrency(remainingToSave))")
                        .font(.system(size: 13, weight: .medium))
                        .foregroundStyle(Color.hf.accent)
                    Text("left to save this month")
                        .font(.system(size: 13))
                        .foregroundStyle(.secondary)
                }
                .frame(maxWidth: .infinity, alignment: .leading)
            }
        }
        .padding(16)
        .background(Color.hf.cardBackground)
        .clipShape(RoundedRectangle(cornerRadius: 16))
    }

    private func formatCurrency(_ amount: Double) -> String {
        let formatter = NumberFormatter()
        formatter.numberStyle = .currency
        formatter.currencySymbol = currency.symbol
        formatter.maximumFractionDigits = currency == .jpy || currency == .krw ? 0 : 0
        return formatter.string(from: NSNumber(value: abs(amount))) ?? "\(currency.symbol)0"
    }
}

// MARK: - Set Savings Goal Button

struct SetSavingsGoalButton: View {
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack(spacing: 8) {
                Image(systemName: "plus.circle.fill")
                    .font(.system(size: 16))
                Text("Set Savings Goal")
                    .font(.system(size: 14, weight: .medium))
            }
            .foregroundStyle(Color.hf.accent)
            .frame(maxWidth: .infinity)
            .padding(.vertical, 14)
            .background(Color.hf.accent.opacity(0.1))
            .clipShape(RoundedRectangle(cornerRadius: 12))
        }
        .buttonStyle(.plain)
    }
}

// MARK: - Savings Goal Sheet

struct SavingsGoalSheet: View {
    @EnvironmentObject var dataManager: DataManager
    @Environment(\.dismiss) var dismiss
    @State private var targetAmount: String
    let existingGoal: SavingsGoal?

    init(existingGoal: SavingsGoal? = nil) {
        self.existingGoal = existingGoal
        _targetAmount = State(initialValue: existingGoal.map { String(Int($0.monthlyTarget)) } ?? "")
    }

    private var isValid: Bool {
        guard let amount = Double(targetAmount.replacingOccurrences(of: ",", with: ".")) else {
            return false
        }
        return amount > 0 && amount <= 999999999
    }

    var body: some View {
        NavigationStack {
            Form {
                Section {
                    HStack {
                        Text(dataManager.profile.currency.symbol)
                            .foregroundStyle(.secondary)
                        TextField("Monthly target", text: $targetAmount)
                            .keyboardType(.numberPad)
                    }
                } header: {
                    Text("Monthly Savings Target")
                } footer: {
                    Text("Set how much you want to save each month. Your progress will be calculated from your income minus expenses.")
                }

                if existingGoal != nil {
                    Section {
                        Button(role: .destructive) {
                            dataManager.deleteSavingsGoal()
                            dismiss()
                        } label: {
                            HStack {
                                Spacer()
                                Text("Delete Goal")
                                Spacer()
                            }
                        }
                    }
                }
            }
            .navigationTitle(existingGoal == nil ? "Set Savings Goal" : "Edit Savings Goal")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Save") {
                        if let amount = Double(targetAmount.replacingOccurrences(of: ",", with: ".")) {
                            dataManager.setSavingsGoal(amount)
                            dismiss()
                        }
                    }
                    .disabled(!isValid)
                }
            }
        }
        .presentationDetents([.medium])
    }
}

import SwiftUI

struct BudgetWatchView: View {
    @EnvironmentObject var dataStore: WatchDataStore

    var body: some View {
        VStack(spacing: 16) {
            // Header
            HStack {
                Image(systemName: "creditcard")
                    .foregroundStyle(.purple)
                Text("Budget")
                    .font(.headline)
            }

            Spacer()

            // Balance
            VStack(spacing: 4) {
                Text("Balance")
                    .font(.caption)
                    .foregroundStyle(.secondary)

                Text("\(dataStore.budget.currencySymbol)\(formatAmount(dataStore.budget.balance))")
                    .font(.system(size: 28, weight: .bold, design: .rounded))
                    .foregroundStyle(dataStore.budget.balance >= 0 ? .primary : .red)
            }

            Spacer()

            // Refresh hint
            Text("Open iPhone for details")
                .font(.caption2)
                .foregroundStyle(.secondary)
        }
        .padding()
        .navigationTitle("Budget")
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

#Preview {
    BudgetWatchView()
        .environmentObject(WatchDataStore.shared)
}

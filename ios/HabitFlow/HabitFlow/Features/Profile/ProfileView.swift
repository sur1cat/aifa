import SwiftUI

struct ProfileView: View {
    @EnvironmentObject var authManager: AuthManager
    @EnvironmentObject var dataManager: DataManager
    @ObservedObject var notificationManager = NotificationManager.shared
    @Environment(\.colorScheme) var colorScheme
    @State private var nameInput = ""
    @State private var isEditingName = false
    @State private var showCurrencyPicker = false
    @State private var showNotificationSettings = false
    @State private var showExportSheet = false
    @State private var exportURL: URL?
    @State private var showDeleteAccountAlert = false
    @State private var isDeletingAccount = false
    @FocusState private var isNameFocused: Bool

    var body: some View {
        NavigationStack {
            List {
                // Tagline
                Section {
                    Text("habits, tasks, money — in one flow")
                        .font(.system(size: 15))
                        .foregroundStyle(.secondary)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .listRowBackground(Color.clear)
                        .listRowInsets(EdgeInsets(top: 0, leading: 16, bottom: 8, trailing: 16))
                }
                .listRowSeparator(.hidden)

                // Account Card
                if let user = authManager.currentUser {
                    Section {
                        ProfileAccountCard(user: user)
                            .listRowBackground(Color.clear)
                            .listRowInsets(EdgeInsets(top: 0, leading: 16, bottom: 12, trailing: 16))
                    }
                    .listRowSeparator(.hidden)
                }

                // Weekly Review Card
                Section {
                    WeeklyReviewCard()
                        .listRowBackground(Color.clear)
                        .listRowInsets(EdgeInsets(top: 0, leading: 16, bottom: 12, trailing: 16))
                }
                .listRowSeparator(.hidden)

                // Name Section
                Section {
                    ProfileNameCard(
                        name: dataManager.profile.name,
                        isEditing: $isEditingName,
                        nameInput: $nameInput,
                        isNameFocused: $isNameFocused,
                        onSave: saveName
                    )
                    .listRowBackground(Color.clear)
                    .listRowInsets(EdgeInsets(top: 0, leading: 16, bottom: 12, trailing: 16))
                }
                .listRowSeparator(.hidden)

                // Settings
                Section {
                    Text("Settings")
                        .font(.system(size: 14, weight: .medium))
                        .foregroundStyle(.secondary)
                        .listRowBackground(Color.clear)
                        .listRowInsets(EdgeInsets(top: 8, leading: 20, bottom: 8, trailing: 16))
                }
                .listRowSeparator(.hidden)

                Section {
                    VStack(spacing: 0) {
                        SettingsRow(
                            icon: "dollarsign.circle",
                            title: "Currency",
                            value: "\(dataManager.profile.currency.symbol) \(dataManager.profile.currency.rawValue)"
                        ) {
                            showCurrencyPicker = true
                        }

                        Divider().padding(.leading, 52)

                        SettingsRow(
                            icon: "bell",
                            title: "Notifications",
                            value: (notificationManager.morningDigestEnabled || notificationManager.eveningDigestEnabled) ? "On" : nil
                        ) {
                            showNotificationSettings = true
                        }

                        Divider().padding(.leading, 52)

                        SettingsRow(
                            icon: "square.and.arrow.up",
                            title: "Export Data",
                            value: nil
                        ) {
                            showExportSheet = true
                        }
                    }
                    .background(Color.hf.cardBackground)
                    .clipShape(RoundedRectangle(cornerRadius: 16))
                    .listRowBackground(Color.clear)
                    .listRowInsets(EdgeInsets(top: 0, leading: 16, bottom: 12, trailing: 16))
                }
                .listRowSeparator(.hidden)

                // Actions
                Section {
                    Text("Account")
                        .font(.system(size: 14, weight: .medium))
                        .foregroundStyle(.secondary)
                        .listRowBackground(Color.clear)
                        .listRowInsets(EdgeInsets(top: 8, leading: 20, bottom: 8, trailing: 16))
                }
                .listRowSeparator(.hidden)

                // Sign Out Button
                Section {
                    Button {
                        Task { await authManager.signOut() }
                    } label: {
                        HStack {
                            Image(systemName: "rectangle.portrait.and.arrow.right")
                            Text("Sign Out")
                        }
                        .font(.system(size: 16, weight: .medium))
                        .foregroundStyle(.primary)
                        .frame(maxWidth: .infinity)
                        .padding()
                        .background(Color.hf.cardBackground)
                        .clipShape(RoundedRectangle(cornerRadius: 14))
                    }
                    .buttonStyle(.plain)
                    .listRowBackground(Color.clear)
                    .listRowInsets(EdgeInsets(top: 0, leading: 16, bottom: 12, trailing: 16))
                }
                .listRowSeparator(.hidden)

                // Delete Account Button
                Section {
                    Button {
                        showDeleteAccountAlert = true
                    } label: {
                        HStack {
                            if isDeletingAccount {
                                ProgressView()
                                    .tint(Color.hf.expense)
                                    .padding(.trailing, 4)
                            } else {
                                Image(systemName: "trash")
                            }
                            Text("Delete Account")
                        }
                        .font(.system(size: 16, weight: .medium))
                        .foregroundStyle(Color.hf.expense)
                        .frame(maxWidth: .infinity)
                        .padding()
                        .background(Color.hf.expense.opacity(0.1))
                        .clipShape(RoundedRectangle(cornerRadius: 14))
                    }
                    .buttonStyle(.plain)
                    .disabled(isDeletingAccount)
                    .listRowBackground(Color.clear)
                    .listRowInsets(EdgeInsets(top: 0, leading: 16, bottom: 24, trailing: 16))
                }
                .listRowSeparator(.hidden)
            }
            .listStyle(.plain)
            .scrollContentBackground(.hidden)
            .background(AppTheme.appBackground(for: colorScheme))
            .navigationTitle("Atoma Profile")
            .sheet(isPresented: $showCurrencyPicker) {
                CurrencyPickerSheet()
            }
            .sheet(isPresented: $showNotificationSettings) {
                NotificationSettingsSheet()
            }
            .sheet(isPresented: $showExportSheet) {
                ExportDataSheet()
            }
            .alert("Delete Account", isPresented: $showDeleteAccountAlert) {
                Button("Cancel", role: .cancel) {}
                Button("Delete", role: .destructive) {
                    Task {
                        isDeletingAccount = true
                        do {
                            try await authManager.deleteAccount()
                        } catch {
                            // Error handled silently
                        }
                        isDeletingAccount = false
                    }
                }
            } message: {
                Text("This will permanently delete your account and all your data. This action cannot be undone.")
            }
        }
    }

    func saveName() {
        dataManager.updateName(nameInput)
        isEditingName = false
        isNameFocused = false
    }

    // MARK: - AI Context

    private func buildProfileContext() -> String {
        let calendar = Calendar.current
        let today = Date()
        let todayString = Habit.dateString(from: today)

        var context = "=== USER'S COMPLETE LIFE DATA ===\n"
        context += "User: \(dataManager.profile.name ?? "User")\n"
        context += "Date: \(todayString)\n\n"

        // Habits overview
        let activeHabits = dataManager.habits.filter { $0.archivedAt == nil }
        let habitsCompletedToday = activeHabits.filter { $0.completedDates.contains(todayString) }.count
        context += "HABITS OVERVIEW:\n"
        context += "- Active habits: \(activeHabits.count)\n"
        context += "- Completed today: \(habitsCompletedToday)\n"

        // Calculate weekly completion rate
        var weekHabitCompleted = 0
        var weekHabitTotal = 0
        for dayOffset in 0..<7 {
            if let date = calendar.date(byAdding: .day, value: -dayOffset, to: today) {
                let dateStr = Habit.dateString(from: date)
                for habit in activeHabits {
                    weekHabitTotal += 1
                    if habit.completedDates.contains(dateStr) {
                        weekHabitCompleted += 1
                    }
                }
            }
        }
        let habitWeekRate = weekHabitTotal > 0 ? Int(Double(weekHabitCompleted) / Double(weekHabitTotal) * 100) : 0
        context += "- Weekly rate: \(habitWeekRate)%\n"

        // Top streaks
        let topStreaks = activeHabits.filter { $0.streak > 0 }.sorted { $0.streak > $1.streak }
        if !topStreaks.isEmpty {
            context += "- Best streaks: "
            context += topStreaks.prefix(3).map { "\($0.icon) \($0.streak)d" }.joined(separator: ", ")
            context += "\n"
        }

        // Goals
        let goals = dataManager.activeGoals
        if !goals.isEmpty {
            context += "\nGOALS (\(goals.count)):\n"
            for goal in goals {
                let goalHabits = activeHabits.filter { $0.goalId == goal.id }
                context += "- \(goal.icon) \(goal.title): \(goalHabits.count) habits\n"
            }
        }

        // Tasks overview
        let todayTasks = dataManager.tasksForDate(today)
        let tasksCompletedToday = todayTasks.filter { $0.isCompleted }.count
        context += "\nTASKS OVERVIEW:\n"
        context += "- Today: \(tasksCompletedToday)/\(todayTasks.count) completed\n"

        // Overdue tasks
        let overdue = dataManager.tasks.filter { !$0.isCompleted && calendar.compare($0.dueDate, to: today, toGranularity: .day) == .orderedAscending }
        if !overdue.isEmpty {
            context += "- Overdue: \(overdue.count) tasks\n"
        }

        // Budget overview
        let monthTransactions = dataManager.transactions.filter {
            calendar.isDate($0.date, equalTo: today, toGranularity: .month)
        }
        let monthIncome = monthTransactions.filter { $0.type == .income }.reduce(0) { $0 + $1.amount }
        let monthExpenses = monthTransactions.filter { $0.type == .expense }.reduce(0) { $0 + $1.amount }
        let currency = dataManager.profile.currency.symbol

        context += "\nBUDGET OVERVIEW:\n"
        context += "- Monthly income: \(currency)\(Int(monthIncome))\n"
        context += "- Monthly expenses: \(currency)\(Int(monthExpenses))\n"
        context += "- Balance: \(currency)\(Int(monthIncome - monthExpenses))\n"

        // Recurring expenses
        let recurring = dataManager.recurringTransactions
        if !recurring.isEmpty {
            var monthlyRecurring: Double = 0
            for r in recurring {
                switch r.frequency {
                case .weekly: monthlyRecurring += r.amount * 4
                case .biweekly: monthlyRecurring += r.amount * 2
                case .monthly: monthlyRecurring += r.amount
                case .quarterly: monthlyRecurring += r.amount / 3
                case .yearly: monthlyRecurring += r.amount / 12
                }
            }
            context += "- Monthly recurring: \(currency)\(Int(monthlyRecurring))\n"
        }

        // Savings goal
        if let goal = dataManager.savingsGoal {
            context += "\nSAVINGS GOAL:\n"
            context += "- Target: \(currency)\(Int(goal.monthlyTarget))\n"
            context += "- Progress: \(Int(goal.progress * 100))%\n"
        }

        return context
    }
}

// MARK: - Profile Account Card

struct ProfileAccountCard: View {
    let user: User
    @Environment(\.colorScheme) var colorScheme

    var body: some View {
        HStack(spacing: 16) {
            ZStack {
                Circle()
                    .fill(user.authProvider == "apple" ? Color.primary.opacity(0.1) : Color.hf.info.opacity(0.15))
                    .frame(width: 56, height: 56)

                Image(systemName: user.providerIcon)
                    .font(.system(size: 24))
                    .foregroundStyle(user.authProvider == "apple" ? .primary : Color.hf.info)
            }

            VStack(alignment: .leading, spacing: 4) {
                Text(user.providerDisplayName)
                    .font(.system(size: 17, weight: .semibold))

                if let email = user.email {
                    Text(email)
                        .font(.system(size: 14))
                        .foregroundStyle(.secondary)
                        .lineLimit(1)
                }
            }

            Spacer()

            Image(systemName: "checkmark.seal.fill")
                .font(.system(size: 20))
                .foregroundStyle(Color.hf.accent)
        }
        .padding(16)
        .background(Color.hf.cardBackground)
        .clipShape(RoundedRectangle(cornerRadius: 16))
    }
}

// MARK: - Profile Streak Card

struct ProfileStreakCard: View {
    let days: Int
    @Environment(\.colorScheme) var colorScheme

    var body: some View {
        HStack(spacing: 16) {
            // Streak circle
            ZStack {
                Circle()
                    .fill(Color.hf.accent.opacity(0.15))
                    .frame(width: 64, height: 64)

                Text("🔥")
                    .font(.system(size: 28))
            }

            VStack(alignment: .leading, spacing: 4) {
                Text("\(days)")
                    .font(.system(size: 36, weight: .bold, design: .rounded))
                    .foregroundStyle(Color.hf.accent)

                Text("days with Atoma")
                    .font(.system(size: 14))
                    .foregroundStyle(.secondary)
            }

            Spacer()
        }
        .padding(16)
        .background(Color.hf.cardBackground)
        .clipShape(RoundedRectangle(cornerRadius: 16))
    }
}

// MARK: - Profile Stats Card

struct ProfileStatsCard: View {
    let habitsCount: Int
    let tasksCompleted: Int
    let tasksTotal: Int
    let transactionsCount: Int

    var body: some View {
        HStack(spacing: 0) {
            StatItem(value: "\(habitsCount)", label: "Habits", icon: "repeat")

            Divider()
                .frame(height: 40)

            StatItem(value: "\(tasksCompleted)/\(tasksTotal)", label: "Tasks", icon: "checkmark.circle")

            Divider()
                .frame(height: 40)

            StatItem(value: "\(transactionsCount)", label: "Transactions", icon: "creditcard")
        }
        .padding(.vertical, 16)
        .background(Color.hf.cardBackground)
        .clipShape(RoundedRectangle(cornerRadius: 16))
    }
}

struct StatItem: View {
    let value: String
    let label: String
    let icon: String

    var body: some View {
        VStack(spacing: 8) {
            Text(value)
                .font(.system(size: 20, weight: .bold, design: .rounded))
                .foregroundStyle(.primary)

            HStack(spacing: 4) {
                Image(systemName: icon)
                    .font(.system(size: 10))
                Text(label)
                    .font(.system(size: 11))
            }
            .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity)
    }
}

// MARK: - Profile Name Card

struct ProfileNameCard: View {
    let name: String
    @Binding var isEditing: Bool
    @Binding var nameInput: String
    var isNameFocused: FocusState<Bool>.Binding
    let onSave: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Your Name")
                .font(.system(size: 13))
                .foregroundStyle(.secondary)

            HStack {
                if isEditing {
                    TextField("Enter your name", text: $nameInput)
                        .font(.system(size: 17))
                        .focused(isNameFocused)
                        .onSubmit {
                            onSave()
                        }

                    Button("Save") {
                        onSave()
                    }
                    .font(.system(size: 15, weight: .medium))
                    .foregroundStyle(Color.hf.accent)
                } else {
                    Text(name.isEmpty ? "Not set" : name)
                        .font(.system(size: 17))
                        .foregroundStyle(name.isEmpty ? .secondary : .primary)

                    Spacer()

                    Button {
                        nameInput = name
                        isEditing = true
                        isNameFocused.wrappedValue = true
                    } label: {
                        Text("Edit")
                            .font(.system(size: 15, weight: .medium))
                            .foregroundStyle(Color.hf.accent)
                    }
                }
            }
        }
        .padding(16)
        .background(Color.hf.cardBackground)
        .clipShape(RoundedRectangle(cornerRadius: 16))
    }
}

// MARK: - Settings Row

struct SettingsRow: View {
    let icon: String
    let title: String
    let value: String?
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack(spacing: 12) {
                Image(systemName: icon)
                    .font(.system(size: 17))
                    .foregroundStyle(Color.hf.accent)
                    .frame(width: 24)

                Text(title)
                    .font(.system(size: 16))
                    .foregroundStyle(.primary)

                Spacer()

                if let value = value {
                    Text(value)
                        .font(.system(size: 15))
                        .foregroundStyle(.secondary)
                }

                Image(systemName: "chevron.right")
                    .font(.system(size: 13, weight: .semibold))
                    .foregroundStyle(.tertiary)
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 14)
        }
        .buttonStyle(.plain)
    }
}

// MARK: - Stat Row (legacy)

struct StatRow: View {
    let label: LocalizedStringKey
    let value: String

    var body: some View {
        HStack {
            Text(label)
                .font(.system(size: 16))
            Spacer()
            Text(value)
                .font(.system(size: 16))
                .foregroundStyle(.secondary)
        }
        .padding()
    }
}

// MARK: - Currency Picker Sheet

struct CurrencyPickerSheet: View {
    @EnvironmentObject var dataManager: DataManager
    @Environment(\.dismiss) var dismiss

    var body: some View {
        NavigationStack {
            List {
                ForEach(Currency.allCases, id: \.self) { currency in
                    Button {
                        dataManager.updateCurrency(currency)
                        dismiss()
                    } label: {
                        HStack {
                            Text(currency.symbol)
                                .font(.system(size: 20))
                                .frame(width: 40)
                            VStack(alignment: .leading) {
                                Text(currency.rawValue)
                                    .font(.system(size: 16, weight: .medium))
                                Text(currency.name)
                                    .font(.system(size: 14))
                                    .foregroundStyle(.secondary)
                            }
                            Spacer()
                            if dataManager.profile.currency == currency {
                                Image(systemName: "checkmark")
                                    .foregroundStyle(Color.hf.accent)
                            }
                        }
                        .foregroundStyle(.primary)
                    }
                }
            }
            .navigationTitle("Currency")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Done") { dismiss() }
                }
            }
        }
        .presentationDetents([.medium])
    }
}

// MARK: - Notification Settings Sheet

struct NotificationSettingsSheet: View {
    @ObservedObject var notificationManager = NotificationManager.shared
    @Environment(\.dismiss) var dismiss
    @State private var showPermissionAlert = false

    let morningHours = Array(5...11)
    let eveningHours = Array(18...23)

    var body: some View {
        NavigationStack {
            Form {
                Section {
                    Toggle(isOn: $notificationManager.morningDigestEnabled) {
                        Label("Morning reminder", systemImage: "sun.horizon")
                    }
                    .onChange(of: notificationManager.morningDigestEnabled) { _, newValue in
                        handleToggle(newValue)
                    }

                    if notificationManager.morningDigestEnabled {
                        Picker("Time", selection: $notificationManager.morningDigestHour) {
                            ForEach(morningHours, id: \.self) { hour in
                                Text("\(hour):00").tag(hour)
                            }
                        }
                        .onChange(of: notificationManager.morningDigestHour) { _, _ in
                            Task { await notificationManager.scheduleMorningDigest() }
                        }
                    }
                } header: {
                    Text("Morning")
                } footer: {
                    Text("Get a reminder to check your habits and tasks")
                }

                Section {
                    Toggle(isOn: $notificationManager.eveningDigestEnabled) {
                        Label("Evening check-in", systemImage: "moon.stars")
                    }
                    .onChange(of: notificationManager.eveningDigestEnabled) { _, newValue in
                        handleToggle(newValue)
                    }

                    if notificationManager.eveningDigestEnabled {
                        Picker("Time", selection: $notificationManager.eveningDigestHour) {
                            ForEach(eveningHours, id: \.self) { hour in
                                Text("\(hour):00").tag(hour)
                            }
                        }
                        .onChange(of: notificationManager.eveningDigestHour) { _, _ in
                            Task { await notificationManager.scheduleEveningDigest() }
                        }
                    }
                } header: {
                    Text("Evening")
                } footer: {
                    Text("Review your day and complete remaining habits")
                }

                Section {
                    HStack {
                        Text("Habit reminders")
                        Spacer()
                        Text("Per habit")
                            .foregroundStyle(.secondary)
                    }
                } footer: {
                    Text("Set individual reminders when editing each habit")
                }
            }
            .navigationTitle("Notifications")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .confirmationAction) {
                    Button("Done") { dismiss() }
                }
            }
            .alert("Notifications Disabled", isPresented: $showPermissionAlert) {
                Button("Open Settings") {
                    if let url = URL(string: UIApplication.openSettingsURLString) {
                        UIApplication.shared.open(url)
                    }
                }
                Button("Cancel", role: .cancel) {
                    notificationManager.morningDigestEnabled = false
                    notificationManager.eveningDigestEnabled = false
                }
            } message: {
                Text("Please enable notifications in Settings to receive reminders.")
            }
        }
        .presentationDetents([.medium])
    }

    private func handleToggle(_ enabled: Bool) {
        if enabled {
            Task {
                if !notificationManager.isAuthorized {
                    let granted = await notificationManager.requestPermission()
                    if !granted {
                        showPermissionAlert = true
                        return
                    }
                }
                await notificationManager.updateDigestSchedule()
            }
        } else {
            Task {
                await notificationManager.updateDigestSchedule()
            }
        }
    }
}

// MARK: - Export Data Sheet

struct ExportDataSheet: View {
    @EnvironmentObject var dataManager: DataManager
    @Environment(\.dismiss) var dismiss
    @State private var isExporting = false
    @State private var exportURL: URL?
    @State private var showShareSheet = false
    @State private var selectedExportType: ExportType = .all

    enum ExportType: String, CaseIterable {
        case all = "All Data"
        case habits = "Habits"
        case tasks = "Tasks"
        case transactions = "Transactions"

        var icon: String {
            switch self {
            case .all: return "square.stack.3d.up"
            case .habits: return "repeat"
            case .tasks: return "checkmark.circle"
            case .transactions: return "creditcard"
            }
        }
    }

    var body: some View {
        NavigationStack {
            Form {
                Section {
                    ForEach(ExportType.allCases, id: \.self) { type in
                        Button {
                            selectedExportType = type
                        } label: {
                            HStack {
                                Label(type.rawValue, systemImage: type.icon)
                                    .foregroundStyle(.primary)
                                Spacer()
                                if selectedExportType == type {
                                    Image(systemName: "checkmark")
                                        .foregroundStyle(Color.hf.accent)
                                }
                            }
                        }
                    }
                } header: {
                    Text("What to export")
                }

                Section {
                    VStack(alignment: .leading, spacing: 8) {
                        Label("Habits: \(dataManager.habits.count)", systemImage: "repeat")
                        Label("Tasks: \(dataManager.tasks.count)", systemImage: "checkmark.circle")
                        Label("Transactions: \(dataManager.transactions.count)", systemImage: "creditcard")
                        if !dataManager.recurringTransactions.isEmpty {
                            Label("Recurring: \(dataManager.recurringTransactions.count)", systemImage: "arrow.clockwise")
                        }
                    }
                    .font(.system(size: 14))
                    .foregroundStyle(.secondary)
                } header: {
                    Text("Data summary")
                }

                Section {
                    Button {
                        exportData()
                    } label: {
                        HStack {
                            Spacer()
                            if isExporting {
                                ProgressView()
                                    .padding(.trailing, 8)
                            }
                            Text(isExporting ? "Preparing..." : "Export as CSV")
                                .font(.system(size: 16, weight: .medium))
                            Spacer()
                        }
                    }
                    .disabled(isExporting)
                } footer: {
                    Text("Data will be exported in CSV format that you can open in Excel or Google Sheets")
                }
            }
            .navigationTitle("Export Data")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
            }
            .sheet(isPresented: $showShareSheet) {
                if let url = exportURL {
                    ShareSheet(items: [url])
                }
            }
        }
        .presentationDetents([.medium, .large])
    }

    private func exportData() {
        isExporting = true

        Task {
            let url: URL?

            switch selectedExportType {
            case .all:
                url = ExportService.shared.exportAllData(
                    habits: dataManager.habits,
                    tasks: dataManager.tasks,
                    transactions: dataManager.transactions,
                    recurring: dataManager.recurringTransactions,
                    currency: dataManager.profile.currency
                )
            case .habits:
                url = ExportService.shared.exportHabitsToCSV(habits: dataManager.habits)
            case .tasks:
                url = ExportService.shared.exportTasksToCSV(tasks: dataManager.tasks)
            case .transactions:
                url = ExportService.shared.exportTransactionsToCSV(
                    transactions: dataManager.transactions,
                    currency: dataManager.profile.currency
                )
            }

            await MainActor.run {
                isExporting = false
                if let url = url {
                    exportURL = url
                    showShareSheet = true
                }
            }
        }
    }
}

// MARK: - Share Sheet

struct ShareSheet: UIViewControllerRepresentable {
    let items: [Any]

    func makeUIViewController(context: Context) -> UIActivityViewController {
        UIActivityViewController(activityItems: items, applicationActivities: nil)
    }

    func updateUIViewController(_ uiViewController: UIActivityViewController, context: Context) {}
}

// MARK: - App Icon Picker Sheet

struct AppIconPickerSheet: View {
    @StateObject private var iconManager = AppIconManager.shared
    @Environment(\.dismiss) var dismiss
    @Environment(\.colorScheme) var colorScheme

    var body: some View {
        NavigationStack {
            ScrollView {
                LazyVGrid(columns: [GridItem(.adaptive(minimum: 100))], spacing: 20) {
                    ForEach(AppIconManager.AppIcon.allCases) { icon in
                        Button {
                            Task {
                                await iconManager.setIcon(icon)
                            }
                        } label: {
                            VStack(spacing: 12) {
                                RoundedRectangle(cornerRadius: 18)
                                    .fill(icon.previewGradient)
                                    .frame(width: 70, height: 70)
                                    .overlay {
                                        Text("A")
                                            .font(.system(size: 32, weight: .bold, design: .rounded))
                                            .foregroundStyle(icon == .minimal ? .black : .white)
                                    }
                                    .shadow(color: .black.opacity(0.1), radius: 4, x: 0, y: 2)
                                    .overlay(
                                        RoundedRectangle(cornerRadius: 18)
                                            .stroke(iconManager.currentIcon == icon ? Color.hf.accent : .clear, lineWidth: 3)
                                    )

                                Text(icon.displayName)
                                    .font(.system(size: 13, weight: .medium))
                                    .foregroundStyle(.primary)

                                if iconManager.currentIcon == icon {
                                    Image(systemName: "checkmark.circle.fill")
                                        .foregroundStyle(Color.hf.accent)
                                        .font(.system(size: 18))
                                } else {
                                    Circle()
                                        .stroke(Color.secondary.opacity(0.3), lineWidth: 1.5)
                                        .frame(width: 18, height: 18)
                                }
                            }
                        }
                        .buttonStyle(.plain)
                    }
                }
                .padding()
            }
            .background(AppTheme.appBackground(for: colorScheme))
            .navigationTitle("App Icon")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .confirmationAction) {
                    Button("Done") { dismiss() }
                }
            }
        }
        .presentationDetents([.medium])
    }
}

import SwiftUI

struct EditHabitSheet: View {
    @EnvironmentObject var dataManager: DataManager
    @ObservedObject var notificationManager = NotificationManager.shared
    @Environment(\.dismiss) var dismiss
    @Environment(\.colorScheme) var colorScheme

    let habit: Habit

    @State private var title: String
    @State private var selectedIcon: String
    @State private var selectedColor: String
    @State private var selectedPeriod: HabitPeriod
    @State private var selectedGoalId: UUID?
    @State private var reminderEnabled: Bool
    @State private var reminderTime: Date
    @State private var hasGoal: Bool
    @State private var targetValue: String
    @State private var unit: String
    @State private var showDeleteConfirmation = false
    @State private var showPermissionAlert = false
    @State private var showArchiveConfirmation = false

    let icons = ["🎯", "💪", "📚", "🧘", "🏃", "💧", "🍎", "😴", "✍️", "🎨", "🎵", "🧹"]
    let commonUnits = ["times", "minutes", "pages", "km", "reps", "glasses", "steps"]

    private var titleError: String? {
        if title.isEmpty { return nil }
        let trimmed = title.trimmingCharacters(in: .whitespaces)
        if trimmed.count < 2 {
            return "Title must be at least 2 characters"
        }
        if title.count > 50 {
            return "Title too long (max 50 characters)"
        }
        return nil
    }

    private var isValid: Bool {
        let trimmed = title.trimmingCharacters(in: .whitespaces)
        return trimmed.count >= 2 && title.count <= 50
    }

    init(habit: Habit) {
        self.habit = habit
        _title = State(initialValue: habit.title)
        _selectedIcon = State(initialValue: habit.icon)
        _selectedColor = State(initialValue: habit.color)
        _selectedPeriod = State(initialValue: habit.period)
        _selectedGoalId = State(initialValue: habit.goalId)
        _reminderEnabled = State(initialValue: habit.reminderEnabled)
        _reminderTime = State(initialValue: habit.reminderTime ?? Self.defaultReminderTime)
        _hasGoal = State(initialValue: habit.hasGoal)
        _targetValue = State(initialValue: habit.targetValue.map { String($0) } ?? "")
        _unit = State(initialValue: habit.unit ?? "")
    }

    private static var defaultReminderTime: Date {
        var components = DateComponents()
        components.hour = 9
        components.minute = 0
        return Calendar.current.date(from: components) ?? Date()
    }

    var body: some View {
        NavigationStack {
            Form {
                Section {
                    TextField("Habit name", text: $title)
                    if let error = titleError {
                        Text(error)
                            .font(.caption)
                            .foregroundStyle(.red)
                    }
                }

                // Goal Picker
                if !dataManager.activeGoals.isEmpty {
                    Section("Link to Goal") {
                        Picker("Goal", selection: $selectedGoalId) {
                            Text("None").tag(nil as UUID?)
                            ForEach(dataManager.activeGoals) { goal in
                                Text("\(goal.icon) \(goal.title)")
                                    .tag(goal.id as UUID?)
                            }
                        }
                    }
                }

                Section("Frequency") {
                    Picker("Frequency", selection: $selectedPeriod) {
                        ForEach(HabitPeriod.allCases, id: \.self) { period in
                            Text(period.title).tag(period)
                        }
                    }
                    .pickerStyle(.segmented)
                }

                Section("Goal") {
                    Toggle("Set a goal", isOn: $hasGoal)

                    if hasGoal {
                        HStack {
                            TextField("Target", text: $targetValue)
                                .keyboardType(.numberPad)
                                .frame(width: 80)

                            Picker("Unit", selection: $unit) {
                                Text("Select").tag("")
                                ForEach(commonUnits, id: \.self) { u in
                                    Text(u).tag(u)
                                }
                            }
                        }
                    }
                }

                Section("Icon") {
                    LazyVGrid(columns: Array(repeating: GridItem(.flexible()), count: 6), spacing: 12) {
                        ForEach(icons, id: \.self) { icon in
                            Text(icon)
                                .font(.system(size: 28))
                                .frame(width: 44, height: 44)
                                .background(selectedIcon == icon ? Color.hf.checkmarkIncomplete.opacity(0.3) : Color.clear)
                                .clipShape(RoundedRectangle(cornerRadius: 10))
                                .onTapGesture {
                                    selectedIcon = icon
                                }
                        }
                    }
                }

                Section("Color") {
                    LazyVGrid(columns: Array(repeating: GridItem(.flexible()), count: 6), spacing: 12) {
                        ForEach(HFColors.habitColors, id: \.name) { item in
                            Circle()
                                .fill(item.color)
                                .frame(width: 36, height: 36)
                                .overlay(
                                    Circle()
                                        .stroke(Color.primary, lineWidth: selectedColor == item.name ? 2 : 0)
                                )
                                .onTapGesture {
                                    selectedColor = item.name
                                }
                        }
                    }
                }

                Section("Reminder") {
                    Toggle(isOn: $reminderEnabled) {
                        Label("Reminder", systemImage: "bell.fill")
                    }
                    .onChange(of: reminderEnabled) { _, newValue in
                        handleReminderToggle(newValue)
                    }

                    if reminderEnabled {
                        DatePicker(
                            "Time",
                            selection: $reminderTime,
                            displayedComponents: .hourAndMinute
                        )
                    }
                }

                Section {
                    if habit.isActive {
                        Button {
                            showArchiveConfirmation = true
                        } label: {
                            Label("Archive Habit", systemImage: "archivebox")
                                .frame(maxWidth: .infinity)
                        }
                        .tint(.orange)
                    } else {
                        Button {
                            dataManager.unarchiveHabit(habit)
                            dismiss()
                        } label: {
                            Label("Unarchive Habit", systemImage: "archivebox")
                                .frame(maxWidth: .infinity)
                        }
                        .tint(.green)
                    }

                    Button(role: .destructive) {
                        showDeleteConfirmation = true
                    } label: {
                        Label("Delete Habit", systemImage: "trash")
                            .frame(maxWidth: .infinity)
                    }
                }
            }
            .navigationTitle("Edit Habit")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Save") {
                        saveHabit()
                    }
                    .disabled(!isValid)
                }
            }
            .alert("Notifications Disabled", isPresented: $showPermissionAlert) {
                Button("Open Settings") {
                    openSettings()
                }
                Button("Cancel", role: .cancel) {
                    reminderEnabled = false
                }
            } message: {
                Text("Please enable notifications in Settings to receive habit reminders.")
            }
            .alert("Delete Habit?", isPresented: $showDeleteConfirmation) {
                Button("Cancel", role: .cancel) {}
                Button("Delete", role: .destructive) {
                    let habitToDelete = habit
                    dismiss()
                    DispatchQueue.main.asyncAfter(deadline: .now() + 0.3) {
                        dataManager.deleteHabit(habitToDelete)
                    }
                }
            } message: {
                Text("This action cannot be undone.")
            }
            .alert("Archive Habit?", isPresented: $showArchiveConfirmation) {
                Button("Cancel", role: .cancel) {}
                Button("Archive") {
                    dataManager.archiveHabit(habit)
                    dismiss()
                }
            } message: {
                Text("Archived habits are hidden from today but historical data is preserved. You can view them by selecting past dates.")
            }
        }
    }

    private func handleReminderToggle(_ enabled: Bool) {
        if enabled {
            Task {
                if !notificationManager.isAuthorized {
                    let granted = await notificationManager.requestPermission()
                    if !granted {
                        showPermissionAlert = true
                        return
                    }
                }
            }
        }
    }

    private func saveHabit() {
        var updatedHabit = habit
        updatedHabit.title = title.trimmingCharacters(in: .whitespaces)
        updatedHabit.icon = selectedIcon
        updatedHabit.color = selectedColor
        updatedHabit.period = selectedPeriod
        updatedHabit.goalId = selectedGoalId
        updatedHabit.reminderEnabled = reminderEnabled
        updatedHabit.reminderTime = reminderEnabled ? reminderTime : nil
        updatedHabit.targetValue = hasGoal ? Int(targetValue) : nil
        updatedHabit.unit = hasGoal && !unit.isEmpty ? unit : nil

        dataManager.updateHabit(updatedHabit)

        if reminderEnabled {
            Task {
                await notificationManager.scheduleHabitReminder(habit: updatedHabit)
            }
        } else {
            notificationManager.cancelHabitReminder(habitID: habit.id)
        }

        dismiss()
    }

    private func openSettings() {
        if let url = URL(string: UIApplication.openSettingsURLString) {
            UIApplication.shared.open(url)
        }
    }
}

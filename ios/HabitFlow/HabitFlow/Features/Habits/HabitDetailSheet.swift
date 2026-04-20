import SwiftUI

struct HabitDetailSheet: View {
    @EnvironmentObject var dataManager: DataManager
    @ObservedObject var notificationManager = NotificationManager.shared
    @Environment(\.dismiss) var dismiss
    @Environment(\.colorScheme) var colorScheme

    let habit: Habit

    @State private var reminderEnabled: Bool
    @State private var reminderTime: Date
    @State private var showPermissionAlert = false
    @State private var showDeleteConfirmation = false

    init(habit: Habit) {
        self.habit = habit
        _reminderEnabled = State(initialValue: habit.reminderEnabled)
        _reminderTime = State(initialValue: habit.reminderTime ?? Self.defaultReminderTime)
    }

    private static var defaultReminderTime: Date {
        var components = DateComponents()
        components.hour = 9
        components.minute = 0
        return Calendar.current.date(from: components) ?? Date()
    }

    var body: some View {
        NavigationStack {
            VStack(spacing: 24) {
                // Header
                VStack(spacing: 8) {
                    Text(habit.icon)
                        .font(.system(size: 56))

                    Text(habit.title)
                        .font(.system(size: 22, weight: .semibold))

                    HStack(spacing: 8) {
                        Text(habit.period.title)
                            .font(.system(size: 13))
                            .padding(.horizontal, 10)
                            .padding(.vertical, 4)
                            .background(habit.swiftUIColor.opacity(0.15))
                            .foregroundStyle(habit.swiftUIColor)
                            .clipShape(Capsule())

                        if habit.streak > 0 {
                            Text("\(habit.streak) \(habit.period.shortTitleString) streak")
                                .font(.system(size: 14))
                                .foregroundStyle(.secondary)
                        }
                    }
                }
                .padding(.top, 16)

                // Reminder Section
                VStack(spacing: 0) {
                    Toggle(isOn: $reminderEnabled) {
                        Label("Reminder", systemImage: "bell.fill")
                    }
                    .padding()
                    .onChange(of: reminderEnabled) { _, newValue in
                        handleReminderToggle(newValue)
                    }

                    if reminderEnabled {
                        Divider()
                            .padding(.leading)

                        DatePicker(
                            "Time",
                            selection: $reminderTime,
                            displayedComponents: .hourAndMinute
                        )
                        .padding()
                        .onChange(of: reminderTime) { _, _ in
                            saveReminder()
                        }
                    }
                }
                .background(Color.hf.cardBackground)
                .clipShape(RoundedRectangle(cornerRadius: 14))

                Spacer()

                // Delete Button
                Button(role: .destructive) {
                    showDeleteConfirmation = true
                } label: {
                    Label("Delete Habit", systemImage: "trash")
                        .frame(maxWidth: .infinity)
                        .padding()
                        .background(Color.hf.cardBackground)
                        .clipShape(RoundedRectangle(cornerRadius: 14))
                }
            }
            .padding()
            .background(AppTheme.appBackground(for: colorScheme))
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button {
                        dismiss()
                    } label: {
                        Image(systemName: "xmark.circle.fill")
                            .foregroundStyle(.secondary)
                    }
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
                    dataManager.deleteHabit(habit)
                    dismiss()
                }
            } message: {
                Text("This action cannot be undone.")
            }
        }
        .presentationDetents([.medium])
        .presentationDragIndicator(.visible)
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
                saveReminder()
            }
        } else {
            saveReminder()
        }
    }

    private func saveReminder() {
        var updatedHabit = habit
        updatedHabit.reminderEnabled = reminderEnabled
        updatedHabit.reminderTime = reminderEnabled ? reminderTime : nil

        dataManager.updateHabitReminder(updatedHabit)

        Task {
            await notificationManager.scheduleHabitReminder(habit: updatedHabit)
        }
    }

    private func openSettings() {
        if let url = URL(string: UIApplication.openSettingsURLString) {
            UIApplication.shared.open(url)
        }
    }
}

import SwiftUI

struct EditTaskSheet: View {
    @EnvironmentObject var dataManager: DataManager
    @Environment(\.dismiss) var dismiss
    @Environment(\.colorScheme) var colorScheme

    let task: DailyTask

    @State private var title: String
    @State private var priority: TaskPriority
    @State private var dueDate: Date
    @State private var showDeleteConfirmation = false

    private var titleError: String? {
        if title.isEmpty { return nil }
        let trimmed = title.trimmingCharacters(in: .whitespaces)
        if trimmed.count < 2 {
            return "Title must be at least 2 characters"
        }
        if title.count > 100 {
            return "Title too long (max 100 characters)"
        }
        return nil
    }

    private var isValid: Bool {
        let trimmed = title.trimmingCharacters(in: .whitespaces)
        return trimmed.count >= 2 && title.count <= 100
    }

    init(task: DailyTask) {
        self.task = task
        _title = State(initialValue: task.title)
        _priority = State(initialValue: task.priority)
        _dueDate = State(initialValue: task.dueDate)
    }

    var body: some View {
        NavigationStack {
            Form {
                Section {
                    TextField("Task name", text: $title)
                    if let error = titleError {
                        Text(error)
                            .font(.caption)
                            .foregroundStyle(.red)
                    }
                }

                Section("Priority") {
                    Picker("Priority", selection: $priority) {
                        ForEach(TaskPriority.allCases, id: \.self) { p in
                            HStack {
                                Circle()
                                    .fill(p.color)
                                    .frame(width: 10, height: 10)
                                Text(p.title)
                            }
                            .tag(p)
                        }
                    }
                    .pickerStyle(.segmented)
                }

                Section("Date") {
                    DatePicker(
                        "Due date",
                        selection: $dueDate,
                        displayedComponents: .date
                    )
                }

                Section {
                    Button(role: .destructive) {
                        showDeleteConfirmation = true
                    } label: {
                        Label("Delete Task", systemImage: "trash")
                            .frame(maxWidth: .infinity)
                    }
                }
            }
            .navigationTitle("Edit Task")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Save") {
                        saveTask()
                    }
                    .disabled(!isValid)
                }
            }
            .alert("Delete Task?", isPresented: $showDeleteConfirmation) {
                Button("Cancel", role: .cancel) {}
                Button("Delete", role: .destructive) {
                    let taskToDelete = task
                    dismiss()
                    DispatchQueue.main.asyncAfter(deadline: .now() + 0.3) {
                        dataManager.deleteTask(taskToDelete)
                    }
                }
            } message: {
                Text("This action cannot be undone.")
            }
        }
        .presentationDetents([.large])
    }

    private func saveTask() {
        var updatedTask = task
        updatedTask.title = title.trimmingCharacters(in: .whitespaces)
        updatedTask.priority = priority
        updatedTask.dueDate = dueDate

        dataManager.updateTask(updatedTask)
        dismiss()
    }
}

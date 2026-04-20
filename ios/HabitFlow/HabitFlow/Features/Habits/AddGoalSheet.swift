import SwiftUI

struct AddGoalSheet: View {
    @EnvironmentObject var dataManager: DataManager
    @Environment(\.dismiss) var dismiss

    @State private var title = ""
    @State private var selectedIcon = "🎯"
    @State private var hasTarget = false
    @State private var targetValue = ""
    @State private var unit = ""
    @State private var hasDeadline = false
    @State private var deadline = Date().addingTimeInterval(30 * 24 * 60 * 60) // 30 days from now

    let icons = ["🎯", "💪", "🏃", "💰", "📚", "🧘", "💧", "🍎", "😴", "✍️", "🎨", "🎵", "🧠", "🏆", "⭐️", "🚀", "💎", "🌟"]
    let commonUnits = ["times", "minutes", "hours", "km", "dollars", "pages", "sessions"]

    var body: some View {
        NavigationStack {
            Form {
                Section {
                    TextField("Goal name", text: $title)
                        .font(.system(size: 17))
                } header: {
                    Text("What do you want to achieve?")
                } footer: {
                    Text("Example: Get healthy, Save $10,000, Read 50 books")
                        .font(.caption)
                }

                Section("Icon") {
                    LazyVGrid(columns: Array(repeating: GridItem(.flexible()), count: 6), spacing: 12) {
                        ForEach(icons, id: \.self) { icon in
                            Text(icon)
                                .font(.system(size: 28))
                                .frame(width: 44, height: 44)
                                .background(selectedIcon == icon ? Color.hf.accent.opacity(0.2) : Color.clear)
                                .clipShape(RoundedRectangle(cornerRadius: 10))
                                .onTapGesture {
                                    selectedIcon = icon
                                }
                        }
                    }
                }

                Section {
                    Toggle("Set a target", isOn: $hasTarget)

                    if hasTarget {
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
                } header: {
                    Text("Measurable Goal")
                } footer: {
                    if hasTarget {
                        Text("Track progress toward a specific target")
                    }
                }

                Section {
                    Toggle("Set a deadline", isOn: $hasDeadline)

                    if hasDeadline {
                        DatePicker("Deadline", selection: $deadline, in: Date()..., displayedComponents: .date)
                    }
                } header: {
                    Text("Timeline")
                }
            }
            .navigationTitle("New Goal")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Add") {
                        let goal = Goal(
                            title: title,
                            icon: selectedIcon,
                            targetValue: hasTarget ? Int(targetValue) : nil,
                            unit: hasTarget && !unit.isEmpty ? unit : nil,
                            deadline: hasDeadline ? deadline : nil
                        )
                        dataManager.addGoal(goal)
                        HapticManager.completionSuccess()
                        dismiss()
                    }
                    .disabled(title.isEmpty)
                }
            }
        }
    }
}

struct EditGoalSheet: View {
    @EnvironmentObject var dataManager: DataManager
    @Environment(\.dismiss) var dismiss

    let goal: Goal

    @State private var title: String
    @State private var selectedIcon: String
    @State private var hasTarget: Bool
    @State private var targetValue: String
    @State private var unit: String
    @State private var hasDeadline: Bool
    @State private var deadline: Date

    let icons = ["🎯", "💪", "🏃", "💰", "📚", "🧘", "💧", "🍎", "😴", "✍️", "🎨", "🎵", "🧠", "🏆", "⭐️", "🚀", "💎", "🌟"]
    let commonUnits = ["times", "minutes", "hours", "km", "dollars", "pages", "sessions"]

    init(goal: Goal) {
        self.goal = goal
        _title = State(initialValue: goal.title)
        _selectedIcon = State(initialValue: goal.icon)
        _hasTarget = State(initialValue: goal.targetValue != nil)
        _targetValue = State(initialValue: goal.targetValue.map { String($0) } ?? "")
        _unit = State(initialValue: goal.unit ?? "")
        _hasDeadline = State(initialValue: goal.deadline != nil)
        _deadline = State(initialValue: goal.deadline ?? Date().addingTimeInterval(30 * 24 * 60 * 60))
    }

    var body: some View {
        NavigationStack {
            Form {
                Section {
                    TextField("Goal name", text: $title)
                        .font(.system(size: 17))
                }

                Section("Icon") {
                    LazyVGrid(columns: Array(repeating: GridItem(.flexible()), count: 6), spacing: 12) {
                        ForEach(icons, id: \.self) { icon in
                            Text(icon)
                                .font(.system(size: 28))
                                .frame(width: 44, height: 44)
                                .background(selectedIcon == icon ? Color.hf.accent.opacity(0.2) : Color.clear)
                                .clipShape(RoundedRectangle(cornerRadius: 10))
                                .onTapGesture {
                                    selectedIcon = icon
                                }
                        }
                    }
                }

                Section("Measurable Goal") {
                    Toggle("Set a target", isOn: $hasTarget)

                    if hasTarget {
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

                Section("Timeline") {
                    Toggle("Set a deadline", isOn: $hasDeadline)

                    if hasDeadline {
                        DatePicker("Deadline", selection: $deadline, in: Date()..., displayedComponents: .date)
                    }
                }

                Section {
                    Button(role: .destructive) {
                        dataManager.archiveGoal(goal)
                        dismiss()
                    } label: {
                        HStack {
                            Image(systemName: "archivebox")
                            Text("Archive Goal")
                        }
                    }
                }
            }
            .navigationTitle("Edit Goal")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Save") {
                        var updatedGoal = goal
                        updatedGoal.title = title
                        updatedGoal.icon = selectedIcon
                        updatedGoal.targetValue = hasTarget ? Int(targetValue) : nil
                        updatedGoal.unit = hasTarget && !unit.isEmpty ? unit : nil
                        updatedGoal.deadline = hasDeadline ? deadline : nil
                        dataManager.updateGoal(updatedGoal)
                        dismiss()
                    }
                    .disabled(title.isEmpty)
                }
            }
        }
    }
}

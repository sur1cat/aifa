import SwiftUI

struct HabitsView: View {
    @EnvironmentObject var dataManager: DataManager
    @EnvironmentObject var authManager: AuthManager
    @Environment(\.colorScheme) var colorScheme
    @State private var showAddSheet = false
    @State private var showAddGoalSheet = false
    @State private var selectedHabit: Habit?
    @State private var selectedGoal: Goal?
    @State private var habitForProgressInput: Habit?
    @State private var progressInputValue = ""
    @State private var selectedDate = Date()
    @State private var showArchived = false
    @State private var showArchivedGoals = false
    @State private var showAIChat = false

    private var archivedHabits: [Habit] {
        dataManager.habits.filter { $0.archivedAt != nil }
    }

    private var activeGoals: [Goal] {
        dataManager.activeGoals
    }

    private var selectedDateString: String {
        Habit.dateString(from: selectedDate)
    }

    private var isToday: Bool {
        Calendar.current.isDateInToday(selectedDate)
    }

    // Filter habits that were active on the selected date
    private var habitsForSelectedDate: [Habit] {
        let calendar = Calendar.current
        return dataManager.habits.filter { habit in
            // Show habit if: createdAt <= selectedDate AND (archivedAt == nil OR selectedDate < archivedAt)
            let createdBefore = calendar.compare(habit.createdAt, to: selectedDate, toGranularity: .day) != .orderedDescending
            let wasActiveOnDate: Bool
            if let archivedAt = habit.archivedAt {
                // Show only if selected date is BEFORE archive date
                wasActiveOnDate = calendar.compare(selectedDate, to: archivedAt, toGranularity: .day) == .orderedAscending
            } else {
                wasActiveOnDate = true
            }
            return createdBefore && wasActiveOnDate
        }
    }

    // Filter habits for a specific goal on the selected date
    private func habitsForGoalOnDate(_ goal: Goal) -> [Habit] {
        habitsForSelectedDate.filter { $0.goalId == goal.id }
    }

    // Habits without goal on the selected date
    private var habitsWithoutGoalOnDate: [Habit] {
        habitsForSelectedDate.filter { $0.goalId == nil }
    }

    // Goals that have habits on the selected date
    private var goalsWithHabitsOnDate: [Goal] {
        activeGoals.filter { goal in
            !habitsForGoalOnDate(goal).isEmpty
        }
    }

    // Add habit from AI suggestion
    private func addHabitFromSuggestion(_ suggestion: SuggestedHabit, _ goal: Goal) {
        let period: HabitPeriod = suggestion.period.lowercased() == "weekly" ? .weekly : .daily
        let habit = Habit(
            goalId: goal.id,
            title: suggestion.title,
            icon: suggestion.icon,
            color: suggestion.color,
            period: period
        )
        dataManager.addHabit(habit)
    }

    var body: some View {
        NavigationStack {
            List {
                // Date Selector
                Section {
                    HabitsDateSelector(selectedDate: $selectedDate)
                        .listRowBackground(Color.clear)
                        .listRowInsets(EdgeInsets(top: 0, leading: 0, bottom: 8, trailing: 0))
                }
                .listRowSeparator(.hidden)

                // Insights Carousel (WHOOP-style)
                if !dataManager.insights(for: .habits).isEmpty {
                    Section {
                        InsightCarousel(
                            insights: dataManager.insights(for: .habits),
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

                if dataManager.isLoading && dataManager.habits.isEmpty {
                    Section {
                        ProgressView()
                            .frame(maxWidth: .infinity)
                            .padding(.top, 100)
                            .listRowBackground(Color.clear)
                    }
                    .listRowSeparator(.hidden)
                } else if habitsForSelectedDate.isEmpty {
                    Section {
                        ContentUnavailableView {
                            Label(dataManager.habits.isEmpty ? "No habits yet" : "No habits on this date", systemImage: "repeat")
                        } description: {
                            Text(dataManager.habits.isEmpty ? "Tap + to add your first habit" : "This habit was created later")
                        }
                        .listRowBackground(Color.clear)
                        .listRowInsets(EdgeInsets(top: 100, leading: 0, bottom: 0, trailing: 0))
                    }
                    .listRowSeparator(.hidden)
                } else {
                    // All active goals (with or without habits)
                    ForEach(activeGoals) { goal in
                        Section {
                            GoalSection(
                                goal: goal,
                                habits: habitsForGoalOnDate(goal),
                                dateString: selectedDateString,
                                isToday: isToday,
                                onHabitToggle: { habit in
                                    dataManager.toggleHabitForDate(habit, date: selectedDate)
                                },
                                onHabitProgressTap: { habit in
                                    dataManager.incrementHabitProgressForDate(habit, date: selectedDate)
                                },
                                onHabitProgressLongPress: { habit in
                                    habitForProgressInput = habit
                                    progressInputValue = "\(habit.progressValues[selectedDateString] ?? 0)"
                                },
                                onHabitTap: { habit in
                                    selectedHabit = habit
                                },
                                onGoalTap: { goal in
                                    selectedGoal = goal
                                },
                                onAddSuggestedHabit: addHabitFromSuggestion
                            )
                            .listRowBackground(Color.clear)
                            .listRowInsets(EdgeInsets(top: 6, leading: 16, bottom: 6, trailing: 16))
                        }
                        .listRowSeparator(.hidden)
                    }

                    // Habits without goals
                    if !habitsWithoutGoalOnDate.isEmpty {
                        Section {
                            OtherHabitsSection(
                                habits: habitsWithoutGoalOnDate,
                                dateString: selectedDateString,
                                isToday: isToday,
                                onHabitToggle: { habit in
                                    dataManager.toggleHabitForDate(habit, date: selectedDate)
                                },
                                onHabitProgressTap: { habit in
                                    dataManager.incrementHabitProgressForDate(habit, date: selectedDate)
                                },
                                onHabitProgressLongPress: { habit in
                                    habitForProgressInput = habit
                                    progressInputValue = "\(habit.progressValues[selectedDateString] ?? 0)"
                                },
                                onHabitTap: { habit in
                                    selectedHabit = habit
                                }
                            )
                            .listRowBackground(Color.clear)
                            .listRowInsets(EdgeInsets(top: 6, leading: 16, bottom: 6, trailing: 16))
                        }
                        .listRowSeparator(.hidden)
                    }
                }
            }
            .listStyle(.plain)
            .scrollContentBackground(.hidden)
            .background(AppTheme.appBackground(for: colorScheme))
            .navigationTitle("Atoma Habits")
            .onAppear {
                dataManager.generateInsights(for: .habits)
            }
            .toolbar {
                ToolbarItem(placement: .topBarLeading) {
                    Button {
                        showAIChat = true
                    } label: {
                        Image(systemName: "atom")
                            .fontWeight(.semibold)
                            .foregroundStyle(Color.hf.accent)
                    }
                }
                ToolbarItem(placement: .topBarTrailing) {
                    HStack(spacing: 16) {
                        if !archivedHabits.isEmpty || !dataManager.archivedGoals.isEmpty {
                            Menu {
                                if !archivedHabits.isEmpty {
                                    Button {
                                        showArchived = true
                                    } label: {
                                        Label("Archived Habits", systemImage: "repeat")
                                    }
                                }
                                if !dataManager.archivedGoals.isEmpty {
                                    Button {
                                        showArchivedGoals = true
                                    } label: {
                                        Label("Archived Goals", systemImage: "flag")
                                    }
                                }
                            } label: {
                                Image(systemName: "archivebox")
                                    .fontWeight(.semibold)
                            }
                            .tint(.secondary)
                        }

                        Menu {
                            Button {
                                showAddSheet = true
                            } label: {
                                Label("Add Habit", systemImage: "repeat")
                            }

                            Button {
                                showAddGoalSheet = true
                            } label: {
                                Label("Add Goal", systemImage: "flag")
                            }
                        } label: {
                            Image(systemName: "plus")
                                .fontWeight(.semibold)
                        }
                        .tint(.primary)
                    }
                }
            }
            .sheet(isPresented: $showAddSheet) {
                AddHabitSheet()
            }
            .sheet(isPresented: $showAddGoalSheet) {
                AddGoalSheet()
            }
            .sheet(isPresented: $showArchived) {
                ArchivedHabitsSheet()
            }
            .sheet(isPresented: $showArchivedGoals) {
                ArchivedGoalsSheet()
            }
            .fullScreenCover(isPresented: $showAIChat) {
                AIChatView(agent: .habitCoach) {
                    buildHabitsContext()
                }
            }
            .sheet(item: $selectedHabit) { habit in
                EditHabitSheet(habit: habit)
            }
            .sheet(item: $selectedGoal) { goal in
                EditGoalSheet(goal: goal)
            }
            .task {
                if authManager.isAuthenticated {
                    await dataManager.syncGoals()
                    await dataManager.syncHabits()
                }
            }
            .refreshable {
                if authManager.isAuthenticated {
                    await dataManager.syncGoals()
                    await dataManager.syncHabits()
                }
            }
            .sheet(item: $habitForProgressInput) { habit in
                ProgressInputSheet(
                    habit: habit,
                    initialValue: progressInputValue,
                    onSave: { value, date in
                        dataManager.setHabitProgress(habit, value: value, date: date)
                    }
                )
            }
        }
    }

    // MARK: - AI Context

    private func buildHabitsContext() -> String {
        let today = Date()
        let todayString = Habit.dateString(from: today)
        let activeHabits = dataManager.habits.filter { $0.archivedAt == nil }

        var context = "=== USER'S HABITS DATA ===\n"
        context += "Date: \(todayString)\n"
        context += "Total active habits: \(activeHabits.count)\n\n"

        // Goals summary
        let goals = dataManager.activeGoals
        if !goals.isEmpty {
            context += "GOALS (\(goals.count)):\n"
            for goal in goals {
                let goalHabits = activeHabits.filter { $0.goalId == goal.id }
                let completed = goalHabits.filter { $0.completedDates.contains(todayString) }.count
                context += "- \(goal.icon) \(goal.title): \(completed)/\(goalHabits.count) habits done today\n"
            }
            context += "\n"
        }

        // Habits details
        context += "HABITS:\n"
        for habit in activeHabits {
            let isCompleted = habit.completedDates.contains(todayString)
            let status = isCompleted ? "✓" : "○"

            var habitInfo = "- \(status) \(habit.icon) \(habit.title)"

            if let target = habit.targetValue, let unit = habit.unit {
                let progress = habit.progressValues[todayString] ?? 0
                habitInfo += " (\(progress)/\(target) \(unit))"
            }

            if habit.streak > 0 {
                habitInfo += " 🔥\(habit.streak) day streak"
            }

            habitInfo += " [\(habit.period.title)]"
            context += habitInfo + "\n"
        }

        // Weekly stats
        let calendar = Calendar.current
        var weekCompleted = 0
        var weekTotal = 0
        for dayOffset in 0..<7 {
            if let date = calendar.date(byAdding: .day, value: -dayOffset, to: today) {
                let dateStr = Habit.dateString(from: date)
                for habit in activeHabits {
                    weekTotal += 1
                    if habit.completedDates.contains(dateStr) {
                        weekCompleted += 1
                    }
                }
            }
        }
        let weekRate = weekTotal > 0 ? Int(Double(weekCompleted) / Double(weekTotal) * 100) : 0
        context += "\nWEEKLY COMPLETION RATE: \(weekRate)%"

        return context
    }
}

struct ProgressInputSheet: View {
    let habit: Habit
    let initialValue: String
    let onSave: (Int, Date) -> Void

    @Environment(\.dismiss) var dismiss
    @State private var inputValue: String
    @State private var selectedDate: Date = Date()

    init(habit: Habit, initialValue: String, onSave: @escaping (Int, Date) -> Void) {
        self.habit = habit
        self.initialValue = initialValue
        self.onSave = onSave
        _inputValue = State(initialValue: initialValue)
    }

    var body: some View {
        NavigationStack {
            VStack(spacing: 20) {
                // Header
                VStack(spacing: 8) {
                    Text(habit.icon)
                        .font(.system(size: 48))
                    Text(habit.title)
                        .font(.headline)
                }
                .padding(.top, 16)

                // Date picker
                HStack {
                    Text("Date")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                    Spacer()
                    DatePicker("", selection: $selectedDate, in: ...Date(), displayedComponents: .date)
                        .labelsHidden()
                }
                .padding(.horizontal)

                // Progress input
                VStack(spacing: 12) {
                    HStack(spacing: 16) {
                        Button {
                            if let current = Int(inputValue), current > 0 {
                                inputValue = "\(current - 1)"
                            }
                        } label: {
                            Image(systemName: "minus.circle.fill")
                                .font(.system(size: 36))
                                .foregroundStyle(habit.swiftUIColor)
                        }

                        TextField("0", text: $inputValue)
                            .keyboardType(.numberPad)
                            .font(.system(size: 48, weight: .bold, design: .rounded))
                            .multilineTextAlignment(.center)
                            .frame(width: 100)

                        Button {
                            if let current = Int(inputValue) {
                                inputValue = "\(current + 1)"
                            } else {
                                inputValue = "1"
                            }
                        } label: {
                            Image(systemName: "plus.circle.fill")
                                .font(.system(size: 36))
                                .foregroundStyle(habit.swiftUIColor)
                        }
                    }

                    if let target = habit.targetValue, let unit = habit.unit {
                        Text("Goal: \(target) \(unit)")
                            .font(.subheadline)
                            .foregroundStyle(.secondary)
                    }
                }

                // Quick buttons
                if let target = habit.targetValue {
                    HStack(spacing: 12) {
                        QuickValueButton(value: target / 4, unit: habit.unit) {
                            inputValue = "\(target / 4)"
                        }
                        QuickValueButton(value: target / 2, unit: habit.unit) {
                            inputValue = "\(target / 2)"
                        }
                        QuickValueButton(value: target, unit: habit.unit) {
                            inputValue = "\(target)"
                        }
                    }
                }

                // Mark Complete button
                if let target = habit.targetValue {
                    Button {
                        inputValue = "\(target)"
                        if let value = Int(inputValue) {
                            onSave(value, selectedDate)
                            HapticManager.completionSuccess()
                        }
                        dismiss()
                    } label: {
                        HStack {
                            Image(systemName: "checkmark.circle.fill")
                            Text("Mark Complete")
                        }
                        .font(.system(size: 16, weight: .semibold))
                        .foregroundStyle(.white)
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 14)
                        .background(habit.swiftUIColor)
                        .clipShape(RoundedRectangle(cornerRadius: 12))
                    }
                    .padding(.horizontal)
                }

                Spacer()
            }
            .padding()
            .navigationTitle("Update Progress")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Save") {
                        if let value = Int(inputValue) {
                            onSave(value, selectedDate)
                            HapticManager.completionSuccess()
                        }
                        dismiss()
                    }
                    .disabled(Int(inputValue) == nil)
                }
            }
        }
        .presentationDetents([.medium])
    }
}

struct QuickValueButton: View {
    let value: Int
    let unit: String?
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            VStack(spacing: 2) {
                Text("\(value)")
                    .font(.system(size: 16, weight: .semibold))
                if let unit = unit {
                    Text(unit)
                        .font(.system(size: 10))
                }
            }
            .foregroundStyle(.primary)
            .frame(width: 60, height: 50)
            .background(Color.hf.cardBackground)
            .clipShape(RoundedRectangle(cornerRadius: 10))
        }
    }
}

struct HabitRow: View {
    let habit: Habit
    let dateString: String
    let isToday: Bool
    let onToggle: () -> Void
    var onProgressTap: (() -> Void)? = nil
    var onProgressLongPress: (() -> Void)? = nil

    @State private var showCelebration = false

    // Computed properties for selected date
    private var isCompletedForDate: Bool {
        if let target = habit.targetValue, target > 0 {
            let progress = habit.progressValues[dateString] ?? 0
            return progress >= target || habit.completedDates.contains(dateString)
        }
        return habit.completedDates.contains(dateString)
    }

    private var progressForDate: Int {
        habit.progressValues[dateString] ?? 0
    }

    private var progressPercentageForDate: Double {
        guard let target = habit.targetValue, target > 0 else { return 0 }
        return min(Double(progressForDate) / Double(target), 1.0)
    }

    var body: some View {
        HStack(spacing: 16) {
            // For habits with goals, show progress ring
            if habit.hasGoal, habit.targetValue != nil {
                Button {
                    HapticManager.toggleOn()
                    onProgressLongPress?()
                } label: {
                    ZStack {
                        Circle()
                            .stroke(Color.hf.checkmarkIncomplete, lineWidth: 3)
                            .frame(width: 32, height: 32)

                        Circle()
                            .trim(from: 0, to: progressPercentageForDate)
                            .stroke(habit.swiftUIColor, style: StrokeStyle(lineWidth: 3, lineCap: .round))
                            .frame(width: 32, height: 32)
                            .rotationEffect(.degrees(-90))

                        if isCompletedForDate {
                            Image(systemName: "checkmark")
                                .font(.system(size: 12, weight: .bold))
                                .foregroundStyle(habit.swiftUIColor)
                        } else {
                            Text("\(progressForDate)")
                                .font(.system(size: 10, weight: .medium))
                                .foregroundStyle(.secondary)
                        }
                    }
                }
                .buttonStyle(.plain)
            } else {
                // Simple checkbox for habits without goals
                Button {
                    let wasCompleted = isCompletedForDate
                    onToggle()
                    if !wasCompleted {
                        triggerCelebration()
                    } else {
                        HapticManager.toggleOff()
                    }
                } label: {
                    AnimatedCheckmark(
                        isCompleted: isCompletedForDate,
                        color: habit.swiftUIColor
                    )
                }
                .buttonStyle(.plain)
            }

            VStack(alignment: .leading, spacing: 4) {
                Text(habit.title)
                    .font(.system(size: 17, weight: .medium))
                    .strikethrough(isCompletedForDate, color: .secondary)
                    .foregroundStyle(isCompletedForDate ? .secondary : .primary)

                HStack(spacing: 8) {
                    // Show goal progress for habits with goals
                    if habit.hasGoal, let target = habit.targetValue, let unit = habit.unit {
                        Button {
                            HapticManager.toggleOn()
                            onProgressLongPress?()
                        } label: {
                            Text("\(progressForDate)/\(target) \(unit)")
                                .font(.system(size: 12))
                                .padding(.horizontal, 8)
                                .padding(.vertical, 2)
                                .background(habit.swiftUIColor.opacity(0.15))
                                .foregroundStyle(habit.swiftUIColor)
                                .clipShape(Capsule())
                        }
                        .buttonStyle(.plain)
                    } else {
                        Text(habit.period.title)
                            .font(.system(size: 12))
                            .padding(.horizontal, 8)
                            .padding(.vertical, 2)
                            .background(habit.swiftUIColor.opacity(0.15))
                            .foregroundStyle(habit.swiftUIColor)
                            .clipShape(Capsule())
                    }

                    if habit.streak > 0 && isToday {
                        HStack(spacing: 2) {
                            Text("\(habit.streak)")
                                .font(.system(size: 12, weight: .semibold))
                            Image(systemName: "flame.fill")
                                .font(.system(size: 10))
                                .foregroundStyle(.orange)
                        }
                        .foregroundStyle(.secondary)
                    }

                    if habit.reminderEnabled, let time = habit.reminderTime, isToday {
                        HStack(spacing: 2) {
                            Image(systemName: "bell.fill")
                                .font(.system(size: 10))
                            Text(time, style: .time)
                                .font(.system(size: 11))
                        }
                        .foregroundStyle(.secondary)
                    }
                }
            }

            Spacer()

            Text(habit.icon)
                .font(.system(size: 24))
        }
        .padding()
        .background(Color.hf.cardBackground)
        .clipShape(RoundedRectangle(cornerRadius: 16))
        .overlay(alignment: .topTrailing) {
            CelebrationEffect(isActive: showCelebration)
                .frame(width: 100, height: 60)
                .offset(x: -20, y: 10)
        }
    }

    private func triggerCelebration() {
        HapticManager.completionSuccess()
        showCelebration = true
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.1) {
            showCelebration = false
        }
    }
}

// MARK: - Habit Templates
struct HabitTemplate: Identifiable {
    let id = UUID()
    let title: String
    let icon: String
    let color: String
    let category: String
    let targetValue: Int?
    let unit: String?

    static let templates: [HabitTemplate] = [
        // Health & Fitness
        HabitTemplate(title: "Morning Meditation", icon: "🧘", color: "purple", category: "Health", targetValue: 10, unit: "minutes"),
        HabitTemplate(title: "Exercise", icon: "🏃", color: "orange", category: "Health", targetValue: 30, unit: "minutes"),
        HabitTemplate(title: "Drink Water", icon: "💧", color: "blue", category: "Health", targetValue: 8, unit: "glasses"),
        HabitTemplate(title: "Sleep 8 Hours", icon: "😴", color: "purple", category: "Health", targetValue: nil, unit: nil),
        HabitTemplate(title: "Take Vitamins", icon: "💊", color: "green", category: "Health", targetValue: nil, unit: nil),

        // Productivity
        HabitTemplate(title: "Read", icon: "📚", color: "blue", category: "Productivity", targetValue: 20, unit: "pages"),
        HabitTemplate(title: "Journal", icon: "✍️", color: "orange", category: "Productivity", targetValue: nil, unit: nil),
        HabitTemplate(title: "No Social Media", icon: "📵", color: "red", category: "Productivity", targetValue: nil, unit: nil),
        HabitTemplate(title: "Learn Something New", icon: "🧠", color: "purple", category: "Productivity", targetValue: 15, unit: "minutes"),

        // Mindfulness
        HabitTemplate(title: "Gratitude", icon: "🙏", color: "green", category: "Mindfulness", targetValue: 3, unit: "things"),
        HabitTemplate(title: "Deep Breathing", icon: "🌬️", color: "blue", category: "Mindfulness", targetValue: 5, unit: "minutes"),
        HabitTemplate(title: "No Phone Before Bed", icon: "🌙", color: "purple", category: "Mindfulness", targetValue: nil, unit: nil),
    ]

    static var categories: [String] {
        Array(Set(templates.map { $0.category })).sorted()
    }

    static func templates(for category: String) -> [HabitTemplate] {
        templates.filter { $0.category == category }
    }
}

struct AddHabitSheet: View {
    @EnvironmentObject var dataManager: DataManager
    @Environment(\.dismiss) var dismiss
    @State private var title = ""
    @State private var selectedIcon = "🎯"
    @State private var selectedColor = "green"
    @State private var selectedPeriod: HabitPeriod = .daily
    @State private var selectedGoalId: UUID?
    @State private var hasGoal = false
    @State private var targetValue = ""
    @State private var unit = ""
    @State private var showCustomForm = false
    @State private var selectedTemplate: HabitTemplate?

    let icons = ["🎯", "💪", "📚", "🧘", "🏃", "💧", "🍎", "😴", "✍️", "🎨", "🎵", "🧹", "💊", "📵", "🧠", "🙏", "🌬️", "🌙"]
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

    var body: some View {
        NavigationStack {
            Form {
                // Quick Templates Section
                if !showCustomForm {
                    Section {
                        Button {
                            withAnimation {
                                showCustomForm = true
                            }
                        } label: {
                            HStack {
                                Image(systemName: "plus.circle.fill")
                                    .foregroundStyle(Color.hf.accent)
                                Text("Create Custom Habit")
                                    .foregroundStyle(.primary)
                                Spacer()
                                Image(systemName: "chevron.right")
                                    .font(.system(size: 13, weight: .semibold))
                                    .foregroundStyle(.tertiary)
                            }
                        }
                    }

                    ForEach(HabitTemplate.categories, id: \.self) { category in
                        Section(category) {
                            ForEach(HabitTemplate.templates(for: category)) { template in
                                Button {
                                    if dataManager.activeGoals.isEmpty {
                                        addFromTemplate(template, goalId: nil)
                                    } else {
                                        selectedTemplate = template
                                    }
                                } label: {
                                    HStack(spacing: 12) {
                                        Text(template.icon)
                                            .font(.system(size: 24))

                                        VStack(alignment: .leading, spacing: 2) {
                                            Text(template.title)
                                                .font(.system(size: 15, weight: .medium))
                                                .foregroundStyle(.primary)

                                            if let target = template.targetValue, let unit = template.unit {
                                                Text("\(target) \(unit)/day")
                                                    .font(.system(size: 12))
                                                    .foregroundStyle(.secondary)
                                            }
                                        }

                                        Spacer()

                                        Image(systemName: "plus.circle")
                                            .foregroundStyle(Color.hf.accent)
                                    }
                                }
                            }
                        }
                    }
                }

                // Custom Form Section
                if showCustomForm {
                    Section {
                        Button {
                            withAnimation {
                                showCustomForm = false
                            }
                        } label: {
                            HStack {
                                Image(systemName: "chevron.left")
                                    .font(.system(size: 13, weight: .semibold))
                                Text("Back to Templates")
                            }
                            .foregroundStyle(Color.hf.accent)
                        }
                    }

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
                }
            }
            .navigationTitle(showCustomForm ? "Custom Habit" : "New Habit")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                if showCustomForm {
                    ToolbarItem(placement: .confirmationAction) {
                        Button("Add") {
                            let habit = Habit(
                                goalId: selectedGoalId,
                                title: title.trimmingCharacters(in: .whitespaces),
                                icon: selectedIcon,
                                color: selectedColor,
                                period: selectedPeriod,
                                targetValue: hasGoal ? Int(targetValue) : nil,
                                unit: hasGoal && !unit.isEmpty ? unit : nil
                            )
                            dataManager.addHabit(habit)
                            HapticManager.completionSuccess()
                            dismiss()
                        }
                        .disabled(!isValid)
                    }
                }
            }
            .sheet(item: $selectedTemplate) { template in
                TemplateGoalPickerSheet(template: template) { goalId in
                    addFromTemplate(template, goalId: goalId)
                }
            }
        }
    }

    private func addFromTemplate(_ template: HabitTemplate, goalId: UUID?) {
        let habit = Habit(
            goalId: goalId,
            title: template.title,
            icon: template.icon,
            color: template.color,
            period: .daily,
            targetValue: template.targetValue,
            unit: template.unit
        )
        dataManager.addHabit(habit)
        HapticManager.completionSuccess()
        dismiss()
    }
}

// MARK: - Template Goal Picker Sheet
struct TemplateGoalPickerSheet: View {
    @EnvironmentObject var dataManager: DataManager
    @Environment(\.dismiss) var dismiss
    let template: HabitTemplate
    let onAdd: (UUID?) -> Void

    var body: some View {
        NavigationStack {
            List {
                Section {
                    HStack(spacing: 12) {
                        Text(template.icon)
                            .font(.system(size: 32))
                        VStack(alignment: .leading) {
                            Text(template.title)
                                .font(.headline)
                            if let target = template.targetValue, let unit = template.unit {
                                Text("\(target) \(unit)/day")
                                    .font(.subheadline)
                                    .foregroundStyle(.secondary)
                            }
                        }
                    }
                    .padding(.vertical, 8)
                }

                Section("Link to Goal") {
                    Button {
                        onAdd(nil)
                        dismiss()
                    } label: {
                        HStack {
                            Text("No Goal")
                            Spacer()
                            Image(systemName: "plus.circle.fill")
                                .foregroundStyle(Color.hf.accent)
                        }
                    }

                    ForEach(dataManager.activeGoals) { goal in
                        Button {
                            onAdd(goal.id)
                            dismiss()
                        } label: {
                            HStack {
                                Text("\(goal.icon) \(goal.title)")
                                Spacer()
                                Image(systemName: "plus.circle.fill")
                                    .foregroundStyle(Color.hf.accent)
                            }
                        }
                    }
                }
            }
            .navigationTitle("Select Goal")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
            }
        }
        .presentationDetents([.medium])
    }
}

// MARK: - Habits Date Selector
struct HabitsDateSelector: View {
    @Binding var selectedDate: Date

    private let calendar = Calendar.current

    private var dates: [Date] {
        (-3...3).compactMap { offset in
            calendar.date(byAdding: .day, value: offset, to: selectedDate)
        }
    }

    private var isViewingToday: Bool {
        calendar.isDateInToday(selectedDate)
    }

    var body: some View {
        VStack(spacing: 8) {
            // Month/Year header with Today button
            HStack {
                Text(monthYearString)
                    .font(.system(size: 14, weight: .medium))
                    .foregroundStyle(.secondary)

                Spacer()

                // Today button integrated into header
                if !isViewingToday {
                    Button {
                        withAnimation {
                            selectedDate = Date()
                        }
                    } label: {
                        Text("Today")
                            .font(.system(size: 13, weight: .medium))
                            .foregroundStyle(Color.hf.accent)
                            .padding(.horizontal, 12)
                            .padding(.vertical, 6)
                            .background(Color.hf.accent.opacity(0.12))
                            .clipShape(Capsule())
                    }
                }
            }
            .padding(.horizontal, 16)

            HStack(spacing: 8) {
                // Left arrow
                Button {
                    withAnimation {
                        if let newDate = calendar.date(byAdding: .day, value: -7, to: selectedDate) {
                            selectedDate = newDate
                        }
                    }
                } label: {
                    Image(systemName: "chevron.left")
                        .font(.system(size: 14, weight: .semibold))
                        .foregroundStyle(.secondary)
                        .frame(width: 24, height: 56)
                }

                ForEach(dates, id: \.self) { date in
                    HabitsDateCell(
                        date: date,
                        isSelected: calendar.isDate(date, inSameDayAs: selectedDate),
                        isToday: calendar.isDateInToday(date)
                    ) {
                        withAnimation(.easeInOut(duration: 0.2)) {
                            selectedDate = date
                        }
                    }
                }

                // Right arrow
                Button {
                    withAnimation {
                        if let newDate = calendar.date(byAdding: .day, value: 7, to: selectedDate) {
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
        }
    }

    private var monthYearString: String {
        let formatter = DateFormatter()
        formatter.dateFormat = "MMMM yyyy"
        return formatter.string(from: selectedDate)
    }
}

struct HabitsDateCell: View {
    let date: Date
    let isSelected: Bool
    let isToday: Bool
    let action: () -> Void

    private let calendar = Calendar.current

    var body: some View {
        Button(action: action) {
            VStack(spacing: 4) {
                Text(dayOfWeek)
                    .font(.system(size: 11, weight: .medium))
                    .foregroundStyle(isSelected ? .white : .secondary)

                Text("\(calendar.component(.day, from: date))")
                    .font(.system(size: 18, weight: isSelected ? .bold : .semibold, design: .rounded))
                    .foregroundStyle(isSelected ? .white : (isToday ? Color.hf.accent : .primary))
            }
            .frame(width: 44, height: 56)
            .background(isSelected ? Color.hf.accent : Color.hf.cardBackground)
            .clipShape(RoundedRectangle(cornerRadius: 12))
            .overlay(
                RoundedRectangle(cornerRadius: 12)
                    .stroke(isToday && !isSelected ? Color.hf.accent : Color.clear, lineWidth: 2)
            )
        }
        .buttonStyle(.plain)
    }

    private var dayOfWeek: String {
        let formatter = DateFormatter()
        formatter.dateFormat = "EEE"
        return formatter.string(from: date).uppercased()
    }
}

// MARK: - Archived Habits Sheet
struct ArchivedHabitsSheet: View {
    @EnvironmentObject var dataManager: DataManager
    @Environment(\.dismiss) var dismiss

    private var archivedHabits: [Habit] {
        dataManager.habits.filter { $0.archivedAt != nil }
            .sorted { ($0.archivedAt ?? Date()) > ($1.archivedAt ?? Date()) }
    }

    var body: some View {
        NavigationStack {
            List {
                if archivedHabits.isEmpty {
                    ContentUnavailableView {
                        Label("No Archived Habits", systemImage: "archivebox")
                    } description: {
                        Text("Archived habits will appear here")
                    }
                } else {
                    ForEach(archivedHabits) { habit in
                        HStack(spacing: 12) {
                            Text(habit.icon)
                                .font(.system(size: 28))

                            VStack(alignment: .leading, spacing: 4) {
                                Text(habit.title)
                                    .font(.system(size: 16, weight: .medium))

                                if let archivedAt = habit.archivedAt {
                                    Text("Archived \(archivedAt.formatted(date: .abbreviated, time: .omitted))")
                                        .font(.system(size: 12))
                                        .foregroundStyle(.secondary)
                                }
                            }

                            Spacer()

                            Image(systemName: "chevron.left")
                                .font(.system(size: 12))
                                .foregroundStyle(.tertiary)
                        }
                        .padding(.vertical, 4)
                        .swipeActions(edge: .trailing) {
                            Button {
                                dataManager.unarchiveHabit(habit)
                            } label: {
                                Label("Restore", systemImage: "arrow.uturn.backward")
                            }
                            .tint(Color.hf.accent)
                        }
                    }
                }
            }
            .navigationTitle("Archived Habits")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .confirmationAction) {
                    Button("Done") {
                        dismiss()
                    }
                }
            }
        }
    }
}

// MARK: - Archived Goals Sheet
struct ArchivedGoalsSheet: View {
    @EnvironmentObject var dataManager: DataManager
    @Environment(\.dismiss) var dismiss

    private var archivedGoals: [Goal] {
        dataManager.archivedGoals
            .sorted { ($0.archivedAt ?? Date()) > ($1.archivedAt ?? Date()) }
    }

    var body: some View {
        NavigationStack {
            List {
                if archivedGoals.isEmpty {
                    ContentUnavailableView {
                        Label("No Archived Goals", systemImage: "flag.slash")
                    } description: {
                        Text("Archived goals will appear here")
                    }
                } else {
                    ForEach(archivedGoals) { goal in
                        HStack(spacing: 12) {
                            Text(goal.icon)
                                .font(.system(size: 28))

                            VStack(alignment: .leading, spacing: 4) {
                                Text(goal.title)
                                    .font(.system(size: 16, weight: .medium))

                                if let archivedAt = goal.archivedAt {
                                    Text("Archived \(archivedAt.formatted(date: .abbreviated, time: .omitted))")
                                        .font(.system(size: 12))
                                        .foregroundStyle(.secondary)
                                }
                            }

                            Spacer()

                            Image(systemName: "chevron.left")
                                .font(.system(size: 12))
                                .foregroundStyle(.tertiary)
                        }
                        .padding(.vertical, 4)
                        .swipeActions(edge: .trailing) {
                            Button {
                                dataManager.unarchiveGoal(goal)
                            } label: {
                                Label("Restore", systemImage: "arrow.uturn.backward")
                            }
                            .tint(Color.hf.accent)
                        }
                    }
                }
            }
            .navigationTitle("Archived Goals")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .confirmationAction) {
                    Button("Done") {
                        dismiss()
                    }
                }
            }
        }
    }
}

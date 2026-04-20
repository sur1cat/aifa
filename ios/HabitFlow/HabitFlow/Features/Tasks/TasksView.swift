import SwiftUI

struct TasksView: View {
    @EnvironmentObject var dataManager: DataManager
    @EnvironmentObject var authManager: AuthManager
    @Environment(\.colorScheme) var colorScheme
    @State private var selectedDate = Date()
    @State private var showAddSheet = false
    @State private var selectedTask: DailyTask?
    @State private var showAIChat = false

    private var tasksForSelectedDate: [DailyTask] {
        let calendar = Calendar.current
        return dataManager.tasks
            .filter { calendar.isDate($0.dueDate, inSameDayAs: selectedDate) }
            .sorted { t1, t2 in
                if t1.isCompleted != t2.isCompleted {
                    return !t1.isCompleted
                }
                return t1.priority.sortOrder < t2.priority.sortOrder
            }
    }

    private var completedCount: Int {
        tasksForSelectedDate.filter { $0.isCompleted }.count
    }

    private var totalCount: Int {
        tasksForSelectedDate.count
    }

    var body: some View {
        NavigationStack {
            List {
                // Tagline
                Section {
                    Text("finish what matters")
                        .font(.system(size: 15))
                        .foregroundStyle(.secondary)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .listRowBackground(Color.clear)
                        .listRowInsets(EdgeInsets(top: 0, leading: 16, bottom: 8, trailing: 16))
                }
                .listRowSeparator(.hidden)

                // Date Selector
                Section {
                    TasksDateSelector(selectedDate: $selectedDate)
                        .listRowBackground(Color.clear)
                        .listRowInsets(EdgeInsets(top: 0, leading: 8, bottom: 12, trailing: 8))
                }
                .listRowSeparator(.hidden)

                // Insights Carousel (WHOOP-style)
                if !dataManager.insights(for: .tasks).isEmpty {
                    Section {
                        InsightCarousel(
                            insights: dataManager.insights(for: .tasks),
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

                if dataManager.isLoading && dataManager.tasks.isEmpty {
                    Section {
                        ProgressView()
                            .frame(maxWidth: .infinity)
                            .padding(.top, 60)
                            .listRowBackground(Color.clear)
                    }
                    .listRowSeparator(.hidden)
                } else if tasksForSelectedDate.isEmpty {
                    Section {
                        TasksEmptyState(isToday: Calendar.current.isDateInToday(selectedDate))
                            .listRowBackground(Color.clear)
                            .listRowInsets(EdgeInsets(top: 40, leading: 16, bottom: 0, trailing: 16))
                    }
                    .listRowSeparator(.hidden)
                } else {
                    // Progress Card
                    Section {
                        TasksProgressCard(
                            completed: completedCount,
                            total: totalCount,
                            isToday: Calendar.current.isDateInToday(selectedDate)
                        )
                        .listRowBackground(Color.clear)
                        .listRowInsets(EdgeInsets(top: 0, leading: 16, bottom: 12, trailing: 16))
                    }
                    .listRowSeparator(.hidden)

                    // Tasks List
                    Section {
                        ForEach(tasksForSelectedDate) { task in
                            TaskRow(task: task) {
                                dataManager.toggleTask(task)
                            }
                            .listRowBackground(Color.clear)
                            .listRowInsets(EdgeInsets(top: 6, leading: 16, bottom: 6, trailing: 16))
                            .onTapGesture {
                                selectedTask = task
                            }
                        }
                    }
                    .listRowSeparator(.hidden)
                }
            }
            .listStyle(.plain)
            .scrollContentBackground(.hidden)
            .background(AppTheme.appBackground(for: colorScheme))
            .navigationTitle("Atoma Tasks")
            .onAppear {
                dataManager.generateInsights(for: .tasks)
            }
            .toolbar {
                ToolbarItem(placement: .topBarLeading) {
                    Button {
                        showAIChat = true
                    } label: {
                        Image(systemName: "atom")
                            .fontWeight(.semibold)
                            .foregroundStyle(Color.hf.info)
                    }
                }
                ToolbarItem(placement: .topBarTrailing) {
                    Button {
                        showAddSheet = true
                    } label: {
                        Image(systemName: "plus")
                            .fontWeight(.semibold)
                    }
                    .tint(.primary)
                }
            }
            .sheet(isPresented: $showAddSheet) {
                AddTaskSheet(initialDate: selectedDate)
            }
            .sheet(item: $selectedTask) { task in
                EditTaskSheet(task: task)
            }
            .fullScreenCover(isPresented: $showAIChat) {
                AIChatView(agent: .taskAssistant) {
                    buildTasksContext()
                }
            }
            .task {
                if authManager.isAuthenticated {
                    await dataManager.syncTasks()
                }
            }
            .refreshable {
                if authManager.isAuthenticated {
                    await dataManager.syncTasks()
                }
            }
        }
    }

    // MARK: - AI Context

    private func buildTasksContext() -> String {
        let today = Date()
        let calendar = Calendar.current
        let dateFormatter = DateFormatter()
        dateFormatter.dateFormat = "yyyy-MM-dd"

        var context = "=== USER'S TASKS DATA ===\n"
        context += "Today: \(dateFormatter.string(from: today))\n\n"

        // Today's tasks
        let todayTasks = dataManager.tasksForDate(today)
        let todayCompleted = todayTasks.filter { $0.isCompleted }.count
        context += "TODAY'S TASKS (\(todayCompleted)/\(todayTasks.count) completed):\n"

        let sortedTasks = todayTasks.sorted { t1, t2 in
            if t1.isCompleted != t2.isCompleted { return !t1.isCompleted }
            return t1.priority.sortOrder < t2.priority.sortOrder
        }

        for task in sortedTasks {
            let status = task.isCompleted ? "✓" : "○"
            let priority = task.priority.rawValue.uppercased()
            context += "- \(status) [\(priority)] \(task.title)\n"
        }

        // Upcoming tasks (next 7 days)
        context += "\nUPCOMING TASKS:\n"
        for dayOffset in 1...7 {
            if let date = calendar.date(byAdding: .day, value: dayOffset, to: today) {
                let tasks = dataManager.tasksForDate(date)
                if !tasks.isEmpty {
                    let dayName = dayOffset == 1 ? "Tomorrow" : dateFormatter.string(from: date)
                    let pending = tasks.filter { !$0.isCompleted }
                    if !pending.isEmpty {
                        context += "\(dayName): \(pending.count) task(s)\n"
                        for task in pending.prefix(3) {
                            context += "  - [\(task.priority.rawValue)] \(task.title)\n"
                        }
                    }
                }
            }
        }

        // Overdue tasks
        let overdueTasks = dataManager.tasks.filter { task in
            !task.isCompleted && calendar.compare(task.dueDate, to: today, toGranularity: .day) == .orderedAscending
        }
        if !overdueTasks.isEmpty {
            context += "\n⚠️ OVERDUE TASKS (\(overdueTasks.count)):\n"
            for task in overdueTasks.prefix(5) {
                context += "- [\(task.priority.rawValue)] \(task.title) (due: \(dateFormatter.string(from: task.dueDate)))\n"
            }
        }

        // Weekly completion rate
        var weekCompleted = 0
        var weekTotal = 0
        for dayOffset in 0..<7 {
            if let date = calendar.date(byAdding: .day, value: -dayOffset, to: today) {
                let tasks = dataManager.tasksForDate(date)
                weekTotal += tasks.count
                weekCompleted += tasks.filter { $0.isCompleted }.count
            }
        }
        let weekRate = weekTotal > 0 ? Int(Double(weekCompleted) / Double(weekTotal) * 100) : 0
        context += "\nWEEKLY COMPLETION RATE: \(weekRate)%"

        return context
    }
}

// MARK: - Tasks Progress Card

struct TasksProgressCard: View {
    let completed: Int
    let total: Int
    let isToday: Bool

    private var progress: Double {
        guard total > 0 else { return 0 }
        return Double(completed) / Double(total)
    }

    private var percentage: Int {
        Int(progress * 100)
    }

    var body: some View {
        HStack(spacing: 16) {
            // Progress Ring
            ZStack {
                Circle()
                    .stroke(Color.hf.accent.opacity(0.2), lineWidth: 6)

                Circle()
                    .trim(from: 0, to: progress)
                    .stroke(Color.hf.accent, style: StrokeStyle(lineWidth: 6, lineCap: .round))
                    .rotationEffect(.degrees(-90))
                    .animation(.easeInOut(duration: 0.3), value: progress)

                if completed == total && total > 0 {
                    Image(systemName: "checkmark")
                        .font(.system(size: 18, weight: .bold))
                        .foregroundStyle(Color.hf.accent)
                } else {
                    Text("\(completed)")
                        .font(.system(size: 18, weight: .bold, design: .rounded))
                        .foregroundStyle(Color.hf.accent)
                }
            }
            .frame(width: 52, height: 52)

            VStack(alignment: .leading, spacing: 4) {
                Text(isToday ? "Today's Progress" : "Progress")
                    .font(.system(size: 15, weight: .semibold))
                    .foregroundStyle(.primary)

                Text("\(completed) of \(total) completed")
                    .font(.system(size: 13))
                    .foregroundStyle(.secondary)
            }

            Spacer()

            Text("\(percentage)%")
                .font(.system(size: 24, weight: .bold, design: .rounded))
                .foregroundStyle(completed == total && total > 0 ? Color.hf.accent : .primary)
        }
        .padding(16)
        .background(Color.hf.cardBackground)
        .clipShape(RoundedRectangle(cornerRadius: 16))
    }
}

// MARK: - Tasks Empty State

struct TasksEmptyState: View {
    let isToday: Bool

    var body: some View {
        VStack(spacing: 12) {
            Image(systemName: isToday ? "checkmark.circle" : "calendar")
                .font(.system(size: 40))
                .foregroundStyle(.secondary)

            Text(isToday ? "No tasks for today" : "No tasks on this date")
                .font(.system(size: 17, weight: .semibold))
                .foregroundStyle(.primary)

            Text(isToday ? "Tap + to add a task" : "Tasks will appear here")
                .font(.system(size: 14))
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity)
        .padding(.vertical, 40)
    }
}

// MARK: - Tasks Date Selector

struct TasksDateSelector: View {
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
                    TasksDateCell(
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

struct TasksDateCell: View {
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

// MARK: - Task Row

struct TaskRow: View {
    let task: DailyTask
    let onToggle: () -> Void

    @State private var showCelebration = false

    var body: some View {
        HStack(spacing: 16) {
            Button {
                let wasCompleted = task.isCompleted
                onToggle()
                if !wasCompleted {
                    triggerCelebration()
                } else {
                    HapticManager.toggleOff()
                }
            } label: {
                AnimatedCheckmark(
                    isCompleted: task.isCompleted,
                    color: Color.hf.checkmarkComplete,
                    size: 24
                )
            }
            .buttonStyle(.plain)

            VStack(alignment: .leading, spacing: 4) {
                Text(task.title)
                    .font(.system(size: 17))
                    .strikethrough(task.isCompleted, color: .secondary)
                    .foregroundStyle(task.isCompleted ? .secondary : .primary)

                Text(task.priority.title)
                    .font(.system(size: 12))
                    .padding(.horizontal, 8)
                    .padding(.vertical, 2)
                    .background(task.priority.color.opacity(0.15))
                    .foregroundStyle(task.priority.color)
                    .clipShape(Capsule())
            }

            Spacer()
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

// MARK: - Add Task Sheet

struct AddTaskSheet: View {
    @EnvironmentObject var dataManager: DataManager
    @Environment(\.dismiss) var dismiss
    @State private var title = ""
    @State private var priority: TaskPriority = .medium
    @State private var dueDate: Date

    init(initialDate: Date = Date()) {
        _dueDate = State(initialValue: initialDate)
    }

    var body: some View {
        NavigationStack {
            Form {
                Section {
                    TextField("Task name", text: $title)
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

                Section {
                    DatePicker("Date", selection: $dueDate, displayedComponents: .date)
                }
            }
            .navigationTitle("New Task")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Add") {
                        let task = DailyTask(title: title, priority: priority, dueDate: dueDate)
                        dataManager.addTask(task)
                        HapticManager.completionSuccess()
                        dismiss()
                    }
                    .disabled(title.isEmpty)
                }
            }
        }
        .presentationDetents([.large])
    }
}

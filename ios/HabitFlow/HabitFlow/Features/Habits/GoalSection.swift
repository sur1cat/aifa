import SwiftUI

struct GoalSection: View {
    let goal: Goal
    let habits: [Habit]
    let dateString: String
    let isToday: Bool
    let onHabitToggle: (Habit) -> Void
    let onHabitProgressTap: (Habit) -> Void
    let onHabitProgressLongPress: (Habit) -> Void
    let onHabitTap: (Habit) -> Void
    let onGoalTap: (Goal) -> Void
    let onAddSuggestedHabit: ((SuggestedHabit, Goal) -> Void)?

    @State private var isExpanded = true
    @State private var showSuggestedHabits = false

    private var completedCount: Int {
        habits.filter { habit in
            if let target = habit.targetValue, target > 0 {
                let progress = habit.progressValues[dateString] ?? 0
                return progress >= target || habit.completedDates.contains(dateString)
            }
            return habit.completedDates.contains(dateString)
        }.count
    }

    private var progress: Double {
        guard !habits.isEmpty else { return 0 }
        return Double(completedCount) / Double(habits.count)
    }

    var body: some View {
        VStack(spacing: 0) {
            // Goal header
            HStack(spacing: 12) {
                // Tappable area for expand/collapse
                HStack(spacing: 12) {
                    // Goal icon
                    Text(goal.icon)
                        .font(.system(size: 24))

                    // Goal title and progress
                    VStack(alignment: .leading, spacing: 4) {
                        Text(goal.title)
                            .font(.system(size: 16, weight: .semibold))
                            .foregroundStyle(.primary)

                        HStack(spacing: 8) {
                            // Progress bar
                            GeometryReader { geometry in
                                ZStack(alignment: .leading) {
                                    RoundedRectangle(cornerRadius: 2)
                                        .fill(Color.hf.checkmarkIncomplete)
                                        .frame(height: 4)

                                    RoundedRectangle(cornerRadius: 2)
                                        .fill(Color.hf.accent)
                                        .frame(width: geometry.size.width * progress, height: 4)
                                }
                            }
                            .frame(height: 4)
                            .frame(maxWidth: 80)

                            // Progress text
                            Text("\(completedCount)/\(habits.count)")
                                .font(.system(size: 12))
                                .foregroundStyle(.secondary)
                        }
                    }
                }
                .contentShape(Rectangle())
                .onTapGesture {
                    withAnimation(.spring(response: 0.3, dampingFraction: 0.8)) {
                        isExpanded.toggle()
                    }
                }

                Spacer()

                // AI generate habits button
                if onAddSuggestedHabit != nil {
                    Button {
                        showSuggestedHabits = true
                    } label: {
                        Image(systemName: "sparkles")
                            .font(.system(size: 16))
                            .foregroundStyle(Color.hf.accent)
                            .padding(8)
                            .background(Color.hf.accent.opacity(0.1))
                            .clipShape(Circle())
                    }
                    .buttonStyle(.plain)
                }

                // Edit button
                Button {
                    onGoalTap(goal)
                } label: {
                    Image(systemName: "pencil.circle")
                        .font(.system(size: 18))
                        .foregroundStyle(.secondary)
                }
                .buttonStyle(.plain)

                // Expand/collapse indicator
                Image(systemName: isExpanded ? "chevron.up" : "chevron.down")
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundStyle(.secondary)
                    .onTapGesture {
                        withAnimation(.spring(response: 0.3, dampingFraction: 0.8)) {
                            isExpanded.toggle()
                        }
                    }
            }
            .padding()
            .background(Color.hf.cardBackground)
            .clipShape(RoundedRectangle(cornerRadius: 16))

            // Habits list (collapsible)
            if isExpanded {
                if habits.isEmpty {
                    // Empty state - prompt to use AI
                    HStack(spacing: 12) {
                        Image(systemName: "sparkles")
                            .font(.system(size: 20))
                            .foregroundStyle(Color.hf.accent)

                        VStack(alignment: .leading, spacing: 4) {
                            Text("No habits yet")
                                .font(.system(size: 14, weight: .medium))
                            Text("Tap ✨ to generate habits from this goal")
                                .font(.system(size: 12))
                                .foregroundStyle(.secondary)
                        }

                        Spacer()
                    }
                    .padding()
                    .background(Color.hf.accent.opacity(0.1))
                    .clipShape(RoundedRectangle(cornerRadius: 12))
                    .padding(.leading, 24)
                    .padding(.top, 8)
                    .onTapGesture {
                        showSuggestedHabits = true
                    }
                } else {
                    VStack(spacing: 8) {
                        ForEach(habits) { habit in
                            HabitRow(
                                habit: habit,
                                dateString: dateString,
                                isToday: isToday,
                                onToggle: { onHabitToggle(habit) },
                                onProgressTap: { onHabitProgressTap(habit) },
                                onProgressLongPress: { onHabitProgressLongPress(habit) }
                            )
                            .onTapGesture {
                                onHabitTap(habit)
                            }
                        }
                    }
                    .padding(.leading, 24)
                    .padding(.top, 8)
                }
                // Removed transition to avoid animation issues
            }
        }
        .sheet(isPresented: $showSuggestedHabits) {
            SuggestedHabitsSheet(goal: goal) { suggestedHabit in
                onAddSuggestedHabit?(suggestedHabit, goal)
            }
        }
    }
}

struct OtherHabitsSection: View {
    let habits: [Habit]
    let dateString: String
    let isToday: Bool
    let onHabitToggle: (Habit) -> Void
    let onHabitProgressTap: (Habit) -> Void
    let onHabitProgressLongPress: (Habit) -> Void
    let onHabitTap: (Habit) -> Void

    @State private var isExpanded = true

    var body: some View {
        VStack(spacing: 0) {
            // Section header
            Button {
                withAnimation(.spring(response: 0.3, dampingFraction: 0.8)) {
                    isExpanded.toggle()
                }
            } label: {
                HStack(spacing: 12) {
                    Image(systemName: "square.stack.3d.up")
                        .font(.system(size: 20))
                        .foregroundStyle(.secondary)

                    Text("Other Habits")
                        .font(.system(size: 16, weight: .semibold))
                        .foregroundStyle(.primary)

                    Spacer()

                    Text("\(habits.count)")
                        .font(.system(size: 14))
                        .foregroundStyle(.secondary)

                    Image(systemName: isExpanded ? "chevron.up" : "chevron.down")
                        .font(.system(size: 14, weight: .semibold))
                        .foregroundStyle(.secondary)
                }
                .padding()
                .background(Color.hf.cardBackground.opacity(0.6))
                .clipShape(RoundedRectangle(cornerRadius: 16))
            }
            .buttonStyle(.plain)

            if isExpanded {
                VStack(spacing: 8) {
                    ForEach(habits) { habit in
                        HabitRow(
                            habit: habit,
                            dateString: dateString,
                            isToday: isToday,
                            onToggle: { onHabitToggle(habit) },
                            onProgressTap: { onHabitProgressTap(habit) },
                            onProgressLongPress: { onHabitProgressLongPress(habit) }
                        )
                        .onTapGesture {
                            onHabitTap(habit)
                        }
                    }
                }
                .padding(.leading, 24)
                .padding(.top, 8)
                .transition(.opacity.combined(with: .move(edge: .top)))
            }
        }
    }
}

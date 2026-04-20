import SwiftUI

struct HabitsWatchView: View {
    @EnvironmentObject var dataStore: WatchDataStore

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 8) {
                // Header
                HStack {
                    Image(systemName: "repeat")
                        .foregroundStyle(.green)
                    Text("Habits")
                        .font(.headline)
                }
                .padding(.horizontal)

                if dataStore.habits.isEmpty {
                    Text("No habits")
                        .foregroundStyle(.secondary)
                        .padding()
                } else {
                    ForEach(dataStore.habits) { habit in
                        HabitWatchRow(habit: habit) {
                            dataStore.toggleHabit(habit)
                        }
                    }
                }
            }
        }
        .navigationTitle("Habits")
        .onAppear {
            dataStore.requestData()
        }
    }
}

struct HabitWatchRow: View {
    let habit: WatchHabit
    let onToggle: () -> Void

    var body: some View {
        Button(action: onToggle) {
            HStack {
                Image(systemName: habit.isCompleted ? "checkmark.circle.fill" : "circle")
                    .foregroundStyle(habit.isCompleted ? habitColor : .gray)

                VStack(alignment: .leading, spacing: 2) {
                    Text(habit.title)
                        .font(.system(size: 14))
                        .lineLimit(1)

                    if habit.streak > 0 {
                        Text("\(habit.streak) days")
                            .font(.system(size: 10))
                            .foregroundStyle(.secondary)
                    }
                }

                Spacer()
            }
            .padding(.horizontal)
            .padding(.vertical, 8)
        }
        .buttonStyle(.plain)
    }

    var habitColor: Color {
        switch habit.color {
        case "green": return .green
        case "blue": return .blue
        case "purple": return .purple
        case "orange": return .orange
        case "pink": return .pink
        case "red": return .red
        default: return .green
        }
    }
}

#Preview {
    HabitsWatchView()
        .environmentObject(WatchDataStore.shared)
}

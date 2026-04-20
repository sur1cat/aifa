import SwiftUI

struct TasksWatchView: View {
    @EnvironmentObject var dataStore: WatchDataStore

    var pendingTasks: [WatchTask] {
        dataStore.tasks.filter { !$0.isCompleted }
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 8) {
                // Header
                HStack {
                    Image(systemName: "checkmark.circle")
                        .foregroundStyle(.blue)
                    Text("Tasks")
                        .font(.headline)

                    Spacer()

                    Text("\(pendingTasks.count)")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
                .padding(.horizontal)

                if pendingTasks.isEmpty {
                    VStack(spacing: 8) {
                        Image(systemName: "checkmark.circle.fill")
                            .font(.title)
                            .foregroundStyle(.green)
                        Text("All done!")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                    .frame(maxWidth: .infinity)
                    .padding()
                } else {
                    ForEach(pendingTasks) { task in
                        TaskWatchRow(task: task) {
                            dataStore.toggleTask(task)
                        }
                    }
                }
            }
        }
        .navigationTitle("Tasks")
    }
}

struct TaskWatchRow: View {
    let task: WatchTask
    let onToggle: () -> Void

    var body: some View {
        Button(action: onToggle) {
            HStack {
                Circle()
                    .fill(priorityColor)
                    .frame(width: 8, height: 8)

                Text(task.title)
                    .font(.system(size: 14))
                    .lineLimit(2)
                    .strikethrough(task.isCompleted)

                Spacer()
            }
            .padding(.horizontal)
            .padding(.vertical, 8)
        }
        .buttonStyle(.plain)
    }

    var priorityColor: Color {
        switch task.priority {
        case "urgent": return .red
        case "high": return .orange
        case "medium": return .yellow
        case "low": return .green
        default: return .gray
        }
    }
}

#Preview {
    TasksWatchView()
        .environmentObject(WatchDataStore.shared)
}

import WidgetKit
import SwiftUI

struct TasksProvider: TimelineProvider {
    func placeholder(in context: Context) -> TasksEntry {
        TasksEntry(date: Date(), tasks: [
            WidgetTask(id: UUID(), title: "Review project", isCompleted: false, priority: "high"),
            WidgetTask(id: UUID(), title: "Call dentist", isCompleted: true, priority: "medium")
        ])
    }

    func getSnapshot(in context: Context, completion: @escaping (TasksEntry) -> Void) {
        let data = WidgetDataManager.shared.loadData()
        let entry = TasksEntry(date: Date(), tasks: Array(data.tasks.prefix(4)))
        completion(entry)
    }

    func getTimeline(in context: Context, completion: @escaping (Timeline<TasksEntry>) -> Void) {
        let data = WidgetDataManager.shared.loadData()
        let entry = TasksEntry(date: Date(), tasks: Array(data.tasks.prefix(4)))

        let nextUpdate = Calendar.current.date(byAdding: .minute, value: 30, to: Date())!
        let timeline = Timeline(entries: [entry], policy: .after(nextUpdate))
        completion(timeline)
    }
}

struct TasksEntry: TimelineEntry {
    let date: Date
    let tasks: [WidgetTask]
}

struct TasksWidgetEntryView: View {
    var entry: TasksProvider.Entry
    @Environment(\.widgetFamily) var family

    var body: some View {
        switch family {
        case .systemSmall:
            SmallTasksView(tasks: entry.tasks)
        case .systemMedium:
            MediumTasksView(tasks: entry.tasks)
        default:
            SmallTasksView(tasks: entry.tasks)
        }
    }
}

struct SmallTasksView: View {
    let tasks: [WidgetTask]

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Image(systemName: "checkmark.circle")
                    .font(.system(size: 14, weight: .semibold))
                Text("Tasks")
                    .font(.system(size: 14, weight: .semibold))
                Spacer()
            }
            .foregroundStyle(.secondary)

            if tasks.isEmpty {
                Spacer()
                Text("No tasks for today")
                    .font(.system(size: 13))
                    .foregroundStyle(.secondary)
                Spacer()
            } else {
                ForEach(tasks.filter { !$0.isCompleted }.prefix(3)) { task in
                    HStack(spacing: 8) {
                        Circle()
                            .fill(priorityColor(task.priority))
                            .frame(width: 8, height: 8)

                        Text(task.title)
                            .font(.system(size: 13))
                            .lineLimit(1)
                            .strikethrough(task.isCompleted)

                        Spacer()
                    }
                }
                Spacer()
            }
        }
        .containerBackground(.fill.tertiary, for: .widget)
    }

    func priorityColor(_ priority: String) -> Color {
        switch priority {
        case "urgent": return .red
        case "high": return .orange
        case "medium": return .yellow
        case "low": return .green
        default: return .gray
        }
    }
}

struct MediumTasksView: View {
    let tasks: [WidgetTask]

    var pendingTasks: [WidgetTask] {
        tasks.filter { !$0.isCompleted }
    }

    var completedCount: Int {
        tasks.filter { $0.isCompleted }.count
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Image(systemName: "checkmark.circle")
                    .font(.system(size: 14, weight: .semibold))
                Text("Tasks")
                    .font(.system(size: 14, weight: .semibold))
                Spacer()
                Text("\(completedCount) done")
                    .font(.system(size: 12))
                    .foregroundStyle(.secondary)
            }
            .foregroundStyle(.secondary)

            if pendingTasks.isEmpty {
                Spacer()
                HStack {
                    Spacer()
                    VStack(spacing: 4) {
                        Image(systemName: "checkmark.circle.fill")
                            .font(.system(size: 24))
                            .foregroundStyle(.green)
                        Text("All done!")
                            .font(.system(size: 14))
                            .foregroundStyle(.secondary)
                    }
                    Spacer()
                }
                Spacer()
            } else {
                LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: 8) {
                    ForEach(pendingTasks.prefix(4)) { task in
                        HStack(spacing: 6) {
                            Circle()
                                .fill(priorityColor(task.priority))
                                .frame(width: 6, height: 6)

                            Text(task.title)
                                .font(.system(size: 12))
                                .lineLimit(1)

                            Spacer()
                        }
                    }
                }
                Spacer()
            }
        }
        .containerBackground(.fill.tertiary, for: .widget)
    }

    func priorityColor(_ priority: String) -> Color {
        switch priority {
        case "urgent": return .red
        case "high": return .orange
        case "medium": return .yellow
        case "low": return .green
        default: return .gray
        }
    }
}

struct TasksWidget: Widget {
    let kind: String = "TasksWidget"

    var body: some WidgetConfiguration {
        StaticConfiguration(kind: kind, provider: TasksProvider()) { entry in
            TasksWidgetEntryView(entry: entry)
        }
        .configurationDisplayName("Tasks")
        .description("View your pending tasks")
        .supportedFamilies([.systemSmall, .systemMedium])
    }
}

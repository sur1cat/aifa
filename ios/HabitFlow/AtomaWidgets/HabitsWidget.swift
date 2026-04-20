import WidgetKit
import SwiftUI

struct HabitsProvider: TimelineProvider {
    func placeholder(in context: Context) -> HabitsEntry {
        HabitsEntry(date: Date(), habits: [
            WidgetHabit(id: UUID(), title: "Exercise", icon: "figure.run", color: "green", isCompleted: true, streak: 7),
            WidgetHabit(id: UUID(), title: "Read", icon: "book.fill", color: "blue", isCompleted: false, streak: 3)
        ])
    }

    func getSnapshot(in context: Context, completion: @escaping (HabitsEntry) -> Void) {
        let data = WidgetDataManager.shared.loadData()
        let entry = HabitsEntry(date: Date(), habits: Array(data.habits.prefix(4)))
        completion(entry)
    }

    func getTimeline(in context: Context, completion: @escaping (Timeline<HabitsEntry>) -> Void) {
        let data = WidgetDataManager.shared.loadData()
        let entry = HabitsEntry(date: Date(), habits: Array(data.habits.prefix(4)))

        // Update every 30 minutes
        let nextUpdate = Calendar.current.date(byAdding: .minute, value: 30, to: Date())!
        let timeline = Timeline(entries: [entry], policy: .after(nextUpdate))
        completion(timeline)
    }
}

struct HabitsEntry: TimelineEntry {
    let date: Date
    let habits: [WidgetHabit]
}

struct HabitsWidgetEntryView: View {
    var entry: HabitsProvider.Entry
    @Environment(\.widgetFamily) var family

    var body: some View {
        switch family {
        case .systemSmall:
            SmallHabitsView(habits: entry.habits)
        case .systemMedium:
            MediumHabitsView(habits: entry.habits)
        default:
            SmallHabitsView(habits: entry.habits)
        }
    }
}

struct SmallHabitsView: View {
    let habits: [WidgetHabit]

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Image(systemName: "repeat")
                    .font(.system(size: 14, weight: .semibold))
                Text("Habits")
                    .font(.system(size: 14, weight: .semibold))
                Spacer()
            }
            .foregroundStyle(.secondary)

            if habits.isEmpty {
                Spacer()
                Text("No habits yet")
                    .font(.system(size: 13))
                    .foregroundStyle(.secondary)
                Spacer()
            } else {
                ForEach(habits.prefix(3)) { habit in
                    HStack(spacing: 8) {
                        Image(systemName: habit.isCompleted ? "checkmark.circle.fill" : "circle")
                            .font(.system(size: 16))
                            .foregroundStyle(habit.isCompleted ? habitColor(habit.color) : .gray)

                        Text(habit.title)
                            .font(.system(size: 13))
                            .lineLimit(1)

                        Spacer()
                    }
                }
                Spacer()
            }
        }
        .containerBackground(.fill.tertiary, for: .widget)
    }

    func habitColor(_ color: String) -> Color {
        switch color {
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

struct MediumHabitsView: View {
    let habits: [WidgetHabit]

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Image(systemName: "repeat")
                    .font(.system(size: 14, weight: .semibold))
                Text("Habits")
                    .font(.system(size: 14, weight: .semibold))
                Spacer()
                Text("\(habits.filter { $0.isCompleted }.count)/\(habits.count)")
                    .font(.system(size: 12))
                    .foregroundStyle(.secondary)
            }
            .foregroundStyle(.secondary)

            if habits.isEmpty {
                Spacer()
                HStack {
                    Spacer()
                    Text("No habits yet")
                        .font(.system(size: 14))
                        .foregroundStyle(.secondary)
                    Spacer()
                }
                Spacer()
            } else {
                LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: 8) {
                    ForEach(habits.prefix(4)) { habit in
                        HStack(spacing: 6) {
                            Image(systemName: habit.isCompleted ? "checkmark.circle.fill" : "circle")
                                .font(.system(size: 14))
                                .foregroundStyle(habit.isCompleted ? habitColor(habit.color) : .gray)

                            Text(habit.title)
                                .font(.system(size: 12))
                                .lineLimit(1)

                            Spacer()

                            if habit.streak > 0 {
                                Text("\(habit.streak)")
                                    .font(.system(size: 10, weight: .medium))
                                    .foregroundStyle(.secondary)
                            }
                        }
                    }
                }
                Spacer()
            }
        }
        .containerBackground(.fill.tertiary, for: .widget)
    }

    func habitColor(_ color: String) -> Color {
        switch color {
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

struct HabitsWidget: Widget {
    let kind: String = "HabitsWidget"

    var body: some WidgetConfiguration {
        StaticConfiguration(kind: kind, provider: HabitsProvider()) { entry in
            HabitsWidgetEntryView(entry: entry)
        }
        .configurationDisplayName("Habits")
        .description("Track your daily habits at a glance")
        .supportedFamilies([.systemSmall, .systemMedium])
    }
}

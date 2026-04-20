import SwiftUI
import os

struct SuggestedHabitsSheet: View {
    let goal: Goal
    let onAddHabit: (SuggestedHabit) -> Void

    @Environment(\.dismiss) private var dismiss
    @State private var suggestedHabits: [SuggestedHabit] = []
    @State private var explanation: String = ""
    @State private var isLoading = false
    @State private var error: String?
    @State private var addedHabits: Set<String> = []

    // Dynamic questions from AI
    @State private var isLoadingQuestions = true
    @State private var clarifyQuestions: [ClarifyQuestion] = []
    @State private var contextHint: String = ""
    @State private var answers: [String: String] = [:]
    @State private var questionsError: String?

    private let aiService = AIService()

    var body: some View {
        NavigationStack {
            Group {
                if isLoadingQuestions {
                    loadingQuestionsView
                } else if let questionsError {
                    questionsErrorView(questionsError)
                } else if isLoading {
                    loadingView
                } else if let error {
                    errorView(error)
                } else if !suggestedHabits.isEmpty {
                    contentView
                } else {
                    contextFormView
                }
            }
            .navigationTitle("Process Habits")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Close") { dismiss() }
                }
            }
        }
        .task {
            await loadClarifyQuestions()
        }
    }

    // MARK: - Loading Questions View

    private var loadingQuestionsView: some View {
        VStack(spacing: 20) {
            ProgressView()
                .scaleEffect(1.2)

            VStack(spacing: 8) {
                Text("Analyzing your goal...")
                    .font(.system(size: 17, weight: .medium))

                Text("Preparing personalized questions")
                    .font(.system(size: 14))
                    .foregroundStyle(.secondary)
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    private func questionsErrorView(_ message: String) -> some View {
        VStack(spacing: 16) {
            Image(systemName: "exclamationmark.triangle")
                .font(.system(size: 40))
                .foregroundStyle(.orange)

            Text("Failed to load questions")
                .font(.system(size: 17, weight: .medium))

            Text(message)
                .font(.system(size: 14))
                .foregroundStyle(.secondary)
                .multilineTextAlignment(.center)

            Button("Try Again") {
                Task { await loadClarifyQuestions() }
            }
            .buttonStyle(.borderedProminent)

            Button("Skip questions") {
                questionsError = nil
                clarifyQuestions = []
                Task { await generateHabits() }
            }
            .font(.system(size: 14))
            .foregroundStyle(.secondary)
        }
        .padding()
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    // MARK: - Context Form (Dynamic Questions)

    private var contextFormView: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 20) {
                // Goal info
                goalCard

                // Explanation
                VStack(alignment: .leading, spacing: 8) {
                    HStack(spacing: 8) {
                        Image(systemName: "sparkles")
                            .foregroundStyle(Color.hf.accent)
                        Text("Help me understand")
                            .font(.system(size: 16, weight: .semibold))
                    }

                    if !contextHint.isEmpty {
                        Text(contextHint)
                            .font(.system(size: 14))
                            .foregroundStyle(.secondary)
                    } else {
                        Text("Answer a few questions to get personalized habits")
                            .font(.system(size: 14))
                            .foregroundStyle(.secondary)
                    }
                }

                // Dynamic questions from AI
                VStack(spacing: 16) {
                    ForEach(clarifyQuestions) { question in
                        VStack(alignment: .leading, spacing: 8) {
                            Text(question.question)
                                .font(.system(size: 14, weight: .medium))
                            TextField(question.placeholder, text: binding(for: question.id))
                                .textFieldStyle(.roundedBorder)
                        }
                    }
                }

                // Generate button
                Button {
                    Task { await generateHabits() }
                } label: {
                    HStack {
                        Image(systemName: "sparkles")
                        Text("Generate Habits")
                    }
                    .font(.system(size: 16, weight: .semibold))
                    .foregroundStyle(.white)
                    .frame(maxWidth: .infinity)
                    .padding()
                    .background(Color.hf.accent)
                    .clipShape(RoundedRectangle(cornerRadius: 12))
                }
                .padding(.top, 8)

                // Skip option
                Button {
                    Task { await generateHabits() }
                } label: {
                    Text("Skip, generate without context")
                        .font(.system(size: 14))
                        .foregroundStyle(.secondary)
                }
                .frame(maxWidth: .infinity)
            }
            .padding()
        }
    }

    private func binding(for questionId: String) -> Binding<String> {
        Binding(
            get: { answers[questionId] ?? "" },
            set: { answers[questionId] = $0 }
        )
    }

    // MARK: - Views

    private var loadingView: some View {
        VStack(spacing: 20) {
            ProgressView()
                .scaleEffect(1.2)

            VStack(spacing: 8) {
                Text("Generating habits...")
                    .font(.system(size: 17, weight: .medium))

                Text("Converting outcome to process habits")
                    .font(.system(size: 14))
                    .foregroundStyle(.secondary)
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    private func errorView(_ message: String) -> some View {
        VStack(spacing: 16) {
            Image(systemName: "exclamationmark.triangle")
                .font(.system(size: 40))
                .foregroundStyle(.orange)

            Text("Failed to generate habits")
                .font(.system(size: 17, weight: .medium))

            Text(message)
                .font(.system(size: 14))
                .foregroundStyle(.secondary)
                .multilineTextAlignment(.center)

            Button("Try Again") {
                Task { await generateHabits() }
            }
            .buttonStyle(.borderedProminent)
        }
        .padding()
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    private var contentView: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 20) {
                // Goal info
                goalCard

                // Explanation
                if !explanation.isEmpty {
                    explanationCard
                }

                // Suggested habits
                VStack(alignment: .leading, spacing: 12) {
                    Text("Suggested Daily Habits")
                        .font(.system(size: 15, weight: .semibold))
                        .foregroundStyle(.secondary)

                    ForEach(suggestedHabits) { habit in
                        suggestedHabitRow(habit)
                    }
                }

                // Info text
                infoCard
            }
            .padding()
        }
    }

    private var goalCard: some View {
        HStack(spacing: 12) {
            Text(goal.icon)
                .font(.system(size: 32))

            VStack(alignment: .leading, spacing: 4) {
                Text("Outcome Goal")
                    .font(.system(size: 12))
                    .foregroundStyle(.secondary)

                Text(goal.title)
                    .font(.system(size: 17, weight: .semibold))

                if let deadline = goal.deadline {
                    Text("Deadline: \(deadline.formatted(date: .abbreviated, time: .omitted))")
                        .font(.system(size: 13))
                        .foregroundStyle(.secondary)
                }
            }

            Spacer()
        }
        .padding()
        .background(Color.hf.cardBackground)
        .clipShape(RoundedRectangle(cornerRadius: 16))
    }

    private var explanationCard: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(spacing: 8) {
                Image(systemName: "lightbulb.fill")
                    .foregroundStyle(.orange)
                Text("How it works")
                    .font(.system(size: 14, weight: .semibold))
            }

            Text(explanation)
                .font(.system(size: 14))
                .foregroundStyle(.secondary)
        }
        .padding()
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(Color.orange.opacity(0.1))
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private func suggestedHabitRow(_ habit: SuggestedHabit) -> some View {
        let isAdded = addedHabits.contains(habit.id)

        return HStack(spacing: 12) {
            Text(habit.icon)
                .font(.system(size: 28))

            VStack(alignment: .leading, spacing: 4) {
                Text(habit.title)
                    .font(.system(size: 16, weight: .medium))

                Text(habit.reason)
                    .font(.system(size: 13))
                    .foregroundStyle(.secondary)
                    .lineLimit(2)

                HStack(spacing: 8) {
                    Label(habit.period.capitalized, systemImage: "repeat")
                        .font(.system(size: 11))
                        .foregroundStyle(.secondary)
                }
            }

            Spacer()

            Button {
                if !isAdded {
                    onAddHabit(habit)
                    addedHabits.insert(habit.id)
                    HapticManager.completionSuccess()
                }
            } label: {
                if isAdded {
                    Image(systemName: "checkmark.circle.fill")
                        .font(.system(size: 28))
                        .foregroundStyle(Color.hf.income)
                } else {
                    Image(systemName: "plus.circle.fill")
                        .font(.system(size: 28))
                        .foregroundStyle(Color.hf.accent)
                }
            }
            .disabled(isAdded)
        }
        .padding()
        .background(Color.hf.cardBackground)
        .clipShape(RoundedRectangle(cornerRadius: 14))
    }

    private var infoCard: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(spacing: 8) {
                Image(systemName: "info.circle.fill")
                    .foregroundStyle(Color.hf.info)
                Text("Process vs Outcome")
                    .font(.system(size: 14, weight: .semibold))
            }

            Text("Research shows process goals are 15x more effective. You control daily actions 100%, while outcomes depend on many factors.")
                .font(.system(size: 13))
                .foregroundStyle(.secondary)
        }
        .padding()
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(Color.hf.info.opacity(0.1))
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    // MARK: - Methods

    private func loadClarifyQuestions() async {
        isLoadingQuestions = true
        questionsError = nil

        AppLogger.ai.info("🎯 Loading clarify questions for goal: \(goal.title)")

        do {
            let response = try await aiService.generateGoalQuestions(title: goal.title)
            AppLogger.ai.info("✅ Got \(response.questions.count) questions")

            await MainActor.run {
                clarifyQuestions = response.questions
                contextHint = response.contextHint
                isLoadingQuestions = false
            }
        } catch {
            AppLogger.ai.error("❌ Failed to load questions: \(error)")
            await MainActor.run {
                questionsError = error.localizedDescription
                isLoadingQuestions = false
            }
        }
    }

    private func generateHabits() async {
        isLoading = true
        error = nil

        AppLogger.ai.info("🎯 Generating habits for goal: \(goal.title)")

        do {
            let deadlineStr = goal.deadline.map { DateFormatters.apiDate.string(from: $0) }
            let targetStr = goal.targetValue.map { "\($0) \(goal.unit ?? "")" }

            // Build context from user answers to dynamic questions
            var contextParts: [String] = []
            for question in clarifyQuestions {
                if let answer = answers[question.id], !answer.isEmpty {
                    contextParts.append("\(question.question): \(answer)")
                }
            }
            let contextStr = contextParts.isEmpty ? nil : contextParts.joined(separator: "\n")

            AppLogger.ai.info("📤 Calling AI service with context: \(contextStr ?? "none")")
            let response = try await aiService.generateHabitsFromGoal(
                title: goal.title,
                deadline: deadlineStr,
                targetValue: targetStr,
                context: contextStr
            )
            AppLogger.ai.info("✅ Got \(response.habits.count) habits")

            await MainActor.run {
                suggestedHabits = response.habits
                explanation = response.explanation
                isLoading = false
            }
        } catch {
            AppLogger.ai.error("❌ Failed to generate habits: \(error)")
            await MainActor.run {
                self.error = error.localizedDescription
                isLoading = false
            }
        }
    }
}

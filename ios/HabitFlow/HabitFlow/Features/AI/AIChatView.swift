import SwiftUI

struct ChatMessage: Identifiable {
    let id = UUID()
    let content: String
    let isUser: Bool
    let timestamp: Date
}

struct AIChatView: View {
    let agent: AIAgent
    let contextProvider: () -> String?

    @State private var messages: [ChatMessage] = []
    @State private var inputText = ""
    @State private var isLoading = false
    @FocusState private var isInputFocused: Bool
    @Environment(\.dismiss) private var dismiss
    @Environment(\.colorScheme) private var colorScheme

    private let aiService = AIService()

    init(agent: AIAgent, contextProvider: @escaping () -> String? = { nil }) {
        self.agent = agent
        self.contextProvider = contextProvider
    }

    var body: some View {
        VStack(spacing: 0) {
            // Header
            chatHeader

            // Messages
            ScrollViewReader { proxy in
                ScrollView {
                    LazyVStack(spacing: 16) {
                        ForEach(messages) { message in
                            MessageBubble(message: message, agent: agent)
                                .id(message.id)
                        }

                        if isLoading {
                            TypingIndicator(agent: agent)
                                .id("typing")
                        }
                    }
                    .padding(.horizontal, 16)
                    .padding(.vertical, 20)
                }
                .scrollDismissesKeyboard(.interactively)
                .onChange(of: messages.count) { _, _ in
                    scrollToBottom(proxy: proxy)
                }
                .onChange(of: isLoading) { _, loading in
                    if loading {
                        scrollToBottom(proxy: proxy, anchor: .bottom)
                    }
                }
            }

            // Input
            chatInputBar
        }
        .background(chatBackground)
        .onAppear {
            addWelcomeMessage()
        }
    }

    private func scrollToBottom(proxy: ScrollViewProxy, anchor: UnitPoint = .bottom) {
        withAnimation(.easeOut(duration: 0.2)) {
            if isLoading {
                proxy.scrollTo("typing", anchor: anchor)
            } else if let lastMessage = messages.last {
                proxy.scrollTo(lastMessage.id, anchor: anchor)
            }
        }
    }

    // MARK: - Header

    private var chatHeader: some View {
        HStack(spacing: 12) {
            Button {
                dismiss()
            } label: {
                Image(systemName: "chevron.down")
                    .font(.system(size: 16, weight: .semibold))
                    .foregroundStyle(.secondary)
                    .frame(width: 36, height: 36)
                    .background(Color.hf.cardBackground)
                    .clipShape(Circle())
            }

            Spacer()

            // Agent info
            VStack(spacing: 2) {
                HStack(spacing: 6) {
                    ZStack {
                        Circle()
                            .fill(agent.color.opacity(0.15))
                            .frame(width: 28, height: 28)
                        Image(systemName: agent.icon)
                            .font(.system(size: 12, weight: .semibold))
                            .foregroundStyle(agent.color)
                    }

                    Text(agent.displayName)
                        .font(.system(size: 16, weight: .semibold))
                }

                Text("AI Assistant")
                    .font(.system(size: 11))
                    .foregroundStyle(.secondary)
            }

            Spacer()

            // Clear chat button
            Button {
                withAnimation {
                    messages.removeAll()
                    addWelcomeMessage()
                }
            } label: {
                Image(systemName: "arrow.counterclockwise")
                    .font(.system(size: 14, weight: .medium))
                    .foregroundStyle(.secondary)
                    .frame(width: 36, height: 36)
                    .background(Color.hf.cardBackground)
                    .clipShape(Circle())
            }
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 12)
        .background(
            Color.hf.cardBackground
                .shadow(color: .black.opacity(0.05), radius: 8, y: 4)
        )
    }

    // MARK: - Input Bar

    private var chatInputBar: some View {
        HStack(alignment: .bottom, spacing: 12) {
            // Text field
            HStack(alignment: .bottom, spacing: 8) {
                TextField("Ask anything...", text: $inputText, axis: .vertical)
                    .textFieldStyle(.plain)
                    .font(.system(size: 16))
                    .lineLimit(1...6)
                    .focused($isInputFocused)
                    .padding(.vertical, 10)
                    .padding(.leading, 16)

                if !inputText.isEmpty {
                    Button {
                        inputText = ""
                    } label: {
                        Image(systemName: "xmark.circle.fill")
                            .font(.system(size: 18))
                            .foregroundStyle(.tertiary)
                    }
                    .padding(.trailing, 8)
                    .padding(.bottom, 10)
                }
            }
            .background(Color.hf.cardBackground)
            .clipShape(RoundedRectangle(cornerRadius: 24))
            .overlay(
                RoundedRectangle(cornerRadius: 24)
                    .stroke(isInputFocused ? agent.color.opacity(0.5) : Color.clear, lineWidth: 2)
            )

            // Send button
            Button {
                sendMessage()
            } label: {
                ZStack {
                    Circle()
                        .fill(canSend ? agent.color : Color.hf.cardBackground)
                        .frame(width: 44, height: 44)

                    Image(systemName: "arrow.up")
                        .font(.system(size: 18, weight: .semibold))
                        .foregroundColor(canSend ? .white : Color.gray.opacity(0.4))
                }
            }
            .disabled(!canSend)
            .animation(.easeInOut(duration: 0.15), value: canSend)
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 12)
        .background(
            Color.hf.cardBackground
                .shadow(color: .black.opacity(0.05), radius: 8, y: -4)
        )
    }

    private var canSend: Bool {
        !inputText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty && !isLoading
    }

    // MARK: - Background

    private var chatBackground: some View {
        Group {
            if colorScheme == .dark {
                Color(.systemBackground)
            } else {
                LinearGradient(
                    colors: [
                        Color(.systemGroupedBackground),
                        Color(.systemGroupedBackground).opacity(0.95)
                    ],
                    startPoint: .top,
                    endPoint: .bottom
                )
            }
        }
        .ignoresSafeArea()
    }

    // MARK: - Actions

    private func addWelcomeMessage() {
        let welcomeText = agent.welcomeMessage
        messages.append(ChatMessage(content: welcomeText, isUser: false, timestamp: Date()))
    }

    private func sendMessage() {
        let text = inputText.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !text.isEmpty else { return }

        // Add user message
        messages.append(ChatMessage(content: text, isUser: true, timestamp: Date()))
        inputText = ""
        isLoading = true

        // Get context
        let context = contextProvider()

        Task {
            do {
                let response = try await aiService.chat(agent: agent, message: text, context: context)
                await MainActor.run {
                    messages.append(ChatMessage(content: response, isUser: false, timestamp: Date()))
                    isLoading = false
                }
            } catch {
                await MainActor.run {
                    messages.append(ChatMessage(content: "Sorry, I couldn't process your request. Please try again.", isUser: false, timestamp: Date()))
                    isLoading = false
                }
            }
        }
    }
}

// MARK: - Message Bubble

struct MessageBubble: View {
    let message: ChatMessage
    let agent: AIAgent

    private var formattedContent: AttributedString {
        if let attributed = try? AttributedString(markdown: message.content, options: AttributedString.MarkdownParsingOptions(interpretedSyntax: .inlineOnlyPreservingWhitespace)) {
            return attributed
        }
        return AttributedString(message.content)
    }

    var body: some View {
        HStack(alignment: .bottom, spacing: 8) {
            if message.isUser {
                Spacer(minLength: 50)
                userBubble
            } else {
                agentBubble
                Spacer(minLength: 50)
            }
        }
    }

    private var userBubble: some View {
        VStack(alignment: .trailing, spacing: 4) {
            Text(formattedContent)
                .font(.system(size: 15))
                .foregroundStyle(.white)
                .padding(.horizontal, 16)
                .padding(.vertical, 12)
                .background(
                    LinearGradient(
                        colors: [agent.color, agent.color.opacity(0.85)],
                        startPoint: .topLeading,
                        endPoint: .bottomTrailing
                    )
                )
                .clipShape(ChatBubbleShape(isUser: true))

            Text(formatTime(message.timestamp))
                .font(.system(size: 10))
                .foregroundStyle(.tertiary)
                .padding(.trailing, 4)
        }
    }

    private var agentBubble: some View {
        HStack(alignment: .bottom, spacing: 8) {
            // Agent avatar
            ZStack {
                Circle()
                    .fill(agent.color.opacity(0.15))
                    .frame(width: 32, height: 32)

                Image(systemName: agent.icon)
                    .font(.system(size: 14, weight: .medium))
                    .foregroundStyle(agent.color)
            }

            VStack(alignment: .leading, spacing: 4) {
                Text(formattedContent)
                    .font(.system(size: 15))
                    .foregroundStyle(.primary)
                    .padding(.horizontal, 16)
                    .padding(.vertical, 12)
                    .background(Color.hf.cardBackground)
                    .clipShape(ChatBubbleShape(isUser: false))
                    .shadow(color: .black.opacity(0.04), radius: 4, y: 2)

                Text(formatTime(message.timestamp))
                    .font(.system(size: 10))
                    .foregroundStyle(.tertiary)
                    .padding(.leading, 4)
            }
        }
    }

    private func formatTime(_ date: Date) -> String {
        let formatter = DateFormatter()
        formatter.dateFormat = "HH:mm"
        return formatter.string(from: date)
    }
}

// MARK: - Chat Bubble Shape

struct ChatBubbleShape: Shape {
    let isUser: Bool

    func path(in rect: CGRect) -> Path {
        let radius: CGFloat = 18
        let tailSize: CGFloat = 6

        var path = Path()

        if isUser {
            // User bubble - tail on right
            path.move(to: CGPoint(x: rect.minX + radius, y: rect.minY))
            path.addLine(to: CGPoint(x: rect.maxX - radius, y: rect.minY))
            path.addQuadCurve(to: CGPoint(x: rect.maxX, y: rect.minY + radius),
                              control: CGPoint(x: rect.maxX, y: rect.minY))
            path.addLine(to: CGPoint(x: rect.maxX, y: rect.maxY - radius))
            // Tail
            path.addQuadCurve(to: CGPoint(x: rect.maxX + tailSize, y: rect.maxY),
                              control: CGPoint(x: rect.maxX, y: rect.maxY))
            path.addQuadCurve(to: CGPoint(x: rect.maxX - radius, y: rect.maxY),
                              control: CGPoint(x: rect.maxX, y: rect.maxY))
            path.addLine(to: CGPoint(x: rect.minX + radius, y: rect.maxY))
            path.addQuadCurve(to: CGPoint(x: rect.minX, y: rect.maxY - radius),
                              control: CGPoint(x: rect.minX, y: rect.maxY))
            path.addLine(to: CGPoint(x: rect.minX, y: rect.minY + radius))
            path.addQuadCurve(to: CGPoint(x: rect.minX + radius, y: rect.minY),
                              control: CGPoint(x: rect.minX, y: rect.minY))
        } else {
            // Agent bubble - tail on left
            path.move(to: CGPoint(x: rect.minX + radius, y: rect.minY))
            path.addLine(to: CGPoint(x: rect.maxX - radius, y: rect.minY))
            path.addQuadCurve(to: CGPoint(x: rect.maxX, y: rect.minY + radius),
                              control: CGPoint(x: rect.maxX, y: rect.minY))
            path.addLine(to: CGPoint(x: rect.maxX, y: rect.maxY - radius))
            path.addQuadCurve(to: CGPoint(x: rect.maxX - radius, y: rect.maxY),
                              control: CGPoint(x: rect.maxX, y: rect.maxY))
            path.addLine(to: CGPoint(x: rect.minX + radius, y: rect.maxY))
            // Tail
            path.addQuadCurve(to: CGPoint(x: rect.minX - tailSize, y: rect.maxY),
                              control: CGPoint(x: rect.minX, y: rect.maxY))
            path.addQuadCurve(to: CGPoint(x: rect.minX, y: rect.maxY - radius),
                              control: CGPoint(x: rect.minX, y: rect.maxY))
            path.addLine(to: CGPoint(x: rect.minX, y: rect.minY + radius))
            path.addQuadCurve(to: CGPoint(x: rect.minX + radius, y: rect.minY),
                              control: CGPoint(x: rect.minX, y: rect.minY))
        }

        return path
    }
}

// MARK: - Typing Indicator

struct TypingIndicator: View {
    let agent: AIAgent
    @State private var animatingDots = [false, false, false]

    var body: some View {
        HStack(alignment: .bottom, spacing: 8) {
            // Agent avatar
            ZStack {
                Circle()
                    .fill(agent.color.opacity(0.15))
                    .frame(width: 32, height: 32)

                Image(systemName: agent.icon)
                    .font(.system(size: 14, weight: .medium))
                    .foregroundStyle(agent.color)
            }

            // Typing dots
            HStack(spacing: 5) {
                ForEach(0..<3, id: \.self) { index in
                    Circle()
                        .fill(agent.color.opacity(animatingDots[index] ? 1 : 0.3))
                        .frame(width: 8, height: 8)
                        .offset(y: animatingDots[index] ? -4 : 0)
                }
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 14)
            .background(Color.hf.cardBackground)
            .clipShape(ChatBubbleShape(isUser: false))
            .shadow(color: .black.opacity(0.04), radius: 4, y: 2)

            Spacer(minLength: 50)
        }
        .onAppear {
            startAnimation()
        }
    }

    private func startAnimation() {
        for i in 0..<3 {
            withAnimation(
                .easeInOut(duration: 0.4)
                .repeatForever(autoreverses: true)
                .delay(Double(i) * 0.15)
            ) {
                animatingDots[i] = true
            }
        }
    }
}

// MARK: - AIAgent Extension

extension AIAgent {
    var color: Color {
        switch self {
        case .habitCoach: return Color.hf.accent
        case .taskAssistant: return Color.hf.info
        case .financeAdvisor: return Color.hf.income
        case .lifeCoach: return Color.hf.warning
        }
    }

    var welcomeMessage: String {
        switch self {
        case .habitCoach:
            return "Hi! I'm your Habit Coach. I can help you build better habits, stay motivated, and track your progress. What would you like to work on?"
        case .taskAssistant:
            return "Hello! I'm your Task Assistant. I can help you prioritize tasks, break down projects, and stay productive. What's on your mind?"
        case .financeAdvisor:
            return "Hi there! I'm your Finance Advisor. I can analyze your spending, suggest ways to save, and help you reach your financial goals. How can I help?"
        case .lifeCoach:
            return "Welcome! I'm your Life Coach. I'm here to help you balance habits, tasks, and finances for a more fulfilling life. What's on your mind?"
        }
    }
}

#Preview {
    AIChatView(agent: .habitCoach)
}

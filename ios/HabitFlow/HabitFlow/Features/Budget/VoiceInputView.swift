import SwiftUI

struct VoiceInputView: View {
    @EnvironmentObject var dataManager: DataManager
    @StateObject private var voiceService = VoiceInputService()
    @Environment(\.dismiss) private var dismiss
    @Environment(\.colorScheme) private var colorScheme

    @State private var parsedTransaction: ParsedTransaction?
    @State private var showConfirmation = false

    var body: some View {
        NavigationStack {
            VStack(spacing: 32) {
                Spacer()

                // Microphone button
                Button {
                    if voiceService.isRecording {
                        voiceService.stopRecording()
                        parseResult()
                    } else {
                        Task {
                            let authorized = await voiceService.requestPermission()
                            if authorized {
                                voiceService.startRecording()
                            }
                        }
                    }
                } label: {
                    ZStack {
                        Circle()
                            .fill(voiceService.isRecording ? Color.hf.expense : Color.hf.accent)
                            .frame(width: 120, height: 120)
                            .shadow(color: voiceService.isRecording ? Color.hf.expense.opacity(0.4) : Color.hf.accent.opacity(0.4), radius: 20)

                        Image(systemName: voiceService.isRecording ? "stop.fill" : "mic.fill")
                            .font(.system(size: 44))
                            .foregroundStyle(.white)
                    }
                }
                .scaleEffect(voiceService.isRecording ? 1.1 : 1.0)
                .animation(.easeInOut(duration: 0.5).repeatForever(autoreverses: true), value: voiceService.isRecording)

                // Instructions or transcribed text
                VStack(spacing: 12) {
                    if voiceService.isRecording {
                        Text("Listening...")
                            .font(.headline)
                            .foregroundStyle(.secondary)

                        if !voiceService.transcribedText.isEmpty {
                            Text(voiceService.transcribedText)
                                .font(.body)
                                .multilineTextAlignment(.center)
                                .padding(.horizontal)
                        }
                    } else if let parsed = parsedTransaction {
                        // Show parsed result
                        VStack(spacing: 16) {
                            Text("Recognized:")
                                .font(.subheadline)
                                .foregroundStyle(.secondary)

                            VStack(spacing: 8) {
                                Text(parsed.title)
                                    .font(.headline)

                                Text("\(parsed.type == .income ? "+" : "-")\(dataManager.profile.currency.symbol)\(String(format: "%.2f", parsed.amount))")
                                    .font(.system(size: 32, weight: .bold, design: .rounded))
                                    .foregroundStyle(parsed.type == .income ? Color.hf.income : Color.hf.expense)

                                Text(parsed.category.capitalized)
                                    .font(.caption)
                                    .padding(.horizontal, 12)
                                    .padding(.vertical, 4)
                                    .background(Color.hf.cardBackground)
                                    .clipShape(Capsule())
                            }
                            .padding()
                            .background(Color.hf.cardBackground)
                            .clipShape(RoundedRectangle(cornerRadius: 16))
                        }
                    } else {
                        Text("Tap to speak")
                            .font(.headline)
                            .foregroundStyle(.secondary)

                        Text("Example: \"Spent 500 on coffee\"")
                            .font(.subheadline)
                            .foregroundStyle(.tertiary)
                    }
                }
                .frame(minHeight: 150)

                Spacer()

                // Action buttons
                if parsedTransaction != nil && !voiceService.isRecording {
                    HStack(spacing: 16) {
                        Button {
                            parsedTransaction = nil
                            voiceService.transcribedText = ""
                        } label: {
                            Text("Cancel")
                                .font(.headline)
                                .frame(maxWidth: .infinity)
                                .padding()
                                .background(Color.hf.cardBackground)
                                .clipShape(RoundedRectangle(cornerRadius: 14))
                        }

                        Button {
                            saveTransaction()
                        } label: {
                            Text("Add")
                                .font(.headline)
                                .foregroundStyle(.white)
                                .frame(maxWidth: .infinity)
                                .padding()
                                .background(Color.hf.accent)
                                .clipShape(RoundedRectangle(cornerRadius: 14))
                        }
                    }
                    .padding(.horizontal)
                }

                if let error = voiceService.error {
                    Text(error)
                        .font(.caption)
                        .foregroundStyle(Color.hf.expense)
                        .padding()
                }
            }
            .padding()
            .background(AppTheme.appBackground(for: colorScheme))
            .navigationTitle("Voice Input")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Close") { dismiss() }
                }
            }
        }
    }

    private func parseResult() {
        guard !voiceService.transcribedText.isEmpty else { return }
        parsedTransaction = voiceService.parseTransaction(from: voiceService.transcribedText)
    }

    private func saveTransaction() {
        guard let parsed = parsedTransaction else { return }

        let transaction = Transaction(
            title: parsed.title,
            amount: parsed.amount,
            type: parsed.type,
            category: parsed.category
        )

        dataManager.addTransaction(transaction)
        dismiss()
    }
}

#Preview {
    VoiceInputView()
        .environmentObject(DataManager.shared)
}

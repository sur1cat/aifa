import Foundation
import Speech
import AVFoundation
import Combine

@MainActor
class VoiceInputService: ObservableObject {
    @Published var isRecording = false
    @Published var transcribedText = ""
    @Published var isAuthorized = false
    @Published var error: String?

    private var audioEngine: AVAudioEngine?
    private var recognitionRequest: SFSpeechAudioBufferRecognitionRequest?
    private var recognitionTask: SFSpeechRecognitionTask?
    private let speechRecognizer: SFSpeechRecognizer?

    init() {
        // Support Russian and English
        speechRecognizer = SFSpeechRecognizer(locale: Locale(identifier: "ru-RU"))
            ?? SFSpeechRecognizer(locale: Locale(identifier: "en-US"))
    }

    func requestPermission() async -> Bool {
        let speechStatus = await withCheckedContinuation { continuation in
            SFSpeechRecognizer.requestAuthorization { status in
                continuation.resume(returning: status)
            }
        }

        guard speechStatus == .authorized else {
            await MainActor.run { isAuthorized = false }
            return false
        }

        let audioStatus = await AVAudioApplication.requestRecordPermission()
        await MainActor.run { isAuthorized = audioStatus }
        return audioStatus
    }

    func startRecording() {
        guard let speechRecognizer = speechRecognizer, speechRecognizer.isAvailable else {
            error = "Speech recognition not available"
            return
        }

        do {
            audioEngine = AVAudioEngine()
            guard let audioEngine = audioEngine else { return }

            let audioSession = AVAudioSession.sharedInstance()
            try audioSession.setCategory(.record, mode: .measurement, options: .duckOthers)
            try audioSession.setActive(true, options: .notifyOthersOnDeactivation)

            recognitionRequest = SFSpeechAudioBufferRecognitionRequest()
            guard let recognitionRequest = recognitionRequest else { return }

            recognitionRequest.shouldReportPartialResults = true

            let inputNode = audioEngine.inputNode
            let recordingFormat = inputNode.outputFormat(forBus: 0)

            inputNode.installTap(onBus: 0, bufferSize: 1024, format: recordingFormat) { buffer, _ in
                self.recognitionRequest?.append(buffer)
            }

            audioEngine.prepare()
            try audioEngine.start()

            isRecording = true
            transcribedText = ""

            recognitionTask = speechRecognizer.recognitionTask(with: recognitionRequest) { [weak self] result, error in
                Task { @MainActor in
                    if let result = result {
                        self?.transcribedText = result.bestTranscription.formattedString
                    }

                    if error != nil || result?.isFinal == true {
                        self?.stopRecording()
                    }
                }
            }
        } catch {
            self.error = error.localizedDescription
            stopRecording()
        }
    }

    func stopRecording() {
        audioEngine?.stop()
        audioEngine?.inputNode.removeTap(onBus: 0)
        recognitionRequest?.endAudio()
        recognitionTask?.cancel()

        audioEngine = nil
        recognitionRequest = nil
        recognitionTask = nil
        isRecording = false
    }

    // Parse voice input to transaction
    // Examples: "потратил 500 рублей на кофе", "spent 20 dollars on lunch"
    func parseTransaction(from text: String) -> ParsedTransaction? {
        let lowercased = text.lowercased()

        // Extract amount - handle spaces in numbers like "10 000" or "1 000 000"
        let amountPattern = #"(\d[\d\s]*\d|\d+)(?:[.,](\d+))?"#
        guard let amountMatch = lowercased.range(of: amountPattern, options: .regularExpression) else {
            return nil
        }

        // Remove spaces from number and convert to Double
        let amountString = String(lowercased[amountMatch])
            .replacingOccurrences(of: " ", with: "")
            .replacingOccurrences(of: ",", with: ".")

        guard let amount = Double(amountString) else {
            return nil
        }

        // Determine type (income or expense)
        let incomeKeywords = ["получил", "заработал", "пришло", "earned", "received", "got paid", "income"]

        var transactionType: TransactionType = .expense
        for keyword in incomeKeywords {
            if lowercased.contains(keyword) {
                transactionType = .income
                break
            }
        }

        // Extract description (words after "на", "on", "for")
        var title = "Voice transaction"
        let descPatterns = ["на (.+)", "on (.+)", "for (.+)"]
        for pattern in descPatterns {
            if let range = lowercased.range(of: pattern, options: .regularExpression) {
                let match = String(lowercased[range])
                let cleaned = match
                    .replacingOccurrences(of: "на ", with: "")
                    .replacingOccurrences(of: "on ", with: "")
                    .replacingOccurrences(of: "for ", with: "")
                    .trimmingCharacters(in: .whitespaces)
                if !cleaned.isEmpty {
                    title = cleaned.capitalized
                }
                break
            }
        }

        // Detect category from keywords
        let category = detectCategory(from: lowercased)

        return ParsedTransaction(
            title: title,
            amount: amount,
            type: transactionType,
            category: category
        )
    }

    private func detectCategory(from text: String) -> String {
        let categories: [String: [String]] = [
            "food": ["кофе", "еда", "обед", "ужин", "завтрак", "ресторан", "кафе", "coffee", "food", "lunch", "dinner", "breakfast", "restaurant"],
            "transport": ["такси", "метро", "автобус", "бензин", "taxi", "metro", "bus", "gas", "uber"],
            "shopping": ["магазин", "одежда", "покупки", "shop", "clothes", "shopping", "store"],
            "entertainment": ["кино", "игры", "развлечения", "movie", "games", "entertainment"],
            "health": ["аптека", "врач", "лекарства", "pharmacy", "doctor", "medicine"],
            "bills": ["счет", "оплата", "коммуналка", "bill", "payment", "utilities"]
        ]

        for (category, keywords) in categories {
            for keyword in keywords {
                if text.contains(keyword) {
                    return category
                }
            }
        }

        return "other"
    }
}

struct ParsedTransaction {
    let title: String
    let amount: Double
    let type: TransactionType
    let category: String
}

import Foundation
import Vision
import UIKit
import Combine

@MainActor
class ReceiptScannerService: ObservableObject {
    @Published var isProcessing = false
    @Published var recognizedText = ""
    @Published var parsedReceipt: ParsedReceipt?
    @Published var error: String?

    func processImage(_ image: UIImage) async {
        guard let cgImage = image.cgImage else {
            error = "Invalid image"
            return
        }

        isProcessing = true
        error = nil

        do {
            let text = try await recognizeText(from: cgImage)
            recognizedText = text
            parsedReceipt = parseReceipt(from: text)
        } catch {
            self.error = error.localizedDescription
        }

        isProcessing = false
    }

    private func recognizeText(from cgImage: CGImage) async throws -> String {
        try await withCheckedThrowingContinuation { continuation in
            let request = VNRecognizeTextRequest { request, error in
                if let error = error {
                    continuation.resume(throwing: error)
                    return
                }

                guard let observations = request.results as? [VNRecognizedTextObservation] else {
                    continuation.resume(returning: "")
                    return
                }

                let text = observations.compactMap { observation in
                    observation.topCandidates(1).first?.string
                }.joined(separator: "\n")

                continuation.resume(returning: text)
            }

            // Support multiple languages
            request.recognitionLanguages = ["ru-RU", "en-US"]
            request.recognitionLevel = .accurate
            request.usesLanguageCorrection = true

            let handler = VNImageRequestHandler(cgImage: cgImage, options: [:])

            do {
                try handler.perform([request])
            } catch {
                continuation.resume(throwing: error)
            }
        }
    }

    func parseReceipt(from text: String) -> ParsedReceipt? {
        let lines = text.lowercased().components(separatedBy: .newlines)

        // Extract store name (usually first non-empty line)
        var storeName = "Receipt"
        for line in lines {
            let trimmed = line.trimmingCharacters(in: .whitespaces)
            if !trimmed.isEmpty && trimmed.count > 2 {
                storeName = trimmed.capitalized
                break
            }
        }

        // Find total amount - look for keywords and extract the largest amount
        let totalKeywords = ["итого", "total", "всего", "к оплате", "sum", "amount", "сумма", "итог"]
        var amounts: [Double] = []

        // Patterns for currency amounts
        let amountPatterns = [
            #"(\d{1,3}(?:[.,\s]\d{3})*(?:[.,]\d{2})?)"#,  // 1,234.56 or 1 234,56
            #"(\d+(?:[.,]\d{1,2})?)"#  // Simple number with optional decimals
        ]

        for line in lines {
            // Check if line contains total keyword
            let isTotalLine = totalKeywords.contains { line.contains($0) }

            for pattern in amountPatterns {
                if let regex = try? NSRegularExpression(pattern: pattern),
                   let match = regex.firstMatch(in: line, range: NSRange(line.startIndex..., in: line)),
                   let matchRange = Range(match.range(at: 1), in: line) {
                    let amountStr = String(line[matchRange])
                        .replacingOccurrences(of: " ", with: "")
                        .replacingOccurrences(of: ",", with: ".")

                    if let amount = Double(amountStr), amount > 0 {
                        if isTotalLine {
                            // Prioritize amounts on total lines
                            amounts.insert(amount, at: 0)
                        } else {
                            amounts.append(amount)
                        }
                    }
                }
            }
        }

        // Get the most likely total (first if we found one on a "total" line, otherwise largest)
        guard let total = amounts.first ?? amounts.max() else {
            return nil
        }

        // Detect category from store name
        let category = detectCategory(from: storeName + " " + text)

        return ParsedReceipt(
            storeName: storeName,
            amount: total,
            category: category,
            rawText: text
        )
    }

    private func detectCategory(from text: String) -> String {
        let lowercased = text.lowercased()

        let categories: [String: [String]] = [
            "food": ["продукты", "магнит", "пятерочка", "перекресток", "дикси", "ашан", "metro", "grocery", "supermarket", "food", "market", "еда", "супермаркет"],
            "restaurant": ["ресторан", "кафе", "кофейня", "бар", "restaurant", "cafe", "coffee", "starbucks", "mcdonalds", "kfc", "burger"],
            "transport": ["такси", "яндекс", "uber", "метро", "транспорт", "бензин", "азс", "gas", "taxi", "transport", "parking"],
            "shopping": ["одежда", "обувь", "zara", "h&m", "uniqlo", "clothes", "shopping", "mall", "fashion"],
            "health": ["аптека", "pharmacy", "медицина", "clinic", "hospital", "doctor", "здоровье"],
            "entertainment": ["кино", "театр", "cinema", "movie", "entertainment", "game", "развлечения"]
        ]

        for (category, keywords) in categories {
            for keyword in keywords {
                if lowercased.contains(keyword) {
                    return category
                }
            }
        }

        return "other"
    }
}

struct ParsedReceipt {
    let storeName: String
    let amount: Double
    let category: String
    let rawText: String
}

import SwiftUI
import PhotosUI

struct ReceiptScannerView: View {
    @EnvironmentObject var dataManager: DataManager
    @StateObject private var scannerService = ReceiptScannerService()
    @Environment(\.dismiss) private var dismiss
    @Environment(\.colorScheme) private var colorScheme

    @State private var selectedItem: PhotosPickerItem?
    @State private var showCamera = false
    @State private var capturedImage: UIImage?

    var body: some View {
        NavigationStack {
            VStack(spacing: 24) {
                if scannerService.isProcessing {
                    // Processing state
                    VStack(spacing: 16) {
                        ProgressView()
                            .scaleEffect(1.5)
                        Text("Scanning receipt...")
                            .font(.headline)
                            .foregroundStyle(.secondary)
                    }
                    .frame(maxHeight: .infinity)
                } else if let parsed = scannerService.parsedReceipt {
                    // Show parsed result
                    VStack(spacing: 20) {
                        Image(systemName: "doc.text.viewfinder")
                            .font(.system(size: 48))
                            .foregroundStyle(Color.hf.accent)

                        VStack(spacing: 12) {
                            Text(parsed.storeName)
                                .font(.title2.bold())

                            Text("-\(dataManager.profile.currency.symbol)\(String(format: "%.2f", parsed.amount))")
                                .font(.system(size: 36, weight: .bold, design: .rounded))
                                .foregroundStyle(Color.hf.expense)

                            Text(parsed.category.capitalized)
                                .font(.caption)
                                .padding(.horizontal, 12)
                                .padding(.vertical, 6)
                                .background(Color.hf.cardBackground)
                                .clipShape(Capsule())
                        }
                        .padding()
                        .frame(maxWidth: .infinity)
                        .background(Color.hf.cardBackground)
                        .clipShape(RoundedRectangle(cornerRadius: 16))

                        Spacer()

                        // Action buttons
                        HStack(spacing: 16) {
                            Button {
                                scannerService.parsedReceipt = nil
                                scannerService.recognizedText = ""
                            } label: {
                                Text("Scan Again")
                                    .font(.headline)
                                    .frame(maxWidth: .infinity)
                                    .padding()
                                    .background(Color.hf.cardBackground)
                                    .clipShape(RoundedRectangle(cornerRadius: 14))
                            }

                            Button {
                                saveTransaction()
                            } label: {
                                Text("Add Expense")
                                    .font(.headline)
                                    .foregroundStyle(.white)
                                    .frame(maxWidth: .infinity)
                                    .padding()
                                    .background(Color.hf.accent)
                                    .clipShape(RoundedRectangle(cornerRadius: 14))
                            }
                        }
                    }
                    .padding()
                } else {
                    // Initial state - choose source
                    VStack(spacing: 32) {
                        Spacer()

                        Image(systemName: "doc.text.viewfinder")
                            .font(.system(size: 80))
                            .foregroundStyle(Color.hf.accent.opacity(0.8))

                        Text("Scan a receipt to add expense")
                            .font(.headline)
                            .foregroundStyle(.secondary)

                        VStack(spacing: 16) {
                            // Camera button
                            Button {
                                showCamera = true
                            } label: {
                                Label("Take Photo", systemImage: "camera.fill")
                                    .font(.headline)
                                    .frame(maxWidth: .infinity)
                                    .padding()
                                    .background(Color.hf.accent)
                                    .foregroundStyle(.white)
                                    .clipShape(RoundedRectangle(cornerRadius: 14))
                            }

                            // Photo picker
                            PhotosPicker(selection: $selectedItem, matching: .images) {
                                Label("Choose from Library", systemImage: "photo.on.rectangle")
                                    .font(.headline)
                                    .frame(maxWidth: .infinity)
                                    .padding()
                                    .background(Color.hf.cardBackground)
                                    .clipShape(RoundedRectangle(cornerRadius: 14))
                            }
                        }
                        .padding(.horizontal)

                        Spacer()
                    }
                }

                if let error = scannerService.error {
                    Text(error)
                        .font(.caption)
                        .foregroundStyle(Color.hf.expense)
                        .padding()
                }
            }
            .padding()
            .background(AppTheme.appBackground(for: colorScheme))
            .navigationTitle("Receipt Scanner")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Close") { dismiss() }
                }
            }
            .onChange(of: selectedItem) { _, newItem in
                Task {
                    if let data = try? await newItem?.loadTransferable(type: Data.self),
                       let image = UIImage(data: data) {
                        await scannerService.processImage(image)
                    }
                }
            }
            .fullScreenCover(isPresented: $showCamera) {
                CameraView { image in
                    Task {
                        await scannerService.processImage(image)
                    }
                }
            }
        }
    }

    private func saveTransaction() {
        guard let parsed = scannerService.parsedReceipt else { return }

        let transaction = Transaction(
            title: parsed.storeName,
            amount: parsed.amount,
            type: .expense,
            category: parsed.category
        )

        dataManager.addTransaction(transaction)
        dismiss()
    }
}

// Camera view using UIImagePickerController
struct CameraView: UIViewControllerRepresentable {
    @Environment(\.dismiss) private var dismiss
    let onCapture: (UIImage) -> Void

    func makeUIViewController(context: Context) -> UIImagePickerController {
        let picker = UIImagePickerController()
        picker.sourceType = .camera
        picker.delegate = context.coordinator
        return picker
    }

    func updateUIViewController(_ uiViewController: UIImagePickerController, context: Context) {}

    func makeCoordinator() -> Coordinator {
        Coordinator(self)
    }

    class Coordinator: NSObject, UIImagePickerControllerDelegate, UINavigationControllerDelegate {
        let parent: CameraView

        init(_ parent: CameraView) {
            self.parent = parent
        }

        func imagePickerController(_ picker: UIImagePickerController, didFinishPickingMediaWithInfo info: [UIImagePickerController.InfoKey: Any]) {
            if let image = info[.originalImage] as? UIImage {
                parent.onCapture(image)
            }
            parent.dismiss()
        }

        func imagePickerControllerDidCancel(_ picker: UIImagePickerController) {
            parent.dismiss()
        }
    }
}

#Preview {
    ReceiptScannerView()
        .environmentObject(DataManager.shared)
}

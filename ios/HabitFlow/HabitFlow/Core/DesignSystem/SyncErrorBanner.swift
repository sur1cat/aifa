import SwiftUI

struct SyncErrorBanner: View {
    @EnvironmentObject var dataManager: DataManager
    @State private var isVisible = false

    var body: some View {
        VStack {
            if isVisible, dataManager.syncError != nil {
                HStack(spacing: 12) {
                    Image(systemName: "exclamationmark.icloud")
                        .font(.system(size: 16))
                        .foregroundStyle(.white)

                    Text("Sync failed")
                        .font(.system(size: 14, weight: .medium))
                        .foregroundStyle(.white)

                    Spacer()

                    Button {
                        withAnimation {
                            dataManager.syncError = nil
                            isVisible = false
                        }
                    } label: {
                        Image(systemName: "xmark")
                            .font(.system(size: 12, weight: .semibold))
                            .foregroundStyle(.white.opacity(0.8))
                    }
                }
                .padding(.horizontal, 16)
                .padding(.vertical, 12)
                .background(Color.hf.expense.opacity(0.9))
                .clipShape(RoundedRectangle(cornerRadius: 12))
                .padding(.horizontal)
                .transition(.move(edge: .top).combined(with: .opacity))
            }

            Spacer()
        }
        .onChange(of: dataManager.syncError) { _, newValue in
            withAnimation(.spring(response: 0.4, dampingFraction: 0.8)) {
                isVisible = newValue != nil
            }

            // Auto-dismiss after 5 seconds
            if newValue != nil {
                DispatchQueue.main.asyncAfter(deadline: .now() + 5) {
                    withAnimation {
                        isVisible = false
                        dataManager.syncError = nil
                    }
                }
            }
        }
    }
}

// ViewModifier for easy application
struct SyncErrorOverlay: ViewModifier {
    func body(content: Content) -> some View {
        content
            .overlay(alignment: .top) {
                SyncErrorBanner()
                    .padding(.top, 8)
            }
    }
}

extension View {
    func withSyncErrorBanner() -> some View {
        modifier(SyncErrorOverlay())
    }
}

import SwiftUI

struct CollapsibleSection<Content: View>: View {
    let title: String
    let icon: String
    let iconColor: Color
    @Binding var isExpanded: Bool
    let actionIcon: String?
    let onAction: (() -> Void)?
    let content: () -> Content

    init(
        title: String,
        icon: String,
        iconColor: Color = .secondary,
        isExpanded: Binding<Bool>,
        actionIcon: String? = nil,
        onAction: (() -> Void)? = nil,
        @ViewBuilder content: @escaping () -> Content
    ) {
        self.title = title
        self.icon = icon
        self.iconColor = iconColor
        self._isExpanded = isExpanded
        self.actionIcon = actionIcon
        self.onAction = onAction
        self.content = content
    }

    var body: some View {
        VStack(spacing: 0) {
            // Header
            HStack(spacing: 10) {
                Button {
                    withAnimation(.easeInOut(duration: 0.2)) {
                        isExpanded.toggle()
                    }
                } label: {
                    HStack(spacing: 10) {
                        Image(systemName: icon)
                            .font(.system(size: 14))
                            .foregroundStyle(iconColor)
                            .frame(width: 20)

                        Text(title)
                            .font(.system(size: 15, weight: .semibold))
                            .foregroundStyle(.primary)

                        Spacer()
                    }
                }
                .buttonStyle(.plain)

                // Action button (e.g., add)
                if let actionIcon = actionIcon, let onAction = onAction {
                    Button(action: onAction) {
                        Image(systemName: actionIcon)
                            .font(.system(size: 18))
                            .foregroundStyle(Color.hf.accent)
                    }
                    .buttonStyle(.plain)
                }

                // Chevron
                Button {
                    withAnimation(.easeInOut(duration: 0.2)) {
                        isExpanded.toggle()
                    }
                } label: {
                    Image(systemName: isExpanded ? "chevron.up" : "chevron.down")
                        .font(.system(size: 12, weight: .semibold))
                        .foregroundStyle(.secondary)
                }
                .buttonStyle(.plain)
            }
            .padding(.vertical, 14)
            .padding(.horizontal, 16)
            .background(Color.hf.cardBackground)

            // Content
            if isExpanded {
                content()
                    .transition(.opacity.combined(with: .move(edge: .top)))
            }
        }
        .clipShape(RoundedRectangle(cornerRadius: 14))
    }
}

// MARK: - Simple Collapsible Header (for inline use)

struct CollapsibleHeader: View {
    let title: String
    let icon: String
    let iconColor: Color
    let isExpanded: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack(spacing: 10) {
                Image(systemName: icon)
                    .font(.system(size: 14))
                    .foregroundStyle(iconColor)
                    .frame(width: 20)

                Text(title)
                    .font(.system(size: 15, weight: .semibold))
                    .foregroundStyle(.primary)

                Spacer()

                Image(systemName: isExpanded ? "chevron.up" : "chevron.down")
                    .font(.system(size: 12, weight: .semibold))
                    .foregroundStyle(.secondary)
            }
            .padding(.vertical, 12)
        }
        .buttonStyle(.plain)
    }
}

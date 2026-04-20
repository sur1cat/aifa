import SwiftUI
import AuthenticationServices

struct LoginView: View {
    @EnvironmentObject var authManager: AuthManager
    @Environment(\.colorScheme) var colorScheme
    @State private var showError = false
    @State private var showTerms = false
    @State private var showPrivacy = false

    var body: some View {
        VStack(spacing: 0) {
            Spacer()

            // Logo - minimal
            VStack(spacing: 16) {
                Image(systemName: "leaf.fill")
                    .font(.system(size: 56))
                    .foregroundStyle(Color.hf.accent.opacity(0.9))

                Text("Atoma")
                    .font(.system(size: 32, weight: .light, design: .default))
                    .tracking(1)

                Text("habits, tasks, money — in one flow")
                    .font(.system(size: 15, weight: .light))
                    .foregroundStyle(.secondary)
            }

            Spacer()
            Spacer()

            // Auth Buttons - minimal
            VStack(spacing: 12) {
                // Google
                Button {
                    Task { await authManager.signInWithGoogle() }
                } label: {
                    HStack(spacing: 12) {
                        GoogleLogo()
                            .frame(width: 20, height: 20)
                        Text("Continue with Google")
                            .font(.system(size: 16, weight: .regular))
                    }
                    .frame(maxWidth: .infinity)
                    .frame(height: 52)
                    .background(Color.hf.cardBackground)
                    .foregroundStyle(Color.hf.textPrimary)
                    .clipShape(RoundedRectangle(cornerRadius: 14))
                }
                .disabled(authManager.isLoading)

                // Apple
                SignInWithAppleButton(.continue) { request in
                    request.requestedScopes = [.fullName, .email]
                } onCompletion: { result in
                    if case .success(let auth) = result {
                        Task { await authManager.handleAppleSignIn(auth) }
                    }
                }
                .signInWithAppleButtonStyle(colorScheme == .dark ? .white : .black)
                .frame(height: 52)
                .clipShape(RoundedRectangle(cornerRadius: 14))
            }
            .padding(.horizontal, 32)

            if authManager.isLoading {
                ProgressView()
                    .padding(.top, 20)
            }

            Spacer()
                .frame(height: 60)

            HStack(spacing: 4) {
                Text("By continuing, you agree to our")
                    .font(.system(size: 12))
                    .foregroundStyle(.tertiary)

                Button("Terms") {
                    showTerms = true
                }
                .font(.system(size: 12, weight: .medium))
                .foregroundStyle(.secondary)

                Text("&")
                    .font(.system(size: 12))
                    .foregroundStyle(.tertiary)

                Button("Privacy") {
                    showPrivacy = true
                }
                .font(.system(size: 12, weight: .medium))
                .foregroundStyle(.secondary)
            }
            .padding(.bottom, 16)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(AppTheme.loginBackground(for: colorScheme))
        .alert("Error", isPresented: $showError) {
            Button("OK") {}
        } message: {
            Text(authManager.error ?? "")
        }
        .onChange(of: authManager.error) { _, new in
            showError = new != nil
        }
        .sheet(isPresented: $showTerms) {
            TermsOfServiceView()
        }
        .sheet(isPresented: $showPrivacy) {
            PrivacyPolicyView()
        }
    }
}

// Google "G" Logo
struct GoogleLogo: View {
    var body: some View {
        GeometryReader { geometry in
            let size = min(geometry.size.width, geometry.size.height)
            ZStack {
                // Blue arc (top-right)
                Circle()
                    .trim(from: 0.0, to: 0.25)
                    .stroke(Color(red: 0.26, green: 0.52, blue: 0.96), lineWidth: size * 0.18)
                    .rotationEffect(.degrees(-45))

                // Green arc (bottom-right)
                Circle()
                    .trim(from: 0.0, to: 0.25)
                    .stroke(Color(red: 0.20, green: 0.66, blue: 0.33), lineWidth: size * 0.18)
                    .rotationEffect(.degrees(45))

                // Yellow arc (bottom-left)
                Circle()
                    .trim(from: 0.0, to: 0.25)
                    .stroke(Color(red: 0.98, green: 0.74, blue: 0.02), lineWidth: size * 0.18)
                    .rotationEffect(.degrees(135))

                // Red arc (top-left)
                Circle()
                    .trim(from: 0.0, to: 0.25)
                    .stroke(Color(red: 0.92, green: 0.26, blue: 0.21), lineWidth: size * 0.18)
                    .rotationEffect(.degrees(225))

                // Horizontal bar for "G"
                Rectangle()
                    .fill(Color(red: 0.26, green: 0.52, blue: 0.96))
                    .frame(width: size * 0.45, height: size * 0.18)
                    .offset(x: size * 0.12, y: 0)
            }
            .frame(width: size, height: size)
        }
    }
}

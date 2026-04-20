import SwiftUI

struct TermsOfServiceView: View {
    @Environment(\.dismiss) var dismiss

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(alignment: .leading, spacing: 24) {
                    Text("Last updated: December 27, 2025")
                        .font(.system(size: 13))
                        .foregroundStyle(.secondary)

                    Group {
                        section(title: "1. Acceptance of Terms") {
                            """
                            By downloading, installing, or using Atoma ("the App"), you agree to be bound by these Terms of Service. If you do not agree to these terms, please do not use the App.
                            """
                        }

                        section(title: "2. Description of Service") {
                            """
                            Atoma is a personal productivity application that helps users track habits, manage daily tasks, and monitor personal finances. The App is provided "as is" for personal, non-commercial use.
                            """
                        }

                        section(title: "3. User Accounts") {
                            """
                            To use certain features, you must create an account using Google Sign-In or Apple Sign-In. You are responsible for maintaining the confidentiality of your account and for all activities under your account. You must provide accurate information and keep it updated.
                            """
                        }

                        section(title: "4. User Data & Privacy") {
                            """
                            We collect and store the following data:
                            • Account information (email, name)
                            • Habits, tasks, and financial transactions you create
                            • App usage analytics

                            Your data is stored securely and is not sold to third parties. You can request deletion of your data at any time by contacting support or deleting your account.
                            """
                        }

                        section(title: "5. Acceptable Use") {
                            """
                            You agree not to:
                            • Use the App for any illegal purpose
                            • Attempt to gain unauthorized access to our systems
                            • Interfere with the proper functioning of the App
                            • Upload malicious content or code
                            • Use automated systems to access the App
                            """
                        }

                        section(title: "6. Financial Data Disclaimer") {
                            """
                            The App provides tools for personal budget tracking. It does not provide financial, investment, or tax advice. All financial decisions are your own responsibility. The App does not connect to bank accounts or process real transactions.
                            """
                        }

                        section(title: "7. AI Features") {
                            """
                            The App includes AI-powered features for insights and recommendations. AI responses are generated automatically and should not be considered professional advice. Use AI suggestions at your own discretion.
                            """
                        }

                        section(title: "8. Intellectual Property") {
                            """
                            All content, features, and functionality of the App are owned by Atoma and are protected by copyright and other intellectual property laws. You may not copy, modify, or distribute any part of the App.
                            """
                        }

                        section(title: "9. Limitation of Liability") {
                            """
                            To the maximum extent permitted by law, Atoma shall not be liable for any indirect, incidental, special, or consequential damages arising from your use of the App. We do not guarantee uninterrupted or error-free service.
                            """
                        }

                        section(title: "10. Changes to Terms") {
                            """
                            We reserve the right to modify these terms at any time. Continued use of the App after changes constitutes acceptance of the new terms. We will notify users of significant changes through the App.
                            """
                        }

                        section(title: "11. Termination") {
                            """
                            We may terminate or suspend your access to the App at any time, without prior notice, for conduct that we believe violates these Terms or is harmful to other users or us.
                            """
                        }

                        section(title: "12. Contact") {
                            """
                            For questions about these Terms, please contact:
                            support@azamatbigali.online
                            """
                        }
                    }
                }
                .padding()
            }
            .navigationTitle("Terms of Service")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button("Done") {
                        dismiss()
                    }
                }
            }
        }
    }

    private func section(title: String, content: () -> String) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(title)
                .font(.system(size: 16, weight: .semibold))

            Text(content())
                .font(.system(size: 14))
                .foregroundStyle(.secondary)
                .fixedSize(horizontal: false, vertical: true)
        }
    }
}

struct PrivacyPolicyView: View {
    @Environment(\.dismiss) var dismiss

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(alignment: .leading, spacing: 24) {
                    Text("Last updated: December 27, 2025")
                        .font(.system(size: 13))
                        .foregroundStyle(.secondary)

                    Group {
                        section(title: "Information We Collect") {
                            """
                            • Account data: Email address and name from Google/Apple Sign-In
                            • User content: Habits, tasks, and transactions you create
                            • Device info: Device type, OS version for app compatibility
                            • Usage data: Anonymous analytics to improve the app
                            """
                        }

                        section(title: "How We Use Your Data") {
                            """
                            • To provide and maintain the App
                            • To sync your data across devices
                            • To generate personalized insights
                            • To improve app functionality
                            • To send important service notifications
                            """
                        }

                        section(title: "Data Storage & Security") {
                            """
                            Your data is stored on secure servers with encryption. We use industry-standard security measures to protect your information. Data is transmitted using HTTPS/TLS encryption.
                            """
                        }

                        section(title: "Data Sharing") {
                            """
                            We do not sell your personal data. We may share data only:
                            • With your consent
                            • To comply with legal obligations
                            • With service providers who assist our operations (under strict confidentiality)
                            """
                        }

                        section(title: "Your Rights") {
                            """
                            You have the right to:
                            • Access your personal data
                            • Correct inaccurate data
                            • Delete your account and data
                            • Export your data
                            • Opt-out of analytics

                            To exercise these rights, contact support@azamatbigali.online
                            """
                        }

                        section(title: "Data Retention") {
                            """
                            We retain your data while your account is active. After account deletion, data is permanently removed within 30 days, except where retention is required by law.
                            """
                        }

                        section(title: "Children's Privacy") {
                            """
                            The App is not intended for children under 13. We do not knowingly collect data from children. If we learn we have collected such data, we will delete it promptly.
                            """
                        }

                        section(title: "Contact Us") {
                            """
                            For privacy questions or concerns:
                            support@azamatbigali.online
                            """
                        }
                    }
                }
                .padding()
            }
            .navigationTitle("Privacy Policy")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button("Done") {
                        dismiss()
                    }
                }
            }
        }
    }

    private func section(title: String, content: () -> String) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(title)
                .font(.system(size: 16, weight: .semibold))

            Text(content())
                .font(.system(size: 14))
                .foregroundStyle(.secondary)
                .fixedSize(horizontal: false, vertical: true)
        }
    }
}

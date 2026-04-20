import SwiftUI
import UIKit

// MARK: - Haptic Feedback Manager
enum HapticManager {
    static func impact(_ style: UIImpactFeedbackGenerator.FeedbackStyle = .medium) {
        let generator = UIImpactFeedbackGenerator(style: style)
        generator.impactOccurred()
    }

    static func notification(_ type: UINotificationFeedbackGenerator.FeedbackType) {
        let generator = UINotificationFeedbackGenerator()
        generator.notificationOccurred(type)
    }

    static func selection() {
        let generator = UISelectionFeedbackGenerator()
        generator.selectionChanged()
    }

    // Specific feedback for app actions
    static func completionSuccess() {
        notification(.success)
    }

    static func toggleOn() {
        impact(.light)
    }

    static func toggleOff() {
        impact(.soft)
    }
}

// MARK: - Celebration Effect View
struct CelebrationEffect: View {
    let isActive: Bool

    @State private var particles: [Particle] = []

    private struct Particle: Identifiable {
        let id = UUID()
        var x: CGFloat
        var y: CGFloat
        var scale: CGFloat
        var opacity: Double
        var rotation: Double
        let emoji: String
    }

    private let emojis = ["✨", "🎉", "⭐️", "💫", "🌟"]

    var body: some View {
        GeometryReader { geometry in
            ZStack {
                ForEach(particles) { particle in
                    Text(particle.emoji)
                        .font(.system(size: 20))
                        .scaleEffect(particle.scale)
                        .opacity(particle.opacity)
                        .rotationEffect(.degrees(particle.rotation))
                        .position(x: particle.x, y: particle.y)
                }
            }
        }
        .onChange(of: isActive) { _, newValue in
            if newValue {
                createParticles()
            }
        }
        .allowsHitTesting(false)
    }

    private func createParticles() {
        particles = (0..<8).map { _ in
            Particle(
                x: CGFloat.random(in: 20...100),
                y: CGFloat.random(in: 0...40),
                scale: CGFloat.random(in: 0.5...1.2),
                opacity: 1.0,
                rotation: Double.random(in: 0...360),
                emoji: emojis.randomElement() ?? "✨"
            )
        }

        // Animate particles
        withAnimation(.easeOut(duration: 0.8)) {
            particles = particles.map { particle in
                var p = particle
                p.y = particle.y - CGFloat.random(in: 30...60)
                p.x = particle.x + CGFloat.random(in: -30...30)
                p.opacity = 0
                p.scale = particle.scale * 0.5
                p.rotation = particle.rotation + Double.random(in: -90...90)
                return p
            }
        }

        // Clear particles after animation
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.8) {
            particles = []
        }
    }
}

// MARK: - Completion Celebration Modifier
struct CompletionCelebration: ViewModifier {
    @Binding var isCompleted: Bool
    @State private var showCelebration = false
    @State private var previousState = false

    func body(content: Content) -> some View {
        content
            .overlay(alignment: .center) {
                CelebrationEffect(isActive: showCelebration)
                    .frame(width: 120, height: 80)
                    .offset(y: -20)
            }
            .onChange(of: isCompleted) { oldValue, newValue in
                // Only celebrate when going from incomplete to complete
                if !previousState && newValue {
                    showCelebration = true
                    HapticManager.completionSuccess()

                    DispatchQueue.main.asyncAfter(deadline: .now() + 0.1) {
                        showCelebration = false
                    }
                }
                previousState = newValue
            }
            .onAppear {
                previousState = isCompleted
            }
    }
}

// MARK: - Checkmark Animation View
struct AnimatedCheckmark: View {
    let isCompleted: Bool
    let color: Color
    var size: CGFloat = 28

    @State private var scale: CGFloat = 1.0

    var body: some View {
        Image(systemName: isCompleted ? "checkmark.circle.fill" : "circle")
            .font(.system(size: size))
            .foregroundStyle(isCompleted ? color : Color.hf.checkmarkIncomplete)
            .scaleEffect(scale)
            .onChange(of: isCompleted) { oldValue, newValue in
                if newValue && !oldValue {
                    // Bounce animation when completing
                    withAnimation(.spring(response: 0.3, dampingFraction: 0.5)) {
                        scale = 1.3
                    }
                    DispatchQueue.main.asyncAfter(deadline: .now() + 0.15) {
                        withAnimation(.spring(response: 0.2, dampingFraction: 0.6)) {
                            scale = 1.0
                        }
                    }
                }
            }
    }
}

// MARK: - View Extension
extension View {
    func celebrateCompletion(isCompleted: Binding<Bool>) -> some View {
        modifier(CompletionCelebration(isCompleted: isCompleted))
    }
}

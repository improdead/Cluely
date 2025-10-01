import SwiftUI

struct ListeningIndicator: View {
  @State private var pulseScale: CGFloat = 1.0

  var body: some View {
    HStack(spacing: 8) {
      Image(systemName: "waveform.circle.fill")
        .font(.system(size: 16))
        .foregroundStyle(
          LinearGradient(
            colors: [.blue, .cyan],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
          )
        )
        .scaleEffect(pulseScale)
        .onAppear {
          withAnimation(.easeInOut(duration: 0.8).repeatForever(autoreverses: true)) {
            pulseScale = 1.15
          }
        }
      Text("Listening...")
        .font(.subheadline.weight(.semibold))
        .foregroundStyle(.white)
    }
    .padding(.horizontal, 16)
    .padding(.vertical, 8)
    .background(
      ZStack {
        Capsule()
          .fill(.ultraThinMaterial)
          .opacity(0.85)
        Capsule()
          .fill(
            LinearGradient(
              colors: [.blue.opacity(0.3), .cyan.opacity(0.2)],
              startPoint: .topLeading,
              endPoint: .bottomTrailing
            )
          )
      }
    )
    .overlay(
      Capsule()
        .stroke(
          LinearGradient(
            colors: [.white.opacity(0.5), .blue.opacity(0.3)],
            startPoint: .top,
            endPoint: .bottom
          ),
          lineWidth: 1.5
        )
    )
    .shadow(color: .blue.opacity(0.5), radius: 10, x: 0, y: 4)
  }
}

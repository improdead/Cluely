import SwiftUI

struct OrnamentControls: View {
  @EnvironmentObject var s: SessionController
  var body: some View {
    HStack(spacing: 16) {
      Button(action: { s.toggleManual() }) {
        HStack(spacing: 6) {
          Image(systemName: s.isRunning ? "stop.circle.fill" : "play.circle.fill")
            .font(.system(size: 18))
          Text(s.isRunning ? "Stop" : "Start")
            .font(.subheadline.weight(.semibold))
        }
        .foregroundStyle(.white)
      }
      .buttonStyle(.borderless)

      Button(action: { s.markMoment() }) {
        HStack(spacing: 6) {
          Image(systemName: "bookmark.fill")
            .font(.system(size: 16))
          Text("Mark")
            .font(.subheadline.weight(.medium))
        }
        .foregroundStyle(.white.opacity(0.9))
      }
      .buttonStyle(.borderless)

      Toggle(isOn: $s.captureFrames) {
        HStack(spacing: 6) {
          Image(systemName: "camera.viewfinder")
            .font(.system(size: 16))
          Text("Frames")
            .font(.subheadline.weight(.medium))
        }
      }
      .toggleStyle(.button)
      .foregroundStyle(s.captureFrames ? .green : .white.opacity(0.9))

      Button("Demo") { s.demoHint() }
        .font(.subheadline.weight(.medium))
        .foregroundStyle(.white.opacity(0.7))
        .buttonStyle(.borderless)
    }
    .padding(.horizontal, 20)
    .padding(.vertical, 12)
    .background(
      ZStack {
        Capsule()
          .fill(.ultraThinMaterial)
          .opacity(0.85)
        Capsule()
          .fill(
            LinearGradient(
              colors: [.white.opacity(0.1), .white.opacity(0.03)],
              startPoint: .top,
              endPoint: .bottom
            )
          )
      }
    )
    .overlay(
      Capsule()
        .stroke(.white.opacity(0.3), lineWidth: 1.5)
    )
    .shadow(color: .black.opacity(0.2), radius: 12, x: 0, y: 4)
  }
}

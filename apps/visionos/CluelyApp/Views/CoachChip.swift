import SwiftUI

struct CoachChip: View {
  var text: String
  var body: some View {
    Text(text)
      .font(.subheadline.weight(.medium))
      .lineLimit(2)
      .foregroundStyle(.white)
      .padding(.horizontal, 20)
      .padding(.vertical, 12)
      .background(
        ZStack {
          Capsule()
            .fill(.ultraThinMaterial)
            .opacity(0.9)
          Capsule()
            .fill(
              LinearGradient(
                colors: [.white.opacity(0.12), .white.opacity(0.05)],
                startPoint: .top,
                endPoint: .bottom
              )
            )
        }
      )
      .overlay(
        Capsule()
          .stroke(
            LinearGradient(
              colors: [.white.opacity(0.5), .white.opacity(0.2)],
              startPoint: .top,
              endPoint: .bottom
            ),
            lineWidth: 1.5
          )
      )
      .shadow(color: .black.opacity(0.25), radius: 15, x: 0, y: 6)
  }
}

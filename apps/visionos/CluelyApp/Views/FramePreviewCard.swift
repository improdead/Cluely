import SwiftUI

struct FramePreviewCard: View {
  var body: some View {
    VStack(alignment: .leading, spacing: 12) {
      HStack {
        Image(systemName: "camera.viewfinder")
          .font(.system(size: 18))
          .foregroundStyle(.green.opacity(0.8))
        Text("Frame Capture")
          .font(.headline.weight(.semibold))
          .foregroundStyle(.white)
        Spacer()
        Circle()
          .fill(.green)
          .frame(width: 8, height: 8)
      }
      Divider()
        .background(.white.opacity(0.2))
      Text("Capturing screen content\nfor visual context...")
        .font(.caption)
        .foregroundStyle(.white.opacity(0.7))
        .lineSpacing(4)
    }
    .padding(20)
    .frame(width: 320, alignment: .leading)
    .background(
      ZStack {
        RoundedRectangle(cornerRadius: 20)
          .fill(.ultraThinMaterial)
          .opacity(0.75)
        RoundedRectangle(cornerRadius: 20)
          .fill(
            LinearGradient(
              colors: [.green.opacity(0.15), .green.opacity(0.05)],
              startPoint: .topLeading,
              endPoint: .bottomTrailing
            )
          )
      }
    )
    .overlay(
      RoundedRectangle(cornerRadius: 20)
        .stroke(
          LinearGradient(
            colors: [.green.opacity(0.5), .green.opacity(0.2)],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
          ),
          lineWidth: 1.5
        )
    )
    .shadow(color: .green.opacity(0.2), radius: 15, x: 0, y: 6)
  }
}

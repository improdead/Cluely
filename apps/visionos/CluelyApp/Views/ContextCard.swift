import SwiftUI

struct ContextCard: View {
  var body: some View {
    VStack(alignment: .leading, spacing: 12) {
      HStack {
        Image(systemName: "doc.text.viewfinder")
          .font(.system(size: 18))
          .foregroundStyle(.white.opacity(0.7))
        Text("Context")
          .font(.headline.weight(.semibold))
          .foregroundStyle(.white)
        Spacer()
      }
      Divider()
        .background(.white.opacity(0.2))
      Text("Visual context and OCR tokens\nappear here when enabled.")
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
              colors: [.white.opacity(0.08), .white.opacity(0.02)],
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
            colors: [.white.opacity(0.4), .white.opacity(0.1)],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
          ),
          lineWidth: 1.5
        )
    )
    .shadow(color: .black.opacity(0.15), radius: 20, x: 0, y: 8)
  }
}

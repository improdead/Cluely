import SwiftUI

struct SuggestionPanel: View {
  var text: String
  var body: some View {
    VStack(alignment: .leading, spacing: 16) {
      HStack {
        Image(systemName: "lightbulb.fill")
          .font(.system(size: 18))
          .foregroundStyle(
            LinearGradient(
              colors: [.yellow, .orange],
              startPoint: .topLeading,
              endPoint: .bottomTrailing
            )
          )
        Text("Suggestion")
          .font(.headline.weight(.semibold))
          .foregroundStyle(.white)
        Spacer()
      }
      Divider()
        .background(.white.opacity(0.2))
      if text.isEmpty || text == "—" {
        Text("Listening for context...")
          .font(.body)
          .foregroundStyle(.white.opacity(0.6))
          .frame(maxWidth: .infinity, alignment: .leading)
      } else {
        Text(text)
          .font(.body.weight(.medium))
          .lineLimit(5)
          .foregroundStyle(.white)
          .frame(maxWidth: .infinity, alignment: .leading)
          .lineSpacing(4)
      }
    }
    .padding(24)
    .frame(width: 380)
    .background(
      ZStack {
        RoundedRectangle(cornerRadius: 24)
          .fill(.ultraThinMaterial)
          .opacity(0.8)
        RoundedRectangle(cornerRadius: 24)
          .fill(
            LinearGradient(
              colors: [.white.opacity(0.1), .white.opacity(0.03)],
              startPoint: .topLeading,
              endPoint: .bottomTrailing
            )
          )
      }
    )
    .overlay(
      RoundedRectangle(cornerRadius: 24)
        .stroke(
          LinearGradient(
            colors: [.white.opacity(0.5), .white.opacity(0.15)],
            startPoint: .top,
            endPoint: .bottom
          ),
          lineWidth: 2
        )
    )
    .shadow(color: .black.opacity(0.2), radius: 25, x: 0, y: 10)
  }
}

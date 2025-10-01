import SwiftUI

struct WarningBadge: View {
  var text: String
  var body: some View {
    HStack(spacing: 6) {
      Image(systemName: "exclamationmark.triangle.fill")
        .font(.system(size: 14))
        .foregroundStyle(.yellow)
      Text(text)
        .font(.caption.weight(.semibold))
        .foregroundStyle(.white)
    }
    .padding(.horizontal, 12)
    .padding(.vertical, 6)
    .background(
      RoundedRectangle(cornerRadius: 12)
        .fill(.ultraThinMaterial)
        .opacity(0.9)
    )
    .overlay(
      RoundedRectangle(cornerRadius: 12)
        .stroke(.white.opacity(0.35), lineWidth: 1)
    )
    .shadow(color: .black.opacity(0.2), radius: 10, x: 0, y: 4)
  }
}

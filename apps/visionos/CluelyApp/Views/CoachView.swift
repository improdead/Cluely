import SwiftUI

struct CoachView: View {
  var text: String
  var body: some View {
    Text(text.isEmpty ? "—" : text)
      .font(.headline)
      .lineLimit(2)
      .padding(8)
      .background(.thinMaterial, in: Capsule())
  }
}

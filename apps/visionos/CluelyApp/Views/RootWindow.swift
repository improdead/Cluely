import SwiftUI

struct RootWindow: View {
  @EnvironmentObject var s: SessionController
  var body: some View {
    ZStack {
      // Main content area
      HStack(alignment: .top, spacing: 32) {
        VStack(alignment: .leading, spacing: 16) {
          ContextCard()
          if s.captureFrames {
            FramePreviewCard()
          }
        }
        Spacer()
        VStack(alignment: .trailing, spacing: 16) {
          SuggestionPanel(text: s.currentHint)
        }
      }
      .padding(32)

      // Top-left listening indicator
      VStack(alignment: .leading, spacing: 8) {
        if s.isRunning {
          ListeningIndicator()
            .padding(.top, 20)
            .padding(.leading, 20)
            .transition(.scale.combined(with: .opacity))
            .animation(.spring(response: 0.3, dampingFraction: 0.7), value: s.isRunning)
        }
        if !s.warningText.isEmpty {
          WarningBadge(text: s.warningText)
            .padding(.leading, 20)
            .transition(.opacity)
        }
        Spacer()
      }

      // Bottom coach chip (optional, shows over everything)
      VStack {
        Spacer()
        if !s.currentHint.isEmpty && s.currentHint != "—" {
          CoachChip(text: s.currentHint)
            .padding(.bottom, 80)
            .transition(.move(edge: .bottom).combined(with: .opacity))
            .animation(.spring(response: 0.4, dampingFraction: 0.75), value: s.currentHint)
        }
      }
    }
    .ornament(attachmentAnchor: .scene(.bottom)) {
      OrnamentControls().environmentObject(s)
    }
  }
}

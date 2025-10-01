import SwiftUI

@main
struct CluelyApp: App {
  @StateObject var session = SessionController()
  var body: some Scene {
    WindowGroup("Cluely") {
      RootWindow()
        .environmentObject(session)
    }
    .windowStyle(.plain)
  }
}

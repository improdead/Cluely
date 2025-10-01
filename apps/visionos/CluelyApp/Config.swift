import Foundation

enum AppConfig {
  static var wsURL: URL {
    if let s = Bundle.main.object(forInfoDictionaryKey: "WSURL") as? String,
       let u = URL(string: s) {
      return u
    }
    // Default for local testing (simulator or same-host)
    return URL(string: "ws://localhost:8080/ws")!
  }
}

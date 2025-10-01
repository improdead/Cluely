import Foundation

final class WSClient {
  private var task: URLSessionWebSocketTask?
  var onEvent: ((String)->Void)?
  func connect() {
    let s = URLSession(configuration: .default)
    task = s.webSocketTask(with: AppConfig.wsURL)
    task?.resume()
    send(text: #"{"type":"hello","app":"cluely-visionos"}"#)
    pump()
  }
  func sendPCM(_ data: Data) {
    task?.send(.data(data)) { _ in }
  }
  func send(text: String) { task?.send(.string(text)) { _ in } }
  func close() { task?.cancel(with: .goingAway, reason: nil) }
  private func pump() {
    task?.receive { [weak self] res in
      guard let self else { return }
      if case .success(let m) = res {
        switch m {
        case .string(let s): self.onEvent?(s)
        case .data(let d): if let s = String(data: d, encoding: .utf8) { self.onEvent?(s) }
        @unknown default: break
        }
        self.pump()
      }
    }
  }
}

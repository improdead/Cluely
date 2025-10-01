import ReplayKit
import Vision
import Foundation

// Broadcast Upload Extension entry point
class SampleHandler: RPBroadcastSampleHandler {
  private var lastCaptureTime = Date.distantPast
  private let captureInterval: TimeInterval = 0.5 // 2 fps
  private var ws: URLSessionWebSocketTask?

  override func broadcastStarted(withSetupInfo setupInfo: [String : NSObject]?) {
    // Connect WS to backend
    let url = Self.wsURL()
    let s = URLSession(configuration: .default)
    ws = s.webSocketTask(with: url)
    ws?.resume()
    send(text: #"{"type":"hello","app":"cluely-broadcast"}"#)
    pump()
  }

  override func broadcastFinished() {
    ws?.cancel(with: .goingAway, reason: nil)
    ws = nil
  }

  override func processSampleBuffer(_ sampleBuffer: CMSampleBuffer, with sampleBufferType: RPSampleBufferType) {
    guard sampleBufferType == .video else { return }
    let now = Date()
    guard now.timeIntervalSince(lastCaptureTime) >= captureInterval else { return }
    lastCaptureTime = now

    guard let pixel = CMSampleBufferGetImageBuffer(sampleBuffer) else { return }
    let req = VNRecognizeTextRequest { [weak self] req, err in
      guard let self else { return }
      guard let results = req.results as? [VNRecognizedTextObservation] else { return }
      var tokens: [String] = []
      for obs in results.prefix(10) {
        if let text = obs.topCandidates(1).first?.string { tokens.append(text) }
      }
      if !tokens.isEmpty {
        let obj: [String: Any] = ["type":"frame_meta", "ocr": tokens]
        if let data = try? JSONSerialization.data(withJSONObject: obj),
           let s = String(data: data, encoding: .utf8) {
          self.send(text: s)
        }
      }
    }
    req.recognitionLevel = .fast
    req.usesLanguageCorrection = false
    let handler = VNImageRequestHandler(cvPixelBuffer: pixel, options: [:])
    try? handler.perform([req])
  }

  private func pump() {
    ws?.receive { [weak self] _ in
      // ignore downstream
      self?.pump()
    }
  }

  private func send(text: String) {
    ws?.send(.string(text)) { _ in }
  }

  static func wsURL() -> URL {
    if let s = Bundle.main.object(forInfoDictionaryKey: "WSURL") as? String,
       let u = URL(string: s) {
      return u
    }
    return URL(string: "ws://localhost:8080/ws")!
  }
}

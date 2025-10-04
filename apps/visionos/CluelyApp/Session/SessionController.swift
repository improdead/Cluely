import Foundation
import AVFAudio

@MainActor
final class SessionController: ObservableObject {
  @Published var isRunning = false // reflects VAD/manual listening state
  @Published var captureFrames = false {
    didSet {
      if captureFrames {
        frames.start()
      } else {
        frames.stop()
      }
    }
  }
  @Published var currentHint = "—"
  @Published var warningText: String = ""
  private var log: [String] = []
  var logText: String { log.joined(separator: "\n") }

  private let audio = AudioEngine()
  private let ws = WSClient()
  private let coach = CoachRuntime()
  private let vad = SimpleVAD()
  private let cfg = VADConfig()
  private let frames = FrameSource()

  private var lastHint: String = ""
  private var lastFollow: String = ""
  private var sendAudio = false // gate for sending PCM to WS
  private var audioStarted = false
  private var firstFramePending = false
  private var lastOCRTokens: [String] = []

  init() {
    ws.onEvent = { [weak self] s in self?.handleEvent(s) }
    ws.connect()
    frames.onTokens = { [weak self] tokens in
      guard let self else { return }
      self.lastOCRTokens = tokens
      if self.firstFramePending {
        self.sendOCRTokens(tokens, first: true, last: false)
        self.firstFramePending = false
      } else if self.captureFrames {
        self.sendOCRTokens(tokens)
      }
    }
    do {
      try audio.prepare()
      try audio.start { [weak self] bufData, db in
        guard let self else { return }
        self.onAudioFrame(bufData: bufData, db: db)
      }
      audioStarted = true
    } catch {
      log.append("Audio start error: \(error)")
    }
  }

  func toggleManual() { isRunning ? stop() : start() }

  func start() {
    // Manual start: enable sending; audio engine already running
    if !audioStarted { do { try audio.start { [weak self] d, db in self?.onAudioFrame(bufData: d, db: db) }; audioStarted = true } catch { log.append("Audio start err: \(error)") } }
    sendAudio = true
    isRunning = true
    firstFramePending = true
    lastOCRTokens = []
    log.append("Start")
  }

  func stop() {
    guard isRunning || sendAudio else { return }
    sendAudio = false
    isRunning = false
    if !lastOCRTokens.isEmpty { sendOCRTokens(lastOCRTokens, first: false, last: true) }
    ws.send(text: #"{"type":"stop"}"#)
    log.append("Stop")
  }

  private func onAudioFrame(bufData: Data, db: Float) {
    // VAD
    let (open, close) = vad.onFrame(db: db, cfg: cfg)
    if open { start() }
    if close { stop() }
    // Send PCM only when listening
    if sendAudio {
      ws.sendPCM(bufData)
    }
  }

  func markMoment() { log.append("• moment \(Date())") }

  func demoHint() {
    ws.send(text: #"{"type":"transcript","text":"We can meet on Friday to review budget.","final":true}"#)
  }

  private func sendOCRTokens(_ tokens: [String], first: Bool = false, last: Bool = false) {
    var obj: [String: Any] = ["type": "frame_meta", "ocr": tokens]
    if first { obj["first"] = true }
    if last { obj["last"] = true }
    if let data = try? JSONSerialization.data(withJSONObject: obj), let s = String(data: data, encoding: .utf8) {
      ws.send(text: s)
    }
  }

  private func handleEvent(_ s: String) {
    guard let data = s.data(using: .utf8) else { return }
    guard let ev = try? JSONDecoder().decode(ServerEvent.self, from: data) else {
      log.append("decode error: \(s.prefix(80))…")
      return
    }
    switch ev.type {
    case "hint_partial":
      if let t = ev.text {
        lastHint = t
        currentHint = compose()
      }
    case "followup_partial":
      if let t = ev.text {
        lastFollow = t
        currentHint = compose()
      }
    case "hint":
      if let t = ev.text, coach.shouldShow(t) {
        lastHint = t
        currentHint = compose()
      }
    case "followup":
      if let t = ev.text, coach.shouldShow(t) {
        lastFollow = t
        currentHint = compose()
      }
    case "partial":
      break
    case "final":
      break
    case "state":
      // could use ev.listening if server controls state
      break
    case "warning":
      if let m = ev.msg {
        log.append("warning: \(m)")
        warningText = m
        // Auto clear after 3 seconds
        Task { @MainActor in
          try? await Task.sleep(nanoseconds: 3_000_000_000)
          if self.warningText == m { self.warningText = "" }
        }
      }
    case "error":
      if let m = ev.msg { log.append("error: \(m)") }
    default:
      break
    }
  }

  private func compose() -> String {
    if lastFollow.isEmpty { return lastHint.isEmpty ? "—" : lastHint }
    if lastHint.isEmpty { return lastFollow }
    return lastHint + "\n" + lastFollow
  }
}

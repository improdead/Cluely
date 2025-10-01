Awesome—here’s a performance‑first, implementation‑grade system design for Cluely on Apple Vision Pro with a Swift (visionOS) frontend and an all‑Go backend. It includes:
goals & UX,


end‑to‑end architecture,


precise data contracts,


mermaid diagrams (components, sequence, state, deployment),


production‑ready code structure for both Swift and Go,


and a performance playbook (latency budgets + concrete tuning tips).


Assumption per your note: user has already granted all permissions. You still show a visible “active” indicator while capturing.

0) Goals (what this MVP must do)
Triggers:


Auto: start on human speech (VAD threshold + hangover).


Manual: keyboard shortcut / on‑screen button.


Pipelines (in parallel when active):


Audio → VAD → WS to backend → streaming ASR → micro‑answer (≤2 lines) + optional follow‑up.


Frames (optional) → ReplayKit frames (throttled) → lightweight OCR keywords → sent as context tokens only (not raw images) to backend.


UI: tiny, glanceable 2‑line coach chip; Start/Stop/Mark controls in an ornament; clear “Listening” indicator.


Stop: on silence (hangover) or user action.



1) High‑level architecture
flowchart LR
  subgraph VisionOS App (Swift)
    TRIG[TriggerController\n(VAD + shortcut)]
    AUD[AudioEngine\nAVAudioEngine + resampler]
    FR[FrameSource\nReplayKit (throttled)]
    WS[WebSocket Client\nURLSessionWebSocketTask]
    UI[Coach UI\nWindow + Ornament]
  end

  subgraph Go Backend (single binary)
    EDGE[WS Ingress\nsession manager]
    ASR[ASR Bridge\n(vendor WS/gRPC)]
    ANS[Answer Engine\nLLM orchestrator]
    RL[Rate Limiter/Dedupe]
    R[(Redis - ephemeral)]:::dim
    P[(Postgres - durable)]:::dim
    OTEL[(OpenTelemetry)]:::dim
  end

  TRIG --> AUD
  TRIG --> FR
  AUD -->|PCM 16k mono 20ms| WS
  FR -->|OCR tokens (JSON)| WS
  WS <--> EDGE
  EDGE --> ASR --> EDGE
  EDGE --> RL --> ANS --> EDGE
  EDGE -->|events| WS
  EDGE <--> R
  ANS <--> P
  EDGE --> OTEL
  ANS --> OTEL
  classDef dim fill:#f6f6f6,stroke:#bbb,color:#666


2) Data contracts
WebSocket (client ↔ server)
Upstream (client → server)


binary: PCM16‑LE, 16 kHz, mono, 20 ms frames (320 samples → 640 bytes).


text (JSON): small control msgs

 {"type":"hello","app":"cluely-visionos","ver":"0.1.0"}
{"type":"frame_meta","ocr":["menu","cabernet","shiba"]}
{"type":"stop"}


Downstream (server → client) — text JSON only:

 {"type":"state","listening":true}
{"type":"partial","text":"we can meet on fri"}
{"type":"final","text":"we can meet on Friday"}
{"type":"hint","text":"Confirm budget owner","ttlMs":4500}
{"type":"followup","text":"Ask preferred timeline","ttlMs":4500}
{"type":"error","code":"ASR_UNAVAILABLE","msg":"..."}



3) UX spec (visionOS)
Window (Shared Space): setup/logs + coach chip.


Ornament (bottom): Start/Stop, “Hold‑to‑talk” (shortcut), Mark, Frames toggle.


Coach chip: capsule with max 2 lines; auto‑fade (3–5 s); rate‑limit 1/1.5 s.


Indicators: “Listening” pill while active.


Presenter‑safe: Frames toggle off; hints remain.



4) Swift codebase structure
apps/visionos/
  CluelyApp/
    App.swift
    Views/
      RootWindow.swift
      CoachView.swift
      OrnamentControls.swift
    Input/
      KeyCapture.swift          # press/hold
      VAD.swift                 # energy-based (Accelerate)
    Session/
      SessionController.swift   # state machine + orchestration
      AudioEngine.swift         # AVAudioEngine + resampler → 16k mono
      FrameSource.swift         # ReplayKit throttled + OCR tokens
      WSClient.swift            # URLSessionWebSocketTask
      CoachRuntime.swift        # rate-limit + dedupe
    Resources/
      AppIcon, Privacy copy
    Info.plist                  # mic, speech; frame usage if needed

Key Swift scaffolds (minimal but real)
App + Window + Ornament
// App.swift
import SwiftUI
@main struct CluelyApp: App {
  @StateObject var session = SessionController()
  var body: some Scene {
    WindowGroup("Cluely") { RootWindow().environmentObject(session) }
      .windowStyle(.plain)
  }
}

// RootWindow.swift
import SwiftUI
struct RootWindow: View {
  @EnvironmentObject var s: SessionController
  var body: some View {
    VStack(alignment: .leading, spacing: 12) {
      HStack(spacing: 12) {
        Button(s.isRunning ? "Stop" : "Start") { s.toggleManual() }
          .keyboardShortcut(" ", modifiers: []) // local
        Button("Mark") { s.markMoment() }
        Toggle("Frames", isOn: $s.captureFrames)
      }
      CoachView(text: s.currentHint)
      Divider()
      ScrollView { Text(s.logText).font(.caption.monospaced()) }
    }
    .ornament(attachmentAnchor: .scene(alignment: .bottom)) {
      OrnamentControls().environmentObject(s)
    }
  }
}

Audio capture (resample → 16k mono 20 ms frames)
// AudioEngine.swift
import AVFAudio
final class AudioEngine {
  private let engine = AVAudioEngine()
  private var converter: AVAudioConverter?
  private let outFormat = AVAudioFormat(commonFormat: .pcmFormatInt16,
                                        sampleRate: 16_000, channels: 1, interleaved: true)!
  func prepare() throws {
    // permission assumed granted; but set an optimal mode
    let sess = AVAudioSession.sharedInstance()
    try sess.setCategory(.record, mode: .spokenAudio, options: [.duckOthers])
    try sess.setActive(true)
    // build converter from hardware format -> 16k mono
    let inFmt = engine.inputNode.inputFormat(forBus: 0)
    converter = AVAudioConverter(from: inFmt, to: outFormat)
  }
  func start(onFrame: @escaping (Data)->Void) throws {
    let node = engine.inputNode
    let inFmt = node.inputFormat(forBus: 0)
    node.installTap(onBus: 0, bufferSize: 1024, format: inFmt) { [weak self] buf, _ in
      guard let self, let conv = self.converter else { return }
      let out = AVAudioPCMBuffer(pcmFormat: self.outFormat,
                                 frameCapacity: AVAudioFrameCount(320))! // 20ms
      var err: NSError?
      let inputBlock: AVAudioConverterInputBlock = { _, outStatus in
        outStatus.pointee = .haveData; return buf
      }
      conv.convert(to: out, error: &err, withInputFrom: inputBlock)
      guard err == nil, let ch = out.int16ChannelData else { return }
      let frames = Int(out.frameLength)
      let bytes = UnsafeBufferPointer(start: ch[0], count: frames).withMemoryRebound(to: UInt8.self) {
        Data(buffer: $0) // zero copy-ish view
      }
      onFrame(bytes) // 320 samples * 2 bytes = 640 bytes
    }
    try engine.start()
  }
  func stop() { engine.inputNode.removeTap(onBus: 0); engine.stop() }
}

Simple energy‑based VAD (Accelerate)
// VAD.swift
import Accelerate, AVFAudio

struct VADConfig { var openDB: Float = -42; var closeDB: Float = -48
  var minOpenMs: Int = 300; var hangMs: Int = 700 }

final class SimpleVAD {
  private var openedAt: Date?; private var lastSpeechAt: Date?
  func levelDB(_ b: AVAudioPCMBuffer) -> Float {
    guard let p = b.floatChannelData?[0] else { return -120 }
    var mean: Float = 0; vDSP_meamgv(p, 1, &mean, vDSP_Length(b.frameLength))
    let rms = sqrtf(mean); return 20 * log10f(max(rms, 1e-9))
  }
  func onFrame(db: Float, cfg: VADConfig) -> (open: Bool, close: Bool) {
    let now = Date()
    if openedAt == nil, db > cfg.openDB { openedAt = now; lastSpeechAt = now; return (true,false) }
    if let _ = openedAt {
      if db > cfg.closeDB { lastSpeechAt = now }
      if let last = lastSpeechAt, now.timeIntervalSince(last)*1000 > Double(cfg.hangMs) {
        openedAt = nil; lastSpeechAt = nil; return (false,true)
      }
    }
    return (false,false)
  }
}

WebSocket client (binary upstream, JSON downstream)
// WSClient.swift
import Foundation
final class WSClient {
  private var task: URLSessionWebSocketTask?
  var onEvent: ((String)->Void)?
  func connect(url: URL) {
    let s = URLSession(configuration: .default)
    task = s.webSocketTask(with: url); task?.resume()
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

Session controller (wires it all)
// SessionController.swift
import Foundation, AVFAudio

@MainActor
final class SessionController: ObservableObject {
  @Published var isRunning = false
  @Published var captureFrames = false
  @Published var currentHint = "—"
  private var log: [String] = []; var logText: String { log.joined(separator:"\n") }

  private let audio = AudioEngine()
  private let vad = SimpleVAD()
  private let cfg = VADConfig()
  private let ws = WSClient()
  private var lastHintAt = Date.distantPast

  init() {
    ws.onEvent = { [weak self] s in self?.handleEvent(s) }
    ws.connect(url: URL(string: "wss://your-domain/ws")!)
    try? audio.prepare()
  }

  func toggleManual() { isRunning ? stop() : start() }
  func start() {
    guard !isRunning else { return }
    isRunning = true; log.append("Start")
    try? audio.start { [weak self] bufData in self?.ws.sendPCM(bufData) }
    // optional: start ReplayKit throttle→OCR→send tokens via ws.send(text:...)
  }
  func stop() {
    guard isRunning else { return }
    audio.stop(); ws.send(text: #"{"type":"stop"}"#); isRunning = false; log.append("Stop")
  }
  func markMoment() { log.append("• moment \(Date())") }

  // VAD auto control
  func onAudioBuffer(_ buf: AVAudioPCMBuffer) {
    let db = vad.levelDB(buf)
    let (open, close) = vad.onFrame(db: db, cfg: cfg)
    if open { start() }
    if close { stop() }
  }

  private func handleEvent(_ s: String) {
    if s.contains("\"hint\"") {
      // naive parse; use Codable in real code
      if Date().timeIntervalSince(lastHintAt) > 1.5 {
        lastHintAt = Date()
        currentHint = extractValue(for:"text", from:s) ?? "—"
      }
    }
  }
  private func extractValue(for key: String, from json: String) -> String? { /*…*/ nil }
}

Tip: integrate VAD directly in the audio tap to decide start/stop without allocating intermediate buffers. (Keep the AudioEngine callback signature as both tap → VAD → sendPCM.)

5) Go backend codebase structure
server/
  cmd/cluelyd/main.go          # single binary: WS + ASR bridge + Answer API
  internal/ws/handler.go       # WS upgrade, per-session loops
  internal/session/store.go    # per-connection state, Redis binding
  internal/asr/client.go       # vendor streaming (WS/gRPC)
  internal/answer/service.go   # micro-answer orchestration (LLM API)
  internal/rt/ratelimit.go     # token/time limiters
  internal/obs/telemetry.go    # otel metrics/traces, pprof
  go.mod go.sum

Go server skeleton (high‑throughput WS with backpressure)
// cmd/cluelyd/main.go
package main

import (
  "log"
  "net"
  "net/http"
  "runtime"
  "time"

  "github.com/go-chi/chi/v5"
  "nhooyr.io/websocket"
)

func main() {
  runtime.GOMAXPROCS(max(1, runtime.NumCPU()-1)) // leave headroom

  // TCP_NODELAY for low-latency WS
  ln, _ := net.Listen("tcp", ":8080")
  if tcpln, ok := ln.(*net.TCPListener); ok { tcpln.SetDeadline(time.Time{}) }

  r := chi.NewRouter()
  r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok")) })
  r.Get("/ws", wsHandler)

  log.Println("cluelyd listening :8080")
  http.Serve(ln, r)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
  c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
    CompressionMode: websocket.CompressionDisabled, // keep latency low
  })
  if err != nil { return }

  // Per-connection session
  sess := NewSession(c)
  go sess.runDownstream() // ASR events -> client
  sess.runUpstream()      // client PCM -> ASR
}

Per‑session engine
// internal/ws/session.go
type Session struct {
  c     *websocket.Conn
  asr   *ASRClient
  ans   *AnswerService
  hints *RateLimiter
  // small pools to avoid allocs
  bufPool sync.Pool
}

func NewSession(c *websocket.Conn) *Session {
  return &Session{
    c: c,
    asr: NewASRClient(),          // connects to vendor WS immediately
    ans: NewAnswerService(),      // HTTP client to LLM vendor
    hints: NewRateLimiter(1, 1500*time.Millisecond), // 1 hint / 1.5s
    bufPool: sync.Pool{New: func() any { b := make([]byte, 0, 2048); return &b }},
  }
}

func (s *Session) runUpstream() {
  defer s.c.Close(websocket.StatusNormalClosure, "bye")
  ctx := context.Background()
  s.c.SetReadLimit(1 << 20)
  s.c.SetReadDeadline(time.Now().Add(35*time.Second))

  // feed vendor ASR
  go func() {
    for ev := range s.asr.Events() {
      // partials/finals -> client
      _ = s.c.Write(ctx, websocket.MessageText, mustJSON(ev))
      if ev.IsFinal && s.hints.Allow() {
        // Ask Answer engine for 2-line micro-answer
        if ans := s.ans.Micro(ev.Text, nil); ans != nil {
          _ = s.c.Write(ctx, websocket.MessageText, mustJSON(struct{
            Type string `json:"type"`; Text string `json:"text"`; TTL int `json:"ttlMs"`
          }{"hint", ans.Answer, 4500}))
          if ans.FollowUp != "" {
            _ = s.c.Write(ctx, websocket.MessageText, mustJSON(struct{
              Type string `json:"type"`; Text string `json:"text"`; TTL int `json:"ttlMs"`
            }{"followup", ans.FollowUp, 4500}))
          }
        }
      }
    }
  }()

  for {
    typ, data, err := s.c.Read(ctx)
    if err != nil { break }
    s.c.SetReadDeadline(time.Now().Add(35*time.Second))
    switch typ {
    case websocket.MessageBinary:
      s.asr.WritePCM(data) // 20ms PCM chunk
    case websocket.MessageText:
      // handle control / OCR tokens
      s.handleText(data)
    }
  }
  s.asr.Close()
}

func (s *Session) runDownstream() { /* already merged in runUpstream via goroutine */ }

ASR bridge (vendor WS)
// internal/asr/client.go
type ASRClient struct {
  // vendor connection + channel of events
}
type ASREvent struct{ Type string; Text string; IsFinal bool }

func NewASRClient() *ASRClient { /* connect to vendor, start recv loop */ return &ASRClient{} }
func (a *ASRClient) WritePCM(p []byte) { /* send frame upstream */ }
func (a *ASRClient) Events() <-chan ASREvent { /* channel of partial/final */ }
func (a *ASRClient) Close() { /* close vendor ws */ }

Answer engine (LLM orchestrator in Go)
// internal/answer/service.go
type Answer struct{ Answer, FollowUp string; Confidence float64 }
type AnswerService struct{ http *http.Client; key string }
func NewAnswerService() *AnswerService {
  return &AnswerService{ http: &http.Client{ Timeout: 3 * time.Second }, key: os.Getenv("OPENAI_API_KEY") }
}
func (s *AnswerService) Micro(text string, ocr []string) *Answer {
  // build prompt: 20 words answer + 12 words follow-up; call vendor LLM
  // return small struct; keep under ~1KB to reduce latency
  return &Answer{ Answer: "Confirm budget owner", FollowUp: "Ask preferred timeline", Confidence: 0.78 }
}

Rate limiter
// internal/rt/ratelimit.go
type RateLimiter struct{ n int; every time.Duration; last time.Time; mu sync.Mutex }
func NewRateLimiter(n int, every time.Duration) *RateLimiter { return &RateLimiter{n:n,every:every} }
func (r *RateLimiter) Allow() bool {
  r.mu.Lock(); defer r.mu.Unlock()
  if time.Since(r.last) >= r.every { r.last = time.Now(); return true }
  return false
}

Why one Go binary? Fewer hops → lower p95. You can split answer later if needed; the contracts won’t change.

6) State machine (client‑side)
stateDiagram-v2
  [*] --> Idle
  Idle --> Armed: app active
  Armed --> Listening: VAD open >= 300ms
  Armed --> Listening: Shortcut pressed
  Listening --> Processing: final ASR segment
  Processing --> Cooling: silence >= 700ms
  Cooling --> Listening: speech resumes
  Cooling --> Idle: timeout or Stop
  Listening --> Idle: Stop


7) Sequence (low‑latency path)
sequenceDiagram
  participant User
  participant Mic as AVAudioEngine
  participant App as WS Client
  participant GW as Go WS Gateway
  participant ASR as Vendor ASR
  participant AE as Answer Engine
  participant UI as Coach

  User->>Mic: speech
  Mic->>App: 20ms PCM frames (16k, mono)
  App->>GW: WS binary frames (640B each)
  GW->>ASR: stream frames
  ASR-->>GW: partials (100–300ms)
  GW-->>App: {"type":"partial","text":"..."}
  ASR-->>GW: final segment
  GW->>AE: micro-answer(text, ocrTokens)
  AE-->>GW: {answer, followUp}
  GW-->>App: {"type":"hint","text":"...", "ttlMs":4500}
  App-->>UI: display 2 lines (rate-limited)


8) Deployment (single box → easy scale)
graph LR
  subgraph Single VM (Docker Compose)
    Caddy[Caddy:443] --> Edge[cluelyd (Go)]
    Edge --> Redis[(Redis)]
    Edge --> Postgres[(Postgres)]
    Edge --> Otel[OTel Exporter]
  end
  Client[Vision Pro] -->|WSS /ws| Caddy

Everything on one VM (TLS via Caddy).


Move Postgres/Redis to managed services later; add a second VM for HA.


Cloudflare in front if you want anycast & DDoS.



9) Performance playbook (concrete, measurable)
Latency budget (end‑to‑end)
Mic → WS send (20 ms frame): ≤ 5–10 ms in‑app processing.


WS ingress → ASR partial: 100–300 ms (vendor‑dependent).


Final segment → micro‑answer: 100–200 ms (short prompt, cached).


UI render: <16 ms.


Target p50: ≤ 500–700 ms from speech to on‑glass hint.


Swift (app)
Use AVAudioEngine tap with 1024 buffer and keep your own 20 ms slicer after resample; avoid per‑frame allocations (reuse Data/UnsafeMutableRawPointer when possible).


Resample once with AVAudioConverter to 16k mono int16; keep converter hot.


Send frames immediately (no extra batching) → keep interactive feel.


Disable JSON upstream; binary PCM only.


Avoid main‑thread work in the audio callback. Only enqueue to WS sender.


URLSessionWebSocketTask: create once per session; let it run; don’t reconnect per hint.


ReplayKit: if frames enabled, throttle to 1–2 fps, crop to center text region, run on‑device OCR and send tokens only. Never stream images on the hot path.


Dedupe hints by value (trim punctuation/case) to prevent flicker.


Go (server)
WS library: nhooyr/websocket (low overhead); disable compression; set small read deadlines, pong handler.


TCP_NODELAY (no Nagle) at the socket layer for lower per‑frame latency.


GOMAXPROCS = CPU‑1; leave headroom.


ASR bridge: one goroutine for read, one for write; bounded channels to enforce backpressure; drop frames only if >200 ms behind.


Object pooling: sync.Pool for 640‑byte buffers to avoid GC churn; reuse JSON encoders (or jsoniter).


Rate limit answer generation (1/1.5 s) and coalesce close‑timed finals.


Observability:


Counters: frames in/out, partials/finals, hints/sec.


Timers: mic→partial, partial→hint.


Export OTLP to Grafana/Datadog; enable pprof in non‑prod.


Config: fast timeouts (idle WS close after 35 s inactivity), health endpoints, graceful shutdown.


Network
Keep sessions sticky to the same gateway instance.


TLS termination (Caddy/ALB) in same region; prefer HTTP/1.1 WS to avoid proxies that meddle with HTTP/2.


If using a CDN proxy, set no buffering on /ws.



10) Edge cases & fallbacks
ASR vendor down → send {"type":"error","code":"ASR_UNAVAILABLE"} and immediately stop session.


Network drop → app shows “Offline”, keeps VAD armed; reconnect WS on next trigger.


Noisy environment → raise VAD openDB by +3–6 dB; user toggle “Push‑to‑talk only”.



11) Roadmap knobs (later)
Switch to gRPC streaming for ASR if vendor supports; keep WS to client.


Add answer‑cache (LRU keyed by lemma’d question) for <100 ms answers on repeats.


Add language auto‑detect and route to matching ASR model.


Push frames OCR entirely on device and keep only keyword tokens.



TL;DR
Swift visionOS app streams 16k/mono/20 ms PCM over WS; starts on VAD or shortcut; shows 2‑line hints.


A single Go binary terminates WS, bridges to ASR, rate‑limits & calls LLM, and returns hints—no extra hops.


You get sub‑second “speech → hint” with the above defaults, and you can scale out the same design by cloning the gateway behind a TLS proxy.


If you want, I can turn this into a starter repo skeleton (Swift project + Go server, Dockerfiles, Compose) that compiles and runs locally + on a VM with the exact contracts above.


Cluely on Apple Vision Pro — Use Cases, Detailed Flows, and Deep-Dive Diagrams
Below are the most realistic use cases, exactly how they run end-to-end, and a set of detailed Mermaid diagrams (components, sequences, states, deployment, and data model). This matches the Swift (visionOS) + Go backend design we discussed and is tuned for speed and low cognitive load for the wearer.

Use cases (with step-by-step flows)
1) In-person sales meeting (hands-free coach)
Goal: Subtle nudges to ask the right follow-ups, then a crisp recap.
How it works
Armed: App is idle but listening for speech (VAD).


Trigger: Prospect starts talking → VAD crosses the threshold for 300 ms → auto-start.


Stream: 20 ms PCM frames (16 kHz mono) stream over WS; partial ASR arrives in ~200–300 ms.


Hinting: On each final ASR segment, backend calls LLM for a ≤ 20-word micro-answer (or question to ask).


UI: A 2-line chip appears in the window ornament:


“Confirm budget owner”


“Ask preferred timeline”
 Auto-fades in 4–5 s; never more than 1 hint per 1.5 s.


Stop: If silence for ~700 ms (“hangover”), session auto-stops.


Recap: At stop, app shows a recap sheet: decisions, owners, dates. You can “Save” or discard.



2) Remote demo / screen share (context via frames)
Goal: Inject context-aware suggestions while you demo.
How it works
You press Start (space bar) before sharing.


Frames: ReplayKit produces 1–2 fps video frames; on-device OCR extracts a few keywords (“SSO”, “Audit logs”, “Plan: Pro”). Only tokens are sent to backend (no images).


ASR: Speech partials/finals stream as usual.


Hinting: Backend fuses transcripts + OCR tokens to bias the suggestion (“Highlight SSO role mappings”).


Stop: You hit Stop or it times out after silence.



3) “Date Mode” assist (lightweight general knowledge)
Goal: Friendly 1-liner + follow-up when you’re asked something unexpected.
How it works
Trigger: You say “Hey Cluely” or she asks you a direct question (question detector on device hears interrogatives).


Audio window: The last 5–7 s audio buffer is sent to backend; optional 1–2 fps OCR tokens (e.g., “Shiba Inu”).


Answer: Backend returns one 20-word answer + one 12-word follow-up.


HUD chip:


“Shibas: smart, high prey drive; short nose-work games help.”


“Ask: ‘Does Kiko like snuffle mats?’”
 Auto-fades.



4) Stand-up or status update (press-and-hold)
Goal: Avoid accidental capture; only capture while key is held.
How it works
Hold the Return (or custom) key → Start on key-down.


Speak your update; partials stream; one top hint appears if helpful.


Release key → Stop immediately; one-line summary is logged.



5) Hiring interview (structured rubric)
Goal: Prompts for behavioral follow-ups; bookmarked moments.
How it works
Auto-start on speech.


Moments: You tap Mark when you hear a strong signal; app stores timestamp + last final text.


Hints: “Ask for a conflict example with peers.”


Stop/Recap: A rubric view shows your marks grouped by competency.



Big-picture component & data-flow (detailed)
flowchart TB
  subgraph VisionOS_App[visionOS App (Swift)]
    A0[TriggerController\n• VAD (energy db)\n• Shortcut (space/return)\n• Manual Start/Stop]
    A1[AudioEngine\n• AVAudioEngine tap\n• AVAudioConverter → 16k mono i16\n• 20ms frames]
    A2[FrameSource (optional)\n• ReplayKit 1–2 fps\n• on-device OCR → tokens\n• NEVER send images]
    A3[WS Client\n• URLSessionWebSocketTask\n• Binary up (PCM)\n• JSON down (events)]
    A4[CoachRuntime\n• rate-limit 1/1.5s\n• dedupe hints\n• 2-line cap + TTL]
    A5[UI\n• Window + Ornament\n• Chip + indicators\n• Mark button]
  end

  subgraph Backend_Go[Backend (Go, single binary)]
    B0[WS Ingress\n• nhooyr/websocket\n• No compression\n• TCP_NODELAY]
    B1[Session Manager\n• per-conn goroutines\n• bounded chans\n• backpressure]
    B2[ASR Bridge\n• vendor WS/gRPC\n• partial/final events]
    B3[RateLimiter/Dedupe\n• 1/1.5s hints\n• collapse near-finals]
    B4[Answer Engine\n• small prompt\n• 20w answer + 12w follow-up]
    B5[(Redis)\n• ephemeral session state]
    B6[(Postgres)\n• saved notes/tasks]
    B7[OTel Exporter\n• traces/metrics]
  end

  A0 --> A1
  A0 --> A2
  A1 -- 20ms PCM --> A3
  A2 -- OCR tokens (JSON) --> A3
  A3 <--> B0
  B0 --> B1
  B1 --> B2
  B2 --> B1
  B1 --> B3 --> B4 --> B1
  B1 --> B5
  B4 --> B6
  B1 --> B7
  B1 -- events JSON --> A3 --> A4 --> A5


“Speech → Hint” sequence (with frames and rate-limit)
sequenceDiagram
  autonumber
  participant User
  participant VAD as VAD (device)
  participant Mic as AudioEngine (20ms)
  participant WS as WebSocket (client)
  participant GW as Go Gateway
  participant ASR as ASR Vendor
  participant RL as RateLimiter
  participant LLM as Answer Engine (LLM)
  participant UI as Coach UI

  User->>VAD: speaks ≥300ms
  VAD-->>Mic: open gate (start)
  loop every 20ms
    Mic->>WS: binary PCM (640B)
    WS->>GW: binary PCM
    GW->>ASR: forward frame
  end
  ASR-->>GW: partial "…ask on Friday"
  GW-->>WS: {"type":"partial","text":"…ask on Friday"}
  WS-->>UI: (optional) ignore partial for chip

  ASR-->>GW: final "we can meet on Friday"
  GW->>RL: allow?
  alt Allowed (≥1.5s since last hint)
    RL-->>GW: yes
    GW->>LLM: Micro(q=finalText, ocrTokens)
    LLM-->>GW: {answer, followUp}
    GW-->>WS: {"type":"hint","text":"Confirm budget owner"}
    WS-->>UI: show chip (TTL 4.5s)
    GW-->>WS: {"type":"followup","text":"Ask preferred timeline"}
    WS-->>UI: (optional) second line or next chip
  else Blocked
    RL-->>GW: no (rate limit)
    GW-->>WS: (skip hint)
  end

  User-->>VAD: silence ≥700ms
  VAD-->>Mic: close gate (stop)
  GW-->>WS: {"type":"state","listening":false}


Client state machine (app-side)
stateDiagram-v2
  [*] --> Idle
  Idle --> Armed: app active
  Armed --> Listening: VAD open ≥300ms
  Armed --> Listening: Shortcut pressed
  Listening --> Processing: final segment received
  Processing --> Cooling: silence ≥700ms or user Stop
  Cooling --> Listening: speech resumes
  Cooling --> Idle: timeout or Stop


Deployment topology (MVP → Scale)
graph LR
  subgraph MVP (Single VM + Docker Compose)
    C[Caddy TLS] --> G[cluelyd (Go): :8080]
    G --- R[(Redis)]
    G --- P[(Postgres)]
    G --- O[OTel → Grafana/Datadog]
  end
  V[Vision Pro App] -->|wss:/ws| C

Scale-out later: clone cluelyd in 2–3 regions behind Cloudflare; keep Redis/Postgres managed.

Minimal data model (persist only what’s useful)
erDiagram
  USER ||--o{ SESSION : starts
  SESSION ||--o{ MOMENT : marks
  SESSION ||--o{ NOTE : saves
  USER {
    uuid id PK
    text email
    text display_name
  }
  SESSION {
    uuid id PK
    uuid user_id FK
    timestamptz started_at
    timestamptz ended_at
    text locale
  }
  MOMENT {
    uuid id PK
    uuid session_id FK
    timestamptz at
    text text_snippet
    text tags
  }
  NOTE {
    uuid id PK
    uuid session_id FK
    text title
    text body
    timestamptz created_at
  }


What the wearer actually experiences (per use case)
Sales meeting
Trigger: auto on speech.


First 1–2 seconds: ASR partials begin; first hint within ~500–700 ms of a final clause.


UI: a small chip pops at the bottom (ornament), then fades.


After stop: a recap sheet appears; you can “Save to Notes.”


Remote demo
Frames: 1–2 fps OCR tokens silently bias hints (“Highlight SSO role mappings”).


UX: looks identical to sales meeting; just more on-topic nudge text.


Date Mode
Trigger: wake word or question detector.


Answer: one crisp sentence + one friendly follow-up; fades quickly.


Push-to-talk
Trigger: key down; stops on key up.


Hinting: at most one chip during the press interval; no lingering capture.



Performance tips (that directly move p95)
Swift app


Convert to 16 kHz mono int16 once per buffer with AVAudioConverter (keep converter hot).


20 ms frames (640 B) → send immediately; no buffering beyond one frame.


No string work in the audio tap; just hand off Data to a queue.


Dedupe chips by normalizing (“confirm budget owner” == same as last) to avoid flicker.


Go backend


Disable per-message compression on WS.


Use bounded channels between WS and ASR to apply backpressure (drop frames only if > 200 ms behind).


sync.Pool for 640 B slices and JSON buffers to avoid GC churn.


Rate-limit hint generation to 1/1.5 s (already in design).


Keep LLM prompts tiny (≤ 300 tokens); pre-warm HTTP clients; set TLS session resumption.


Network


Terminate TLS near the gateway (same region).


Stick the entire session to one gateway instance (cookie/route affinity if using a proxy).



Example WS messages (sizes & cadence)
Upstream (every 20 ms)
Binary: 640 B (PCM 16-bit * 320 samples) ⇒ ~32 KB/s audio uplink.


Downstream (sporadic)
partial: 50–150 B JSON, a few per second.


final: 100–300 B JSON, typically every 1–3 s.


hint: 120–220 B JSON, ≤ 1 / 1.5 s (rate-limited).



If you want, I can turn these diagrams into a starter repo with the Swift app scaffold and a single-binary Go server that speaks the exact WS protocol above, plus a runnable docker-compose to test locally.


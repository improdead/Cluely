# Cluely Server (MVP)

This is a minimal Go backend that implements the WebSocket contract from `general_guide.md`. It now ships with a self-contained heuristic hint engine and leaves automatic speech recognition (ASR) unplugged so the client can supply transcripts directly.

What works now:
- WebSocket endpoint at ws://localhost:8080/ws
- Accepts control JSON messages and optional binary audio frames (PCM16 16kHz mono, 20ms)
- Generates hint/follow-up pairs using lightweight heuristics (no external APIs)
- Rate limiting (1 hint / 1.5s) and backpressure warnings when audio buffers overflow

Protocol (subset):
- Upstream (client → server)
  - {"type":"hello"}
  - {"type":"frame_meta","ocr":["token1","token2"]}
  - {"type":"stop"}
  - Binary: PCM16-LE, 16 kHz mono, 20ms frames (640 bytes)
  - {"type":"transcript","text":"...","final":true} ← primary input for hints
- Downstream (server → client)
  - {"type":"state","listening":false}
  - {"type":"hint","text":"Confirm budget owner","ttlMs":4500}
  - {"type":"followup","text":"Ask preferred timeline","ttlMs":4500}
  - {"type":"warning","code":"AUDIO_BACKPRESSURE","msg":"Audio quality degraded (dropping frames)."}

Configuration:
- No API keys are required. Hints rely on local heuristics.
- ASR is disabled unless you supply your own backend. Set `ASR_PROVIDER=stub` to exercise the no-op dropper, or leave it unset and stream transcripts over WebSocket.
- Optional tuning knobs remain for PCM buffer sizing, port, and metrics interval. See `.env.example` for details.

Observability:
- Server logs cover ASR wiring (if enabled), hint generation, and session events
- Basic metrics logged every 30s: active sessions, PCM frames (in, drop), ASR events, hints sent, errors

Quick start:
1. Copy `.env.example` to `.env` and adjust if needed.
2. Build the server:
   `go build -o bin/cluelyd ./cmd/cluelyd`
3. Run the server:
   `./bin/cluelyd`

Quick test with the included dev client:
- Build and run the test client which connects to ws://localhost:8080/ws, sends sample transcripts, and prints responses.
- `go build -o bin/wsdev ./cmd/wsdev`
- `./bin/wsdev`

Next steps (to reach the full guide):
- Wire in a real ASR backend if live audio transcription is required.
- Add TLS via Caddy and optional Redis/Postgres for session/recap persistence.
- Expand observability (OTel, pprof) and multi-region deployment.
- Replace visionOS client's Demo button with real VAD-driven audio streaming.

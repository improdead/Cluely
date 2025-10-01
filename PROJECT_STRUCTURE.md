# Cluely — Project Structure

Full MVP implementation following `general_guide.md` with liquid-glass UI and E2E audio → hint pipeline.

## Directory Structure

```
Cluely/
├── general_guide.md           # Original comprehensive design spec
├── implementation_status.md   # What's implemented vs. planned
├── agent.md                   # Gemini agent spec and prompting
├── QUICKSTART.md              # E2E setup and run guide
├── PROJECT_STRUCTURE.md       # This file
│
├── server/                    # Go backend (single binary)
│   ├── go.mod
│   ├── README.md
│   ├── cmd/
│   │   ├── cluelyd/
│   │   │   └── main.go        # Main server entrypoint
│   │   └── wsdev/
│   │       └── main.go        # Dev WS test client
│   └── internal/
│       ├── ws/
│       │   └── session.go     # WebSocket session handler
│       ├── asr/
│       │   └── client.go      # Google Cloud STT v2 bridge
│       ├── answer/
│       │   └── service.go     # Gemini answer engine
│       └── rt/
│           └── ratelimit.go   # Rate limiter utility
│
└── apps/
    └── visionos/
        └── CluelyApp/         # Swift visionOS client
            ├── App.swift
            ├── Config.swift
            ├── Views/
            │   ├── RootWindow.swift
            │   ├── CoachView.swift
            │   ├── OrnamentControls.swift
            │   ├── ListeningIndicator.swift
            │   ├── ContextCard.swift
            │   └── SuggestionPanel.swift
            ├── Input/
            │   └── VAD.swift
            ├── Session/
            │   ├── AudioEngine.swift
            │   ├── WSClient.swift
            │   ├── CoachRuntime.swift
            │   ├── SessionController.swift
            │   └── FrameSource.swift   # Placeholder
            └── Resources/
                └── Info.plist
```

## File Manifest

### Documentation
- **general_guide.md**: The original comprehensive design (visionOS + Go backend, ASR, LLM, deployment)
- **implementation_status.md**: Current state of implementation (what's done, half-done, future)
- **agent.md**: Gemini agent spec, prompt, constraints, and API usage
- **QUICKSTART.md**: Step-by-step guide to run the full E2E stack
- **PROJECT_STRUCTURE.md**: This file

### Backend (Go)
- **server/go.mod**: Module dependencies (chi, nhooyr/websocket, Google Cloud Speech v2, etc.)
- **server/README.md**: Server-specific setup and env var requirements
- **server/cmd/cluelyd/main.go**: Main HTTP/WS server on `:8080`
- **server/cmd/wsdev/main.go**: Simple WebSocket dev client for testing
- **server/internal/ws/session.go**: Per-connection session, WS read/write loop, ASR relay
- **server/internal/asr/client.go**: Google Cloud STT v2 streaming client with backpressure
- **server/internal/answer/service.go**: Gemini API integration for micro-answer + follow-up
- **server/internal/rt/ratelimit.go**: Simple rate limiter (1/1.5s)

### Frontend (visionOS/Swift)
- **apps/visionos/CluelyApp/App.swift**: Main app entry, SessionController in environment
- **apps/visionos/CluelyApp/Config.swift**: Reads WSURL from Info.plist
- **apps/visionos/CluelyApp/Views/RootWindow.swift**: Main window with glass UI layout
- **apps/visionos/CluelyApp/Views/CoachView.swift**: 2-line hint chip (bottom)
- **apps/visionos/CluelyApp/Views/OrnamentControls.swift**: Ornament with controls
- **apps/visionos/CluelyApp/Views/ListeningIndicator.swift**: Top-left "Listening..." pill
- **apps/visionos/CluelyApp/Views/ContextCard.swift**: Left glass card (for OCR/context)
- **apps/visionos/CluelyApp/Views/SuggestionPanel.swift**: Right glass panel for hints
- **apps/visionos/CluelyApp/Input/VAD.swift**: Energy-based voice activity detection
- **apps/visionos/CluelyApp/Session/AudioEngine.swift**: AVAudioEngine → 16k mono PCM + dB
- **apps/visionos/CluelyApp/Session/WSClient.swift**: WebSocket client (URLSession)
- **apps/visionos/CluelyApp/Session/CoachRuntime.swift**: Rate-limit and dedupe for hints
- **apps/visionos/CluelyApp/Session/SessionController.swift**: Main orchestrator (audio, WS, VAD, state)
- **apps/visionos/CluelyApp/Session/FrameSource.swift**: Placeholder for ReplayKit + OCR
- **apps/visionos/CluelyApp/Resources/Info.plist**: Bundle config, mic permission, WSURL

## Key Features Implemented

### Backend
- ✅ WebSocket server (nhooyr/websocket, no compression)
- ✅ Google Cloud Speech-to-Text v2 streaming with bounded channel backpressure
- ✅ Gemini (gemini-1.5-flash) for micro-answer + follow-up
- ✅ Rate limiting (1 hint / 1.5s per session)
- ✅ Graceful fallback if ASR or Gemini unavailable

### Frontend
- ✅ "Liquid glass" UI (ultraThinMaterial, white borders, transparent overlays)
- ✅ VAD-driven auto start/stop (energy-based, adjustable thresholds)
- ✅ Audio capture → 16 kHz mono PCM → WebSocket streaming
- ✅ Hint display in suggestion panel + coach chip
- ✅ Manual controls (Start/Stop/Mark/Frames toggle/Demo)
- ✅ Configurable server URL (Info.plist)

## What's Not Yet Implemented (Future)
- ReplayKit frame capture + on-device OCR tokenization
- Persistence (Redis/Postgres for session history, recap/notes)
- TLS/wss:// (Caddy or cloud proxy)
- Observability (OTel, pprof, metrics)
- Multi-region deployment, answer cache, language auto-detect
- Unit/integration tests and CI
- Polished UI animations (TTL fade, presenter-safe mode)

## How to Use This Project

1. **Read**: `QUICKSTART.md` for setup and run instructions.
2. **Check**: `implementation_status.md` to see what's implemented vs. planned.
3. **Reference**: `general_guide.md` for the full design spec.
4. **Agent**: `agent.md` for Gemini prompt and constraints.
5. **Build**: Follow server/README.md and QUICKSTART.md to run E2E.

## Contact / Next Steps

- For ASR setup, ensure you have a GCP project with Speech-to-Text API enabled and service account JSON.
- For Gemini, get an API key from Google AI Studio or Vertex AI.
- For visionOS development, you need macOS + Xcode 15+ and visionOS SDK.

Enjoy building Cluely! 🎯

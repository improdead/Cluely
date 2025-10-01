# Recent Changes & Enhancements

## 🎨 UI Overhaul — Liquid Glass Design

Completely redesigned the visionOS UI to match Vision Pro's design language with beautiful glass effects:

### New Components
- **ListeningIndicator**: Animated pill with pulsing blue gradient glow, appears top-left when VAD is active
- **ContextCard**: Left-side glass card with gradient borders and depth shadows for context/profile info
- **SuggestionPanel**: Right-side panel with rich glass effects, orange lightbulb icon, shows Gemini hints
- **FramePreviewCard**: Green glass card that appears when frame capture is enabled
- **CoachChip**: Bottom overlay chip with frosted glass for quick hints
- **OrnamentControls**: Enhanced bottom controls with icons (play/stop, bookmark, camera) and better glass styling

### Visual Enhancements
- **Multi-layer glass effects**: ultraThinMaterial + gradient overlays for depth
- **Gradient borders**: White-to-transparent gradients on all cards for that premium Vision Pro look
- **Depth shadows**: Subtle black/colored shadows with blur radius for floating effect
- **Smooth animations**: Spring animations for listening indicator, hint chips
- **Color accents**: Blue for listening, yellow/orange for suggestions, green for frame capture

## 🔧 Backend Enhancements

### Gemini Integration
- **Model**: Switched to `gemini-flash-lite-latest` for faster responses (configurable via `GEMINI_MODEL` env var)
- **Enhanced logging**: Detailed error logs for every Gemini API call, including status codes and response bodies
- **Fallback behavior**: Graceful degradation with static hints when API key is missing

### Observability & Metrics
- **New metrics module** (`internal/obs/metrics.go`):
  - Tracks active sessions, total sessions
  - PCM frames received
  - ASR partials and finals
  - Hints and follow-ups sent
  - ASR and answer errors
- **Auto-logging**: Metrics printed every 30 seconds to console
- **Error tracking**: Detailed logs for ASR initialization, Gemini requests, WS send errors

### Error Handling
- All critical paths now log errors with context (`[answer]`, `[session]`, `[asr]` prefixes)
- Non-blocking error handling (server continues if ASR/Gemini unavailable)
- HTTP status code checking for Gemini API responses

## 📹 ReplayKit + OCR Implementation

### FrameSource (visionOS)
- **ReplayKit integration**: Captures screen at 2 FPS when enabled
- **Vision framework OCR**: Runs `VNRecognizeTextRequest` with fast recognition level
- **Token extraction**: Sends top 10 OCR tokens to server via `frame_meta` JSON
- **Toggle control**: Frames toggle in ornament controls starts/stops capture

### SessionController Integration
- Automatically starts/stops FrameSource when `captureFrames` toggle changes
- Sends OCR tokens to server in real-time
- Visual feedback: FramePreviewCard appears when capture is active

## 🎯 VAD & Audio Pipeline

### AudioEngine Enhancements
- **dB metering**: Now emits audio level (dB) alongside PCM data
- **VAD integration**: Energy-based voice activity detection with configurable thresholds
- **Auto start/stop**: VAD triggers listening state automatically on speech detection
- **Manual override**: Start/Stop buttons still work for manual control

### SessionController Flow
- Audio engine runs continuously (VAD armed)
- When VAD detects speech → starts sending PCM to server
- After 700ms silence → stops sending automatically
- Clean state management with `sendAudio` gate flag

## 📊 Current Feature Status

### ✅ Fully Working
- Beautiful liquid glass UI with animations
- VAD-driven auto start/stop
- Google Cloud STT v2 streaming ASR
- Gemini micro-answer generation (flash-lite-latest)
- ReplayKit + Vision OCR (when enabled)
- Error logging and basic metrics
- Manual controls and demo mode

### 🔄 Next Steps (if needed)
- TLS/wss:// for production
- Redis/Postgres for session persistence
- OTel/pprof for production observability
- Docker Compose + Caddyfile for deployment
- Auto-fade animations for hint chips
- Presenter-safe mode (disable OCR in presentations)

## 🚀 How to Run

1. **Backend** (with all new features):
   ```bash
   cd server
   # Set environment variables:
   # GEMINI_API_KEY, GEMINI_MODEL (optional), GOOGLE_APPLICATION_CREDENTIALS, GCP_PROJECT_ID
   go build -o bin/cluelyd ./cmd/cluelyd
   ./bin/cluelyd
   # Watch for metrics logged every 30s
   ```

2. **Frontend** (visionOS):
   - Drag all Swift files into Xcode (including new View components)
   - Build and run in visionOS Simulator
   - Toggle "Frames" to enable ReplayKit+OCR
   - Speak to trigger VAD → see beautiful glass UI animate!

## 📝 Environment Variables

- `GEMINI_API_KEY`: Required for LLM hints
- `GEMINI_MODEL`: Optional, defaults to `gemini-flash-lite-latest`
- `GOOGLE_APPLICATION_CREDENTIALS`: Path to GCP service account JSON
- `GCP_PROJECT_ID`: Your GCP project ID
- `GCP_RECOGNIZER`: Optional, defaults to global recognizer

## 🎨 UI Files Added/Modified

New files:
- `CoachChip.swift`
- `FramePreviewCard.swift`

Enhanced with glass effects:
- `ListeningIndicator.swift`
- `ContextCard.swift`
- `SuggestionPanel.swift`
- `OrnamentControls.swift`
- `RootWindow.swift`

## 🔍 Observability Output Example

```
[metrics] sessions=1 pcm=15234 asr(p=42 f=8) hints=5 followups=3 errors(asr=0 ans=0)
[answer] success (JSON): Confirm budget owner
[session] send hint error: ...
```

Every log now has a prefix for easy filtering!

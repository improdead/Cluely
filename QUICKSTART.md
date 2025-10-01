# Cluely — Quick Start (E2E MVP)

This guide walks you through running the full Cluely stack: Go backend + visionOS client.

## What you have now
- **Backend (Go)**: WebSocket server with Google Cloud STT v2 + Gemini (flash-lite-latest) for micro-answers, with detailed error logs and metrics
- **Frontend (Swift/visionOS)**: Gorgeous "liquid glass" UI with VAD-driven auto start/stop, audio streaming, ReplayKit+OCR, and animated hint displays

## Prerequisites

### 1. Backend (Go server on Windows or Mac)
- Go 1.21+ installed
- GEMINI_API_KEY (Google AI Studio or Vertex AI)
- GEMINI_MODEL (optional, defaults to gemini-flash-lite-latest)
- Google Cloud credentials (Speech-to-Text v2):
  - A GCP project with Speech-to-Text API enabled
  - Service account JSON with appropriate permissions
  - Set env vars:
    - `GOOGLE_APPLICATION_CREDENTIALS` = path to service account JSON
    - `GCP_PROJECT_ID` = your GCP project ID
    - `GEMINI_API_KEY` = your Gemini API key

### 2. Frontend (visionOS client on macOS)
- macOS with Xcode 15+
- visionOS SDK installed
- Vision Pro device or visionOS Simulator

## Setup & Run

### Backend (Go server)

From the `Cluely` directory (Windows or macOS):

1. Set environment variables (Windows System Properties → Environment Variables or shell):
   ```powershell
   # Windows PowerShell example (set permanently via System Properties)
   $env:GEMINI_API_KEY = "your-key"
   $env:GOOGLE_APPLICATION_CREDENTIALS = "C:\path\to\service-account.json"
   $env:GCP_PROJECT_ID = "your-gcp-project"
   ```

2. Build and run:
   ```bash
   cd server
   go build -o bin/cluelyd ./cmd/cluelyd
   ./bin/cluelyd
   ```

   Server listens on `:8080`. If ASR initialization fails (missing GCP creds), it logs a warning and continues; you can test with the Demo button or transcript helper.

### Frontend (visionOS app)

On macOS:

1. Open Xcode and create a new visionOS App project named "Cluely".

2. Drag the contents of `apps/visionos/CluelyApp/` into your Xcode project (preserve folder structure).

3. Update `Info.plist` with your server URL:
   - If server is on the same Mac: `ws://localhost:8080/ws`
   - If server is on Windows PC: `ws://<YOUR_WINDOWS_IP>:8080/ws`

4. Build and run in the visionOS Simulator (or device).

5. Grant microphone permission when prompted.

## How to test E2E

### Option A: Real audio (VAD-driven)
- Let the app start with the audio engine running (VAD is armed).
- **Speak** near the mic; VAD should auto-detect and trigger "Listening..." pill.
- PCM streams to the server → Google STT → partial/final transcripts → Gemini hint → displayed in the suggestion panel.
- After ~700ms of silence, VAD should auto-stop.

### Option B: Manual start/stop
- Tap **Start** in the app to begin streaming.
- Speak → see partial/final + hints.
- Tap **Stop** to end.

### Option C: Demo button (fallback)
- If ASR is not working, tap **Demo** to send a test transcript to the server.
- You should see a hint appear in the suggestion panel.

## Troubleshooting

- **No ASR transcripts?**
  - Check server logs for ASR errors.
  - Verify `GOOGLE_APPLICATION_CREDENTIALS` and `GCP_PROJECT_ID` are correct.
  - Ensure Speech-to-Text API is enabled in your GCP project.

- **No hints?**
  - Check server logs for Gemini errors.
  - Verify `GEMINI_API_KEY` is set and valid.
  - If missing, server falls back to static hints.

- **Client can't connect?**
  - Ensure server is running on the expected host/port.
  - Check firewall rules (Windows Firewall may block `:8080`).
  - Update `Info.plist` WSURL to match server address.

- **VAD not triggering?**
  - Try adjusting VAD thresholds in `VAD.swift` (`openDB`, `closeDB`).
  - Verify microphone permission in Settings → Privacy & Security.

## What's next?

To reach full production parity per `general_guide.md`:
1. Add FrameSource + OCR tokens (ReplayKit + Vision framework).
2. Deploy behind Caddy with TLS for wss://.
3. Add Redis/Postgres for session persistence and recap/notes.
4. Add observability (OTel, pprof).
5. Polish UI (TTL fade animations, presenter-safe mode).

Enjoy your MVP! 🚀

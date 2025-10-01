# Cluely Server (MVP)

This is a minimal Go backend that implements the WebSocket contract from `general_guide.md`, with Gemini used for the micro‑answer generation and Google Cloud Speech-to-Text v2 for streaming ASR.

What works now:
- WebSocket endpoint at ws://localhost:8080/ws
- Accepts control JSON messages and binary audio frames (PCM16 16kHz mono, 20ms)
- Streams PCM to Google Cloud STT v2 and relays partial/final transcripts to client
- Generates hint/follow-up using Gemini for each final transcript
- Rate limiting (1 hint / 1.5s), backpressure (drop frames if >200ms behind)

Protocol (subset):
- Upstream (client → server)
  - {"type":"hello"}
  - {"type":"frame_meta","ocr":["token1","token2"]}
  - {"type":"stop"}
  - Binary: PCM16-LE, 16 kHz mono, 20ms frames (640 bytes)
  - {"type":"transcript","text":"...","final":true}   ← fallback if ASR unavailable
- Downstream (server → client)
  - {"type":"state","listening":false}
  - {"type":"partial","text":"..."}
  - {"type":"final","text":"..."}
  - {"type":"hint","text":"Confirm budget owner","ttlMs":4500}
  - {"type":"followup","text":"Ask preferred timeline","ttlMs":4500}

Requirements you'll need to provide:
- GEMINI_API_KEY: A Google AI Studio/Vertex API key for calling Gemini. Without it, the server returns a deterministic fallback hint.
- Google Cloud credentials for Speech-to-Text v2:
  - GOOGLE_APPLICATION_CREDENTIALS: Path to service account JSON (must have Speech-to-Text API enabled).
  - GCP_PROJECT_ID: Your GCP project ID.
  - GCP_RECOGNIZER (optional): defaults to projects/{project}/locations/global/recognizers/_
  - ASR_PCM_BUFFER (optional): size of the upstream PCM buffer (default 128). Raise this if you see AUDIO_BACKPRESSURE warnings.
  - If ASR fails to initialize (e.g., missing credentials), the server will log a warning and continue without ASR; you can use the transcript helper or Demo button to test.

Observability:
- Server logs detailed errors for ASR, Gemini, and session issues
- Basic metrics logged every 30s: active sessions, PCM frames (in, drop), ASR events, hints sent, errors
- Client receives a warning event when backpressure drops audio frames:
  {"type":"warning","code":"AUDIO_BACKPRESSURE","msg":"Audio quality degraded (dropping frames)."}
   - GEMINI_API_KEY
   - GOOGLE_APPLICATION_CREDENTIALS (path to your GCP service account JSON)
   - GCP_PROJECT_ID
3) Build and run:
   go build -o bin/cluelyd ./cmd/cluelyd
   ./bin/cluelyd

Quick test with the included dev client:
- Build and run the test client which connects to ws://localhost:8080/ws and sends a final transcript, then prints responses.
- go build -o bin/wsdev ./cmd/wsdev
- ./bin/wsdev

Next steps (to reach the full guide):
- Add TLS via Caddy and optional Redis/Postgres for session/recap persistence.
- Add observability (OTel, pprof) and multi-region deployment.
- Replace visionOS client's Demo button with real VAD-driven audio streaming.

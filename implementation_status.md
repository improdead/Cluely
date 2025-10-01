# Implementation Status

Scope follows general_guide.md (visionOS Swift app + single-binary Go backend).

Implemented
- Backend (Go)
  - /ws WebSocket endpoint (nhooyr/websocket), /healthz
  - Upstream JSON: hello, frame_meta (OCR tokens), stop, and fallback transcript helper
  - Downstream JSON: state, partial, final, hint, followup
  - ASR Bridge: Google Cloud Speech-to-Text v2 streaming with backpressure (drop frames if buffer full)
  - Answer Engine: Gemini (gemini-flash-lite-latest by default) via REST GenerateContent; JSON-first prompt; fallback if GEMINI_API_KEY missing
  - Hint rate limiting: 1 hint / 1.5s
  - Enhanced error logging: detailed logs for ASR, Gemini, and session errors
  - Basic observability: in-memory metrics logged every 30s (sessions, PCM frames, ASR events, hints, errors)
  - Dev WS client for local testing
- Frontend (visionOS scaffold in Swift)
  - **Liquid glass UI** inspired by Vision Pro design language:
    - Animated "Listening..." pill with pulsing blue gradient glow
    - Context card with glass material, gradient borders, and depth shadows
    - Suggestion panel with rich glass effects and orange accent
    - Frame capture indicator with green glass styling when enabled
    - Bottom coach chip with frosted glass overlay
    - Enhanced ornament controls with icons and better glass styling
  - WebSocket client wired to server; shows hints in multiple locations
  - Audio capture to 16 kHz mono int16 (20 ms frames) + dB metering for VAD
  - VAD (energy-based) wired to auto start/stop; manual Start/Stop also works
  - **ReplayKit + Vision OCR**: Captures screen at 2fps, runs fast OCR, sends tokens to server
  - Simple coach runtime (rate-limit + dedupe)
  - Configurable WSURL via Info.plist

Half-implemented (placeholders or simplified)
- Robust JSON parsing: ✅ Replaced with Codable structs in SessionController; server still uses simple maps internally (safe).
- Error handling & reconnect: basic; no offline state UI yet
- Security/TLS: Local ws:// only; production wss:// via Caddy/Cloud proxy is not set up here
- UI polish: chip TTL could auto-fade after 4.5s; presenter-safe mode not implemented
- Observability: basic metrics only; add OTel, pprof, distributed tracing later

Not implemented (future per guide)
- Persistence (Redis/Postgres) and recap/notes model
- Production observability (OTel, pprof, distributed tracing) and deployment (Caddy, Docker Compose)
- Multi-region, rate limiting at gateway, stickiness, and CDN/proxy tuning
- Answer cache and language auto-detect
- Unit/integration tests and CI
- Advanced UI: auto-fade hints, presenter-safe mode, accessibility features

How to reach full E2E parity
1) Add TLS and deploy behind Caddy; optional Redis/Postgres and recap persistence
2) Replace naive parsing with Codable and add error handling + reconnect
3) Add full observability (OTel, pprof, distributed tracing) and production deployment (Docker Compose, Caddyfile)
4) Polish UI with auto-fade animations, presenter-safe mode, and accessibility

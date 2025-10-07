# Code Issues Identified

## 1. Rate limiter ignores configured burst size
The `RateLimiter` struct exposes an `N` field, but `Allow()` only tracks a single `last` timestamp and never accounts for the requested burst capacity. Any value other than `1` is effectively ignored, so callers cannot allow multiple events within the window even if they ask for it. 【F:server/internal/rt/ratelimit.go†L8-L26】

### Remediation plan
1. Replace the single `last` timestamp with a token-bucket style counter that tracks how many events are currently available, replenishing `N` tokens every `Every` interval using `time.Since` and clamping to the configured burst size.
2. Adjust `Allow()` to decrement the bucket when a request is allowed and to deny when the bucket is empty, ensuring concurrent safety under the existing mutex.
3. Add unit tests that cover default (`N=1`) and multi-token scenarios, verifying refill timing and contention behaviour using a fake clock or controllable `time.Now` hook.
4. Audit call sites to confirm that any expectations about single-shot rate limiting still hold once bursts are respected; update documentation/comments accordingly.

## 2. Idle WebSocket sessions close after 35 seconds
`Session.run` wraps every `Read` call in a `context.WithTimeout(..., 35*time.Second)`. When the deadline expires (for example, during a quiet meeting), `Read` returns an error that is logged and the session is terminated. There is no ping/pong keep-alive to reset the timer, so long-lived connections are dropped unnecessarily. 【F:server/internal/ws/session.go†L93-L144】

### Remediation plan
1. Remove the read timeout from `Session.run` and instead rely on WebSocket-level ping/pong to detect dead peers; configure the connection with `EnableCompression` and `PingInterval` if supported by `nhooyr.io/websocket`.
2. Implement a periodic ping goroutine (e.g., every 25 seconds) that sends control frames and closes the connection only after a configurable grace period without `Pong` responses.
3. Update `pump()` and any other listeners to ensure downstream responses keep the connection alive and to handle `websocket.CloseStatus` properly.
4. Create integration or unit tests using a mock WebSocket connection to prove that idle sessions stay open while unresponsive clients still time out gracefully.

## 3. Broadcast OCR feed never marks first/last frames
The broadcast upload extension emits `frame_meta` messages without the `first` or `last` flags. As a result, the server never populates the "First frame" or "Last frame" buckets when this capture path is used, degrading the prompt context that relies on those markers. 【F:apps/visionos/BroadcastUpload/SampleHandler.swift†L33-L45】

### Remediation plan
1. Track whether the broadcast handler has emitted any OCR frames and annotate the first payload with `"first": true`; when `broadcastFinished` is invoked, flush a final `frame_meta` carrying the most recent tokens and `"last": true`.
2. Debounce OCR captures so that repeated frames do not incorrectly mark multiple `first` flags, potentially by caching a boolean `hasSentFirst` and the latest token batch.
3. Coordinate with the backend to ensure `frame_meta` messages bearing both flags are idempotent and do not require additional schema changes.
4. Add unit tests (using a small helper around `SampleHandler`) or UI tests to confirm the correct sequencing of flags during start/stop events.

## 4. VisionOS VAD ignores `minOpenMs`
`VADConfig` defines a `minOpenMs` threshold, but `SimpleVAD.onFrame` never uses it. Any audio sample above `openDB` immediately triggers a start event, even if it lasts only a few milliseconds, which increases false-positive activations from transient noise. 【F:apps/visionos/CluelyApp/Input/VAD.swift†L4-L24】

### Remediation plan
1. Extend `SimpleVAD` with a short-lived buffer that only transitions to the open state after the signal has remained above `openDB` for at least `minOpenMs`, accumulating frame durations to handle variable callback cadences.
2. Reset the pending-open buffer when the signal drops below `openDB` before reaching the minimum duration to avoid spurious openings.
3. Ensure `hangMs` logic continues to work by updating `lastSpeechAt` only once the VAD is truly open, and add a configuration-driven unit test suite that exercises various frame rates and noise bursts.
4. Document the expected frame duration assumptions in `VADConfig` so callers know how to tune `minOpenMs` relative to capture cadence.

## 5. Prompt normalisation strips important casing
`contextualTokens` lowercases every OCR token before deduplicating. The prompt therefore loses capitalization for proper nouns (e.g., "AWS", "CRM"), weakening grounding for the LLM. Preserving original casing while deduplicating would keep that information available. 【F:server/internal/answer/service.go†L223-L241】

### Remediation plan
1. Modify `contextualTokens` to track deduplication using a case-insensitive key (e.g., `strings.ToLower`) while storing the first-seen original token so the casing that reached the user is preserved.
2. Trim whitespace before deduplication but avoid lowercasing the emitted token; consider normalizing Unicode forms to prevent accidental duplicates.
3. Add tests that feed in mixed-case tokens and ensure the output retains canonical casing while still deduplicating case-insensitively.
4. Review downstream consumers (LLM prompts, logging) to confirm that case-sensitive tokens do not break existing assumptions.

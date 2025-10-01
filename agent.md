# Cluely Agent (Gemini) — Spec and Prompting

Goal
- Generate a concise on-glass micro-answer (≤ 20 words) and a follow-up (≤ 12 words) from final ASR segments, optionally biased by OCR tokens.

Inputs
- final transcript text (string)
- optional OCR tokens (array<string>)
- optional lightweight session context (future)

Outputs
- JSON only:
  { "answer": "...", "followUp": "..." }

System style and constraints
- Be crisp, supportive, and specific; avoid hedging and filler words
- Respect hard length caps; prefer nouns and actionable verbs
- Avoid sensitive content; defer when unsafe/unknown rather than inventing facts
- No markdown, no emojis, no extra prose — JSON only

API implementation (MVP)
- Model: gemini-1.5-flash (low latency)
- Endpoint: POST https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent?key=<GEMINI_API_KEY>
- Request body:
  - contents: one part with a single text prompt (see below)
  - generationConfig: temperature ~0.6, maxOutputTokens ≤ 128
- Parsing: prefer strict JSON response; fallback to splitting first two lines if plain text is returned

Prompt (JSON-first)
You are Cluely, a concise on-glass coach.
Given the final transcript below, produce exactly two helpful items: a concise micro-answer (<=20 words) and a follow-up (<=12 words).
Return strict JSON with keys 'answer' and 'followUp'.
Transcript: <FINAL_TEXT>
Context tokens: <TOK1, TOK2, ...> (omit if none)
Output JSON only.

Latency & quality controls
- Keep prompts short (≤ 300 tokens total)
- Rate limit requests to ≤ 1/1.5s per session (enforced in gateway)
- Dedupe by normalized text to reduce flicker

Error handling
- If the API fails or returns empty, use deterministic fallback suggestions
- Do not block the UI waiting for retries; prefer best-effort and move on

Safety
- Do not produce medical, legal, or explicit content
- If asked about risky or unknown topics, return a safe generic prompt to clarify context

Roadmap (later)
- Add answer cache (lemma-keyed) for repeated questions
- Add per-user domain knowledge/context windows with grounding
- Add multilingual support and tone adaptation

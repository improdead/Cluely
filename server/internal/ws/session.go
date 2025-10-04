package ws

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"cluely/server/internal/answer"
	"cluely/server/internal/asr"
	"cluely/server/internal/obs"
	"cluely/server/internal/rt"

	"nhooyr.io/websocket"
)

// Upstream message (client -> server) minimal schema
// Matches general_guide.md plus a "transcript" helper for MVP testing
// {"type":"hello"}
// {"type":"frame_meta","ocr":["token1","token2"]}
// {"type":"stop"}
// {"type":"transcript","text":"...","final":true}

type upMsg struct {
	Type  string   `json:"type"`
	Text  string   `json:"text,omitempty"`
	Final bool     `json:"final,omitempty"`
	OCR   []string `json:"ocr,omitempty"`
	First bool     `json:"first,omitempty"`
	Last  bool     `json:"last,omitempty"`
}

type Session struct {
	c            *websocket.Conn
	ans          *answer.Service
	asr          asr.Client
	ocrTokens    []string
	firstOCR     []string
	lastOCR      []string
	hints        *rt.RateLimiter
	mu           sync.Mutex
	listening    bool
	lastDropWarn time.Time
	closedOnce   sync.Once
}

func Handle(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		CompressionMode: websocket.CompressionDisabled,
	})
	if err != nil {
		return
	}
	log.Printf("ws client connected: %s", r.RemoteAddr)
	// Build ASR client (only if explicitly requested)
	var asrClient asr.Client
	switch provider := strings.ToLower(strings.TrimSpace(os.Getenv("ASR_PROVIDER"))); provider {
	case "", "disabled", "none":
		log.Printf("asr disabled via ASR_PROVIDER=%s", provider)
	case "stub":
		asrClient, err = asr.New("stub")
		if err != nil {
			log.Printf("asr init failed (fallback to transcript helper): %v", err)
		}
	case "gemini":
		asrClient, err = asr.New("gemini")
		if err != nil {
			log.Printf("asr init failed (fallback to transcript helper): %v", err)
		}
	default:
		log.Printf("asr provider %q not supported; disabling ASR", provider)
	}
	// Build session
	s := &Session{
		c:         c,
		ans:       answer.NewServiceFromEnv(),
		asr:       asrClient,
		hints:     rt.NewRateLimiter(1, 1500*time.Millisecond),
		listening: false,
	}
	// Send initial state
	_ = s.sendJSON(map[string]any{"type": "state", "listening": s.listening})
	obs.IncSessionActive()
	// Start main loop
	s.run()
}

func (s *Session) run() {
	ctx := context.Background()
	s.c.SetReadLimit(1 << 20) // 1MB
	defer func() {
		if s.asr != nil {
			s.asr.Flush()
			s.asr.Close()
		}
		obs.DecSessionActive()
		s.close(websocket.StatusNormalClosure, "bye")
	}()

	// Spawn goroutine to relay ASR events -> client
	if s.asr != nil {
		go s.relayASR()
	}

	for {
		readCtx, cancel := context.WithTimeout(ctx, 35*time.Second)
		typ, data, err := s.c.Read(readCtx)
		cancel()
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure || websocket.CloseStatus(err) == websocket.StatusGoingAway {
				return
			}
			log.Printf("ws read error: %v", err)
			return
		}
		switch typ {
		case websocket.MessageBinary:
			// Forward PCM to ASR if available
			obs.IncPCMFrame()
			if s.asr != nil {
				ok := s.asr.WritePCM(data)
				if !ok {
					// Rate-limit warnings to once every 2s
					if time.Since(s.lastDropWarn) > 2*time.Second {
						s.lastDropWarn = time.Now()
						_ = s.sendJSON(map[string]any{
							"type": "warning",
							"code": "AUDIO_BACKPRESSURE",
							"msg":  "Audio quality degraded (dropping frames).",
						})
					}
				}
			}
		case websocket.MessageText:
			if err := s.handleText(data); err != nil {
				log.Printf("handleText: %v", err)
			}
		}
	}
}

func (s *Session) relayASR() {
	for ev := range s.asr.Events() {
		// Track metrics
		if ev.IsFinal {
			obs.IncASRFinal()
		} else {
			obs.IncASRPartial()
		}
		// Send partial/final to client
		if err := s.sendJSON(map[string]any{"type": ev.Type, "text": ev.Text}); err != nil {
			log.Printf("[session] send %s error: %v", ev.Type, err)
		}
		// On final, generate and stream hint if rate-limit allows
		if ev.IsFinal && s.hints.Allow() {
			ocr, first, last := s.snapshotOCRContext()
			if ans := s.ans.Micro(ev.Text, ocr, first, last); ans != nil {
				s.streamAnswer(ans)
			} else {
				obs.IncErrorAnswer()
			}
		}
	}
}

func (s *Session) handleText(data []byte) error {
	log.Printf("[session] received: %s", string(data))
	var m upMsg
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	switch strings.ToLower(m.Type) {
	case "hello":
		return s.sendJSON(map[string]any{"type": "state", "listening": s.listening})
	case "frame_meta":
		s.mu.Lock()
		s.ocrTokens = m.OCR
		if m.First {
			s.firstOCR = m.OCR
		}
		if m.Last {
			s.lastOCR = m.OCR
		}
		s.mu.Unlock()
		return nil
	case "stop":
		s.setListening(false)
		if s.asr != nil {
			s.asr.Flush()
		}
		return s.sendJSON(map[string]any{"type": "state", "listening": s.listening})
	case "transcript":
		if m.Text == "" {
			return errors.New("empty transcript.text")
		}
		// Echo a final/partial to match contract
		kind := "partial"
		if m.Final {
			kind = "final"
		}
		if err := s.sendJSON(map[string]any{"type": kind, "text": m.Text}); err != nil {
			return err
		}
		if m.Final && s.hints.Allow() {
			ocr, first, last := s.snapshotOCRContext()
			if ans := s.ans.Micro(m.Text, ocr, first, last); ans != nil {
				s.streamAnswer(ans)
			}
		}
		return nil
	default:
		return nil
	}
}

func (s *Session) snapshotOCRContext() (ocr, first, last []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ocr = make([]string, len(s.ocrTokens))
	copy(ocr, s.ocrTokens)
	first = make([]string, len(s.firstOCR))
	copy(first, s.firstOCR)
	last = make([]string, len(s.lastOCR))
	copy(last, s.lastOCR)
	return
}

func (s *Session) setListening(v bool) { s.mu.Lock(); s.listening = v; s.mu.Unlock() }

func (s *Session) sendJSON(v any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	b, _ := json.Marshal(v)
	log.Printf("[session] sending: %s", string(b))
	return s.c.Write(ctx, websocket.MessageText, b)
}

func (s *Session) close(code websocket.StatusCode, reason string) {
	s.closedOnce.Do(func() { _ = s.c.Close(code, reason) })
}

// streamAnswer sends token-by-token partial updates to the client for a smoother UI,
// then emits the final hint/followup events for stability.
func (s *Session) streamAnswer(ans *answer.Answer) {
	go func() {
		// stream answer tokens
		if strings.TrimSpace(ans.Answer) != "" {
			var partial string
			toks := strings.Fields(ans.Answer)
			for i, t := range toks {
				if i > 0 {
					partial += " "
				}
				partial += t
				_ = s.sendJSON(map[string]any{"type": "hint_partial", "text": partial})
				time.Sleep(50 * time.Millisecond)
			}
			obs.IncHint()
			_ = s.sendJSON(map[string]any{"type": "hint", "text": ans.Answer, "ttlMs": 4500})
		}
		// stream follow-up tokens
		if strings.TrimSpace(ans.FollowUp) != "" {
			var partial string
			toks := strings.Fields(ans.FollowUp)
			for i, t := range toks {
				if i > 0 {
					partial += " "
				}
				partial += t
				_ = s.sendJSON(map[string]any{"type": "followup_partial", "text": partial})
				time.Sleep(50 * time.Millisecond)
			}
			obs.IncFollowup()
			_ = s.sendJSON(map[string]any{"type": "followup", "text": ans.FollowUp, "ttlMs": 4500})
		}
	}()
}

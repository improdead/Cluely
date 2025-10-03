package answer

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Answer struct {
	Answer     string  `json:"answer"`
	FollowUp   string  `json:"followUp"`
	Confidence float64 `json:"confidence,omitempty"`
}

type Service struct {
	apiKey string
	model  string
	http   *http.Client
}

func NewServiceFromEnv() *Service {
	key := os.Getenv("GEMINI_API_KEY")
	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-flash-lite-latest"
	}
	return &Service{
		apiKey: key,
		model:  model,
		http:   &http.Client{Timeout: 4 * time.Second},
	}
}

// Micro produces a concise two-line hint and follow-up using Gemini.
// If no API key is present, returns a deterministic fallback.
// firstOCR and lastOCR are optional OCR tokens from start/end of speech segment.
func (s *Service) Micro(text string, ocr []string, firstOCR []string, lastOCR []string) *Answer {
	if strings.TrimSpace(text) == "" {
		log.Println("[answer] empty text, skipping")
		return nil
	}
	if s.apiKey == "" {
		log.Println("[answer] GEMINI_API_KEY missing, using fallback")
		return &Answer{Answer: "Confirm budget owner", FollowUp: "Ask preferred timeline", Confidence: 0.5}
	}

	prompt := buildPrompt(text, ocr, firstOCR, lastOCR)
	endpoint := "https://generativelanguage.googleapis.com/v1beta/models/" + s.model + ":generateContent?key=" + s.apiKey
	body := map[string]any{
		"contents": []any{
			map[string]any{
				"parts": []any{map[string]any{"text": prompt}},
			},
		},
		"generationConfig": map[string]any{
			"temperature":     0.6,
			"maxOutputTokens": 128,
		},
	}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", endpoint, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.http.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	var gr geminiResp
	if err := json.NewDecoder(resp.Body).Decode(&gr); err != nil {
		return nil
	}
	textOut := gr.FirstText()
	if textOut == "" {
		return nil
	}
	// Attempt strict JSON first
	var a Answer
	if json.Unmarshal([]byte(textOut), &a) == nil && a.Answer != "" {
		return &a
	}
	// Fallback: split lines
	lines := strings.Split(strings.TrimSpace(textOut), "\n")
	ans := strings.TrimSpace(lines[0])
	follow := ""
	if len(lines) > 1 {
		follow = strings.TrimSpace(lines[1])
	}
	return &Answer{Answer: ans, FollowUp: follow, Confidence: 0.7}
}

type geminiResp struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func (r geminiResp) FirstText() string {
	if len(r.Candidates) == 0 || len(r.Candidates[0].Content.Parts) == 0 {
		return ""
	}
	return r.Candidates[0].Content.Parts[0].Text
}

func buildPrompt(text string, ocr []string, firstOCR []string, lastOCR []string) string {
	// Build rich context from OCR tokens
	var contextParts []string
	if len(firstOCR) > 0 {
		contextParts = append(contextParts, "Screen at start: "+strings.Join(firstOCR, ", "))
	}
	if len(lastOCR) > 0 && !equalSlices(firstOCR, lastOCR) {
		contextParts = append(contextParts, "Screen at end: "+strings.Join(lastOCR, ", "))
	}
	if len(ocr) > 0 && len(firstOCR) == 0 && len(lastOCR) == 0 {
		contextParts = append(contextParts, "Visible text: "+strings.Join(ocr, ", "))
	}
	visualContext := ""
	if len(contextParts) > 0 {
		visualContext = "\n\nVisual Context:\n" + strings.Join(contextParts, "\n")
	}

	return `You are Cluely, an elite executive coach powered by real-time context.

Your Role:
- You provide instant, actionable micro-coaching during live conversations, meetings, presentations, and interviews
- You see what the user sees (via screen OCR) and hear what they hear (via transcription)
- Your guidance helps users navigate critical moments with confidence and strategic insight

Your Coaching Principles:
1. BE SPECIFIC: Reference actual content from the transcript and visual context
2. BE ACTIONABLE: Suggest concrete next steps, not generic advice
3. BE STRATEGIC: Think like an executive coach—identify opportunities, risks, and power dynamics
4. BE CONCISE: Deliver maximum insight in minimum words
5. BE CONTEXT-AWARE: Use screen content to understand the situation deeply

Output Format:
{
  "answer": "<Your primary insight or suggestion (≤20 words)>",
  "followUp": "<Your tactical next step or clarifying question (≤12 words)>"
}

Examples:

Transcript: "So our Q4 revenue was down 8% but user engagement is up"
Screen: "Q4 Financial Summary, Revenue: $2.1M, MAU: 45K"
{
  "answer": "Pivot to engagement growth story—higher LTV potential",
  "followUp": "Ask: 'What's driving the MAU increase?'"
}

Transcript: "I'm not sure I understand the technical architecture you're proposing"
Screen: "System Design Diagram, Microservices, API Gateway"
{
  "answer": "They're confused—simplify the explanation now",
  "followUp": "Use analogy: 'Like separate apps talking via messages'"
}

Transcript: "We can probably meet Friday to discuss the terms"
{
  "answer": "'Probably' signals low commitment—confirm availability now",
  "followUp": "Ask: 'Friday 2pm works—shall I send invite?'"
}

Now, analyze this interaction:

Transcript: "` + text + `"` + visualContext + `

Provide your coaching in strict JSON format (keys: "answer", "followUp"):`
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

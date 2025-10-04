package answer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	apiKey  string
	model   string
	client  *http.Client
	baseURL string
}

const (
	geminiBaseURL  = "https://generativelanguage.googleapis.com/v1beta"
	defaultGemini  = "gemini-1.5-flash"
	requestTimeout = 8 * time.Second
)

func NewServiceFromEnv() *Service {
	apiKey := strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
	model := strings.TrimSpace(os.Getenv("GEMINI_MODEL"))
	base := strings.TrimSpace(os.Getenv("GEMINI_BASE_URL"))
	if model == "" {
		model = defaultGemini
	}
	if base == "" {
		base = geminiBaseURL
	}
	return &Service{
		apiKey:  apiKey,
		model:   model,
		client:  &http.Client{Timeout: requestTimeout},
		baseURL: base,
	}
}

func (s *Service) Micro(text string, ocr []string, firstOCR []string, lastOCR []string) *Answer {
	transcript := strings.TrimSpace(text)
	if transcript == "" {
		log.Println("[answer] empty text, skipping")
		return nil
	}
	if s.apiKey == "" {
		log.Println("[answer] GEMINI_API_KEY is not set; cannot generate hint")
		return nil
	}

	prompt := buildPrompt(transcript, ocr, firstOCR, lastOCR)
	ans, err := s.callGemini(prompt)
	if err != nil {
		log.Printf("[answer] gemini request failed: %v", err)
		return nil
	}
	return ans
}

func buildPrompt(transcript string, ocr []string, first []string, last []string) string {
	uniqTokens := contextualTokens(ocr, first, last)
	var sb strings.Builder

	// Core identity: who you are and what to do
	sb.WriteString("<core_identity> You are Cluely, a real-time on-glass sales coach created by Cluely. Your sole purpose is to analyze the conversation and what's on the screen, then deliver exactly one tactical coaching hint and one crisp follow-up question that advances the deal. Be specific, accurate, and immediately actionable. </core_identity> ")

	// Hard rules and safety constraints
	sb.WriteString("<rules> NEVER use meta-phrases or pleasantries. NEVER reveal or mention models/providers. NEVER mention 'screenshot' or 'image'—say 'the screen' if needed. NEVER summarize the transcript unless explicitly asked. DO NOT add explanations, markdown, code fences, or keys beyond answer and followUp. Do not invent names, figures, or commitments. Avoid double quotes inside values to keep JSON valid; paraphrase instead. If uncertain, state that briefly and ask the minimum clarifier. </rules> ")

	// How to interpret context and adapt coaching
	sb.WriteString("<context_model> Focus on meaning over keywords. Weight OCR by recency: 'Last frame tokens' = what's visible now (highest signal); 'First frame tokens' = persistent session context; 'Unique context tokens' = disambiguation. Infer stage: discovery (goals, pain, why-now), evaluation (architecture, pilot, metrics), negotiation (pricing, budget, procurement, legal). Adapt: discovery -> clarify outcome + next step; evaluation -> tie feature to their outcome and propose pilot/measure; negotiation -> surface blockers, decision path/owners, and close timeline. </context_model> ")

	// Output contract
	sb.WriteString("<output_contract> Return EXACTLY one compact JSON object only: {\"answer\":\"<=22 words, directive, empathetic, concrete\",\"followUp\":\"<=16 words, one open-ended question\"}. No newlines, no extra whitespace, no code fences, no other keys. </output_contract> ")

	// Quality bar and examples
	sb.WriteString("<quality> The answer proposes one next move (e.g., anchor ROI to budget owner, confirm risk mitigation, propose time-bound step). The followUp asks one specific question that progresses approval, scope, or timeline. Examples—answer: 'Tie uptime risk to your SLOs; propose a 2-week pilot'. followUp: 'Who owns final approval on this?'. </quality> ")

	// Inject live context
	sb.WriteString("Transcript:\n")
	sb.WriteString(transcript)
	sb.WriteString("\n\nRecent OCR tokens: ")
	if len(ocr) == 0 {
		sb.WriteString("none")
	} else {
		sb.WriteString(strings.Join(ocr, ", "))
	}
	sb.WriteString("\nFirst frame tokens: ")
	if len(first) == 0 {
		sb.WriteString("none")
	} else {
		sb.WriteString(strings.Join(first, ", "))
	}
	sb.WriteString("\nLast frame tokens: ")
	if len(last) == 0 {
		sb.WriteString("none")
	} else {
		sb.WriteString(strings.Join(last, ", "))
	}
	sb.WriteString("\nUnique context tokens: ")
	if len(uniqTokens) == 0 {
		sb.WriteString("none")
	} else {
		sb.WriteString(strings.Join(uniqTokens, ", "))
	}
	return sb.String()
}

func (s *Service) callGemini(prompt string) (*Answer, error) {
	requestPayload := geminiRequest{
		Contents: []geminiContent{
			{
				Role:  "user",
				Parts: []geminiPart{{Text: prompt}},
			},
		},
		GenerationConfig: &geminiGenerationConfig{
			Temperature:     0.7,
			TopP:            0.95,
			TopK:            32,
			MaxOutputTokens: 120,
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(requestPayload); err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", strings.TrimRight(s.baseURL, "/"), s.model, s.apiKey)
	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, parseGeminiError(resp)
	}

	var genResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	candidateText := extractCandidateText(genResp.Candidates)
	if candidateText == "" {
		return nil, errors.New("gemini returned empty candidate text")
	}
	candidateText = trimCodeFence(candidateText)

	var ans Answer
	if err := json.Unmarshal([]byte(candidateText), &ans); err != nil {
		return nil, fmt.Errorf("unmarshal candidate: %w", err)
	}
	if ans.Answer == "" && ans.FollowUp == "" {
		return nil, errors.New("gemini returned empty payload")
	}
	if ans.Confidence == 0 {
		ans.Confidence = 0.8
	}
	return &ans, nil
}

func parseGeminiError(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("gemini http %d: failed to read error body: %w", resp.StatusCode, err)
	}
	var apiErr geminiError
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Error.Message != "" {
		return fmt.Errorf("gemini http %d: %s", resp.StatusCode, apiErr.Error.Message)
	}
	return fmt.Errorf("gemini http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
}

func extractCandidateText(candidates []geminiCandidate) string {
	for _, c := range candidates {
		for _, part := range c.Content.Parts {
			if t := strings.TrimSpace(part.Text); t != "" {
				return t
			}
		}
	}
	return ""
}

func trimCodeFence(text string) string {
	trimmed := strings.TrimSpace(text)
	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```")
		trimmed = strings.TrimSpace(trimmed)
		if strings.HasPrefix(strings.ToLower(trimmed), "json") {
			trimmed = strings.TrimSpace(trimmed[4:])
		}
		if idx := strings.LastIndex(trimmed, "```"); idx != -1 {
			trimmed = trimmed[:idx]
		}
		trimmed = strings.TrimSpace(trimmed)
	}
	return trimmed
}

func contextualTokens(ocr []string, firstOCR []string, lastOCR []string) []string {
	merged := append([]string{}, ocr...)
	merged = append(merged, firstOCR...)
	merged = append(merged, lastOCR...)

	seen := make(map[string]struct{}, len(merged))
	uniq := make([]string, 0, len(merged))
	for _, token := range merged {
		normalized := strings.TrimSpace(strings.ToLower(token))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		uniq = append(uniq, normalized)
	}
	return uniq
}

type geminiRequest struct {
	Contents         []geminiContent         `json:"contents"`
	GenerationConfig *geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text       string            `json:"text,omitempty"`
	InlineData *geminiInlineData `json:"inlineData,omitempty"`
}

type geminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type geminiGenerationConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	TopP            float64 `json:"topP,omitempty"`
	TopK            int     `json:"topK,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}

type geminiResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
}

type geminiCandidate struct {
	Content struct {
		Parts []geminiPart `json:"parts"`
	} `json:"content"`
}

type geminiError struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

package asr

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type geminiConfig struct {
	APIKey  string
	Model   string
	BaseURL string
	Timeout time.Duration
}

type geminiClient struct {
	cfg       geminiConfig
	http      *http.Client
	events    chan Event
	mu        sync.Mutex
	buf       []byte
	closeOnce sync.Once
	closed    bool
	wg        sync.WaitGroup
}

func newGeminiClient(cfg geminiConfig) (Client, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("gemini ASR requires API key")
	}
	if cfg.Model == "" {
		cfg.Model = defaultGeminiASRModel
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultGeminiBaseURL
	}
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	if cfg.Timeout == 0 {
		cfg.Timeout = 12 * time.Second
	}
	client := &geminiClient{
		cfg:    cfg,
		http:   &http.Client{Timeout: cfg.Timeout},
		events: make(chan Event, 4),
	}
	return client, nil
}

func (c *geminiClient) WritePCM(data []byte) bool {
	if len(data) == 0 {
		return true
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return false
	}
	c.buf = append(c.buf, data...)
	return true
}

func (c *geminiClient) Events() <-chan Event {
	return c.events
}

func (c *geminiClient) Dropped() int64 {
	return 0
}

func (c *geminiClient) Flush() {
	c.mu.Lock()
	if c.closed || len(c.buf) == 0 {
		c.mu.Unlock()
		return
	}
	audio := make([]byte, len(c.buf))
	copy(audio, c.buf)
	c.buf = c.buf[:0]
	c.mu.Unlock()

	c.launchTranscription(audio, true)
}

func (c *geminiClient) Close() {
	c.closeOnce.Do(func() {
		c.mu.Lock()
		c.closed = true
		audio := make([]byte, len(c.buf))
		copy(audio, c.buf)
		c.buf = nil
		c.mu.Unlock()
		if len(audio) > 0 {
			c.launchTranscription(audio, true)
		}
		c.wg.Wait()
		close(c.events)
	})
}

func (c *geminiClient) launchTranscription(audio []byte, final bool) {
	if len(audio) == 0 {
		return
	}
	c.wg.Add(1)
	go func(buf []byte, isFinal bool) {
		defer c.wg.Done()
		if err := c.streamTranscribe(buf, isFinal); err != nil {
			log.Printf("[asr][gemini] transcribe error: %v", err)
		}
	}(audio, final)
}

func (c *geminiClient) emitEvent(evt Event) {
	select {
	case c.events <- evt:
	default:
		log.Printf("[asr][gemini] dropping %s event (channel full)", evt.Type)
	}
}

func (c *geminiClient) streamTranscribe(audio []byte, expectFinal bool) error {
	inline := base64.StdEncoding.EncodeToString(audio)

	payload := geminiASRRequest{
		Contents: []geminiASRContent{
			{
				Role: "user",
				Parts: []geminiASRPart{
					{Text: "Transcribe the provided audio into clear English text. Return only the transcript."},
					{InlineData: &geminiInlineData{MimeType: "audio/pcm;rate=16000", Data: inline}},
				},
			},
		},
		GenerationConfig: &geminiGenerationConfig{
			Temperature:     0.1,
			TopP:            0.9,
			MaxOutputTokens: 256,
		},
	}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return fmt.Errorf("encode request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:streamGenerateContent?key=%s&alt=sse", c.cfg.BaseURL, c.cfg.Model, c.cfg.APIKey)
	req, err := http.NewRequest(http.MethodPost, url, &body)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return c.parseError(resp)
	}

	scanner := bufio.NewScanner(resp.Body)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1<<20)

	var (
		eventData     strings.Builder
		lastText      string
		emittedFinal  bool
		streamClosing bool
	)

	flushEvent := func(data string) {
		if data == "" {
			return
		}
		if data == "[DONE]" {
			streamClosing = true
			return
		}
		var chunk geminiASRResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			log.Printf("[asr][gemini] failed to parse stream chunk: %v", err)
			return
		}
		text := extractCandidateText(chunk.Candidates)
		if text == "" {
			return
		}

		isChunkFinal := candidateFinished(chunk.Candidates)

		if isChunkFinal && expectFinal {
			if text != lastText {
				c.emitEvent(Event{Type: "partial", Text: text, IsFinal: false})
				lastText = text
			}
			c.emitEvent(Event{Type: "final", Text: text, IsFinal: true})
			emittedFinal = true
			return
		}

		if text != lastText {
			c.emitEvent(Event{Type: "partial", Text: text, IsFinal: false})
			lastText = text
		}
	}

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data:") {
			if eventData.Len() > 0 {
				eventData.WriteByte('\n')
			}
			eventData.WriteString(strings.TrimSpace(strings.TrimPrefix(line, "data:")))
			continue
		}
		if strings.TrimSpace(line) == "" {
			flushEvent(strings.TrimSpace(eventData.String()))
			eventData.Reset()
		}
		if streamClosing {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read stream: %w", err)
	}

	if eventData.Len() > 0 {
		flushEvent(strings.TrimSpace(eventData.String()))
	}

	if expectFinal && !emittedFinal && lastText != "" {
		c.emitEvent(Event{Type: "final", Text: lastText, IsFinal: true})
		emittedFinal = true
	}

	if lastText == "" {
		return fmt.Errorf("no transcription returned")
	}

	return nil
}

func (c *geminiClient) parseError(resp *http.Response) error {
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("gemini http %d: failed to read error body: %w", resp.StatusCode, err)
	}
	var apiErr geminiError
	if err := json.Unmarshal(data, &apiErr); err == nil && apiErr.Error.Message != "" {
		return fmt.Errorf("gemini http %d: %s", resp.StatusCode, apiErr.Error.Message)
	}
	return fmt.Errorf("gemini http %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
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

func candidateFinished(candidates []geminiCandidate) bool {
	for _, c := range candidates {
		if reason := strings.TrimSpace(strings.ToUpper(c.FinishReason)); reason != "" && reason != "INCOMPLETE" {
			return true
		}
	}
	return false
}

type geminiASRRequest struct {
	Contents         []geminiASRContent      `json:"contents"`
	GenerationConfig *geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiASRContent struct {
	Role  string          `json:"role,omitempty"`
	Parts []geminiASRPart `json:"parts"`
}

type geminiASRPart struct {
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

type geminiASRResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
}

type geminiCandidate struct {
	Content struct {
		Parts []geminiASRPart `json:"parts"`
	} `json:"content"`
	FinishReason string `json:"finishReason"`
}

type geminiError struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

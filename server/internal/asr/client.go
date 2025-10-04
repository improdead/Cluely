package asr

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"cluely/server/internal/obs"
)

type Event struct {
	Type    string
	Text    string
	IsFinal bool
}

type Client interface {
	WritePCM([]byte) bool
	Events() <-chan Event
	Dropped() int64
	Flush()
	Close()
}

func New(provider string) (Client, error) {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "", "none", "disabled":
		return nil, nil
	case "stub":
		return newStubClient(), nil
	case "gemini":
		return newGeminiClientFromEnv()
	default:
		return nil, fmt.Errorf("asr provider %q not supported", provider)
	}
}

type stubClient struct {
	events    chan Event
	dropped   int64
	closeOnce sync.Once
}

func newStubClient() Client {
	ch := make(chan Event)
	close(ch)
	return &stubClient{events: ch}
}

func (c *stubClient) WritePCM(_ []byte) bool {
	atomic.AddInt64(&c.dropped, 1)
	obs.IncPCMFrameDrop()
	return false
}

func (c *stubClient) Events() <-chan Event { return c.events }

func (c *stubClient) Dropped() int64 { return atomic.LoadInt64(&c.dropped) }

func (c *stubClient) Flush() {}

func (c *stubClient) Close() {
	c.closeOnce.Do(func() {})
}

const (
	defaultGeminiASRModel = "gemini-1.5-flash"
	defaultGeminiBaseURL  = "https://generativelanguage.googleapis.com/v1beta"
)

func newGeminiClientFromEnv() (Client, error) {
	apiKey := strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
	if apiKey == "" {
		return nil, errors.New("GEMINI_API_KEY is required for ASR provider gemini")
	}

	model := firstNonEmpty(
		os.Getenv("GEMINI_ASR_MODEL"),
		os.Getenv("GEMINI_MODEL"),
		defaultGeminiASRModel,
	)
	base := firstNonEmpty(
		os.Getenv("GEMINI_ASR_BASE_URL"),
		os.Getenv("GEMINI_BASE_URL"),
		defaultGeminiBaseURL,
	)

	cfg := geminiConfig{
		APIKey:  strings.TrimSpace(apiKey),
		Model:   strings.TrimSpace(model),
		BaseURL: strings.TrimRight(strings.TrimSpace(base), "/"),
		Timeout: 12 * time.Second,
	}
	return newGeminiClient(cfg)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

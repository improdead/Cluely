package asr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type requestPayload struct {
	Contents []struct {
		Parts []struct {
			InlineData *struct {
				MimeType string `json:"mimeType"`
				Data     string `json:"data"`
			} `json:"inlineData,omitempty"`
		} `json:"parts"`
	} `json:"contents"`
}

func TestGeminiClientFlushEmitsFinalEvent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/models/test-model:streamGenerateContent" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if q := r.URL.Query().Get("alt"); q != "sse" {
			t.Fatalf("expected alt=sse, got %q", q)
		}
		var payload requestPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if len(payload.Contents) == 0 || len(payload.Contents[0].Parts) < 2 {
			t.Fatalf("unexpected payload structure: %#v", payload)
		}
		inline := payload.Contents[0].Parts[1].InlineData
		if inline == nil || inline.MimeType != "audio/pcm;rate=16000" {
			t.Fatalf("missing inline audio: %#v", inline)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)
		_, _ = fmt.Fprintf(w, "data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"hello\"}]}}]}\n\n")
		if flusher != nil {
			flusher.Flush()
		}
		_, _ = fmt.Fprintf(w, "data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"hello world\"}]},\"finishReason\":\"STOP\"}]}\n\n")
		if flusher != nil {
			flusher.Flush()
		}
		_, _ = fmt.Fprintf(w, "data: [DONE]\n\n")
		if flusher != nil {
			flusher.Flush()
		}
	}))
	defer srv.Close()

	t.Setenv("GEMINI_API_KEY", "test-key")
	t.Setenv("GEMINI_ASR_MODEL", "test-model")
	t.Setenv("GEMINI_ASR_BASE_URL", srv.URL)

	client, err := New("gemini")
	if err != nil {
		t.Fatalf("New gemini client failed: %v", err)
	}
	defer client.Close()

	if !client.WritePCM([]byte{0x00, 0x01, 0x02, 0x03}) {
		t.Fatalf("WritePCM returned false")
	}
	client.Flush()

	select {
	case evt := <-client.Events():
		if evt.IsFinal || evt.Type != "partial" {
			t.Fatalf("unexpected first event: %#v", evt)
		}
		if evt.Text != "hello" {
			t.Fatalf("unexpected transcript: %q", evt.Text)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for first partial event")
	}

	select {
	case evt := <-client.Events():
		if evt.IsFinal || evt.Type != "partial" {
			t.Fatalf("unexpected second event: %#v", evt)
		}
		if evt.Text != "hello world" {
			t.Fatalf("unexpected transcript: %q", evt.Text)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for second partial event")
	}

	select {
	case evt := <-client.Events():
		if !evt.IsFinal || evt.Type != "final" {
			t.Fatalf("unexpected final event: %#v", evt)
		}
		if evt.Text != "hello world" {
			t.Fatalf("unexpected transcript: %q", evt.Text)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for final event")
	}
}

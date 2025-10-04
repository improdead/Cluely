package answer

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestCallGeminiSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/models/gemini-1.5-flash:generateContent" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("key") != "test-key" {
			t.Fatalf("missing or incorrect key: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"{\"answer\":\"Anchor ROI to their uptime risk\",\"followUp\":\"Who signs off on this?\"}"}]}}]}`))
	}))
	defer srv.Close()

	svc := NewServiceFromEnv()
	svc.apiKey = "test-key"
	svc.model = "gemini-1.5-flash"
	svc.baseURL = srv.URL
	svc.client = srv.Client()

	ans, err := svc.callGemini("prompt")
	if err != nil {
		t.Fatalf("callGemini returned error: %v", err)
	}
	if ans.Answer != "Anchor ROI to their uptime risk" {
		t.Fatalf("unexpected answer: %q", ans.Answer)
	}
	if ans.FollowUp != "Who signs off on this?" {
		t.Fatalf("unexpected followUp: %q", ans.FollowUp)
	}
	if ans.Confidence != 0.8 {
		t.Fatalf("expected confidence 0.8, got %v", ans.Confidence)
	}
}

func TestBuildPromptIncludesContext(t *testing.T) {
	got := buildPrompt("We need approval from finance soon", []string{"budget", "renewal"}, []string{"kickoff"}, []string{"procurement"})

	checks := []string{
		"<core_identity>",
		"<rules>",
		"Last frame tokens: procurement",
		"First frame tokens: kickoff",
		"Recent OCR tokens: budget, renewal",
		"Unique context tokens: budget, renewal, kickoff, procurement",
	}
	for _, substr := range checks {
		if !strings.Contains(got, substr) {
			t.Fatalf("prompt missing %q\nfull prompt:\n%s", substr, got)
		}
	}
}

func TestCallGeminiHandlesError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":{"code":400,"message":"bad"}}`, http.StatusBadRequest)
	}))
	defer srv.Close()

	svc := NewServiceFromEnv()
	svc.apiKey = "test-key"
	svc.model = "gemini-1.5-flash"
	svc.baseURL = srv.URL
	svc.client = srv.Client()

	if _, err := svc.callGemini("prompt"); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestMain(m *testing.M) {
	os.Setenv("GEMINI_API_KEY", "")
	os.Setenv("GEMINI_MODEL", "")
	os.Setenv("GEMINI_BASE_URL", "")
	code := m.Run()
	os.Exit(code)
}

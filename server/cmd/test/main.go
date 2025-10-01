package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"nhooyr.io/websocket"
)

// TestClient simulates a visionOS client sending transcripts and OCR tokens
func main() {
	url := envOr("TEST_WS_URL", "ws://localhost:8080/ws")
	fmt.Printf("🧪 Cluely Backend Test\n")
	fmt.Printf("Connecting to: %s\n\n", url)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		log.Fatalf("❌ Failed to connect: %v\nMake sure the server is running!", err)
	}
	defer c.Close(websocket.StatusNormalClosure, "test complete")

	fmt.Println("✅ Connected to server")

	// Start receiving messages in background
	go receiveMessages(ctx, c)

	// Test 1: Simple transcript (no OCR)
	fmt.Println("\n📝 Test 1: Simple transcript")
	sendJSON(ctx, c, map[string]any{
		"type":  "hello",
		"app":   "cluely-test",
		"ver":   "test",
	})
	time.Sleep(200 * time.Millisecond)

	sendJSON(ctx, c, map[string]any{
		"type":  "transcript",
		"text":  "We can probably meet Friday to discuss the budget",
		"final": true,
	})
	time.Sleep(3 * time.Second)

	// Test 2: Transcript with OCR context (first + last)
	fmt.Println("\n📝 Test 2: Transcript with visual context")
	sendJSON(ctx, c, map[string]any{
		"type":  "frame_meta",
		"ocr":   []string{"Q4", "Revenue", "$2.1M", "Profit", "-8%", "MAU", "45K"},
		"first": true,
	})
	time.Sleep(200 * time.Millisecond)

	sendJSON(ctx, c, map[string]any{
		"type":  "transcript",
		"text":  "So our Q4 revenue was down 8 percent but user engagement is way up",
		"final": true,
	})
	time.Sleep(3 * time.Second)

	sendJSON(ctx, c, map[string]any{
		"type": "frame_meta",
		"ocr":  []string{"Q4", "Summary", "Next", "Steps", "Action", "Items"},
		"last": true,
	})
	time.Sleep(2 * time.Second)

	// Test 3: Meeting context
	fmt.Println("\n📝 Test 3: Sales meeting scenario")
	sendJSON(ctx, c, map[string]any{
		"type":  "frame_meta",
		"ocr":   []string{"LinkedIn", "Roy", "Senior", "Engineer", "Bananazon", "8", "years"},
		"first": true,
	})
	time.Sleep(200 * time.Millisecond)

	sendJSON(ctx, c, map[string]any{
		"type":  "transcript",
		"text":  "I'm not sure if we have the budget for this right now",
		"final": true,
	})
	time.Sleep(3 * time.Second)

	// Test 4: Technical discussion
	fmt.Println("\n📝 Test 4: Technical explanation")
	sendJSON(ctx, c, map[string]any{
		"type": "frame_meta",
		"ocr":  []string{"System", "Design", "Microservices", "API", "Gateway", "Database", "Redis"},
	})
	time.Sleep(200 * time.Millisecond)

	sendJSON(ctx, c, map[string]any{
		"type":  "transcript",
		"text":  "I don't really understand how this architecture works",
		"final": true,
	})
	time.Sleep(3 * time.Second)

	fmt.Println("\n✅ All tests sent! Check responses above.")
	fmt.Println("💡 Tip: If you see hints, Gemini is working correctly!")
	fmt.Println("⚠️  If you see fallback hints, check your GEMINI_API_KEY")
	time.Sleep(1 * time.Second)
}

func receiveMessages(ctx context.Context, c *websocket.Conn) {
	for {
		_, data, err := c.Read(ctx)
		if err != nil {
			return
		}
		var msg map[string]any
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}
		typ, _ := msg["type"].(string)
		switch typ {
		case "hint":
			text, _ := msg["text"].(string)
			fmt.Printf("💡 HINT: %s\n", text)
		case "followup":
			text, _ := msg["text"].(string)
			fmt.Printf("   ↳ Follow-up: %s\n", text)
		case "partial":
			text, _ := msg["text"].(string)
			fmt.Printf("   (partial: %s)\n", text)
		case "final":
			text, _ := msg["text"].(string)
			fmt.Printf("📄 Final transcript: %s\n", text)
		case "warning":
			msg, _ := msg["msg"].(string)
			fmt.Printf("⚠️  Warning: %s\n", msg)
		case "error":
			msg, _ := msg["msg"].(string)
			fmt.Printf("❌ Error: %s\n", msg)
		}
	}
}

func sendJSON(ctx context.Context, c *websocket.Conn, v any) {
	b, _ := json.Marshal(v)
	wctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	_ = c.Write(wctx, websocket.MessageText, b)
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

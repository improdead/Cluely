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

// A tiny dev client to test the server without a visionOS app.
func main() {
	url := envOr("WS_URL", "ws://localhost:8080/ws")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer c.Close(websocket.StatusNormalClosure, "bye")

	// hello
	_ = writeJSON(ctx, c, map[string]any{"type": "hello"})
	// send a final transcript to trigger the heuristic hint engine
	_ = writeJSON(ctx, c, map[string]any{
		"type":  "transcript",
		"text":  "We can meet on Friday to review the budget and next steps.",
		"final": true,
	})

	// read a few messages
	readCtx, cancelRead := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancelRead()
	for i := 0; i < 5; i++ {
		typ, data, err := c.Read(readCtx)
		if err != nil {
			break
		}
		fmt.Printf("<- [%v] %s\n", typ, string(data))
	}
}

func writeJSON(ctx context.Context, c *websocket.Conn, v any) error {
	b, _ := json.Marshal(v)
	wctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return c.Write(wctx, websocket.MessageText, b)
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

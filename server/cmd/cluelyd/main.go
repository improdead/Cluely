package main

import (
	"log"
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/go-chi/chi/v5"

	"cluely/server/internal/obs"
	wsHandler "cluely/server/internal/ws"
)

func main() {
	// Leave a CPU for other processes to keep latency sensible
	runtime.GOMAXPROCS(max(1, runtime.NumCPU()-1))

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	// No special deadlines: aim for low-latency WS
	if tcpln, ok := ln.(*net.TCPListener); ok {
		_ = tcpln.SetDeadline(time.Time{})
	}

	r := chi.NewRouter()
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok")) })
	r.Get("/ws", wsHandler.Handle)

	// Start metrics logger (every 30s)
	obs.StartMetricsLogger(30 * time.Second)

	log.Println("cluelyd listening :8080")
	if err := http.Serve(ln, r); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

func max(a, b int) int { if a > b { return a }; return b }

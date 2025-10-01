package obs

import (
	"log"
	"sync/atomic"
	"time"
)

// Simple in-memory metrics (replace with OTel later)
var (
	SessionsActive     int64
	SessionsTotal      int64
	PCMFramesReceived  int64
	PCMFramesDropped   int64
	ASRPartialsRecv    int64
	ASRFinalsRecv      int64
	HintsSent          int64
	FollowupsSent      int64
	ErrorsASR          int64
	ErrorsAnswer       int64
)

func IncSessionActive() { atomic.AddInt64(&SessionsActive, 1); atomic.AddInt64(&SessionsTotal, 1) }
func DecSessionActive() { atomic.AddInt64(&SessionsActive, -1) }
func IncPCMFrame()      { atomic.AddInt64(&PCMFramesReceived, 1) }
func IncASRPartial()    { atomic.AddInt64(&ASRPartialsRecv, 1) }
func IncASRFinal()      { atomic.AddInt64(&ASRFinalsRecv, 1) }
func IncHint()          { atomic.AddInt64(&HintsSent, 1) }
func IncFollowup()      { atomic.AddInt64(&FollowupsSent, 1) }
func IncErrorASR()      { atomic.AddInt64(&ErrorsASR, 1) }
func IncErrorAnswer()   { atomic.AddInt64(&ErrorsAnswer, 1) }
func IncPCMFrameDrop()  { atomic.AddInt64(&PCMFramesDropped, 1) }

// LogMetrics prints current metrics (call periodically)
func LogMetrics() {
log.Printf("[metrics] sessions=%d pcm(in=%d drop=%d) asr(p=%d f=%d) hints=%d followups=%d errors(asr=%d ans=%d)",
		atomic.LoadInt64(&SessionsActive),
		atomic.LoadInt64(&PCMFramesReceived),
		atomic.LoadInt64(&PCMFramesDropped),
		atomic.LoadInt64(&ASRPartialsRecv),
		atomic.LoadInt64(&ASRFinalsRecv),
		atomic.LoadInt64(&HintsSent),
		atomic.LoadInt64(&FollowupsSent),
		atomic.LoadInt64(&ErrorsASR),
		atomic.LoadInt64(&ErrorsAnswer),
	)
}

// StartMetricsLogger logs metrics every interval
func StartMetricsLogger(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			LogMetrics()
		}
	}()
}

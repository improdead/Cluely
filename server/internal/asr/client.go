package asr

import (
	"context"
	"io"
	"log"
	"os"
	"strconv"
	"sync"
	"sync/atomic"

	speech "cloud.google.com/go/speech/apiv2"
	"cloud.google.com/go/speech/apiv2/speechpb"

	"cluely/server/internal/obs"
)

type Event struct {
	Type    string // "partial" or "final"
	Text    string
	IsFinal bool
}

type Client struct {
	ctx      context.Context
	cancel   context.CancelFunc
	events   chan Event
	pcmIn    chan []byte
	closeOne sync.Once
	dropped  int64
}

func NewClient() (*Client, error) {
	ctx, cancel := context.WithCancel(context.Background())
bufSize := envOrInt("ASR_PCM_BUFFER", 128)
	c := &Client{
		ctx:    ctx,
		cancel: cancel,
		events: make(chan Event, 8),
		pcmIn:  make(chan []byte, bufSize), // buffered to handle bursts
	}
	if err := c.start(); err != nil {
		cancel()
		return nil, err
	}
	return c, nil
}

func (c *Client) start() error {
	// Read env: GOOGLE_APPLICATION_CREDENTIALS, GCP_PROJECT_ID, GCP_RECOGNIZER
	projectID := envOr("GCP_PROJECT_ID", "your-gcp-project")
	recognizer := envOr("GCP_RECOGNIZER", "projects/"+projectID+"/locations/global/recognizers/_")

	client, err := speech.NewClient(c.ctx)
	if err != nil {
		return err
	}
	stream, err := client.StreamingRecognize(c.ctx)
	if err != nil {
		return err
	}
	// Send initial config (v2 StreamingRecognizeRequest)
	err = stream.Send(&speechpb.StreamingRecognizeRequest{
		Recognizer: recognizer,
		StreamingRequest: &speechpb.StreamingRecognizeRequest_StreamingConfig{
			StreamingConfig: &speechpb.StreamingRecognitionConfig{
				Config: &speechpb.RecognitionConfig{
					ExplicitDecodingConfig: &speechpb.ExplicitDecodingConfig{
						Encoding:        speechpb.ExplicitDecodingConfig_LINEAR16,
						SampleRateHertz: 16000,
						AudioChannelCount: 1,
					},
					LanguageCodes: []string{"en-US"},
					Model: "latest_long", // or "latest_short" for lower latency
				},
				StreamingFeatures: &speechpb.StreamingRecognitionFeatures{
					InterimResults: true,
				},
			},
		},
	})
	if err != nil {
		return err
	}

	// Writer goroutine: pcmIn -> stream.Send
	go func() {
		defer stream.CloseSend()
		for data := range c.pcmIn {
			if err := stream.Send(&speechpb.StreamingRecognizeRequest{
				Recognizer: recognizer,
				StreamingRequest: &speechpb.StreamingRecognizeRequest_Audio{
					Audio: data,
				},
			}); err != nil {
				log.Printf("asr send: %v", err)
				return
			}
		}
	}()

	// Reader goroutine: stream.Recv -> events
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF || err == context.Canceled {
				break
			}
			if err != nil {
				log.Printf("asr recv: %v", err)
				break
			}
			for _, res := range resp.Results {
				if len(res.Alternatives) == 0 {
					continue
				}
				alt := res.Alternatives[0]
				ev := Event{Text: alt.Transcript, IsFinal: res.IsFinal}
				if res.IsFinal {
					ev.Type = "final"
				} else {
					ev.Type = "partial"
				}
				select {
				case c.events <- ev:
				case <-c.ctx.Done():
					return
				}
			}
		}
		close(c.events)
	}()

	return nil
}

func (c *Client) WritePCM(data []byte) bool {
	// Non-blocking send; drop frame if buffer full (backpressure)
	select {
	case c.pcmIn <- data:
		return true
	default:
		atomic.AddInt64(&c.dropped, 1)
		obs.IncPCMFrameDrop()
		return false
	}
}

func (c *Client) Events() <-chan Event { return c.events }

func (c *Client) Dropped() int64 { return atomic.LoadInt64(&c.dropped) }

func (c *Client) Close() {
	c.closeOne.Do(func() {
		c.cancel()
		close(c.pcmIn)
	})
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func envOrInt(k string, d int) int {
	if v := os.Getenv(k); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return d
}

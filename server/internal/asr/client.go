package asr

import (
	"sync"
	"sync/atomic"

	"cluely/server/internal/obs"
)

type Event struct {
	Type    string
	Text    string
	IsFinal bool
}

type Client struct {
	closeOnce sync.Once
	events    chan Event
	dropped   int64
}

func NewClient() (*Client, error) {
	events := make(chan Event)
	close(events)

	return &Client{
		events: events,
	}, nil
}

func (c *Client) WritePCM(_ []byte) bool {
	atomic.AddInt64(&c.dropped, 1)
	obs.IncPCMFrameDrop()
	return false
}

func (c *Client) Events() <-chan Event {
	return c.events
}

func (c *Client) Dropped() int64 {
	return atomic.LoadInt64(&c.dropped)
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {})
}

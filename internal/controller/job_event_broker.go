package controller

import "sync"

// JobEventBroker fans out in-memory events to SSE clients.
type JobEventBroker struct {
	mu          sync.RWMutex
	subscribers map[chan any]struct{}
}

func NewJobEventBroker() *JobEventBroker {
	return &JobEventBroker{
		subscribers: make(map[chan any]struct{}),
	}
}

func (b *JobEventBroker) Subscribe() (<-chan any, func()) {
	ch := make(chan any, 1)

	b.mu.Lock()
	b.subscribers[ch] = struct{}{}
	b.mu.Unlock()

	return ch, func() {
		b.mu.Lock()
		if _, ok := b.subscribers[ch]; ok {
			delete(b.subscribers, ch)
			close(ch)
		}
		b.mu.Unlock()
	}
}

func (b *JobEventBroker) Publish(payload any) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.subscribers {
		select {
		case ch <- payload:
		default:
			// Drop stale events for slow subscribers.
		}
	}
}

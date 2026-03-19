package event

import "sync"

type Broker interface {
	Publish(Record)
	Subscribe(sessionID, threadID, projectID string) (<-chan Record, func())
}

type Hub struct {
	mu          sync.RWMutex
	subscribers map[uint64]subscription
	nextID      uint64
}

type subscription struct {
	sessionID string
	threadID  string
	projectID string
	ch        chan Record
}

func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[uint64]subscription),
	}
}

func (h *Hub) Publish(record Record) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, sub := range h.subscribers {
		if sub.sessionID != "" && sub.sessionID != record.SessionID {
			continue
		}
		if sub.threadID != "" {
			threadID, _ := record.Payload["thread_id"].(string)
			if threadID != sub.threadID {
				continue
			}
		}
		if sub.projectID != "" {
			projectID, _ := record.Payload["project_id"].(string)
			if projectID != sub.projectID {
				continue
			}
		}
		select {
		case sub.ch <- record:
		default:
		}
	}
}

func (h *Hub) Subscribe(sessionID, threadID, projectID string) (<-chan Record, func()) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.nextID++
	id := h.nextID
	ch := make(chan Record, 16)
	h.subscribers[id] = subscription{
		sessionID: sessionID,
		threadID:  threadID,
		projectID: projectID,
		ch:        ch,
	}

	cancel := func() {
		h.mu.Lock()
		defer h.mu.Unlock()

		sub, ok := h.subscribers[id]
		if !ok {
			return
		}
		delete(h.subscribers, id)
		close(sub.ch)
	}

	return ch, cancel
}

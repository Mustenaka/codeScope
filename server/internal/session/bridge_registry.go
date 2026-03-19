package session

import (
	"errors"
	"sync"
)

var ErrBridgeNotConnected = errors.New("bridge is not connected")

type BridgeRegistry struct {
	mu      sync.RWMutex
	clients map[string]chan any
}

func NewBridgeRegistry() *BridgeRegistry {
	return &BridgeRegistry{
		clients: make(map[string]chan any),
	}
}

func (r *BridgeRegistry) Register(sessionID string, outbound chan any) func() bool {
	r.mu.Lock()
	r.clients[sessionID] = outbound
	r.mu.Unlock()

	return func() bool {
		r.mu.Lock()
		defer r.mu.Unlock()
		current, ok := r.clients[sessionID]
		if !ok || current != outbound {
			return false
		}
		delete(r.clients, sessionID)
		return true
	}
}

func (r *BridgeRegistry) IsConnected(sessionID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.clients[sessionID]
	return ok
}

func (r *BridgeRegistry) ConnectedSessionIDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0, len(r.clients))
	for sessionID := range r.clients {
		ids = append(ids, sessionID)
	}
	return ids
}

func (r *BridgeRegistry) Send(sessionID string, message any) error {
	r.mu.RLock()
	outbound, ok := r.clients[sessionID]
	r.mu.RUnlock()
	if !ok {
		return ErrBridgeNotConnected
	}

	select {
	case outbound <- message:
		return nil
	default:
		return errors.New("bridge outbound queue is full")
	}
}

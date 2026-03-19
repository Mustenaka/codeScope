package capture

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"codescope/bridge/internal/discovery"
	"codescope/bridge/internal/session"
)

type Adapter interface {
	Attach(ctx context.Context, candidate discovery.Candidate, sink Sink) error
}

type DiscoverySource struct {
	scanner      discovery.Scanner
	pollInterval time.Duration
	bridgeMeta   session.Metadata
	logger       *log.Logger
	adapters     []Adapter
}

type activeSession struct {
	cancel context.CancelFunc
}

func NewDiscoverySource(scanner discovery.Scanner, pollInterval time.Duration, bridgeMeta session.Metadata, logger *log.Logger, adapters ...Adapter) *DiscoverySource {
	if logger == nil {
		logger = log.Default()
	}
	return &DiscoverySource{
		scanner:      scanner,
		pollInterval: pollInterval,
		bridgeMeta:   bridgeMeta,
		logger:       logger,
		adapters:     adapters,
	}
}

func (s *DiscoverySource) Start(ctx context.Context, sink Sink) error {
	active := make(map[string]activeSession)
	var mu sync.Mutex

	if err := s.scanOnceWithState(ctx, sink, active, &mu); err != nil {
		return err
	}

	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			mu.Lock()
			for _, session := range active {
				session.cancel()
			}
			mu.Unlock()
			return ctx.Err()
		case <-ticker.C:
			if err := s.scanOnceWithState(ctx, sink, active, &mu); err != nil {
				return err
			}
		}
	}
}

func (s *DiscoverySource) scanOnce(ctx context.Context, sink Sink) error {
	return s.scanOnceWithState(ctx, sink, map[string]activeSession{}, &sync.Mutex{})
}

func (s *DiscoverySource) scanOnceWithState(ctx context.Context, sink Sink, active map[string]activeSession, mu *sync.Mutex) error {
	candidates, err := s.scanner.Scan(ctx)
	if err != nil {
		s.logger.Printf("discovery scan failed: %v", err)
		return sink.Emit(ctx, ObservedEvent{
			Meta: s.bridgeMeta,
			Event: session.Event{
				Type: session.EventTypeError,
				Payload: map[string]any{
					"kind":    "discovery_failure",
					"message": err.Error(),
					"source":  "discovery",
				},
			},
		})
	}

	seen := make(map[string]discovery.Candidate, len(candidates))
	for _, candidate := range candidates {
		seen[candidate.Key()] = candidate
	}

	mu.Lock()
	for key, running := range active {
		if _, ok := seen[key]; ok {
			continue
		}
		running.cancel()
		delete(active, key)
	}

	for _, candidate := range candidates {
		key := candidate.Key()
		if _, ok := active[key]; ok {
			continue
		}
		sessionCtx, cancel := context.WithCancel(ctx)
		active[key] = activeSession{cancel: cancel}
		go s.runAdapters(sessionCtx, candidate, sink)
	}
	mu.Unlock()

	return nil
}

func (s *DiscoverySource) runAdapters(ctx context.Context, candidate discovery.Candidate, sink Sink) {
	for _, adapter := range s.adapters {
		adapter := adapter
		go func() {
			if err := adapter.Attach(ctx, candidate, sink); err != nil && ctx.Err() == nil {
				emitErr := sink.Emit(ctx, ObservedEvent{
					Meta: candidate.Meta,
					Event: session.Event{
						Type: session.EventTypeError,
						Payload: map[string]any{
							"kind":         "capture_adapter_failure",
							"message":      err.Error(),
							"adapter_type": fmt.Sprintf("%T", adapter),
							"source":       "capture",
						},
					},
				})
				if emitErr != nil {
					s.logger.Printf("emit capture adapter failure: %v", emitErr)
				}
			}
		}()
	}
}

type ProcessSnapshotAdapter struct{}

func (ProcessSnapshotAdapter) Attach(ctx context.Context, candidate discovery.Candidate, sink Sink) error {
	commandPayload := map[string]any{
		"pid":               candidate.PID,
		"process_name":      candidate.ProcessName,
		"command_line":      candidate.CommandLine,
		"source":            "process_discovery",
		"semantic_kind":     "debug_event",
		"observed":          true,
		"debug_category":    "process_observation",
		"thread_id":         candidate.Meta.SessionID,
		"source_session_id": candidate.Meta.SessionID,
	}
	if err := sink.Emit(ctx, ObservedEvent{
		Meta: candidate.Meta,
		Event: session.Event{
			Type:    session.EventTypeCommand,
			Payload: commandPayload,
		},
	}); err != nil {
		return err
	}

	return sink.Emit(ctx, ObservedEvent{
		Meta: candidate.Meta,
		Event: session.Event{
			Type: session.EventTypeTerminalOutput,
			Payload: map[string]any{
				"content":           fmt.Sprintf("[bridge] observing %s session pid=%d", candidate.Meta.AgentName, candidate.PID),
				"source":            "process_discovery",
				"semantic_kind":     "debug_event",
				"observed":          true,
				"debug_category":    "process_observation",
				"thread_id":         candidate.Meta.SessionID,
				"source_session_id": candidate.Meta.SessionID,
			},
		},
	})
}

type SessionHeartbeatAdapter struct {
	interval time.Duration
}

func NewSessionHeartbeatAdapter(interval time.Duration) SessionHeartbeatAdapter {
	return SessionHeartbeatAdapter{interval: interval}
}

func (a SessionHeartbeatAdapter) Attach(ctx context.Context, candidate discovery.Candidate, sink Sink) error {
	if a.interval <= 0 {
		return nil
	}

	if err := sink.Emit(ctx, ObservedEvent{
		Meta: candidate.Meta,
		Event: session.Event{
			Type:    session.EventTypeHeartbeat,
			Payload: map[string]any{"source": "discovery"},
		},
	}); err != nil {
		return err
	}

	ticker := time.NewTicker(a.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := sink.Emit(ctx, ObservedEvent{
				Meta: candidate.Meta,
				Event: session.Event{
					Type:    session.EventTypeHeartbeat,
					Payload: map[string]any{"source": "discovery"},
				},
			}); err != nil {
				return err
			}
		}
	}
}

type WorkspaceFileWatcherAdapter struct{}

func (WorkspaceFileWatcherAdapter) Attach(ctx context.Context, candidate discovery.Candidate, sink Sink) error {
	source, err := NewFileWatcherSource(candidate.Meta, candidate.Meta.WorkspaceRoot)
	if err != nil {
		return err
	}
	return source.Start(ctx, sink)
}

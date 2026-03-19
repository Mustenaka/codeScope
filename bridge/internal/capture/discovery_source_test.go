package capture

import (
	"context"
	"sync"
	"testing"
	"time"

	"codescope/bridge/internal/discovery"
	"codescope/bridge/internal/session"
)

type observedSinkRecorder struct {
	mu     sync.Mutex
	events []ObservedEvent
}

func (s *observedSinkRecorder) Emit(_ context.Context, event ObservedEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, event)
	return nil
}

func (s *observedSinkRecorder) snapshot() []ObservedEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]ObservedEvent, len(s.events))
	copy(out, s.events)
	return out
}

type scriptedScanner struct {
	mu        sync.Mutex
	snapshots [][]discovery.Candidate
	index     int
}

func (s *scriptedScanner) Scan(context.Context) ([]discovery.Candidate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.snapshots) == 0 {
		return nil, nil
	}
	if s.index >= len(s.snapshots) {
		return s.snapshots[len(s.snapshots)-1], nil
	}
	current := s.snapshots[s.index]
	s.index++
	return current, nil
}

type blockingAdapter struct {
	started chan string
	stopped chan string
}

func (a *blockingAdapter) Attach(ctx context.Context, candidate discovery.Candidate, sink Sink) error {
	select {
	case a.started <- candidate.Meta.SessionID:
	default:
	}
	select {
	case <-ctx.Done():
		select {
		case a.stopped <- candidate.Meta.SessionID:
		default:
		}
		return ctx.Err()
	}
}

func TestDiscoverySourceStartsNewSessionsOnceAndStopsMissingSessions(t *testing.T) {
	candidate := discovery.Candidate{
		Meta: session.Metadata{
			AgentName:     "codex",
			WorkspaceRoot: "D:/repo",
			MachineID:     "machine-1",
			SessionID:     "session-1",
		},
		PID: 101,
	}

	adapter := &blockingAdapter{
		started: make(chan string, 4),
		stopped: make(chan string, 4),
	}
	source := NewDiscoverySource(
		&scriptedScanner{snapshots: [][]discovery.Candidate{{candidate}, {candidate}, {}}},
		25*time.Millisecond,
		session.Metadata{SessionID: "bridge-session"},
		nil,
		adapter,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- source.Start(ctx, &observedSinkRecorder{})
	}()

	select {
	case started := <-adapter.started:
		if started != "session-1" {
			t.Fatalf("expected session-1 to start, got %q", started)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for session start")
	}

	select {
	case stopped := <-adapter.stopped:
		if stopped != "session-1" {
			t.Fatalf("expected session-1 to stop, got %q", stopped)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for session stop")
	}

	cancel()
	<-done

	select {
	case duplicate := <-adapter.started:
		t.Fatalf("expected session to start only once, got duplicate start for %q", duplicate)
	default:
	}
}

func TestDiscoverySourceReportsScanFailuresAsErrorEvents(t *testing.T) {
	source := NewDiscoverySource(failingScanner{}, 10*time.Millisecond, session.Metadata{
		AgentName:     "bridge",
		WorkspaceRoot: "D:/repo",
		MachineID:     "machine-1",
		SessionID:     "bridge-session",
	}, nil)
	recorder := &observedSinkRecorder{}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	if err := source.scanOnce(ctx, recorder); err != nil {
		t.Fatalf("scanOnce should degrade to error event, got %v", err)
	}

	events := recorder.snapshot()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Event.Type != session.EventTypeError {
		t.Fatalf("expected error event, got %q", events[0].Event.Type)
	}
}

func TestProcessSnapshotAdapterEmitsSemanticDebugFields(t *testing.T) {
	candidate := discovery.Candidate{
		Meta: session.Metadata{
			AgentName:     "codex",
			WorkspaceRoot: "D:/repo/codeScope",
			MachineID:     "machine-1",
			SessionID:     "session-1",
		},
		PID:         101,
		ProcessName: "codex.exe",
		CommandLine: "codex --cwd D:/repo/codeScope",
	}
	recorder := &observedSinkRecorder{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := (ProcessSnapshotAdapter{}).Attach(ctx, candidate, recorder); err != nil {
		t.Fatalf("attach process snapshot adapter: %v", err)
	}

	events := recorder.snapshot()
	if len(events) != 2 {
		t.Fatalf("expected 2 snapshot events, got %d", len(events))
	}

	commandPayload := events[0].Event.Payload
	if commandPayload["debug_category"] != "process_observation" {
		t.Fatalf("expected debug process_observation category, got %#v", commandPayload["debug_category"])
	}
	if commandPayload["semantic_kind"] != "debug_event" {
		t.Fatalf("expected debug semantic kind, got %#v", commandPayload["semantic_kind"])
	}
	if _, ok := commandPayload["thread_state"]; ok {
		t.Fatalf("expected process observation to avoid release thread_state, got %#v", commandPayload["thread_state"])
	}
	if commandPayload["thread_id"] != "session-1" {
		t.Fatalf("expected thread_id=session-1, got %#v", commandPayload["thread_id"])
	}
}

type failingScanner struct{}

func (failingScanner) Scan(context.Context) ([]discovery.Candidate, error) {
	return nil, context.DeadlineExceeded
}

package capture

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"codescope/bridge/internal/session"
)

type sinkRecorder struct {
	mu     sync.Mutex
	events []ObservedEvent
}

func (s *sinkRecorder) Emit(_ context.Context, event ObservedEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, event)
	return nil
}

func (s *sinkRecorder) snapshot() []ObservedEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]ObservedEvent, len(s.events))
	copy(out, s.events)
	return out
}

func TestReaderSourceEmitsLineEvents(t *testing.T) {
	meta := session.Metadata{SessionID: "session-1"}
	source := NewReaderSource(meta, "stdin", strings.NewReader("line one\nline two\n"), session.EventTypeTerminalOutput)
	recorder := &sinkRecorder{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := source.Start(ctx, recorder); err != nil {
		t.Fatalf("start reader source: %v", err)
	}

	if len(recorder.events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(recorder.events))
	}

	if recorder.events[0].Event.Payload["content"] != "line one" {
		t.Fatalf("expected first line payload, got %#v", recorder.events[0].Event.Payload["content"])
	}

	if recorder.events[1].Event.Payload["source"] != "stdin" {
		t.Fatalf("expected source label to be propagated, got %#v", recorder.events[1].Event.Payload["source"])
	}

	if recorder.events[0].Meta.SessionID != "session-1" {
		t.Fatalf("expected metadata to be propagated, got %#v", recorder.events[0].Meta)
	}
}

func TestJSONLSourceEmitsStructuredEvents(t *testing.T) {
	source := NewJSONLSource(session.Metadata{SessionID: "session-1"}, strings.NewReader(`{"event_type":"ai_output","payload":{"content":"planning next patch"}}
{"message_type":"heartbeat","payload":{"kind":"keepalive"}}`))
	recorder := &sinkRecorder{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := source.Start(ctx, recorder); err != nil {
		t.Fatalf("start jsonl source: %v", err)
	}

	if len(recorder.events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(recorder.events))
	}
	if recorder.events[0].Event.Type != session.EventTypeAIOutput {
		t.Fatalf("expected ai_output, got %q", recorder.events[0].Event.Type)
	}
	if recorder.events[1].Event.Type != session.EventTypeHeartbeat {
		t.Fatalf("expected heartbeat, got %q", recorder.events[1].Event.Type)
	}
}

func TestJSONLSourceRejectsInvalidRecords(t *testing.T) {
	source := NewJSONLSource(session.Metadata{}, strings.NewReader(`{"payload":{"content":"missing event type"}}`))
	recorder := &sinkRecorder{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := source.Start(ctx, recorder); err == nil {
		t.Fatal("expected invalid jsonl source input to fail")
	}
}

type semanticAdapterStub struct {
	name   string
	event  ObservedEvent
	called chan string
}

func (s semanticAdapterStub) Name() string {
	return s.name
}

func (s semanticAdapterStub) Start(ctx context.Context, sink Sink) error {
	select {
	case s.called <- s.name:
	default:
	}
	if err := sink.Emit(ctx, s.event); err != nil {
		return err
	}
	<-ctx.Done()
	return ctx.Err()
}

func TestSemanticCaptureSourceStartsAllAdapters(t *testing.T) {
	recorder := &sinkRecorder{}
	called := make(chan string, 4)
	source := NewSemanticCaptureSource(nil,
		semanticAdapterStub{
			name:   "codex",
			called: called,
			event: ObservedEvent{
				Meta:  session.Metadata{SessionID: "session-1"},
				Event: session.Event{Type: session.EventTypeCommand, Payload: map[string]any{"semantic_kind": "thread_message"}},
			},
		},
		semanticAdapterStub{
			name:   "claude",
			called: called,
			event: ObservedEvent{
				Meta:  session.Metadata{SessionID: "session-2"},
				Event: session.Event{Type: session.EventTypeAIOutput, Payload: map[string]any{"semantic_kind": "thread_message"}},
			},
		},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()

	err := source.Start(ctx, recorder)
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context termination, got %v", err)
	}

	started := map[string]bool{}
	for len(called) > 0 {
		started[<-called] = true
	}
	if !started["codex"] || !started["claude"] {
		t.Fatalf("expected both adapters to start, got %#v", started)
	}

	events := recorder.snapshot()
	if len(events) != 2 {
		t.Fatalf("expected 2 semantic events, got %d", len(events))
	}
}

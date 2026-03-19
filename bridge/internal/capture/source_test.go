package capture

import (
	"context"
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

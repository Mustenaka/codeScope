package capture

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"codescope/bridge/internal/session"
)

func TestCodexSessionSourceEmitsRealUserAndAssistantMessages(t *testing.T) {
	root := t.TempDir()
	codexRoot := filepath.Join(root, ".codex")
	sessionDir := filepath.Join(codexRoot, "sessions", "2026", "03", "19")
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("mkdir session dir: %v", err)
	}

	if err := os.WriteFile(
		filepath.Join(codexRoot, "session_index.jsonl"),
		[]byte("{\"id\":\"session-real-1\",\"thread_name\":\"Real capture thread\"}\n"),
		0o644,
	); err != nil {
		t.Fatalf("write session index: %v", err)
	}

	sessionFile := filepath.Join(sessionDir, "rollout-example-session-real-1.jsonl")
	contents := "" +
		"{\"timestamp\":\"2026-03-19T09:00:00Z\",\"type\":\"session_meta\",\"payload\":{\"id\":\"session-real-1\",\"cwd\":\"D:/Work/Code/Cross/codeScope\"}}\n" +
		"{\"timestamp\":\"2026-03-19T09:01:00Z\",\"type\":\"event_msg\",\"payload\":{\"type\":\"user_message\",\"message\":\"请修复真实消息采集。\"}}\n" +
		"{\"timestamp\":\"2026-03-19T09:02:00Z\",\"type\":\"event_msg\",\"payload\":{\"type\":\"agent_message\",\"message\":\"已开始接入真实消息采集。\"}}\n"
	if err := os.WriteFile(sessionFile, []byte(contents), 0o644); err != nil {
		t.Fatalf("write session file: %v", err)
	}

	source := NewCodexSessionSource(filepath.Join(root, ".codex"), "machine-1", 10*time.Millisecond, log.New(os.Stdout, "", 0))
	recorder := &sinkRecorder{}

	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- source.Start(ctx, recorder)
	}()

	deadline := time.Now().Add(60 * time.Millisecond)
	for time.Now().Before(deadline) {
		if len(recorder.snapshot()) >= 2 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	cancel()
	<-done

	events := recorder.snapshot()
	if len(events) != 2 {
		t.Fatalf("expected 2 message events, got %d (%#v)", len(events), events)
	}

	if events[0].Meta.SessionID != "session-real-1" {
		t.Fatalf("expected real session id, got %q", events[0].Meta.SessionID)
	}
	if events[0].Meta.WorkspaceRoot != "D:/Work/Code/Cross/codeScope" {
		t.Fatalf("expected workspace root from session meta, got %q", events[0].Meta.WorkspaceRoot)
	}
	if events[0].Event.Type != session.EventTypeCommand {
		t.Fatalf("expected first event to be command, got %q", events[0].Event.Type)
	}
	if events[0].Event.Payload["role"] != "user" {
		t.Fatalf("expected user role payload, got %#v", events[0].Event.Payload["role"])
	}
	if events[0].Event.Payload["thread_title"] != "Real capture thread" {
		t.Fatalf("expected thread title from session index, got %#v", events[0].Event.Payload["thread_title"])
	}
	if events[0].Event.Payload["semantic_kind"] != "thread_message" {
		t.Fatalf("expected semantic thread_message kind, got %#v", events[0].Event.Payload["semantic_kind"])
	}
	if events[0].Event.Payload["thread_state"] != session.ThreadStateRunning {
		t.Fatalf("expected user message to keep thread running, got %#v", events[0].Event.Payload["thread_state"])
	}
	if events[1].Event.Type != session.EventTypeAIOutput {
		t.Fatalf("expected second event to be ai_output, got %q", events[1].Event.Type)
	}
	if events[1].Event.Payload["semantic_kind"] != "thread_message" {
		t.Fatalf("expected assistant semantic thread_message kind, got %#v", events[1].Event.Payload["semantic_kind"])
	}
	if events[1].Event.Payload["thread_state"] != session.ThreadStateWaitingPrompt {
		t.Fatalf("expected assistant full-turn capture to mark waiting_prompt, got %#v", events[1].Event.Payload["thread_state"])
	}
}

func TestCodexSessionSourceEmitsAppendedMessagesWithoutReplayingOldOnSameRun(t *testing.T) {
	root := t.TempDir()
	codexRoot := filepath.Join(root, ".codex")
	sessionDir := filepath.Join(codexRoot, "sessions", "2026", "03", "19")
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("mkdir session dir: %v", err)
	}

	sessionFile := filepath.Join(sessionDir, "rollout-example-session-real-2.jsonl")
	initial := "" +
		"{\"timestamp\":\"2026-03-19T09:00:00Z\",\"type\":\"session_meta\",\"payload\":{\"id\":\"session-real-2\",\"cwd\":\"D:/Work/Code/Cross/codeScope\"}}\n" +
		"{\"timestamp\":\"2026-03-19T09:01:00Z\",\"type\":\"event_msg\",\"payload\":{\"type\":\"user_message\",\"message\":\"第一条消息。\"}}\n"
	if err := os.WriteFile(sessionFile, []byte(initial), 0o644); err != nil {
		t.Fatalf("write initial session file: %v", err)
	}

	source := NewCodexSessionSource(filepath.Join(root, ".codex"), "machine-1", 10*time.Millisecond, log.New(os.Stdout, "", 0))
	recorder := &sinkRecorder{}

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- source.Start(ctx, recorder)
	}()

	time.Sleep(30 * time.Millisecond)
	updated := initial + "{\"timestamp\":\"2026-03-19T09:02:00Z\",\"type\":\"event_msg\",\"payload\":{\"type\":\"agent_message\",\"message\":\"第二条消息。\"}}\n"
	if err := os.WriteFile(sessionFile, []byte(updated), 0o644); err != nil {
		t.Fatalf("append session file: %v", err)
	}

	deadline := time.Now().Add(80 * time.Millisecond)
	for time.Now().Before(deadline) {
		if len(recorder.snapshot()) >= 2 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	cancel()
	<-done

	events := recorder.snapshot()
	if len(events) != 2 {
		t.Fatalf("expected 2 total events, got %d", len(events))
	}
	if events[1].Event.Payload["content"] != "第二条消息。" {
		t.Fatalf("expected appended assistant content, got %#v", events[1].Event.Payload["content"])
	}
}

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

func TestClaudeSessionSourceEmitsRealUserAndAssistantMessages(t *testing.T) {
	root := t.TempDir()
	claudeRoot := filepath.Join(root, ".claude")
	projectDir := filepath.Join(claudeRoot, "projects", "d--Work-Code-Cross-codeScope")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project dir: %v", err)
	}

	sessionID := "claude-session-1"
	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")
	index := `{
  "version": 1,
  "entries": [
    {
      "sessionId": "claude-session-1",
      "fullPath": "` + filepath.ToSlash(sessionFile) + `",
      "projectPath": "D:/Work/Code/Cross/codeScope",
      "firstPrompt": "Continue the real message capture rollout",
      "created": "2026-03-19T09:00:00Z",
      "modified": "2026-03-19T09:03:00Z"
    }
  ]
}`
	if err := os.WriteFile(filepath.Join(projectDir, "sessions-index.json"), []byte(index), 0o644); err != nil {
		t.Fatalf("write session index: %v", err)
	}

	contents := "" +
		"{\"type\":\"user\",\"message\":{\"role\":\"user\",\"content\":[{\"type\":\"text\",\"text\":\"Continue the real message capture rollout\"}]},\"timestamp\":\"2026-03-19T09:01:00Z\",\"cwd\":\"D:/Work/Code/Cross/codeScope\",\"sessionId\":\"claude-session-1\"}\n" +
		"{\"type\":\"assistant\",\"message\":{\"role\":\"assistant\",\"content\":[{\"type\":\"tool_use\",\"id\":\"toolu_1\",\"name\":\"Read\",\"input\":{\"file_path\":\"D:/Work/Code/Cross/codeScope/README.md\"}}],\"stop_reason\":\"tool_use\"},\"timestamp\":\"2026-03-19T09:01:05Z\",\"cwd\":\"D:/Work/Code/Cross/codeScope\",\"sessionId\":\"claude-session-1\"}\n" +
		"{\"type\":\"user\",\"message\":{\"role\":\"user\",\"content\":[{\"type\":\"tool_result\",\"tool_use_id\":\"toolu_1\",\"content\":\"ok\"}]},\"timestamp\":\"2026-03-19T09:01:06Z\",\"cwd\":\"D:/Work/Code/Cross/codeScope\",\"sessionId\":\"claude-session-1\"}\n" +
		"{\"type\":\"assistant\",\"message\":{\"role\":\"assistant\",\"content\":[{\"type\":\"text\",\"text\":\"I wired the semantic adapter and skipped tool noise.\"}],\"stop_reason\":\"end_turn\"},\"timestamp\":\"2026-03-19T09:02:00Z\",\"cwd\":\"D:/Work/Code/Cross/codeScope\",\"sessionId\":\"claude-session-1\"}\n"
	if err := os.WriteFile(sessionFile, []byte(contents), 0o644); err != nil {
		t.Fatalf("write session file: %v", err)
	}

	source := NewClaudeSessionSource(claudeRoot, "machine-1", 10*time.Millisecond, log.New(os.Stdout, "", 0))
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

	if events[0].Meta.SessionID != sessionID {
		t.Fatalf("expected real session id, got %q", events[0].Meta.SessionID)
	}
	if events[0].Meta.WorkspaceRoot != "D:/Work/Code/Cross/codeScope" {
		t.Fatalf("expected workspace root from transcript/index, got %q", events[0].Meta.WorkspaceRoot)
	}
	if events[0].Event.Type != session.EventTypeCommand {
		t.Fatalf("expected user event to be command, got %q", events[0].Event.Type)
	}
	if events[0].Event.Payload["semantic_kind"] != "thread_message" {
		t.Fatalf("expected semantic thread_message kind, got %#v", events[0].Event.Payload["semantic_kind"])
	}
	if events[0].Event.Payload["thread_title"] != "Continue the real message capture rollout" {
		t.Fatalf("expected thread title from sessions index, got %#v", events[0].Event.Payload["thread_title"])
	}
	if events[1].Event.Type != session.EventTypeAIOutput {
		t.Fatalf("expected assistant event to be ai_output, got %q", events[1].Event.Type)
	}
	if events[1].Event.Payload["content"] != "I wired the semantic adapter and skipped tool noise." {
		t.Fatalf("expected assistant text output, got %#v", events[1].Event.Payload["content"])
	}
	if events[1].Event.Payload["thread_state"] != session.ThreadStateWaitingPrompt {
		t.Fatalf("expected assistant final text to mark waiting_prompt, got %#v", events[1].Event.Payload["thread_state"])
	}
}

func TestClaudeSessionSourceEmitsAppendedMessagesWithoutReplayingOldOnSameRun(t *testing.T) {
	root := t.TempDir()
	claudeRoot := filepath.Join(root, ".claude")
	projectDir := filepath.Join(claudeRoot, "projects", "d--Work-Code-Cross-codeScope")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project dir: %v", err)
	}

	sessionID := "claude-session-2"
	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")
	initial := "" +
		"{\"type\":\"user\",\"message\":{\"role\":\"user\",\"content\":[{\"type\":\"text\",\"text\":\"First prompt\"}]},\"timestamp\":\"2026-03-19T09:01:00Z\",\"cwd\":\"D:/Work/Code/Cross/codeScope\",\"sessionId\":\"claude-session-2\"}\n"
	if err := os.WriteFile(sessionFile, []byte(initial), 0o644); err != nil {
		t.Fatalf("write initial session file: %v", err)
	}

	source := NewClaudeSessionSource(claudeRoot, "machine-1", 10*time.Millisecond, log.New(os.Stdout, "", 0))
	recorder := &sinkRecorder{}

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- source.Start(ctx, recorder)
	}()

	time.Sleep(30 * time.Millisecond)
	updated := initial + "{\"type\":\"assistant\",\"message\":{\"role\":\"assistant\",\"content\":[{\"type\":\"text\",\"text\":\"Second answer\"}],\"stop_reason\":\"end_turn\"},\"timestamp\":\"2026-03-19T09:02:00Z\",\"cwd\":\"D:/Work/Code/Cross/codeScope\",\"sessionId\":\"claude-session-2\"}\n"
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
	if events[1].Event.Payload["content"] != "Second answer" {
		t.Fatalf("expected appended assistant content, got %#v", events[1].Event.Payload["content"])
	}
}

func TestClaudeSessionSourceMarksPauseTurnAsWaitingReview(t *testing.T) {
	root := t.TempDir()
	claudeRoot := filepath.Join(root, ".claude")
	projectDir := filepath.Join(claudeRoot, "projects", "d--Work-Code-Cross-codeScope")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project dir: %v", err)
	}

	sessionID := "claude-session-review-1"
	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")
	contents := "" +
		"{\"type\":\"assistant\",\"message\":{\"role\":\"assistant\",\"content\":[{\"type\":\"text\",\"text\":\"Please approve the proposed refactor before I continue.\"}],\"stop_reason\":\"pause_turn\"},\"timestamp\":\"2026-03-19T09:02:00Z\",\"cwd\":\"D:/Work/Code/Cross/codeScope\",\"sessionId\":\"claude-session-review-1\"}\n"
	if err := os.WriteFile(sessionFile, []byte(contents), 0o644); err != nil {
		t.Fatalf("write session file: %v", err)
	}

	source := NewClaudeSessionSource(claudeRoot, "machine-1", 10*time.Millisecond, log.New(os.Stdout, "", 0))
	recorder := &sinkRecorder{}

	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- source.Start(ctx, recorder)
	}()

	deadline := time.Now().Add(60 * time.Millisecond)
	for time.Now().Before(deadline) {
		if len(recorder.snapshot()) >= 1 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	cancel()
	<-done

	events := recorder.snapshot()
	if len(events) != 1 {
		t.Fatalf("expected 1 assistant event, got %d", len(events))
	}
	if events[0].Event.Payload["thread_state"] != session.ThreadStateWaitingReview {
		t.Fatalf("expected pause_turn to map to waiting_review, got %#v", events[0].Event.Payload["thread_state"])
	}
}

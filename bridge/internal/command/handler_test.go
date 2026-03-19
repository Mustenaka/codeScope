package command

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"codescope/bridge/internal/session"
)

type publisherRecorder struct {
	messages []session.Message
}

func (p *publisherRecorder) Publish(_ context.Context, msg session.Message) error {
	p.messages = append(p.messages, msg)
	return nil
}

func TestHandleSendPromptWritesToLocalInboxAndPublishesResult(t *testing.T) {
	tempDir := t.TempDir()
	inbox := filepath.Join(tempDir, "inbox.jsonl")
	meta := session.Metadata{
		AgentName:     "codex",
		WorkspaceRoot: tempDir,
		MachineID:     "machine-1",
		SessionID:     "session-1",
	}

	handler := NewHandler(meta, nil, NewFilePromptSink(inbox))
	handler.now = func() time.Time {
		return time.Date(2026, 3, 17, 10, 0, 0, 0, time.UTC)
	}

	command := session.Message{
		MessageID:   "msg-1",
		SessionID:   "session-1",
		MessageType: session.MessageTypeCommand,
		CommandID:   "cmd-1",
		CommandType: session.CommandTypeSendPrompt,
		Timestamp:   "2026-03-17T10:00:00Z",
		Payload: map[string]any{
			"content": "continue fixing tests",
		},
	}

	recorder := &publisherRecorder{}
	if err := handler.Handle(context.Background(), command, recorder); err != nil {
		t.Fatalf("handle command: %v", err)
	}

	if len(recorder.messages) != 2 {
		t.Fatalf("expected 2 published messages, got %d", len(recorder.messages))
	}

	result := recorder.messages[1]
	if result.MessageType != session.MessageTypeCommandResult {
		t.Fatalf("expected command_result message, got %q", result.MessageType)
	}

	if result.Status != session.StatusSuccess {
		t.Fatalf("expected success status, got %q", result.Status)
	}

	content, err := os.ReadFile(inbox)
	if err != nil {
		t.Fatalf("read prompt inbox: %v", err)
	}

	if !strings.Contains(string(content), "continue fixing tests") {
		t.Fatalf("expected prompt file to contain command content, got %q", string(content))
	}

	var record map[string]any
	line := strings.TrimSpace(string(content))
	if err := json.Unmarshal([]byte(line), &record); err != nil {
		t.Fatalf("unmarshal prompt record: %v", err)
	}

	if record["command_id"] != "cmd-1" {
		t.Fatalf("expected command_id in prompt record, got %#v", record["command_id"])
	}
}

func TestHandleSendPromptFailsClearlyWhenInjectionIsUnsupported(t *testing.T) {
	meta := session.Metadata{
		AgentName:     "bridge",
		WorkspaceRoot: "D:/repo",
		MachineID:     "machine-1",
		SessionID:     "bridge-session",
	}

	handler := NewHandler(meta, nil, NewUnsupportedPromptSink("side-channel mode does not support prompt injection"))
	command := session.Message{
		MessageID:   "msg-1",
		SessionID:   "session-42",
		MessageType: session.MessageTypeCommand,
		CommandID:   "cmd-42",
		CommandType: session.CommandTypeSendPrompt,
		Timestamp:   "2026-03-17T10:00:00Z",
		Payload: map[string]any{
			"content": "continue",
		},
	}

	recorder := &publisherRecorder{}
	if err := handler.Handle(context.Background(), command, recorder); err != nil {
		t.Fatalf("handle command: %v", err)
	}

	if len(recorder.messages) != 2 {
		t.Fatalf("expected 2 published messages, got %d", len(recorder.messages))
	}

	result := recorder.messages[1]
	if result.SessionID != "session-42" {
		t.Fatalf("expected command result to target original session, got %q", result.SessionID)
	}
	if result.Status != session.StatusFailed {
		t.Fatalf("expected failed status, got %q", result.Status)
	}
	if accepted, _ := result.Payload["accepted"].(bool); accepted {
		t.Fatalf("expected accepted=false, got payload %#v", result.Payload)
	}
	if !strings.Contains(result.Payload["error"].(string), "side-channel") {
		t.Fatalf("expected side-channel failure message, got %#v", result.Payload["error"])
	}
}

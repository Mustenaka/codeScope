package session

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewEventMessageIncludesSessionMetadata(t *testing.T) {
	now := time.Date(2026, 3, 17, 10, 0, 0, 0, time.UTC)
	meta := Metadata{
		AgentName:     "codex",
		WorkspaceRoot: "D:/Work/Code/Cross/codeScope",
		MachineID:     "machine-1",
		SessionID:     "session-1",
	}

	msg := NewEventMessage(meta, EventTypeTerminalOutput, map[string]any{
		"content": "go test ./...",
		"stream":  "stdout",
	}, now)

	if msg.SessionID != "session-1" {
		t.Fatalf("expected session id to be propagated, got %q", msg.SessionID)
	}

	if msg.MessageType != MessageTypeEvent {
		t.Fatalf("expected message type %q, got %q", MessageTypeEvent, msg.MessageType)
	}

	if msg.EventType != EventTypeTerminalOutput {
		t.Fatalf("expected event type %q, got %q", EventTypeTerminalOutput, msg.EventType)
	}

	if msg.Timestamp != now.UTC().Format(time.RFC3339Nano) {
		t.Fatalf("expected timestamp %q, got %q", now.UTC().Format(time.RFC3339Nano), msg.Timestamp)
	}

	if msg.MessageID == "" {
		t.Fatal("expected generated message id")
	}

	if msg.Payload["agent_name"] != "codex" {
		t.Fatalf("expected payload to include agent_name, got %#v", msg.Payload["agent_name"])
	}

	if msg.Payload["project_id"] == "" {
		t.Fatalf("expected payload to include project_id, got %#v", msg.Payload["project_id"])
	}
	if msg.Payload["thread_id"] != "session-1" {
		t.Fatalf("expected payload to include thread_id=session-1, got %#v", msg.Payload["thread_id"])
	}
	if msg.Payload["source_session_id"] != "session-1" {
		t.Fatalf("expected payload to include source_session_id=session-1, got %#v", msg.Payload["source_session_id"])
	}
	if msg.Payload["thread_state"] != ThreadStateRunning {
		t.Fatalf("expected terminal output to default to running thread_state, got %#v", msg.Payload["thread_state"])
	}
}

func TestNewHeartbeatMessageIncludesProjectAndThreadMetadata(t *testing.T) {
	now := time.Date(2026, 3, 19, 10, 0, 0, 0, time.UTC)
	meta := Metadata{
		AgentName:     "claude",
		WorkspaceRoot: "D:/Work/Code/Cross/codeScope",
		MachineID:     "machine-1",
		SessionID:     "session-2",
	}

	msg := NewHeartbeatMessage(meta, now)

	if msg.Payload["thread_id"] != "session-2" {
		t.Fatalf("expected thread_id=session-2, got %#v", msg.Payload["thread_id"])
	}
	if msg.Payload["thread_state"] != ThreadStateRunning {
		t.Fatalf("expected heartbeat thread_state running, got %#v", msg.Payload["thread_state"])
	}
	if msg.Payload["project_name"] != "codeScope" {
		t.Fatalf("expected project_name codeScope, got %#v", msg.Payload["project_name"])
	}
}

func TestNewCommandResultMessageMarksFailedPromptAsWaitingPrompt(t *testing.T) {
	now := time.Date(2026, 3, 19, 10, 0, 0, 0, time.UTC)
	meta := Metadata{
		AgentName:     "codex",
		WorkspaceRoot: "D:/Work/Code/Cross/codeScope",
		MachineID:     "machine-1",
		SessionID:     "session-3",
	}

	msg := NewCommandResultMessage(meta, "cmd-1", CommandTypeSendPrompt, StatusFailed, map[string]any{
		"accepted": false,
		"error":    "side-channel mode does not support prompt injection",
	}, now)

	if msg.Payload["thread_id"] != "session-3" {
		t.Fatalf("expected thread_id=session-3, got %#v", msg.Payload["thread_id"])
	}
	if msg.Payload["thread_state"] != ThreadStateWaitingPrompt {
		t.Fatalf("expected failed send_prompt to mark waiting_prompt, got %#v", msg.Payload["thread_state"])
	}
}

func TestMessageJSONRoundTrip(t *testing.T) {
	in := Message{
		MessageID:   "msg-1",
		SessionID:   "session-1",
		MessageType: MessageTypeCommandResult,
		Timestamp:   "2026-03-17T10:00:00Z",
		CommandID:   "cmd-1",
		CommandType: CommandTypeSendPrompt,
		Status:      StatusSuccess,
		Payload: map[string]any{
			"accepted": true,
		},
	}

	data, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal message: %v", err)
	}

	var out Message
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal message: %v", err)
	}

	if out.MessageID != in.MessageID {
		t.Fatalf("expected message id %q, got %q", in.MessageID, out.MessageID)
	}

	if out.CommandType != in.CommandType {
		t.Fatalf("expected command type %q, got %q", in.CommandType, out.CommandType)
	}

	if accepted, ok := out.Payload["accepted"].(bool); !ok || !accepted {
		t.Fatalf("expected accepted payload to survive round trip, got %#v", out.Payload["accepted"])
	}
}

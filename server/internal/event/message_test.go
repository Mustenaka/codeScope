package event

import "testing"

func TestMessageValidateSuccess(t *testing.T) {
	message := Message{
		MessageID:   "message-1",
		SessionID:   "session-1",
		MessageType: MessageTypeEvent,
		EventType:   TypeTerminalOutput,
		Timestamp:   "2026-03-17T10:00:00Z",
		Payload:     map[string]any{"content": "running"},
	}

	if err := message.Validate(); err != nil {
		t.Fatalf("validate message: %v", err)
	}
}

func TestMessageValidateRequiresEventTypeForEvent(t *testing.T) {
	message := Message{
		MessageID:   "message-1",
		SessionID:   "session-1",
		MessageType: MessageTypeEvent,
		Timestamp:   "2026-03-17T10:00:00Z",
		Payload:     map[string]any{"content": "running"},
	}

	if err := message.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestMessageValidateRejectsInvalidMessageType(t *testing.T) {
	message := Message{
		MessageID:   "message-1",
		SessionID:   "session-1",
		MessageType: "bad",
		Timestamp:   "2026-03-17T10:00:00Z",
	}

	if err := message.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

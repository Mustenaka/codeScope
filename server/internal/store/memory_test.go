package store

import (
	"testing"
	"time"

	"codescope/server/internal/event"
	"codescope/server/internal/session"
)

func TestMemorySessionStoreCRUD(t *testing.T) {
	store := NewMemorySessionStore()
	now := time.Date(2026, 3, 17, 10, 0, 0, 0, time.UTC)

	record := session.Session{
		ID:            "session-1",
		ProjectName:   "codeScope",
		WorkspaceRoot: "/workspace",
		MachineID:     "machine-1",
		Status:        session.StatusCreated,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := store.Create(record); err != nil {
		t.Fatalf("create session: %v", err)
	}

	saved, err := store.Get("session-1")
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if saved.ProjectName != "codeScope" {
		t.Fatalf("expected project name codeScope, got %q", saved.ProjectName)
	}

	record.Status = session.StatusRunning
	record.UpdatedAt = now.Add(time.Minute)
	if err := store.Update(record); err != nil {
		t.Fatalf("update session: %v", err)
	}

	list, err := store.List()
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 session, got %d", len(list))
	}
	if list[0].Status != session.StatusRunning {
		t.Fatalf("expected running status, got %q", list[0].Status)
	}
}

func TestMemoryEventStoreAppendAndListBySession(t *testing.T) {
	store := NewMemoryEventStore()
	base := time.Date(2026, 3, 17, 10, 0, 0, 0, time.UTC)

	first := event.Record{
		ID:          "event-1",
		MessageID:   "message-1",
		SessionID:   "session-1",
		MessageType: event.MessageTypeEvent,
		EventType:   event.TypeTerminalOutput,
		Timestamp:   base,
		Payload:     map[string]any{"content": "first"},
		CreatedAt:   base,
	}
	second := event.Record{
		ID:          "event-2",
		MessageID:   "message-2",
		SessionID:   "session-1",
		MessageType: event.MessageTypeEvent,
		EventType:   event.TypeAIOutput,
		Timestamp:   base.Add(time.Second),
		Payload:     map[string]any{"content": "second"},
		CreatedAt:   base.Add(time.Second),
	}
	other := event.Record{
		ID:          "event-3",
		MessageID:   "message-3",
		SessionID:   "session-2",
		MessageType: event.MessageTypeEvent,
		EventType:   event.TypeAIOutput,
		Timestamp:   base.Add(2 * time.Second),
		Payload:     map[string]any{"content": "other"},
		CreatedAt:   base.Add(2 * time.Second),
	}

	for _, item := range []event.Record{second, other, first} {
		if err := store.Append(item); err != nil {
			t.Fatalf("append event %s: %v", item.ID, err)
		}
	}

	list, err := store.ListBySession("session-1")
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 events, got %d", len(list))
	}
	if list[0].ID != "event-1" || list[1].ID != "event-2" {
		t.Fatalf("expected chronological order, got %q then %q", list[0].ID, list[1].ID)
	}
}

package event

import (
	"errors"
	"testing"
	"time"

	"codescope/server/internal/session"
)

func TestServiceIngestAutoCreatesSession(t *testing.T) {
	sessionStore := newTestSessionStore()
	eventStore := newTestEventStore()
	hub := NewHub()
	service := NewService(eventStore, sessionStore, hub)

	record, err := service.Ingest(Message{
		MessageID:   "msg-1",
		SessionID:   "session-1",
		MessageType: MessageTypeEvent,
		EventType:   TypeTerminalOutput,
		Timestamp:   "2026-03-17T10:00:00Z",
		Payload: map[string]any{
			"content":        "go test ./...",
			"project_name":   "codeScope",
			"workspace_root": "/workspace",
			"machine_id":     "machine-1",
		},
	})
	if err != nil {
		t.Fatalf("ingest message: %v", err)
	}
	if record.SessionID != "session-1" {
		t.Fatalf("expected record session-1, got %q", record.SessionID)
	}

	saved, err := sessionStore.Get("session-1")
	if err != nil {
		t.Fatalf("get auto-created session: %v", err)
	}
	if saved.ProjectName != "codeScope" {
		t.Fatalf("expected project name codeScope, got %q", saved.ProjectName)
	}
	if saved.Status != session.StatusRunning {
		t.Fatalf("expected running session, got %q", saved.Status)
	}
}

func TestServiceIngestHeartbeatKeepsSessionQueryable(t *testing.T) {
	sessionStore := newTestSessionStore()
	eventStore := newTestEventStore()
	hub := NewHub()
	service := NewService(eventStore, sessionStore, hub)

	record, err := service.Ingest(Message{
		MessageID:   "msg-heartbeat",
		SessionID:   "session-heartbeat",
		MessageType: MessageTypeHeartbeat,
		Timestamp:   "2026-03-17T10:00:05Z",
		Payload: map[string]any{
			"agent_name":     "fake-source",
			"workspace_root": "/workspace",
			"machine_id":     "machine-2",
		},
	})
	if err != nil {
		t.Fatalf("ingest heartbeat: %v", err)
	}
	if record.EventType != TypeHeartbeat {
		t.Fatalf("expected heartbeat event type, got %q", record.EventType)
	}

	saved, err := sessionStore.Get("session-heartbeat")
	if err != nil {
		t.Fatalf("get auto-created heartbeat session: %v", err)
	}
	if saved.Status != session.StatusCreated {
		t.Fatalf("expected created session after heartbeat, got %q", saved.Status)
	}
	expected := time.Date(2026, 3, 17, 10, 0, 5, 0, time.UTC)
	if !saved.UpdatedAt.Equal(expected) {
		t.Fatalf("expected updated_at %s, got %s", expected, saved.UpdatedAt)
	}
}

func TestServiceIngestRefreshesSessionMetadataWhenPayloadGetsBetter(t *testing.T) {
	sessionStore := newTestSessionStore()
	sessionStore.sessions["session-1"] = session.Session{
		ID:            "session-1",
		ProjectName:   "bridge",
		WorkspaceRoot: "/workspace/codeScope/bridge",
		MachineID:     "machine-1",
		Status:        session.StatusCreated,
	}
	eventStore := newTestEventStore()
	hub := NewHub()
	service := NewService(eventStore, sessionStore, hub)

	_, err := service.Ingest(Message{
		MessageID:   "msg-refresh",
		SessionID:   "session-1",
		MessageType: MessageTypeEvent,
		EventType:   TypeCommand,
		Timestamp:   "2026-03-19T10:00:00Z",
		Payload: map[string]any{
			"project_name":   "codeScope",
			"workspace_root": "/workspace/codeScope",
			"machine_id":     "machine-1",
		},
	})
	if err != nil {
		t.Fatalf("ingest message: %v", err)
	}

	saved, err := sessionStore.Get("session-1")
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if saved.ProjectName != "codeScope" {
		t.Fatalf("expected refreshed project name, got %q", saved.ProjectName)
	}
	if saved.WorkspaceRoot != "/workspace/codeScope" {
		t.Fatalf("expected refreshed workspace root, got %q", saved.WorkspaceRoot)
	}
}

type testSessionStore struct {
	sessions map[string]session.Session
}

func newTestSessionStore() *testSessionStore {
	return &testSessionStore{sessions: make(map[string]session.Session)}
}

func (s *testSessionStore) Create(record session.Session) error {
	if _, exists := s.sessions[record.ID]; exists {
		return errors.New("conflict")
	}
	s.sessions[record.ID] = record
	return nil
}

func (s *testSessionStore) Get(id string) (session.Session, error) {
	record, exists := s.sessions[id]
	if !exists {
		return session.Session{}, errors.New("not found")
	}
	return record, nil
}

func (s *testSessionStore) Update(record session.Session) error {
	if _, exists := s.sessions[record.ID]; !exists {
		return errors.New("not found")
	}
	s.sessions[record.ID] = record
	return nil
}

type testEventStore struct {
	records []Record
}

func newTestEventStore() *testEventStore {
	return &testEventStore{}
}

func (s *testEventStore) Append(record Record) error {
	s.records = append(s.records, record)
	return nil
}

func (s *testEventStore) ListBySession(sessionID string) ([]Record, error) {
	var filtered []Record
	for _, record := range s.records {
		if record.SessionID == sessionID {
			filtered = append(filtered, record)
		}
	}
	return filtered, nil
}

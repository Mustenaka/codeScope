package event

import (
	"fmt"
	"log"
	"path"
	"sync/atomic"

	"codescope/server/internal/session"
)

type EventStore interface {
	Append(Record) error
	ListBySession(sessionID string) ([]Record, error)
}

type SessionReader interface {
	Create(session.Session) error
	Get(id string) (session.Session, error)
	Update(session.Session) error
}

type Service struct {
	events   EventStore
	sessions SessionReader
	broker   Broker
	idGen    func() string
}

func NewService(events EventStore, sessions SessionReader, broker Broker) *Service {
	return &Service{
		events:   events,
		sessions: sessions,
		broker:   broker,
		idGen: func() string {
			return fmt.Sprintf("event-%d", recordCounter())
		},
	}
}

func (s *Service) Ingest(message Message) (Record, error) {
	record, err := message.ToRecord(s.idGen())
	if err != nil {
		return Record{}, fmt.Errorf("validate event message: %w", err)
	}

	current, err := s.ensureSession(record)
	if err != nil {
		return Record{}, err
	}

	if err := s.events.Append(record); err != nil {
		return Record{}, fmt.Errorf("append event: %w", err)
	}

	if record.MessageType == MessageTypeHeartbeat {
		current.LastActivityAt = record.Timestamp
		current.UpdatedAt = record.Timestamp
		if err := s.sessions.Update(current); err != nil {
			return Record{}, fmt.Errorf("update heartbeat session %s: %w", current.ID, err)
		}
	} else if current.Status == session.StatusCreated {
		current.Status = session.StatusRunning
		if current.StartedAt.IsZero() {
			current.StartedAt = record.Timestamp
		}
		current.LastActivityAt = record.Timestamp
		current.UpdatedAt = record.Timestamp
		if err := s.sessions.Update(current); err != nil {
			return Record{}, fmt.Errorf("mark session running: %w", err)
		}
	} else {
		current.LastActivityAt = record.Timestamp
		current.UpdatedAt = record.Timestamp
		if err := s.sessions.Update(current); err != nil {
			return Record{}, fmt.Errorf("refresh session activity %s: %w", current.ID, err)
		}
	}

	s.broker.Publish(record)
	return record, nil
}

func (s *Service) ListBySession(sessionID string) ([]Record, error) {
	records, err := s.events.ListBySession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("list events for session %s: %w", sessionID, err)
	}
	return records, nil
}

var sequence atomic.Uint64

func recordCounter() uint64 {
	return sequence.Add(1)
}

func (s *Service) ensureSession(record Record) (session.Session, error) {
	current, err := s.sessions.Get(record.SessionID)
	if err == nil {
		updated := applyRecordMetadata(current, record)
		if updated != current {
			if err := s.sessions.Update(updated); err != nil {
				return session.Session{}, fmt.Errorf("refresh session metadata %s: %w", current.ID, err)
			}
			return updated, nil
		}
		return current, nil
	}

	created := session.Session{
		ID:            record.SessionID,
		ProjectName:   stringValue(record.Payload, "project_name", fallbackProjectName(record)),
		WorkspaceRoot: stringValue(record.Payload, "workspace_root", "."),
		MachineID:     stringValue(record.Payload, "machine_id", "unknown-machine"),
		Status:        session.StatusCreated,
		CreatedAt:     record.Timestamp,
		UpdatedAt:     record.Timestamp,
	}
	if err := s.sessions.Create(created); err != nil {
		current, reloadErr := s.sessions.Get(record.SessionID)
		if reloadErr == nil {
			return current, nil
		}
		return session.Session{}, fmt.Errorf("create session %s: %w", created.ID, err)
	}
	log.Printf("session auto-created: session_id=%s workspace_root=%s machine_id=%s", created.ID, created.WorkspaceRoot, created.MachineID)

	current, err = s.sessions.Get(record.SessionID)
	if err != nil {
		return session.Session{}, fmt.Errorf("reload session %s: %w", record.SessionID, err)
	}
	return current, nil
}

func stringValue(values map[string]any, key, fallback string) string {
	if value, ok := values[key].(string); ok && value != "" {
		return value
	}
	return fallback
}

func fallbackProjectName(record Record) string {
	if workspace := stringValue(record.Payload, "workspace_root", ""); workspace != "" && workspace != "." {
		return path.Base(workspace)
	}
	if agent := stringValue(record.Payload, "agent_name", ""); agent != "" {
		return agent
	}
	return record.SessionID
}

func applyRecordMetadata(current session.Session, record Record) session.Session {
	current.ProjectName = stringValue(record.Payload, "project_name", current.ProjectName)
	current.WorkspaceRoot = stringValue(record.Payload, "workspace_root", current.WorkspaceRoot)
	current.MachineID = stringValue(record.Payload, "machine_id", current.MachineID)
	return current
}

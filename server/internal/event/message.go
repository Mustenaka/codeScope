package event

import (
	"errors"
	"fmt"
	"time"
)

type Message struct {
	MessageID   string         `json:"message_id"`
	SessionID   string         `json:"session_id"`
	MessageType MessageType    `json:"message_type"`
	EventType   Type           `json:"event_type,omitempty"`
	CommandID   string         `json:"command_id,omitempty"`
	CommandType CommandType    `json:"command_type,omitempty"`
	Status      CommandStatus  `json:"status,omitempty"`
	Timestamp   string         `json:"timestamp"`
	Payload     map[string]any `json:"payload,omitempty"`
}

func (m Message) Validate() error {
	if m.MessageID == "" {
		return errors.New("message_id is required")
	}
	if m.SessionID == "" {
		return errors.New("session_id is required")
	}
	if !m.MessageType.Valid() {
		return fmt.Errorf("invalid message_type %q", m.MessageType)
	}
	if _, err := time.Parse(time.RFC3339, m.Timestamp); err != nil {
		return fmt.Errorf("invalid timestamp: %w", err)
	}
	if m.MessageType == MessageTypeEvent {
		if !m.EventType.Valid() {
			return fmt.Errorf("invalid event_type %q", m.EventType)
		}
	}
	if m.MessageType == MessageTypeCommand {
		if m.CommandID == "" {
			return errors.New("command_id is required")
		}
		if !m.CommandType.Valid() {
			return fmt.Errorf("invalid command_type %q", m.CommandType)
		}
	}
	if m.MessageType == MessageTypeHeartbeat {
		if m.EventType != "" && m.EventType != TypeHeartbeat {
			return fmt.Errorf("heartbeat message cannot use event_type %q", m.EventType)
		}
	}
	if m.MessageType == MessageTypeCommandResult {
		if m.CommandID == "" {
			return errors.New("command_id is required")
		}
		if !m.CommandType.Valid() {
			return fmt.Errorf("invalid command_type %q", m.CommandType)
		}
		if m.Status != CommandStatusSuccess && m.Status != CommandStatusFailed {
			return fmt.Errorf("invalid command status %q", m.Status)
		}
	}
	return nil
}

func (m Message) ToRecord(id string) (Record, error) {
	if err := m.Validate(); err != nil {
		return Record{}, err
	}

	ts, err := time.Parse(time.RFC3339, m.Timestamp)
	if err != nil {
		return Record{}, fmt.Errorf("parse timestamp: %w", err)
	}

	record := Record{
		ID:          id,
		MessageID:   m.MessageID,
		SessionID:   m.SessionID,
		MessageType: m.MessageType,
		EventType:   m.EventType,
		CommandID:   m.CommandID,
		CommandType: m.CommandType,
		Status:      m.Status,
		Timestamp:   ts.UTC(),
		Payload:     cloneMap(m.Payload),
		CreatedAt:   ts.UTC(),
	}
	if m.MessageType == MessageTypeHeartbeat {
		record.EventType = TypeHeartbeat
	}
	return record, nil
}

func cloneMap(input map[string]any) map[string]any {
	if len(input) == 0 {
		return map[string]any{}
	}
	output := make(map[string]any, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

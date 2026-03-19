package event

import "time"

type Type string

const (
	TypeAIOutput       Type = "ai_output"
	TypeTerminalOutput Type = "terminal_output"
	TypeCommand        Type = "command"
	TypeFileChange     Type = "file_change"
	TypeDiff           Type = "diff"
	TypeError          Type = "error"
	TypeHeartbeat      Type = "heartbeat"
)

type MessageType string

const (
	MessageTypeEvent         MessageType = "event"
	MessageTypeHeartbeat     MessageType = "heartbeat"
	MessageTypeCommand       MessageType = "command"
	MessageTypeCommandResult MessageType = "command_result"
)

type CommandType string

const (
	CommandTypeSendPrompt CommandType = "send_prompt"
)

type CommandStatus string

const (
	CommandStatusSuccess CommandStatus = "success"
	CommandStatusFailed  CommandStatus = "failed"
)

type Record struct {
	ID          string         `json:"id"`
	MessageID   string         `json:"message_id"`
	SessionID   string         `json:"session_id"`
	MessageType MessageType    `json:"message_type"`
	EventType   Type           `json:"event_type,omitempty"`
	CommandID   string         `json:"command_id,omitempty"`
	CommandType CommandType    `json:"command_type,omitempty"`
	Status      CommandStatus  `json:"status,omitempty"`
	Timestamp   time.Time      `json:"timestamp"`
	Payload     map[string]any `json:"payload,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

func (t Type) Valid() bool {
	switch t {
	case TypeAIOutput, TypeTerminalOutput, TypeCommand, TypeFileChange, TypeDiff, TypeError, TypeHeartbeat:
		return true
	default:
		return false
	}
}

func (m MessageType) Valid() bool {
	switch m {
	case MessageTypeEvent, MessageTypeHeartbeat, MessageTypeCommand, MessageTypeCommandResult:
		return true
	default:
		return false
	}
}

func (c CommandType) Valid() bool {
	switch c {
	case CommandTypeSendPrompt:
		return true
	default:
		return false
	}
}

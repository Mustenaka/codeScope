package session

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"path/filepath"
	"strings"
	"time"
)

const (
	MessageTypeEvent         = "event"
	MessageTypeHeartbeat     = "heartbeat"
	MessageTypeCommand       = "command"
	MessageTypeCommandResult = "command_result"
)

const (
	EventTypeTerminalOutput = "terminal_output"
	EventTypeAIOutput       = "ai_output"
	EventTypeCommand        = "command"
	EventTypeFileChange     = "file_change"
	EventTypeHeartbeat      = "heartbeat"
	EventTypeError          = "error"
)

const (
	ThreadStateRunning       = "running"
	ThreadStateWaitingPrompt = "waiting_prompt"
	ThreadStateWaitingReview = "waiting_review"
	ThreadStateCompleted     = "completed"
	ThreadStateBlocked       = "blocked"
)

const (
	CommandTypeSendPrompt = "send_prompt"
)

const (
	StatusSuccess = "success"
	StatusFailed  = "failed"
)

type Metadata struct {
	AgentName     string
	WorkspaceRoot string
	MachineID     string
	SessionID     string
}

type Event struct {
	Type    string
	Payload map[string]any
}

type Message struct {
	MessageID   string         `json:"message_id"`
	SessionID   string         `json:"session_id,omitempty"`
	MessageType string         `json:"message_type"`
	EventType   string         `json:"event_type,omitempty"`
	CommandID   string         `json:"command_id,omitempty"`
	CommandType string         `json:"command_type,omitempty"`
	Status      string         `json:"status,omitempty"`
	Timestamp   string         `json:"timestamp"`
	Payload     map[string]any `json:"payload,omitempty"`
}

func NewEventMessage(meta Metadata, eventType string, payload map[string]any, now time.Time) Message {
	enriched := clonePayload(payload)
	enrichSemanticPayload(enriched, meta, eventType, "")

	return Message{
		MessageID:   NewMessageID(),
		SessionID:   meta.SessionID,
		MessageType: MessageTypeEvent,
		EventType:   eventType,
		Timestamp:   formatTimestamp(now),
		Payload:     enriched,
	}
}

func NewHeartbeatMessage(meta Metadata, now time.Time) Message {
	payload := map[string]any{}
	enrichSemanticPayload(payload, meta, EventTypeHeartbeat, "")
	return Message{
		MessageID:   NewMessageID(),
		SessionID:   meta.SessionID,
		MessageType: MessageTypeHeartbeat,
		EventType:   EventTypeHeartbeat,
		Timestamp:   formatTimestamp(now),
		Payload:     payload,
	}
}

func NewCommandResultMessage(meta Metadata, commandID, commandType, status string, payload map[string]any, now time.Time) Message {
	enriched := clonePayload(payload)
	threadState := ""
	if commandType == CommandTypeSendPrompt && status == StatusFailed {
		threadState = ThreadStateWaitingPrompt
	}
	enrichSemanticPayload(enriched, meta, "", threadState)
	return Message{
		MessageID:   NewMessageID(),
		SessionID:   meta.SessionID,
		MessageType: MessageTypeCommandResult,
		CommandID:   commandID,
		CommandType: commandType,
		Status:      status,
		Timestamp:   formatTimestamp(now),
		Payload:     enriched,
	}
}

func NewMessageID() string {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return "msg-fallback"
	}
	return "msg-" + hex.EncodeToString(buf)
}

func clonePayload(payload map[string]any) map[string]any {
	if len(payload) == 0 {
		return map[string]any{}
	}
	cloned := make(map[string]any, len(payload))
	for key, value := range payload {
		cloned[key] = value
	}
	return cloned
}

func formatTimestamp(now time.Time) string {
	return now.UTC().Format(time.RFC3339Nano)
}

func enrichSemanticPayload(payload map[string]any, meta Metadata, eventType, overrideThreadState string) {
	projectID := stableProjectID(meta.MachineID, meta.WorkspaceRoot)
	projectName := projectDisplayName(meta.WorkspaceRoot)

	payload["agent_name"] = meta.AgentName
	payload["workspace_root"] = meta.WorkspaceRoot
	payload["machine_id"] = meta.MachineID
	payload["project_id"] = projectID
	if _, ok := payload["project_name"]; !ok || strings.TrimSpace(asString(payload["project_name"])) == "" {
		payload["project_name"] = projectName
	}
	if _, ok := payload["thread_id"]; !ok || strings.TrimSpace(asString(payload["thread_id"])) == "" {
		payload["thread_id"] = meta.SessionID
	}
	if _, ok := payload["source_session_id"]; !ok || strings.TrimSpace(asString(payload["source_session_id"])) == "" {
		payload["source_session_id"] = meta.SessionID
	}

	threadState := overrideThreadState
	if threadState == "" {
		threadState = defaultThreadStateForEvent(eventType, payload)
	}
	if threadState != "" {
		payload["thread_state"] = threadState
	}
}

func defaultThreadStateForEvent(eventType string, payload map[string]any) string {
	switch eventType {
	case EventTypeError:
		return ThreadStateBlocked
	case EventTypeHeartbeat, EventTypeTerminalOutput, EventTypeAIOutput, EventTypeCommand:
		return ThreadStateRunning
	default:
		if observed, ok := payload["observed"].(bool); ok && observed {
			return ThreadStateRunning
		}
		return ""
	}
}

func stableProjectID(machineID, workspaceRoot string) string {
	sum := sha1.Sum([]byte(machineID + "|" + filepath.ToSlash(workspaceRoot)))
	return "project-" + hex.EncodeToString(sum[:8])
}

func projectDisplayName(workspaceRoot string) string {
	clean := filepath.ToSlash(filepath.Clean(workspaceRoot))
	if clean == "." || clean == "/" {
		return clean
	}
	return filepath.Base(clean)
}

func asString(value any) string {
	text, _ := value.(string)
	return text
}

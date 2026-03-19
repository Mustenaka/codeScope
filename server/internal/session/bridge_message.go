package session

type BridgeMessage struct {
	MessageID   string         `json:"message_id"`
	SessionID   string         `json:"session_id"`
	MessageType string         `json:"message_type"`
	EventType   string         `json:"event_type,omitempty"`
	CommandID   string         `json:"command_id,omitempty"`
	CommandType string         `json:"command_type,omitempty"`
	Status      string         `json:"status,omitempty"`
	Timestamp   string         `json:"timestamp"`
	Payload     map[string]any `json:"payload,omitempty"`
}

const (
	BridgeMessageTypeCommand       = "command"
	BridgeMessageTypeCommandResult = "command_result"
	BridgeCommandTypeSendPrompt    = "send_prompt"
	BridgeCommandStatusSuccess     = "success"
	BridgeCommandStatusFailed      = "failed"
)

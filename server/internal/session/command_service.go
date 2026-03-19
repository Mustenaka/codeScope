package session

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type CommandTaskStore interface {
	Create(task CommandTask) error
	ListBySession(sessionID string) ([]CommandTask, error)
	Get(sessionID, taskID string) (CommandTask, error)
	Update(task CommandTask) error
}

type SessionReader interface {
	Get(id string) (Session, error)
}

type CommandResultRecorder interface {
	RecordCommandResult(message BridgeMessage) error
}

type PromptCommandInput struct {
	Content         string `json:"content" binding:"required"`
	ProjectID       string `json:"project_id,omitempty"`
	ThreadID        string `json:"thread_id,omitempty"`
	ThreadTitle     string `json:"thread_title,omitempty"`
	SourceSessionID string `json:"source_session_id,omitempty"`
}

type CommandService struct {
	sessions SessionReader
	tasks    CommandTaskStore
	bridges  *BridgeRegistry
	results  CommandResultRecorder
	now      func() time.Time
	idGen    func() string
}

func NewCommandService(sessions SessionReader, tasks CommandTaskStore, bridges *BridgeRegistry, results CommandResultRecorder) *CommandService {
	return &CommandService{
		sessions: sessions,
		tasks:    tasks,
		bridges:  bridges,
		results:  results,
		now:      time.Now().UTC,
		idGen: func() string {
			return fmt.Sprintf("cmd-%d", time.Now().UTC().UnixNano())
		},
	}
}

func (s *CommandService) Bridges() *BridgeRegistry {
	return s.bridges
}

func (s *CommandService) CreatePrompt(sessionID string, input PromptCommandInput) (CommandTask, error) {
	if _, err := s.sessions.Get(sessionID); err != nil {
		return CommandTask{}, fmt.Errorf("load session %s: %w", sessionID, err)
	}

	now := s.now()
	task := CommandTask{
		ID:        s.idGen(),
		SessionID: sessionID,
		TaskType:  BridgeCommandTypeSendPrompt,
		Payload:   promptPayload(input),
		Status:    CommandTaskPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.tasks.Create(task); err != nil {
		return CommandTask{}, fmt.Errorf("create command task: %w", err)
	}

	message := BridgeMessage{
		MessageID:   fmt.Sprintf("msg-%d", now.UnixNano()),
		SessionID:   sessionID,
		MessageType: BridgeMessageTypeCommand,
		CommandID:   task.ID,
		CommandType: BridgeCommandTypeSendPrompt,
		Timestamp:   now.Format(time.RFC3339Nano),
		Payload:     promptPayload(input),
	}

	if err := s.bridges.Send(sessionID, message); err != nil {
		task.Status = CommandTaskFailed
		task.Result = err.Error()
		task.UpdatedAt = s.now()
		_ = s.tasks.Update(task)
		log.Printf("command failed: session_id=%s command_id=%s reason=%v", sessionID, task.ID, err)
		return CommandTask{}, fmt.Errorf("send prompt command: %w", err)
	}

	task.Status = CommandTaskRunning
	task.UpdatedAt = s.now()
	if err := s.tasks.Update(task); err != nil {
		return CommandTask{}, fmt.Errorf("mark command task running: %w", err)
	}
	log.Printf("command dispatched: session_id=%s command_id=%s command_type=%s", sessionID, task.ID, BridgeCommandTypeSendPrompt)
	return task, nil
}

func promptPayload(input PromptCommandInput) map[string]any {
	payload := map[string]any{
		"content": input.Content,
	}
	if input.ProjectID != "" {
		payload["project_id"] = input.ProjectID
	}
	if input.ThreadID != "" {
		payload["thread_id"] = input.ThreadID
	}
	if input.ThreadTitle != "" {
		payload["thread_title"] = input.ThreadTitle
	}
	if input.SourceSessionID != "" {
		payload["source_session_id"] = input.SourceSessionID
	}
	return payload
}

func (s *CommandService) ListBySession(sessionID string) ([]CommandTask, error) {
	tasks, err := s.tasks.ListBySession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("list command tasks for session %s: %w", sessionID, err)
	}
	return tasks, nil
}

func (s *CommandService) Complete(message BridgeMessage) (CommandTask, error) {
	task, err := s.tasks.Get(message.SessionID, message.CommandID)
	if err != nil {
		return CommandTask{}, fmt.Errorf("load command task %s: %w", message.CommandID, err)
	}

	switch message.Status {
	case BridgeCommandStatusSuccess:
		task.Status = CommandTaskSuccess
	default:
		task.Status = CommandTaskFailed
	}
	task.UpdatedAt = s.now()
	task.Result = commandResultSummary(message.Payload)

	if err := s.tasks.Update(task); err != nil {
		return CommandTask{}, fmt.Errorf("update command task %s: %w", task.ID, err)
	}

	if err := s.results.RecordCommandResult(message); err != nil {
		return CommandTask{}, fmt.Errorf("ingest command result event: %w", err)
	}
	if task.Status == CommandTaskSuccess {
		log.Printf("command completed: session_id=%s command_id=%s", task.SessionID, task.ID)
	} else {
		log.Printf("command failed: session_id=%s command_id=%s result=%s", task.SessionID, task.ID, task.Result)
	}
	return task, nil
}

func commandResultSummary(payload map[string]any) string {
	if result, ok := payload["result"].(string); ok && result != "" {
		return result
	}
	if result, ok := payload["error"].(string); ok && result != "" {
		return result
	}
	if len(payload) == 0 {
		return ""
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(encoded)
}

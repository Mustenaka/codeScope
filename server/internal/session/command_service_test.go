package session

import (
	"errors"
	"testing"
	"time"
)

func TestCommandServiceCreatePromptDispatchesToBridge(t *testing.T) {
	sessionStore := &stubSessionReader{
		session: Session{
			ID:            "session-1",
			ProjectName:   "codeScope",
			WorkspaceRoot: "/workspace",
			MachineID:     "machine-1",
			Status:        StatusRunning,
		},
	}
	taskStore := newStubCommandTaskStore()
	results := &stubCommandResultRecorder{}
	registry := NewBridgeRegistry()
	outbound := make(chan any, 1)
	registry.Register("session-1", outbound)

	service := NewCommandService(sessionStore, taskStore, registry, results)
	now := time.Date(2026, 3, 17, 10, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	service.idGen = func() string { return "cmd-1" }

	task, err := service.CreatePrompt("session-1", PromptCommandInput{Content: "continue fixing tests"})
	if err != nil {
		t.Fatalf("create prompt: %v", err)
	}
	if task.Status != CommandTaskRunning {
		t.Fatalf("expected running command task, got %q", task.Status)
	}

	select {
	case raw := <-outbound:
		message, ok := raw.(BridgeMessage)
		if !ok {
			t.Fatalf("expected bridge message, got %#v", raw)
		}
		if message.CommandID != "cmd-1" {
			t.Fatalf("expected command id cmd-1, got %q", message.CommandID)
		}
		if message.CommandType != BridgeCommandTypeSendPrompt {
			t.Fatalf("expected send_prompt command, got %q", message.CommandType)
		}
	default:
		t.Fatalf("expected prompt command to be dispatched to bridge")
	}

	saved, err := taskStore.Get("session-1", "cmd-1")
	if err != nil {
		t.Fatalf("load saved task: %v", err)
	}
	if saved.Status != CommandTaskRunning {
		t.Fatalf("expected saved running task, got %q", saved.Status)
	}
}

func TestCommandServiceCompleteUpdatesTaskAndRecordsResult(t *testing.T) {
	taskStore := newStubCommandTaskStore()
	now := time.Date(2026, 3, 17, 10, 0, 0, 0, time.UTC)
	if err := taskStore.Create(CommandTask{
		ID:        "cmd-1",
		SessionID: "session-1",
		TaskType:  BridgeCommandTypeSendPrompt,
		Payload:   map[string]any{"content": "continue"},
		Status:    CommandTaskRunning,
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("seed command task: %v", err)
	}

	results := &stubCommandResultRecorder{}
	service := NewCommandService(&stubSessionReader{}, taskStore, NewBridgeRegistry(), results)
	service.now = func() time.Time { return now.Add(time.Second) }

	task, err := service.Complete(BridgeMessage{
		MessageID:   "msg-1",
		SessionID:   "session-1",
		MessageType: BridgeMessageTypeCommandResult,
		CommandID:   "cmd-1",
		CommandType: BridgeCommandTypeSendPrompt,
		Status:      BridgeCommandStatusFailed,
		Timestamp:   now.Add(time.Second).Format(time.RFC3339),
		Payload: map[string]any{
			"accepted": false,
			"error":    "managed process stdin not ready",
		},
	})
	if err != nil {
		t.Fatalf("complete command: %v", err)
	}
	if task.Status != CommandTaskFailed {
		t.Fatalf("expected failed task, got %q", task.Status)
	}
	if task.Result != "managed process stdin not ready" {
		t.Fatalf("expected failure result to be recorded, got %q", task.Result)
	}
	if len(results.messages) != 1 {
		t.Fatalf("expected command result recorder to receive one message, got %d", len(results.messages))
	}
}

type stubSessionReader struct {
	session Session
	err     error
}

func (s *stubSessionReader) Get(id string) (Session, error) {
	if s.err != nil {
		return Session{}, s.err
	}
	if s.session.ID == "" {
		return Session{}, errors.New("not found")
	}
	return s.session, nil
}

type stubCommandTaskStore struct {
	tasks map[string]map[string]CommandTask
}

func newStubCommandTaskStore() *stubCommandTaskStore {
	return &stubCommandTaskStore{tasks: make(map[string]map[string]CommandTask)}
}

func (s *stubCommandTaskStore) Create(task CommandTask) error {
	if _, ok := s.tasks[task.SessionID]; !ok {
		s.tasks[task.SessionID] = make(map[string]CommandTask)
	}
	s.tasks[task.SessionID][task.ID] = task
	return nil
}

func (s *stubCommandTaskStore) ListBySession(sessionID string) ([]CommandTask, error) {
	bucket := s.tasks[sessionID]
	result := make([]CommandTask, 0, len(bucket))
	for _, task := range bucket {
		result = append(result, task)
	}
	return result, nil
}

func (s *stubCommandTaskStore) Get(sessionID, taskID string) (CommandTask, error) {
	bucket, ok := s.tasks[sessionID]
	if !ok {
		return CommandTask{}, errors.New("not found")
	}
	task, ok := bucket[taskID]
	if !ok {
		return CommandTask{}, errors.New("not found")
	}
	return task, nil
}

func (s *stubCommandTaskStore) Update(task CommandTask) error {
	if _, ok := s.tasks[task.SessionID]; !ok {
		return errors.New("not found")
	}
	s.tasks[task.SessionID][task.ID] = task
	return nil
}

type stubCommandResultRecorder struct {
	messages []BridgeMessage
}

func (s *stubCommandResultRecorder) RecordCommandResult(message BridgeMessage) error {
	s.messages = append(s.messages, message)
	return nil
}

package project

import (
	"errors"
	"strings"
	"testing"
	"time"

	"codescope/server/internal/event"
	"codescope/server/internal/session"
)

var errTestNotFound = errors.New("not found")

type testSessionReader struct {
	sessions []session.Session
}

func (r testSessionReader) List() ([]session.Session, error) {
	return r.sessions, nil
}

func (r testSessionReader) Get(id string) (session.Session, error) {
	for _, item := range r.sessions {
		if item.ID == id {
			return item, nil
		}
	}
	return session.Session{}, errTestNotFound
}

type testEventReader struct {
	bySession map[string][]event.Record
}

func (r testEventReader) ListBySession(sessionID string) ([]event.Record, error) {
	return r.bySession[sessionID], nil
}

func TestServiceListProjectsGroupsSessionsIntoProjects(t *testing.T) {
	base := time.Date(2026, 3, 19, 9, 0, 0, 0, time.UTC)
	service := NewService(
		testSessionReader{sessions: []session.Session{
			{
				ID:             "session-1",
				ProjectName:    "codeScope",
				WorkspaceRoot:  "/workspace/codeScope",
				MachineID:      "machine-1",
				Status:         session.StatusRunning,
				LastActivityAt: base.Add(2 * time.Minute),
				CreatedAt:      base,
				UpdatedAt:      base.Add(2 * time.Minute),
			},
			{
				ID:             "session-2",
				ProjectName:    "codeScope",
				WorkspaceRoot:  "/workspace/codeScope",
				MachineID:      "machine-1",
				Status:         session.StatusCreated,
				BridgeOnline:   true,
				LastActivityAt: base.Add(3 * time.Minute),
				CreatedAt:      base.Add(time.Minute),
				UpdatedAt:      base.Add(3 * time.Minute),
			},
			{
				ID:             "session-3",
				ProjectName:    "other",
				WorkspaceRoot:  "/workspace/other",
				MachineID:      "machine-1",
				Status:         session.StatusStopped,
				LastActivityAt: base.Add(4 * time.Minute),
				CreatedAt:      base.Add(2 * time.Minute),
				UpdatedAt:      base.Add(4 * time.Minute),
			},
		}},
		testEventReader{bySession: map[string][]event.Record{}},
	)

	projects, err := service.ListProjects()
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projects))
	}

	var found Project
	for _, item := range projects {
		if item.Name == "codeScope" {
			found = item
			break
		}
	}
	if found.ID == "" {
		t.Fatalf("expected to find codeScope project in %#v", projects)
	}
	if found.ThreadCount != 0 {
		t.Fatalf("expected 0 visible threads without readable history, got %d", found.ThreadCount)
	}
	if found.RunningThreadCount != 0 {
		t.Fatalf("expected 0 running visible threads, got %d", found.RunningThreadCount)
	}
}

func TestServiceListThreadsDerivesStateAndSummary(t *testing.T) {
	base := time.Date(2026, 3, 19, 9, 0, 0, 0, time.UTC)
	sessions := []session.Session{
		{
			ID:             "session-1",
			ProjectName:    "codeScope",
			WorkspaceRoot:  "/workspace/codeScope",
			MachineID:      "machine-1",
			Status:         session.StatusRunning,
			LastActivityAt: base.Add(2 * time.Minute),
			StartedAt:      base.Add(time.Minute),
			CreatedAt:      base,
			UpdatedAt:      base.Add(2 * time.Minute),
		},
		{
			ID:             "session-2",
			ProjectName:    "codeScope",
			WorkspaceRoot:  "/workspace/codeScope",
			MachineID:      "machine-1",
			Status:         session.StatusCreated,
			BridgeOnline:   true,
			LastActivityAt: base.Add(3 * time.Minute),
			CreatedAt:      base.Add(time.Minute),
			UpdatedAt:      base.Add(3 * time.Minute),
		},
	}
	service := NewService(
		testSessionReader{sessions: sessions},
		testEventReader{bySession: map[string][]event.Record{
			"session-1": {
				{
					ID:          "event-1",
					SessionID:   "session-1",
					MessageType: event.MessageTypeEvent,
					EventType:   event.TypeAIOutput,
					Timestamp:   base.Add(2 * time.Minute),
					Payload:     map[string]any{"content": "Implemented server API", "agent_name": "codex"},
				},
			},
			"session-2": {
				{
					ID:          "event-2",
					SessionID:   "session-2",
					MessageType: event.MessageTypeCommandResult,
					CommandType: event.CommandTypeSendPrompt,
					Status:      event.CommandStatusFailed,
					Timestamp:   base.Add(3 * time.Minute),
					Payload:     map[string]any{"accepted": false, "error": "side-channel mode"},
				},
			},
		}},
	)

	projectID := ProjectID(sessions[0])
	threads, err := service.ListThreads(projectID)
	if err != nil {
		t.Fatalf("list threads: %v", err)
	}
	if len(threads) != 2 {
		t.Fatalf("expected 2 threads, got %d", len(threads))
	}

	byID := make(map[string]Thread, len(threads))
	for _, item := range threads {
		byID[item.ID] = item
	}
	if byID["session-1"].Summary != "Implemented server API" {
		t.Fatalf("expected assistant summary, got %q", byID["session-1"].Summary)
	}
	if byID["session-1"].Status != ThreadStateRunning {
		t.Fatalf("expected running state, got %q", byID["session-1"].Status)
	}
	if byID["session-2"].Status != ThreadStateWaitingPrompt {
		t.Fatalf("expected waiting_prompt state, got %q", byID["session-2"].Status)
	}
}

func TestServiceListThreadsFallsBackToLatestPromptWhenAssistantSummaryMissing(t *testing.T) {
	base := time.Date(2026, 3, 19, 9, 0, 0, 0, time.UTC)
	sessionRecord := session.Session{
		ID:             "session-1",
		ProjectName:    "codeScope",
		WorkspaceRoot:  "/workspace/codeScope",
		MachineID:      "machine-1",
		Status:         session.StatusRunning,
		LastActivityAt: base.Add(2 * time.Minute),
		StartedAt:      base,
		CreatedAt:      base,
		UpdatedAt:      base.Add(2 * time.Minute),
	}
	service := NewService(
		testSessionReader{sessions: []session.Session{sessionRecord}},
		testEventReader{bySession: map[string][]event.Record{
			"session-1": {
				{
					ID:          "event-command",
					SessionID:   "session-1",
					MessageType: event.MessageTypeCommand,
					CommandType: event.CommandTypeSendPrompt,
					Timestamp:   base.Add(2 * time.Minute),
					Payload:     map[string]any{"content": "Please summarize the repo"},
				},
			},
		}},
	)

	threads, err := service.ListThreads(ProjectID(sessionRecord))
	if err != nil {
		t.Fatalf("list threads: %v", err)
	}
	if len(threads) != 1 {
		t.Fatalf("expected 1 thread, got %d", len(threads))
	}
	if threads[0].Summary != "Please summarize the repo" {
		t.Fatalf("expected prompt fallback summary, got %q", threads[0].Summary)
	}
}

func TestServiceListThreadsUsesPromptDerivedTitleInsteadOfProjectName(t *testing.T) {
	base := time.Date(2026, 3, 19, 9, 0, 0, 0, time.UTC)
	sessionRecord := session.Session{
		ID:             "session-1",
		ProjectName:    "codeScope",
		WorkspaceRoot:  "/workspace/codeScope",
		MachineID:      "machine-1",
		Status:         session.StatusRunning,
		LastActivityAt: base.Add(2 * time.Minute),
		StartedAt:      base,
		CreatedAt:      base,
		UpdatedAt:      base.Add(2 * time.Minute),
	}
	service := NewService(
		testSessionReader{sessions: []session.Session{sessionRecord}},
		testEventReader{bySession: map[string][]event.Record{
			"session-1": {
				{
					ID:          "event-command",
					SessionID:   "session-1",
					MessageType: event.MessageTypeCommand,
					CommandType: event.CommandTypeSendPrompt,
					Timestamp:   base.Add(2 * time.Minute),
					Payload: map[string]any{
						"content": "Fix the bridge project-name derivation so the project list stops showing subdirectory names.",
					},
				},
			},
		}},
	)

	threads, err := service.ListThreads(ProjectID(sessionRecord))
	if err != nil {
		t.Fatalf("list threads: %v", err)
	}
	if len(threads) != 1 {
		t.Fatalf("expected 1 thread, got %d", len(threads))
	}
	if threads[0].Title == "codeScope" {
		t.Fatalf("expected thread title to avoid project-name fallback, got %q", threads[0].Title)
	}
	if !strings.HasPrefix(threads[0].Title, "Fix the bridge project-name derivation") {
		t.Fatalf("expected prompt-derived title, got %q", threads[0].Title)
	}
}

func TestServiceListThreadsFallsBackToGenericThreadTitleWhenNoReadableContent(t *testing.T) {
	base := time.Date(2026, 3, 19, 9, 0, 0, 0, time.UTC)
	sessionRecord := session.Session{
		ID:             "session-abcdef12",
		ProjectName:    "codeScope",
		WorkspaceRoot:  "/workspace/codeScope",
		MachineID:      "machine-1",
		Status:         session.StatusRunning,
		LastActivityAt: base.Add(2 * time.Minute),
		StartedAt:      base,
		CreatedAt:      base,
		UpdatedAt:      base.Add(2 * time.Minute),
	}
	service := NewService(
		testSessionReader{sessions: []session.Session{sessionRecord}},
		testEventReader{bySession: map[string][]event.Record{
			"session-abcdef12": {
				{
					ID:          "event-debug",
					SessionID:   "session-abcdef12",
					MessageType: event.MessageTypeEvent,
					EventType:   event.TypeTerminalOutput,
					Timestamp:   base.Add(2 * time.Minute),
					Payload:     map[string]any{"content": "[bridge] observing", "observed": true},
				},
			},
		}},
	)

	threads, err := service.ListThreads(ProjectID(sessionRecord))
	if err != nil {
		t.Fatalf("list threads: %v", err)
	}
	if len(threads) != 0 {
		t.Fatalf("expected debug-only thread to be filtered, got %d", len(threads))
	}
}

func TestServicePrefersBridgeSemanticThreadFields(t *testing.T) {
	base := time.Date(2026, 3, 19, 9, 0, 0, 0, time.UTC)
	sessionRecord := session.Session{
		ID:             "session-9",
		ProjectName:    "codeScope",
		WorkspaceRoot:  "/workspace/codeScope",
		MachineID:      "machine-1",
		Status:         session.StatusRunning,
		LastActivityAt: base.Add(2 * time.Minute),
		CreatedAt:      base,
		UpdatedAt:      base.Add(2 * time.Minute),
	}
	service := NewService(
		testSessionReader{sessions: []session.Session{sessionRecord}},
		testEventReader{bySession: map[string][]event.Record{
			"session-9": {
				{
					ID:          "event-9",
					SessionID:   "session-9",
					MessageType: event.MessageTypeEvent,
					EventType:   event.TypeTerminalOutput,
					Timestamp:   base.Add(2 * time.Minute),
					Payload: map[string]any{
						"content":      "waiting for next prompt",
						"thread_id":    "thread-semantic-1",
						"thread_state": "waiting_prompt",
					},
				},
			},
		}},
	)

	threads, err := service.ListThreads(ProjectID(sessionRecord))
	if err != nil {
		t.Fatalf("list threads: %v", err)
	}
	if len(threads) != 1 {
		t.Fatalf("expected 1 thread, got %d", len(threads))
	}
	if threads[0].ID != "thread-semantic-1" {
		t.Fatalf("expected semantic thread id, got %q", threads[0].ID)
	}
	if threads[0].Status != ThreadStateWaitingPrompt {
		t.Fatalf("expected semantic waiting_prompt state, got %q", threads[0].Status)
	}
}

func TestServiceListMessagesFiltersDebugEventsAndMapsRoles(t *testing.T) {
	base := time.Date(2026, 3, 19, 9, 0, 0, 0, time.UTC)
	sessionRecord := session.Session{
		ID:             "session-1",
		ProjectName:    "codeScope",
		WorkspaceRoot:  "/workspace/codeScope",
		MachineID:      "machine-1",
		Status:         session.StatusRunning,
		LastActivityAt: base.Add(2 * time.Minute),
		CreatedAt:      base,
		UpdatedAt:      base.Add(2 * time.Minute),
	}
	service := NewService(
		testSessionReader{sessions: []session.Session{sessionRecord}},
		testEventReader{bySession: map[string][]event.Record{
			"session-1": {
				{
					ID:          "event-observed",
					SessionID:   "session-1",
					MessageType: event.MessageTypeEvent,
					EventType:   event.TypeTerminalOutput,
					Timestamp:   base,
					Payload:     map[string]any{"content": "[bridge] observing", "observed": true},
				},
				{
					ID:          "event-command",
					SessionID:   "session-1",
					MessageType: event.MessageTypeCommand,
					CommandID:   "cmd-1",
					CommandType: event.CommandTypeSendPrompt,
					Timestamp:   base.Add(time.Minute),
					Payload:     map[string]any{"content": "Please summarize the repo"},
				},
				{
					ID:          "event-ai",
					SessionID:   "session-1",
					MessageType: event.MessageTypeEvent,
					EventType:   event.TypeAIOutput,
					Timestamp:   base.Add(2 * time.Minute),
					Payload:     map[string]any{"content": "Repo summary here"},
				},
			},
		}},
	)

	messages, err := service.ListMessages(ThreadID(sessionRecord))
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}
	if messages[0].Role != RoleUser {
		t.Fatalf("expected first role user, got %q", messages[0].Role)
	}
	if messages[1].Role != RoleAssistant {
		t.Fatalf("expected second role assistant, got %q", messages[1].Role)
	}
	if messages[1].Content != "Repo summary here" {
		t.Fatalf("expected assistant content, got %q", messages[1].Content)
	}
}

func TestServiceListMessagesMapsRealCapturedEventCommandAsUserMessage(t *testing.T) {
	base := time.Date(2026, 3, 19, 9, 0, 0, 0, time.UTC)
	sessionRecord := session.Session{
		ID:             "session-real-1",
		ProjectName:    "codeScope",
		WorkspaceRoot:  "/workspace/codeScope",
		MachineID:      "machine-1",
		Status:         session.StatusRunning,
		LastActivityAt: base.Add(2 * time.Minute),
		CreatedAt:      base,
		UpdatedAt:      base.Add(2 * time.Minute),
	}
	service := NewService(
		testSessionReader{sessions: []session.Session{sessionRecord}},
		testEventReader{bySession: map[string][]event.Record{
			"session-real-1": {
				{
					ID:          "event-user-real",
					SessionID:   "session-real-1",
					MessageType: event.MessageTypeEvent,
					EventType:   event.TypeCommand,
					Timestamp:   base.Add(time.Minute),
					Payload: map[string]any{
						"role":    "user",
						"content": "Please implement real prompt capture.",
					},
				},
				{
					ID:          "event-assistant-real",
					SessionID:   "session-real-1",
					MessageType: event.MessageTypeEvent,
					EventType:   event.TypeAIOutput,
					Timestamp:   base.Add(2 * time.Minute),
					Payload: map[string]any{
						"role":    "assistant",
						"content": "Implemented real prompt capture.",
					},
				},
			},
		}},
	)

	messages, err := service.ListMessages(ThreadID(sessionRecord))
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}
	if messages[0].Role != RoleUser {
		t.Fatalf("expected first role user, got %q", messages[0].Role)
	}
	if messages[0].Content != "Please implement real prompt capture." {
		t.Fatalf("expected captured user content, got %q", messages[0].Content)
	}
	if messages[1].Role != RoleAssistant {
		t.Fatalf("expected second role assistant, got %q", messages[1].Role)
	}
}

func TestServiceListThreadsAggregatesSessionsWithSameSemanticThreadID(t *testing.T) {
	base := time.Date(2026, 3, 19, 9, 0, 0, 0, time.UTC)
	sessions := []session.Session{
		{
			ID:             "session-1",
			ProjectName:    "codeScope",
			WorkspaceRoot:  "/workspace/codeScope",
			MachineID:      "machine-1",
			Status:         session.StatusRunning,
			LastActivityAt: base.Add(time.Minute),
			StartedAt:      base,
			CreatedAt:      base,
			UpdatedAt:      base.Add(time.Minute),
		},
		{
			ID:             "session-2",
			ProjectName:    "codeScope",
			WorkspaceRoot:  "/workspace/codeScope",
			MachineID:      "machine-1",
			Status:         session.StatusRunning,
			LastActivityAt: base.Add(2 * time.Minute),
			StartedAt:      base.Add(90 * time.Second),
			CreatedAt:      base.Add(90 * time.Second),
			UpdatedAt:      base.Add(2 * time.Minute),
		},
	}
	service := NewService(
		testSessionReader{sessions: sessions},
		testEventReader{bySession: map[string][]event.Record{
			"session-1": {
				{
					ID:          "event-user-1",
					SessionID:   "session-1",
					MessageType: event.MessageTypeEvent,
					EventType:   event.TypeCommand,
					Timestamp:   base.Add(time.Minute),
					Payload: map[string]any{
						"thread_id":    "thread-semantic-1",
						"thread_title": "cli",
						"role":         "user",
						"content":      "cli",
					},
				},
			},
			"session-2": {
				{
					ID:          "event-ai-1",
					SessionID:   "session-2",
					MessageType: event.MessageTypeEvent,
					EventType:   event.TypeAIOutput,
					Timestamp:   base.Add(2 * time.Minute),
					Payload: map[string]any{
						"thread_id":    "thread-semantic-1",
						"thread_title": "cli",
						"content":      "assistant reply",
					},
				},
			},
		}},
	)

	threads, err := service.ListThreads(ProjectID(sessions[0]))
	if err != nil {
		t.Fatalf("list threads: %v", err)
	}
	if len(threads) != 1 {
		t.Fatalf("expected merged single thread, got %d", len(threads))
	}
	if threads[0].ID != "thread-semantic-1" {
		t.Fatalf("expected semantic thread id, got %q", threads[0].ID)
	}

	messages, err := service.ListMessages("thread-semantic-1")
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("expected merged messages from both sessions, got %d", len(messages))
	}
}

func TestServiceListThreadsFiltersDebugOnlyGhostThreads(t *testing.T) {
	base := time.Date(2026, 3, 19, 9, 0, 0, 0, time.UTC)
	sessionRecord := session.Session{
		ID:             "session-ghost-1",
		ProjectName:    "codeScope",
		WorkspaceRoot:  "/workspace/codeScope",
		MachineID:      "machine-1",
		Status:         session.StatusRunning,
		LastActivityAt: base.Add(time.Minute),
		CreatedAt:      base,
		UpdatedAt:      base.Add(time.Minute),
	}
	service := NewService(
		testSessionReader{sessions: []session.Session{sessionRecord}},
		testEventReader{bySession: map[string][]event.Record{
			"session-ghost-1": {
				{
					ID:          "event-debug-1",
					SessionID:   "session-ghost-1",
					MessageType: event.MessageTypeEvent,
					EventType:   event.TypeTerminalOutput,
					Timestamp:   base.Add(time.Minute),
					Payload: map[string]any{
						"content":  "[bridge] observing codex session",
						"observed": true,
					},
				},
			},
		}},
	)

	threads, err := service.ListThreads(ProjectID(sessionRecord))
	if err != nil {
		t.Fatalf("list threads: %v", err)
	}
	if len(threads) != 0 {
		t.Fatalf("expected ghost thread to be filtered, got %d", len(threads))
	}
}

func TestServiceListThreadsMarksOldInactiveThreadAsWaitingPrompt(t *testing.T) {
	base := time.Now().UTC().Add(-2 * activeThreadWindow)
	sessionRecord := session.Session{
		ID:             "session-old-1",
		ProjectName:    "codeScope",
		WorkspaceRoot:  "/workspace/codeScope",
		MachineID:      "machine-1",
		Status:         session.StatusRunning,
		LastActivityAt: base,
		StartedAt:      base.Add(-10 * time.Minute),
		CreatedAt:      base.Add(-10 * time.Minute),
		UpdatedAt:      base,
	}
	service := NewService(
		testSessionReader{sessions: []session.Session{sessionRecord}},
		testEventReader{bySession: map[string][]event.Record{
			"session-old-1": {
				{
					ID:          "event-user-1",
					SessionID:   "session-old-1",
					MessageType: event.MessageTypeEvent,
					EventType:   event.TypeCommand,
					Timestamp:   base,
					Payload: map[string]any{
						"role":    "user",
						"content": "old prompt",
					},
				},
			},
		}},
	)

	threads, err := service.ListThreads(ProjectID(sessionRecord))
	if err != nil {
		t.Fatalf("list threads: %v", err)
	}
	if len(threads) != 1 {
		t.Fatalf("expected 1 thread, got %d", len(threads))
	}
	if threads[0].Status != ThreadStateWaitingPrompt {
		t.Fatalf("expected stale thread to become waiting_prompt, got %q", threads[0].Status)
	}
}

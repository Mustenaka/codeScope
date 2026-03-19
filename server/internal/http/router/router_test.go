package router_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"codescope/server/internal/app"
	"codescope/server/internal/config"
	"codescope/server/internal/event"
	"codescope/server/internal/http/router"
	"codescope/server/internal/session"
)

func newTestRuntime(t *testing.T) (http.Handler, app.Dependencies) {
	t.Helper()

	container := app.NewWithConfig(config.Config{
		AppName: "codeScope Server",
	})

	deps := container.Dependencies()
	return router.New(router.Dependencies{
		Config:         deps.Config,
		SessionService: deps.SessionService,
		EventService:   deps.EventService,
		EventHub:       deps.EventHub,
		FileService:    deps.FileService,
		ProjectService: deps.ProjectService,
		PromptService:  deps.PromptService,
		CommandService: deps.CommandService,
	}), deps
}

func newTestEngine(t *testing.T) http.Handler {
	engine, _ := newTestRuntime(t)
	return engine
}

func TestHealthRoute(t *testing.T) {
	engine := newTestEngine(t)

	request := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var payload map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload["status"] != "ok" {
		t.Fatalf("expected status ok, got %q", payload["status"])
	}
}

func TestSessionRoutes(t *testing.T) {
	engine := newTestEngine(t)

	body := bytes.NewBufferString(`{"id":"session-1","project_name":"codeScope","workspace_root":"/workspace","machine_id":"machine-1"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/sessions", body)
	createReq.Header.Set("Content-Type", "application/json")
	createResp := httptest.NewRecorder()
	engine.ServeHTTP(createResp, createReq)

	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d", http.StatusCreated, createResp.Code)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	listResp := httptest.NewRecorder()
	engine.ServeHTTP(listResp, listReq)

	if listResp.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d", http.StatusOK, listResp.Code)
	}

	var sessions []map[string]any
	if err := json.Unmarshal(listResp.Body.Bytes(), &sessions); err != nil {
		t.Fatalf("decode sessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}

	detailReq := httptest.NewRequest(http.MethodGet, "/api/sessions/session-1", nil)
	detailResp := httptest.NewRecorder()
	engine.ServeHTTP(detailResp, detailReq)

	if detailResp.Code != http.StatusOK {
		t.Fatalf("expected detail status %d, got %d", http.StatusOK, detailResp.Code)
	}
}

func TestSessionEventsRoute(t *testing.T) {
	engine := newTestEngine(t)

	createReq := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewBufferString(`{"id":"session-1","project_name":"codeScope","workspace_root":"/workspace","machine_id":"machine-1"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createResp := httptest.NewRecorder()
	engine.ServeHTTP(createResp, createReq)

	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d", http.StatusCreated, createResp.Code)
	}

	updateReq := httptest.NewRequest(http.MethodPatch, "/api/sessions/session-1/status", bytes.NewBufferString(`{"status":"running"}`))
	updateReq.Header.Set("Content-Type", "application/json")
	updateResp := httptest.NewRecorder()
	engine.ServeHTTP(updateResp, updateReq)

	if updateResp.Code != http.StatusOK {
		t.Fatalf("expected update status %d, got %d", http.StatusOK, updateResp.Code)
	}

	eventsReq := httptest.NewRequest(http.MethodGet, "/api/sessions/session-1/events", nil)
	eventsResp := httptest.NewRecorder()
	engine.ServeHTTP(eventsResp, eventsReq)

	if eventsResp.Code != http.StatusOK {
		t.Fatalf("expected events status %d, got %d", http.StatusOK, eventsResp.Code)
	}
}

func TestProjectThreadRoutesHideDebugOnlySessionsFromReleaseView(t *testing.T) {
	engine := newTestEngine(t)

	createReq := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewBufferString(`{"id":"session-1","project_name":"codeScope","workspace_root":"/workspace","machine_id":"machine-1"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createResp := httptest.NewRecorder()
	engine.ServeHTTP(createResp, createReq)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d", http.StatusCreated, createResp.Code)
	}
	projectsReq := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	projectsResp := httptest.NewRecorder()
	engine.ServeHTTP(projectsResp, projectsReq)
	if projectsResp.Code != http.StatusOK {
		t.Fatalf("expected projects status %d, got %d", http.StatusOK, projectsResp.Code)
	}

	var projects []map[string]any
	if err := json.Unmarshal(projectsResp.Body.Bytes(), &projects); err != nil {
		t.Fatalf("decode projects: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}

	projectID, ok := projects[0]["id"].(string)
	if !ok || projectID == "" {
		t.Fatalf("expected project id, got %#v", projects[0]["id"])
	}

	threadsReq := httptest.NewRequest(http.MethodGet, "/api/projects/"+projectID+"/threads", nil)
	threadsResp := httptest.NewRecorder()
	engine.ServeHTTP(threadsResp, threadsReq)
	if threadsResp.Code != http.StatusOK {
		t.Fatalf("expected threads status %d, got %d", http.StatusOK, threadsResp.Code)
	}

	var threads []map[string]any
	if err := json.Unmarshal(threadsResp.Body.Bytes(), &threads); err != nil {
		t.Fatalf("decode threads: %v", err)
	}
	if len(threads) != 0 {
		t.Fatalf("expected 0 visible threads without readable history, got %d", len(threads))
	}
}

func TestThreadPromptRouteDispatchesThroughReleaseThreadID(t *testing.T) {
	engine, deps := newTestRuntime(t)

	createReq := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewBufferString(`{"id":"session-1","project_name":"codeScope","workspace_root":"/workspace","machine_id":"machine-1"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createResp := httptest.NewRecorder()
	engine.ServeHTTP(createResp, createReq)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d", http.StatusCreated, createResp.Code)
	}

	updateReq := httptest.NewRequest(http.MethodPatch, "/api/sessions/session-1/status", bytes.NewBufferString(`{"status":"running"}`))
	updateReq.Header.Set("Content-Type", "application/json")
	updateResp := httptest.NewRecorder()
	engine.ServeHTTP(updateResp, updateReq)
	if updateResp.Code != http.StatusOK {
		t.Fatalf("expected update status %d, got %d", http.StatusOK, updateResp.Code)
	}
	if _, err := deps.EventService.Ingest(event.Message{
		MessageID:   "event-thread-visible",
		SessionID:   "session-1",
		MessageType: event.MessageTypeEvent,
		EventType:   event.TypeAIOutput,
		Timestamp:   time.Date(2026, 3, 19, 9, 1, 0, 0, time.UTC).Format(time.RFC3339Nano),
		Payload: map[string]any{
			"content": "assistant reply",
		},
	}); err != nil {
		t.Fatalf("seed visible thread: %v", err)
	}

	projectsReq := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	projectsResp := httptest.NewRecorder()
	engine.ServeHTTP(projectsResp, projectsReq)
	if projectsResp.Code != http.StatusOK {
		t.Fatalf("expected projects status %d, got %d", http.StatusOK, projectsResp.Code)
	}

	var projects []map[string]any
	if err := json.Unmarshal(projectsResp.Body.Bytes(), &projects); err != nil {
		t.Fatalf("decode projects: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	projectID, _ := projects[0]["id"].(string)

	threadsReq := httptest.NewRequest(http.MethodGet, "/api/projects/"+projectID+"/threads", nil)
	threadsResp := httptest.NewRecorder()
	engine.ServeHTTP(threadsResp, threadsReq)
	if threadsResp.Code != http.StatusOK {
		t.Fatalf("expected threads status %d, got %d", http.StatusOK, threadsResp.Code)
	}

	var threads []map[string]any
	if err := json.Unmarshal(threadsResp.Body.Bytes(), &threads); err != nil {
		t.Fatalf("decode threads: %v", err)
	}
	if len(threads) != 1 {
		t.Fatalf("expected 1 thread, got %d", len(threads))
	}
	threadID, _ := threads[0]["id"].(string)
	if threadID == "" {
		t.Fatalf("expected thread id in %#v", threads[0])
	}
	if _, _, err := deps.SessionService.EnsureBridgeSession("session-1", session.BridgeMetadata{
		ProjectName:   "codeScope",
		WorkspaceRoot: "/workspace",
		MachineID:     "machine-1",
	}, time.Date(2026, 3, 19, 9, 2, 0, 0, time.UTC)); err != nil {
		t.Fatalf("ensure bridge session: %v", err)
	}
	unregister := deps.CommandService.Bridges().Register("session-1", make(chan any, 1))
	defer unregister()

	promptReq := httptest.NewRequest(http.MethodPost, "/api/threads/"+threadID+"/commands/prompt", strings.NewReader(`{"content":"continue from release thread route"}`))
	promptReq.Header.Set("Content-Type", "application/json")
	promptResp := httptest.NewRecorder()
	engine.ServeHTTP(promptResp, promptReq)

	if promptResp.Code != http.StatusCreated {
		t.Fatalf("expected thread prompt status %d, got %d body=%s", http.StatusCreated, promptResp.Code, promptResp.Body.String())
	}

	var task map[string]any
	if err := json.Unmarshal(promptResp.Body.Bytes(), &task); err != nil {
		t.Fatalf("decode task: %v", err)
	}
	if task["session_id"] != "session-1" {
		t.Fatalf("expected backing session session-1, got %#v", task["session_id"])
	}
}

func TestProjectFileRoutesExposeReleaseProjectWorkspace(t *testing.T) {
	engine := newTestEngine(t)

	createReq := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewBufferString(`{"id":"session-1","project_name":"codeScope","workspace_root":"D:/Work/Code/Cross/codeScope/server","machine_id":"machine-1"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createResp := httptest.NewRecorder()
	engine.ServeHTTP(createResp, createReq)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d", http.StatusCreated, createResp.Code)
	}

	projectsReq := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	projectsResp := httptest.NewRecorder()
	engine.ServeHTTP(projectsResp, projectsReq)
	if projectsResp.Code != http.StatusOK {
		t.Fatalf("expected projects status %d, got %d", http.StatusOK, projectsResp.Code)
	}

	var projects []map[string]any
	if err := json.Unmarshal(projectsResp.Body.Bytes(), &projects); err != nil {
		t.Fatalf("decode projects: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	projectID, _ := projects[0]["id"].(string)

	treeReq := httptest.NewRequest(http.MethodGet, "/api/projects/"+projectID+"/files/tree", nil)
	treeResp := httptest.NewRecorder()
	engine.ServeHTTP(treeResp, treeReq)
	if treeResp.Code != http.StatusOK {
		t.Fatalf("expected tree status %d, got %d body=%s", http.StatusOK, treeResp.Code, treeResp.Body.String())
	}

	contentReq := httptest.NewRequest(http.MethodGet, "/api/projects/"+projectID+"/files/content?path=go.mod", nil)
	contentResp := httptest.NewRecorder()
	engine.ServeHTTP(contentResp, contentReq)
	if contentResp.Code != http.StatusOK {
		t.Fatalf("expected content status %d, got %d body=%s", http.StatusOK, contentResp.Code, contentResp.Body.String())
	}

	var content map[string]any
	if err := json.Unmarshal(contentResp.Body.Bytes(), &content); err != nil {
		t.Fatalf("decode content: %v", err)
	}
	if content["path"] != "go.mod" {
		t.Fatalf("expected go.mod content path, got %#v", content["path"])
	}
}

func TestProjectCreateThreadRouteDispatchesInitialPromptThroughProjectExecutor(t *testing.T) {
	engine, deps := newTestRuntime(t)

	createReq := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewBufferString(`{"id":"session-1","project_name":"codeScope","workspace_root":"/workspace","machine_id":"machine-1"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createResp := httptest.NewRecorder()
	engine.ServeHTTP(createResp, createReq)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d", http.StatusCreated, createResp.Code)
	}

	updateReq := httptest.NewRequest(http.MethodPatch, "/api/sessions/session-1/status", bytes.NewBufferString(`{"status":"running"}`))
	updateReq.Header.Set("Content-Type", "application/json")
	updateResp := httptest.NewRecorder()
	engine.ServeHTTP(updateResp, updateReq)
	if updateResp.Code != http.StatusOK {
		t.Fatalf("expected update status %d, got %d", http.StatusOK, updateResp.Code)
	}
	bridgeConnectedAt := time.Date(2026, 3, 19, 9, 1, 0, 0, time.UTC)
	if _, _, err := deps.SessionService.EnsureBridgeSession("session-1", session.BridgeMetadata{
		ProjectName:   "codeScope",
		WorkspaceRoot: "/workspace",
		MachineID:     "machine-1",
	}, bridgeConnectedAt); err != nil {
		t.Fatalf("ensure bridge session: %v", err)
	}

	unregister := deps.CommandService.Bridges().Register("session-1", make(chan any, 1))
	defer unregister()

	projectsReq := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	projectsResp := httptest.NewRecorder()
	engine.ServeHTTP(projectsResp, projectsReq)
	if projectsResp.Code != http.StatusOK {
		t.Fatalf("expected projects status %d, got %d", http.StatusOK, projectsResp.Code)
	}

	var projects []map[string]any
	if err := json.Unmarshal(projectsResp.Body.Bytes(), &projects); err != nil {
		t.Fatalf("decode projects: %v", err)
	}
	projectID, _ := projects[0]["id"].(string)

	createThreadReq := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID+"/threads", strings.NewReader(`{"content":"Start a project-level formal thread"}`))
	createThreadReq.Header.Set("Content-Type", "application/json")
	createThreadResp := httptest.NewRecorder()
	engine.ServeHTTP(createThreadResp, createThreadReq)
	if createThreadResp.Code != http.StatusCreated {
		t.Fatalf("expected create thread status %d, got %d body=%s", http.StatusCreated, createThreadResp.Code, createThreadResp.Body.String())
	}

	var thread map[string]any
	if err := json.Unmarshal(createThreadResp.Body.Bytes(), &thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}
	if thread["project_id"] != projectID {
		t.Fatalf("expected response project_id %q, got %#v", projectID, thread["project_id"])
	}
	if thread["session_id"] != "session-1" {
		t.Fatalf("expected project executor session-1, got %#v", thread["session_id"])
	}
	if thread["title"] == "" {
		t.Fatalf("expected non-empty thread title, got %#v", thread["title"])
	}

	tasks, err := deps.CommandService.ListBySession("session-1")
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 prompt task, got %d", len(tasks))
	}
	if tasks[0].Payload["thread_id"] != thread["id"] {
		t.Fatalf("expected command payload thread_id to match response id, got %#v", tasks[0].Payload["thread_id"])
	}
	if tasks[0].Payload["project_id"] != projectID {
		t.Fatalf("expected command payload project_id %q, got %#v", projectID, tasks[0].Payload["project_id"])
	}
}

func TestThreadPromptRouteUsesWritableSessionWithinSemanticThread(t *testing.T) {
	engine, deps := newTestRuntime(t)

	for _, body := range []string{
		`{"id":"session-online-old","project_name":"codeScope","workspace_root":"/workspace","machine_id":"machine-1"}`,
		`{"id":"session-offline-new","project_name":"codeScope","workspace_root":"/workspace","machine_id":"machine-1"}`,
	} {
		createReq := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewBufferString(body))
		createReq.Header.Set("Content-Type", "application/json")
		createResp := httptest.NewRecorder()
		engine.ServeHTTP(createResp, createReq)
		if createResp.Code != http.StatusCreated {
			t.Fatalf("expected create status %d, got %d", http.StatusCreated, createResp.Code)
		}
	}

	for _, update := range []struct {
		id     string
		status string
	}{
		{id: "session-online-old", status: "running"},
		{id: "session-offline-new", status: "running"},
	} {
		updateReq := httptest.NewRequest(http.MethodPatch, "/api/sessions/"+update.id+"/status", bytes.NewBufferString(`{"status":"`+update.status+`"}`))
		updateReq.Header.Set("Content-Type", "application/json")
		updateResp := httptest.NewRecorder()
		engine.ServeHTTP(updateResp, updateReq)
		if updateResp.Code != http.StatusOK {
			t.Fatalf("expected update status %d, got %d", http.StatusOK, updateResp.Code)
		}
	}

	if _, _, err := deps.SessionService.EnsureBridgeSession("session-online-old", session.BridgeMetadata{
		ProjectName:   "codeScope",
		WorkspaceRoot: "/workspace",
		MachineID:     "machine-1",
	}, time.Date(2026, 3, 19, 9, 2, 0, 0, time.UTC)); err != nil {
		t.Fatalf("ensure online bridge session: %v", err)
	}

	if _, err := deps.EventService.Ingest(event.Message{
		MessageID:   "msg-online-old",
		SessionID:   "session-online-old",
		MessageType: event.MessageTypeEvent,
		EventType:   event.TypeCommand,
		Timestamp:   time.Date(2026, 3, 19, 9, 2, 0, 0, time.UTC).Format(time.RFC3339Nano),
		Payload: map[string]any{
			"thread_id":    "thread-semantic-1",
			"thread_title": "formal thread",
			"role":         "user",
			"content":      "first prompt",
		},
	}); err != nil {
		t.Fatalf("seed old semantic session: %v", err)
	}
	if _, err := deps.EventService.Ingest(event.Message{
		MessageID:   "msg-offline-new",
		SessionID:   "session-offline-new",
		MessageType: event.MessageTypeEvent,
		EventType:   event.TypeAIOutput,
		Timestamp:   time.Date(2026, 3, 19, 9, 4, 0, 0, time.UTC).Format(time.RFC3339Nano),
		Payload: map[string]any{
			"thread_id":    "thread-semantic-1",
			"thread_title": "formal thread",
			"content":      "latest assistant output",
		},
	}); err != nil {
		t.Fatalf("seed new semantic session: %v", err)
	}

	unregister := deps.CommandService.Bridges().Register("session-online-old", make(chan any, 1))
	defer unregister()

	promptReq := httptest.NewRequest(http.MethodPost, "/api/threads/thread-semantic-1/commands/prompt", strings.NewReader(`{"content":"continue on the writable session"}`))
	promptReq.Header.Set("Content-Type", "application/json")
	promptResp := httptest.NewRecorder()
	engine.ServeHTTP(promptResp, promptReq)
	if promptResp.Code != http.StatusCreated {
		t.Fatalf("expected thread prompt status %d, got %d body=%s", http.StatusCreated, promptResp.Code, promptResp.Body.String())
	}

	var task map[string]any
	if err := json.Unmarshal(promptResp.Body.Bytes(), &task); err != nil {
		t.Fatalf("decode task: %v", err)
	}
	if task["session_id"] != "session-online-old" {
		t.Fatalf("expected writable online session, got %#v", task["session_id"])
	}
	if payload, _ := task["payload"].(map[string]any); payload["thread_id"] != "thread-semantic-1" {
		t.Fatalf("expected thread_id to remain semantic in command payload, got %#v", payload["thread_id"])
	}
}

func TestThreadCommandsRouteAggregatesTasksAcrossSemanticThreadSessions(t *testing.T) {
	engine, deps := newTestRuntime(t)

	for _, body := range []string{
		`{"id":"session-a","project_name":"codeScope","workspace_root":"/workspace","machine_id":"machine-1"}`,
		`{"id":"session-b","project_name":"codeScope","workspace_root":"/workspace","machine_id":"machine-1"}`,
	} {
		createReq := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewBufferString(body))
		createReq.Header.Set("Content-Type", "application/json")
		createResp := httptest.NewRecorder()
		engine.ServeHTTP(createResp, createReq)
		if createResp.Code != http.StatusCreated {
			t.Fatalf("expected create status %d, got %d", http.StatusCreated, createResp.Code)
		}
	}

	for _, sessionID := range []string{"session-a", "session-b"} {
		updateReq := httptest.NewRequest(http.MethodPatch, "/api/sessions/"+sessionID+"/status", bytes.NewBufferString(`{"status":"running"}`))
		updateReq.Header.Set("Content-Type", "application/json")
		updateResp := httptest.NewRecorder()
		engine.ServeHTTP(updateResp, updateReq)
		if updateResp.Code != http.StatusOK {
			t.Fatalf("expected update status %d, got %d", http.StatusOK, updateResp.Code)
		}
		if _, _, err := deps.SessionService.EnsureBridgeSession(sessionID, session.BridgeMetadata{
			ProjectName:   "codeScope",
			WorkspaceRoot: "/workspace",
			MachineID:     "machine-1",
		}, time.Date(2026, 3, 19, 9, 2, 0, 0, time.UTC)); err != nil {
			t.Fatalf("ensure bridge session %s: %v", sessionID, err)
		}
		unregister := deps.CommandService.Bridges().Register(sessionID, make(chan any, 1))
		defer unregister()
	}

	for _, seed := range []event.Message{
		{
			MessageID:   "msg-a",
			SessionID:   "session-a",
			MessageType: event.MessageTypeEvent,
			EventType:   event.TypeCommand,
			Timestamp:   time.Date(2026, 3, 19, 9, 2, 0, 0, time.UTC).Format(time.RFC3339Nano),
			Payload: map[string]any{
				"thread_id":    "thread-semantic-agg",
				"thread_title": "aggregated thread",
				"role":         "user",
				"content":      "first prompt",
			},
		},
		{
			MessageID:   "msg-b",
			SessionID:   "session-b",
			MessageType: event.MessageTypeEvent,
			EventType:   event.TypeAIOutput,
			Timestamp:   time.Date(2026, 3, 19, 9, 3, 0, 0, time.UTC).Format(time.RFC3339Nano),
			Payload: map[string]any{
				"thread_id":    "thread-semantic-agg",
				"thread_title": "aggregated thread",
				"content":      "latest reply",
			},
		},
	} {
		if _, err := deps.EventService.Ingest(seed); err != nil {
			t.Fatalf("seed thread event: %v", err)
		}
	}

	if _, err := deps.CommandService.CreatePrompt("session-a", session.PromptCommandInput{
		Content:   "old command",
		ThreadID:  "thread-semantic-agg",
		ProjectID: "project-7c6c07f9c4b59dcd",
	}); err != nil {
		t.Fatalf("create prompt for session-a: %v", err)
	}
	if _, err := deps.CommandService.CreatePrompt("session-b", session.PromptCommandInput{
		Content:   "new command",
		ThreadID:  "thread-semantic-agg",
		ProjectID: "project-7c6c07f9c4b59dcd",
	}); err != nil {
		t.Fatalf("create prompt for session-b: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/threads/thread-semantic-agg/commands", nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected commands status %d, got %d body=%s", http.StatusOK, resp.Code, resp.Body.String())
	}

	var tasks []map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &tasks); err != nil {
		t.Fatalf("decode tasks: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 aggregated tasks, got %d", len(tasks))
	}
}

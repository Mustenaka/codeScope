package router_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"codescope/server/internal/app"
	"codescope/server/internal/config"
	"codescope/server/internal/http/router"
)

func newTestEngine(t *testing.T) http.Handler {
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
	})
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

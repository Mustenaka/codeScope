package router_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"codescope/server/internal/app"
	"codescope/server/internal/config"
	"codescope/server/internal/http/router"

	"github.com/gorilla/websocket"
)

func TestBridgeEventBroadcastToMobileSubscriber(t *testing.T) {
	container := app.NewWithConfig(config.Config{AppName: "codeScope Server"})
	deps := container.Dependencies()
	engine := router.New(router.Dependencies{
		Config:         deps.Config,
		SessionService: deps.SessionService,
		EventService:   deps.EventService,
		EventHub:       deps.EventHub,
		FileService:    deps.FileService,
		ProjectService: deps.ProjectService,
		PromptService:  deps.PromptService,
		CommandService: deps.CommandService,
	})
	server := httptest.NewServer(engine)
	defer server.Close()

	httpURL := strings.TrimPrefix(server.URL, "http")
	mobileConn, _, err := websocket.DefaultDialer.Dial("ws"+httpURL+"/ws/mobile?session_id=session-1", nil)
	if err != nil {
		t.Fatalf("dial mobile websocket: %v", err)
	}
	defer mobileConn.Close()

	bridgeConn, _, err := websocket.DefaultDialer.Dial("ws"+httpURL+"/ws/bridge", nil)
	if err != nil {
		t.Fatalf("dial bridge websocket: %v", err)
	}
	defer bridgeConn.Close()

	message := map[string]any{
		"message_id":   "message-1",
		"session_id":   "session-1",
		"message_type": "event",
		"event_type":   "terminal_output",
		"timestamp":    "2026-03-17T10:00:00Z",
		"payload": map[string]any{
			"content":        "go test ./...",
			"agent_name":     "fake-source",
			"workspace_root": "/workspace",
			"machine_id":     "machine-1",
			"project_name":   "codeScope",
		},
	}
	if err := bridgeConn.WriteJSON(message); err != nil {
		t.Fatalf("write bridge message: %v", err)
	}

	var ack map[string]any
	if err := bridgeConn.ReadJSON(&ack); err != nil {
		t.Fatalf("read bridge ack: %v", err)
	}
	if ack["type"] != "ack" {
		t.Fatalf("expected ack response, got %#v", ack)
	}

	if err := mobileConn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set deadline: %v", err)
	}

	var received map[string]any
	if err := mobileConn.ReadJSON(&received); err != nil {
		t.Fatalf("read mobile message: %v", err)
	}

	if received["session_id"] != "session-1" {
		t.Fatalf("expected session-1, got %v", received["session_id"])
	}

	payload, ok := received["payload"].(map[string]any)
	if !ok {
		raw, _ := json.Marshal(received["payload"])
		t.Fatalf("expected payload object, got %s", string(raw))
	}
	if payload["content"] != "go test ./..." {
		t.Fatalf("expected payload content go test ./..., got %v", payload["content"])
	}

	sessionResp := httptest.NewRecorder()
	sessionReq := httptest.NewRequest("GET", "/api/sessions/session-1", nil)
	engine.ServeHTTP(sessionResp, sessionReq)
	if sessionResp.Code != 200 {
		t.Fatalf("expected auto-created session to be queryable, got %d", sessionResp.Code)
	}
}

func TestBridgeEventBroadcastToThreadSubscriber(t *testing.T) {
	container := app.NewWithConfig(config.Config{AppName: "codeScope Server"})
	deps := container.Dependencies()
	engine := router.New(router.Dependencies{
		Config:         deps.Config,
		SessionService: deps.SessionService,
		EventService:   deps.EventService,
		EventHub:       deps.EventHub,
		FileService:    deps.FileService,
		ProjectService: deps.ProjectService,
		PromptService:  deps.PromptService,
		CommandService: deps.CommandService,
	})
	server := httptest.NewServer(engine)
	defer server.Close()

	httpURL := strings.TrimPrefix(server.URL, "http")
	mobileConn, _, err := websocket.DefaultDialer.Dial("ws"+httpURL+"/ws/mobile?thread_id=thread-001", nil)
	if err != nil {
		t.Fatalf("dial mobile websocket: %v", err)
	}
	defer mobileConn.Close()

	bridgeConn, _, err := websocket.DefaultDialer.Dial("ws"+httpURL+"/ws/bridge", nil)
	if err != nil {
		t.Fatalf("dial bridge websocket: %v", err)
	}
	defer bridgeConn.Close()

	message := map[string]any{
		"message_id":   "message-2",
		"session_id":   "session-1",
		"message_type": "event",
		"event_type":   "ai_output",
		"timestamp":    "2026-03-19T10:05:00Z",
		"payload": map[string]any{
			"content":        "Applied follow-up patch",
			"agent_name":     "codex",
			"workspace_root": "/workspace",
			"machine_id":     "machine-1",
			"project_name":   "codeScope",
			"thread_id":      "thread-001",
			"thread_state":   "running",
		},
	}
	if err := bridgeConn.WriteJSON(message); err != nil {
		t.Fatalf("write bridge message: %v", err)
	}

	var ack map[string]any
	if err := bridgeConn.ReadJSON(&ack); err != nil {
		t.Fatalf("read bridge ack: %v", err)
	}

	if err := mobileConn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set deadline: %v", err)
	}

	var received map[string]any
	if err := mobileConn.ReadJSON(&received); err != nil {
		t.Fatalf("read mobile message: %v", err)
	}

	payload, ok := received["payload"].(map[string]any)
	if !ok {
		t.Fatalf("expected payload object, got %#v", received["payload"])
	}
	if payload["thread_id"] != "thread-001" {
		t.Fatalf("expected thread-001, got %#v", payload["thread_id"])
	}
}

func TestBridgeEventBroadcastToProjectSubscriber(t *testing.T) {
	container := app.NewWithConfig(config.Config{AppName: "codeScope Server"})
	deps := container.Dependencies()
	engine := router.New(router.Dependencies{
		Config:         deps.Config,
		SessionService: deps.SessionService,
		EventService:   deps.EventService,
		EventHub:       deps.EventHub,
		FileService:    deps.FileService,
		ProjectService: deps.ProjectService,
		PromptService:  deps.PromptService,
		CommandService: deps.CommandService,
	})
	server := httptest.NewServer(engine)
	defer server.Close()

	httpURL := strings.TrimPrefix(server.URL, "http")
	mobileConn, _, err := websocket.DefaultDialer.Dial("ws"+httpURL+"/ws/mobile?project_id=project-001", nil)
	if err != nil {
		t.Fatalf("dial mobile websocket: %v", err)
	}
	defer mobileConn.Close()

	bridgeConn, _, err := websocket.DefaultDialer.Dial("ws"+httpURL+"/ws/bridge", nil)
	if err != nil {
		t.Fatalf("dial bridge websocket: %v", err)
	}
	defer bridgeConn.Close()

	message := map[string]any{
		"message_id":   "message-3",
		"session_id":   "session-1",
		"message_type": "event",
		"event_type":   "ai_output",
		"timestamp":    "2026-03-19T10:06:00Z",
		"payload": map[string]any{
			"content":        "Updated project thread summary",
			"agent_name":     "codex",
			"workspace_root": "/workspace",
			"machine_id":     "machine-1",
			"project_name":   "codeScope",
			"project_id":     "project-001",
			"thread_id":      "thread-001",
			"thread_state":   "running",
		},
	}
	if err := bridgeConn.WriteJSON(message); err != nil {
		t.Fatalf("write bridge message: %v", err)
	}

	var ack map[string]any
	if err := bridgeConn.ReadJSON(&ack); err != nil {
		t.Fatalf("read bridge ack: %v", err)
	}

	if err := mobileConn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set deadline: %v", err)
	}

	var received map[string]any
	if err := mobileConn.ReadJSON(&received); err != nil {
		t.Fatalf("read mobile message: %v", err)
	}

	payload, ok := received["payload"].(map[string]any)
	if !ok {
		t.Fatalf("expected payload object, got %#v", received["payload"])
	}
	if payload["project_id"] != "project-001" {
		t.Fatalf("expected project-001, got %#v", payload["project_id"])
	}
}

func TestBridgeLifecycleAndCommandRoundTrip(t *testing.T) {
	engine := newTestEngine(t)
	server := httptest.NewServer(engine)
	defer server.Close()

	httpURL := strings.TrimPrefix(server.URL, "http")
	wsURL := "ws" + httpURL

	bridgeConn, _, err := websocket.DefaultDialer.Dial(wsURL+"/ws/bridge", nil)
	if err != nil {
		t.Fatalf("dial bridge websocket: %v", err)
	}

	heartbeat := map[string]any{
		"message_id":   "message-heartbeat-1",
		"session_id":   "session-roundtrip",
		"message_type": "heartbeat",
		"timestamp":    "2026-03-17T10:00:00Z",
		"payload": map[string]any{
			"agent_name":     "codex",
			"workspace_root": "/workspace",
			"machine_id":     "machine-1",
			"project_name":   "codeScope",
		},
	}
	if err := bridgeConn.WriteJSON(heartbeat); err != nil {
		t.Fatalf("write heartbeat: %v", err)
	}

	var ack map[string]any
	if err := bridgeConn.ReadJSON(&ack); err != nil {
		t.Fatalf("read heartbeat ack: %v", err)
	}

	sessionResp := httptest.NewRecorder()
	sessionReq := httptest.NewRequest(http.MethodGet, "/api/sessions/session-roundtrip", nil)
	engine.ServeHTTP(sessionResp, sessionReq)
	if sessionResp.Code != http.StatusOK {
		t.Fatalf("expected session detail after heartbeat, got %d", sessionResp.Code)
	}

	var sessionDetail map[string]any
	if err := json.Unmarshal(sessionResp.Body.Bytes(), &sessionDetail); err != nil {
		t.Fatalf("decode session detail: %v", err)
	}
	if sessionDetail["bridge_online"] != true {
		t.Fatalf("expected bridge_online true, got %#v", sessionDetail["bridge_online"])
	}

	commandReq := httptest.NewRequest(http.MethodPost, "/api/sessions/session-roundtrip/commands/prompt", strings.NewReader(`{"content":"continue fixing tests"}`))
	commandReq.Header.Set("Content-Type", "application/json")
	commandResp := httptest.NewRecorder()
	engine.ServeHTTP(commandResp, commandReq)
	if commandResp.Code != http.StatusCreated {
		t.Fatalf("expected prompt command create status 201, got %d: %s", commandResp.Code, commandResp.Body.String())
	}

	var commandTask map[string]any
	if err := json.Unmarshal(commandResp.Body.Bytes(), &commandTask); err != nil {
		t.Fatalf("decode command task: %v", err)
	}
	commandID, _ := commandTask["id"].(string)
	if commandID == "" {
		t.Fatalf("expected command id in response")
	}

	if err := bridgeConn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set bridge read deadline: %v", err)
	}

	var dispatched map[string]any
	if err := bridgeConn.ReadJSON(&dispatched); err != nil {
		t.Fatalf("read dispatched command: %v", err)
	}
	if dispatched["message_type"] != "command" {
		t.Fatalf("expected command dispatch, got %#v", dispatched)
	}
	if dispatched["command_id"] != commandID {
		t.Fatalf("expected command id %q, got %#v", commandID, dispatched["command_id"])
	}

	result := map[string]any{
		"message_id":   "message-result-1",
		"session_id":   "session-roundtrip",
		"message_type": "command_result",
		"command_id":   commandID,
		"command_type": "send_prompt",
		"status":       "success",
		"timestamp":    "2026-03-17T10:00:01Z",
		"payload": map[string]any{
			"accepted":   true,
			"local_path": "D:/bridge-data/prompts/inbox.jsonl",
		},
	}
	if err := bridgeConn.WriteJSON(result); err != nil {
		t.Fatalf("write command result: %v", err)
	}
	if err := bridgeConn.ReadJSON(&ack); err != nil {
		t.Fatalf("read command result ack: %v", err)
	}

	listResp := httptest.NewRecorder()
	listReq := httptest.NewRequest(http.MethodGet, "/api/sessions/session-roundtrip/commands", nil)
	engine.ServeHTTP(listResp, listReq)
	if listResp.Code != http.StatusOK {
		t.Fatalf("expected command list status 200, got %d", listResp.Code)
	}

	var tasks []map[string]any
	if err := json.Unmarshal(listResp.Body.Bytes(), &tasks); err != nil {
		t.Fatalf("decode command list: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 command task, got %d", len(tasks))
	}
	if tasks[0]["status"] != "success" {
		t.Fatalf("expected successful task, got %#v", tasks[0]["status"])
	}

	bridgeConn.Close()
	time.Sleep(150 * time.Millisecond)

	sessionResp = httptest.NewRecorder()
	sessionReq = httptest.NewRequest(http.MethodGet, "/api/sessions/session-roundtrip", nil)
	engine.ServeHTTP(sessionResp, sessionReq)
	if sessionResp.Code != http.StatusOK {
		t.Fatalf("expected session detail after disconnect, got %d", sessionResp.Code)
	}
	if err := json.Unmarshal(sessionResp.Body.Bytes(), &sessionDetail); err != nil {
		t.Fatalf("decode session detail after disconnect: %v", err)
	}
	if sessionDetail["bridge_online"] != false {
		t.Fatalf("expected bridge_online false after disconnect, got %#v", sessionDetail["bridge_online"])
	}

	commandReq = httptest.NewRequest(http.MethodPost, "/api/sessions/session-roundtrip/commands/prompt", strings.NewReader(`{"content":"retry while offline"}`))
	commandReq.Header.Set("Content-Type", "application/json")
	commandResp = httptest.NewRecorder()
	engine.ServeHTTP(commandResp, commandReq)
	if commandResp.Code != http.StatusConflict {
		t.Fatalf("expected conflict while bridge offline, got %d", commandResp.Code)
	}

	oldBridge, _, err := websocket.DefaultDialer.Dial(wsURL+"/ws/bridge", nil)
	if err != nil {
		t.Fatalf("dial old bridge websocket: %v", err)
	}
	defer oldBridge.Close()
	if err := oldBridge.WriteJSON(heartbeat); err != nil {
		t.Fatalf("write old bridge heartbeat: %v", err)
	}
	if err := oldBridge.ReadJSON(&ack); err != nil {
		t.Fatalf("read old bridge heartbeat ack: %v", err)
	}

	newBridge, _, err := websocket.DefaultDialer.Dial(wsURL+"/ws/bridge", nil)
	if err != nil {
		t.Fatalf("dial replacement bridge websocket: %v", err)
	}
	defer newBridge.Close()
	replacementHeartbeat := map[string]any{
		"message_id":   "message-heartbeat-2",
		"session_id":   "session-roundtrip",
		"message_type": "heartbeat",
		"timestamp":    "2026-03-17T10:00:02Z",
		"payload": map[string]any{
			"agent_name":     "codex",
			"workspace_root": "/workspace",
			"machine_id":     "machine-1",
			"project_name":   "codeScope",
		},
	}
	if err := newBridge.WriteJSON(replacementHeartbeat); err != nil {
		t.Fatalf("write replacement heartbeat: %v", err)
	}
	if err := newBridge.ReadJSON(&ack); err != nil {
		t.Fatalf("read replacement heartbeat ack: %v", err)
	}

	statusResp := httptest.NewRecorder()
	statusReq := httptest.NewRequest(http.MethodGet, "/api/bridge/status", nil)
	engine.ServeHTTP(statusResp, statusReq)
	if statusResp.Code != http.StatusOK {
		t.Fatalf("expected bridge status endpoint to succeed, got %d", statusResp.Code)
	}

	oldBridge.Close()
	time.Sleep(150 * time.Millisecond)

	sessionResp = httptest.NewRecorder()
	sessionReq = httptest.NewRequest(http.MethodGet, "/api/sessions/session-roundtrip", nil)
	engine.ServeHTTP(sessionResp, sessionReq)
	if sessionResp.Code != http.StatusOK {
		t.Fatalf("expected session detail after old bridge disconnect, got %d", sessionResp.Code)
	}
	if err := json.Unmarshal(sessionResp.Body.Bytes(), &sessionDetail); err != nil {
		t.Fatalf("decode session detail after old bridge disconnect: %v", err)
	}
	if sessionDetail["bridge_online"] != true {
		t.Fatalf("expected replacement bridge to keep session online, got %#v", sessionDetail["bridge_online"])
	}

	commandReq = httptest.NewRequest(http.MethodPost, "/api/sessions/session-roundtrip/commands/prompt", strings.NewReader(`{"content":"dispatch to replacement bridge"}`))
	commandReq.Header.Set("Content-Type", "application/json")
	commandResp = httptest.NewRecorder()
	engine.ServeHTTP(commandResp, commandReq)
	if commandResp.Code != http.StatusCreated {
		t.Fatalf("expected prompt command create after reconnect, got %d: %s", commandResp.Code, commandResp.Body.String())
	}

	if err := newBridge.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set replacement bridge deadline: %v", err)
	}
	if err := newBridge.ReadJSON(&dispatched); err != nil {
		t.Fatalf("expected replacement bridge to receive command: %v", err)
	}
}

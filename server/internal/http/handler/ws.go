package handler

import (
	"errors"
	"log"
	"net/http"
	"time"

	"codescope/server/internal/event"
	"codescope/server/internal/session"
	"codescope/server/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const bridgeReadTimeout = 90 * time.Second

type EventIngestService interface {
	Ingest(message event.Message) (event.Record, error)
}

type WebSocketHandler struct {
	ingest   EventIngestService
	sessions *session.Service
	commands *session.CommandService
	bridges  *session.BridgeRegistry
	broker   event.Broker
	upgrader websocket.Upgrader
}

func NewWebSocketHandler(ingest EventIngestService, sessions *session.Service, commands *session.CommandService, bridges *session.BridgeRegistry, broker event.Broker) *WebSocketHandler {
	return &WebSocketHandler{
		ingest:   ingest,
		sessions: sessions,
		commands: commands,
		bridges:  bridges,
		broker:   broker,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (h *WebSocketHandler) Bridge(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	outbound := make(chan any, 32)
	writeErr := make(chan error, 1)
	go func() {
		for payload := range outbound {
			if err := conn.WriteJSON(payload); err != nil {
				writeErr <- err
				return
			}
		}
	}()

	var unregister func() bool
	var registeredSessionID string
	defer func() {
		removed := false
		if unregister != nil {
			removed = unregister()
		}
		if removed && registeredSessionID != "" {
			if err := h.sessions.MarkBridgeDisconnected(registeredSessionID, time.Now().UTC()); err != nil {
				log.Printf("bridge disconnect state update failed: session_id=%s err=%v", registeredSessionID, err)
			} else {
				log.Printf("bridge disconnected: session_id=%s", registeredSessionID)
			}
		}
		close(outbound)
	}()

	if err := conn.SetReadDeadline(time.Now().Add(bridgeReadTimeout)); err != nil {
		log.Printf("bridge websocket initial read deadline failed: %v", err)
	}

	for {
		select {
		case err := <-writeErr:
			if err != nil {
				return
			}
		default:
		}

		var message event.Message
		if err := conn.ReadJSON(&message); err != nil {
			return
		}
		if err := conn.SetReadDeadline(time.Now().Add(bridgeReadTimeout)); err != nil {
			log.Printf("bridge websocket refresh read deadline failed: session_id=%s err=%v", registeredSessionID, err)
		}

		if unregister == nil && message.SessionID != "" {
			at, parseErr := time.Parse(time.RFC3339, message.Timestamp)
			if parseErr != nil {
				at = time.Now().UTC()
			}
			metadata := session.BridgeMetadata{
				ProjectName:   payloadString(message.Payload, "project_name"),
				WorkspaceRoot: payloadString(message.Payload, "workspace_root"),
				MachineID:     payloadString(message.Payload, "machine_id"),
			}
			_, created, err := h.sessions.EnsureBridgeSession(message.SessionID, metadata, at.UTC())
			if err != nil {
				outbound <- gin.H{"type": "error", "error": err.Error()}
				continue
			}
			unregister = h.bridges.Register(message.SessionID, outbound)
			registeredSessionID = message.SessionID
			log.Printf("bridge connected: session_id=%s", message.SessionID)
			if created {
				log.Printf("bridge session ready: session_id=%s source=connect", message.SessionID)
			}
		}

		switch message.MessageType {
		case event.MessageTypeEvent, event.MessageTypeHeartbeat:
			record, err := h.ingest.Ingest(message)
			if err != nil {
				outbound <- gin.H{"type": "error", "error": err.Error()}
				continue
			}
			outbound <- gin.H{
				"type":       "ack",
				"message_id": record.MessageID,
				"session_id": record.SessionID,
			}
		case event.MessageTypeCommandResult:
			task, err := h.commands.Complete(session.BridgeMessage{
				MessageID:   message.MessageID,
				SessionID:   message.SessionID,
				MessageType: string(message.MessageType),
				EventType:   string(message.EventType),
				CommandID:   message.CommandID,
				CommandType: string(message.CommandType),
				Status:      string(message.Status),
				Timestamp:   message.Timestamp,
				Payload:     message.Payload,
			})
			if err != nil {
				state := "error"
				if errors.Is(err, store.ErrNotFound) {
					state = "not_found"
				}
				outbound <- gin.H{"type": state, "error": err.Error()}
				continue
			}
			if ts, parseErr := time.Parse(time.RFC3339, message.Timestamp); parseErr == nil {
				if err := h.sessions.TouchActivity(message.SessionID, ts.UTC()); err != nil {
					log.Printf("command result activity update failed: session_id=%s err=%v", message.SessionID, err)
				}
			}
			outbound <- gin.H{
				"type":       "ack",
				"message_id": message.MessageID,
				"session_id": task.SessionID,
				"command_id": task.ID,
			}
		default:
			outbound <- gin.H{"type": "error", "error": "unsupported message_type"}
		}
	}
}

func payloadString(payload map[string]any, key string) string {
	value, ok := payload[key].(string)
	if !ok {
		return ""
	}
	return value
}

func (h *WebSocketHandler) Mobile(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	sessionID := c.Query("session_id")
	threadID := c.Query("thread_id")
	projectID := c.Query("project_id")
	stream, cancel := h.broker.Subscribe(sessionID, threadID, projectID)
	defer cancel()

	closeSignal := make(chan struct{})
	go func() {
		defer close(closeSignal)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-closeSignal:
			return
		case record, ok := <-stream:
			if !ok {
				return
			}
			if err := conn.WriteJSON(record); err != nil {
				return
			}
		case <-ticker.C:
			if err := conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(5*time.Second)); err != nil {
				log.Printf("mobile websocket ping failed: %v", err)
				return
			}
		}
	}
}

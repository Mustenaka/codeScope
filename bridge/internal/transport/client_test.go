package transport

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"codescope/bridge/internal/session"
	"github.com/gorilla/websocket"
)

func TestClientPublishesAndReceivesCommand(t *testing.T) {
	var (
		upgrader = websocket.Upgrader{}
		mu       sync.Mutex
		received []session.Message
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade websocket: %v", err)
			return
		}
		defer conn.Close()

		command := session.Message{
			MessageID:   "msg-command",
			SessionID:   "session-1",
			MessageType: session.MessageTypeCommand,
			CommandID:   "cmd-1",
			CommandType: session.CommandTypeSendPrompt,
			Timestamp:   time.Now().UTC().Format(time.RFC3339Nano),
			Payload: map[string]any{
				"content": "continue",
			},
		}
		if err := conn.WriteJSON(command); err != nil {
			t.Errorf("write command: %v", err)
			return
		}

		for {
			var msg session.Message
			if err := conn.ReadJSON(&msg); err != nil {
				return
			}
			mu.Lock()
			received = append(received, msg)
			mu.Unlock()
			if msg.MessageType == session.MessageTypeHeartbeat {
				return
			}
		}
	}))
	defer server.Close()

	client := NewClient(serverURLToWebSocket(server.URL),
		WithReconnectInterval(10*time.Millisecond),
		WithHeartbeatInterval(20*time.Millisecond),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Start(ctx); err != nil {
		t.Fatalf("start client: %v", err)
	}

	event := session.Message{
		MessageID:   "msg-event",
		SessionID:   "session-1",
		MessageType: session.MessageTypeEvent,
		EventType:   session.EventTypeTerminalOutput,
		Timestamp:   time.Now().UTC().Format(time.RFC3339Nano),
		Payload: map[string]any{
			"content": "hello",
		},
	}

	if err := client.Publish(ctx, event); err != nil {
		t.Fatalf("publish event: %v", err)
	}

	select {
	case cmd := <-client.Commands():
		if cmd.CommandType != session.CommandTypeSendPrompt {
			t.Fatalf("expected send_prompt command, got %q", cmd.CommandType)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for command")
	}

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		count := len(received)
		var first session.Message
		if count > 0 {
			first = received[0]
		}
		mu.Unlock()

		if count >= 2 {
			if first.MessageID != "msg-event" {
				t.Fatalf("expected first sent message to be event, got %#v", first)
			}
			return
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("expected server to receive event and heartbeat, got %d message(s)", len(received))
}

func TestClientReconnectsAndFlushesBufferedMessages(t *testing.T) {
	var (
		upgrader = websocket.Upgrader{}
		addrMu   sync.Mutex
		received []session.Message
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade websocket: %v", err)
			return
		}
		defer conn.Close()

		for {
			var msg session.Message
			if err := conn.ReadJSON(&msg); err != nil {
				return
			}
			addrMu.Lock()
			received = append(received, msg)
			addrMu.Unlock()
		}
	}))
	defer server.Close()

	wsURL := serverURLToWebSocket(server.URL)

	client := NewClient(wsURL, WithReconnectInterval(20*time.Millisecond))
	client.wsDialer = &flakyDialer{
		delegate:  websocket.DefaultDialer,
		failCount: 1,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Start(ctx); err != nil {
		t.Fatalf("start client: %v", err)
	}

	event := session.Message{
		MessageID:   "msg-buffered",
		SessionID:   "session-1",
		MessageType: session.MessageTypeEvent,
		EventType:   session.EventTypeAIOutput,
		Timestamp:   time.Now().UTC().Format(time.RFC3339Nano),
		Payload: map[string]any{
			"content": "buffered",
		},
	}

	if err := client.Publish(ctx, event); err != nil {
		t.Fatalf("publish buffered event: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		addrMu.Lock()
		count := len(received)
		addrMu.Unlock()
		if count > 0 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}

	t.Fatal("expected buffered message to be delivered after reconnect")
}

type flakyDialer struct {
	delegate  dialer
	failCount int32
	attempts  int32
}

func (d *flakyDialer) DialContext(ctx context.Context, urlStr string, requestHeader http.Header) (*websocket.Conn, *http.Response, error) {
	attempt := atomic.AddInt32(&d.attempts, 1)
	if attempt <= d.failCount {
		return nil, nil, errors.New("simulated dial failure")
	}
	return d.delegate.DialContext(ctx, urlStr, requestHeader)
}

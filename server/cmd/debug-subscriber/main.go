package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"codescope/server/internal/event"
	"github.com/gorilla/websocket"
)

func main() {
	var (
		serverURL = flag.String("server-url", "ws://localhost:8080/ws/mobile", "server websocket url")
		sessionID = flag.String("session-id", "", "optional session id filter")
	)
	flag.Parse()

	target, err := buildTarget(*serverURL, *sessionID)
	if err != nil {
		log.Fatal(err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)

	for {
		if err := subscribe(target, stop); err != nil {
			log.Printf("subscriber disconnected: %v", err)
		}

		select {
		case <-stop:
			return
		case <-time.After(2 * time.Second):
		}
	}
}

func subscribe(target string, stop <-chan os.Signal) error {
	conn, _, err := websocket.DefaultDialer.Dial(target, nil)
	if err != nil {
		return fmt.Errorf("dial websocket %s: %w", target, err)
	}
	defer conn.Close()

	log.Printf("subscribed %s", target)
	for {
		select {
		case <-stop:
			_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "shutdown"), time.Now().Add(time.Second))
			return nil
		default:
		}

		var record event.Record
		if err := conn.ReadJSON(&record); err != nil {
			return err
		}
		fmt.Println(formatRecord(record))
	}
}

func buildTarget(raw, sessionID string) (string, error) {
	target, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("parse server url: %w", err)
	}
	query := target.Query()
	if sessionID != "" {
		query.Set("session_id", sessionID)
	}
	target.RawQuery = query.Encode()
	return target.String(), nil
}

func formatRecord(record event.Record) string {
	content := firstNonEmpty(
		asString(record.Payload["content"]),
		asString(record.Payload["path"]),
		compactPayload(record.Payload),
	)
	return fmt.Sprintf("[%s] %s %s: %s",
		record.SessionID,
		record.Timestamp.Format(time.RFC3339),
		record.EventType,
		content,
	)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return "-"
}

func asString(value any) string {
	text, _ := value.(string)
	return text
}

func compactPayload(payload map[string]any) string {
	if len(payload) == 0 {
		return ""
	}
	parts := make([]string, 0, len(payload))
	for key, value := range payload {
		parts = append(parts, fmt.Sprintf("%s=%v", key, value))
	}
	return strings.Join(parts, " ")
}

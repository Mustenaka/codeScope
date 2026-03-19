package transport

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"codescope/bridge/internal/session"
	"github.com/gorilla/websocket"
)

type dialer interface {
	DialContext(ctx context.Context, urlStr string, requestHeader http.Header) (*websocket.Conn, *http.Response, error)
}

type heartbeatFactory func() session.Message

type Client struct {
	target            string
	logger            *log.Logger
	wsDialer          dialer
	reconnectInterval time.Duration
	heartbeatInterval time.Duration
	newHeartbeat      heartbeatFactory
	outgoing          chan session.Message
	commands          chan session.Message
	startOnce         sync.Once
}

type Option func(*Client)

func WithReconnectInterval(interval time.Duration) Option {
	return func(c *Client) {
		c.reconnectInterval = interval
	}
}

func WithHeartbeatInterval(interval time.Duration) Option {
	return func(c *Client) {
		c.heartbeatInterval = interval
	}
}

func WithHeartbeatFactory(factory func() session.Message) Option {
	return func(c *Client) {
		c.newHeartbeat = factory
	}
}

func WithLogger(logger *log.Logger) Option {
	return func(c *Client) {
		c.logger = logger
	}
}

func NewClient(target string, opts ...Option) *Client {
	client := &Client{
		target:            target,
		logger:            log.Default(),
		wsDialer:          websocket.DefaultDialer,
		reconnectInterval: 2 * time.Second,
		heartbeatInterval: 15 * time.Second,
		outgoing:          make(chan session.Message, 256),
		commands:          make(chan session.Message, 64),
		newHeartbeat: func() session.Message {
			return session.Message{
				MessageID:   session.NewMessageID(),
				MessageType: session.MessageTypeHeartbeat,
				EventType:   session.EventTypeHeartbeat,
				Timestamp:   time.Now().UTC().Format(time.RFC3339Nano),
			}
		},
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func (c *Client) Target() string {
	return c.target
}

func (c *Client) Start(ctx context.Context) error {
	c.startOnce.Do(func() {
		go c.run(ctx)
	})
	return nil
}

func (c *Client) Publish(ctx context.Context, msg session.Message) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.outgoing <- msg:
		return nil
	}
}

func (c *Client) Commands() <-chan session.Message {
	return c.commands
}

func (c *Client) run(ctx context.Context) {
	var pending *session.Message

	for {
		if ctx.Err() != nil {
			return
		}

		conn, _, err := c.wsDialer.DialContext(ctx, c.target, nil)
		if err != nil {
			c.logger.Printf("websocket connect failed target=%s err=%v", c.target, err)
			if !sleepContext(ctx, c.reconnectInterval) {
				return
			}
			continue
		}

		c.logger.Printf("websocket connected target=%s", c.target)
		pending, err = c.handleConnection(ctx, conn, pending)
		conn.Close()
		if err != nil && !errors.Is(err, context.Canceled) {
			c.logger.Printf("websocket connection dropped target=%s err=%v", c.target, err)
		}

		if !sleepContext(ctx, c.reconnectInterval) {
			return
		}
	}
}

func (c *Client) handleConnection(ctx context.Context, conn *websocket.Conn, pending *session.Message) (*session.Message, error) {
	readErr := make(chan error, 1)
	go c.readLoop(ctx, conn, readErr)

	var ticker *time.Ticker
	var heartbeatC <-chan time.Time
	if c.heartbeatInterval > 0 {
		ticker = time.NewTicker(c.heartbeatInterval)
		heartbeatC = ticker.C
		defer ticker.Stop()
	}

	for {
		if pending != nil {
			if err := conn.WriteJSON(*pending); err != nil {
				return pending, err
			}
			pending = nil
			continue
		}

		select {
		case <-ctx.Done():
			_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "shutdown"), time.Now().Add(time.Second))
			return nil, ctx.Err()
		case err := <-readErr:
			return nil, err
		case msg := <-c.outgoing:
			pending = &msg
		case <-heartbeatC:
			if err := conn.WriteJSON(c.newHeartbeat()); err != nil {
				return nil, err
			}
		}
	}
}

func (c *Client) readLoop(ctx context.Context, conn *websocket.Conn, readErr chan<- error) {
	for {
		var msg session.Message
		if err := conn.ReadJSON(&msg); err != nil {
			select {
			case readErr <- err:
			default:
			}
			return
		}

		if msg.MessageType != session.MessageTypeCommand {
			continue
		}

		select {
		case <-ctx.Done():
			select {
			case readErr <- ctx.Err():
			default:
			}
			return
		case c.commands <- msg:
		}
	}
}

func sleepContext(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func serverURLToWebSocket(raw string) string {
	switch {
	case strings.HasPrefix(raw, "https://"):
		return "wss://" + strings.TrimPrefix(raw, "https://")
	case strings.HasPrefix(raw, "http://"):
		return "ws://" + strings.TrimPrefix(raw, "http://")
	default:
		return raw
	}
}

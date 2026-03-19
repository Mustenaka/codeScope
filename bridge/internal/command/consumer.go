package command

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

type PromptTarget interface {
	SendPrompt(ctx context.Context, content string) error
}

type InboxConsumer struct {
	path         string
	statePath    string
	target       PromptTarget
	logger       *log.Logger
	pollInterval time.Duration
	offset       int64
	stateLoaded  bool
	lastLineHash string
	lastLineSize int64
}

type promptRecord struct {
	Payload struct {
		Content string `json:"content"`
	} `json:"payload"`
}

type consumerState struct {
	Offset       int64  `json:"offset"`
	LastLineHash string `json:"last_line_hash,omitempty"`
	LastLineSize int64  `json:"last_line_size,omitempty"`
}

func NewInboxConsumer(path, statePath string, target PromptTarget, logger *log.Logger) *InboxConsumer {
	if logger == nil {
		logger = log.Default()
	}
	consumer := &InboxConsumer{
		path:         path,
		statePath:    statePath,
		target:       target,
		logger:       logger,
		pollInterval: 500 * time.Millisecond,
	}
	_ = consumer.loadState()
	return consumer
}

func (c *InboxConsumer) Run(ctx context.Context) error {
	if c.target == nil {
		return nil
	}
	if err := c.ensureStateLoaded(); err != nil {
		return err
	}

	ticker := time.NewTicker(c.pollInterval)
	defer ticker.Stop()

	for {
		if err := c.drain(ctx); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (c *InboxConsumer) drain(ctx context.Context) error {
	if err := c.ensureStateLoaded(); err != nil {
		return err
	}

	file, err := os.Open(c.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer file.Close()

	if err := c.ensureOffsetStillValid(file); err != nil {
		return err
	}

	if _, err := file.Seek(c.offset, io.SeekStart); err != nil {
		return err
	}

	reader := bufio.NewReader(file)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			c.offset += int64(len(line))
			c.lastLineHash = hashLine(line)
			c.lastLineSize = int64(len(line))
			if err := c.handleLine(ctx, line); err != nil {
				return err
			}
			if err := c.persistState(); err != nil {
				return err
			}
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
	}
}

func (c *InboxConsumer) handleLine(ctx context.Context, line []byte) error {
	var record promptRecord
	if err := json.Unmarshal(line, &record); err != nil {
		c.logger.Printf("skip malformed prompt inbox record err=%v", err)
		return nil
	}
	if record.Payload.Content == "" {
		return nil
	}
	return c.target.SendPrompt(ctx, record.Payload.Content)
}

func (c *InboxConsumer) ensureStateLoaded() error {
	if c.stateLoaded {
		return nil
	}
	if err := c.loadState(); err != nil {
		return err
	}
	c.stateLoaded = true
	return nil
}

func (c *InboxConsumer) loadState() error {
	if c.statePath == "" {
		return nil
	}
	data, err := os.ReadFile(c.statePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	var state consumerState
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}
	c.offset = state.Offset
	c.lastLineHash = state.LastLineHash
	c.lastLineSize = state.LastLineSize
	return nil
}

func (c *InboxConsumer) persistState() error {
	if c.statePath == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(c.statePath), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(consumerState{
		Offset:       c.offset,
		LastLineHash: c.lastLineHash,
		LastLineSize: c.lastLineSize,
	})
	if err != nil {
		return err
	}
	return os.WriteFile(c.statePath, data, 0o644)
}

func (c *InboxConsumer) ensureOffsetStillValid(file *os.File) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.Size() < c.offset {
		c.resetState()
		return c.persistState()
	}
	if c.offset == 0 || c.lastLineSize == 0 || c.lastLineHash == "" {
		return nil
	}

	start := c.offset - c.lastLineSize
	if start < 0 {
		start = 0
	}
	if _, err := file.Seek(start, io.SeekStart); err != nil {
		return err
	}
	buf := make([]byte, c.lastLineSize)
	if _, err := io.ReadFull(file, buf); err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			c.resetState()
			return c.persistState()
		}
		return err
	}
	if hashLine(buf) != c.lastLineHash {
		c.logger.Printf("prompt inbox changed before offset=%d, resetting consumer state", c.offset)
		c.resetState()
		return c.persistState()
	}
	return nil
}

func (c *InboxConsumer) resetState() {
	c.offset = 0
	c.lastLineHash = ""
	c.lastLineSize = 0
}

func hashLine(line []byte) string {
	sum := sha256.Sum256(line)
	return fmt.Sprintf("%x", sum[:])
}

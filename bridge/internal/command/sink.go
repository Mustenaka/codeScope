package command

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"codescope/bridge/internal/session"
)

type PromptSink interface {
	WritePrompt(ctx context.Context, msg session.Message) (map[string]any, error)
}

type FilePromptSink struct {
	path string
	mu   sync.Mutex
}

type UnsupportedPromptSink struct {
	reason string
}

func NewFilePromptSink(path string) *FilePromptSink {
	return &FilePromptSink{path: path}
}

func NewUnsupportedPromptSink(reason string) *UnsupportedPromptSink {
	return &UnsupportedPromptSink{reason: reason}
}

func (s *UnsupportedPromptSink) WritePrompt(_ context.Context, _ session.Message) (map[string]any, error) {
	return nil, fmt.Errorf("%s", s.reason)
}

func (s *FilePromptSink) WritePrompt(_ context.Context, msg session.Message) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record := map[string]any{
		"command_id":   msg.CommandID,
		"command_type": msg.CommandType,
		"session_id":   msg.SessionID,
		"timestamp":    msg.Timestamp,
		"payload":      msg.Payload,
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return nil, fmt.Errorf("create prompt inbox dir: %w", err)
	}

	data, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("marshal prompt record: %w", err)
	}
	data = append(data, '\n')

	file, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open prompt inbox: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return nil, fmt.Errorf("write prompt inbox: %w", err)
	}

	return map[string]any{
		"accepted":   true,
		"local_path": s.path,
	}, nil
}

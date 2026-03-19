package capture

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"codescope/bridge/internal/session"
)

type CodexSessionSource struct {
	codexRoot    string
	machineID    string
	pollInterval time.Duration
	logger       *log.Logger
	states       map[string]*codexSessionFileState
	titles       map[string]string
}

func (s *CodexSessionSource) Name() string {
	return "codex_session_file"
}

type codexSessionFileState struct {
	offset      int64
	meta        session.Metadata
	threadTitle string
	threadID    string
}

type codexIndexEntry struct {
	ID         string `json:"id"`
	ThreadName string `json:"thread_name"`
}

type codexEnvelope struct {
	Timestamp string          `json:"timestamp"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
}

type codexSessionMetaPayload struct {
	ID  string `json:"id"`
	Cwd string `json:"cwd"`
}

type codexEventMessagePayload struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func NewCodexSessionSource(codexRoot, machineID string, pollInterval time.Duration, logger *log.Logger) *CodexSessionSource {
	if logger == nil {
		logger = log.Default()
	}
	if pollInterval <= 0 {
		pollInterval = time.Second
	}
	return &CodexSessionSource{
		codexRoot:    codexRoot,
		machineID:    machineID,
		pollInterval: pollInterval,
		logger:       logger,
		states:       make(map[string]*codexSessionFileState),
		titles:       make(map[string]string),
	}
}

func (s *CodexSessionSource) Start(ctx context.Context, sink Sink) error {
	if err := s.scanOnce(ctx, sink); err != nil {
		return err
	}

	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := s.scanOnce(ctx, sink); err != nil {
				return err
			}
		}
	}
}

func (s *CodexSessionSource) scanOnce(ctx context.Context, sink Sink) error {
	s.titles = s.loadTitles()

	files, err := s.listSessionFiles()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	seen := make(map[string]struct{}, len(files))
	for _, path := range files {
		seen[path] = struct{}{}
		if err := s.readFile(ctx, path, sink); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			s.logger.Printf("codex session capture skipped path=%q err=%v", path, err)
		}
	}

	for path := range s.states {
		if _, ok := seen[path]; !ok {
			delete(s.states, path)
		}
	}
	return nil
}

func (s *CodexSessionSource) loadTitles() map[string]string {
	indexPath := filepath.Join(s.codexRoot, "session_index.jsonl")
	file, err := os.Open(indexPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			s.logger.Printf("codex session index unavailable path=%q err=%v", indexPath, err)
		}
		return map[string]string{}
	}
	defer file.Close()

	titles := make(map[string]string)
	scanner := bufio.NewScanner(file)
	buffer := make([]byte, 0, 64*1024)
	scanner.Buffer(buffer, 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var entry codexIndexEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		if entry.ID == "" || strings.TrimSpace(entry.ThreadName) == "" {
			continue
		}
		titles[entry.ID] = strings.TrimSpace(entry.ThreadName)
	}
	if err := scanner.Err(); err != nil {
		s.logger.Printf("codex session index read failed path=%q err=%v", indexPath, err)
	}
	return titles
}

func (s *CodexSessionSource) listSessionFiles() ([]string, error) {
	sessionsRoot := filepath.Join(s.codexRoot, "sessions")
	if _, err := os.Stat(sessionsRoot); err != nil {
		return nil, err
	}

	files := make([]string, 0, 16)
	cutoff := time.Now().Add(-72 * time.Hour)
	err := filepath.WalkDir(sessionsRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".jsonl" {
			return nil
		}
		if _, tracked := s.states[path]; !tracked {
			info, infoErr := d.Info()
			if infoErr == nil && info.ModTime().Before(cutoff) {
				return nil
			}
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk codex sessions: %w", err)
	}
	return files, nil
}

func (s *CodexSessionSource) readFile(ctx context.Context, path string, sink Sink) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open codex session file: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat codex session file: %w", err)
	}

	state := s.states[path]
	if state == nil {
		state = &codexSessionFileState{}
		s.states[path] = state
	}
	if info.Size() < state.offset {
		state.offset = 0
	}

	if _, err := file.Seek(state.offset, io.SeekStart); err != nil {
		return fmt.Errorf("seek codex session file: %w", err)
	}

	reader := bufio.NewReader(file)
	offset := state.offset
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		lineBytes, err := reader.ReadBytes('\n')
		if len(lineBytes) == 0 && errors.Is(err, io.EOF) {
			break
		}
		offset += int64(len(lineBytes))
		line := strings.TrimSpace(string(lineBytes))
		if line != "" {
			if emitErr := s.handleLine(ctx, state, line, info.ModTime(), sink); emitErr != nil {
				return emitErr
			}
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("read codex session file: %w", err)
		}
	}

	state.offset = offset
	return nil
}

func (s *CodexSessionSource) handleLine(ctx context.Context, state *codexSessionFileState, line string, fallbackTime time.Time, sink Sink) error {
	var envelope codexEnvelope
	if err := json.Unmarshal([]byte(line), &envelope); err != nil {
		return fmt.Errorf("decode codex session line: %w", err)
	}

	switch envelope.Type {
	case "session_meta":
		var payload codexSessionMetaPayload
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
			return fmt.Errorf("decode codex session meta: %w", err)
		}
		if payload.ID == "" || strings.TrimSpace(payload.Cwd) == "" {
			return nil
		}
		state.meta = session.Metadata{
			AgentName:     "codex",
			WorkspaceRoot: filepath.ToSlash(filepath.Clean(strings.TrimSpace(payload.Cwd))),
			MachineID:     s.machineID,
			SessionID:     payload.ID,
		}
		if title := strings.TrimSpace(s.titles[state.meta.SessionID]); title != "" {
			state.threadTitle = title
			state.threadID = stableSemanticThreadID(state.meta.MachineID, state.meta.WorkspaceRoot, title)
		}
		return nil
	case "event_msg":
		if state.meta.SessionID == "" {
			return nil
		}
		var payload codexEventMessagePayload
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
			return fmt.Errorf("decode codex event payload: %w", err)
		}
		role, eventType, ok := codexMessageRole(payload.Type)
		if !ok {
			return nil
		}
		content := strings.TrimSpace(payload.Message)
		if content == "" {
			return nil
		}
		if state.threadTitle == "" {
			state.threadTitle = deriveSemanticThreadTitle(content)
		}
		if state.threadID == "" {
			state.threadID = stableSemanticThreadID(state.meta.MachineID, state.meta.WorkspaceRoot, state.threadTitle)
		}

		eventTime := parseCodexTimestamp(envelope.Timestamp, fallbackTime)
		payloadMap := map[string]any{
			"content":           content,
			"role":              role,
			"source":            "codex_session_file",
			"semantic_kind":     "thread_message",
			"thread_state":      codexThreadState(role),
			"thread_id":         state.threadID,
			"source_session_id": state.meta.SessionID,
		}
		if title := strings.TrimSpace(state.threadTitle); title != "" {
			payloadMap["thread_title"] = title
		}

		return sink.Emit(ctx, ObservedEvent{
			Meta: state.meta,
			Event: session.Event{
				Type: eventType,
				Payload: map[string]any{
					"content":           payloadMap["content"],
					"role":              payloadMap["role"],
					"source":            payloadMap["source"],
					"semantic_kind":     payloadMap["semantic_kind"],
					"thread_state":      payloadMap["thread_state"],
					"thread_id":         payloadMap["thread_id"],
					"source_session_id": payloadMap["source_session_id"],
					"thread_title":      payloadMap["thread_title"],
					"timestamp":         eventTime.UTC().Format(time.RFC3339Nano),
				},
			},
		})
	default:
		return nil
	}
}

func codexThreadState(role string) string {
	switch strings.TrimSpace(role) {
	case "assistant":
		// Session-file assistant turns are observed after the turn is fully emitted,
		// so the most useful default hint is that the thread is ready for the next prompt.
		return session.ThreadStateWaitingPrompt
	default:
		return session.ThreadStateRunning
	}
}

func codexMessageRole(messageType string) (role string, eventType string, ok bool) {
	switch strings.TrimSpace(messageType) {
	case "user_message":
		return "user", session.EventTypeCommand, true
	case "agent_message":
		return "assistant", session.EventTypeAIOutput, true
	default:
		return "", "", false
	}
}

func parseCodexTimestamp(value string, fallback time.Time) time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback.UTC()
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return fallback.UTC()
	}
	return parsed.UTC()
}

func deriveSemanticThreadTitle(content string) string {
	content = strings.Join(strings.Fields(strings.TrimSpace(content)), " ")
	if content == "" {
		return ""
	}
	runes := []rune(content)
	if len(runes) <= 72 {
		return content
	}
	return strings.TrimSpace(string(runes[:72]))
}

func stableSemanticThreadID(machineID, workspaceRoot, title string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return ""
	}
	sum := sha1.Sum([]byte(machineID + "|" + filepath.ToSlash(workspaceRoot) + "|" + strings.ToLower(title)))
	return "thread-" + fmt.Sprintf("%x", sum[:8])
}

package capture

import (
	"bufio"
	"context"
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

type ClaudeSessionSource struct {
	claudeRoot    string
	machineID     string
	pollInterval  time.Duration
	logger        *log.Logger
	states        map[string]*claudeSessionFileState
	threadTitles  map[string]string
	workspaceRoot map[string]string
}

type claudeSessionFileState struct {
	offset      int64
	meta        session.Metadata
	threadTitle string
	threadID    string
}

type claudeSessionsIndex struct {
	Entries []claudeSessionIndexEntry `json:"entries"`
}

type claudeSessionIndexEntry struct {
	SessionID   string `json:"sessionId"`
	FullPath    string `json:"fullPath"`
	ProjectPath string `json:"projectPath"`
	FirstPrompt string `json:"firstPrompt"`
}

type claudeTranscriptRecord struct {
	Type        string                  `json:"type"`
	Timestamp   string                  `json:"timestamp"`
	SessionID   string                  `json:"sessionId"`
	Cwd         string                  `json:"cwd"`
	IsSidechain bool                    `json:"isSidechain"`
	Message     claudeTranscriptMessage `json:"message"`
}

type claudeTranscriptMessage struct {
	Role       string                    `json:"role"`
	StopReason string                    `json:"stop_reason"`
	Content    []claudeTranscriptContent `json:"content"`
}

type claudeTranscriptContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func NewClaudeSessionSource(claudeRoot, machineID string, pollInterval time.Duration, logger *log.Logger) *ClaudeSessionSource {
	if logger == nil {
		logger = log.Default()
	}
	if pollInterval <= 0 {
		pollInterval = time.Second
	}
	return &ClaudeSessionSource{
		claudeRoot:    claudeRoot,
		machineID:     machineID,
		pollInterval:  pollInterval,
		logger:        logger,
		states:        make(map[string]*claudeSessionFileState),
		threadTitles:  make(map[string]string),
		workspaceRoot: make(map[string]string),
	}
}

func (s *ClaudeSessionSource) Name() string {
	return "claude_session_file"
}

func (s *ClaudeSessionSource) Start(ctx context.Context, sink Sink) error {
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

func (s *ClaudeSessionSource) scanOnce(ctx context.Context, sink Sink) error {
	s.threadTitles, s.workspaceRoot = s.loadIndexMetadata()

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
			s.logger.Printf("claude session capture skipped path=%q err=%v", path, err)
		}
	}

	for path := range s.states {
		if _, ok := seen[path]; !ok {
			delete(s.states, path)
		}
	}
	return nil
}

func (s *ClaudeSessionSource) loadIndexMetadata() (map[string]string, map[string]string) {
	projectsRoot := filepath.Join(s.claudeRoot, "projects")
	if _, err := os.Stat(projectsRoot); err != nil {
		return map[string]string{}, map[string]string{}
	}

	titles := make(map[string]string)
	workspaces := make(map[string]string)
	_ = filepath.WalkDir(projectsRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() || !strings.EqualFold(d.Name(), "sessions-index.json") {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			s.logger.Printf("claude session index unavailable path=%q err=%v", path, readErr)
			return nil
		}
		var index claudeSessionsIndex
		if decodeErr := json.Unmarshal(data, &index); decodeErr != nil {
			s.logger.Printf("claude session index decode failed path=%q err=%v", path, decodeErr)
			return nil
		}
		for _, entry := range index.Entries {
			sessionID := strings.TrimSpace(entry.SessionID)
			if sessionID == "" {
				continue
			}
			if title := deriveClaudeThreadTitle(entry.FirstPrompt); title != "" {
				titles[sessionID] = title
			}
			if workspace := normalizeClaudeWorkspace(entry.ProjectPath); workspace != "" {
				workspaces[sessionID] = workspace
			}
		}
		return nil
	})
	return titles, workspaces
}

func (s *ClaudeSessionSource) listSessionFiles() ([]string, error) {
	projectsRoot := filepath.Join(s.claudeRoot, "projects")
	if _, err := os.Stat(projectsRoot); err != nil {
		return nil, err
	}

	files := make([]string, 0, 16)
	cutoff := time.Now().Add(-72 * time.Hour)
	err := filepath.WalkDir(projectsRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := strings.ToLower(d.Name())
			if name == "subagents" || name == "tool-results" {
				return filepath.SkipDir
			}
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
		return nil, fmt.Errorf("walk claude sessions: %w", err)
	}
	return files, nil
}

func (s *ClaudeSessionSource) readFile(ctx context.Context, path string, sink Sink) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open claude session file: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat claude session file: %w", err)
	}

	state := s.states[path]
	if state == nil {
		state = &claudeSessionFileState{}
		s.states[path] = state
	}
	if info.Size() < state.offset {
		state.offset = 0
	}

	if _, err := file.Seek(state.offset, io.SeekStart); err != nil {
		return fmt.Errorf("seek claude session file: %w", err)
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
			return fmt.Errorf("read claude session file: %w", err)
		}
	}

	state.offset = offset
	return nil
}

func (s *ClaudeSessionSource) handleLine(ctx context.Context, state *claudeSessionFileState, line string, fallbackTime time.Time, sink Sink) error {
	var record claudeTranscriptRecord
	if err := json.Unmarshal([]byte(line), &record); err != nil {
		return fmt.Errorf("decode claude session line: %w", err)
	}
	if record.IsSidechain {
		return nil
	}

	role, eventType, ok := claudeMessageRole(record)
	if !ok {
		return nil
	}
	content := extractClaudeMessageText(record.Message.Content)
	if content == "" {
		return nil
	}

	sessionID := strings.TrimSpace(record.SessionID)
	if sessionID == "" {
		return nil
	}
	workspaceRoot := normalizeClaudeWorkspace(record.Cwd)
	if workspaceRoot == "" {
		workspaceRoot = s.workspaceRoot[sessionID]
	}
	if workspaceRoot == "" {
		return nil
	}

	state.meta = session.Metadata{
		AgentName:     "claude",
		WorkspaceRoot: workspaceRoot,
		MachineID:     s.machineID,
		SessionID:     sessionID,
	}
	if state.threadTitle == "" {
		state.threadTitle = strings.TrimSpace(s.threadTitles[sessionID])
	}
	if state.threadTitle == "" && role == "user" {
		state.threadTitle = deriveSemanticThreadTitle(content)
	}
	if state.threadID == "" && state.threadTitle != "" {
		state.threadID = stableSemanticThreadID(state.meta.MachineID, state.meta.WorkspaceRoot, state.threadTitle)
	}

	payload := map[string]any{
		"content":           content,
		"role":              role,
		"source":            s.Name(),
		"semantic_kind":     "thread_message",
		"thread_state":      claudeThreadState(role, record.Message.StopReason),
		"source_session_id": state.meta.SessionID,
		"timestamp":         parseCodexTimestamp(record.Timestamp, fallbackTime).UTC().Format(time.RFC3339Nano),
	}
	if state.threadID != "" {
		payload["thread_id"] = state.threadID
	}
	if title := strings.TrimSpace(state.threadTitle); title != "" {
		payload["thread_title"] = title
	}

	return sink.Emit(ctx, ObservedEvent{
		Meta: state.meta,
		Event: session.Event{
			Type:    eventType,
			Payload: payload,
		},
	})
}

func claudeMessageRole(record claudeTranscriptRecord) (role string, eventType string, ok bool) {
	switch strings.TrimSpace(record.Type) {
	case "user":
		if strings.TrimSpace(record.Message.Role) != "user" {
			return "", "", false
		}
		return "user", session.EventTypeCommand, true
	case "assistant":
		if strings.TrimSpace(record.Message.Role) != "assistant" {
			return "", "", false
		}
		return "assistant", session.EventTypeAIOutput, true
	default:
		return "", "", false
	}
}

func extractClaudeMessageText(items []claudeTranscriptContent) string {
	parts := make([]string, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.Type) != "text" {
			continue
		}
		text := strings.TrimSpace(item.Text)
		if text == "" || isClaudeNoiseText(text) {
			continue
		}
		parts = append(parts, text)
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}

func isClaudeNoiseText(text string) bool {
	return strings.HasPrefix(text, "<ide_opened_file>") && strings.HasSuffix(text, "</ide_opened_file>")
}

func deriveClaudeThreadTitle(firstPrompt string) string {
	lines := make([]string, 0, 2)
	for _, line := range strings.Split(firstPrompt, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || isClaudeNoiseText(line) {
			continue
		}
		lines = append(lines, line)
	}
	return deriveSemanticThreadTitle(strings.Join(lines, " "))
}

func normalizeClaudeWorkspace(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return filepath.ToSlash(filepath.Clean(value))
}

func claudeThreadState(role, stopReason string) string {
	if strings.TrimSpace(role) != "assistant" {
		return session.ThreadStateRunning
	}
	switch strings.TrimSpace(stopReason) {
	case "end_turn":
		return session.ThreadStateWaitingPrompt
	case "pause_turn":
		return session.ThreadStateWaitingReview
	default:
		return session.ThreadStateRunning
	}
}

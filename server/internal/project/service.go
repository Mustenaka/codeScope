package project

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"codescope/server/internal/event"
	"codescope/server/internal/session"
)

var ErrNotFound = errors.New("project resource not found")

type SessionReader interface {
	List() ([]session.Session, error)
}

type EventReader interface {
	ListBySession(sessionID string) ([]event.Record, error)
}

type ThreadState string

const (
	ThreadStateRunning       ThreadState = "running"
	ThreadStateWaitingPrompt ThreadState = "waiting_prompt"
	ThreadStateWaitingReview ThreadState = "waiting_review"
	ThreadStateCompleted     ThreadState = "completed"
	ThreadStateBlocked       ThreadState = "blocked"
)

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
)

type Project struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	WorkspaceRoot      string    `json:"workspace_root"`
	MachineID          string    `json:"machine_id"`
	ThreadCount        int       `json:"thread_count"`
	RunningThreadCount int       `json:"running_thread_count"`
	LastActivityAt     time.Time `json:"last_activity_at,omitempty"`
}

type Thread struct {
	ID             string      `json:"id"`
	ProjectID      string      `json:"project_id"`
	SessionID      string      `json:"session_id"`
	Title          string      `json:"title"`
	AgentKind      string      `json:"agent_kind,omitempty"`
	Status         ThreadState `json:"status"`
	Summary        string      `json:"summary,omitempty"`
	LastActivityAt time.Time   `json:"last_activity_at,omitempty"`
	StartedAt      time.Time   `json:"started_at,omitempty"`
	EndedAt        time.Time   `json:"ended_at,omitempty"`
}

type Message struct {
	ID         string    `json:"id"`
	ThreadID   string    `json:"thread_id"`
	Role       Role      `json:"role"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	Sequence   int       `json:"sequence"`
	SourceType string    `json:"source_type,omitempty"`
}

type Service struct {
	sessions SessionReader
	events   EventReader
}

const activeThreadWindow = 20 * time.Minute

type threadSnapshot struct {
	session        session.Session
	records        []event.Record
	projectID      string
	threadID       string
	title          string
	agentKind      string
	status         ThreadState
	summary        string
	lastActivityAt time.Time
	startedAt      time.Time
	endedAt        time.Time
}

func NewService(sessions SessionReader, events EventReader) *Service {
	return &Service{
		sessions: sessions,
		events:   events,
	}
}

func ProjectID(item session.Session) string {
	sum := sha1.Sum([]byte(item.MachineID + "|" + item.WorkspaceRoot))
	return "project-" + hex.EncodeToString(sum[:8])
}

func ThreadID(item session.Session) string {
	return item.ID
}

func (s *Service) ListProjects() ([]Project, error) {
	sessions, err := s.sessions.List()
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}

	threadsByProject, err := s.collectVisibleThreads(sessions)
	if err != nil {
		return nil, err
	}
	grouped := make(map[string]*Project)
	for _, item := range sessions {
		id := ProjectID(item)
		current, ok := grouped[id]
		if !ok {
			grouped[id] = &Project{
				ID:            id,
				Name:          firstNonEmpty(item.ProjectName, baseName(item.WorkspaceRoot), item.ID),
				WorkspaceRoot: item.WorkspaceRoot,
				MachineID:     item.MachineID,
				LastActivityAt: latestTime(
					item.LastActivityAt,
					item.UpdatedAt,
					item.CreatedAt,
				),
			}
			current = grouped[id]
		}
		if activity := latestTime(item.LastActivityAt, item.UpdatedAt, item.CreatedAt); activity.After(current.LastActivityAt) {
			current.LastActivityAt = activity
			if item.ProjectName != "" {
				current.Name = item.ProjectName
			}
		}
	}
	for projectID, threads := range threadsByProject {
		current, ok := grouped[projectID]
		if !ok {
			continue
		}
		current.ThreadCount = len(threads)
		current.RunningThreadCount = 0
		for _, thread := range threads {
			if thread.Status == ThreadStateRunning {
				current.RunningThreadCount++
			}
			if thread.LastActivityAt.After(current.LastActivityAt) {
				current.LastActivityAt = thread.LastActivityAt
			}
		}
	}

	projects := make([]Project, 0, len(grouped))
	for _, item := range grouped {
		projects = append(projects, *item)
	}
	sort.Slice(projects, func(i, j int) bool {
		if projects[i].LastActivityAt.Equal(projects[j].LastActivityAt) {
			return projects[i].Name < projects[j].Name
		}
		return projects[i].LastActivityAt.After(projects[j].LastActivityAt)
	})
	return projects, nil
}

func (s *Service) GetProject(projectID string) (Project, error) {
	projects, err := s.ListProjects()
	if err != nil {
		return Project{}, err
	}
	for _, item := range projects {
		if item.ID == projectID {
			return item, nil
		}
	}
	return Project{}, ErrNotFound
}

func (s *Service) ListThreads(projectID string) ([]Thread, error) {
	sessions, err := s.sessions.List()
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	threadsByProject, err := s.collectVisibleThreads(sessions)
	if err != nil {
		return nil, err
	}
	return threadsByProject[projectID], nil
}

func (s *Service) GetThread(threadID string) (Thread, error) {
	sessions, err := s.sessions.List()
	if err != nil {
		return Thread{}, fmt.Errorf("list sessions: %w", err)
	}
	threadsByProject, err := s.collectVisibleThreads(sessions)
	if err != nil {
		return Thread{}, err
	}
	for _, threads := range threadsByProject {
		for _, thread := range threads {
			if thread.ID == threadID {
				return thread, nil
			}
		}
	}
	return Thread{}, ErrNotFound
}

func (s *Service) ListMessages(threadID string) ([]Message, error) {
	thread, err := s.GetThread(threadID)
	if err != nil {
		return nil, err
	}
	records, err := s.events.ListBySession(thread.SessionID)
	if err != nil {
		return nil, fmt.Errorf("list events for session %s: %w", thread.SessionID, err)
	}

	sessions, err := s.sessions.List()
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}

	messages := make([]Message, 0, len(records))
	for _, item := range sessions {
		if ProjectID(item) != thread.ProjectID {
			continue
		}
		sessionRecords, listErr := s.events.ListBySession(item.ID)
		if listErr != nil {
			return nil, fmt.Errorf("list events for session %s: %w", item.ID, listErr)
		}
		if deriveThreadID(item, sessionRecords) != thread.ID {
			continue
		}
		for _, record := range sessionRecords {
			message, ok := mapRecordToMessage(thread.ID, record)
			if !ok {
				continue
			}
			messages = append(messages, message)
		}
	}

	sort.Slice(messages, func(i, j int) bool {
		if messages[i].CreatedAt.Equal(messages[j].CreatedAt) {
			return messages[i].ID < messages[j].ID
		}
		return messages[i].CreatedAt.Before(messages[j].CreatedAt)
	})
	for i := range messages {
		messages[i].Sequence = i + 1
	}
	return messages, nil
}

func (s *Service) collectVisibleThreads(sessions []session.Session) (map[string][]Thread, error) {
	snapshots := make([]threadSnapshot, 0, len(sessions))
	for _, item := range sessions {
		records, err := s.events.ListBySession(item.ID)
		if err != nil {
			return nil, fmt.Errorf("list events for session %s: %w", item.ID, err)
		}
		snapshot := buildThreadSnapshot(item, records)
		if !isReleaseVisibleSnapshot(snapshot) {
			continue
		}
		snapshots = append(snapshots, snapshot)
	}

	grouped := make(map[string]map[string]*Thread)
	for _, snapshot := range snapshots {
		projectID := snapshot.projectID
		if _, ok := grouped[projectID]; !ok {
			grouped[projectID] = make(map[string]*Thread)
		}
		existing, ok := grouped[projectID][snapshot.threadID]
		if !ok {
			grouped[projectID][snapshot.threadID] = &Thread{
				ID:             snapshot.threadID,
				ProjectID:      snapshot.projectID,
				SessionID:      snapshot.session.ID,
				Title:          snapshot.title,
				AgentKind:      snapshot.agentKind,
				Status:         snapshot.status,
				Summary:        snapshot.summary,
				LastActivityAt: snapshot.lastActivityAt,
				StartedAt:      snapshot.startedAt,
				EndedAt:        snapshot.endedAt,
			}
			continue
		}
		if snapshot.lastActivityAt.After(existing.LastActivityAt) {
			existing.SessionID = snapshot.session.ID
			existing.Title = firstNonEmpty(snapshot.title, existing.Title)
			existing.AgentKind = firstNonEmpty(snapshot.agentKind, existing.AgentKind)
			existing.Status = snapshot.status
			existing.Summary = firstNonEmpty(snapshot.summary, existing.Summary)
			existing.LastActivityAt = snapshot.lastActivityAt
			if !snapshot.endedAt.IsZero() {
				existing.EndedAt = snapshot.endedAt
			}
		}
		if existing.StartedAt.IsZero() || (!snapshot.startedAt.IsZero() && snapshot.startedAt.Before(existing.StartedAt)) {
			existing.StartedAt = snapshot.startedAt
		}
		if existing.Summary == "" {
			existing.Summary = snapshot.summary
		}
	}

	result := make(map[string][]Thread, len(grouped))
	for projectID, threadsByID := range grouped {
		threads := make([]Thread, 0, len(threadsByID))
		for _, thread := range threadsByID {
			threads = append(threads, *thread)
		}
		sort.Slice(threads, func(i, j int) bool {
			if threads[i].LastActivityAt.Equal(threads[j].LastActivityAt) {
				return threads[i].ID < threads[j].ID
			}
			return threads[i].LastActivityAt.After(threads[j].LastActivityAt)
		})
		result[projectID] = threads
	}
	return result, nil
}

func buildThreadSnapshot(item session.Session, records []event.Record) threadSnapshot {
	return threadSnapshot{
		session:        item,
		records:        records,
		projectID:      ProjectID(item),
		threadID:       deriveThreadID(item, records),
		title:          buildThreadTitle(item, records),
		agentKind:      deriveAgentKind(records),
		status:         deriveThreadState(item, records),
		summary:        deriveSummary(records),
		lastActivityAt: latestTime(item.LastActivityAt, item.UpdatedAt, item.CreatedAt),
		startedAt:      item.StartedAt,
		endedAt:        item.EndedAt,
	}
}

func isReleaseVisibleSnapshot(snapshot threadSnapshot) bool {
	if hasReadableMessages(snapshot.records) {
		return true
	}
	return !isDebugOnlyThread(snapshot)
}

func isDebugOnlyThread(snapshot threadSnapshot) bool {
	if snapshot.summary != "" {
		return false
	}
	if isSyntheticFallbackTitle(snapshot.title, snapshot.session.ID) {
		return true
	}
	for _, record := range snapshot.records {
		if isMeaningfulNonDebugRecord(record) {
			return false
		}
	}
	return true
}

func isSyntheticFallbackTitle(title, sessionID string) bool {
	return title == fallbackThreadTitle(session.Session{ID: sessionID})
}

func hasReadableMessages(records []event.Record) bool {
	for _, record := range records {
		if _, ok := userSummary(record); ok {
			return true
		}
		if _, ok := assistantSummary(record); ok {
			return true
		}
	}
	return false
}

func isMeaningfulNonDebugRecord(record event.Record) bool {
	switch {
	case record.MessageType == event.MessageTypeCommand && payloadRole(record.Payload) == string(RoleUser):
		return true
	case record.MessageType == event.MessageTypeCommand && record.CommandType == event.CommandTypeSendPrompt:
		return true
	case record.MessageType == event.MessageTypeEvent && record.EventType == event.TypeAIOutput:
		return true
	case record.MessageType == event.MessageTypeEvent && record.EventType == event.TypeCommand && payloadRole(record.Payload) == string(RoleUser):
		return true
	default:
		return false
	}
}

func buildThreadTitle(item session.Session, records []event.Record) string {
	if title := deriveThreadTitleText(records); title != "" {
		return title
	}
	return fallbackThreadTitle(item)
}

func deriveAgentKind(records []event.Record) string {
	for i := len(records) - 1; i >= 0; i-- {
		if value, ok := records[i].Payload["agent_name"].(string); ok && value != "" {
			return value
		}
	}
	return ""
}

func deriveThreadID(item session.Session, records []event.Record) string {
	for i := len(records) - 1; i >= 0; i-- {
		if value, ok := payloadString(records[i].Payload, "thread_id"); ok {
			return value
		}
	}
	return ThreadID(item)
}

func deriveSummary(records []event.Record) string {
	for i := len(records) - 1; i >= 0; i-- {
		if value, ok := assistantSummary(records[i]); ok {
			return value
		}
	}
	for i := len(records) - 1; i >= 0; i-- {
		if value, ok := userSummary(records[i]); ok {
			return value
		}
	}
	for i := len(records) - 1; i >= 0; i-- {
		if value, ok := commandResultSummary(records[i]); ok {
			return value
		}
	}
	return ""
}

func deriveThreadTitleText(records []event.Record) string {
	for i := len(records) - 1; i >= 0; i-- {
		if value, ok := payloadString(records[i].Payload, "thread_title"); ok {
			return clampTitle(value)
		}
	}
	for _, record := range records {
		if value, ok := userSummary(record); ok {
			return clampTitle(value)
		}
	}
	for i := len(records) - 1; i >= 0; i-- {
		if value, ok := assistantSummary(records[i]); ok {
			return clampTitle(value)
		}
	}
	return ""
}

func clampTitle(value string) string {
	value = strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	if value == "" {
		return ""
	}
	const maxRunes = 72
	runes := []rune(value)
	if len(runes) <= maxRunes {
		return value
	}
	return strings.TrimSpace(string(runes[:maxRunes])) + "..."
}

func fallbackThreadTitle(item session.Session) string {
	suffix := item.ID
	if len(suffix) > 8 {
		suffix = suffix[len(suffix)-8:]
	}
	return "Thread " + suffix
}

func assistantSummary(record event.Record) (string, bool) {
	switch {
	case record.MessageType == event.MessageTypeEvent && record.EventType == event.TypeAIOutput:
		return payloadContent(record.Payload)
	case record.MessageType == event.MessageTypeEvent && record.EventType == event.TypeTerminalOutput:
		if observed, _ := record.Payload["observed"].(bool); observed {
			return "", false
		}
		return payloadContent(record.Payload)
	default:
		return "", false
	}
}

func userSummary(record event.Record) (string, bool) {
	if record.MessageType == event.MessageTypeCommand && record.CommandType == event.CommandTypeSendPrompt {
		return payloadString(record.Payload, "content")
	}
	if record.MessageType == event.MessageTypeEvent &&
		record.EventType == event.TypeCommand &&
		payloadRole(record.Payload) == string(RoleUser) {
		return payloadString(record.Payload, "content")
	}
	return "", false
}

func commandResultSummary(record event.Record) (string, bool) {
	if record.MessageType != event.MessageTypeCommandResult {
		return "", false
	}
	if value, ok := payloadString(record.Payload, "result"); ok {
		return value, true
	}
	return payloadString(record.Payload, "error")
}

func deriveThreadState(item session.Session, records []event.Record) ThreadState {
	now := time.Now().UTC()
	for i := len(records) - 1; i >= 0; i-- {
		record := records[i]
		if value, ok := payloadString(record.Payload, "thread_state"); ok {
			switch value {
			case string(ThreadStateRunning):
				return ThreadStateRunning
			case string(ThreadStateWaitingPrompt):
				return ThreadStateWaitingPrompt
			case string(ThreadStateWaitingReview):
				return ThreadStateWaitingReview
			case string(ThreadStateCompleted):
				return ThreadStateCompleted
			case string(ThreadStateBlocked):
				return ThreadStateBlocked
			}
		}
		if record.EventType == event.TypeError {
			return ThreadStateBlocked
		}
		if record.MessageType == event.MessageTypeCommandResult &&
			record.CommandType == event.CommandTypeSendPrompt &&
			record.Status == event.CommandStatusFailed {
			return ThreadStateWaitingPrompt
		}
	}

	switch item.Status {
	case session.StatusFailed:
		return ThreadStateBlocked
	case session.StatusStopped:
		return ThreadStateCompleted
	case session.StatusCreated:
		return ThreadStateWaitingPrompt
	default:
		if activity := latestTime(item.LastActivityAt, item.UpdatedAt, item.CreatedAt); !activity.IsZero() && now.Sub(activity) > activeThreadWindow {
			if hasReadableMessages(records) {
				return ThreadStateWaitingPrompt
			}
			return ThreadStateCompleted
		}
		return ThreadStateRunning
	}
}

func mapRecordToMessage(threadID string, record event.Record) (Message, bool) {
	switch {
	case record.MessageType == event.MessageTypeCommand && record.CommandType == event.CommandTypeSendPrompt:
		content, ok := payloadString(record.Payload, "content")
		if !ok {
			return Message{}, false
		}
		return Message{
			ID:         record.ID,
			ThreadID:   threadID,
			Role:       RoleUser,
			Content:    content,
			CreatedAt:  record.Timestamp,
			SourceType: string(record.MessageType),
		}, true
	case record.MessageType == event.MessageTypeEvent &&
		record.EventType == event.TypeCommand &&
		payloadRole(record.Payload) == string(RoleUser):
		content, ok := payloadString(record.Payload, "content")
		if !ok {
			return Message{}, false
		}
		return Message{
			ID:         record.ID,
			ThreadID:   threadID,
			Role:       RoleUser,
			Content:    content,
			CreatedAt:  record.Timestamp,
			SourceType: string(record.EventType),
		}, true
	case record.MessageType == event.MessageTypeEvent && record.EventType == event.TypeAIOutput:
		content, ok := payloadContent(record.Payload)
		if !ok {
			return Message{}, false
		}
		return Message{
			ID:         record.ID,
			ThreadID:   threadID,
			Role:       RoleAssistant,
			Content:    content,
			CreatedAt:  record.Timestamp,
			SourceType: string(record.EventType),
		}, true
	case record.MessageType == event.MessageTypeEvent && record.EventType == event.TypeTerminalOutput:
		if observed, _ := record.Payload["observed"].(bool); observed {
			return Message{}, false
		}
		content, ok := payloadContent(record.Payload)
		if !ok {
			return Message{}, false
		}
		return Message{
			ID:         record.ID,
			ThreadID:   threadID,
			Role:       RoleAssistant,
			Content:    content,
			CreatedAt:  record.Timestamp,
			SourceType: string(record.EventType),
		}, true
	default:
		return Message{}, false
	}
}

func payloadContent(payload map[string]any) (string, bool) {
	return payloadString(payload, "content")
}

func payloadString(payload map[string]any, key string) (string, bool) {
	value, ok := payload[key].(string)
	if !ok {
		return "", false
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", false
	}
	return value, true
}

func payloadRole(payload map[string]any) string {
	value, _ := payload["role"].(string)
	return strings.TrimSpace(value)
}

func latestTime(values ...time.Time) time.Time {
	var latest time.Time
	for _, value := range values {
		if value.After(latest) {
			latest = value
		}
	}
	return latest
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func baseName(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")
	path = strings.TrimRight(path, "/")
	if path == "" {
		return ""
	}
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

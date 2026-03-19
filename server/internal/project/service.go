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
var ErrNoWritableSession = errors.New("project has no writable bridge session")

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
	ThreadStateOffline       ThreadState = "offline"
	ThreadStateStale         ThreadState = "stale"
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
	CreatedAt          time.Time `json:"created_at,omitempty"`
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
	AgentKind  string    `json:"agent_kind,omitempty"`
	SourceType string    `json:"source_type,omitempty"`
}

type Service struct {
	sessions SessionReader
	events   EventReader
	now      func() time.Time
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

type CreateThreadInput struct {
	Content string `json:"content" binding:"required"`
}

type ThreadLaunch struct {
	Thread           Thread `json:"thread"`
	BackingSessionID string `json:"backing_session_id"`
}

type ThreadExecution struct {
	Thread            Thread   `json:"thread"`
	WritableSessionID string   `json:"writable_session_id"`
	SessionIDs        []string `json:"session_ids"`
}

func NewService(sessions SessionReader, events EventReader) *Service {
	return &Service{
		sessions: sessions,
		events:   events,
		now:      time.Now().UTC,
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
				CreatedAt:     item.CreatedAt,
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
		if current.CreatedAt.IsZero() || (!item.CreatedAt.IsZero() && item.CreatedAt.Before(current.CreatedAt)) {
			current.CreatedAt = item.CreatedAt
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
			if message.AgentKind == "" && message.Role == RoleAssistant {
				message.AgentKind = firstNonEmpty(payloadAgentKind(record.Payload), thread.AgentKind)
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

func (s *Service) PrepareThreadLaunch(projectID string, input CreateThreadInput) (ThreadLaunch, error) {
	content := strings.TrimSpace(input.Content)
	if content == "" {
		return ThreadLaunch{}, errors.New("content is required")
	}

	sessions, err := s.sessions.List()
	if err != nil {
		return ThreadLaunch{}, fmt.Errorf("list sessions: %w", err)
	}

	projectFound := false
	var selected *session.Session
	for i := range sessions {
		item := sessions[i]
		if ProjectID(item) != projectID {
			continue
		}
		projectFound = true
		if !item.BridgeOnline {
			continue
		}
		if selected == nil || betterWritableSession(item, *selected) {
			candidate := item
			selected = &candidate
		}
	}
	if !projectFound {
		return ThreadLaunch{}, ErrNotFound
	}
	if selected == nil {
		return ThreadLaunch{}, ErrNoWritableSession
	}

	now := s.now()
	title := clampTitle(content)
	threadID := newThreadID(projectID, selected.ID, content, now)
	thread := Thread{
		ID:             threadID,
		ProjectID:      projectID,
		SessionID:      selected.ID,
		Title:          title,
		Status:         ThreadStateRunning,
		Summary:        content,
		LastActivityAt: now,
		StartedAt:      now,
	}
	return ThreadLaunch{
		Thread:           thread,
		BackingSessionID: selected.ID,
	}, nil
}

func (s *Service) ResolveThreadExecution(threadID string) (ThreadExecution, error) {
	sessions, err := s.sessions.List()
	if err != nil {
		return ThreadExecution{}, fmt.Errorf("list sessions: %w", err)
	}

	snapshots := make([]threadSnapshot, 0, len(sessions))
	for _, item := range sessions {
		records, listErr := s.events.ListBySession(item.ID)
		if listErr != nil {
			return ThreadExecution{}, fmt.Errorf("list events for session %s: %w", item.ID, listErr)
		}
		snapshot := buildThreadSnapshot(item, records, s.now())
		if snapshot.threadID != threadID {
			continue
		}
		snapshots = append(snapshots, snapshot)
	}
	if len(snapshots) == 0 {
		return ThreadExecution{}, ErrNotFound
	}

	sort.Slice(snapshots, func(i, j int) bool {
		if snapshots[i].lastActivityAt.Equal(snapshots[j].lastActivityAt) {
			return snapshots[i].session.ID < snapshots[j].session.ID
		}
		return snapshots[i].lastActivityAt.After(snapshots[j].lastActivityAt)
	})

	thread := Thread{
		ID:             threadID,
		ProjectID:      snapshots[0].projectID,
		SessionID:      snapshots[0].session.ID,
		Title:          snapshots[0].title,
		AgentKind:      snapshots[0].agentKind,
		Status:         snapshots[0].status,
		Summary:        snapshots[0].summary,
		LastActivityAt: snapshots[0].lastActivityAt,
		StartedAt:      snapshots[0].startedAt,
		EndedAt:        snapshots[0].endedAt,
	}

	sessionIDs := make([]string, 0, len(snapshots))
	var writable *session.Session
	for _, snapshot := range snapshots {
		sessionIDs = append(sessionIDs, snapshot.session.ID)
		if !snapshot.session.BridgeOnline {
			continue
		}
		if writable == nil || betterWritableSession(snapshot.session, *writable) {
			candidate := snapshot.session
			writable = &candidate
		}
	}
	if writable == nil {
		return ThreadExecution{
			Thread:     thread,
			SessionIDs: sessionIDs,
		}, ErrNoWritableSession
	}

	return ThreadExecution{
		Thread:            thread,
		WritableSessionID: writable.ID,
		SessionIDs:        sessionIDs,
	}, nil
}

func (s *Service) collectVisibleThreads(sessions []session.Session) (map[string][]Thread, error) {
	snapshots := make([]threadSnapshot, 0, len(sessions))
	for _, item := range sessions {
		records, err := s.events.ListBySession(item.ID)
		if err != nil {
			return nil, fmt.Errorf("list events for session %s: %w", item.ID, err)
		}
		snapshot := buildThreadSnapshot(item, records, s.now())
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

func buildThreadSnapshot(item session.Session, records []event.Record, now time.Time) threadSnapshot {
	return threadSnapshot{
		session:        item,
		records:        records,
		projectID:      ProjectID(item),
		threadID:       deriveThreadID(item, records),
		title:          buildThreadTitle(item, records),
		agentKind:      deriveAgentKind(records),
		status:         deriveThreadState(item, records, now),
		summary:        deriveSummary(records),
		lastActivityAt: latestTime(item.LastActivityAt, item.UpdatedAt, item.CreatedAt),
		startedAt:      item.StartedAt,
		endedAt:        item.EndedAt,
	}
}

func betterWritableSession(candidate, current session.Session) bool {
	candidateConnectedAt := latestTime(candidate.BridgeConnectedAt, candidate.LastActivityAt, candidate.UpdatedAt, candidate.CreatedAt)
	currentConnectedAt := latestTime(current.BridgeConnectedAt, current.LastActivityAt, current.UpdatedAt, current.CreatedAt)
	if !candidateConnectedAt.Equal(currentConnectedAt) {
		return candidateConnectedAt.After(currentConnectedAt)
	}
	if candidate.BridgeOnline != current.BridgeOnline {
		return candidate.BridgeOnline
	}
	return candidate.ID > current.ID
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
	var fallback string
	for i := len(records) - 1; i >= 0; i-- {
		if value, ok := semanticThreadID(records[i], item.ID); ok {
			return value
		}
		if fallback == "" {
			if value, ok := payloadString(records[i].Payload, "thread_id"); ok {
				fallback = value
			}
		}
	}
	if strings.TrimSpace(fallback) != "" {
		return fallback
	}
	return ThreadID(item)
}

func semanticThreadID(record event.Record, sessionID string) (string, bool) {
	threadID, ok := payloadString(record.Payload, "thread_id")
	if !ok {
		return "", false
	}
	if !isSessionFallbackThreadID(threadID, record.Payload, sessionID) {
		return threadID, true
	}
	return "", false
}

func isSessionFallbackThreadID(threadID string, payload map[string]any, sessionID string) bool {
	threadID = strings.TrimSpace(threadID)
	if threadID == "" {
		return true
	}
	if threadID == strings.TrimSpace(sessionID) {
		return true
	}
	if sourceSessionID, ok := payloadString(payload, "source_session_id"); ok && threadID == sourceSessionID {
		return true
	}
	return false
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

func newThreadID(projectID, sessionID, content string, now time.Time) string {
	sum := sha1.Sum([]byte(projectID + "|" + sessionID + "|" + content + "|" + now.UTC().Format(time.RFC3339Nano)))
	return "thread-" + hex.EncodeToString(sum[:8])
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

func deriveThreadState(item session.Session, records []event.Record, now time.Time) ThreadState {
	now = now.UTC()
	base, ok := latestLifecycleState(records)
	if !ok {
		base = fallbackThreadState(item)
	}
	return applyThreadStateOverlays(base, item, records, now)
}

func latestLifecycleState(records []event.Record) (ThreadState, bool) {
	for i := len(records) - 1; i >= 0; i-- {
		record := records[i]
		if state, ok := explicitThreadStateForRecord(record); ok {
			return state, true
		}
		if state, ok := lifecycleThreadState(record); ok {
			return state, true
		}
	}
	return "", false
}

func explicitThreadStateForRecord(record event.Record) (ThreadState, bool) {
	state, ok := explicitThreadState(record.Payload)
	if !ok {
		return "", false
	}
	if !isAuthoritativeThreadStateRecord(record, state) {
		return "", false
	}
	return state, true
}

func isAuthoritativeThreadStateRecord(record event.Record, state ThreadState) bool {
	if record.MessageType == event.MessageTypeHeartbeat {
		return false
	}
	if semanticKind, _ := record.Payload["semantic_kind"].(string); strings.TrimSpace(semanticKind) == "debug_event" {
		return false
	}
	if observed, _ := record.Payload["observed"].(bool); observed {
		return false
	}
	if state == ThreadStateRunning {
		switch {
		case record.MessageType == event.MessageTypeCommandResult:
			return false
		case record.MessageType == event.MessageTypeEvent &&
			record.EventType == event.TypeTerminalOutput:
			return false
		}
	}
	return true
}

func fallbackThreadState(item session.Session) ThreadState {
	switch item.Status {
	case session.StatusFailed:
		return ThreadStateBlocked
	case session.StatusStopped:
		return ThreadStateCompleted
	case session.StatusCreated:
		return ThreadStateWaitingPrompt
	default:
		return ThreadStateRunning
	}
}

func applyThreadStateOverlays(base ThreadState, item session.Session, records []event.Record, now time.Time) ThreadState {
	if isTerminalThreadState(base) {
		return base
	}
	if bridgeConnectivityKnown(item) && !item.BridgeOnline {
		return ThreadStateOffline
	}
	if base == ThreadStateRunning {
		activity := latestTime(item.LastActivityAt, item.UpdatedAt, item.CreatedAt, latestRecordTime(records))
		if !activity.IsZero() && now.Sub(activity) > activeThreadWindow {
			return ThreadStateStale
		}
	}
	return base
}

func bridgeConnectivityKnown(item session.Session) bool {
	return item.BridgeOnline || !item.BridgeConnectedAt.IsZero() || !item.BridgeDisconnectedAt.IsZero()
}

func isTerminalThreadState(state ThreadState) bool {
	switch state {
	case ThreadStateCompleted, ThreadStateBlocked:
		return true
	default:
		return false
	}
}

func latestRecordTime(records []event.Record) time.Time {
	var latest time.Time
	for _, record := range records {
		if record.Timestamp.After(latest) {
			latest = record.Timestamp
		}
	}
	return latest
}

func explicitThreadState(payload map[string]any) (ThreadState, bool) {
	value, ok := payloadString(payload, "thread_state")
	if !ok {
		return "", false
	}
	switch value {
	case string(ThreadStateRunning):
		return ThreadStateRunning, true
	case string(ThreadStateWaitingPrompt):
		return ThreadStateWaitingPrompt, true
	case string(ThreadStateWaitingReview):
		return ThreadStateWaitingReview, true
	case string(ThreadStateCompleted):
		return ThreadStateCompleted, true
	case string(ThreadStateBlocked):
		return ThreadStateBlocked, true
	case string(ThreadStateOffline):
		return ThreadStateOffline, true
	case string(ThreadStateStale):
		return ThreadStateStale, true
	default:
		return "", false
	}
}

func lifecycleThreadState(record event.Record) (ThreadState, bool) {
	switch {
	case record.EventType == event.TypeError:
		return ThreadStateBlocked, true
	case record.MessageType == event.MessageTypeCommandResult &&
		record.CommandType == event.CommandTypeSendPrompt &&
		record.Status == event.CommandStatusFailed:
		return ThreadStateWaitingPrompt, true
	case record.MessageType == event.MessageTypeCommand &&
		record.CommandType == event.CommandTypeSendPrompt:
		return ThreadStateRunning, true
	case record.MessageType == event.MessageTypeEvent &&
		record.EventType == event.TypeCommand &&
		payloadRole(record.Payload) == string(RoleUser):
		return ThreadStateRunning, true
	case record.MessageType == event.MessageTypeEvent &&
		record.EventType == event.TypeAIOutput:
		return ThreadStateRunning, true
	case record.MessageType == event.MessageTypeEvent &&
		record.EventType == event.TypeTerminalOutput:
		if observed, _ := record.Payload["observed"].(bool); observed {
			return "", false
		}
		if _, ok := payloadContent(record.Payload); ok {
			return ThreadStateRunning, true
		}
		return "", false
	default:
		return "", false
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
			AgentKind:  payloadAgentKind(record.Payload),
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
			AgentKind:  payloadAgentKind(record.Payload),
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
			AgentKind:  payloadAgentKind(record.Payload),
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
			AgentKind:  payloadAgentKind(record.Payload),
			SourceType: string(record.EventType),
		}, true
	default:
		return Message{}, false
	}
}

func payloadContent(payload map[string]any) (string, bool) {
	return payloadString(payload, "content")
}

func payloadAgentKind(payload map[string]any) string {
	return firstNonEmpty(
		payloadStringValue(payload, "agent_kind"),
		payloadStringValue(payload, "agent_name"),
	)
}

func payloadStringValue(payload map[string]any, key string) string {
	value, _ := payloadString(payload, key)
	return value
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

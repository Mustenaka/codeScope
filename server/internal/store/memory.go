package store

import (
	"sort"
	"sync"

	"codescope/server/internal/event"
	"codescope/server/internal/prompt"
	"codescope/server/internal/session"
)

type MemorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string]session.Session
}

func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{
		sessions: make(map[string]session.Session),
	}
}

func (s *MemorySessionStore) Create(record session.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[record.ID]; exists {
		return ErrConflict
	}
	s.sessions[record.ID] = record
	return nil
}

func (s *MemorySessionStore) List() ([]session.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]session.Session, 0, len(s.sessions))
	for _, record := range s.sessions {
		result = append(result, record)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})
	return result, nil
}

func (s *MemorySessionStore) Get(id string) (session.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, exists := s.sessions[id]
	if !exists {
		return session.Session{}, ErrNotFound
	}
	return record, nil
}

func (s *MemorySessionStore) Update(record session.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[record.ID]; !exists {
		return ErrNotFound
	}
	s.sessions[record.ID] = record
	return nil
}

type MemoryEventStore struct {
	mu     sync.RWMutex
	events []event.Record
}

func NewMemoryEventStore() *MemoryEventStore {
	return &MemoryEventStore{
		events: make([]event.Record, 0),
	}
}

func (s *MemoryEventStore) Append(record event.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.events = append(s.events, record)
	return nil
}

func (s *MemoryEventStore) ListBySession(sessionID string) ([]event.Record, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]event.Record, 0)
	for _, record := range s.events {
		if record.SessionID == sessionID {
			result = append(result, record)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.Before(result[j].Timestamp)
	})
	return result, nil
}

type MemoryPromptStore struct {
	mu      sync.RWMutex
	prompts []prompt.Template
}

func NewMemoryPromptStore() *MemoryPromptStore {
	return &MemoryPromptStore{
		prompts: make([]prompt.Template, 0),
	}
}

func (s *MemoryPromptStore) Create(record prompt.Template) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.prompts = append(s.prompts, record)
	return nil
}

func (s *MemoryPromptStore) List() ([]prompt.Template, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]prompt.Template, len(s.prompts))
	copy(result, s.prompts)
	return result, nil
}

type MemoryCommandTaskStore struct {
	mu    sync.RWMutex
	tasks map[string]map[string]session.CommandTask
}

func NewMemoryCommandTaskStore() *MemoryCommandTaskStore {
	return &MemoryCommandTaskStore{
		tasks: make(map[string]map[string]session.CommandTask),
	}
}

func (s *MemoryCommandTaskStore) Create(task session.CommandTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[task.SessionID]; !ok {
		s.tasks[task.SessionID] = make(map[string]session.CommandTask)
	}
	if _, ok := s.tasks[task.SessionID][task.ID]; ok {
		return ErrConflict
	}
	s.tasks[task.SessionID][task.ID] = task
	return nil
}

func (s *MemoryCommandTaskStore) ListBySession(sessionID string) ([]session.CommandTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bucket := s.tasks[sessionID]
	result := make([]session.CommandTask, 0, len(bucket))
	for _, task := range bucket {
		result = append(result, task)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})
	return result, nil
}

func (s *MemoryCommandTaskStore) Get(sessionID, taskID string) (session.CommandTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bucket, ok := s.tasks[sessionID]
	if !ok {
		return session.CommandTask{}, ErrNotFound
	}
	task, ok := bucket[taskID]
	if !ok {
		return session.CommandTask{}, ErrNotFound
	}
	return task, nil
}

func (s *MemoryCommandTaskStore) Update(task session.CommandTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	bucket, ok := s.tasks[task.SessionID]
	if !ok {
		return ErrNotFound
	}
	if _, ok := bucket[task.ID]; !ok {
		return ErrNotFound
	}
	bucket[task.ID] = task
	return nil
}

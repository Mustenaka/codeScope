package store

import (
	"errors"
	"fmt"
	"sync/atomic"

	"codescope/server/internal/event"
	"codescope/server/internal/prompt"
	"codescope/server/internal/session"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

type SessionStore interface {
	Create(session.Session) error
	List() ([]session.Session, error)
	Get(id string) (session.Session, error)
	Update(session.Session) error
}

type EventStore interface {
	Append(event.Record) error
	ListBySession(sessionID string) ([]event.Record, error)
}

type PromptStore interface {
	Create(prompt.Template) error
	List() ([]prompt.Template, error)
}

type CommandTaskStore interface {
	Create(session.CommandTask) error
	ListBySession(sessionID string) ([]session.CommandTask, error)
	Get(sessionID, taskID string) (session.CommandTask, error)
	Update(session.CommandTask) error
}

var idCounter atomic.Uint64

func NewID() string {
	value := idCounter.Add(1)
	return fmt.Sprintf("id-%d", value)
}

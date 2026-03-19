package session

import (
	"errors"
	"fmt"
	"log"
	"time"
)

type Store interface {
	Create(Session) error
	List() ([]Session, error)
	Get(id string) (Session, error)
	Update(Session) error
}

type Service struct {
	store Store
	now   func() time.Time
	idGen func() string
}

type BridgeMetadata struct {
	ProjectName   string
	WorkspaceRoot string
	MachineID     string
}

func NewService(store Store) *Service {
	return &Service{
		store: store,
		now:   time.Now().UTC,
		idGen: func() string {
			return fmt.Sprintf("session-%d", time.Now().UTC().UnixNano())
		},
	}
}

func (s *Service) Create(input CreateInput) (Session, error) {
	if input.ProjectName == "" {
		return Session{}, errors.New("project_name is required")
	}
	if input.WorkspaceRoot == "" {
		return Session{}, errors.New("workspace_root is required")
	}
	if input.MachineID == "" {
		return Session{}, errors.New("machine_id is required")
	}

	id := input.ID
	if id == "" {
		id = s.idGen()
	}

	now := s.now()
	record := Session{
		ID:             id,
		ProjectName:    input.ProjectName,
		WorkspaceRoot:  input.WorkspaceRoot,
		MachineID:      input.MachineID,
		Status:         StatusCreated,
		LastActivityAt: now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.store.Create(record); err != nil {
		return Session{}, fmt.Errorf("create session: %w", err)
	}
	return record, nil
}

func (s *Service) List() ([]Session, error) {
	records, err := s.store.List()
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	return records, nil
}

func (s *Service) Get(id string) (Session, error) {
	record, err := s.store.Get(id)
	if err != nil {
		return Session{}, fmt.Errorf("get session %s: %w", id, err)
	}
	return record, nil
}

func (s *Service) UpdateStatus(id string, status Status) (Session, error) {
	if !status.Valid() {
		return Session{}, fmt.Errorf("invalid session status %q", status)
	}

	record, err := s.store.Get(id)
	if err != nil {
		return Session{}, fmt.Errorf("get session %s: %w", id, err)
	}

	now := s.now()
	record.Status = status
	record.UpdatedAt = now
	switch status {
	case StatusRunning:
		if record.StartedAt.IsZero() {
			record.StartedAt = now
		}
		record.EndedAt = time.Time{}
	case StatusStopped, StatusFailed:
		record.EndedAt = now
	}

	if err := s.store.Update(record); err != nil {
		return Session{}, fmt.Errorf("update session %s: %w", id, err)
	}
	return record, nil
}

func (s *Service) EnsureBridgeSession(id string, metadata BridgeMetadata, at time.Time) (Session, bool, error) {
	record, err := s.store.Get(id)
	if err == nil {
		updated := applyBridgeMetadata(record, metadata)
		updated.BridgeOnline = true
		if updated.BridgeConnectedAt.IsZero() {
			updated.BridgeConnectedAt = at
		}
		updated.LastActivityAt = at
		updated.UpdatedAt = at
		if err := s.store.Update(updated); err != nil {
			return Session{}, false, fmt.Errorf("update bridge session %s: %w", id, err)
		}
		return updated, false, nil
	}

	record = Session{
		ID:                id,
		ProjectName:       firstNonEmpty(metadata.ProjectName, id),
		WorkspaceRoot:     firstNonEmpty(metadata.WorkspaceRoot, "."),
		MachineID:         firstNonEmpty(metadata.MachineID, "unknown-machine"),
		Status:            StatusCreated,
		BridgeOnline:      true,
		BridgeConnectedAt: at,
		LastActivityAt:    at,
		CreatedAt:         at,
		UpdatedAt:         at,
	}
	if err := s.store.Create(record); err != nil {
		reloaded, reloadErr := s.store.Get(id)
		if reloadErr != nil {
			return Session{}, false, fmt.Errorf("create bridge session %s: %w", id, err)
		}
		updated := applyBridgeMetadata(reloaded, metadata)
		updated.BridgeOnline = true
		if updated.BridgeConnectedAt.IsZero() {
			updated.BridgeConnectedAt = at
		}
		updated.LastActivityAt = at
		updated.UpdatedAt = at
		if updateErr := s.store.Update(updated); updateErr != nil {
			return Session{}, false, fmt.Errorf("update reloaded bridge session %s: %w", id, updateErr)
		}
		return updated, false, nil
	}

	log.Printf("session auto-created: session_id=%s workspace_root=%s machine_id=%s", record.ID, record.WorkspaceRoot, record.MachineID)
	return record, true, nil
}

func (s *Service) MarkBridgeDisconnected(id string, at time.Time) error {
	record, err := s.store.Get(id)
	if err != nil {
		return fmt.Errorf("get session %s: %w", id, err)
	}
	record.BridgeOnline = false
	record.BridgeDisconnectedAt = at
	record.UpdatedAt = at
	if err := s.store.Update(record); err != nil {
		return fmt.Errorf("update bridge disconnect %s: %w", id, err)
	}
	return nil
}

func (s *Service) TouchActivity(id string, at time.Time) error {
	record, err := s.store.Get(id)
	if err != nil {
		return fmt.Errorf("get session %s: %w", id, err)
	}
	if at.After(record.LastActivityAt) {
		record.LastActivityAt = at
	}
	if at.After(record.UpdatedAt) {
		record.UpdatedAt = at
	}
	if err := s.store.Update(record); err != nil {
		return fmt.Errorf("touch session %s: %w", id, err)
	}
	return nil
}

func applyBridgeMetadata(record Session, metadata BridgeMetadata) Session {
	record.ProjectName = firstNonEmpty(metadata.ProjectName, record.ProjectName)
	record.WorkspaceRoot = firstNonEmpty(metadata.WorkspaceRoot, record.WorkspaceRoot)
	record.MachineID = firstNonEmpty(metadata.MachineID, record.MachineID)
	return record
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

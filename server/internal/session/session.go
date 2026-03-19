package session

import "time"

type Status string

const (
	StatusCreated Status = "created"
	StatusRunning Status = "running"
	StatusStopped Status = "stopped"
	StatusFailed  Status = "failed"
)

type Session struct {
	ID                   string    `json:"id"`
	ProjectName          string    `json:"project_name"`
	WorkspaceRoot        string    `json:"workspace_root"`
	MachineID            string    `json:"machine_id"`
	Status               Status    `json:"status"`
	BridgeOnline         bool      `json:"bridge_online"`
	BridgeConnectedAt    time.Time `json:"bridge_connected_at,omitempty"`
	BridgeDisconnectedAt time.Time `json:"bridge_disconnected_at,omitempty"`
	LastActivityAt       time.Time `json:"last_activity_at,omitempty"`
	StartedAt            time.Time `json:"started_at,omitempty"`
	EndedAt              time.Time `json:"ended_at,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type CreateInput struct {
	ID            string `json:"id"`
	ProjectName   string `json:"project_name" binding:"required"`
	WorkspaceRoot string `json:"workspace_root" binding:"required"`
	MachineID     string `json:"machine_id" binding:"required"`
}

type UpdateStatusInput struct {
	Status Status `json:"status" binding:"required"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusCreated, StatusRunning, StatusStopped, StatusFailed:
		return true
	default:
		return false
	}
}
